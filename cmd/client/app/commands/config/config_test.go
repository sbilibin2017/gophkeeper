package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateSelfSignedCert creates a temporary self-signed cert and key files and returns their paths.
func generateSelfSignedCert(t *testing.T) (certPath, keyPath string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	certFile, err := os.CreateTemp("", "cert-*.pem")
	require.NoError(t, err)
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	require.NoError(t, err)

	keyFile, err := os.CreateTemp("", "key-*.pem")
	require.NoError(t, err)
	defer keyFile.Close()

	privBytes, err := x509.MarshalECPrivateKey(priv)
	require.NoError(t, err)

	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	require.NoError(t, err)

	return certFile.Name(), keyFile.Name()
}

func TestNewClientConfig(t *testing.T) {
	certPath, keyPath := generateSelfSignedCert(t)
	defer os.Remove(certPath)
	defer os.Remove(keyPath)

	defer func() {
		if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
			t.Logf("failed to remove client.db: %v", err)
		}
	}()

	tests := []struct {
		name        string
		authURL     string
		tlsCert     string
		tlsKey      string
		expectError bool
		expectDB    bool
		expectHTTP  bool
		expectGRPC  bool
	}{
		{
			name:        "empty authURL returns only DB config",
			authURL:     "",
			expectError: false,
			expectDB:    true,
			expectHTTP:  false,
			expectGRPC:  false,
		},
		{
			name:        "http scheme no tls",
			authURL:     "http://example.com",
			expectError: false,
			expectDB:    true,
			expectHTTP:  true,
			expectGRPC:  false,
		},
		{
			name:        "https scheme with tls",
			authURL:     "https://example.com",
			tlsCert:     certPath,
			tlsKey:      keyPath,
			expectError: false,
			expectDB:    true,
			expectHTTP:  true,
			expectGRPC:  false,
		},
		{
			name:        "grpc scheme no tls",
			authURL:     "grpc://example.com",
			expectError: false,
			expectDB:    true,
			expectHTTP:  false,
			expectGRPC:  true,
		},
		{
			name:        "grpc scheme with tls",
			authURL:     "grpc://example.com",
			tlsCert:     certPath,
			tlsKey:      keyPath,
			expectError: false,
			expectDB:    true,
			expectHTTP:  false,
			expectGRPC:  true,
		},
		{
			name:        "unsupported scheme returns error",
			authURL:     "ftp://example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewClientConfig(tt.authURL, tt.tlsCert, tt.tlsKey)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)

			if tt.expectDB {
				assert.NotNil(t, cfg.DB, "expected DB connection")
			} else {
				assert.Nil(t, cfg.DB, "expected no DB connection")
			}

			if tt.expectHTTP {
				assert.NotNil(t, cfg.HTTPClient, "expected HTTP client")
			} else {
				assert.Nil(t, cfg.HTTPClient, "expected no HTTP client")
			}

			if tt.expectGRPC {
				assert.NotNil(t, cfg.GRPCClient, "expected gRPC client")
			} else {
				assert.Nil(t, cfg.GRPCClient, "expected no gRPC client")
			}
		})
	}
}
