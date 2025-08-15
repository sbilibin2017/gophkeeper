package models

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	DeviceID      string `json:"device_id"`
	Token         string `json:"token"`
	PrivateKeyPEM []byte `json:"private_key_pem"`
}

type LoginRequest struct {
	DeviceID string `json:"device_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
