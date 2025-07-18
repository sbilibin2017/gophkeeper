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

func setupUsernamePasswordTestDB(t *testing.T) *sqlx.DB {
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

func TestGetUsernamePasswordLocal(t *testing.T) {
	db := setupUsernamePasswordTestDB(t)
	defer db.Close()

	meta := "local-meta"
	_, err := db.Exec(`
		INSERT INTO secret_username_password_request (secret_name, username, password, meta)
		VALUES (?, ?, ?, ?)`,
		"up1", "localuser", "localpass", meta)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := GetUsernamePasswordLocal(ctx, db, "up1")

	require.NoError(t, err)
	assert.Equal(t, "up1", result.SecretName)
	assert.Equal(t, "localuser", result.Username)
	assert.Equal(t, "localpass", result.Password)
	require.NotNil(t, result.Meta)
	assert.Equal(t, meta, *result.Meta)
}

func TestGetUsernamePasswordHTTP(t *testing.T) {
	meta := "http-meta"
	expected := struct {
		Item models.UsernamePasswordResponse `json:"item"`
	}{
		Item: models.UsernamePasswordResponse{
			SecretName:  "up-http",
			SecretOwner: "user-http",
			Username:    "http-user",
			Password:    "http-pass",
			Meta:        &meta,
			UpdatedAt:   time.Now(),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		assert.Equal(t, "/get/username-password/up-http", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(expected))
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	result, err := GetUsernamePasswordHTTP(ctx, client, "token123", "up-http")
	require.NoError(t, err)

	assert.Equal(t, expected.Item.SecretName, result.SecretName)
	assert.Equal(t, expected.Item.SecretOwner, result.SecretOwner)
	assert.Equal(t, expected.Item.Username, result.Username)
	assert.Equal(t, expected.Item.Password, result.Password)
	require.NotNil(t, result.Meta)
	assert.Equal(t, *expected.Item.Meta, *result.Meta)
	assert.WithinDuration(t, expected.Item.UpdatedAt, result.UpdatedAt, time.Second)
}

type stubUsernamePasswordClient struct{}

func (s *stubUsernamePasswordClient) Get(ctx context.Context, in *pb.UsernamePasswordGetRequest, opts ...grpc.CallOption) (*pb.UsernamePasswordGetResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer token456" {
		return nil, fmt.Errorf("unauthorized")
	}

	meta := "grpc-meta"
	return &pb.UsernamePasswordGetResponse{
		SecretName:  "up-grpc",
		SecretOwner: "user-grpc",
		Username:    "grpc-user",
		Password:    "grpc-pass",
		Meta:        meta,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func TestGetUsernamePasswordGRPC(t *testing.T) {
	clientStub := &stubUsernamePasswordClient{}
	ctx := context.Background()

	result, err := GetUsernamePasswordGRPC(ctx, clientStub, "token456", "up-grpc")
	require.NoError(t, err)

	assert.Equal(t, "up-grpc", result.SecretName)
	assert.Equal(t, "user-grpc", result.SecretOwner)
	assert.Equal(t, "grpc-user", result.Username)
	assert.Equal(t, "grpc-pass", result.Password)
	require.NotNil(t, result.Meta)
	assert.Equal(t, "grpc-meta", *result.Meta)
	assert.False(t, result.UpdatedAt.IsZero())
}
