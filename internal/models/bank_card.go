package models

import "time"

// BankCardAddRequest represents a request to add a bank card secret.
type BankCardAddRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Number     string  `json:"number"`         // Card number
	Owner      string  `json:"owner"`          // Name on the card
	Exp        string  `json:"exp"`            // Expiration date (MM/YY)
	CVV        string  `json:"cvv"`            // Card CVV code
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// BankCardGetRequest represents a request to retrieve a bank card secret by name.
type BankCardGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// BankCardResponse contains the retrieved bank card secret and metadata.
type BankCardResponse struct {
	SecretName  string    `json:"secret_name"`    // Secret name
	SecretOwner string    `json:"secret_owner"`   // Username of the secret's owner
	Number      string    `json:"number"`         // Card number
	Owner       string    `json:"owner"`          // Name on the card
	Exp         string    `json:"exp"`            // Expiration date (MM/YY)
	CVV         string    `json:"cvv"`            // Card CVV code
	Meta        *string   `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   time.Time `json:"updated_at"`     // Last modification timestamp
}

// BankCardListResponse contains a list of all bank card secrets.
type BankCardListResponse struct {
	Items []BankCardResponse `json:"items"` // List of bank card secrets
}
