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

func setupGetBinaryTestDB(t *testing.T) *sqlx.DB {
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

	return db
}

func TestGetBinaryLocal(t *testing.T) {
	db := setupGetBinaryTestDB(t)
	defer db.Close()

	meta := "local-meta"
	_, err := db.Exec(`
		INSERT INTO secret_binary_request (secret_name, data, meta)
		VALUES (?, ?, ?)`,
		"binary1", []byte("test-data"), meta)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := GetBinaryLocal(ctx, db, "binary1")

	require.NoError(t, err)
	require.Equal(t, "binary1", result.SecretName)
	require.Equal(t, []byte("test-data"), result.Data)
	require.NotNil(t, result.Meta)
	assert.Equal(t, meta, *result.Meta)
}

func TestGetBinaryHTTP(t *testing.T) {
	meta := "http-meta"
	expected := struct {
		Item models.BinaryResponse `json:"item"`
	}{
		Item: models.BinaryResponse{
			SecretName:  "bin123",
			SecretOwner: "user1",
			Data:        []byte("binary-content"),
			Meta:        &meta,
			UpdatedAt:   time.Now(),
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/get/binary/bin123", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		require.NoError(t, json.NewEncoder(w).Encode(expected))
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	result, err := GetBinaryHTTP(ctx, client, "test-token", "bin123")
	require.NoError(t, err)

	assert.Equal(t, expected.Item.SecretName, result.SecretName)
	assert.Equal(t, expected.Item.SecretOwner, result.SecretOwner)
	assert.Equal(t, expected.Item.Data, result.Data)
	assert.Equal(t, *expected.Item.Meta, *result.Meta)
	assert.WithinDuration(t, expected.Item.UpdatedAt, result.UpdatedAt, time.Second)
}

type stubBinaryClient struct{}

func (s *stubBinaryClient) Get(ctx context.Context, in *pb.BinaryGetRequest, opts ...grpc.CallOption) (*pb.BinaryGetResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer test-token" {
		return nil, fmt.Errorf("unauthorized")
	}

	meta := "grpc-meta"
	return &pb.BinaryGetResponse{
		SecretName:  "grpc-bin",
		SecretOwner: "grpc-user",
		Data:        []byte("grpc-binary"),
		Meta:        meta,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func TestGetBinaryGRPC(t *testing.T) {
	clientStub := &stubBinaryClient{}
	ctx := context.Background()

	result, err := GetBinaryGRPC(ctx, clientStub, "test-token", "grpc-bin")
	require.NoError(t, err)

	assert.Equal(t, "grpc-bin", result.SecretName)
	assert.Equal(t, "grpc-user", result.SecretOwner)
	assert.Equal(t, []byte("grpc-binary"), result.Data)
	require.NotNil(t, result.Meta)
	assert.Equal(t, "grpc-meta", *result.Meta)
	assert.False(t, result.UpdatedAt.IsZero())
}
