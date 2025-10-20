-- +goose Up
ALTER TABLE users ADD COLUMN hashed_password TEXT NOT NULL DEFAULT 'unset';


-- +goose Down
alter table users drop column hashed_password;
