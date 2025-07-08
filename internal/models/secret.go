package models

import "time"

const (
	SecretTypeLoginPassword = "login_password"
	SecretTypeText          = "text"
	SecretTypeBinary        = "binary"
	SecretTypeCard          = "card"
)

// --- LoginPassword ---

// LoginPassword represents login-password credentials with optional metadata.
type LoginPassword struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // Unique identifier for the secret
	Login     string            `json:"login" db:"login"`           // The login or username
	Password  string            `json:"password" db:"password"`     // The associated password
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // Optional metadata as key-value pairs
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// LoginPasswordOpt defines a functional option for configuring LoginPassword.
type LoginPasswordOpt func(*LoginPassword)

// NewLoginPassword creates a new LoginPassword with given options.
func NewLoginPassword(opts ...LoginPasswordOpt) *LoginPassword {
	lp := &LoginPassword{
		UpdatedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(lp)
	}
	return lp
}

func WithLoginPasswordSecretID(id string) LoginPasswordOpt {
	return func(lp *LoginPassword) {
		lp.SecretID = id
	}
}

func WithLoginPasswordLogin(login string) LoginPasswordOpt {
	return func(lp *LoginPassword) {
		lp.Login = login
	}
}

func WithLoginPasswordPassword(password string) LoginPasswordOpt {
	return func(lp *LoginPassword) {
		lp.Password = password
	}
}

func WithLoginPasswordMeta(meta map[string]string) LoginPasswordOpt {
	return func(lp *LoginPassword) {
		lp.Meta = meta
	}
}

func WithLoginPasswordUpdatedAt(t time.Time) LoginPasswordOpt {
	return func(lp *LoginPassword) {
		lp.UpdatedAt = t
	}
}

// --- Text ---

// Text represents a textual secret with optional metadata.
type Text struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // Unique identifier for the secret
	Content   string            `json:"content" db:"content"`       // The main text content of the secret
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // Optional metadata as key-value pairs
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// TextOpt defines a functional option for configuring Text.
type TextOpt func(*Text)

// NewText creates a new Text secret with given options.
func NewText(opts ...TextOpt) *Text {
	t := &Text{
		UpdatedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func WithTextSecretID(id string) TextOpt {
	return func(t *Text) {
		t.SecretID = id
	}
}

func WithTextContent(content string) TextOpt {
	return func(t *Text) {
		t.Content = content
	}
}

func WithTextMeta(meta map[string]string) TextOpt {
	return func(t *Text) {
		t.Meta = meta
	}
}

func WithTextUpdatedAt(tme time.Time) TextOpt {
	return func(t *Text) {
		t.UpdatedAt = tme
	}
}

// --- Binary ---

// Binary represents a binary secret (e.g., file data) with optional metadata.
type Binary struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // Unique identifier for the secret
	Data      []byte            `json:"data" db:"data"`             // Raw binary data
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // Optional metadata as key-value pairs
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// BinaryOpt defines a functional option for configuring Binary.
type BinaryOpt func(*Binary)

// NewBinary creates a new Binary secret with given options.
func NewBinary(opts ...BinaryOpt) *Binary {
	b := &Binary{
		UpdatedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func WithBinarySecretID(id string) BinaryOpt {
	return func(b *Binary) {
		b.SecretID = id
	}
}

func WithBinaryData(data []byte) BinaryOpt {
	return func(b *Binary) {
		b.Data = data
	}
}

func WithBinaryMeta(meta map[string]string) BinaryOpt {
	return func(b *Binary) {
		b.Meta = meta
	}
}

func WithBinaryUpdatedAt(tme time.Time) BinaryOpt {
	return func(b *Binary) {
		b.UpdatedAt = tme
	}
}

// --- Card ---

// Card represents sensitive card information with optional metadata.
type Card struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // Unique identifier for the secret
	Number    string            `json:"number" db:"number"`         // Card number
	Holder    string            `json:"holder" db:"holder"`         // Name of the cardholder
	ExpMonth  int               `json:"exp_month" db:"exp_month"`   // Expiration month (1–12)
	ExpYear   int               `json:"exp_year" db:"exp_year"`     // Expiration year (4-digit)
	CVV       string            `json:"cvv" db:"cvv"`               // Card verification value (CVV)
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // Optional metadata as key-value pairs
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // Timestamp of the last update
}

// CardOpt defines a functional option for configuring Card.
type CardOpt func(*Card)

// NewCard creates a new Card secret with given options.
func NewCard(opts ...CardOpt) *Card {
	c := &Card{
		UpdatedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithCardSecretID(id string) CardOpt {
	return func(c *Card) {
		c.SecretID = id
	}
}

func WithCardNumber(number string) CardOpt {
	return func(c *Card) {
		c.Number = number
	}
}

func WithCardHolder(holder string) CardOpt {
	return func(c *Card) {
		c.Holder = holder
	}
}

func WithCardExpMonth(month int) CardOpt {
	return func(c *Card) {
		c.ExpMonth = month
	}
}

func WithCardExpYear(year int) CardOpt {
	return func(c *Card) {
		c.ExpYear = year
	}
}

func WithCardCVV(cvv string) CardOpt {
	return func(c *Card) {
		c.CVV = cvv
	}
}

func WithCardMeta(meta map[string]string) CardOpt {
	return func(c *Card) {
		c.Meta = meta
	}
}

func WithCardUpdatedAt(tme time.Time) CardOpt {
	return func(c *Card) {
		c.UpdatedAt = tme
	}
}
