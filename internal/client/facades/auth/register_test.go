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

// --- HTTP Test Server ---

func startRegisterHTTPTestServer(t *testing.T) (string, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var req models.AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		if req.Username == "fail" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		resp := models.AuthResponse{Token: "http-token-for-" + req.Username}
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

// --- gRPC Test Server ---

type registerAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *registerAuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	if req.Username == "fail" {
		return nil, status.Error(400, "invalid user")
	}
	return &pb.AuthResponse{Token: "grpc-token-for-" + req.Username}, nil
}

func startRegisterGRPCTestServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, &registerAuthServer{})

	go func() {
		_ = grpcSrv.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		grpcSrv.GracefulStop()
	}
}

// --- Tests ---

func TestRegisterHTTPFacade_Register(t *testing.T) {
	url, cleanup := startRegisterHTTPTestServer(t)
	defer cleanup()

	client := resty.New()
	client.SetBaseURL(url)

	facade := NewRegisterHTTPFacade(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Success case
	req := &models.AuthRequest{
		Username: "alice",
		Password: "password123",
	}

	resp, err := facade.Register(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "http-token-for-alice", resp.Token)

	// Failure case: server returns 400
	reqFail := &models.AuthRequest{
		Username: "fail",
		Password: "nopass",
	}
	resp, err = facade.Register(ctx, reqFail)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "registration failed with status")
}

func TestRegisterGRPCFacade_Register(t *testing.T) {
	addr, cleanup := startRegisterGRPCTestServer(t)
	defer cleanup()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	facade := NewRegisterGRPCFacade(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Success case
	req := &models.AuthRequest{
		Username: "bob",
		Password: "secretpass",
	}

	resp, err := facade.Register(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "grpc-token-for-bob", resp.Token)

	// Failure case: grpc returns error
	reqFail := &models.AuthRequest{
		Username: "fail",
		Password: "nopass",
	}
	resp, err = facade.Register(ctx, reqFail)
	require.Error(t, err)
	require.Nil(t, resp)
	var grpcErr interface{ GRPCStatus() *status.Status }
	require.True(t, errors.As(err, &grpcErr))
}
