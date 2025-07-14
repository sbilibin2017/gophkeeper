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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// --- HTTP tests ---

func TestSecretBankCardListHTTPFacade_List(t *testing.T) {
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
	facade := NewSecretBankCardListHTTPFacade(client)

	secrets, err := facade.List(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "card1", secrets[0].SecretName)
	assert.Equal(t, "John Doe", secrets[0].Owner)
}

func TestSecretBankCardGetHTTPFacade_Get(t *testing.T) {
	expected := models.SecretBankCardClient{
		SecretName: "card1",
		Owner:      "John Doe",
		Number:     "1234567890123456",
		Exp:        "12/25",
		CVV:        "123",
		UpdatedAt:  time.Now(),
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/get/secret-bank-card/card1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretBankCardGetHTTPFacade(client)

	secret, err := facade.Get(context.Background(), "token", "card1")
	assert.NoError(t, err)
	assert.Equal(t, expected.SecretName, secret.SecretName)
}

func TestSecretBankCardSaveHTTPFacade_Save(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/save/secret-bank-card", r.URL.Path)
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer token", auth)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretBankCardSaveHTTPFacade(client)

	secret := models.SecretBankCardClient{
		SecretName: "card1",
		Owner:      "John Doe",
		Number:     "1234567890123456",
		Exp:        "12/25",
		CVV:        "123",
		UpdatedAt:  time.Now(),
	}

	err := facade.Save(context.Background(), "token", secret)
	assert.NoError(t, err)
}

// --- gRPC mocks and tests ---

type mockBankCardServer struct {
	pb.UnimplementedSecretBankCardServiceServer
}

func (s *mockBankCardServer) List(ctx context.Context, req *pb.SecretBankCardListRequest) (*pb.SecretBankCardListResponse, error) {
	// Токен в Metadata, так что req.Token пустой — проверяем контекст
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata in context")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
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

func (s *mockBankCardServer) Get(ctx context.Context, req *pb.SecretBankCardGetRequest) (*pb.SecretBankCardGetResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata in context")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretBankCardGetResponse{
		Card: &pb.SecretBankCard{
			SecretName: "card1",
			Owner:      "John Doe",
			Number:     "1234567890123456",
			Exp:        "12/25",
			Cvv:        "123",
			Meta:       "",
			UpdatedAt:  time.Now().Format(time.RFC3339),
		},
	}, nil
}

func (s *mockBankCardServer) Save(ctx context.Context, req *pb.SecretBankCardSaveRequest) (*pb.SecretBankCardSaveResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata in context")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}
	return &pb.SecretBankCardSaveResponse{}, nil
}

func setupGRPCServer(t *testing.T) *grpc.ClientConn {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretBankCardServiceServer(server, &mockBankCardServer{})
	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(func() { server.Stop() })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

	return conn
}

func TestSecretBankCardListGRPCFacade_List(t *testing.T) {
	conn := setupGRPCServer(t)
	client := pb.NewSecretBankCardServiceClient(conn)
	facade := NewSecretBankCardListGRPCFacade(client)

	secrets, err := facade.List(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "card1", secrets[0].SecretName)
}

func TestSecretBankCardGetGRPCFacade_Get(t *testing.T) {
	conn := setupGRPCServer(t)
	client := pb.NewSecretBankCardServiceClient(conn)
	facade := NewSecretBankCardGetGRPCFacade(client)

	secret, err := facade.Get(context.Background(), "test-token", "card1")
	assert.NoError(t, err)
	assert.Equal(t, "card1", secret.SecretName)
}

func TestSecretBankCardSaveGRPCFacade_Save(t *testing.T) {
	conn := setupGRPCServer(t)
	client := pb.NewSecretBankCardServiceClient(conn)
	facade := NewSecretBankCardSaveGRPCFacade(client)

	secret := models.SecretBankCardClient{
		SecretName: "card1",
		Owner:      "John Doe",
		Number:     "1234567890123456",
		Exp:        "12/25",
		CVV:        "123",
		UpdatedAt:  time.Now(),
	}

	err := facade.Save(context.Background(), "test-token", secret)
	assert.NoError(t, err)
}
