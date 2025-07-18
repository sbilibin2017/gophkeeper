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

// ---------- Test for DeleteBinaryLocal ----------

func setupDeleteBinaryLocalTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_binary_request (
		secret_name TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		meta TEXT
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO secret_binary_request (secret_name, data, meta)
		VALUES (?, ?, ?)`,
		"mybinary", []byte{0x01, 0x02, 0x03}, "test-meta")
	require.NoError(t, err)

	return db
}

func TestDeleteBinaryLocal(t *testing.T) {
	db := setupDeleteBinaryLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := DeleteBinaryLocal(ctx, db, "mybinary")
	require.NoError(t, err)

	// Verify deletion
	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM secret_binary_request WHERE secret_name = ?", "mybinary")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteBinaryLocal_NotFound(t *testing.T) {
	db := setupDeleteBinaryLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := DeleteBinaryLocal(ctx, db, "unknown_binary")
	require.NoError(t, err)
}

// ---------- Test for DeleteBinaryHTTP ----------

func TestDeleteBinaryHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/delete/binary/mybinary", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	err := DeleteBinaryHTTP(ctx, client, "test-token", "mybinary")
	require.NoError(t, err)
}

func TestDeleteBinaryHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	err := DeleteBinaryHTTP(ctx, client, "bad-token", "nonexistent")
	assert.Error(t, err)
}

// ---------- Test for DeleteBinaryGRPC ----------

type stubBinaryDeleteClient struct{}

func (s *stubBinaryDeleteClient) Delete(
	ctx context.Context,
	in *pb.BinaryDeleteRequest,
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

func TestDeleteBinaryGRPC(t *testing.T) {
	client := &stubBinaryDeleteClient{}
	ctx := context.Background()

	err := DeleteBinaryGRPC(ctx, client, "grpc-token", "mybinary")
	require.NoError(t, err)
}

func TestDeleteBinaryGRPC_Unauthorized(t *testing.T) {
	client := &stubBinaryDeleteClient{}
	ctx := context.Background()

	err := DeleteBinaryGRPC(ctx, client, "bad-token", "mybinary")
	assert.Error(t, err)
}

func TestDeleteBinaryGRPC_NotFound(t *testing.T) {
	client := &stubBinaryDeleteClient{}
	ctx := context.Background()

	err := DeleteBinaryGRPC(ctx, client, "grpc-token", "nonexistent")
	assert.Error(t, err)
}
