-- MFA/TOTP table
-- Run manually or add to Go migration runner

CREATE TABLE IF NOT EXISTS user_mfa (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    secret_encrypted TEXT NOT NULL,
    backup_codes_hash TEXT[],
    enabled_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_mfa_user_id ON user_mfa(user_id);
