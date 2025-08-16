/*
Package crypto предоставляет функции для симметричного и асимметричного шифрования данных,
генерации ключей и работы с RSA ключами в формате PEM.

AES (Advanced Encryption Standard) — это симметричный блочный шифр, который используется для
шифрования данных с одним секретным ключом.

GCM (Galois/Counter Mode) — это режим работы AES, обеспечивающий как конфиденциальность,
так и целостность данных. Для каждого шифрования используется nonce (одноразовый случайный вектор),
который обеспечивает уникальность шифрования для одного ключа и предотвращает повторное использование ключа
для одинаковых данных.

PEM (Privacy Enhanced Mail) — это текстовый формат кодирования ключей и сертификатов.
Он представляет бинарные данные в Base64, заключённые между заголовком и футером, например:
"-----BEGIN PUBLIC KEY-----" ... "-----END PUBLIC KEY-----". PEM широко используется для
обмена публичными и приватными ключами RSA.

В этом пакете:

- AESGCM.Key — случайный 256-битный ключ AES.
- AESGCM.Encrypt — шифрует данные (plaintext) с помощью ключа AESGCM.Key. Возвращает:
  - ciphertext — зашифрованные данные, которые безопасно передавать или хранить.
  - nonce — случайный одноразовый вектор, необходимый для расшифровки данных. Он должен храниться вместе с ciphertext.

- AESGCM.Decrypt — расшифровывает данные, используя ciphertext и nonce.

- RSAEncryptor.PublicKey — публичный ключ RSA для шифрования AES ключа.
- RSAEncryptor.PrivateKey — приватный ключ RSA для расшифровки AES ключа.
- RSAEncryptor.EncryptAESKey — шифрует AES ключ, чтобы его можно было безопасно передавать.
- RSAEncryptor.DecryptAESKey — расшифровывает AES ключ, полученный от EncryptAESKey.

Логика использования:
1. Генерируется AESGCM (случайный AES ключ 256 бит).
2. С помощью AESGCM.Key шифруются данные методом Encrypt.
3. AES ключ шифруется публичным ключом RSA с помощью EncryptAESKey.
4. Полученный ciphertext и nonce можно безопасно хранить/передавать вместе с зашифрованным AES ключом.
5. На стороне получателя AES ключ расшифровывается приватным ключом RSA с помощью DecryptAESKey.
6. AES ключ используется для расшифровки данных методом Decrypt, используя nonce.

- ParseRSAPublicKeyPEM и ParseRSAPrivateKeyPEM — функции для чтения ключей из PEM.
- GenerateID — генерация уникального идентификатора UUID v4.
*/

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"

	"github.com/google/uuid"
)

// AESGCM хранит AES ключ и позволяет шифровать/расшифровывать данные в режиме GCM.
type AESGCM struct {
	Key []byte
}

// NewAESGCM создаёт новый AESGCM с случайным ключом 256 бит.
func NewAESGCM() (aesGCM *AESGCM, err error) {
	key := make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		return
	}
	aesGCM = &AESGCM{Key: key}
	return
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
	return
}

// RSAEncryptor хранит публичный и/или приватный ключ для шифрования AES ключей.
type RSAEncryptor struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// NewRSAEncryptor создаёт новый RSAEncryptor с заданными ключами.
func NewRSAEncryptor(pub *rsa.PublicKey, priv *rsa.PrivateKey) (rsaEnc *RSAEncryptor) {
	rsaEnc = &RSAEncryptor{
		PublicKey:  pub,
		PrivateKey: priv,
	}
	return
}

// EncryptAESKey шифрует AES ключ с помощью публичного ключа RSA.
func (r *RSAEncryptor) EncryptAESKey(aesKey []byte) (encryptedKey []byte, err error) {
	encryptedKey, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, r.PublicKey, aesKey, nil)
	return
}

// DecryptAESKey расшифровывает AES ключ с помощью приватного ключа RSA.
func (r *RSAEncryptor) DecryptAESKey(encryptedAESKey []byte) (aesKey []byte, err error) {
	aesKey, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, r.PrivateKey, encryptedAESKey, nil)
	return
}

// ParseRSAPublicKeyPEM парсит PEM публичный ключ RSA.
func ParseRSAPublicKeyPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key PEM")
	}
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return parsedKey.(*rsa.PublicKey), nil
}

// ParseRSAPrivateKeyPEM парсит PEM приватный ключ RSA.
func ParseRSAPrivateKeyPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("invalid private key PEM")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// GenerateID генерирует случайный уникальный UUID v4.
func GenerateID() string {
	return uuid.New().String()
}
