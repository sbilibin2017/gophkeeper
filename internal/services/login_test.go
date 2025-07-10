package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestLoginHTTP_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/login", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"token": "mocked-token",
		})
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	token, err := LoginHTTP(context.Background(), client, "user", "pass")
	assert.NoError(t, err)
	assert.Equal(t, "mocked-token", token)
}

func TestLoginHTTP_ServerErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid login", http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	token, err := LoginHTTP(context.Background(), client, "user", "wrongpass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
	assert.Empty(t, token)
}

func TestLoginHTTP_NetworkError(t *testing.T) {
	client := resty.New().SetBaseURL("http://127.0.0.1:0") // несуществующий порт

	token, err := LoginHTTP(context.Background(), client, "user", "pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send login request")
	assert.Empty(t, token)
}

const loginBufSize = 1024 * 1024

var loginLis *bufconn.Listener

type testLoginServer struct {
	pb.UnimplementedLoginServiceServer
}

func (s *testLoginServer) Login(ctx context.Context, creds *pb.Credentials) (*pb.LoginResponse, error) {
	switch creds.Username {
	case "fail":
		return &pb.LoginResponse{
			Error: "invalid credentials",
		}, nil
	case "grpc_error":
		return nil, fmt.Errorf("grpc transport error")
	default:
		return &pb.LoginResponse{
			Token: "valid-token",
		}, nil
	}
}

func loginBufDialer(context.Context, string) (net.Conn, error) {
	return loginLis.Dial()
}

func startTestLoginGRPCServer(t *testing.T) *grpc.ClientConn {
	loginLis = bufconn.Listen(loginBufSize)
	s := grpc.NewServer()
	pb.RegisterLoginServiceServer(s, &testLoginServer{})

	go func() {
		if err := s.Serve(loginLis); err != nil {
			t.Fatalf("Login gRPC server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(loginBufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial login bufnet: %v", err)
	}

	return conn
}

func TestLoginGRPC_Success(t *testing.T) {
	conn := startTestLoginGRPCServer(t)
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)
	token, err := LoginGRPC(context.Background(), client, "testuser", "testpass")

	assert.NoError(t, err)
	assert.Equal(t, "valid-token", token)
}

func TestLoginGRPC_Error(t *testing.T) {
	conn := startTestLoginGRPCServer(t)
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)
	token, err := LoginGRPC(context.Background(), client, "fail", "wrongpass")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestLoginGRPC_RPCError(t *testing.T) {
	conn := startTestLoginGRPCServer(t)
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)
	token, err := LoginGRPC(context.Background(), client, "grpc_error", "pass")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "grpc transport error")
}
