package repository

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWalletResetCardPostgres(t *testing.T) {
	dsn := os.Getenv("SUB2API_WALLET_TEST_DSN")
	if dsn == "" {
		t.Skip("set SUB2API_WALLET_TEST_DSN to a disposable PostgreSQL test database")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, ApplyMigrations(ctx, db))
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	users := NewUserRepository(client, db)
	codes := NewRedeemCodeRepository(client)
	billing := NewUsageBillingRepository(client, db)
	redeem := service.NewRedeemService(codes, users, nil, nil, nil, client, nil, nil)

	suffix := uuid.NewString()
	user, err := client.User.Create().SetEmail(suffix + "@example.test").SetPasswordHash("hash").SetBalance(20).Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().SetName("reset-" + suffix).SetSubscriptionType("standard").Save(ctx)
	require.NoError(t, err)
	other, err := client.Group.Create().SetName("other-" + suffix).SetSubscriptionType("standard").Save(ctx)
	require.NoError(t, err)
	key, err := client.APIKey.Create().SetUserID(user.ID).SetGroupID(group.ID).SetKey("reset-" + suffix).SetName("wallet").Save(ctx)
	require.NoError(t, err)
	otherKey, err := client.APIKey.Create().SetUserID(user.ID).SetGroupID(other.ID).SetKey("other-" + suffix).SetName("wallet").Save(ctx)
	require.NoError(t, err)
	charge := func(keyID, groupID int64, amount float64) *service.UsageBillingCommand {
		return &service.UsageBillingCommand{RequestID: uuid.NewString(), APIKeyID: keyID, UserID: user.ID, WalletGroupID: groupID, BalanceCost: amount}
	}
	newCard := func() *service.RedeemCode {
		code, err := service.GenerateRedeemCode()
		require.NoError(t, err)
		card := &service.RedeemCode{Code: code, Type: service.RedeemTypeResetCard, Status: service.StatusUnused,
			GroupID: &group.ID, Value: 40, ResetCapUSD: 3, ValidityDays: 7}
		require.NoError(t, codes.Create(ctx, card))
		return card
	}
	card := newCard()
	_, err = redeem.ApplyWalletResetCard(ctx, user.ID, card.Code)
	require.ErrorIs(t, err, service.ErrWalletResetNoSpend)
	stored, err := codes.GetByID(ctx, card.ID)
	require.NoError(t, err)
	require.True(t, stored.CanUse(), "failed redemption must not consume the card")

	initial := charge(key.ID, group.ID, 10)
	_, err = billing.Apply(ctx, initial)
	require.NoError(t, err)
	duplicate, err := billing.Apply(ctx, initial)
	require.NoError(t, err)
	require.False(t, duplicate.Applied)
	preview, err := redeem.PreviewWalletResetCard(ctx, user.ID, card.Code)
	require.NoError(t, err)
	require.Equal(t, 10.0, preview.EligibleUSD)
	require.Equal(t, 3.0, preview.CreditUSD)

	// Different cards racing to redeem the same period cannot duplicate the credit.
	second := newCard()
	var wg sync.WaitGroup
	var issued [2]*service.WalletResetCredit
	var errs [2]error
	for i, candidate := range []*service.RedeemCode{card, second} {
		wg.Go(func() { issued[i], errs[i] = redeem.ApplyWalletResetCard(ctx, user.ID, candidate.Code) })
	}
	wg.Wait()
	win := 0
	if issued[0] == nil {
		win = 1
	}
	require.NoError(t, errs[win])
	require.ErrorIs(t, errs[1-win], service.ErrWalletResetNoSpend)
	require.Equal(t, 3.0, issued[win].AmountUSD)
	summary, err := redeem.ListWalletResetCredits(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 3.0, summary.ActiveUSD)
	require.Len(t, summary.Credits, 1)
	require.Equal(t, "active", summary.Credits[0].Status)
	winCode := []*service.RedeemCode{card, second}[win].Code
	_, err = redeem.ApplyWalletResetCard(ctx, user.ID, winCode)
	require.ErrorIs(t, err, service.ErrRedeemCodeUsed)

	outs, err := billing.ApplyBatch(ctx, []*service.UsageBillingCommand{
		charge(otherKey.ID, other.ID, 2), charge(key.ID, group.ID, 1), charge(key.ID, group.ID, 3),
	})
	require.NoError(t, err)
	require.Equal(t, 0.0, outs[0].Result.WalletResetCreditUsed, "credit cannot pay another group")
	require.Equal(t, 1.0, outs[1].Result.WalletResetCreditUsed)
	require.Equal(t, 2.0, outs[2].Result.WalletResetCreditUsed)
	require.Equal(t, 7.0, *outs[2].Result.NewBalance)

	third := newCard()
	preview, err = redeem.PreviewWalletResetCard(ctx, user.ID, third.Code)
	require.NoError(t, err)
	require.Equal(t, 1.0, preview.EligibleUSD, "credit-funded spend must not yield a new credit")
	credit, err := redeem.ApplyWalletResetCard(ctx, user.ID, third.Code)
	require.NoError(t, err)
	require.Equal(t, 0.4, credit.AmountUSD)
	reader := users.(service.WalletResetCreditReader)
	available, err := reader.ActiveWalletResetCredit(ctx, user.ID, group.ID)
	require.NoError(t, err)
	require.Equal(t, 0.4, available)
	summary, err = redeem.ListWalletResetCredits(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, 0.4, summary.ActiveUSD)
	require.Len(t, summary.Credits, 2)
	require.Equal(t, "spent", summary.Credits[1].Status)
	_, err = db.ExecContext(ctx, `UPDATE wallet_reset_credits SET expires_at = $1 WHERE redeem_code_id = $2`, time.Now().Add(-time.Hour), third.ID)
	require.NoError(t, err)
	available, err = reader.ActiveWalletResetCredit(ctx, user.ID, group.ID)
	require.NoError(t, err)
	require.Zero(t, available)
	result, err := billing.Apply(ctx, charge(key.ID, group.ID, 8))
	require.NoError(t, err)
	require.True(t, result.BalanceOverdrafted)
	require.Equal(t, -1.0, *result.NewBalance)
	require.Zero(t, result.WalletResetCreditUsed)
	preview, err = redeem.PreviewWalletResetCard(ctx, user.ID, newCard().Code)
	require.NoError(t, err)
	require.Equal(t, 7.0, preview.EligibleUSD, "overdraft is not settled spend")
	credits, err := redeem.ListWalletResetCredits(ctx, user.ID)
	require.NoError(t, err)
	require.Zero(t, credits.ActiveUSD)
	require.Len(t, credits.Credits, 2)
	require.Equal(t, "expired", credits.Credits[0].Status)
	require.Equal(t, "spent", credits.Credits[1].Status)
}
