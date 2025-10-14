-- +goose Up
CREATE TABLE users (
  id UUID PRIMARY KEY NOT NULL,
  created_at timestamp not null,
  updated_at timestamp not null,
  email text not null unique
);

-- +goose Down
drop table users;
