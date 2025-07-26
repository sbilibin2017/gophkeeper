package configs

// ClientConfig holds configuration settings for the GophKeeper client.
type ClientConfig struct {
	ServerURL        string  `json:"server_url"`
	ClientPubKeyFile string  `json:"client_pub_key_file"`
	Token            *string `json:"token,omitempty"`
}
