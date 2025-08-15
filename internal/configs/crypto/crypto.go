package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword хеширует пароль с помощью bcrypt.
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// GenerateRSAKeys создаёт RSA приватный и публичный ключи.
func GenerateRSAKeys(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

// GenerateDEK создаёт случайный Data Encryption Key указанной длины.
func GenerateDEK(size int) ([]byte, error) {
	dek := make([]byte, size)
	if _, err := rand.Read(dek); err != nil {
		return nil, err
	}
	return dek, nil
}

// EncryptDEK шифрует DEK с помощью публичного RSA ключа.
func EncryptDEK(pubKey *rsa.PublicKey, dek []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, dek, nil)
}

// RSAPrivateKeyToPEM конвертирует приватный RSA ключ в PEM формат.
func RSAPrivateKeyToPEM(privKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})
}
