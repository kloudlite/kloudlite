-- Migration: Add magic link tokens table for passwordless email authentication
-- Created: 2026-01-31

CREATE TABLE IF NOT EXISTS magic_link_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,
  token TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ip_address TEXT,
  user_agent TEXT
);

-- Indexes for fast token lookup and cleanup
CREATE INDEX IF NOT EXISTS idx_magic_link_tokens_token ON magic_link_tokens(token);
CREATE INDEX IF NOT EXISTS idx_magic_link_tokens_email ON magic_link_tokens(email);
CREATE INDEX IF NOT EXISTS idx_magic_link_tokens_expires_at ON magic_link_tokens(expires_at);

-- Optional: Add comment for documentation
COMMENT ON TABLE magic_link_tokens IS 'Stores one-time tokens for passwordless email authentication with 15-minute expiration';
COMMENT ON COLUMN magic_link_tokens.token IS 'Cryptographically secure 256-bit token (43 characters, base64url)';
COMMENT ON COLUMN magic_link_tokens.used_at IS 'Timestamp when token was verified - prevents reuse';
