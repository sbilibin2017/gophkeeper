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

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const logoutBufSize = 1024 * 1024

var logoutLis *bufconn.Listener

// Mock gRPC server for LogoutService
type logoutTestServer struct {
	pb.UnimplementedLogoutServiceServer
	ReceivedReq *pb.LogoutRequest
}

func (s *logoutTestServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	s.ReceivedReq = req
	return &emptypb.Empty{}, nil
}

func logoutBufDialer(context.Context, string) (net.Conn, error) {
	return logoutLis.Dial()
}

// Setup a bufconn gRPC server and client for Logout tests
func logoutSetupGRPCServer(t *testing.T) (pb.LogoutServiceClient, *logoutTestServer, func()) {
	logoutLis = bufconn.Listen(logoutBufSize)
	s := grpc.NewServer()
	serverImpl := &logoutTestServer{}
	pb.RegisterLogoutServiceServer(s, serverImpl)

	go func() {
		if err := s.Serve(logoutLis); err != nil {
			t.Logf("gRPC server exited: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(logoutBufDialer), grpc.WithInsecure())
	require.NoError(t, err)

	client := pb.NewLogoutServiceClient(conn)
	cleanup := func() {
		conn.Close()
		s.Stop()
	}

	return client, serverImpl, cleanup
}

func TestLogoutHTTP_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/logout", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`OK`))
	}))
	defer ts.Close()

	restyClient := resty.New().SetBaseURL(ts.URL)

	req := &models.LogoutRequest{
		Token: "valid-token",
	}

	err := LogoutHTTP(context.Background(), restyClient, req)
	assert.NoError(t, err)
}

func TestLogoutHTTP_FailureStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer ts.Close()

	restyClient := resty.New().SetBaseURL(ts.URL)

	req := &models.LogoutRequest{
		Token: "invalid-token",
	}

	err := LogoutHTTP(context.Background(), restyClient, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "logout failed: 401 Unauthorized")
}

func TestLogoutHTTP_RequestError(t *testing.T) {
	restyClient := resty.New().SetBaseURL("http://invalid.localhost")

	req := &models.LogoutRequest{
		Token: "any-token",
	}

	err := LogoutHTTP(context.Background(), restyClient, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect: connection refused")
}

func TestLogoutGRPC_Success(t *testing.T) {
	client, serverImpl, cleanup := logoutSetupGRPCServer(t)
	defer cleanup()

	req := &models.LogoutRequest{
		Token: "grpc-valid-token",
	}

	err := LogoutGRPC(context.Background(), client, req)
	assert.NoError(t, err)
	assert.NotNil(t, serverImpl.ReceivedReq)
	assert.Equal(t, "grpc-valid-token", serverImpl.ReceivedReq.Token)
}
