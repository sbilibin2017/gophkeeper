package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	_ "modernc.org/sqlite"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func setupListBankCardsLocalTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE secret_bank_card_request (
		secret_name TEXT PRIMARY KEY,
		number TEXT NOT NULL,
		owner TEXT NOT NULL,
		exp TEXT NOT NULL,
		cvv TEXT NOT NULL,
		meta TEXT
	);`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestListBankCardsLocal(t *testing.T) {
	db := setupListBankCardsLocalTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO secret_bank_card_request (secret_name, number, owner, exp, cvv, meta)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"card1", "1111222233334444", "John Doe", "12/25", "123", "test-meta")
	assert.NoError(t, err)

	ctx := context.Background()
	results, err := ListBankCardsLocal(ctx, db)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "card1", results[0].SecretName)
	assert.Equal(t, "1111222233334444", results[0].Number)
	assert.Equal(t, "John Doe", results[0].Owner)
	assert.NotNil(t, results[0].Meta)
	assert.Equal(t, "test-meta", *results[0].Meta)
}

func TestListBankCardsHTTP(t *testing.T) {
	meta := "from-server"
	expected := struct {
		Items []models.BankCardResponse `json:"items"`
	}{
		Items: []models.BankCardResponse{
			{
				SecretName:  "card_123",
				SecretOwner: "alice",
				Number:      "9999888877776666",
				Owner:       "Alice Smith",
				Exp:         "11/27",
				CVV:         "999",
				Meta:        &meta,
				UpdatedAt:   time.Now(),
			},
		},
	}

	// Create a test HTTP server to mock the /list/bank-card endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/list/bank-card", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(expected)
		require.NoError(t, err)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	results, err := ListBankCardsHTTP(ctx, client, "test-token")
	require.NoError(t, err)
	require.Len(t, results, 1)

	require.NotNil(t, results[0].Meta)
	require.NotNil(t, expected.Items[0].Meta)

	assert.Equal(t, expected.Items[0].SecretName, results[0].SecretName)
	assert.Equal(t, expected.Items[0].SecretOwner, results[0].SecretOwner)
	assert.Equal(t, *expected.Items[0].Meta, *results[0].Meta)
	assert.Equal(t, expected.Items[0].Number, results[0].Number)
	assert.Equal(t, expected.Items[0].Owner, results[0].Owner)
	assert.Equal(t, expected.Items[0].Exp, results[0].Exp)
	assert.Equal(t, expected.Items[0].CVV, results[0].CVV)
}

type stubBankCardListClient struct{}

func (s *stubBankCardListClient) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.BankCardListResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer test-token" {
		return nil, assert.AnError
	}

	meta := "grpc-meta"
	return &pb.BankCardListResponse{
		Items: []*pb.BankCardGetResponse{
			{
				SecretName:  "card_001",
				SecretOwner: "john",
				Number:      "1234567890123456",
				Owner:       "John Doe",
				Exp:         "10/24",
				Cvv:         "123",
				Meta:        meta,
				UpdatedAt:   time.Now().Format(time.RFC3339), // Set updated_at string
			},
			{
				SecretName:  "card_002",
				SecretOwner: "jane",
				Number:      "6543210987654321",
				Owner:       "Jane Roe",
				Exp:         "09/23",
				Cvv:         "321",
				Meta:        "",
				UpdatedAt:   "",
			},
		},
	}, nil
}

func TestListBankCardsGRPC(t *testing.T) {
	clientStub := &stubBankCardListClient{}
	ctx := context.Background()

	result, err := ListBankCardsGRPC(ctx, clientStub, "test-token")
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, "card_001", result[0].SecretName)
	assert.Equal(t, "john", result[0].SecretOwner)
	assert.Equal(t, "1234567890123456", result[0].Number)
	assert.Equal(t, "John Doe", result[0].Owner)
	assert.Equal(t, "10/24", result[0].Exp)
	assert.Equal(t, "123", result[0].CVV)
	require.NotNil(t, result[0].Meta)
	assert.Equal(t, "grpc-meta", *result[0].Meta)
	assert.False(t, result[0].UpdatedAt.IsZero())

	assert.Equal(t, "card_002", result[1].SecretName)
	assert.Equal(t, "jane", result[1].SecretOwner)
	assert.Equal(t, "6543210987654321", result[1].Number)
	assert.Equal(t, "Jane Roe", result[1].Owner)
	assert.Equal(t, "09/23", result[1].Exp)
	assert.Equal(t, "321", result[1].CVV)
	assert.Nil(t, result[1].Meta)
	assert.True(t, result[1].UpdatedAt.IsZero())
}
