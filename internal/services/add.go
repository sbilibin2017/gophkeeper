package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"gopkg.in/yaml.v3"
)

// ------- HTTP Options --------

type addHTTPOptions struct {
	HTTPClient *resty.Client
	Encoders   []func(data []byte) ([]byte, error)
	File       *os.File
}

type addHTTPOpt func(*addHTTPOptions)

func WithAddHTTPClient(client *resty.Client) addHTTPOpt {
	return func(opts *addHTTPOptions) {
		opts.HTTPClient = client
	}
}

func WithAddHTTPFile(file *os.File) addHTTPOpt {
	return func(opts *addHTTPOptions) {
		opts.File = file
	}
}

func WithAddHTTPHMACEncoder(enc func([]byte) []byte) addHTTPOpt {
	return func(opts *addHTTPOptions) {
		opts.Encoders = append(opts.Encoders, func(data []byte) ([]byte, error) {
			return enc(data), nil
		})
	}
}

func WithAddHTTPRSAEncoder(enc func([]byte) ([]byte, error)) addHTTPOpt {
	return func(opts *addHTTPOptions) {
		opts.Encoders = append(opts.Encoders, enc)
	}
}

// ------- GRPC Options --------

type addGRPCOptions struct {
	GRPCClient pb.AddServiceClient
	Encoders   []func(data []byte) ([]byte, error)
	File       *os.File
}

type addGRPCOpt func(*addGRPCOptions)

func WithAddGRPCClient(client pb.AddServiceClient) addGRPCOpt {
	return func(opts *addGRPCOptions) {
		opts.GRPCClient = client
	}
}

func WithAddGRPCFile(file *os.File) addGRPCOpt {
	return func(opts *addGRPCOptions) {
		opts.File = file
	}
}

func WithAddGRPCHMACEncoder(enc func([]byte) []byte) addGRPCOpt {
	return func(opts *addGRPCOptions) {
		opts.Encoders = append(opts.Encoders, func(data []byte) ([]byte, error) {
			return enc(data), nil
		})
	}
}

func WithAddGRPCRSAEncoder(enc func([]byte) ([]byte, error)) addGRPCOpt {
	return func(opts *addGRPCOptions) {
		opts.Encoders = append(opts.Encoders, enc)
	}
}

// -------- Helpers --------

