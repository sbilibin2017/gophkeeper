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

func TestRegisterHTTP_Success(t *testing.T) {
	// Start a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register" {
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

	err := RegisterHTTP(context.Background(), user,
		WithRegisterHTTPClient(client),
		WithRegisterHTTPEncoders(nil),
	)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRegisterHTTP_ServerError(t *testing.T) {
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

	err := RegisterHTTP(context.Background(), user,
		WithRegisterHTTPClient(client),
		WithRegisterHTTPEncoders(nil),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type testRegisterServer struct {
	pb.UnimplementedRegisterServiceServer
}

func (s *testRegisterServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return &pb.RegisterResponse{Error: "missing fields"}, nil
	}
	return &pb.RegisterResponse{}, nil
}

func TestRegisterGRPC_Success(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0") // random available port
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRegisterServiceServer(s, &testRegisterServer{})

	go s.Serve(lis)
	defer s.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)

	user := &models.User{
		Username: "testuser",
		Password: "testpass",
	}

	err = RegisterGRPC(context.Background(), user,
		WithRegisterGRPCClient(client),
		WithRegisterGRPCEncoders(nil), // no encoding for simplicity
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegisterGRPC_ErrorResponse(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRegisterServiceServer(s, &testRegisterServer{})

	go s.Serve(lis)
	defer s.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)

	user := &models.User{
		Username: "", // missing username triggers error
		Password: "pass",
	}

	err = RegisterGRPC(context.Background(), user,
		WithRegisterGRPCClient(client),
		WithRegisterGRPCEncoders(nil),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
