package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// --- AddLoginPassword ---

type addLoginPasswordConfig struct {
	encoders []func([]byte) ([]byte, error)
	db       *sqlx.DB
}

type AddLoginPasswordOpt func(*addLoginPasswordConfig)

func WithAddLoginPasswordEncoders(enc []func([]byte) ([]byte, error)) AddLoginPasswordOpt {
	return func(c *addLoginPasswordConfig) {
		c.encoders = enc
	}
}

func WithAddLoginPasswordDB(db *sqlx.DB) AddLoginPasswordOpt {
	return func(c *addLoginPasswordConfig) {
		c.db = db
	}
}

func AddLoginPassword(ctx context.Context, secret *models.LoginPassword, opts ...AddLoginPasswordOpt) error {
	config := &addLoginPasswordConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.db == nil {
		return fmt.Errorf("database client is not configured")
	}

	encode := func(data string) (string, error) {
		if len(config.encoders) == 0 {
			// No encoders: return raw data as is
			return data, nil
		}
		b := []byte(data)
		var err error
		for _, enc := range config.encoders {
			b, err = enc(b)
			if err != nil {
				return "", err
			}
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}

	encodedSecretID, err := encode(secret.SecretID)
	if err != nil {
		return fmt.Errorf("encoding secret ID failed: %w", err)
	}
	encodedLogin, err := encode(secret.Login)
	if err != nil {
		return fmt.Errorf("encoding login failed: %w", err)
	}
	encodedPassword, err := encode(secret.Password)
	if err != nil {
		return fmt.Errorf("encoding password failed: %w", err)
	}

	encodedMeta := make(map[string]string, len(secret.Meta))
	for k, v := range secret.Meta {
		ek, err := encode(k)
		if err != nil {
			return fmt.Errorf("encoding meta key failed: %w", err)
		}
		ev, err := encode(v)
		if err != nil {
			return fmt.Errorf("encoding meta value failed: %w", err)
		}
		encodedMeta[ek] = ev
	}

	metaJSON, err := json.Marshal(encodedMeta)
	if err != nil {
		return fmt.Errorf("marshalling meta JSON failed: %w", err)
	}

	query := `
		INSERT INTO login_passwords (secret_id, login, password, meta)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_id) DO UPDATE SET
			login = EXCLUDED.login,
			password = EXCLUDED.password,
			meta = EXCLUDED.meta
	`

	_, err = config.db.ExecContext(ctx, query, encodedSecretID, encodedLogin, encodedPassword, string(metaJSON))
	if err != nil {
		return fmt.Errorf("db insert/update failed: %w", err)
	}

	return nil
}

// --- AddText ---

type addTextConfig struct {
	encoders []func([]byte) ([]byte, error)
	db       *sqlx.DB
}

type AddTextOpt func(*addTextConfig)

func WithAddTextEncoders(enc []func([]byte) ([]byte, error)) AddTextOpt {
	return func(c *addTextConfig) {
		c.encoders = enc
	}
}

func WithAddTextDB(db *sqlx.DB) AddTextOpt {
	return func(c *addTextConfig) {
		c.db = db
	}
}

func AddText(ctx context.Context, secret *models.Text, opts ...AddTextOpt) error {
	config := &addTextConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.db == nil {
		return fmt.Errorf("database client is not configured")
	}

	encode := func(data string) (string, error) {
		if len(config.encoders) == 0 {
			return data, nil
		}
		b := []byte(data)
		var err error
		for _, enc := range config.encoders {
			b, err = enc(b)
			if err != nil {
				return "", err
			}
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}

	encodedSecretID, err := encode(secret.SecretID)
	if err != nil {
		return fmt.Errorf("encoding secret ID failed: %w", err)
	}
	encodedContent, err := encode(secret.Content)
	if err != nil {
		return fmt.Errorf("encoding content failed: %w", err)
	}

	encodedMeta := make(map[string]string, len(secret.Meta))
	for k, v := range secret.Meta {
		ek, err := encode(k)
		if err != nil {
			return fmt.Errorf("encoding meta key failed: %w", err)
		}
		ev, err := encode(v)
		if err != nil {
			return fmt.Errorf("encoding meta value failed: %w", err)
		}
		encodedMeta[ek] = ev
	}

	metaJSON, err := json.Marshal(encodedMeta)
	if err != nil {
		return fmt.Errorf("marshalling meta JSON failed: %w", err)
	}

	query := `
		INSERT INTO texts (secret_id, content, meta, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_id) DO UPDATE SET
			content = EXCLUDED.content,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err = config.db.ExecContext(ctx, query, encodedSecretID, encodedContent, string(metaJSON), secret.UpdatedAt)
	if err != nil {
		return fmt.Errorf("db insert/update failed: %w", err)
	}

	return nil
}

// --- AddCard ---

type addCardConfig struct {
	encoders []func([]byte) ([]byte, error)
	db       *sqlx.DB
}

type AddCardOpt func(*addCardConfig)

func WithAddCardEncoders(enc []func([]byte) ([]byte, error)) AddCardOpt {
	return func(c *addCardConfig) {
		c.encoders = enc
	}
}

func WithAddCardDB(db *sqlx.DB) AddCardOpt {
	return func(c *addCardConfig) {
		c.db = db
	}
}

func AddCard(ctx context.Context, secret *models.Card, opts ...AddCardOpt) error {
	config := &addCardConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.db == nil {
		return fmt.Errorf("database client is not configured")
	}

	encode := func(data string) (string, error) {
		if len(config.encoders) == 0 {
			return data, nil
		}
		b := []byte(data)
		var err error
		for _, enc := range config.encoders {
			b, err = enc(b)
			if err != nil {
				return "", err
			}
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}

	encodedSecretID, err := encode(secret.SecretID)
	if err != nil {
		return fmt.Errorf("encoding secret ID failed: %w", err)
	}
	encodedNumber, err := encode(secret.Number)
	if err != nil {
		return fmt.Errorf("encoding number failed: %w", err)
	}
	encodedHolder, err := encode(secret.Holder)
	if err != nil {
		return fmt.Errorf("encoding holder failed: %w", err)
	}
	encodedCVV, err := encode(secret.CVV)
	if err != nil {
		return fmt.Errorf("encoding CVV failed: %w", err)
	}

	encodedMeta := make(map[string]string, len(secret.Meta))
	for k, v := range secret.Meta {
		ek, err := encode(k)
		if err != nil {
			return fmt.Errorf("encoding meta key failed: %w", err)
		}
		ev, err := encode(v)
		if err != nil {
			return fmt.Errorf("encoding meta value failed: %w", err)
		}
		encodedMeta[ek] = ev
	}

	metaJSON, err := json.Marshal(encodedMeta)
	if err != nil {
		return fmt.Errorf("marshalling meta JSON failed: %w", err)
	}

	query := `
		INSERT INTO cards (secret_id, number, holder, exp_month, exp_year, cvv, meta, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (secret_id) DO UPDATE SET
			number = EXCLUDED.number,
			holder = EXCLUDED.holder,
			exp_month = EXCLUDED.exp_month,
			exp_year = EXCLUDED.exp_year,
			cvv = EXCLUDED.cvv,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err = config.db.ExecContext(ctx, query,
		encodedSecretID,
		encodedNumber,
		encodedHolder,
		secret.ExpMonth,
		secret.ExpYear,
		encodedCVV,
		string(metaJSON),
		secret.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("db insert/update failed: %w", err)
	}

	return nil
}

// --- AddBinary ---

type addBinaryConfig struct {
	encoders []func([]byte) ([]byte, error)
	db       *sqlx.DB
}

type AddBinaryOpt func(*addBinaryConfig)

func WithAddBinaryEncoders(enc []func([]byte) ([]byte, error)) AddBinaryOpt {
	return func(c *addBinaryConfig) {
		c.encoders = enc
	}
}

func WithAddBinaryDB(db *sqlx.DB) AddBinaryOpt {
	return func(c *addBinaryConfig) {
		c.db = db
	}
}

func AddBinary(ctx context.Context, secret *models.Binary, opts ...AddBinaryOpt) error {
	config := &addBinaryConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.db == nil {
		return fmt.Errorf("database client is not configured")
	}

	encode := func(data []byte) (string, error) {
		if len(config.encoders) == 0 {
			// No encoders: return base64 encoded raw data
			return base64.StdEncoding.EncodeToString(data), nil
		}
		b := data
		var err error
		for _, enc := range config.encoders {
			b, err = enc(b)
			if err != nil {
				return "", err
			}
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}

	encodedSecretID, err := encode([]byte(secret.SecretID))
	if err != nil {
		return fmt.Errorf("encoding secret ID failed: %w", err)
	}

	encodedData, err := encode(secret.Data)
	if err != nil {
		return fmt.Errorf("encoding binary data failed: %w", err)
	}

	encodedMeta := make(map[string]string, len(secret.Meta))
	for k, v := range secret.Meta {
		ek, err := encode([]byte(k))
		if err != nil {
			return fmt.Errorf("encoding meta key failed: %w", err)
		}
		ev, err := encode([]byte(v))
		if err != nil {
			return fmt.Errorf("encoding meta value failed: %w", err)
		}
		encodedMeta[ek] = ev
	}

	metaJSON, err := json.Marshal(encodedMeta)
	if err != nil {
		return fmt.Errorf("marshalling meta JSON failed: %w", err)
	}

	query := `
		INSERT INTO binaries (secret_id, data, meta, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_id) DO UPDATE SET
			data = EXCLUDED.data,
			meta = EXCLUDED.meta,
			updated_at = EXCLUDED.updated_at
	`

	_, err = config.db.ExecContext(ctx, query,
		encodedSecretID,
		encodedData,
		string(metaJSON),
		secret.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("db insert/update failed: %w", err)
	}

	return nil
}
