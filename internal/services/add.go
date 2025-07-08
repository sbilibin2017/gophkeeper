package services

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- AddLoginPasswordHTTP with functional options ---

type addLoginPasswordHTTPConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   *resty.Client
}

type AddLoginPasswordHTTPOpt func(*addLoginPasswordHTTPConfig)

func WithAddLoginPasswordHTTPEncoders(enc []func([]byte) ([]byte, error)) AddLoginPasswordHTTPOpt {
	return func(c *addLoginPasswordHTTPConfig) {
		c.encoders = enc
	}
}

func WithAddLoginPasswordHTTPClient(client *resty.Client) AddLoginPasswordHTTPOpt {
	return func(c *addLoginPasswordHTTPConfig) {
		c.client = client
	}
}

func AddLoginPasswordHTTP(ctx context.Context, secret *models.LoginPassword, opts ...AddLoginPasswordHTTPOpt) error {
	config := &addLoginPasswordHTTPConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data string) (string, error) {
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

	// encode metadata keys and values
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

	data := map[string]interface{}{
		"secret_id": encodedSecretID,
		"login":     encodedLogin,
		"password":  encodedPassword,
		"meta":      encodedMeta,
	}

	resp, err := config.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post("/secrets/login-password")
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.String())
	}

	return nil
}

type addLoginPasswordGRPCConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   pb.AddLoginPasswordServiceClient
}

type AddLoginPasswordGRPCOpt func(*addLoginPasswordGRPCConfig)

func WithAddLoginPasswordGRPCEncoders(enc []func([]byte) ([]byte, error)) AddLoginPasswordGRPCOpt {
	return func(c *addLoginPasswordGRPCConfig) {
		c.encoders = enc
	}
}

func WithAddLoginPasswordGRPCClient(client pb.AddLoginPasswordServiceClient) AddLoginPasswordGRPCOpt {
	return func(c *addLoginPasswordGRPCConfig) {
		c.client = client
	}
}

func AddLoginPasswordGRPC(ctx context.Context, secret *models.LoginPassword, opts ...AddLoginPasswordGRPCOpt) error {
	config := &addLoginPasswordGRPCConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data string) (string, error) {
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

	req := &pb.LoginPassword{
		SecretId:  encodedSecretID,
		Login:     encodedLogin,
		Password:  encodedPassword,
		Meta:      encodedMeta,
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	resp, err := config.client.AddLoginPassword(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("AddLoginPassword failed: %s", resp.Error)
	}

	return nil
}

// --- AddTextSecretHTTP with functional options ---

type addTextSecretHTTPConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   *resty.Client
}

type AddTextSecretHTTPOpt func(*addTextSecretHTTPConfig)

func WithAddTextSecretHTTPEncoders(enc []func([]byte) ([]byte, error)) AddTextSecretHTTPOpt {
	return func(c *addTextSecretHTTPConfig) {
		c.encoders = enc
	}
}

func WithAddTextSecretHTTPClient(client *resty.Client) AddTextSecretHTTPOpt {
	return func(c *addTextSecretHTTPConfig) {
		c.client = client
	}
}

func AddTextSecretHTTP(ctx context.Context, secret *models.Text, opts ...AddTextSecretHTTPOpt) error {
	config := &addTextSecretHTTPConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data string) (string, error) {
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

	data := map[string]interface{}{
		"secret_id": encodedSecretID,
		"content":   encodedContent,
		"meta":      encodedMeta,
	}

	resp, err := config.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post("/secrets/text")
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.String())
	}

	return nil
}

// --- AddTextSecretGRPC with functional options ---

type addTextSecretGRPCConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   pb.AddTextServiceClient
}

type AddTextSecretGRPCOpt func(*addTextSecretGRPCConfig)

func WithAddTextSecretGRPCEncoders(enc []func([]byte) ([]byte, error)) AddTextSecretGRPCOpt {
	return func(c *addTextSecretGRPCConfig) {
		c.encoders = enc
	}
}

func WithAddTextSecretGRPCClient(client pb.AddTextServiceClient) AddTextSecretGRPCOpt {
	return func(c *addTextSecretGRPCConfig) {
		c.client = client
	}
}

func AddTextSecretGRPC(ctx context.Context, secret *models.Text, opts ...AddTextSecretGRPCOpt) error {
	config := &addTextSecretGRPCConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data string) (string, error) {
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

	req := &pb.Text{
		SecretId:  encodedSecretID,
		Content:   encodedContent,
		Meta:      encodedMeta,
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	resp, err := config.client.AddText(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("AddTextSecret failed: %s", resp.Error)
	}

	return nil
}

// --- AddBinarySecretHTTP with functional options ---

type addBinarySecretHTTPConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   *resty.Client
}

type AddBinarySecretHTTPOpt func(*addBinarySecretHTTPConfig)

func WithAddBinarySecretHTTPEncoders(enc []func([]byte) ([]byte, error)) AddBinarySecretHTTPOpt {
	return func(c *addBinarySecretHTTPConfig) {
		c.encoders = enc
	}
}

func WithAddBinarySecretHTTPClient(client *resty.Client) AddBinarySecretHTTPOpt {
	return func(c *addBinarySecretHTTPConfig) {
		c.client = client
	}
}

func AddBinarySecretHTTP(ctx context.Context, secret *models.Binary, opts ...AddBinarySecretHTTPOpt) error {
	config := &addBinarySecretHTTPConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data []byte) (string, error) {
		var err error
		for _, enc := range config.encoders {
			data, err = enc(data)
			if err != nil {
				return "", err
			}
		}
		return base64.StdEncoding.EncodeToString(data), nil
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

	data := map[string]interface{}{
		"secret_id": encodedSecretID,
		"data":      encodedData,
		"meta":      encodedMeta,
	}

	resp, err := config.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post("/secrets/binary")
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.String())
	}

	return nil
}

