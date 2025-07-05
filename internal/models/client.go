package models

// Credentials stores user credentials.
// Contains username and password.
type Credentials struct {
	Username string `json:"username"` // Username — the user's name.
	Password string `json:"password"` // Password — the user's password.
}

// CredentialsOpt defines a functional parameter to configure Credentials.
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
