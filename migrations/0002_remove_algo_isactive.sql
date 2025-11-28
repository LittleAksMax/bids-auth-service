-- +goose Up
ALTER TABLE users
DROP COLUMN IF EXISTS password_algo,
    DROP COLUMN IF EXISTS is_active;

-- +goose Down
ALTER TABLE users
    ADD COLUMN password_algo TEXT NOT NULL DEFAULT 'argon2id',
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;