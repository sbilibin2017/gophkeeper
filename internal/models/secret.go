package models

// AddSecretBinaryRequest represents a request to add a binary secret.
type AddSecretBinaryRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Data       []byte  `json:"data"`           // Binary data
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// GetSecretBinaryRequest represents a request to get a binary secret.
type GetSecretBinaryRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// GetSecretBinaryResponse represents a response with a binary secret.
type GetSecretBinaryResponse struct {
	SecretName  string  `json:"secret_name"`    // Secret name
	SecretOwner string  `json:"secret_owner"`   // Secret owner
	Data        []byte  `json:"data"`           // Binary data
	Meta        *string `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   string  `json:"updated_at"`     // Last update timestamp
}

// ListSecretBinaryResponse contains a list of binary secrets.
type ListSecretBinaryResponse struct {
	Items []GetSecretBinaryResponse `json:"items"` // List of binary secrets
}

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

// TextAddRequest represents a request to add a text secret.
type TextAddRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Content    string  `json:"content"`        // Text content
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// TextGetRequest represents a request to get a text secret.
type TextGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// TextResponse represents a response with text secret information.
type TextResponse struct {
	SecretName  string  `json:"secret_name"`    // Secret name
	SecretOwner string  `json:"secret_owner"`   // Secret owner
	Content     string  `json:"content"`        // Text content
	Meta        *string `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   string  `json:"updated_at"`     // Last update timestamp
}

// TextListResponse contains a list of text secrets.
type TextListResponse struct {
	Items []TextResponse `json:"items"` // List of text secrets
}

// UsernamePasswordAddRequest represents a request to add a username-password secret.
type UsernamePasswordAddRequest struct {
	SecretName string  `json:"secret_name"`    // Secret name
	Username   string  `json:"username"`       // Username
	Password   string  `json:"password"`       // Password
	Meta       *string `json:"meta,omitempty"` // Optional metadata
}

// UsernamePasswordGetRequest represents a request to get a username-password secret.
type UsernamePasswordGetRequest struct {
	SecretName string `json:"secret_name"` // Secret name
}

// UsernamePasswordResponse represents a response with a username-password secret.
type UsernamePasswordResponse struct {
	SecretName  string  `json:"secret_name"`    // Secret name
	SecretOwner string  `json:"secret_owner"`   // Secret owner
	Username    string  `json:"username"`       // Username
	Password    string  `json:"password"`       // Password
	Meta        *string `json:"meta,omitempty"` // Optional metadata
	UpdatedAt   string  `json:"updated_at"`     // Last update timestamp
}

// UsernamePasswordListResponse contains a list of username-password secrets.
type UsernamePasswordListResponse struct {
	Items []UsernamePasswordResponse `json:"items"` // List of username-password secrets
}
