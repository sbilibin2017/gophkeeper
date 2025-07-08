package services

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

// --- HTTP tests for Login ---

func TestLoginHTTP_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		if body["username"] == "" || body["password"] == "" {
			t.Errorf("missing username or password")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New()
	client.SetBaseURL(server.URL)

	user := &models.User{
		Username: "testuser",
		Password: "testpass",
	}

	err := LoginHTTP(context.Background(), user,
		WithLoginHTTPClient(client),
		WithLoginHTTPEncoders(nil),
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestLoginHTTP_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := resty.New()
	client.SetBaseURL(server.URL)

	user := &models.User{
		Username: "testuser",
		Password: "testpass",
	}

	err := LoginHTTP(context.Background(), user,
		WithLoginHTTPClient(client),
		WithLoginHTTPEncoders(nil),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- gRPC server implementation for Login ---

type testLoginServer struct {
	pb.UnimplementedLoginServiceServer
}

func (s *testLoginServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return &pb.LoginResponse{Error: "missing fields"}, nil
	}
	return &pb.LoginResponse{}, nil
}

// --- gRPC tests for Login ---

func TestLoginGRPC_Success(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLoginServiceServer(s, &testLoginServer{})

	go s.Serve(lis)
	defer s.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)

	user := &models.User{
		Username: "testuser",
		Password: "testpass",
	}

	err = LoginGRPC(context.Background(), user,
		WithLoginGRPCClient(client),
		WithLoginGRPCEncoders(nil),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestLoginGRPC_ErrorResponse(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLoginServiceServer(s, &testLoginServer{})

	go s.Serve(lis)
	defer s.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)

	user := &models.User{
		Username: "", // missing username triggers error
		Password: "pass",
	}

	err = LoginGRPC(context.Background(), user,
		WithLoginGRPCClient(client),
		WithLoginGRPCEncoders(nil),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
