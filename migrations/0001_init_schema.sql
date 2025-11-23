-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $func$
BEGIN
    NEW.updated_at = NOW();
RETURN NEW;
END;
$func$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       username CITEXT NOT NULL UNIQUE CHECK (username <> ''),
                       email CITEXT NOT NULL UNIQUE CHECK (email <> ''),
                       password_hash TEXT NOT NULL CHECK (password_hash <> ''),
                       password_algo TEXT NOT NULL DEFAULT 'argon2id',
                       is_active BOOLEAN NOT NULL DEFAULT TRUE,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS users_set_updated_at ON users;

-- +goose StatementBegin
DROP FUNCTION IF EXISTS set_updated_at();
-- +goose StatementEnd

DROP TABLE IF EXISTS users;