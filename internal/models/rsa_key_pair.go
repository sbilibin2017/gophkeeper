package models

// RSAKeyPair представляет пару ключей RSA в PEM-формате.
type RSAKeyPair struct {
	PrivateKeyPEM []byte `json:"private_key_pem"`
	PublicKeyPEM  []byte `json:"public_key_pem"`
}
