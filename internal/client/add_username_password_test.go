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

// ---------- Test for AddUsernamePasswordHTTP ----------

func TestAddUsernamePasswordHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/add/username-password", r.URL.Path) // Corrected path here

		// Optionally decode and check the JSON body here if needed.

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	meta := "http-meta"
	req := models.UsernamePasswordAddRequest{
		SecretName: "login_abc",
		Username:   "user_http",
		Password:   "pass_http",
		Meta:       &meta,
	}

	err := AddUsernamePasswordHTTP(ctx, client, "test-token", req)
	require.NoError(t, err)
}

func TestAddUsernamePasswordHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	req := models.UsernamePasswordAddRequest{
		SecretName: "login_err",
		Username:   "user",
		Password:   "pass",
	}

	err := AddUsernamePasswordHTTP(ctx, client, "bad-token", req)
	assert.Error(t, err)
}

// ---------- Test for AddUsernamePasswordGRPC ----------

type stubUsernamePasswordAddClient struct{}

func (s *stubUsernamePasswordAddClient) Add(
	ctx context.Context,
	in *pb.UsernamePasswordAddRequest,
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

func TestAddUsernamePasswordGRPC(t *testing.T) {
	client := &stubUsernamePasswordAddClient{}
	ctx := context.Background()

	meta := "grpc-meta"
	req := models.UsernamePasswordAddRequest{
		SecretName: "grpc-login",
		Username:   "grpcuser",
		Password:   "grpcpass",
		Meta:       &meta,
	}

	err := AddUsernamePasswordGRPC(ctx, client, "grpc-token", req)
	require.NoError(t, err)
}

func TestAddUsernamePasswordGRPC_Unauthorized(t *testing.T) {
	client := &stubUsernamePasswordAddClient{}
	ctx := context.Background()

	req := models.UsernamePasswordAddRequest{
		SecretName: "grpc-login",
		Username:   "user",
		Password:   "pass",
	}

	err := AddUsernamePasswordGRPC(ctx, client, "bad-token", req)
	assert.Error(t, err)
}

func TestAddUsernamePasswordGRPC_ValidationError(t *testing.T) {
	client := &stubUsernamePasswordAddClient{}
	ctx := context.Background()

	req := models.UsernamePasswordAddRequest{
		SecretName: "",
		Username:   "user",
		Password:   "pass",
	}

	err := AddUsernamePasswordGRPC(ctx, client, "grpc-token", req)
	assert.Error(t, err)
}
