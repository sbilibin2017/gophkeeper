package facades

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTP тест ---

func TestSecretTextListFacade_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/secret-text" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		secrets := []models.SecretTextClient{
			{
				SecretName: "text1",
				Content:    "Hello, world!",
				Meta:       nil,
				UpdatedAt:  time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
	facade := NewTextListFacade(client)

	secrets, err := facade.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "text1", secrets[0].SecretName)
	assert.Equal(t, "Hello, world!", secrets[0].Content)
}

// --- gRPC мок сервер ---

type mockTextServiceServer struct {
	pb.UnimplementedSecretTextServiceServer
}

func (s *mockTextServiceServer) ListTextSecrets(ctx context.Context, req *pb.SecretTextListRequest) (*pb.SecretTextListResponse, error) {
	if req.Token != "test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretTextListResponse{
		Items: []*pb.SecretText{
			{
				SecretName: "text1",
				Content:    "Hello, world!",
				Meta:       "",
				UpdatedAt:  time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func TestSecretTextListGRPCFacade_List(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterSecretTextServiceServer(server, &mockTextServiceServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretTextServiceClient(conn)
	facade := NewTextListGRPCFacade(client)

	secrets, err := facade.List(ctx, "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "text1", secrets[0].SecretName)
	assert.Equal(t, "Hello, world!", secrets[0].Content)
}

// --- gRPC мок сервер с ошибочным UpdatedAt ---

type badTextServer struct {
	pb.UnimplementedSecretTextServiceServer
}

func (s *badTextServer) ListTextSecrets(ctx context.Context, req *pb.SecretTextListRequest) (*pb.SecretTextListResponse, error) {
	return &pb.SecretTextListResponse{
		Items: []*pb.SecretText{
			{
				SecretName: "text1",
				Content:    "Hello, world!",
				Meta:       "",
				UpdatedAt:  "bad-format",
			},
		},
	}, nil
}

func TestSecretTextListGRPCFacade_List_InvalidUpdatedAt(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterSecretTextServiceServer(server, &badTextServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretTextServiceClient(conn)
	facade := NewTextListGRPCFacade(client)

	_, err = facade.List(ctx, "test-token")
	assert.Error(t, err)
	assert.Equal(t, "invalid updated_at format in response", err.Error())
}
