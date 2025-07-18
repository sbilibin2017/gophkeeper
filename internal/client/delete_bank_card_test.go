package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	_ "modernc.org/sqlite"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// ---------- Test for DeleteBankCardLocal ----------

func setupDeleteBankCardLocalTestDB(t *testing.T) *sqlx.DB {
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

func TestDeleteBankCardLocal(t *testing.T) {
	db := setupDeleteBankCardLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := DeleteBankCardLocal(ctx, db, "mycard")
	require.NoError(t, err)

	// Verify deletion
	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM secret_bank_card_request WHERE secret_name = ?", "mycard")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteBankCardLocal_NotFound(t *testing.T) {
	db := setupDeleteBankCardLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := DeleteBankCardLocal(ctx, db, "unknown_card")
	// Deleting nonexistent row should not error, just affect zero rows
	require.NoError(t, err)
}

// ---------- Test for DeleteBankCardHTTP ----------

func TestDeleteBankCardHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/delete/bank-card/mycard", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	err := DeleteBankCardHTTP(ctx, client, "test-token", "mycard")
	require.NoError(t, err)
}

func TestDeleteBankCardHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	err := DeleteBankCardHTTP(ctx, client, "bad-token", "nonexistent")
	assert.Error(t, err)
}

// ---------- Test for DeleteBankCardGRPC ----------

type stubBankCardDeleteClient struct{}

func (s *stubBankCardDeleteClient) Delete(
	ctx context.Context,
	in *pb.BankCardDeleteRequest,
	opts ...grpc.CallOption,
) (*emptypb.Empty, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer grpc-token" {
		return nil, fmt.Errorf("unauthorized")
	}

	if in.SecretName == "nonexistent" {
		return nil, fmt.Errorf("not found")
	}

	return &emptypb.Empty{}, nil
}

func TestDeleteBankCardGRPC(t *testing.T) {
	client := &stubBankCardDeleteClient{}
	ctx := context.Background()

	err := DeleteBankCardGRPC(ctx, client, "grpc-token", "mycard")
	require.NoError(t, err)
}

func TestDeleteBankCardGRPC_Unauthorized(t *testing.T) {
	client := &stubBankCardDeleteClient{}
	ctx := context.Background()

	err := DeleteBankCardGRPC(ctx, client, "bad-token", "mycard")
	assert.Error(t, err)
}

func TestDeleteBankCardGRPC_NotFound(t *testing.T) {
	client := &stubBankCardDeleteClient{}
	ctx := context.Background()

	err := DeleteBankCardGRPC(ctx, client, "grpc-token", "nonexistent")
	assert.Error(t, err)
}
