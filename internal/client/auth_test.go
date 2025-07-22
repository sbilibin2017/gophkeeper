package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// --- gRPC server implementation for testing ---

type authServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *authServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "token_for_" + req.Username}, nil
}

func (s *authServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "token_for_" + req.Username}, nil
}

func (s *authServer) Logout(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// --- helper to start HTTP server for testing ---

func startTestHTTPServer(t *testing.T) (url string, shutdown func()) {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var req models.AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		resp := models.AuthResponse{Token: "token_for_" + req.Username}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var req models.AuthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		resp := models.AuthResponse{Token: "token_for_" + req.Username}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &http.Server{Handler: mux}
	go server.Serve(ln)

	return "http://" + ln.Addr().String(), func() {
		server.Close()
		ln.Close()
	}
}

// --- Tests ---

func TestRegisterHTTP(t *testing.T) {
	url, shutdown := startTestHTTPServer(t)
	defer shutdown()

	client := resty.New().SetHostURL(url)

	resp, err := RegisterHTTP(context.Background(), client, &models.AuthRequest{
		Username: "user1",
		Password: "pass1",
	})
	require.NoError(t, err)
	assert.Equal(t, "token_for_user1", resp.Token)
}

func TestLoginHTTP(t *testing.T) {
	url, shutdown := startTestHTTPServer(t)
	defer shutdown()

	client := resty.New().SetHostURL(url)

	resp, err := LoginHTTP(context.Background(), client, &models.AuthRequest{
		Username: "user2",
		Password: "pass2",
	})
	require.NoError(t, err)
	assert.Equal(t, "token_for_user2", resp.Token)
}

func TestLogoutHTTP(t *testing.T) {
	url, shutdown := startTestHTTPServer(t)
	defer shutdown()

	client := resty.New().SetHostURL(url)

	err := LogoutHTTP(context.Background(), client)
	require.NoError(t, err)
}

func TestRegisterGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &authServer{})
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := pb.NewAuthServiceClient(conn)

	resp, err := RegisterGRPC(context.Background(), grpcClient, &models.AuthRequest{
		Username: "user1",
		Password: "pass1",
	})
	require.NoError(t, err)
	assert.Equal(t, "token_for_user1", resp.Token)
}

func TestLoginGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &authServer{})
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := pb.NewAuthServiceClient(conn)

	resp, err := LoginGRPC(context.Background(), grpcClient, &models.AuthRequest{
		Username: "user2",
		Password: "pass2",
	})
	require.NoError(t, err)
	assert.Equal(t, "token_for_user2", resp.Token)
}

func TestLogoutGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &authServer{})
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := pb.NewAuthServiceClient(conn)

	err = LogoutGRPC(context.Background(), grpcClient)
	require.NoError(t, err)
}
