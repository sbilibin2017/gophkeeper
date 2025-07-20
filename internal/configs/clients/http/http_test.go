package http

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func generateTestCertKeyFiles(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatalf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(time.Hour * 24),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certOut, err := os.CreateTemp("", "cert.pem")
	if err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.CreateTemp("", "key.pem")
	if err != nil {
		t.Fatalf("failed to create temp key file: %v", err)
	}
	defer keyOut.Close()

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("failed to marshal private key: %v", err)
	}

	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	return certOut.Name(), keyOut.Name()
}

func TestWithHTTPTLSClientCert(t *testing.T) {
	certFile, keyFile := generateTestCertKeyFiles(t)

	client := resty.New()
	opt := WithTLSClientCert(certFile, keyFile)
	err := opt(client)
	assert.NoError(t, err)

	transport, err := client.Transport()
	assert.NoError(t, err)
	assert.NotNil(t, transport.TLSClientConfig)
}

func TestWithHTTPRetryCount(t *testing.T) {
	client := resty.New()
	err := WithRetryCount(5)(client)
	assert.NoError(t, err)
	assert.Equal(t, 5, client.RetryCount)

	err = WithRetryCount(0)(client)
	assert.NoError(t, err)
	assert.Equal(t, 5, client.RetryCount) // unchanged
}

func TestWithHTTPRetryWaitTime(t *testing.T) {
	client := resty.New()
	dur := 123 * time.Millisecond
	err := WithRetryWaitTime(dur)(client)
	assert.NoError(t, err)
	assert.Equal(t, dur, client.RetryWaitTime)

	err = WithRetryWaitTime(0)(client)
	assert.NoError(t, err)
	assert.Equal(t, dur, client.RetryWaitTime) // unchanged
}

func TestWithHTTPRetryMaxWaitTime(t *testing.T) {
	client := resty.New()
	dur := 3 * time.Second
	err := WithRetryMaxWaitTime(dur)(client)
	assert.NoError(t, err)
	assert.Equal(t, dur, client.RetryMaxWaitTime)

	err = WithRetryMaxWaitTime(0)(client)
	assert.NoError(t, err)
	assert.Equal(t, dur, client.RetryMaxWaitTime) // unchanged
}

func TestWithHTTPToken(t *testing.T) {
	token := "mytoken"

	// Start a local test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+token, auth)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	client, err := New(ts.URL, WithToken(token))
	assert.NoError(t, err)

	resp, err := client.R().Get("/")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())
}

func TestWithHTTPTLSClientCert_Errors(t *testing.T) {
	// Case 1: empty certFile
	client := resty.New()
	err := WithTLSClientCert("", "key.pem")(client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")

	// Case 2: empty keyFile
	client = resty.New()
	err = WithTLSClientCert("cert.pem", "")(client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")

	// Case 3: invalid cert/key files (non-existent files)
	client = resty.New()
	err = WithTLSClientCert("nonexistent-cert.pem", "nonexistent-key.pem")(client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client certificate/key")
}

func TestNew_ReturnsErrorFromOption(t *testing.T) {
	errOption := func(client *resty.Client) error {
		return fmt.Errorf("option error")
	}

	client, err := New("http://example.com", errOption)
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Equal(t, "option error", err.Error())
}
