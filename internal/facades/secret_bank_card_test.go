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

func TestSecretBankCardListHTTPFacade_List(t *testing.T) {
	// Мок HTTP сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/secret-bank-card" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		secrets := []models.SecretBankCardClient{
			{
				SecretName: "card1",
				Owner:      "John Doe",
				Number:     "1234567890123456",
				Exp:        "12/25",
				CVV:        "123",
				Meta:       nil,
				UpdatedAt:  time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
	facade := NewBankCardListHTTPFacade(client)

	secrets, err := facade.List(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "card1", secrets[0].SecretName)
	assert.Equal(t, "John Doe", secrets[0].Owner)
}

// --- gRPC мок сервер ---

type mockBankCardServiceServer struct {
	pb.UnimplementedSecretBankCardServiceServer
}

func (s *mockBankCardServiceServer) ListBankCards(ctx context.Context, req *pb.SecretBankCardListRequest) (*pb.SecretBankCardListResponse, error) {
	if req.Token != "test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretBankCardListResponse{
		Items: []*pb.SecretBankCard{
			{
				SecretName: "card1",
				Owner:      "John Doe",
				Number:     "1234567890123456",
				Exp:        "12/25",
				Cvv:        "123",
				Meta:       "",
				UpdatedAt:  time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestSecretBankCardListGRPCFacade_List(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretBankCardServiceServer(server, &mockBankCardServiceServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretBankCardServiceClient(conn)
	facade := NewBankCardListGRPCFacade(client)

	secrets, err := facade.List(ctx, "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "card1", secrets[0].SecretName)
	assert.Equal(t, "John Doe", secrets[0].Owner)
}

// --- gRPC мок сервер с ошибочным UpdatedAt ---

type badServer struct {
	pb.UnimplementedSecretBankCardServiceServer
}

func (s *badServer) ListBankCards(ctx context.Context, req *pb.SecretBankCardListRequest) (*pb.SecretBankCardListResponse, error) {
	return &pb.SecretBankCardListResponse{
		Items: []*pb.SecretBankCard{
			{
				SecretName: "card1",
				Owner:      "John Doe",
				Number:     "1234567890123456",
				Exp:        "12/25",
				Cvv:        "123",
				Meta:       "",
				UpdatedAt:  "bad-format", // Некорректный формат времени
			},
		},
	}, nil
}

func TestSecretBankCardListGRPCFacade_List_InvalidUpdatedAt(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretBankCardServiceServer(server, &badServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretBankCardServiceClient(conn)
	facade := NewBankCardListGRPCFacade(client)

	_, err = facade.List(ctx, "test-token")
	assert.Error(t, err)
	assert.Equal(t, "invalid updated_at format in response", err.Error())
}
