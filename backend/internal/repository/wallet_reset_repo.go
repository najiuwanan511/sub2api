package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

func (r *userRepository) ActiveWalletResetCredit(ctx context.Context, userID, groupID int64) (float64, error) {
	rows, err := r.sql.QueryContext(ctx, `SELECT COALESCE(SUM(c.remaining_usd), 0)
		FROM wallet_reset_credits c JOIN groups g ON g.id = c.group_id
		WHERE c.user_id = $1 AND c.group_id = $2 AND c.remaining_usd > 0
		  AND c.expires_at > statement_timestamp() AND g.deleted_at IS NULL AND g.subscription_type = 'standard'`, userID, groupID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, rows.Err()
	}
	var balance float64
	err = rows.Scan(&balance)
	return balance, err
}

// Settle credits before permanent balance, earliest expiration first. The user
// lock serializes all credit settlement and card redemptions for this wallet.
func settleWalletResetBalance(ctx context.Context, tx *sql.Tx, userID, groupID int64, amount float64) (float64, bool, float64, error) {
	var before decimal.Decimal
	err := tx.QueryRowContext(ctx, `SELECT balance FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, userID).Scan(&before)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, 0, service.ErrUserNotFound
	}
	if err != nil {
		return 0, false, 0, err
	}
	cost := decimal.NewFromFloat(amount)
	var used decimal.Decimal
	err = tx.QueryRowContext(ctx, `WITH ranked AS (
		SELECT c.redeem_code_id, LEAST(c.remaining_usd, GREATEST(0, $3::numeric -
			COALESCE(SUM(c.remaining_usd) OVER (ORDER BY c.expires_at, c.redeem_code_id
			ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING), 0))) AS take
		FROM wallet_reset_credits c JOIN groups g ON g.id = c.group_id
		WHERE c.user_id = $1 AND c.group_id = $2 AND c.remaining_usd > 0
		  AND c.expires_at > statement_timestamp() AND g.deleted_at IS NULL AND g.subscription_type = 'standard'
	), used AS (
		UPDATE wallet_reset_credits c SET remaining_usd = c.remaining_usd - ranked.take
		FROM ranked WHERE c.redeem_code_id = ranked.redeem_code_id AND ranked.take > 0 RETURNING ranked.take
	) SELECT COALESCE(SUM(take), 0) FROM used`, userID, groupID, cost).Scan(&used)
	if err != nil {
		return 0, false, 0, err
	}
	cash := cost.Sub(used)
	paid := decimal.Min(cash, decimal.Max(before, decimal.Zero))
	if paid.Sign() > 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO wallet_reset_spend (user_id, group_id, eligible_usd)
			SELECT $1, id, $3 FROM groups WHERE id = $2 AND deleted_at IS NULL AND subscription_type = 'standard'
			ON CONFLICT (user_id, group_id) DO UPDATE SET eligible_usd = wallet_reset_spend.eligible_usd + EXCLUDED.eligible_usd`, userID, groupID, paid)
		if err != nil {
			return 0, false, 0, err
		}
	}
	if cash.Sign() > 0 {
		if _, err = tx.ExecContext(ctx, `UPDATE users SET balance = balance - $1, updated_at = NOW() WHERE id = $2`, cash, userID); err != nil {
			return 0, false, 0, err
		}
	}
	return before.Sub(cash).InexactFloat64(), cash.IsZero() || before.GreaterThanOrEqual(cash), used.InexactFloat64(), nil
}

func applyWalletBalanceCharge(ctx context.Context, tx *sql.Tx, cmd *service.UsageBillingCommand, result *service.UsageBillingApplyResult) error {
	var balance, credit float64
	var sufficient bool
	var err error
	if cmd.WalletGroupID > 0 {
		balance, sufficient, credit, err = settleWalletResetBalance(ctx, tx, cmd.UserID, cmd.WalletGroupID, cmd.BalanceCost)
	} else {
		balance, sufficient, err = deductUsageBillingBalance(ctx, tx, cmd.UserID, cmd.BalanceCost)
	}
	if err != nil {
		return err
	}
	result.NewBalance = &balance
	result.BalanceOverdrafted = !sufficient
	result.WalletResetCreditUsed = credit
	return nil
}
