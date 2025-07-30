-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS secrets (
    secret_name TEXT NOT NULL,
    secret_type TEXT NOT NULL,
    secret_owner TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    ciphertext BLOB NOT NULL,
    aes_key_enc BLOB NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (secret_name, secret_type, secret_owner)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS secrets;
-- +goose StatementEnd
