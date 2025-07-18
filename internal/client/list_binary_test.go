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
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	_ "modernc.org/sqlite"
)

func setupListBinaryLocalTestDB(t *testing.T) *sqlx.DB {
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

func TestListBinaryLocal(t *testing.T) {
	db := setupListBinaryLocalTestDB(t)
	defer db.Close()

	binaryData := []byte{0x01, 0x02, 0x03, 0x04}
	_, err := db.Exec(`
		INSERT INTO secret_binary_request (secret_name, data, meta)
		VALUES (?, ?, ?)`,
		"bin1", binaryData, "local-meta")
	require.NoError(t, err)

	ctx := context.Background()
	results, err := ListBinaryLocal(ctx, db)
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, "bin1", results[0].SecretName)
	assert.Equal(t, binaryData, results[0].Data)
	require.NotNil(t, results[0].Meta)
	assert.Equal(t, "local-meta", *results[0].Meta)
}

func TestListBinaryHTTP(t *testing.T) {
	meta := "http-meta"
	updatedAt := time.Date(2025, 7, 18, 0, 0, 0, 0, time.UTC)

	expected := struct {
		Items []models.BinaryResponse `json:"items"`
	}{
		Items: []models.BinaryResponse{
			{
				SecretName:  "bin_123",
				SecretOwner: "owner-xyz",
				Data:        []byte{0x0a, 0x0b, 0x0c},
				Meta:        &meta,
				UpdatedAt:   updatedAt,
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/list/binary", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(expected)
		require.NoError(t, err)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	results, err := ListBinaryHTTP(ctx, client, "test-token")
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, expected.Items[0].SecretName, results[0].SecretName)
	assert.Equal(t, expected.Items[0].Data, results[0].Data)
	require.NotNil(t, results[0].Meta)
	assert.Equal(t, *expected.Items[0].Meta, *results[0].Meta)
	assert.True(t, results[0].UpdatedAt.Equal(updatedAt))
}

// stubBinaryListClient implements pb.BinaryListServiceClient for testing
type stubBinaryListClient struct{}

func (s *stubBinaryListClient) List(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.BinaryListResponse, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer test-token" {
		return nil, fmt.Errorf("unauthorized")
	}

	meta := "grpc-meta"
	return &pb.BinaryListResponse{
		Items: []*pb.BinaryGetResponse{
			{
				SecretName: "bin_001",
				Data:       []byte{0x10, 0x20, 0x30},
				Meta:       meta,
				UpdatedAt:  "2025-07-18T15:04:05Z",
			},
			{
				SecretName: "bin_002",
				Data:       []byte{0x40, 0x50, 0x60},
				Meta:       "",
				UpdatedAt:  "invalid-time-format",
			},
		},
	}, nil
}

func TestListBinaryGRPC(t *testing.T) {
	clientStub := &stubBinaryListClient{}
	ctx := context.Background()

	result, err := ListBinaryGRPC(ctx, clientStub, "test-token")
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, "bin_001", result[0].SecretName)
	assert.Equal(t, []byte{0x10, 0x20, 0x30}, result[0].Data)
	require.NotNil(t, result[0].Meta)
	assert.Equal(t, "grpc-meta", *result[0].Meta)

	// Valid updatedAt parsed
	expectedTime, _ := time.Parse(time.RFC3339, "2025-07-18T15:04:05Z")
	assert.True(t, result[0].UpdatedAt.Equal(expectedTime))

	assert.Equal(t, "bin_002", result[1].SecretName)
	assert.Equal(t, []byte{0x40, 0x50, 0x60}, result[1].Data)
	assert.Nil(t, result[1].Meta)

	// Invalid time format defaults to zero time
	assert.True(t, result[1].UpdatedAt.IsZero())
}
