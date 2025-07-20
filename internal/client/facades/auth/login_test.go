package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/client/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// --- HTTP Test Server for Login with error case ---

func startLoginHTTPTestServer(t *testing.T) (string, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var req models.AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		if req.Username == "fail" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		resp := models.AuthResponse{Token: "http-login-token-for-" + req.Username}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &http.Server{Handler: mux}

	go func() {
		_ = srv.Serve(ln)
	}()

	return "http://" + ln.Addr().String(), func() {
		_ = srv.Shutdown(context.Background())
	}
}

// --- gRPC Test Server for Login with error case ---

type loginAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *loginAuthServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	if req.Username == "fail" {
		return nil, status.Error(401, "unauthorized")
	}
	return &pb.AuthResponse{Token: "grpc-login-token-for-" + req.Username}, nil
}

func startLoginGRPCTestServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, &loginAuthServer{})

	go func() {
		_ = grpcSrv.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		grpcSrv.GracefulStop()
	}
}

// --- Tests ---

func TestLoginHTTPFacade_Login(t *testing.T) {
	url, cleanup := startLoginHTTPTestServer(t)
	defer cleanup()

	client := resty.New()
	client.SetBaseURL(url)

	facade := NewLoginHTTPFacade(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Success case
	req := &models.AuthRequest{
		Username: "alice",
		Password: "password123",
	}

	resp, err := facade.Login(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "http-login-token-for-alice", resp.Token)

	// Failure case: HTTP 401 Unauthorized
	reqFail := &models.AuthRequest{
		Username: "fail",
		Password: "nopass",
	}

	resp, err = facade.Login(ctx, reqFail)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "login failed with status")
}

func TestLoginGRPCFacade_Login(t *testing.T) {
	addr, cleanup := startLoginGRPCTestServer(t)
	defer cleanup()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	facade := NewLoginGRPCFacade(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Success case
	req := &models.AuthRequest{
		Username: "bob",
		Password: "secretpass",
	}

	resp, err := facade.Login(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "grpc-login-token-for-bob", resp.Token)

	// Failure case: gRPC unauthorized error
	reqFail := &models.AuthRequest{
		Username: "fail",
		Password: "nopass",
	}

	resp, err = facade.Login(ctx, reqFail)
	require.Error(t, err)
	require.Nil(t, resp)

	// Confirm error is a gRPC status error
	var grpcErr interface{ GRPCStatus() *status.Status }
	require.True(t, errors.As(err, &grpcErr))
}
