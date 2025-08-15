-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    user_id       TEXT PRIMARY KEY,     -- уникальный идентификатор пользователя (UUID)
    username      TEXT NOT NULL UNIQUE, -- логин
    password_hash TEXT NOT NULL,        -- хэш пароля
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
