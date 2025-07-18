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

// ---------- Test for AddBinaryHTTP ----------

func TestAddBinaryHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/add/binary", r.URL.Path)

		// Optionally: you could decode and check JSON body here

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	meta := "some-meta"
	req := models.BinaryAddRequest{
		SecretName: "bin001",
		Data:       []byte{0xAA, 0xBB, 0xCC},
		Meta:       &meta,
	}

	err := AddBinaryHTTP(ctx, client, "test-token", req)
	require.NoError(t, err)
}

func TestAddBinaryHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	req := models.BinaryAddRequest{
		SecretName: "bin001",
		Data:       []byte{0x00},
	}

	err := AddBinaryHTTP(ctx, client, "bad-token", req)
	assert.Error(t, err)
}

// ---------- Test for AddBinaryGRPC ----------

type stubBinaryAddClient struct{}

func (s *stubBinaryAddClient) Add(
	ctx context.Context,
	in *pb.BinaryAddRequest,
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

func TestAddBinaryGRPC(t *testing.T) {
	client := &stubBinaryAddClient{}
	ctx := context.Background()

	meta := "grpc-meta"
	req := models.BinaryAddRequest{
		SecretName: "grpc-bin",
		Data:       []byte{0x11, 0x22, 0x33},
		Meta:       &meta,
	}

	err := AddBinaryGRPC(ctx, client, "grpc-token", req)
	require.NoError(t, err)
}

func TestAddBinaryGRPC_Unauthorized(t *testing.T) {
	client := &stubBinaryAddClient{}
	ctx := context.Background()

	req := models.BinaryAddRequest{
		SecretName: "grpc-bin",
		Data:       []byte{0x11, 0x22, 0x33},
	}

	err := AddBinaryGRPC(ctx, client, "bad-token", req)
	assert.Error(t, err)
}

func TestAddBinaryGRPC_ValidationError(t *testing.T) {
	client := &stubBinaryAddClient{}
	ctx := context.Background()

	req := models.BinaryAddRequest{
		SecretName: "",
		Data:       []byte{0x11, 0x22, 0x33},
	}

	err := AddBinaryGRPC(ctx, client, "grpc-token", req)
	assert.Error(t, err)
}
