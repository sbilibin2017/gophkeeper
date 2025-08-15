-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS secrets (
    secret_id   TEXT PRIMARY KEY, -- уникальный идентификатор секрета (UUID)
    user_id     TEXT NOT NULL,    -- владелец секрета
    secret_type TEXT NOT NULL,    -- тип секрета: password, text, card, binary
    title       TEXT,             -- название или метка
    data        TEXT NOT NULL,    -- зашифрованные данные (Base64 или hex)
    meta        TEXT,             -- метаинформация, JSON
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS secrets;
-- +goose StatementEnd
