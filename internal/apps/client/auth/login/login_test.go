package login

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestRegisterCommand_RunE(t *testing.T) {
	runHTTPFunc := func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
		if username == "httpuser" && password == "httppass" {
			return &models.AuthResponse{Token: "http_token"}, nil
		}
		return nil, errors.New("http login failed")
	}

	runGRPCFunc := func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
		if username == "grpcuser" && password == "grpcpass" {
			return &models.AuthResponse{Token: "grpc_token"}, nil
		}
		return nil, errors.New("grpc login failed")
	}

	tests := []struct {
		name        string
		args        []string
		wantOutput  string
		wantErrPart string
	}{
		{
			name:       "Successful gRPC login",
			args:       []string{"login", "--username", "grpcuser", "--password", "grpcpass", "--auth-url", "grpc://localhost", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantOutput: "grpc_token\n",
		},
		{
			name:       "Successful HTTP login",
			args:       []string{"login", "--username", "httpuser", "--password", "httppass", "--auth-url", "https://example.com", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantOutput: "http_token\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{Use: "root"}
			RegisterCommand(root, runHTTPFunc, runGRPCFunc)

			var output bytes.Buffer
			root.SetOut(&output)
			root.SetErr(&output)
			root.SetArgs(tt.args)

			err := root.Execute()

			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				assert.Empty(t, output.String())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutput, output.String())
			}
		})
	}
}

// minimalGRPCAuthServer starts a gRPC test server with a simple AuthService implementation
func minimalGRPCAuthServer(t *testing.T, certFile, keyFile string) (*grpc.Server, string) {
	// Load TLS cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		t.Fatalf("failed to load key pair: %v", err)
	}
	creds := credentials.NewServerTLSFromCert(&cert)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(grpc.Creds(creds))

	// Implement the AuthService server
	pb.RegisterAuthServiceServer(server, &testAuthService{})

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	return server, lis.Addr().String()
}

type testAuthService struct {
	pb.UnimplementedAuthServiceServer
}

func (s *testAuthService) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	// Return a single token for consistency with HTTP test
	return &pb.AuthResponse{
		Token: "test_token",
	}, nil
}

func TestNewRunGRPC_Integration(t *testing.T) {
	// Remove old DB if exists
	if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove client.db: %v", err)
	}
	// Schedule deletion after test
	defer func() {
		if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to remove client.db: %v", err)
		}
	}()

	certFile, keyFile := generateSelfSignedCert(t)
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	server, addr := minimalGRPCAuthServer(t, certFile, keyFile)
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	authURL := fmt.Sprintf("grpc://%s", addr)
	run := NewRunGRPC(authURL, certFile, keyFile)

	resp, err := run(ctx, "dummyuser", "dummypass")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test_token", resp.Token)
}

// generateSelfSignedCert generates a self-signed TLS cert and key and writes to temp files
func generateSelfSignedCert(t *testing.T) (certFile, keyFile string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	certOut, err := os.CreateTemp("", "cert-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer certOut.Close()

	keyOut, err := os.CreateTemp("", "key-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer keyOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatal(err)
	}

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}); err != nil {
		t.Fatal(err)
	}

	return certOut.Name(), keyOut.Name()
}

// minimalAuthServer runs a HTTPS test server simulating the login endpoint,
// returns the server and its base URL.
func minimalAuthServer(t *testing.T, certFile, keyFile string) (*http.Server, string) {
	handler := http.NewServeMux()
	handler.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Changed response JSON to match test expectation: a single "token"
		_, _ = w.Write([]byte(`{"token":"test_token"}`))
	})

	srv := &http.Server{
		Handler: handler,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go func() {
		err := srv.ServeTLS(ln, certFile, keyFile)
		if err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()

	baseURL := fmt.Sprintf("https://%s", ln.Addr().String())
	return srv, baseURL
}

func TestNewRunHTTP_Integration(t *testing.T) {
	// Remove old DB if exists
	if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove client.db: %v", err)
	}

	// Schedule deletion of client.db after test finishes
	defer func() {
		if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
			t.Fatalf("failed to remove client.db: %v", err)
		}
	}()

	certFile, keyFile := generateSelfSignedCert(t)
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	srv, authURL := minimalAuthServer(t, certFile, keyFile)
	defer srv.Shutdown(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	run := NewRunHTTP(authURL, certFile, keyFile)

	resp, err := run(ctx, "dummyuser", "dummypass")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test_token", resp.Token)
}
