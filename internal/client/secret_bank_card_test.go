package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	_ "modernc.org/sqlite"
)

// --- HTTP tests ---

func TestSaveSecretBankCardHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/save/secret-bank-card" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.NotEmpty(t, body["secret_name"])
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)
	card := models.SecretBankCardSaveRequest{
		SecretName: "card1",
		Owner:      "John",
		Number:     "4111111111111111",
		Exp:        "12/25",
		CVV:        "123",
	}

	err := SaveSecretBankCardHTTP(context.Background(), clientHTTP, "testtoken", card)
	assert.NoError(t, err)
}

func TestGetSecretBankCardHTTP(t *testing.T) {
	cardResponse := models.SecretBankCardGetResponse{
		SecretName:  "card1",
		Owner:       "John",
		Number:      "4111111111111111",
		Exp:         "12/25",
		CVV:         "123",
		SecretOwner: "John",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/get/secret-bank-card/card1" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		respBytes, _ := json.Marshal(cardResponse)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)

	secret, err := GetSecretBankCardHTTP(context.Background(), clientHTTP, "testtoken", "card1")
	require.NoError(t, err)
	assert.Equal(t, cardResponse.SecretName, secret.SecretName)
	assert.Equal(t, cardResponse.Owner, secret.Owner)
	assert.Equal(t, cardResponse.Number, secret.Number)
	assert.Equal(t, cardResponse.Exp, secret.Exp)
	assert.Equal(t, cardResponse.CVV, secret.CVV)
}

func TestListSecretBankCardHTTP(t *testing.T) {
	listResponse := []models.SecretBankCardGetResponse{
		{SecretName: "card1", Owner: "John", Number: "1111", Exp: "01/23", CVV: "123"},
		{SecretName: "card2", Owner: "Jane", Number: "2222", Exp: "02/24", CVV: "234"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/secret-bank-card" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		respBytes, _ := json.Marshal(listResponse)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)

	secrets, err := ListSecretBankCardHTTP(context.Background(), clientHTTP, "testtoken")
	require.NoError(t, err)
	require.Len(t, secrets, 2)
	assert.Equal(t, "card1", secrets[0].SecretName)
	assert.Equal(t, "card2", secrets[1].SecretName)
}

type mockSecretBankCardService struct {
	pb.UnimplementedSecretBankCardServiceServer
}

func (m *mockSecretBankCardService) Save(ctx context.Context, req *pb.SecretBankCardSaveRequest) (*pb.SecretBankCardSaveResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	auth := md.Get("authorization")
	if len(auth) == 0 || auth[0] != "Bearer testtoken" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return &pb.SecretBankCardSaveResponse{}, nil
}

func (m *mockSecretBankCardService) Get(ctx context.Context, req *pb.SecretBankCardGetRequest) (*pb.SecretBankCardGetResponse, error) {
	return &pb.SecretBankCardGetResponse{
		SecretName:  req.SecretName,
		SecretOwner: "John Doe",
		Number:      "4111111111111111",
		Owner:       "John Doe",
		Exp:         "12/25",
		Cvv:         "123",
		Meta:        "",
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func (m *mockSecretBankCardService) List(ctx context.Context, req *pb.SecretBankCardListRequest) (*pb.SecretBankCardListResponse, error) {
	return &pb.SecretBankCardListResponse{
		Items: []*pb.SecretBankCardGetResponse{
			{
				SecretName:  "card1",
				SecretOwner: "John Doe",
				Number:      "4111111111111111",
				Owner:       "John Doe",
				Exp:         "12/25",
				Cvv:         "123",
				Meta:        "",
				UpdatedAt:   time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func startMockGRPCServer(t *testing.T) (pb.SecretBankCardServiceClient, func()) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pb.RegisterSecretBankCardServiceServer(server, &mockSecretBankCardService{})

	go server.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)

	client := pb.NewSecretBankCardServiceClient(conn)

	cleanup := func() {
		server.Stop()
		conn.Close()
	}

	return client, cleanup
}

func TestSaveSecretBankCardGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCServer(t)
	defer cleanup()

	card := models.SecretBankCardSaveRequest{
		SecretName: "card1",
		Owner:      "John Doe",
		Number:     "4111111111111111",
		Exp:        "12/25",
		CVV:        "123",
	}

	err := SaveSecretBankCardGRPC(context.Background(), clientGRPC, "testtoken", card)
	assert.NoError(t, err)
}

func TestGetSecretBankCardGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCServer(t)
	defer cleanup()

	secret, err := GetSecretBankCardGRPC(context.Background(), clientGRPC, "testtoken", "card1")
	assert.NoError(t, err)
	assert.Equal(t, "card1", secret.SecretName)
	assert.Equal(t, "John Doe", secret.Owner)
}

func TestListSecretBankCardGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCServer(t)
	defer cleanup()

	secrets, err := ListSecretBankCardGRPC(context.Background(), clientGRPC, "testtoken")
	assert.NoError(t, err)
	assert.NotEmpty(t, secrets)
	assert.Equal(t, "card1", secrets[0].SecretName)
}

func secretBankCardDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_bank_card_request (
		secret_name TEXT PRIMARY KEY,
		number TEXT,
		owner TEXT,
		exp TEXT,
		cvv TEXT,
		meta TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSaveAndGetSecretBankCardRequest(t *testing.T) {
	ctx := context.Background()
	db := secretBankCardDB(t)
	defer db.Close()

	metaJSON := `{"some":"metadata"}`
	card := models.SecretBankCardSaveRequest{
		SecretName: "card1",
		Number:     "1234567890123456",
		Owner:      "John Doe",
		Exp:        "12/25",
		CVV:        "123",
		Meta:       &metaJSON,
	}

	// Сохраняем карту
	err := SaveSecretBankCardRequest(ctx, db, card)
	require.NoError(t, err)

	// Получаем список всех секретов (только имена)
	secrets, err := GetAllSecretsBankCardRequest(ctx, db)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "card1", secrets[0].SecretName)

	// Получаем карту по имени
	cardResp, err := GetSecretBankCardByNameRequest(ctx, db, "card1")
	require.NoError(t, err)
	assert.Equal(t, card.SecretName, cardResp.SecretName)
	assert.Equal(t, card.Number, cardResp.Number)
	assert.Equal(t, card.Owner, cardResp.Owner)
	assert.Equal(t, card.Exp, cardResp.Exp)
	assert.Equal(t, card.CVV, cardResp.CVV)

	assert.NotNil(t, cardResp.Meta)
	assert.Equal(t, metaJSON, *cardResp.Meta)

	// Обновляем карту с другим meta и number
	newMetaJSON := `{"updated":"data"}`
	card.Number = "9999888877776666"
	card.Meta = &newMetaJSON

	err = SaveSecretBankCardRequest(ctx, db, card)
	require.NoError(t, err)

	updatedCardResp, err := GetSecretBankCardByNameRequest(ctx, db, "card1")
	require.NoError(t, err)
	assert.Equal(t, "9999888877776666", updatedCardResp.Number)
	assert.NotNil(t, updatedCardResp.Meta)
	assert.Equal(t, newMetaJSON, *updatedCardResp.Meta)
}

func TestGetSecretBankCardByNameRequest_NotFound(t *testing.T) {
	ctx := context.Background()
	db := secretBankCardDB(t)
	defer db.Close()

	cardResp, err := GetSecretBankCardByNameRequest(ctx, db, "missing_card")
	assert.Nil(t, cardResp)
	assert.Error(t, err)
	assert.Equal(t, "secret not found or error fetching", err.Error())
}
