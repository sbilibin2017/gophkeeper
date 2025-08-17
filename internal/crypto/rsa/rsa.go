package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// RSA хранит публичный и приватный ключи RSA и предоставляет методы для работы с ними.
type RSA struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

// New создаёт новый объект RSA с пустыми ключами.
func New() *RSA {
	return &RSA{}
}

// SetPrivateKey задаёт приватный ключ вручную
func (r *RSA) SetPrivateKey(priv *rsa.PrivateKey) {
	r.PrivateKey = priv
}

// SetPublicKey задаёт публичный ключ вручную
func (r *RSA) SetPublicKey(pub *rsa.PublicKey) {
	r.PublicKey = pub
}

// Generate создает новую пару ключей RSA и возвращает PEM-кодированные строки
func (r *RSA) GenerateKeyPair() (privatePEM string, publicPEM string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	r.PrivateKey = priv
	r.PublicKey = &priv.PublicKey

	// Приватный ключ
	privDER := x509.MarshalPKCS1PrivateKey(priv)
	privBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER}
	privatePEM = string(pem.EncodeToMemory(privBlock))

	// Публичный ключ
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}
	publicPEM = string(pem.EncodeToMemory(pubBlock))

	return privatePEM, publicPEM, nil
}

// Encrypt шифрует данные с помощью публичного ключа RSA
func (r *RSA) Encrypt(data []byte) ([]byte, error) {
	if r.PublicKey == nil {
		return nil, errors.New("public key is nil")
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, r.PublicKey, data, nil)
}

// Decrypt расшифровывает данные с помощью приватного ключа RSA
func (r *RSA) Decrypt(data []byte) ([]byte, error) {
	if r.PrivateKey == nil {
		return nil, errors.New("private key is nil")
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, r.PrivateKey, data, nil)
}

// ParsePublicKey парсит PEM публичный ключ и сохраняет его в структуре
func (r *RSA) ParsePublicKey(pemData []byte) error {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return errors.New("invalid public key PEM")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	pubKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}

	r.PublicKey = pubKey
	return nil
}

// ParsePrivateKey парсит PEM приватный ключ и сохраняет его в структуре
func (r *RSA) ParsePrivateKey(pemData []byte) error {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return errors.New("invalid private key PEM")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	r.PrivateKey = priv
	return nil
}
