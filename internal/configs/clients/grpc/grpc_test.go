package grpc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func startBufServer(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	s := gogrpc.NewServer()
	go func() {
		require.NoError(t, s.Serve(lis))
	}()
}

// generateTestCertFiles creates a self-signed root CA cert for TLS testing
func generateTestCertFiles() (certFile, keyFile string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	certOut, err := os.CreateTemp("", "cert-*.pem")
	if err != nil {
		return "", "", err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyOut, err := os.CreateTemp("", "key-*.pem")
	if err != nil {
		return "", "", err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certOut.Name(), keyOut.Name(), nil
}

func TestWithToken(t *testing.T) {
	opt := WithAuthToken("abc123")
	dialOpt, err := opt()
	require.NoError(t, err)
	require.NotNil(t, dialOpt)
}

func TestWithToken_Empty(t *testing.T) {
	opt := WithAuthToken("")
	dialOpt, err := opt()
	require.NoError(t, err)
	assert.Nil(t, dialOpt)
}

func TestWithRetryPolicy(t *testing.T) {
	opt := WithRetryPolicy(RetryPolicy{
		Count:   4,
		Wait:    100 * time.Millisecond,
		MaxWait: 300 * time.Millisecond,
	})
	dialOpt, err := opt()
	require.NoError(t, err)
	require.NotNil(t, dialOpt)
}

func TestWithRetryPolicy_Empty(t *testing.T) {
	opt := WithRetryPolicy(RetryPolicy{})
	dialOpt, err := opt()
	require.NoError(t, err)
	assert.Nil(t, dialOpt)
}

func TestWithTLSCert_Success(t *testing.T) {
	certFile, _, err := generateTestCertFiles()
	require.NoError(t, err)
	defer os.Remove(certFile)

	opt := WithTLSCert(certFile)
	dialOpt, err := opt()
	require.NoError(t, err)
	require.NotNil(t, dialOpt)
}

func TestWithTLSCert_Failure(t *testing.T) {
	opt := WithTLSCert("invalid-cert.pem")
	dialOpt, err := opt()
	require.Error(t, err)
	assert.Nil(t, dialOpt)
}

func TestWithTLSCert_Empty(t *testing.T) {
	opt := WithTLSCert()
	dialOpt, err := opt()
	require.NoError(t, err)
	assert.Nil(t, dialOpt)
}

func TestNew_WithOptions(t *testing.T) {
	startBufServer(t)

	certFile, _, err := generateTestCertFiles()
	require.NoError(t, err)
	defer os.Remove(certFile)

	conn, err := New("bufnet",
		func() (gogrpc.DialOption, error) {
			return gogrpc.WithContextDialer(bufDialer), nil
		},
		WithTLSCert(certFile),
		WithAuthToken("abc123"),
		WithRetryPolicy(RetryPolicy{
			Count:   2,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, conn)
	conn.Close()
}

func TestNew_ErrorInOption(t *testing.T) {
	errOpt := func() (gogrpc.DialOption, error) {
		return nil, assert.AnError
	}
	conn, err := New("target", errOpt)
	require.Error(t, err)
	assert.Nil(t, conn)
}

func TestTokenAuth_GetRequestMetadata(t *testing.T) {
	token := "sometoken123"
	auth := tokenAuth{token: token}

	md, err := auth.GetRequestMetadata(context.Background(), "someURI")
	assert.NoError(t, err)
	assert.NotNil(t, md)
	assert.Equal(t, "Bearer "+token, md["authorization"])
}

func TestTokenAuth_RequireTransportSecurity(t *testing.T) {
	auth := tokenAuth{token: "anytoken"}

	requireTLS := auth.RequireTransportSecurity()
	assert.True(t, requireTLS)
}
