package models

import "errors"

// Credentials stores user credentials.
// Contains username and password.
type Credentials struct {
	Username string `json:"username"` // Username — the user's login name.
	Password string `json:"password"` // Password — the user's password.
}

// CredentialsOpt defines a functional option for configuring Credentials.
// Allows setting struct fields via options.
type CredentialsOpt func(*Credentials)

// NewCredentials creates a new Credentials object and applies the given options.
// Returns a pointer to the configured Credentials struct.
func NewCredentials(opts ...CredentialsOpt) *Credentials {
	c := &Credentials{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithUsername returns a CredentialsOpt option that sets the Username field.
func WithUsername(username string) CredentialsOpt {
	return func(c *Credentials) {
		c.Username = username
	}
}

// WithPassword returns a CredentialsOpt option that sets the Password field.
func WithPassword(password string) CredentialsOpt {
	return func(c *Credentials) {
		c.Password = password
	}
}

// SecretAddRequest defines the parameters required to add a new secret to the server.
type SecretAddRequest struct {
	ServerURL        string // ServerURL is the address of the remote server.
	SType            string // SType is the type of the secret (e.g., "login", "card", "file").
	File             string // File is the path to the file if the secret is file-based.
	Interactive      bool   // Interactive enables interactive input mode.
	HMACKey          string // HMACKey is the key used for HMAC encryption.
	RSAPublicKeyPath string // RSAPublicKeyPath is the path to the RSA public key for encryption.
}

// SecretAddOption defines a functional option for configuring SecretAddRequest.
type SecretAddOption func(*SecretAddRequest)

// NewSecretAddRequest creates a new SecretAddRequest and applies the given options.
func NewSecretAddRequest(opts ...SecretAddOption) (*SecretAddRequest, error) {
	req := &SecretAddRequest{}
	for _, opt := range opts {
		opt(req)
	}
	if req.File == "" && !req.Interactive {
		return nil, errors.New("either file or interactive must be specified")
	}
	if req.File != "" && req.Interactive {
		return nil, errors.New("file and interactive cannot be used together")
	}
	return req, nil
}

// WithServerURL returns a SecretAddOption that sets the ServerURL field.
func WithServerURL(url string) SecretAddOption {
	return func(r *SecretAddRequest) {
		r.ServerURL = url
	}
}

// WithSType returns a SecretAddOption that sets the SType field.
func WithSType(stype string) SecretAddOption {
	return func(r *SecretAddRequest) {
		r.SType = stype
	}
}

// WithFile returns a SecretAddOption that sets the File field.
func WithFile(file string) SecretAddOption {
	return func(r *SecretAddRequest) {
		r.File = file
	}
}

// WithInteractive returns a SecretAddOption that sets the Interactive field.
func WithInteractive(interactive bool) SecretAddOption {
	return func(r *SecretAddRequest) {
		r.Interactive = interactive
	}
}

// WithHMACKey returns a SecretAddOption that sets the HMACKey field.
func WithHMACKey(key string) SecretAddOption {
	return func(r *SecretAddRequest) {
		r.HMACKey = key
	}
}

// WithRSAPublicKeyPath returns a SecretAddOption that sets the RSAPublicKeyPath field.
func WithRSAPublicKeyPath(path string) SecretAddOption {
	return func(r *SecretAddRequest) {
		r.RSAPublicKeyPath = path
	}
}
