package aesgsm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// AESGCM хранит AES ключ и позволяет шифровать/расшифровывать данные в режиме GCM.
type AESGCM struct {
	Key []byte
}

// New создаёт новый AESGCM с случайным ключом 256 бит.
func New() (*AESGCM, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return &AESGCM{Key: key}, nil
}

// Encrypt шифрует данные с помощью AES-GCM.
func (a *AESGCM) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	var block cipher.Block
	if block, err = aes.NewCipher(a.Key); err != nil {
		return
	}
	var gcm cipher.AEAD
	if gcm, err = cipher.NewGCM(block); err != nil {
		return
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return
}

// Decrypt расшифровывает данные с помощью AES-GCM.
func (a *AESGCM) Decrypt(ciphertext, nonce []byte) (plaintext []byte, err error) {
	var block cipher.Block
	if block, err = aes.NewCipher(a.Key); err != nil {
		return
	}
	var gcm cipher.AEAD
	if gcm, err = cipher.NewGCM(block); err != nil {
		return
	}

	plaintext, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err == nil && plaintext == nil {
		// Возвращаем пустой слайс вместо nil
		plaintext = []byte{}
	}
	return
}
