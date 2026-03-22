-- Add invite token and expiry to business_members
ALTER TABLE business_members
ADD COLUMN IF NOT EXISTS invite_token VARCHAR(255) UNIQUE,
ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_invite_token ON business_members (invite_token);
