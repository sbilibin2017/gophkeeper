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

// ---------- Test for AddBankCardHTTP ----------

func TestAddBankCardHTTP(t *testing.T) {
	// Setup test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "/add/bank-card", r.URL.Path)

		// Optional: check the JSON body
		// but skipping for brevity

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	meta := "meta-info"
	req := models.BankCardAddRequest{
		SecretName: "card123",
		Number:     "1234567890123456",
		Owner:      "John Doe",
		Exp:        "12/24",
		CVV:        "321",
		Meta:       &meta,
	}

	err := AddBankCardHTTP(ctx, client, "test-token", req)
	require.NoError(t, err)
}

func TestAddBankCardHTTP_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	ctx := context.Background()

	req := models.BankCardAddRequest{
		SecretName: "card123",
		Number:     "0000000000000000",
		Owner:      "John Doe",
		Exp:        "12/24",
		CVV:        "321",
	}

	err := AddBankCardHTTP(ctx, client, "bad-token", req)
	assert.Error(t, err)
}

// ---------- Test for AddBankCardGRPC ----------

// stub client implementing BankCardAddServiceClient
type stubBankCardAddClient struct{}

func (s *stubBankCardAddClient) Add(
	ctx context.Context,
	in *pb.BankCardAddRequest,
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

func TestAddBankCardGRPC(t *testing.T) {
	client := &stubBankCardAddClient{}
	ctx := context.Background()

	meta := "grpc-meta"
	req := models.BankCardAddRequest{
		SecretName: "grpc-card",
		Number:     "9999888877776666",
		Owner:      "Grpc User",
		Exp:        "11/26",
		CVV:        "123",
		Meta:       &meta,
	}

	err := AddBankCardGRPC(ctx, client, "grpc-token", req)
	require.NoError(t, err)
}

func TestAddBankCardGRPC_Unauthorized(t *testing.T) {
	client := &stubBankCardAddClient{}
	ctx := context.Background()

	req := models.BankCardAddRequest{
		SecretName: "grpc-card",
		Number:     "9999888877776666",
		Owner:      "Grpc User",
		Exp:        "11/26",
		CVV:        "123",
	}

	err := AddBankCardGRPC(ctx, client, "bad-token", req)
	assert.Error(t, err)
}

func TestAddBankCardGRPC_ValidationError(t *testing.T) {
	client := &stubBankCardAddClient{}
	ctx := context.Background()

	req := models.BankCardAddRequest{
		SecretName: "",
		Number:     "9999888877776666",
		Owner:      "Grpc User",
		Exp:        "11/26",
		CVV:        "123",
	}

	err := AddBankCardGRPC(ctx, client, "grpc-token", req)
	assert.Error(t, err)
}