// --- AddBinarySecretGRPC with functional options ---

type addBinarySecretGRPCConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   pb.AddBinaryServiceClient
}

type AddBinarySecretGRPCOpt func(*addBinarySecretGRPCConfig)

func WithAddBinarySecretGRPCEncoders(enc []func([]byte) ([]byte, error)) AddBinarySecretGRPCOpt {
	return func(c *addBinarySecretGRPCConfig) {
		c.encoders = enc
	}
}

func WithAddBinarySecretGRPCClient(client pb.AddBinaryServiceClient) AddBinarySecretGRPCOpt {
	return func(c *addBinarySecretGRPCConfig) {
		c.client = client
	}
}

func AddBinarySecretGRPC(ctx context.Context, secret *models.Binary, opts ...AddBinarySecretGRPCOpt) error {
	config := &addBinarySecretGRPCConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data []byte) ([]byte, error) {
		var err error
		for _, enc := range config.encoders {
			data, err = enc(data)
			if err != nil {
				return nil, err
			}
		}
		return data, nil
	}

	// Encode secretID, then convert to base64 string (since pb.Binary.SecretId is string)
	encodedSecretIDBytes, err := encode([]byte(secret.SecretID))
	if err != nil {
		return fmt.Errorf("encoding secret ID failed: %w", err)
	}
	encodedSecretID := base64.StdEncoding.EncodeToString(encodedSecretIDBytes)

	// Encode data (pb.Binary.Data is []byte), assign directly
	encodedData, err := encode(secret.Data)
	if err != nil {
		return fmt.Errorf("encoding binary data failed: %w", err)
	}

	// Encode meta keys and values, convert encoded bytes to base64 strings (since meta map[string]string)
	encodedMeta := make(map[string]string, len(secret.Meta))
	for k, v := range secret.Meta {
		ekBytes, err := encode([]byte(k))
		if err != nil {
			return fmt.Errorf("encoding meta key failed: %w", err)
		}
		evBytes, err := encode([]byte(v))
		if err != nil {
			return fmt.Errorf("encoding meta value failed: %w", err)
		}
		ek := base64.StdEncoding.EncodeToString(ekBytes)
		ev := base64.StdEncoding.EncodeToString(evBytes)
		encodedMeta[ek] = ev
	}

	req := &pb.Binary{
		SecretId:  encodedSecretID, // string
		Data:      encodedData,     // []byte
		Meta:      encodedMeta,     // map[string]string
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	resp, err := config.client.AddBinary(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("AddBinary failed: %s", resp.Error)
	}

	return nil
}

// --- AddCardSecretHTTP with functional options ---

type addCardSecretHTTPConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   *resty.Client
}

type AddCardSecretHTTPOpt func(*addCardSecretHTTPConfig)

func WithAddCardSecretHTTPEncoders(enc []func([]byte) ([]byte, error)) AddCardSecretHTTPOpt {
	return func(c *addCardSecretHTTPConfig) {
		c.encoders = enc
	}
}

func WithAddCardSecretHTTPClient(client *resty.Client) AddCardSecretHTTPOpt {
	return func(c *addCardSecretHTTPConfig) {
		c.client = client
	}
}

func AddCardSecretHTTP(ctx context.Context, secret *models.Card, opts ...AddCardSecretHTTPOpt) error {
	config := &addCardSecretHTTPConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data string) (string, error) {
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
		return fmt.Errorf("encoding card number failed: %w", err)
	}
	encodedHolder, err := encode(secret.Holder)
	if err != nil {
		return fmt.Errorf("encoding cardholder failed: %w", err)
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

	data := map[string]interface{}{
		"secret_id": encodedSecretID,
		"number":    encodedNumber,
		"holder":    encodedHolder,
		"exp_month": secret.ExpMonth,
		"exp_year":  secret.ExpYear,
		"cvv":       encodedCVV,
		"meta":      encodedMeta,
	}

	resp, err := config.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post("/secrets/card")
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.String())
	}

	return nil
}

// --- AddCardSecretGRPC with functional options ---

type addCardSecretGRPCConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   pb.AddCardServiceClient
}

type AddCardSecretGRPCOpt func(*addCardSecretGRPCConfig)

func WithAddCardSecretGRPCEncoders(enc []func([]byte) ([]byte, error)) AddCardSecretGRPCOpt {
	return func(c *addCardSecretGRPCConfig) {
		c.encoders = enc
	}
}

func WithAddCardSecretGRPCClient(client pb.AddCardServiceClient) AddCardSecretGRPCOpt {
	return func(c *addCardSecretGRPCConfig) {
		c.client = client
	}
}

func AddCardSecretGRPC(ctx context.Context, secret *models.Card, opts ...AddCardSecretGRPCOpt) error {
	config := &addCardSecretGRPCConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data string) (string, error) {
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
		return fmt.Errorf("encoding card number failed: %w", err)
	}
	encodedHolder, err := encode(secret.Holder)
	if err != nil {
		return fmt.Errorf("encoding cardholder failed: %w", err)
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

	req := &pb.Card{
		SecretId:  encodedSecretID,
		Number:    encodedNumber,
		Holder:    encodedHolder,
		ExpMonth:  int32(secret.ExpMonth),
		ExpYear:   int32(secret.ExpYear),
		Cvv:       encodedCVV,
		Meta:      encodedMeta,
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	resp, err := config.client.AddCard(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("AddCardSecret failed: %s", resp.Error)
	}

	return nil
}
