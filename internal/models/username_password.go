package models

// UsernamePassword хранит пару логин и пароль.
type UsernamePassword struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
