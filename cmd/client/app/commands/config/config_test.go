package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
)

// GenerateSelfSignedCert creates temp cert and key files and returns their paths.
func generateTempCertFiles(t *testing.T) (certPath, keyPath string) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})

	tmpCertFile, err := ioutil.TempFile("", "cert-*.pem")
	require.NoError(t, err)
	_, err = tmpCertFile.Write(certPEM)
	require.NoError(t, err)
	err = tmpCertFile.Close()
	require.NoError(t, err)

	tmpKeyFile, err := ioutil.TempFile("", "key-*.pem")
	require.NoError(t, err)
	_, err = tmpKeyFile.Write(keyPEM)
	require.NoError(t, err)
	err = tmpKeyFile.Close()
	require.NoError(t, err)

	return tmpCertFile.Name(), tmpKeyFile.Name()
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name           string
		authURL        string
		tlsCert        string
		tlsKey         string
		expectedScheme string
		expectError    bool
	}{
		{
			name:           "HTTP scheme no TLS",
			authURL:        "http://example.com",
			tlsCert:        "",
			tlsKey:         "",
			expectedScheme: scheme.HTTP,
			expectError:    false,
		},
		{
			name:           "HTTPS scheme with TLS",
			authURL:        "https://example.com",
			expectedScheme: scheme.HTTPS,
			expectError:    false,
		},
		{
			name:           "gRPC scheme with TLS",
			authURL:        "grpc://example.com",
			expectedScheme: scheme.GRPC,
			expectError:    false,
		},
		{
			name:        "Unsupported scheme returns error",
			authURL:     "ftp://example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Remove client.db after test, ignore error if file doesn't exist
			defer func() {
				_ = os.Remove("client.db")
			}()

			// Generate TLS cert/key files if needed
			if !tt.expectError && (tt.expectedScheme == scheme.HTTPS || tt.expectedScheme == scheme.GRPC) {
				certFile, keyFile := generateTempCertFiles(t)
				defer os.Remove(certFile)
				defer os.Remove(keyFile)
				tt.tlsCert = certFile
				tt.tlsKey = keyFile
			}

			cfg, err := NewConfig(tt.authURL, tt.tlsCert, tt.tlsKey)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, cfg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)

			// Adjust this check depending on how your config struct exposes clients
			switch tt.expectedScheme {
			case scheme.HTTP, scheme.HTTPS:
				require.NotNil(t, cfg.HTTPClient)
				require.Nil(t, cfg.GRPCClient)
			case scheme.GRPC:
				require.NotNil(t, cfg.GRPCClient)
				require.Nil(t, cfg.HTTPClient)
			}
		})
	}
}
