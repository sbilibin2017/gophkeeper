package configs

// ClientConfig содержит конфигурацию для подключения к серверу.
// Включает URL сервера, путь к файлу публичного RSA-ключа и ключ HMAC.
type ClientConfig struct {
	ServerURL        string // ServerURL — адрес сервера для подключения.
	RSAPublicKeyPath string // RSAPublicKeyPath — путь к файлу публичного RSA-ключа на локальной файловой системе.
	HMACKey          string // HMACKey — ключ HMAC для подписи или аутентификации.
}

// ClientConfigOpt определяет функцию, изменяющую ClientConfig.
// Используется для настройки ClientConfig с помощью функциональных опций.
type ClientConfigOpt func(*ClientConfig)

// NewClientConfig создаёт новый экземпляр ClientConfig и применяет к нему переданные опции.
// Возвращает указатель на сконфигурированный ClientConfig.
func NewClientConfig(opts ...ClientConfigOpt) *ClientConfig {
	config := &ClientConfig{}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

// WithServerURL возвращает опцию ClientConfigOpt, которая устанавливает поле ServerURL.
func WithServerURL(u string) ClientConfigOpt {
	return func(c *ClientConfig) {
		c.ServerURL = u
	}
}

// WithRSAPublicKeyPath возвращает опцию ClientConfigOpt, которая устанавливает путь к публичному RSA-ключу.
func WithRSAPublicKeyPath(path string) ClientConfigOpt {
	return func(c *ClientConfig) {
		c.RSAPublicKeyPath = path
	}
}

// WithHMACKey возвращает опцию ClientConfigOpt, которая устанавливает ключ HMAC.
func WithHMACKey(key string) ClientConfigOpt {
	return func(c *ClientConfig) {
		c.HMACKey = key
	}
}