func applyAddEncoders(data []byte, encoders []func([]byte) ([]byte, error)) ([]byte, error) {
	var err error
	for _, enc := range encoders {
		data, err = enc(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// Helper to encode a string field with encoders
func encodeAddField(field string, encoders []func([]byte) ([]byte, error)) (string, error) {
	encodedBytes, err := applyAddEncoders([]byte(field), encoders)
	if err != nil {
		return "", err
	}
	return string(encodedBytes), nil
}

func parseValidTill(validTill string) (int, int, error) {
	parts := strings.Split(validTill, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid valid till format, expected MM/YY")
	}
	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("invalid month in valid till")
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil || year < 0 || year > 99 {
		return 0, 0, fmt.Errorf("invalid year in valid till")
	}
	year += 2000
	return month, year, nil
}

func AddLoginPasswordHTTPFile(ctx context.Context, opts ...addHTTPOpt) error {
	options := &addHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.File == nil {
		return fmt.Errorf("file option must be provided")
	}

	data, err := io.ReadAll(options.File)
	if err != nil {
		return err
	}

	var secret models.LoginPassword
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}

	secret.UpdatedAt = time.Now().UTC()

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	// Inline encoder application (replaces applyAddEncoders)
	encodedData := rawData
	for _, enc := range options.Encoders {
		encodedData, err = enc(encodedData)
		if err != nil {
			return err
		}
	}

	fmt.Println("AddLoginPasswordHTTPFile Encoded Data:", string(encodedData))

	_, err = options.HTTPClient.R().
		SetContext(ctx).
		SetBody(encodedData).
		Post("/loginpassword")

	return err
}

func AddLoginPasswordHTTPInteractive(ctx context.Context, opts ...addHTTPOpt) error {
	options := &addHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var login, password string
	fmt.Print("Enter login: ")
	if _, err := fmt.Scanln(&login); err != nil {
		return err
	}
	fmt.Print("Enter password: ")
	if _, err := fmt.Scanln(&password); err != nil {
		return err
	}

	secret := &models.LoginPassword{
		Login:     login,
		Password:  password,
		UpdatedAt: time.Now().UTC(),
	}

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	// Inline encoder application (replaces applyAddEncoders)
	encodedData := rawData
	for _, enc := range options.Encoders {
		encodedData, err = enc(encodedData)
		if err != nil {
			return err
		}
	}

	fmt.Println("AddLoginPasswordHTTPInteractive Encoded Data:", string(encodedData))

	if options.HTTPClient == nil {
		return fmt.Errorf("HTTP client option must be provided")
	}

	_, err = options.HTTPClient.R().
		SetContext(ctx).
		SetBody(encodedData).
		Post("/loginpassword")

	return err
}

func AddLoginPasswordGRPCFile(ctx context.Context, opts ...addGRPCOpt) error {
	options := &addGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.File == nil {
		return fmt.Errorf("file option must be provided")
	}

	data, err := io.ReadAll(options.File)
	if err != nil {
		return err
	}

	var secret models.LoginPassword
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}

	secret.UpdatedAt = time.Now().UTC()

	// Inline encoding of login
	encodedLogin := []byte(secret.Login)
	for _, enc := range options.Encoders {
		encodedLogin, err = enc(encodedLogin)
		if err != nil {
			return fmt.Errorf("failed to encode login: %w", err)
		}
	}

	// Inline encoding of password
	encodedPassword := []byte(secret.Password)
	for _, enc := range options.Encoders {
		encodedPassword, err = enc(encodedPassword)
		if err != nil {
			return fmt.Errorf("failed to encode password: %w", err)
		}
	}

	// Inline encoding of each Meta value
	encodedMeta := make(map[string]string)
	for k, v := range secret.Meta {
		encoded := []byte(v)
		for _, enc := range options.Encoders {
			encoded, err = enc(encoded)
			if err != nil {
				return fmt.Errorf("failed to encode meta value for key %q: %w", k, err)
			}
		}
		encodedMeta[k] = string(encoded)
	}

	req := &pb.LoginPassword{
		SecretId:  secret.SecretID,
		Login:     string(encodedLogin),
		Password:  string(encodedPassword),
		Meta:      encodedMeta,
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	fmt.Println("AddLoginPasswordGRPCFile: Sending encrypted fields")

	_, err = options.GRPCClient.AddLoginPassword(ctx, req)
	return err
}

func AddLoginPasswordGRPCInteractive(ctx context.Context, opts ...addGRPCOpt) error {
	options := &addGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.GRPCClient == nil {
		return fmt.Errorf("grpc client option must be provided")
	}

	var login, password string
	fmt.Print("Enter login: ")
	if _, err := fmt.Scanln(&login); err != nil {
		return err
	}
	fmt.Print("Enter password: ")
	if _, err := fmt.Scanln(&password); err != nil {
		return err
	}

	// Inline encoding of login
	encodedLogin := []byte(login)
	var err error
	for _, enc := range options.Encoders {
		encodedLogin, err = enc(encodedLogin)
		if err != nil {
			return fmt.Errorf("failed to encode login: %w", err)
		}
	}

	// Inline encoding of password
	encodedPassword := []byte(password)
	for _, enc := range options.Encoders {
		encodedPassword, err = enc(encodedPassword)
		if err != nil {
			return fmt.Errorf("failed to encode password: %w", err)
		}
	}

	// Meta is empty in interactive mode
	encodedMeta := make(map[string]string)

	req := &pb.LoginPassword{
		Login:     string(encodedLogin),
		Password:  string(encodedPassword),
		Meta:      encodedMeta,
		UpdatedAt: time.Now().UTC().Unix(),
	}

	fmt.Println("AddLoginPasswordGRPCInteractive: Sending encrypted fields")

	_, err = options.GRPCClient.AddLoginPassword(ctx, req)
	return err
}

func AddTextHTTPFile(ctx context.Context, opts ...addHTTPOpt) error {
	options := &addHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.File == nil {
		return fmt.Errorf("file option must be provided")
	}

	data, err := io.ReadAll(options.File)
	if err != nil {
		return err
	}

	var secret models.Text
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}
	secret.UpdatedAt = time.Now().UTC()

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	// Inline encoder application
	encodedData := rawData
	for _, enc := range options.Encoders {
		encodedData, err = enc(encodedData)
		if err != nil {
			return fmt.Errorf("failed to encode data: %w", err)
		}
	}

	fmt.Println("AddTextHTTPFile Encoded Data:", string(encodedData))

	if options.HTTPClient == nil {
		return fmt.Errorf("HTTP client option must be provided")
	}

	_, err = options.HTTPClient.R().
		SetContext(ctx).
		SetBody(encodedData).
		Post("/text")

	return err
}

func AddTextHTTPInteractive(ctx context.Context, opts ...addHTTPOpt) error {
	options := &addHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var text string
	fmt.Print("Enter text: ")
	if _, err := fmt.Scanln(&text); err != nil {
		return err
	}

	secret := &models.Text{
		Content:   text,
		UpdatedAt: time.Now().UTC(),
	}

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	// Apply encoders inline
	encodedData := rawData
	for _, enc := range options.Encoders {
		encodedData, err = enc(encodedData)
		if err != nil {
			return fmt.Errorf("failed to encode data: %w", err)
		}
	}

	fmt.Println("AddTextHTTPInteractive Encoded Data:", string(encodedData))

	if options.HTTPClient == nil {
		return fmt.Errorf("HTTP client option must be provided")
	}

	_, err = options.HTTPClient.R().
		SetContext(ctx).
		SetBody(encodedData).
		Post("/text")

	return err
}

// -------- AddText GRPC --------

func AddTextGRPCFile(ctx context.Context, opts ...addGRPCOpt) error {
	options := &addGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.File == nil {
		return fmt.Errorf("file option must be provided")
	}

	data, err := io.ReadAll(options.File)
	if err != nil {
		return err
	}

	var secret models.Text
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}

	secret.UpdatedAt = time.Now().UTC()

	// Inline encoding of Content field
	encoded := []byte(secret.Content)
	for _, enc := range options.Encoders {
		encoded, err = enc(encoded)
		if err != nil {
			return fmt.Errorf("failed to encode content: %w", err)
		}
	}

	req := &pb.Text{
		Content:   string(encoded),
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	fmt.Println("AddTextGRPCFile: Sending encrypted fields")

	if options.GRPCClient == nil {
		return fmt.Errorf("grpc client option must be provided")
	}

	_, err = options.GRPCClient.AddText(ctx, req)
	return err
}

func AddTextGRPCInteractive(ctx context.Context, opts ...addGRPCOpt) error {
	options := &addGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.GRPCClient == nil {
		return fmt.Errorf("grpc client option must be provided")
	}

	var text string
	fmt.Print("Enter text: ")
	if _, err := fmt.Scanln(&text); err != nil {
		return err
	}

	// Inline encoding of text
	encoded := []byte(text)
	var err error
	for _, enc := range options.Encoders {
		encoded, err = enc(encoded)
		if err != nil {
			return fmt.Errorf("failed to encode content: %w", err)
		}
	}

	req := &pb.Text{
		Content:   string(encoded),
		UpdatedAt: time.Now().UTC().Unix(),
	}

	fmt.Println("AddTextGRPCInteractive: Sending encrypted fields")

	_, err = options.GRPCClient.AddText(ctx, req)
	return err
}

func AddCardHTTPFile(ctx context.Context, opts ...addHTTPOpt) error {
	options := &addHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.File == nil {
		return fmt.Errorf("file option must be provided")
	}

	data, err := io.ReadAll(options.File)
	if err != nil {
		return err
	}

	var secret models.Card
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}
	secret.UpdatedAt = time.Now().UTC()

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	// Inline encoding logic
	encoded := rawData
	for _, enc := range options.Encoders {
		encoded, err = enc(encoded)
		if err != nil {
			return err
		}
	}

	fmt.Println("AddCardHTTPFile Encoded Data:", string(encoded))

	if options.HTTPClient != nil {
		_, err = options.HTTPClient.R().
			SetContext(ctx).
			SetBody(encoded).
			Post("/card")
		return err
	}

	return fmt.Errorf("HTTP client option must be provided")
}

