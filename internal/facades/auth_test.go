package facades

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Minimal HTTP handler example
func startTestHTTPServer(t *testing.T) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token":"http_test_token"}`))
	})
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token":"http_test_token"}`))
	})
	mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("HTTP server error: %v", err)
		}
	}()

	// wait briefly for server to start
	time.Sleep(100 * time.Millisecond)

	return srv
}

// Minimal gRPC AuthService implementation
type authServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *authServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc_test_token"}, nil
}

func (s *authServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc_test_token"}, nil
}

func (s *authServer) Logout(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func startTestGRPCServer(t *testing.T) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", ":9090")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &authServer{})

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			t.Errorf("gRPC server error: %v", err)
		}
	}()

	// wait briefly for server to start
	time.Sleep(100 * time.Millisecond)

	return grpcServer, lis
}

func TestWithEmbeddedServers(t *testing.T) {
	httpServer := startTestHTTPServer(t)
	defer func() {
		err := httpServer.Shutdown(context.Background())
		require.NoError(t, err)
	}()

	grpcServer, lis := startTestGRPCServer(t)
	defer func() {
		grpcServer.Stop()
		lis.Close()
	}()

	// HTTP facade setup
	httpFacade := &AuthHTTPFacade{client: resty.New().SetBaseURL("http://localhost:8080")}

	// gRPC facade setup
	grpcConn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer grpcConn.Close()
	grpcFacade := NewAuthGRPCFacade(grpcConn)

	// HTTP Register test
	resp, err := httpFacade.Register(context.Background(), &models.AuthRequest{Login: "user", Password: "pass"})
	require.NoError(t, err)
	require.Equal(t, "http_test_token", resp.Token)

	// HTTP Login test
	resp, err = httpFacade.Login(context.Background(), &models.AuthRequest{Login: "user", Password: "pass"})
	require.NoError(t, err)
	require.Equal(t, "http_test_token", resp.Token)

	// HTTP Logout test
	err = httpFacade.Logout(context.Background())
	require.NoError(t, err)

	// gRPC Register test
	grpcResp, err := grpcFacade.Register(context.Background(), &models.AuthRequest{Login: "user", Password: "pass"})
	require.NoError(t, err)
	require.Equal(t, "grpc_test_token", grpcResp.Token)

	// gRPC Login test
	grpcResp, err = grpcFacade.Login(context.Background(), &models.AuthRequest{Login: "user", Password: "pass"})
	require.NoError(t, err)
	require.Equal(t, "grpc_test_token", grpcResp.Token)

	// gRPC Logout test
	err = grpcFacade.Logout(context.Background())
	require.NoError(t, err)
}

// -- HTTP Error Tests --

func TestAuthHTTPFacade_Register_Error(t *testing.T) {
	// Use a client pointing to an invalid URL to force connection error
	f := &AuthHTTPFacade{client: resty.New().SetBaseURL("http://127.0.0.1:0")}
	_, err := f.Register(context.Background(), &models.AuthRequest{Login: "x", Password: "x"})
	require.Error(t, err)
}

func TestAuthHTTPFacade_Login_Error(t *testing.T) {
	f := &AuthHTTPFacade{client: resty.New().SetBaseURL("http://127.0.0.1:0")}
	_, err := f.Login(context.Background(), &models.AuthRequest{Login: "x", Password: "x"})
	require.Error(t, err)
}

func TestAuthHTTPFacade_Logout_Error(t *testing.T) {
	f := &AuthHTTPFacade{client: resty.New().SetBaseURL("http://127.0.0.1:0")}
	err := f.Logout(context.Background())
	require.Error(t, err)
}

func TestAuthHTTPFacade_Register_HTTPErrorStatus(t *testing.T) {
	// Setup server that returns error status
	srv := &http.Server{
		Addr: ":8081",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "bad request", http.StatusBadRequest)
		}),
	}
	go srv.ListenAndServe()
	defer srv.Close()
	time.Sleep(50 * time.Millisecond)

	f := &AuthHTTPFacade{client: resty.New().SetBaseURL("http://localhost:8081")}
	_, err := f.Register(context.Background(), &models.AuthRequest{Login: "x", Password: "x"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to register")
}

func TestAuthHTTPFacade_Login_HTTPErrorStatus(t *testing.T) {
	srv := &http.Server{
		Addr: ":8082",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}),
	}
	go srv.ListenAndServe()
	defer srv.Close()
	time.Sleep(50 * time.Millisecond)

	f := &AuthHTTPFacade{client: resty.New().SetBaseURL("http://localhost:8082")}
	_, err := f.Login(context.Background(), &models.AuthRequest{Login: "x", Password: "x"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to login")
}

func TestAuthHTTPFacade_Logout_HTTPErrorStatus(t *testing.T) {
	srv := &http.Server{
		Addr: ":8083",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}),
	}
	go srv.ListenAndServe()
	defer srv.Close()
	time.Sleep(50 * time.Millisecond)

	f := &AuthHTTPFacade{client: resty.New().SetBaseURL("http://localhost:8083")}
	err := f.Logout(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to logout")
}

// -- gRPC Error Tests --

type authServerError struct {
	pb.UnimplementedAuthServiceServer
}

func (s *authServerError) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return nil, status.Error(codes.Internal, "internal error")
}

func (s *authServerError) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return nil, status.Error(codes.Unauthenticated, "unauthenticated")
}

func (s *authServerError) Logout(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unknown, "unknown error")
}

func startTestGRPCServerError(t *testing.T) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", ":9091")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &authServerError{})

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			t.Errorf("gRPC server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return grpcServer, lis
}

func TestAuthGRPCFacade_Register_Error(t *testing.T) {
	grpcServer, lis := startTestGRPCServerError(t)
	defer func() {
		grpcServer.Stop()
		lis.Close()
	}()

	conn, err := grpc.Dial("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	f := NewAuthGRPCFacade(conn)
	_, err = f.Register(context.Background(), &models.AuthRequest{Login: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "grpc Register failed")
}

func TestAuthGRPCFacade_Login_Error(t *testing.T) {
	grpcServer, lis := startTestGRPCServerError(t)
	defer func() {
		grpcServer.Stop()
		lis.Close()
	}()

	conn, err := grpc.Dial("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	f := NewAuthGRPCFacade(conn)
	_, err = f.Login(context.Background(), &models.AuthRequest{Login: "user", Password: "pass"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "grpc Login failed")
}

func TestAuthGRPCFacade_Logout_Error(t *testing.T) {
	grpcServer, lis := startTestGRPCServerError(t)
	defer func() {
		grpcServer.Stop()
		lis.Close()
	}()

	conn, err := grpc.Dial("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	f := NewAuthGRPCFacade(conn)
	err = f.Logout(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "grpc Logout failed")
}
