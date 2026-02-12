-- +goose Up
ALTER TABLE users 
    ADD COLUMN role VARCHAR(5) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'));

-- +goose Down
ALTER TABLE users 
    DROP COLUMN role;

