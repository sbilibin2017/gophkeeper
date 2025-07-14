package facades

import (
	"context"
	"encoding/json"
	"errors"
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

func TestSecretBinaryListFacade_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/secret-binary" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		secrets := []models.SecretBinaryClient{
			{
				SecretName: "binary1",
				Data:       []byte{0x01, 0x02, 0x03},
				Meta:       nil,
				UpdatedAt:  time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
	facade := NewBinaryListFacade(client)

	secrets, err := facade.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "binary1", secrets[0].SecretName)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, secrets[0].Data)
}

// --- gRPC мок сервер ---

type mockBinaryServiceServer struct {
	pb.UnimplementedSecretBinaryServiceServer
}

func (s *mockBinaryServiceServer) ListBinarySecrets(ctx context.Context, req *pb.SecretBinaryListRequest) (*pb.SecretBinaryListResponse, error) {
	if req.Token != "test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretBinaryListResponse{
		Items: []*pb.SecretBinary{
			{
				SecretName: "binary1",
				Data:       []byte{0x01, 0x02, 0x03},
				Meta:       "",
				UpdatedAt:  time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func TestSecretBinaryListGRPCFacade_List(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretBinaryServiceServer(server, &mockBinaryServiceServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretBinaryServiceClient(conn)
	facade := NewBinaryListGRPCFacade(client)

	secrets, err := facade.List(ctx, "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "binary1", secrets[0].SecretName)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, secrets[0].Data)
}

// --- gRPC мок сервер с ошибочным UpdatedAt ---

type badBinaryServer struct {
	pb.UnimplementedSecretBinaryServiceServer
}

func (s *badBinaryServer) ListBinarySecrets(ctx context.Context, req *pb.SecretBinaryListRequest) (*pb.SecretBinaryListResponse, error) {
	return &pb.SecretBinaryListResponse{
		Items: []*pb.SecretBinary{
			{
				SecretName: "binary1",
				Data:       []byte{0x01, 0x02, 0x03},
				Meta:       "",
				UpdatedAt:  "bad-format",
			},
		},
	}, nil
}

func TestSecretBinaryListGRPCFacade_List_InvalidUpdatedAt(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretBinaryServiceServer(server, &badBinaryServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretBinaryServiceClient(conn)
	facade := NewBinaryListGRPCFacade(client)

	_, err = facade.List(ctx, "test-token")
	assert.Error(t, err)
	assert.Equal(t, "invalid updated_at format in response", err.Error())
}
