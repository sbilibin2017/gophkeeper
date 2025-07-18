package client

import (
	"context"
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

func setupTextTestDB(t *testing.T) *sqlx.DB {
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

func TestGetTextLocal(t *testing.T) {
	db := setupTextTestDB(t)
	defer db.Close()

	meta := "meta-local"
	_, err := db.Exec(`
		INSERT INTO secret_text_request (secret_name, content, meta)
		VALUES (?, ?, ?)`,
		"text1", "secret-content", meta)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := GetTextLocal(ctx, db, "text1")

	require.NoError(t, err)
	require.Equal(t, "text1", result.SecretName)
	require.Equal(t, "secret-content", result.Content)
	require.NotNil(t, result.Meta)
	assert.Equal(t, meta, *result.Meta)
}

func TestGetTextHTTP(t *testing.T) {
	meta := "meta-http"
	expected := struct {
		Item models.TextResponse `json:"item"`
	}{
		Item: models.TextResponse{
			SecretName:  "text123",
			SecretOwner: "user-http",
			Content:     "http-content",
			Meta:        &meta,
			UpdatedAt:   time.Now(),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/get/text/text123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(expected))
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	result, err := GetTextHTTP(ctx, client, "test-token", "text123")
	require.NoError(t, err)

	assert.Equal(t, expected.Item.SecretName, result.SecretName)
	assert.Equal(t, expected.Item.SecretOwner, result.SecretOwner)
	assert.Equal(t, expected.Item.Content, result.Content)
	require.NotNil(t, result.Meta)
	assert.Equal(t, *expected.Item.Meta, *result.Meta)
	assert.WithinDuration(t, expected.Item.UpdatedAt, result.UpdatedAt, time.Second)
}

type stubTextClient struct{}

func (s *stubTextClient) Get(ctx context.Context, in *pb.TextGetRequest, opts ...grpc.CallOption) (*pb.TextGetResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer test-token" {
		return nil, fmt.Errorf("unauthorized")
	}

	meta := "grpc-meta"
	return &pb.TextGetResponse{
		SecretName:  "grpc-text",
		SecretOwner: "grpc-user",
		Content:     "grpc-content",
		Meta:        meta,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func TestGetTextGRPC(t *testing.T) {
	clientStub := &stubTextClient{}
	ctx := context.Background()

	result, err := GetTextGRPC(ctx, clientStub, "test-token", "grpc-text")
	require.NoError(t, err)

	assert.Equal(t, "grpc-text", result.SecretName)
	assert.Equal(t, "grpc-user", result.SecretOwner)
	assert.Equal(t, "grpc-content", result.Content)
	require.NotNil(t, result.Meta)
	assert.Equal(t, "grpc-meta", *result.Meta)
	assert.False(t, result.UpdatedAt.IsZero())
}
