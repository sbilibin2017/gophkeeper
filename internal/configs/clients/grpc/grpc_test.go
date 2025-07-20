package grpc

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// generateTestCertKeyFiles creates temp cert and key files for TLS tests
func generateTestCertKeyFiles(t *testing.T) (certFile, keyFile string) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	assert.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	assert.NoError(t, err)

	certOut, err := os.CreateTemp("", "cert.pem")
	assert.NoError(t, err)
	defer certOut.Close()
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	assert.NoError(t, err)

	keyOut, err := os.CreateTemp("", "key.pem")
	assert.NoError(t, err)
	defer keyOut.Close()
	b, err := x509.MarshalECPrivateKey(priv)
	assert.NoError(t, err)
	err = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	assert.NoError(t, err)

	return certOut.Name(), keyOut.Name()
}

func TestWithKeepaliveParams(t *testing.T) {
	params := keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             2 * time.Second,
		PermitWithoutStream: true,
	}

	opt := WithKeepaliveParams(params)
	opts, err := opt(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, opts)
}

func TestWithTLSClientCert_ErrorCases(t *testing.T) {
	opt := WithTLSClientCert("", "key.pem")
	_, err := opt(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client cert/key")

	opt = WithTLSClientCert("cert.pem", "")
	_, err = opt(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client cert/key")

	opt = WithTLSClientCert("nonexistent-cert.pem", "nonexistent-key.pem")
	_, err = opt(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client cert/key")
}

func TestWithTLSClientCert_Success(t *testing.T) {
	certFile, keyFile := generateTestCertKeyFiles(t)
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	opt := WithTLSClientCert(certFile, keyFile)
	opts, err := opt(nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, opts)
}

type dummyCreds struct{}

func (d *dummyCreds) ClientHandshake(ctx context.Context, authority string, rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return rawConn, nil, nil
}

func (d *dummyCreds) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return rawConn, nil, nil
}

func (d *dummyCreds) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{}
}

func (d *dummyCreds) Clone() credentials.TransportCredentials {
	return d
}

func (d *dummyCreds) OverrideServerName(serverName string) error {
	return nil
}

func TestWithTransportCredentials(t *testing.T) {
	creds := &dummyCreds{}
	opt := WithTransportCredentials(creds)

	opts, err := opt(nil)
	assert.NoError(t, err)
	assert.Len(t, opts, 1)
	assert.NotNil(t, opts[0])
}
func startTestServer(t *testing.T) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", "localhost:0") // random free port
	assert.NoError(t, err)

	s := grpc.NewServer()
	go func() {
		_ = s.Serve(lis)
	}()
	return s, lis
}

func TestNew(t *testing.T) {
	srv, lis := startTestServer(t)
	defer srv.Stop()

	conn, err := New(lis.Addr().String(), func(opts []grpc.DialOption) ([]grpc.DialOption, error) {
		// Use insecure transport credentials for test server without TLS
		return append(opts, grpc.WithTransportCredentials(insecure.NewCredentials())), nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()
}

func TestNew_ReturnsErrorIfOptionFails(t *testing.T) {
	errOpt := func(opts []grpc.DialOption) ([]grpc.DialOption, error) {
		return nil, fmt.Errorf("option failure")
	}

	conn, err := New("dummy:1234", errOpt)
	assert.Nil(t, conn)
	assert.Error(t, err)
	assert.Equal(t, "option failure", err.Error())
}

func TestUnaryInterceptorWithToken_AddsAuthorizationMetadata(t *testing.T) {
	const token = "test-token"
	interceptor := WithUnaryInterceptorToken(token)

	fakeInvoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "Bearer "+token, md["authorization"][0])
		return nil
	}

	err := interceptor(context.Background(), "TestMethod", nil, nil, nil, fakeInvoker)
	assert.NoError(t, err)
}

func TestStreamInterceptorWithToken_AddsAuthorizationMetadata(t *testing.T) {
	const token = "test-token"
	interceptor := WithStreamInterceptorToken(token)

	fakeStreamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, "Bearer "+token, md["authorization"][0])
		return nil, nil
	}

	_, err := interceptor(context.Background(), nil, nil, "TestMethod", fakeStreamer)
	assert.NoError(t, err)
}
