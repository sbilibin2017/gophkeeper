package facades

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// --- HTTP facade integration tests ---

func TestAuthHTTPFacade_RegisterAndLogin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/register" || r.URL.Path == "/login" {
			var body struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil || body.Username == "" || body.Password == "" {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			w.Header().Set("Authorization", "Bearer test-token-"+body.Username)
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/username" {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer test-token-") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			username := strings.TrimPrefix(auth, "Bearer test-token-")
			resp := struct {
				Username string `json:"username"`
			}{
				Username: username,
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewAuthHTTPFacade(client)

	ctx := context.Background()

	token, err := facade.Register(ctx, "user1", "pass1")
	assert.NoError(t, err)
	assert.Equal(t, "test-token-user1", token)

	token, err = facade.Login(ctx, "user2", "pass2")
	assert.NoError(t, err)
	assert.Equal(t, "test-token-user2", token)

	username, err := facade.GetUsername(ctx, "test-token-user2")
	assert.NoError(t, err)
	assert.Equal(t, "user2", username)
}

// --- gRPC facade integration test ---

type testAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *testAuthServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc-token-" + req.Username}, nil
}

func (s *testAuthServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	return &pb.AuthResponse{Token: "grpc-token-" + req.Username}, nil
}

func (s *testAuthServer) GetUsername(ctx context.Context, _ *emptypb.Empty) (*pb.GetUsernameResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata")
	}
	auth := md["authorization"]
	if len(auth) == 0 || !strings.HasPrefix(auth[0], "Bearer grpc-token-") {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	username := strings.TrimPrefix(auth[0], "Bearer grpc-token-")
	return &pb.GetUsernameResponse{Username: username}, nil
}

func TestAuthGRPCFacade_RegisterLoginGetUsername(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0") // random free port
	assert.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &testAuthServer{})

	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	facade := NewAuthGRPCFacade(conn) // Use your constructor here

	ctx := context.Background()

	token, err := facade.Register(ctx, "grpcuser1", "pass1")
	assert.NoError(t, err)
	assert.Equal(t, "grpc-token-grpcuser1", token)

	token, err = facade.Login(ctx, "grpcuser2", "pass2")
	assert.NoError(t, err)
	assert.Equal(t, "grpc-token-grpcuser2", token)

	username, err := facade.GetUsername(ctx, "grpc-token-grpcuser2")
	assert.NoError(t, err)
	assert.Equal(t, "grpcuser2", username)
}