func AddCardHTTPInteractive(ctx context.Context, opts ...addHTTPOpt) error {
	options := &addHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var secretID, cardNumber, cardHolder, cvc string
	var expMonth, expYear int

	fmt.Print("Enter SecretID: ")
	if _, err := fmt.Scanln(&secretID); err != nil {
		return err
	}

	fmt.Print("Enter card number: ")
	if _, err := fmt.Scanln(&cardNumber); err != nil {
		return err
	}

	fmt.Print("Enter expiration month (1-12): ")
	if _, err := fmt.Scanln(&expMonth); err != nil {
		return err
	}
	if expMonth < 1 || expMonth > 12 {
		return fmt.Errorf("invalid expiration month: %d", expMonth)
	}

	fmt.Print("Enter expiration year (4-digit): ")
	if _, err := fmt.Scanln(&expYear); err != nil {
		return err
	}
	if expYear < 1000 || expYear > 9999 {
		return fmt.Errorf("invalid expiration year: %d", expYear)
	}

	fmt.Print("Enter card holder: ")
	if _, err := fmt.Scanln(&cardHolder); err != nil {
		return err
	}

	fmt.Print("Enter CVC: ")
	if _, err := fmt.Scanln(&cvc); err != nil {
		return err
	}

	meta := make(map[string]string)
	fmt.Println("Enter metadata key-value pairs (empty key to finish):")
	for {
		var key, value string
		fmt.Print("  Key: ")
		if _, err := fmt.Scanln(&key); err != nil {
			return err
		}
		if key == "" {
			break
		}
		fmt.Print("  Value: ")
		if _, err := fmt.Scanln(&value); err != nil {
			return err
		}
		meta[key] = value
	}

	secret := &models.Card{
		SecretID:  secretID,
		Number:    cardNumber,
		Holder:    cardHolder,
		ExpMonth:  expMonth,
		ExpYear:   expYear,
		CVV:       cvc,
		Meta:      meta,
		UpdatedAt: time.Now().UTC(),
	}

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	// Inline encoder application
	encoded := rawData
	for _, enc := range options.Encoders {
		encoded, err = enc(encoded)
		if err != nil {
			return err
		}
	}

	fmt.Println("AddCardHTTPInteractive Encoded Data:", string(encoded))

	if options.HTTPClient != nil {
		_, err = options.HTTPClient.R().
			SetContext(ctx).
			SetBody(encoded).
			Post("/card")
		return err
	}

	return fmt.Errorf("HTTP client option must be provided")
}

