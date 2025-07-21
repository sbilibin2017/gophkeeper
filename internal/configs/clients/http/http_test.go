package http

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestTLSCert creates a self-signed TLS certificate and key, writes them to temp files, and returns their paths.
func generateTestTLSCert() (certPath, keyPath string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"TestOrg"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	certFile, err := os.CreateTemp("", "cert-*.pem")
	if err != nil {
		return "", "", err
	}
	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return "", "", err
	}

	keyFile, err := os.CreateTemp("", "key-*.pem")
	if err != nil {
		return "", "", err
	}
	defer keyFile.Close()
	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		return "", "", err
	}

	return certFile.Name(), keyFile.Name(), nil
}

func TestNewClient_Default(t *testing.T) {
	client, err := New("https://api.example.com")
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.Equal(t, "https://api.example.com", client.BaseURL)
}

func TestWithRetryPolicy(t *testing.T) {
	policy := RetryPolicy{
		Count:   3,
		Wait:    100 * time.Millisecond,
		MaxWait: 500 * time.Millisecond,
	}

	client, err := New("https://retry.test", WithRetryPolicy(policy))
	require.NoError(t, err)

	assert.Equal(t, 3, client.RetryCount)
	assert.Equal(t, 100*time.Millisecond, client.RetryWaitTime)
	assert.Equal(t, 500*time.Millisecond, client.RetryMaxWaitTime)
}

func TestWithToken(t *testing.T) {
	expectedToken := "sometoken123"

	// Start a test server that asserts Authorization header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+expectedToken, auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client, err := New(ts.URL, WithToken(expectedToken))
	require.NoError(t, err)

	resp, err := client.R().
		Get("/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestWithTLSCert_Generated(t *testing.T) {
	certPath, keyPath, err := generateTestTLSCert()
	require.NoError(t, err)
	defer os.Remove(certPath)
	defer os.Remove(keyPath)

	client, err := New("https://tls.test", WithTLSCert(TLSCert{
		CertFile: certPath,
		KeyFile:  keyPath,
	}))
	require.NoError(t, err)

	transport, ok := client.GetClient().Transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, transport.TLSClientConfig)
	assert.Len(t, transport.TLSClientConfig.Certificates, 1)
}

func TestWithTLSCert_FileNotFound(t *testing.T) {
	_, err := New("https://tls.test", WithTLSCert(TLSCert{
		CertFile: "nonexistent-cert.pem",
		KeyFile:  "nonexistent-key.pem",
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load TLS cert/key")
}
