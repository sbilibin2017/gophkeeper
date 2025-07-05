package models

// Credentials хранит учетные данные пользователя.
// Содержит имя пользователя и пароль.
type Credentials struct {
	Username string `json:"username"` // Username — имя пользователя.
	Password string `json:"password"` // Password — пароль пользователя.
}

// CredentialsOpt определяет функциональный параметр для настройки Credentials.
// Позволяет задавать поля структуры через опции.
type CredentialsOpt func(*Credentials)

// NewCredentials создаёт новый объект Credentials и применяет к нему переданные опции.
// Возвращает указатель на сконфигурированную структуру Credentials.
func NewCredentials(opts ...CredentialsOpt) *Credentials {
	c := &Credentials{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithUsername возвращает опцию CredentialsOpt, которая задаёт поле Username.
func WithUsername(username string) CredentialsOpt {
	return func(c *Credentials) {
		c.Username = username
	}
}

// WithPassword возвращает опцию CredentialsOpt, которая задаёт поле Password.
func WithPassword(password string) CredentialsOpt {
	return func(c *Credentials) {
		c.Password = password
	}
}
