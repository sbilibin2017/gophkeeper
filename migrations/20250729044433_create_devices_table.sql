-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS devices (
    device_id      TEXT PRIMARY KEY,  -- уникальный идентификатор устройства (UUID)
    user_id        TEXT NOT NULL,     -- связка с пользователем
    device_name    TEXT,              -- например, "iPhone" или "PC"
    public_key     TEXT NOT NULL,     -- публичный ключ устройства
    encrypted_dek  TEXT NOT NULL,     -- DEK зашифрованный публичным ключом устройства
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS devices;
-- +goose StatementEnd
