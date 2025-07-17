package models

// AddSecretBankCardRequest represents a request to add a bank card.
type BankCardAddRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Number     string  `json:"number"`         // Card number
	Owner      string  `json:"owner"`          // Card owner
	Exp        string  `json:"exp"`            // Expiration date
	CVV        string  `json:"cvv"`            // CVV code
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// GetSecretBankCardRequest represents a request to get a bank card.
type BankCardGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// GetSecretBankCardResponse represents a response with bank card information.
type BankCardResponse struct {
	SecretName  string  `json:"secret_name"`    // Secret name
	SecretOwner string  `json:"secret_owner"`   // Secret owner
	Number      string  `json:"number"`         // Card number
	Owner       string  `json:"owner"`          // Card owner
	Exp         string  `json:"exp"`            // Expiration date
	CVV         string  `json:"cvv"`            // CVV code
	Meta        *string `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   string  `json:"updated_at"`     // Last update timestamp
}

// ListSecretBankCardResponse contains a list of bank cards.
type BankCardListResponse struct {
	Items []BankCardResponse `json:"items"` // List of bank cards
}
