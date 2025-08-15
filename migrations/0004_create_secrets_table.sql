-- +goose Up
CREATE TABLE secrets (
    secret_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL,
    payload_encrypted_with_symmetric_key BLOB NOT NULL, -- данные, зашифрованные симметричным ключом.
    meta TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- +goose Down
DROP TABLE IF EXISTS secrets;