func AddCardGRPCFile(ctx context.Context, opts ...addGRPCOpt) error {
	options := &addGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.File == nil {
		return fmt.Errorf("file option must be provided")
	}
	if options.GRPCClient == nil {
		return fmt.Errorf("grpc client option must be provided")
	}

	data, err := io.ReadAll(options.File)
	if err != nil {
		return err
	}

	var secret models.Card
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}

	secret.UpdatedAt = time.Now().UTC()

	// Inline encoder application for Number
	encodedNumber := []byte(secret.Number)
	for _, enc := range options.Encoders {
		encodedNumber, err = enc(encodedNumber)
		if err != nil {
			return fmt.Errorf("failed to encode card number: %w", err)
		}
	}

	// Inline encoder application for Holder
	encodedHolder := []byte(secret.Holder)
	for _, enc := range options.Encoders {
		encodedHolder, err = enc(encodedHolder)
		if err != nil {
			return fmt.Errorf("failed to encode card holder: %w", err)
		}
	}

	// Inline encoder application for CVV
	encodedCVV := []byte(secret.CVV)
	for _, enc := range options.Encoders {
		encodedCVV, err = enc(encodedCVV)
		if err != nil {
			return fmt.Errorf("failed to encode CVV: %w", err)
		}
	}

	// Inline encoder application for each Meta value
	encodedMeta := make(map[string]string)
	for k, v := range secret.Meta {
		encodedValue := []byte(v)
		for _, enc := range options.Encoders {
			encodedValue, err = enc(encodedValue)
			if err != nil {
				return fmt.Errorf("failed to encode meta value for key %q: %w", k, err)
			}
		}
		encodedMeta[k] = string(encodedValue)
	}

	req := &pb.Card{
		SecretId:  secret.SecretID,
		Number:    string(encodedNumber),
		Holder:    string(encodedHolder),
		ExpMonth:  int32(secret.ExpMonth),
		ExpYear:   int32(secret.ExpYear),
		Cvv:       string(encodedCVV),
		Meta:      encodedMeta,
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	fmt.Println("AddCardGRPCFile: Sending encrypted proto request")

	_, err = options.GRPCClient.AddCard(ctx, req)
	return err
}

