-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS secrets (
    secret_name TEXT NOT NULL,
    secret_type TEXT NOT NULL,
    secret_owner TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    ciphertext BYTEA NOT NULL,
    aes_key_enc BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (secret_name, secret_type, secret_owner)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS secrets;
-- +goose StatementEnd
