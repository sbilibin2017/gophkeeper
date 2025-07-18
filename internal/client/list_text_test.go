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

// Setup an in-memory SQLite DB for local tests
func setupListTextLocalTestDB(t *testing.T) *sqlx.DB {
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

	return db
}

func TestListTextLocal(t *testing.T) {
	db := setupListTextLocalTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO secret_text_request (secret_name, content, meta)
		VALUES (?, ?, ?)`,
		"text1", "my secret content", "test-meta")
	require.NoError(t, err)

	ctx := context.Background()
	results, err := ListTextLocal(ctx, db)
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, "text1", results[0].SecretName)
	assert.Equal(t, "my secret content", results[0].Content)
	require.NotNil(t, results[0].Meta)
	assert.Equal(t, "test-meta", *results[0].Meta)
}

func TestListTextHTTP(t *testing.T) {
	meta := "http-meta"
	respItems := []models.TextResponse{
		{
			SecretName: "text_123",
			Content:    "http secret",
			Meta:       &meta,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/list/text", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(struct {
			Items []models.TextResponse `json:"items"`
		}{Items: respItems})
		require.NoError(t, err)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	results, err := ListTextHTTP(ctx, client, "test-token")
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, respItems[0].SecretName, results[0].SecretName)
	assert.Equal(t, respItems[0].Content, results[0].Content)
	require.NotNil(t, results[0].Meta)
	assert.Equal(t, *respItems[0].Meta, *results[0].Meta)
}

// Stub for pb.TextListServiceClient with correct signature
type stubTextListClient struct{}

func (s *stubTextListClient) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.TextListResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer test-token" {
		return nil, fmt.Errorf("unauthorized")
	}

	meta := "grpc-meta"
	return &pb.TextListResponse{
		Items: []*pb.TextGetResponse{
			{
				SecretName: "text_001",
				Content:    "grpc content 1",
				Meta:       meta,
			},
			{
				SecretName: "text_002",
				Content:    "grpc content 2",
				Meta:       "",
			},
		},
	}, nil
}

func TestListTextGRPC(t *testing.T) {
	clientStub := &stubTextListClient{}
	ctx := context.Background()

	result, err := ListTextGRPC(ctx, clientStub, "test-token")
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, "text_001", result[0].SecretName)
	assert.Equal(t, "grpc content 1", result[0].Content)
	require.NotNil(t, result[0].Meta)
	assert.Equal(t, "grpc-meta", *result[0].Meta)

	assert.Equal(t, "text_002", result[1].SecretName)
	assert.Equal(t, "grpc content 2", result[1].Content)
	assert.Nil(t, result[1].Meta)
}
