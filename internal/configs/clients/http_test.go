package clients

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestCACert(t *testing.T, path string) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	certOut, err := os.Create(path)
	require.NoError(t, err)
	defer certOut.Close()

	require.NoError(t, pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))
}

func generateTestClientCertAndKey(t *testing.T, certPath, keyPath string) {
	// Генерация ключа клиента
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Простая заготовка сертификата (self-signed)
	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Test Client"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	// Запись сертификата
	certOut, err := os.Create(certPath)
	require.NoError(t, err)
	defer certOut.Close()
	require.NoError(t, pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}))

	// Запись ключа
	keyOut, err := os.Create(keyPath)
	require.NoError(t, err)
	defer keyOut.Close()
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	require.NoError(t, pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}))
}

func TestNewHTTPClient(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "ca.crt")
	clientCertPath := filepath.Join(tmpDir, "client.crt")
	clientKeyPath := filepath.Join(tmpDir, "client.key")

	generateTestCACert(t, certPath)
	generateTestClientCertAndKey(t, clientCertPath, clientKeyPath)

	tests := []struct {
		name        string
		baseURL     string
		options     []HTTPClientOption
		expectError bool
		assertFn    func(t *testing.T, c *resty.Client)
	}{
		{
			name:    "Default client",
			baseURL: "http://localhost",
			assertFn: func(t *testing.T, c *resty.Client) {
				assert.Equal(t, 3, c.RetryCount)
				assert.Equal(t, 500*time.Millisecond, c.RetryWaitTime)
				assert.Equal(t, 2*time.Second, c.RetryMaxWaitTime)
				assert.Equal(t, "http://localhost", c.BaseURL)
			},
		},
		{
			name:    "Custom retry count",
			baseURL: "http://localhost",
			options: []HTTPClientOption{
				WithHTTPRetryCount(5),
			},
			assertFn: func(t *testing.T, c *resty.Client) {
				assert.Equal(t, 5, c.RetryCount)
			},
		},
		{
			name:    "Custom retry wait time",
			baseURL: "http://localhost",
			options: []HTTPClientOption{
				WithHTTPRetryWaitTime(1 * time.Second),
			},
			assertFn: func(t *testing.T, c *resty.Client) {
				assert.Equal(t, 1*time.Second, c.RetryWaitTime)
			},
		},
		{
			name:    "Custom max retry wait time",
			baseURL: "http://localhost",
			options: []HTTPClientOption{
				WithHTTPRetryMaxWaitTime(10 * time.Second),
			},
			assertFn: func(t *testing.T, c *resty.Client) {
				assert.Equal(t, 10*time.Second, c.RetryMaxWaitTime)
			},
		},
		{
			name:    "Valid client TLS cert",
			baseURL: "https://localhost",
			options: []HTTPClientOption{
				WithHTTPTLSClientCert(clientCertPath, clientKeyPath),
			},
			assertFn: func(t *testing.T, c *resty.Client) {
				tlsCfg := c.GetClient().Transport.(*http.Transport).TLSClientConfig
				assert.NotNil(t, tlsCfg)
				assert.IsType(t, &tls.Config{}, tlsCfg)
				assert.NotEmpty(t, tlsCfg.Certificates)
			},
		},
		{
			name:    "Invalid client TLS cert path",
			baseURL: "https://localhost",
			options: []HTTPClientOption{
				WithHTTPTLSClientCert("nonexistent.crt", "nonexistent.key"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewHTTPClient(tt.baseURL, tt.options...)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if tt.assertFn != nil {
					tt.assertFn(t, client)
				}
			}
		})
	}
}
