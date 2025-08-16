-- +goose Up
CREATE TABLE secret_keys (
    secret_key_id TEXT PRIMARY KEY,
    secret_id TEXT NOT NULL,
    device_id TEXT NOT NULL,
    encrypted_aes_key BLOB NOT NULL,  -- AES ключ, зашифрованный публичным ключом устройства
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (secret_id) REFERENCES secrets(secret_id),
    FOREIGN KEY (device_id) REFERENCES devices(device_id),
    UNIQUE(secret_id, device_id)       -- один ключ на одно устройство
);

-- +goose Down
DROP TABLE IF EXISTS secret_keys;