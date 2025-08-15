package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

type RSA struct {
	bitSize int
}

// New создаёт сервис для генерации ключей с указанным размером (2048, 4096 и т.д.)
func New(bitSize int) *RSA {
	return &RSA{bitSize: bitSize}
}

// GenerateRSAKeyPair генерирует пару ключей и возвращает KeyPair с PEM-кодированными ключами
func (s *RSA) Generate() (*models.RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, s.bitSize)
	if err != nil {
		return nil, err
	}

	// Кодирование приватного ключа в PEM
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Получение публичного ключа
	publicKey := &privateKey.PublicKey
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	// Кодирование публичного ключа в PEM
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	return &models.RSAKeyPair{
		PrivateKeyPEM: privPEM,
		PublicKeyPEM:  pubPEM,
	}, nil
}
