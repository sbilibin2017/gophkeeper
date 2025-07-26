package configs

import "time"

// ServerConfig holds configuration settings for the GophKeeper server.
type ServerConfig struct {
	ServerURL    string        `json:"server_url"`
	DatabaseDSN  string        `json:"database_dsn"`
	JWTSecretKey string        `json:"jwt_secret_key"`
	JWTExp       time.Duration `json:"jwt_exp"`
}
