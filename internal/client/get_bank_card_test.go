package client

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	_ "modernc.org/sqlite"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// ---------- Test for GetBankCardLocal ----------

func setupGetBankCardLocalTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

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
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO secret_bank_card_request (secret_name, number, owner, exp, cvv, meta)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"mycard", "4444333322221111", "Mark Smith", "08/26", "456", "test-meta")
	require.NoError(t, err)

	return db
}

func TestGetBankCardLocal(t *testing.T) {
	db := setupGetBankCardLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	result, err := GetBankCardLocal(ctx, db, "mycard")

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "mycard", result.SecretName)
	assert.Equal(t, "4444333322221111", result.Number)
	assert.Equal(t, "Mark Smith", result.Owner)
	assert.Equal(t, "08/26", result.Exp)
	assert.Equal(t, "456", result.CVV)
	assert.NotNil(t, result.Meta)
	assert.Equal(t, "test-meta", *result.Meta)
}

func TestGetBankCardLocal_NotFound(t *testing.T) {
	db := setupGetBankCardLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	result, err := GetBankCardLocal(ctx, db, "unknown_card")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, sql.ErrNoRows, err)
}

// ---------- Test for GetBankCardHTTP ----------

func TestGetBankCardHTTP(t *testing.T) {
	meta := "some-meta"
	expected := struct {
		Item models.BankCardResponse `json:"item"`
	}{
		Item: models.BankCardResponse{
			SecretName:  "card_test",
			SecretOwner: "tester",
			Number:      "1234123412341234",
			Owner:       "Test Owner",
			Exp:         "05/30",
			CVV:         "111",
			Meta:        &meta,
			UpdatedAt:   time.Now(),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/get/bank-card/card_test", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(expected)
		require.NoError(t, err)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	result, err := GetBankCardHTTP(ctx, client, "test-token", "card_test")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, expected.Item.SecretName, result.SecretName)
	assert.Equal(t, expected.Item.SecretOwner, result.SecretOwner)
	assert.Equal(t, expected.Item.Number, result.Number)
	assert.Equal(t, expected.Item.Owner, result.Owner)
	assert.Equal(t, expected.Item.Exp, result.Exp)
	assert.Equal(t, expected.Item.CVV, result.CVV)
	assert.NotNil(t, result.Meta)
	assert.Equal(t, *expected.Item.Meta, *result.Meta)
}

func TestGetBankCardHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	result, err := GetBankCardHTTP(ctx, client, "bad-token", "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---------- Test for GetBankCardGRPC ----------

type stubBankCardGetClient struct{}

func (s *stubBankCardGetClient) Get(
	ctx context.Context,
	in *pb.BankCardGetRequest,
	opts ...grpc.CallOption,
) (*pb.BankCardGetResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer grpc-token" {
		return nil, fmt.Errorf("unauthorized")
	}

	return &pb.BankCardGetResponse{
		SecretName:  "grpc_card",
		SecretOwner: "grpc_user",
		Number:      "5555666677778888",
		Owner:       "GRPC Owner",
		Exp:         "07/29",
		Cvv:         "321",
		Meta:        "grpc-meta",
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func TestGetBankCardGRPC(t *testing.T) {
	client := &stubBankCardGetClient{}
	ctx := context.Background()

	resp, err := GetBankCardGRPC(ctx, client, "grpc-token", "grpc_card")
	require.NoError(t, err)

	assert.Equal(t, "grpc_card", resp.SecretName)
	assert.Equal(t, "grpc_user", resp.SecretOwner)
	assert.Equal(t, "5555666677778888", resp.Number)
	assert.Equal(t, "GRPC Owner", resp.Owner)
	assert.Equal(t, "07/29", resp.Exp)
	assert.Equal(t, "321", resp.CVV)
	assert.NotNil(t, resp.Meta)
	assert.Equal(t, "grpc-meta", *resp.Meta)
	assert.False(t, resp.UpdatedAt.IsZero())
}

func TestGetBankCardGRPC_Unauthorized(t *testing.T) {
	client := &stubBankCardGetClient{}
	ctx := context.Background()

	_, err := GetBankCardGRPC(ctx, client, "invalid-token", "grpc_card")
	assert.Error(t, err)
}
