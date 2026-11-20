ALTER TABLE redeem_codes ADD COLUMN IF NOT EXISTS reset_cap_usd NUMERIC(20,8) NOT NULL DEFAULT 0;

-- Only settled wallet spend (excluding temporary credit and overdraft) contributes.
CREATE TABLE IF NOT EXISTS wallet_reset_spend (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    eligible_usd NUMERIC(20,8) NOT NULL CHECK (eligible_usd >= 0),
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS wallet_reset_credits (
    redeem_code_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    amount_usd NUMERIC(20,8) NOT NULL CHECK (amount_usd > 0),
    remaining_usd NUMERIC(20,8) NOT NULL CHECK (remaining_usd >= 0 AND remaining_usd <= amount_usd),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_wallet_reset_credits_active
    ON wallet_reset_credits (user_id, group_id, expires_at, redeem_code_id) WHERE remaining_usd > 0;

CREATE INDEX IF NOT EXISTS idx_wallet_reset_credits_user_history
    ON wallet_reset_credits (user_id, redeem_code_id DESC);
