package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrWalletResetCardInvalid = infraerrors.BadRequest("WALLET_RESET_CARD_INVALID", "Reset card requires a group, a percentage between 0 and 100, a positive USD cap, and 1–365 days of validity")
	ErrWalletResetCardGroup   = infraerrors.BadRequest("WALLET_RESET_CARD_GROUP", "Reset cards require a wallet-billed group and its group cannot be changed")
	ErrWalletResetNoSpend     = infraerrors.BadRequest("WALLET_RESET_NO_SPEND", "No eligible wallet spend is available for this group yet")
)

type WalletResetCredit struct {
	RedeemCodeID int64     `json:"redeem_code_id"`
	GroupID      *int64    `json:"group_id"`
	GroupName    string    `json:"group_name"`
	AmountUSD    float64   `json:"amount_usd"`
	RemainingUSD float64   `json:"remaining_usd"`
	ExpiresAt    time.Time `json:"expires_at"`
	Status       string    `json:"status"`
}

type WalletResetPreview struct {
	GroupName    string  `json:"group_name"`
	Percent      float64 `json:"percent"`
	EligibleUSD  float64 `json:"eligible_usd"`
	CreditUSD    float64 `json:"credit_usd"`
	ValidityDays int     `json:"validity_days"`
}

// WalletResetCreditReader is optional for substitute repositories used by service tests.
type WalletResetCreditReader interface {
	ActiveWalletResetCredit(ctx context.Context, userID, groupID int64) (float64, error)
}

func walletResetCredit(ctx context.Context, repo UserRepository, userID int64, groupID *int64) (float64, error) {
	if groupID == nil || *groupID <= 0 {
		return 0, nil
	}
	if r, ok := repo.(WalletResetCreditReader); ok {
		return r.ActiveWalletResetCredit(ctx, userID, *groupID)
	}
	return 0, nil
}

func validateWalletResetCard(groupID *int64, percent, capUSD float64, days int) error {
	if groupID == nil || *groupID <= 0 || !(percent >= 0.00000001 && percent <= 100) ||
		!(capUSD >= 0.00000001 && capUSD <= 1000000000) || days < 1 || days > 365 {
		return ErrWalletResetCardInvalid
	}
	return nil
}

func resetCardCredit(spend decimal.Decimal, percent, capUSD float64) decimal.Decimal {
	return decimal.Min(spend.Mul(decimal.NewFromFloat(percent)).Shift(-2).Truncate(UsageBillingMonetaryScale),
		decimal.NewFromFloat(capUSD).Truncate(UsageBillingMonetaryScale))
}

func (s *RedeemService) PreviewWalletResetCard(ctx context.Context, userID int64, code string) (*WalletResetPreview, error) {
	if err := s.checkRedeemRateLimit(ctx, userID); err != nil {
		return nil, err
	}
	card, err := s.redeemRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrRedeemCodeNotFound) {
			s.incrementRedeemErrorCount(ctx, userID)
		}
		return nil, err
	}
	if card.Type != RedeemTypeResetCard {
		return nil, unsupportedRedeemTypeError(card.Type)
	}
	if card.IsExpired() {
		return nil, ErrRedeemCodeExpired
	}
	if !card.CanUse() {
		return nil, ErrRedeemCodeUsed
	}
	if err := validateWalletResetCard(card.GroupID, card.Value, card.ResetCapUSD, card.ValidityDays); err != nil {
		return nil, err
	}
	spend, groupName, err := loadWalletResetSpend(ctx, s.entClient, userID, *card.GroupID)
	if err != nil {
		return nil, err
	}
	return &WalletResetPreview{GroupName: groupName, Percent: card.Value, EligibleUSD: spend.InexactFloat64(),
		CreditUSD: resetCardCredit(spend, card.Value, card.ResetCapUSD).InexactFloat64(), ValidityDays: card.ValidityDays}, nil
}