func AddCardGRPCInteractive(ctx context.Context, opts ...addGRPCOpt) error {
	options := &addGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var cardNumber, validTill, cardHolder, cvc string

	fmt.Print("Enter card number: ")
	if _, err := fmt.Scanln(&cardNumber); err != nil {
		return err
	}

	fmt.Print("Enter valid till (MM/YY): ")
	if _, err := fmt.Scanln(&validTill); err != nil {
		return err
	}

	fmt.Print("Enter card holder: ")
	if _, err := fmt.Scanln(&cardHolder); err != nil {
		return err
	}

	fmt.Print("Enter CVC: ")
	if _, err := fmt.Scanln(&cvc); err != nil {
		return err
	}

	month, year, err := parseValidTill(validTill)
	if err != nil {
		return err
	}

	encodeField := func(field string) (string, error) {
		encoded := []byte(field)
		for _, enc := range options.Encoders {
			var err error
			encoded, err = enc(encoded)
			if err != nil {
				return "", err
			}
		}
		return string(encoded), nil
	}

	encodedNumber, err := encodeField(cardNumber)
	if err != nil {
		return fmt.Errorf("failed to encode card number: %w", err)
	}

	encodedHolder, err := encodeField(cardHolder)
	if err != nil {
		return fmt.Errorf("failed to encode card holder: %w", err)
	}

	encodedCVV, err := encodeField(cvc)
	if err != nil {
		return fmt.Errorf("failed to encode CVV: %w", err)
	}

	req := &pb.Card{
		Number:    encodedNumber,
		ExpMonth:  int32(month),
		ExpYear:   int32(year),
		Holder:    encodedHolder,
		Cvv:       encodedCVV,
		UpdatedAt: time.Now().UTC().Unix(),
	}

	fmt.Println("AddCardGRPCInteractive: Sending encrypted proto fields")

	_, err = options.GRPCClient.AddCard(ctx, req)
	return err
}

// -------- AddBinary HTTP --------

func AddBinaryHTTPFile(ctx context.Context, opts ...addHTTPOpt) error {
	options := &addHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	data, err := io.ReadAll(options.File)
	if err != nil {
		return err
	}

	var secret models.Binary
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}

	secret.UpdatedAt = time.Now().UTC()

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	encodedData, err := applyAddEncoders(rawData, options.Encoders)
	if err != nil {
		return err
	}

	fmt.Println("AddBinaryHTTPFile Encoded Data:", string(encodedData))

	_, err = options.HTTPClient.R().
		SetContext(ctx).
		SetBody(encodedData).
		Post("/binary")

	return err
}

func AddBinaryHTTPInteractive(ctx context.Context, opts ...addHTTPOpt) error {
	options := &addHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var binaryData string
	fmt.Print("Enter binary data (base64 or raw string): ")
	if _, err := fmt.Scanln(&binaryData); err != nil {
		return err
	}

	secret := &models.Binary{
		Data:      []byte(binaryData),
		UpdatedAt: time.Now().UTC(),
	}

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	encodedData, err := applyAddEncoders(rawData, options.Encoders)
	if err != nil {
		return err
	}

	fmt.Println("AddBinaryHTTPInteractive Encoded Data:", string(encodedData))

	_, err = options.HTTPClient.R().
		SetContext(ctx).
		SetBody(encodedData).
		Post("/binary")

	return err
}

// -------- AddBinary GRPC --------

func AddBinaryGRPCFile(ctx context.Context, opts ...addGRPCOpt) error {
	options := &addGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	data, err := io.ReadAll(options.File)
	if err != nil {
		return err
	}

	var secret models.Binary
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}

	secret.UpdatedAt = time.Now().UTC()

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	encodedData, err := applyAddEncoders(rawData, options.Encoders)
	if err != nil {
		return err
	}

	req := &pb.Binary{
		Data:      encodedData,
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	fmt.Println("AddBinaryGRPCFile: Sending encoded data in proto message")

	_, err = options.GRPCClient.AddBinary(ctx, req)
	return err
}

func AddBinaryGRPCInteractive(ctx context.Context, opts ...addGRPCOpt) error {
	options := &addGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var binaryData string
	fmt.Print("Enter binary data (base64 or raw string): ")
	if _, err := fmt.Scanln(&binaryData); err != nil {
		return err
	}

	secret := &models.Binary{
		Data:      []byte(binaryData),
		UpdatedAt: time.Now().UTC(),
	}

	rawData, err := yaml.Marshal(secret)
	if err != nil {
		return err
	}

	encodedData, err := applyAddEncoders(rawData, options.Encoders)
	if err != nil {
		return err
	}

	req := &pb.Binary{
		Data:      encodedData, // Send encoded data here too
		UpdatedAt: secret.UpdatedAt.Unix(),
	}

	fmt.Println("AddBinaryGRPCInteractive: Sending encoded data in proto message")

	_, err = options.GRPCClient.AddBinary(ctx, req)
	return err
}
