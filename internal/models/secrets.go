package models

import "time"

const (
	SecretTypeLoginPassword = "login_password" // secret with login and password
	SecretTypeText          = "text"           // text secret
	SecretTypeBinary        = "binary"         // binary secret
	SecretTypeCard          = "card"           // secret containing bank card data
)

// LoginPassword represents a secret containing a user's login and password.
type LoginPassword struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // unique secret identifier
	Login     string            `json:"login" db:"login"`           // user login
	Password  string            `json:"password" db:"password"`     // user password
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // additional metadata
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // last update time
}

// Text represents a text secret containing arbitrary textual data.
type Text struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // unique secret identifier
	Content   string            `json:"content" db:"content"`       // textual content
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // additional metadata
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // last update time
}

// Binary represents a binary secret containing arbitrary binary data.
type Binary struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // unique secret identifier
	Data      []byte            `json:"data" db:"data"`             // binary data
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // additional metadata
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // last update time
}

// Card represents a secret containing bank card data.
type Card struct {
	SecretID  string            `json:"secret_id" db:"secret_id"`   // unique secret identifier
	Number    string            `json:"number" db:"number"`         // card number
	Holder    string            `json:"holder" db:"holder"`         // cardholder name
	ExpMonth  int               `json:"exp_month" db:"exp_month"`   // expiration month
	ExpYear   int               `json:"exp_year" db:"exp_year"`     // expiration year
	CVV       string            `json:"cvv" db:"cvv"`               // CVV code
	Meta      map[string]string `json:"meta,omitempty" db:"meta"`   // additional metadata
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"` // last update time
}
