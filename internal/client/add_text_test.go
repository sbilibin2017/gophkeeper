package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// ---------- Test for AddTextHTTP ----------

func TestAddTextHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/add/text", r.URL.Path)

		// Optionally decode and validate JSON body here.

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	meta := "http-meta"
	req := models.TextAddRequest{
		SecretName: "text_123",
		Content:    "some text content",
		Meta:       &meta,
	}

	err := AddTextHTTP(ctx, client, "test-token", req)
	require.NoError(t, err)
}

func TestAddTextHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	req := models.TextAddRequest{
		SecretName: "text_123",
		Content:    "some text",
	}

	err := AddTextHTTP(ctx, client, "bad-token", req)
	assert.Error(t, err)
}

// ---------- Test for AddTextGRPC ----------

type stubTextAddClient struct{}

func (s *stubTextAddClient) Add(
	ctx context.Context,
	in *pb.TextAddRequest,
	opts ...grpc.CallOption,
) (*emptypb.Empty, error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	auth := md["authorization"]
	if len(auth) != 1 || auth[0] != "Bearer grpc-token" {
		return nil, fmt.Errorf("unauthorized")
	}

	if in.SecretName == "" {
		return nil, fmt.Errorf("secret_name required")
	}

	return &emptypb.Empty{}, nil
}

func TestAddTextGRPC(t *testing.T) {
	client := &stubTextAddClient{}
	ctx := context.Background()

	meta := "grpc-meta"
	req := models.TextAddRequest{
		SecretName: "grpc-text",
		Content:    "grpc text content",
		Meta:       &meta,
	}

	err := AddTextGRPC(ctx, client, "grpc-token", req)
	require.NoError(t, err)
}

func TestAddTextGRPC_Unauthorized(t *testing.T) {
	client := &stubTextAddClient{}
	ctx := context.Background()

	req := models.TextAddRequest{
		SecretName: "grpc-text",
		Content:    "some content",
	}

	err := AddTextGRPC(ctx, client, "bad-token", req)
	assert.Error(t, err)
}

func TestAddTextGRPC_ValidationError(t *testing.T) {
	client := &stubTextAddClient{}
	ctx := context.Background()

	req := models.TextAddRequest{
		SecretName: "",
		Content:    "some content",
	}

	err := AddTextGRPC(ctx, client, "grpc-token", req)
	assert.Error(t, err)
}
