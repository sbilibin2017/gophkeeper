package models

// AuthRegisterRequest — структура запроса на регистрацию пользователя
type AuthRegisterRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	ClientPubKeyFile string `json:"client_pub_key_file"`
}

// AuthLoginRequest — структура запроса на вход пользователя
type AuthLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse — структура ответа с токеном после регистрации или входа
type AuthResponse struct {
	Token string `json:"token"`
}
