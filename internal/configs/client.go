package configs

// ClientConfig contains the configuration for connecting to the server.
// Includes the server URL, path to the public RSA key file, and HMAC key.
type ClientConfig struct {
	ServerURL        string // ServerURL — address of the server to connect to.
	RSAPublicKeyPath string // RSAPublicKeyPath — path to the public RSA key file on the local filesystem.
	HMACKey          string // HMACKey — HMAC key for signing or authentication.
}

// ClientConfigOpt defines a function that modifies ClientConfig.
// Used to configure ClientConfig via functional options.
type ClientConfigOpt func(*ClientConfig)

// NewClientConfig creates a new instance of ClientConfig and applies the given options.
// Returns a pointer to the configured ClientConfig.
func NewClientConfig(opts ...ClientConfigOpt) *ClientConfig {
	config := &ClientConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// WithServerURL returns a ClientConfigOpt option that sets the ServerURL field.
func WithServerURL(u string) ClientConfigOpt {
	return func(c *ClientConfig) {
		c.ServerURL = u
	}
}

// WithRSAPublicKeyPath returns a ClientConfigOpt option that sets the path to the public RSA key.
func WithRSAPublicKeyPath(path string) ClientConfigOpt {
	return func(c *ClientConfig) {
		c.RSAPublicKeyPath = path
	}
}

// WithHMACKey returns a ClientConfigOpt option that sets the HMAC key.
func WithHMACKey(key string) ClientConfigOpt {
	return func(c *ClientConfig) {
		c.HMACKey = key
	}
}