func loadWalletResetSpend(ctx context.Context, client *dbent.Client, userID, groupID int64) (decimal.Decimal, string, error) {
	rows, err := client.QueryContext(ctx, `
		SELECT COALESCE(s.eligible_usd, 0), g.name
		FROM groups g LEFT JOIN wallet_reset_spend s ON s.group_id = g.id AND s.user_id = $1
		WHERE g.id = $2 AND g.deleted_at IS NULL AND g.subscription_type = 'standard'`, userID, groupID)
	if err != nil {
		return decimal.Zero, "", err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return decimal.Zero, "", err
		}
		return decimal.Zero, "", ErrWalletResetCardGroup
	}
	var spend decimal.Decimal
	var groupName string
	err = rows.Scan(&spend, &groupName)
	return spend, groupName, err
}

// issueWalletResetCredit runs in the same transaction as the one-use code claim.
// The user's row is also locked by gateway settlement before it modifies spend or credits.
func issueWalletResetCredit(ctx context.Context, tx *dbent.Tx, userID int64, card *RedeemCode) (*WalletResetCredit, error) {
	if err := validateWalletResetCard(card.GroupID, card.Value, card.ResetCapUSD, card.ValidityDays); err != nil {
		return nil, err
	}
	client := tx.Client()
	rows, err := client.QueryContext(ctx, `SELECT id FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, userID)
	if err != nil {
		return nil, err
	}
	found := rows.Next()
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrUserNotFound
	}
	groupID := *card.GroupID
	spend, groupName, err := loadWalletResetSpend(ctx, client, userID, groupID)
	if err != nil {
		return nil, err
	}
	credit := resetCardCredit(spend, card.Value, card.ResetCapUSD)
	if credit.Sign() <= 0 {
		return nil, ErrWalletResetNoSpend
	}
	if _, err := client.ExecContext(ctx, `UPDATE wallet_reset_spend SET eligible_usd = 0 WHERE user_id = $1 AND group_id = $2`, userID, groupID); err != nil {
		return nil, fmt.Errorf("reset eligible spend: %w", err)
	}
	expires := time.Now().UTC().Add(time.Duration(card.ValidityDays) * 24 * time.Hour)
	_, err = client.ExecContext(ctx, `INSERT INTO wallet_reset_credits
		(redeem_code_id, user_id, group_id, amount_usd, remaining_usd, expires_at) VALUES ($1,$2,$3,$4,$4,$5)`,
		card.ID, userID, groupID, credit, expires)
	if err != nil {
		return nil, fmt.Errorf("issue reset card credit: %w", err)
	}
	amount := credit.InexactFloat64()
	return &WalletResetCredit{RedeemCodeID: card.ID, GroupID: &groupID, GroupName: groupName,
		AmountUSD: amount, RemainingUSD: amount, ExpiresAt: expires, Status: "active"}, nil
}

type WalletResetCreditSummary struct {
	ActiveUSD float64             `json:"active_usd"`
	Credits   []WalletResetCredit `json:"credits"`
}

func (s *RedeemService) ListWalletResetCredits(ctx context.Context, userID int64) (*WalletResetCreditSummary, error) {
	rows, err := s.entClient.QueryContext(ctx, `WITH credits AS (
		SELECT c.redeem_code_id, c.group_id, COALESCE(g.name,'') AS group_name,
		c.amount_usd, c.remaining_usd, c.expires_at,
		CASE WHEN c.remaining_usd = 0 THEN 'spent'
		     WHEN c.expires_at <= statement_timestamp() OR g.id IS NULL OR g.subscription_type <> 'standard' THEN 'expired'
		     ELSE 'active' END AS status
		FROM wallet_reset_credits c LEFT JOIN groups g ON g.id = c.group_id AND g.deleted_at IS NULL
		WHERE c.user_id = $1
	) SELECT redeem_code_id, group_id, group_name, amount_usd, remaining_usd, expires_at, status,
		COALESCE(SUM(remaining_usd) FILTER (WHERE status = 'active') OVER (), 0)
	FROM credits ORDER BY status ASC, redeem_code_id DESC LIMIT 100`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := &WalletResetCreditSummary{Credits: make([]WalletResetCredit, 0)}
	for rows.Next() {
		var c WalletResetCredit
		if err := rows.Scan(&c.RedeemCodeID, &c.GroupID, &c.GroupName, &c.AmountUSD, &c.RemainingUSD, &c.ExpiresAt, &c.Status, &out.ActiveUSD); err != nil {
			return nil, err
		}
		out.Credits = append(out.Credits, c)
	}
	return out, rows.Err()
}
