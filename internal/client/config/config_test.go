package config

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWithDB(t *testing.T) {
	cfg, err := NewConfig(WithDB())
	require.NoError(t, err)
	require.NotNil(t, cfg.DB)

	// Cleanup: remove the DB file after test finishes
	t.Cleanup(func() {
		cfg.DB.Close() // close DB before deleting file
		err := os.Remove("client.db")
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("failed to remove client.db: %v", err)
		}
	})

	// Simple ping test to ensure DB is alive.
	err = cfg.DB.Ping()
	assert.NoError(t, err)
}

// Test WithHTTPClient basic functionality
func TestWithHTTPClient_AuthorizationHeader(t *testing.T) {
	token := "mytoken"

	// Setup a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer "+token, auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg, err := NewConfig(WithHTTPClient(ts.URL, "", "", token))
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)

	// Make a GET request to the test server
	resp, err := cfg.HTTPClient.R().Get("/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

// Test WithHTTPClient with TLS certs (simulate loading certs)
func TestWithHTTPClient_TLS(t *testing.T) {
	// We'll create dummy cert and key files for test or skip this in tests since it's complex to mock TLS
	// For simplicity, we skip actual cert loading here or you can prepare test certs and pass paths

	// Just test that WithHTTPClient does not error when no cert paths given
	_, err := NewConfig(WithHTTPClient("https://example.com", "", "", ""))
	assert.NoError(t, err)
}

func TestWithGRPCClient_NoToken(t *testing.T) {
	addr, cleanup := startTestGRPCServer(t)
	defer cleanup()

	cfg, err := NewConfig(WithGRPCClient(addr, "", "", ""))
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)

	// You can optionally try a simple invoke or just check that connection is established.
}

// Start a minimal dummy gRPC server on a random port and return the address and a cleanup function
func startTestGRPCServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0") // Listen on random free port
	require.NoError(t, err)

	s := grpc.NewServer()
	// Optionally register your gRPC services here, or just run empty server.

	go func() {
		_ = s.Serve(lis) // run server in goroutine
	}()

	return lis.Addr().String(), func() {
		s.Stop()
		lis.Close()
	}
}

func TestWithGRPCClient_WithToken(t *testing.T) {
	addr, cleanup := startTestGRPCServer(t)
	defer cleanup()

	token := "testtoken"
	cfg, err := NewConfig(WithGRPCClient(addr, "", "", token))
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)

	// Now the connection is successful, you can try to invoke something or just ensure it's ready.
}

// Test retryInterceptor retry logic
func TestRetryInterceptor(t *testing.T) {
	count := 0
	maxRetries := 3
	interceptor := retryInterceptor(maxRetries, 10*time.Millisecond)

	err := interceptor(
		context.Background(),
		"/test",
		nil,
		nil,
		nil,
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			count++
			if count < maxRetries {
				// Return a grpc error with code Unavailable
				return status.Error(codes.Unavailable, "unavailable")
			}
			return nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, maxRetries, count)
}

// generateSelfSignedCert creates a self-signed TLS certificate and private key PEM bytes.
func generateSelfSignedCert() (certPEM []byte, keyPEM []byte, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // valid for 1 year

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return certPEM, keyPEM, nil
}

func writeTempFile(t *testing.T, content []byte, pattern string) string {
	f, err := os.CreateTemp("", pattern)
	require.NoError(t, err)
	_, err = f.Write(content)
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)
	return f.Name()
}

func TestWithHTTPClient_WithGeneratedCert(t *testing.T) {
	certPEM, keyPEM, err := generateSelfSignedCert()
	require.NoError(t, err)

	certFile := writeTempFile(t, certPEM, "cert.pem")
	keyFile := writeTempFile(t, keyPEM, "key.pem")
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	cfg, err := NewConfig(WithHTTPClient("https://example.com", certFile, keyFile, ""))
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
}

func TestWithGRPCClient_WithGeneratedCert(t *testing.T) {
	certPEM, keyPEM, err := generateSelfSignedCert()
	require.NoError(t, err)

	certFile := writeTempFile(t, certPEM, "cert.pem")
	keyFile := writeTempFile(t, keyPEM, "key.pem")
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	_, err = NewConfig(WithGRPCClient("localhost:0", certFile, keyFile, ""))
	require.Error(t, err) // expected to fail dial but cert loading is fine
}
