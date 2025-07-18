package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	_ "modernc.org/sqlite"
)

// Setup in-memory SQLite for local tests
func setupListUsernamePasswordLocalTestDB(t *testing.T) *sqlx.DB {
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

	return db
}

func TestListUsernamePasswordLocal(t *testing.T) {
	db := setupListUsernamePasswordLocalTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO secret_username_password_request (secret_name, username, password, meta)
		VALUES (?, ?, ?, ?)`,
		"login1", "user1", "pass1", "local-meta")
	require.NoError(t, err)

	ctx := context.Background()
	results, err := ListUsernamePasswordLocal(ctx, db)
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, "login1", results[0].SecretName)
	assert.Equal(t, "user1", results[0].Username)
	assert.Equal(t, "pass1", results[0].Password)
	require.NotNil(t, results[0].Meta)
	assert.Equal(t, "local-meta", *results[0].Meta)
}

func TestListUsernamePasswordHTTP(t *testing.T) {
	meta := "http-meta"
	respItems := []models.UsernamePasswordResponse{
		{
			SecretName: "login_123",
			Username:   "http-user",
			Password:   "http-pass",
			Meta:       &meta,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/list/username-password", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(struct {
			Items []models.UsernamePasswordResponse `json:"items"`
		}{Items: respItems})
		require.NoError(t, err)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	results, err := ListUsernamePasswordHTTP(ctx, client, "test-token")
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, respItems[0].SecretName, results[0].SecretName)
	assert.Equal(t, respItems[0].Username, results[0].Username)
	assert.Equal(t, respItems[0].Password, results[0].Password)
	require.NotNil(t, results[0].Meta)
	assert.Equal(t, *respItems[0].Meta, *results[0].Meta)
}

// Stub implementing pb.UsernamePasswordListServiceClient with correct signature
type stubUsernamePasswordListClient struct{}

func (s *stubUsernamePasswordListClient) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.UsernamePasswordListResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer test-token" {
		return nil, fmt.Errorf("unauthorized")
	}

	meta := "grpc-meta"
	return &pb.UsernamePasswordListResponse{
		Items: []*pb.UsernamePasswordGetResponse{
			{
				SecretName: "login_001",
				Username:   "grpc-user1",
				Password:   "grpc-pass1",
				Meta:       meta,
			},
			{
				SecretName: "login_002",
				Username:   "grpc-user2",
				Password:   "grpc-pass2",
				Meta:       "",
			},
		},
	}, nil
}

func TestListUsernamePasswordGRPC(t *testing.T) {
	clientStub := &stubUsernamePasswordListClient{}
	ctx := context.Background()

	result, err := ListUsernamePasswordGRPC(ctx, clientStub, "test-token")
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, "login_001", result[0].SecretName)
	assert.Equal(t, "grpc-user1", result[0].Username)
	assert.Equal(t, "grpc-pass1", result[0].Password)
	require.NotNil(t, result[0].Meta)
	assert.Equal(t, "grpc-meta", *result[0].Meta)

	assert.Equal(t, "login_002", result[1].SecretName)
	assert.Equal(t, "grpc-user2", result[1].Username)
	assert.Equal(t, "grpc-pass2", result[1].Password)
	assert.Nil(t, result[1].Meta)
}
