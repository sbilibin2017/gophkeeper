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

func TestRegisterHTTP_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/register", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "application/json") // ✅ ВАЖНО
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"token": "mocked-token",
		})
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	token, err := RegisterHTTP(context.Background(), client, "user", "pass")
	assert.NoError(t, err)
	assert.Equal(t, "mocked-token", token)
}

func TestRegisterHTTP_ServerErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	token, err := RegisterHTTP(context.Background(), client, "user", "pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registration failed")
	assert.Empty(t, token)
}

func TestRegisterHTTP_NetworkError(t *testing.T) {
	// несуществующий порт, чтобы симулировать сетевую ошибку
	client := resty.New().SetBaseURL("http://127.0.0.1:0")

	token, err := RegisterHTTP(context.Background(), client, "user", "pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send register request")
	assert.Empty(t, token)
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// Тестовая реализация gRPC сервера
type testRegisterServer struct {
	pb.UnimplementedRegisterServiceServer
}

func (s *testRegisterServer) Register(ctx context.Context, creds *pb.Credentials) (*pb.RegisterResponse, error) {
	switch creds.Username {
	case "fail":
		return &pb.RegisterResponse{
			Error: "username not allowed",
		}, nil
	case "grpc_error":
		return nil, fmt.Errorf("grpc transport error")
	default:
		return &pb.RegisterResponse{
			Token: "valid-token",
		}, nil
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func startTestGRPCServer(t *testing.T) *grpc.ClientConn {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterRegisterServiceServer(s, &testRegisterServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	return conn
}

func TestRegisterGRPC_Success(t *testing.T) {
	conn := startTestGRPCServer(t)
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)
	token, err := RegisterGRPC(context.Background(), client, "testuser", "testpass")

	assert.NoError(t, err)
	assert.Equal(t, "valid-token", token)
}

func TestRegisterGRPC_Error(t *testing.T) {
	conn := startTestGRPCServer(t)
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)
	token, err := RegisterGRPC(context.Background(), client, "fail", "testpass")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "username not allowed")
}

func TestRegisterGRPC_RPCError(t *testing.T) {
	conn := startTestGRPCServer(t)
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)
	token, err := RegisterGRPC(context.Background(), client, "grpc_error", "testpass")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "grpc transport error")
}
