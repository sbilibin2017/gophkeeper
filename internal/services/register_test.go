package services

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func TestRegisterHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New()
	client.SetBaseURL(server.URL)

	err := RegisterHTTP(
		context.Background(),
		"user",
		"pass",
		WithRegisterHTTPClient(client),
		// nil encoders can be omitted, or you can explicitly pass them:
		// WithHMACEncoder(nil),
		// WithRSAEncoder(nil),
	)
	assert.NoError(t, err)
}

// --- gRPC test setup ---

type testRegisterServer struct {
	pb.UnimplementedRegisterServiceServer
	shouldFail bool
}

func (s *testRegisterServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if s.shouldFail {
		return nil, errors.New("registration failed")
	}
	return &pb.RegisterResponse{Error: ""}, nil
}

func TestRegisterGRPC(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	testSrv := &testRegisterServer{}
	pb.RegisterRegisterServiceServer(srv, testSrv)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	defer srv.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)

	t.Run("successful registration", func(t *testing.T) {
		testSrv.shouldFail = false

		err := RegisterGRPC(
			ctx,
			"grpcuser",
			"grpcpass",
			WithRegisterGRPCClient(client),
			// omit encoders or explicitly pass nil:
			// WithGRPCHMACEncoder(nil),
			// WithGRPCRSAEncoder(nil),
		)
		assert.NoError(t, err)
	})

	t.Run("failed registration", func(t *testing.T) {
		testSrv.shouldFail = true

		err := RegisterGRPC(
			ctx,
			"grpcuser",
			"grpcpass",
			WithRegisterGRPCClient(client),
		)
		assert.Error(t, err)
	})
}
