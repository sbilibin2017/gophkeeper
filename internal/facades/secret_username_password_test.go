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

func TestSecretUsernamePasswordListFacade_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/secret-username-password" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		secrets := []models.SecretUsernamePasswordClient{
			{
				SecretName: "login1",
				Username:   "user1",
				Password:   "pass1",
				Meta:       nil,
				UpdatedAt:  time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
	facade := NewUsernamePasswordListFacade(client)

	secrets, err := facade.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "login1", secrets[0].SecretName)
	assert.Equal(t, "user1", secrets[0].Username)
	assert.Equal(t, "pass1", secrets[0].Password)
}

// --- gRPC мок сервер ---

type mockUsernamePasswordServer struct {
	pb.UnimplementedSecretUsernamePasswordServiceServer
}

func (s *mockUsernamePasswordServer) ListUsernamePasswordSecrets(ctx context.Context, req *pb.SecretUsernamePasswordListRequest) (*pb.SecretUsernamePasswordListResponse, error) {
	if req.Token != "test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretUsernamePasswordListResponse{
		Items: []*pb.SecretUsernamePassword{
			{
				SecretName: "login1",
				Username:   "user1",
				Password:   "pass1",
				Meta:       "",
				UpdatedAt:  time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func TestSecretUsernamePasswordListGRPCFacade_List(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterSecretUsernamePasswordServiceServer(server, &mockUsernamePasswordServer{})

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

	client := pb.NewSecretUsernamePasswordServiceClient(conn)
	facade := NewUsernamePasswordListGRPCFacade(client)

	secrets, err := facade.List(ctx, "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "login1", secrets[0].SecretName)
	assert.Equal(t, "user1", secrets[0].Username)
	assert.Equal(t, "pass1", secrets[0].Password)
}

// --- gRPC мок сервер с ошибочным UpdatedAt ---

type badUsernamePasswordServer struct {
	pb.UnimplementedSecretUsernamePasswordServiceServer
}

func (s *badUsernamePasswordServer) ListUsernamePasswordSecrets(ctx context.Context, req *pb.SecretUsernamePasswordListRequest) (*pb.SecretUsernamePasswordListResponse, error) {
	return &pb.SecretUsernamePasswordListResponse{
		Items: []*pb.SecretUsernamePassword{
			{
				SecretName: "login1",
				Username:   "user1",
				Password:   "pass1",
				Meta:       "",
				UpdatedAt:  "bad-format",
			},
		},
	}, nil
}

func TestSecretUsernamePasswordListGRPCFacade_List_InvalidUpdatedAt(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterSecretUsernamePasswordServiceServer(server, &badUsernamePasswordServer{})

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

	client := pb.NewSecretUsernamePasswordServiceClient(conn)
	facade := NewUsernamePasswordListGRPCFacade(client)

	_, err = facade.List(ctx, "test-token")
	assert.Error(t, err)
	assert.Equal(t, "invalid updated_at format in response", err.Error())
}
