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

// ---------- Test for DeleteTextLocal ----------

func setupDeleteTextLocalTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_text_request (
		secret_name TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		meta TEXT
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO secret_text_request (secret_name, content, meta)
		VALUES (?, ?, ?)`,
		"mytext", "sample content", "test-meta")
	require.NoError(t, err)

	return db
}

func TestDeleteTextLocal(t *testing.T) {
	db := setupDeleteTextLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := DeleteTextLocal(ctx, db, "mytext")
	require.NoError(t, err)

	// Verify deletion
	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM secret_text_request WHERE secret_name = ?", "mytext")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteTextLocal_NotFound(t *testing.T) {
	db := setupDeleteTextLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := DeleteTextLocal(ctx, db, "unknown_text")
	require.NoError(t, err)
}

// ---------- Test for DeleteTextHTTP ----------

func TestDeleteTextHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/delete/text/mytext", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	err := DeleteTextHTTP(ctx, client, "test-token", "mytext")
	require.NoError(t, err)
}

func TestDeleteTextHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	err := DeleteTextHTTP(ctx, client, "bad-token", "nonexistent")
	assert.Error(t, err)
}

// ---------- Test for DeleteTextGRPC ----------

type stubTextDeleteClient struct{}

func (s *stubTextDeleteClient) Delete(
	ctx context.Context,
	in *pb.TextDeleteRequest,
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

func TestDeleteTextGRPC(t *testing.T) {
	client := &stubTextDeleteClient{}
	ctx := context.Background()

	err := DeleteTextGRPC(ctx, client, "grpc-token", "mytext")
	require.NoError(t, err)
}

func TestDeleteTextGRPC_Unauthorized(t *testing.T) {
	client := &stubTextDeleteClient{}
	ctx := context.Background()

	err := DeleteTextGRPC(ctx, client, "bad-token", "mytext")
	assert.Error(t, err)
}

func TestDeleteTextGRPC_NotFound(t *testing.T) {
	client := &stubTextDeleteClient{}
	ctx := context.Background()

	err := DeleteTextGRPC(ctx, client, "grpc-token", "nonexistent")
	assert.Error(t, err)
}
