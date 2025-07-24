package cryptor

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// helper to generate RSA keys for testing (2048 bits)
func generateRSAKeys(t *testing.T) (pubPEM, privPEM []byte) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)
	pubPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return pubPEM, privPEM
}

func TestCryptor_EncryptDecrypt(t *testing.T) {
	pubPEM, privPEM := generateRSAKeys(t)

	// Create Cryptor with public and private key options from PEM bytes (simulate file reading)
	c := &Cryptor{}

	// Load public key from PEM bytes (simulate WithPublicKeyFromCert but simplified)
	block, _ := pem.Decode(pubPEM)
	require.NotNil(t, block)
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err)
	rsaPub, ok := pubKey.(*rsa.PublicKey)
	require.True(t, ok)
	c.publicKey = rsaPub

	blockPriv, _ := pem.Decode(privPEM)
	require.NotNil(t, blockPriv)
	privKey, err := x509.ParsePKCS1PrivateKey(blockPriv.Bytes)
	require.NoError(t, err)
	c.privateKey = privKey

	// Test encryption
	plaintext := []byte("Hello, Testify!")
	encrypted, err := c.Encrypt(plaintext)
	require.NoError(t, err)
	require.NotNil(t, encrypted)
	require.NotEmpty(t, encrypted.Ciphertext)
	require.NotEmpty(t, encrypted.AESKeyEnc)

	// Test decryption
	decrypted, err := c.Decrypt(encrypted)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestCryptor_NewCryptor_WithOpts(t *testing.T) {
	pubPEM, privPEM := generateRSAKeys(t)

	// Save keys to temp files to test options reading from files
	pubFile := t.TempDir() + "/pub.pem"
	privFile := t.TempDir() + "/priv.pem"

	err := os.WriteFile(pubFile, pubPEM, 0644)
	require.NoError(t, err)
	err = os.WriteFile(privFile, privPEM, 0600)
	require.NoError(t, err)

	// Note: The public key in your original code expects a certificate, not just a public key PEM
	// So here we simulate a certificate by wrapping the public key in a self-signed cert for test.
	// But to keep it simple here, we'll skip the test for WithPublicKeyFromCert (since it expects cert).

	// Instead, let's test only WithPrivateKeyFromFile here, as example:

	c, err := NewCryptor(
		WithPrivateKeyFromFile(privFile),
	)
	require.NoError(t, err)
	require.NotNil(t, c.privateKey)
	require.Nil(t, c.publicKey)

	// Without public key, Encrypt should error
	_, err = c.Encrypt([]byte("data"))
	require.Error(t, err)

	// We can add more tests if you add a WithPublicKeyFromFile or similar

}

func TestCryptor_Encrypt_NoPublicKey(t *testing.T) {
	c := &Cryptor{}
	_, err := c.Encrypt([]byte("test"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "public key is not set")
}

func TestCryptor_Decrypt_NoPrivateKey(t *testing.T) {
	c := &Cryptor{}
	_, err := c.Decrypt(&Encrypted{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "private key is not set")
}

// helper: создаёт временный файл с содержимым, возвращает путь
func writeTempFile(t *testing.T, dir, name string, data []byte) string {
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, data, 0600)
	require.NoError(t, err)
	return path
}

// helper: генерирует самоподписанный сертификат с RSA ключом в PEM формате
func generateCertPEM(t *testing.T) []byte {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          newSerialNumber(t),
		Subject:               pkix.Name{CommonName: "Test Cert"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}

// helper: генерация серийного номера для сертификата
func newSerialNumber(t *testing.T) *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(t, err)
	return serialNumber
}

// helper: проверяет содержит ли строка хотя бы один из подстрок
func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func TestWithPublicKeyFromCert(t *testing.T) {
	tmpDir := t.TempDir()

	// 1) Корректный сертификат
	certPEM := generateCertPEM(t)
	certFile := writeTempFile(t, tmpDir, "cert.pem", certPEM)

	c := &Cryptor{}
	err := WithPublicKeyFromCert(certFile)(c)
	require.NoError(t, err)
	require.NotNil(t, c.publicKey)

	// 2) Файл отсутствует
	err = WithPublicKeyFromCert("not_exist.pem")(c)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read cert file")

	// 3) Некорректный PEM
	badPEM := []byte("not a pem")
	badFile := writeTempFile(t, tmpDir, "bad.pem", badPEM)
	err = WithPublicKeyFromCert(badFile)(c)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid PEM block in cert")

	// 4) PEM не сертификат
	notCertPEM := pem.EncodeToMemory(&pem.Block{Type: "NOTCERT", Bytes: []byte("data")})
	notCertFile := writeTempFile(t, tmpDir, "not_cert.pem", notCertPEM)
	err = WithPublicKeyFromCert(notCertFile)(c)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse certificate")

	// 5) Сертификат без RSA ключа — эмулируем ошибку (вводим неверные байты)
	tamperedPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte{1, 2, 3}})
	tamperedFile := writeTempFile(t, tmpDir, "tampered.pem", tamperedPEM)
	err = WithPublicKeyFromCert(tamperedFile)(c)
	require.Error(t, err)
	require.True(t, containsAny(err.Error(), []string{
		"failed to parse certificate",
		"certificate does not contain RSA public key",
	}))
}

func TestWithPrivateKeyFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Генерируем RSA приватный ключ PKCS#1
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privPKCS1 := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})
	privFilePKCS1 := writeTempFile(t, tmpDir, "priv_pkcs1.pem", privPKCS1)

	c := &Cryptor{}
	err = WithPrivateKeyFromFile(privFilePKCS1)(c)
	require.NoError(t, err)
	require.NotNil(t, c.privateKey)

	// PKCS#8 формат
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	require.NoError(t, err)
	privPKCS8 := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes})
	privFilePKCS8 := writeTempFile(t, tmpDir, "priv_pkcs8.pem", privPKCS8)

	c2 := &Cryptor{}
	err = WithPrivateKeyFromFile(privFilePKCS8)(c2)
	require.NoError(t, err)
	require.NotNil(t, c2.privateKey)

	// 2) Файл отсутствует
	err = WithPrivateKeyFromFile("not_exist.pem")(c)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read private key file")

	// 3) Некорректный PEM
	badPEM := []byte("not a pem")
	badFile := writeTempFile(t, tmpDir, "bad_priv.pem", badPEM)
	err = WithPrivateKeyFromFile(badFile)(c)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid PEM block in private key")

	// 4) Невалидные байты приватного ключа
	badPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte("bad bytes")})
	badPrivFile := writeTempFile(t, tmpDir, "bad_priv2.pem", badPrivPEM)
	err = WithPrivateKeyFromFile(badPrivFile)(c)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse private key")

	// 5) Приватный ключ не RSA (эмулируем)
	nonRSAPrivPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not rsa")})
	nonRSAFile := writeTempFile(t, tmpDir, "nonrsa.pem", nonRSAPrivPEM)
	err = WithPrivateKeyFromFile(nonRSAFile)(c)
	require.Error(t, err)
	require.True(t, containsAny(err.Error(), []string{
		"failed to parse private key",
		"not an RSA private key",
	}))
}
