-- +goose Up
-- Assumes prior migrations already created pgcrypto/citext + users table + updated_at trigger.
-- No backfill needed (database is empty).

-- 1) Create password_credentials table (separate from users)
CREATE TABLE IF NOT EXISTS password_credentials (
                                                    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL CHECK (password_hash <> ''),
    password_salt TEXT NOT NULL CHECK (password_salt <> '')
    );

-- 2) Create refresh_tokens table (supports rotation via replaced_by_token_id)
CREATE TABLE IF NOT EXISTS refresh_tokens (
                                              token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE CHECK (token_hash <> ''),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    replaced_by_token_id UUID NULL REFERENCES refresh_tokens(token_id) ON DELETE SET NULL
    );

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS refresh_tokens_expires_at_idx ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS refresh_tokens_revoked_at_idx ON refresh_tokens(revoked_at);

-- 3) Remove password_hash from users (password now lives in password_credentials)
ALTER TABLE users
DROP COLUMN IF EXISTS password_hash;


-- +goose Down
-- Revert to previous (pre-simplification) layout: password_hash on users, no separate tables.

-- 1) Re-add password_hash to users (as it existed after Script 2)
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '' CHECK (password_hash <> '');

-- 2) Drop refresh tokens + password credentials
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS password_credentials;