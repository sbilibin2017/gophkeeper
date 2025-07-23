package models

import "time"

// BankCard represents a bank card secret with sensitive information.
// It includes card details such as number, owner, expiry date, CVV, optional metadata,
// and an additional SecretOwner field to indicate the owner of this secret.
// UpdatedAt indicates the last time this secret was modified.
type BankCard struct {
	SecretName  string    `json:"secret_name" db:"secret_name"`   // Unique identifier for the secret
	SecretOwner string    `json:"secret_owner" db:"secret_owner"` // Owner of the secret (e.g., user ID or username)
	Number      string    `json:"number" db:"number"`             // Bank card number
	Owner       string    `json:"owner" db:"owner"`               // Card owner name
	Exp         string    `json:"exp" db:"exp"`                   // Expiry date of the card
	CVV         string    `json:"cvv" db:"cvv"`                   // Card CVV code
	Meta        *string   `json:"meta,omitempty" db:"meta"`       // Optional metadata
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`     // Timestamp of the last update
}

// GetSecretName returns the unique secret name of the bank card.
func (b *BankCard) GetSecretName() string {
	return b.SecretName
}

// GetUpdatedAt returns the timestamp when the bank card was last updated.
func (b *BankCard) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

// BankCardData contains bank card details excluding secret ownership information.
// It includes sensitive fields like number, owner, expiry date, CVV, and optional metadata.
type BankCardData struct {
	Number string  `json:"number"`         // Bank card number
	Owner  string  `json:"owner"`          // Card owner name
	Exp    string  `json:"exp"`            // Expiry date of the card
	CVV    string  `json:"cvv"`            // Card CVV code
	Meta   *string `json:"meta,omitempty"` // Optional metadata
}
