package clients

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func generateSelfSignedCert(t *testing.T, dir string) (certPath, keyPath string) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		DNSNames: []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	certOut := filepath.Join(dir, "cert.pem")
	keyOut := filepath.Join(dir, "key.pem")

	err = os.WriteFile(certOut, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}), 0644)
	require.NoError(t, err)

	keyBytes := x509.MarshalPKCS1PrivateKey(priv)
	err = os.WriteFile(keyOut, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}), 0644)
	require.NoError(t, err)

	return certOut, keyOut
}

func TestNewGRPCClient(t *testing.T) {
	t.Run("With keepalive params", func(t *testing.T) {
		conn, err := NewGRPCClient("localhost:50051",
			WithGRPCKeepaliveParams(keepalive.ClientParameters{
				Time:                20 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			}),
		)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		_ = conn.Close()
	})

	t.Run("With TLS client cert", func(t *testing.T) {
		dir := t.TempDir()
		certPath, keyPath := generateSelfSignedCert(t, dir)

		conn, err := NewGRPCClient("localhost:50051",
			WithGRPCTLSClientCert(certPath, keyPath),
		)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		_ = conn.Close()
	})

	t.Run("Invalid TLS cert path", func(t *testing.T) {
		_, err := NewGRPCClient("localhost:50051",
			WithGRPCTLSClientCert("bad_cert_path.pem", "bad_key_path.pem"),
		)
		assert.Error(t, err)
	})

	t.Run("With custom transport credentials", func(t *testing.T) {
		creds := credentials.NewTLS(nil) // nil config = default TLS config
		conn, err := NewGRPCClient("localhost:50051",
			WithGRPCTransportCredentials(creds),
		)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		_ = conn.Close()
	})
}
