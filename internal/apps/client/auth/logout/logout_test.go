package logout

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

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// createClientTables creates all required client tables in DB before tests run
func createClientTables(ctx context.Context, dbConn *sqlx.DB) error {
	if err := bankcard.CreateClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := text.CreateClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := binary.CreateClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := user.CreateClientTable(ctx, dbConn); err != nil {
		return err
	}
	return nil
}

func TestRegisterCommand_RunE(t *testing.T) {
	runLogoutHTTPFunc := func(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, token string) error {
		if token == "valid_http_token" {
			return nil
		}
		return errors.New("http logout failed")
	}

	runLogoutGRPCFunc := func(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, token string) error {
		if token == "valid_grpc_token" {
			return nil
		}
		return errors.New("grpc logout failed")
	}

	tests := []struct {
		name        string
		args        []string
		wantOutput  string
		wantErrPart string
	}{
		{
			name:       "Successful gRPC logout",
			args:       []string{"logout", "--auth-url", "grpc://localhost", "--token", "valid_grpc_token", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantOutput: "Logout successful.\n",
		},
		{
			name:       "Successful HTTP logout",
			args:       []string{"logout", "--auth-url", "https://example.com", "--token", "valid_http_token", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantOutput: "Logout successful.\n",
		},
		{
			name:        "Unsupported auth URL scheme",
			args:        []string{"logout", "--auth-url", "ftp://example.com", "--token", "token", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantErrPart: "unsupported auth URL scheme",
		},
		{
			name:        "Logout failure HTTP",
			args:        []string{"logout", "--auth-url", "https://example.com", "--token", "bad_token", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantErrPart: "logout failed: http logout failed",
		},
		{
			name:        "Logout failure gRPC",
			args:        []string{"logout", "--auth-url", "grpc://localhost", "--token", "bad_token", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantErrPart: "logout failed: grpc logout failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{Use: "root"}
			RegisterLogoutCommand(root, runLogoutHTTPFunc, runLogoutGRPCFunc)

			var output bytes.Buffer
			root.SetOut(&output)
			root.SetErr(&output)
			root.SetArgs(tt.args)

			err := root.Execute()

			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				assert.NotEmpty(t, output.String())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutput, output.String())
			}
		})
	}
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

// minimalGRPCAuthServer starts a gRPC test server with a simple AuthService implementation
func minimalGRPCAuthServer(t *testing.T, certFile, keyFile string) (*grpc.Server, string) {
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
	pb.RegisterAuthServiceServer(server, &testAuthService{})

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	return server, "grpc://" + lis.Addr().String()
}

type testAuthService struct {
	pb.UnimplementedAuthServiceServer
}

func (s *testAuthService) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func TestRunGRPC_LogoutIntegration(t *testing.T) {
	if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove client.db: %v", err)
	}
	defer os.Remove("client.db")

	certFile, keyFile := generateSelfSignedCert(t)
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	server, authURL := minimalGRPCAuthServer(t, certFile, keyFile)
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	defer dbConn.Close()

	if err := createClientTables(ctx, dbConn); err != nil {
		t.Fatalf("failed to create client tables: %v", err)
	}

	err = RunLogoutGRPC(ctx, authURL, certFile, keyFile, "dummy_token")
	assert.NoError(t, err)
}

// minimalLogoutServer starts a HTTPS server with a /logout endpoint that returns 200 OK
func minimalLogoutServer(t *testing.T, certFile, keyFile string) (*http.Server, string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Handler: mux,
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

func TestRunHTTP_LogoutIntegration(t *testing.T) {
	if err := os.Remove("client.db"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove client.db: %v", err)
	}
	defer os.Remove("client.db")

	certFile, keyFile := generateSelfSignedCert(t)
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	srv, authURL := minimalLogoutServer(t, certFile, keyFile)
	defer srv.Shutdown(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	defer dbConn.Close()

	if err := createClientTables(ctx, dbConn); err != nil {
		t.Fatalf("failed to create client tables: %v", err)
	}

	err = RunLogoutHTTP(ctx, authURL, certFile, keyFile, "dummy_token")
	assert.NoError(t, err)
}
