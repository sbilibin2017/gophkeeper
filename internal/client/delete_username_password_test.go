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

// ---------- Test for DeleteUsernamePasswordLocal ----------

func setupDeleteUsernamePasswordLocalTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_username_password_request (
		secret_name TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		meta TEXT
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO secret_username_password_request (secret_name, username, password, meta)
		VALUES (?, ?, ?, ?)`,
		"mysecret", "user1", "pass1", "test-meta")
	require.NoError(t, err)

	return db
}

func TestDeleteUsernamePasswordLocal(t *testing.T) {
	db := setupDeleteUsernamePasswordLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := DeleteUsernamePasswordLocal(ctx, db, "mysecret")
	require.NoError(t, err)

	// Verify deletion
	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM secret_username_password_request WHERE secret_name = ?", "mysecret")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestDeleteUsernamePasswordLocal_NotFound(t *testing.T) {
	db := setupDeleteUsernamePasswordLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()
	err := DeleteUsernamePasswordLocal(ctx, db, "unknown_secret")
	require.NoError(t, err)
}

// ---------- Test for DeleteUsernamePasswordHTTP ----------

func TestDeleteUsernamePasswordHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/delete/username-password/mysecret", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	err := DeleteUsernamePasswordHTTP(ctx, client, "test-token", "mysecret")
	require.NoError(t, err)
}

func TestDeleteUsernamePasswordHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	err := DeleteUsernamePasswordHTTP(ctx, client, "bad-token", "nonexistent")
	assert.Error(t, err)
}

// ---------- Test for DeleteUsernamePasswordGRPC ----------

type stubUsernamePasswordDeleteClient struct{}

func (s *stubUsernamePasswordDeleteClient) Delete(
	ctx context.Context,
	in *pb.UsernamePasswordDeleteRequest,
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

func TestDeleteUsernamePasswordGRPC(t *testing.T) {
	client := &stubUsernamePasswordDeleteClient{}
	ctx := context.Background()

	err := DeleteUsernamePasswordGRPC(ctx, client, "grpc-token", "mysecret")
	require.NoError(t, err)
}

func TestDeleteUsernamePasswordGRPC_Unauthorized(t *testing.T) {
	client := &stubUsernamePasswordDeleteClient{}
	ctx := context.Background()

	err := DeleteUsernamePasswordGRPC(ctx, client, "bad-token", "mysecret")
	assert.Error(t, err)
}

func TestDeleteUsernamePasswordGRPC_NotFound(t *testing.T) {
	client := &stubUsernamePasswordDeleteClient{}
	ctx := context.Background()

	err := DeleteUsernamePasswordGRPC(ctx, client, "grpc-token", "nonexistent")
	assert.Error(t, err)
}
