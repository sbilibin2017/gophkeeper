-- +goose Up
CREATE TABLE secrets (
    secret_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    secret_name TEXT NOT NULL,            -- человекочитаемое имя секрета
    secret_type TEXT NOT NULL,            -- тип секрета: password, card, note и т.д.
    encrypted_payload BLOB NOT NULL,      -- зашифрованные данные AES
    nonce BLOB NOT NULL,                  -- nonce для AES-GCM
    meta TEXT,                            -- JSON метаданные
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    UNIQUE(user_id, secret_name)          -- уникальность имени секрета для пользователя
);

-- +goose Down
DROP TABLE IF EXISTS secrets;
