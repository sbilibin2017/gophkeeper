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

func TestLoginHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New()
	client.SetBaseURL(server.URL)

	// Call with required args first, then options to set client & encoders
	err := LoginHTTP(
		context.Background(),
		"user",
		"pass",
		WithLoginHTTPClient(client),
		// encoders can be omitted or set to nil explicitly if you want
	)
	assert.NoError(t, err)
}

type testLoginServer struct {
	pb.UnimplementedLoginServiceServer
	shouldFail bool
}

func (s *testLoginServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if s.shouldFail {
		return nil, errors.New("login failed")
	}
	return &pb.LoginResponse{}, nil
}

func TestLoginGRPC(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	testSrv := &testLoginServer{}
	pb.RegisterLoginServiceServer(srv, testSrv)

	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
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

	client := pb.NewLoginServiceClient(conn)

	t.Run("successful login", func(t *testing.T) {
		testSrv.shouldFail = false

		err := LoginGRPC(
			ctx,
			"grpcuser",
			"grpcpass",
			WithLoginGRPCClient(client),
			// encoders can be omitted or set explicitly if needed
		)
		assert.NoError(t, err)
	})

	t.Run("failed login", func(t *testing.T) {
		testSrv.shouldFail = true

		err := LoginGRPC(
			ctx,
			"grpcuser",
			"grpcpass",
			WithLoginGRPCClient(client),
		)
		assert.Error(t, err)
	})
}
