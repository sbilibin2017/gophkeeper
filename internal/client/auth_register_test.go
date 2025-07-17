package client

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

const registerBufSize = 1024 * 1024

var registerLis *bufconn.Listener

type registerTestServer struct {
	pb.UnimplementedRegisterServiceServer
	ReceivedReq *pb.RegisterRequest
}

// Fix: Return *emptypb.Empty (matches proto)
func (s *registerTestServer) Register(ctx context.Context, req *pb.RegisterRequest) (*emptypb.Empty, error) {
	s.ReceivedReq = req
	return &emptypb.Empty{}, nil
}

func registerBufDialer(context.Context, string) (net.Conn, error) {
	return registerLis.Dial()
}

func registerSetupGRPCServer(t *testing.T) (pb.RegisterServiceClient, *registerTestServer, func()) {
	registerLis = bufconn.Listen(registerBufSize)
	s := grpc.NewServer()
	serverImpl := &registerTestServer{}
	pb.RegisterRegisterServiceServer(s, serverImpl)

	go func() {
		if err := s.Serve(registerLis); err != nil {
			t.Logf("gRPC server exited: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(registerBufDialer), grpc.WithInsecure())
	require.NoError(t, err)

	client := pb.NewRegisterServiceClient(conn)
	cleanup := func() {
		conn.Close()
		s.Stop()
	}

	return client, serverImpl, cleanup
}

func TestRegisterRegisterHTTP_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/register", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`OK`))
	}))
	defer ts.Close()

	restyClient := resty.New().SetBaseURL(ts.URL)

	req := &models.RegisterRequest{
		Username: "validuser",
		Password: "StrongPass1",
	}

	err := RegisterHTTP(context.Background(), restyClient, req)
	assert.NoError(t, err)
}

func TestRegisterRegisterHTTP_FailureStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	defer ts.Close()

	restyClient := resty.New().SetBaseURL(ts.URL)

	req := &models.RegisterRequest{
		Username: "validuser",
		Password: "StrongPass1",
	}

	err := RegisterHTTP(context.Background(), restyClient, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registration failed with status 403")
}

func TestRegisterRegisterHTTP_RequestError(t *testing.T) {
	restyClient := resty.New().SetBaseURL("http://invalid.localhost")

	req := &models.RegisterRequest{
		Username: "validuser",
		Password: "StrongPass1",
	}

	err := RegisterHTTP(context.Background(), restyClient, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send HTTP register request")
}

func TestRegisterRegisterGRPC_Success(t *testing.T) {
	client, serverImpl, cleanup := registerSetupGRPCServer(t)
	defer cleanup()

	req := &models.RegisterRequest{
		Username: "grpcuser",
		Password: "StrongPass1",
	}

	err := RegisterGRPC(context.Background(), client, req)
	assert.NoError(t, err)
	assert.NotNil(t, serverImpl.ReceivedReq)
	assert.Equal(t, req.Username, serverImpl.ReceivedReq.Username)
	assert.Equal(t, req.Password, serverImpl.ReceivedReq.Password)
}

func TestRegisterValidateRegisterUsername(t *testing.T) {
	tests := []struct {
		username string
		wantErr  bool
	}{
		{"abc", false},
		{"a1_b2", false},
		{"ab", true},
		{"a_very_long_username_more_than_30_chars", true},
		{"invalid-char!", true},
	}

	for _, tt := range tests {
		err := ValidateRegisterUsername(tt.username)
		if tt.wantErr {
			assert.Error(t, err, "username: %s", tt.username)
		} else {
			assert.NoError(t, err, "username: %s", tt.username)
		}
	}
}

func TestRegisterValidateRegisterPassword(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"Strong1A", false},
		{"weak", true},
		{"nouppercase1", true},
		{"NOLOWERCASE1", true},
		{"NoDigits", true},
	}

	for _, tt := range tests {
		err := ValidateRegisterPassword(tt.password)
		if tt.wantErr {
			assert.Error(t, err, "password: %s", tt.password)
		} else {
			assert.NoError(t, err, "password: %s", tt.password)
		}
	}
}
