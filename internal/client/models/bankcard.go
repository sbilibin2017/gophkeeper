package models

// BankCardAddRequest represents the request payload for adding a new bank card secret.
type BankCardAddRequest struct {
	SecretName string  `json:"secret_name,omitempty" validate:"required"`       // Unique name of the secret
	Number     string  `json:"number,omitempty" validate:"required,luhn"`       // Card number (Luhn validated)
	Owner      string  `json:"owner,omitempty" validate:"required"`             // Card owner name
	Exp        string  `json:"exp,omitempty" validate:"required,len=5"`         // Expiration date (e.g., MM/YY)
	CVV        string  `json:"cvv,omitempty" validate:"required,len=3,numeric"` // Card CVV code
	Meta       *string `json:"meta,omitempty"`                                  // Additional metadata or notes
}

// BankCardGetRequest represents the request to get a bank card secret by its name.
type BankCardGetRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to retrieve
}

// BankCardDeleteRequest represents the request to delete a bank card secret by its name.
type BankCardDeleteRequest struct {
	SecretName string `json:"secret_name,omitempty"` // Unique name of the secret to delete
}
