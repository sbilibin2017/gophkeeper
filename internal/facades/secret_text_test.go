package facades

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTP тест с авторизацией ---

func TestSecretTextListFacade_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/list/secret-text", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		secrets := []models.SecretTextClient{
			{
				SecretName: "text1",
				Content:    "Hello, world!",
				Meta:       nil,
				UpdatedAt:  time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
	facade := NewTextListFacade(client)

	secrets, err := facade.List(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "text1", secrets[0].SecretName)
	assert.Equal(t, "Hello, world!", secrets[0].Content)
}

// --- gRPC мок сервер ---
// Теперь метод называется List (не ListTextSecrets), и токен приходит в metadata, не в запросе

type mockTextServiceServer struct {
	pb.UnimplementedSecretTextServiceServer
}

func (s *mockTextServiceServer) List(ctx context.Context, req *pb.SecretTextListRequest) (*pb.SecretTextListResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretTextListResponse{
		Items: []*pb.SecretText{
			{
				SecretName: "text1",
				Content:    "Hello, world!",
				Meta:       "",
				UpdatedAt:  time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func TestSecretTextListGRPCFacade_List(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterSecretTextServiceServer(server, &mockTextServiceServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretTextServiceClient(conn)
	facade := NewTextListGRPCFacade(client)

	secrets, err := facade.List(ctx, "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "text1", secrets[0].SecretName)
	assert.Equal(t, "Hello, world!", secrets[0].Content)
}

// --- gRPC мок сервер с ошибочным UpdatedAt ---

type badTextServer struct {
	pb.UnimplementedSecretTextServiceServer
}

func (s *badTextServer) List(ctx context.Context, req *pb.SecretTextListRequest) (*pb.SecretTextListResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretTextListResponse{
		Items: []*pb.SecretText{
			{
				SecretName: "text1",
				Content:    "Hello, world!",
				Meta:       "",
				UpdatedAt:  "bad-format",
			},
		},
	}, nil
}

func TestSecretTextListGRPCFacade_List_InvalidUpdatedAt(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterSecretTextServiceServer(server, &badTextServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretTextServiceClient(conn)
	facade := NewTextListGRPCFacade(client)

	_, err = facade.List(ctx, "test-token")
	assert.Error(t, err)
	assert.Equal(t, "invalid updated_at format in response", err.Error())
}

func TestSecretTextSaveHTTPFacade_Save(t *testing.T) {
	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/save/secret-text", r.URL.Path)
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer token", auth)

		err := json.NewDecoder(r.Body).Decode(&receivedBody)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
	facade := NewSecretTextSaveHTTPFacade(client)

	meta := "some-meta"
	secret := models.SecretTextClient{
		SecretName: "text1",
		Content:    "Hello, world!",
		Meta:       &meta,
		UpdatedAt:  time.Now(),
	}

	err := facade.Save(context.Background(), "token", secret)
	assert.NoError(t, err)
	assert.Equal(t, secret.SecretName, receivedBody["secret_name"])
	assert.Equal(t, secret.Content, receivedBody["content"])
	assert.Equal(t, *secret.Meta, receivedBody["meta"])
}

func TestSecretTextGetHTTPFacade_Get(t *testing.T) {
	expected := models.SecretTextClient{
		SecretName: "text1",
		Content:    "Hello, world!",
		UpdatedAt:  time.Now().Truncate(time.Second),
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/get/secret-text/text1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretTextGetHTTPFacade(client)

	secret, err := facade.Get(context.Background(), "token", "text1")
	assert.NoError(t, err)
	assert.Equal(t, expected.SecretName, secret.SecretName)
	assert.Equal(t, expected.Content, secret.Content)
	assert.WithinDuration(t, expected.UpdatedAt, secret.UpdatedAt, time.Second)
}

type mockSecretTextServer struct {
	pb.UnimplementedSecretTextServiceServer
}

func (s *mockSecretTextServer) Get(ctx context.Context, req *pb.SecretTextGetRequest) (*pb.SecretTextGetResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretTextGetResponse{
		Secret: &pb.SecretText{
			SecretName: req.SecretName,
			Content:    "my secret",
			Meta:       "meta-text",
			UpdatedAt:  time.Now().Format(time.RFC3339),
		},
	}, nil
}

func (s *mockSecretTextServer) Save(ctx context.Context, req *pb.SecretTextSaveRequest) (*pb.SecretTextSaveResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}
	if req.Secret.SecretName == "" || req.Secret.Content == "" {
		return nil, errors.New("invalid input")
	}
	return &pb.SecretTextSaveResponse{}, nil
}

// --- helper ---

func setupGRPCServerSecretText(t *testing.T) *grpc.ClientConn {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretTextServiceServer(server, &mockSecretTextServer{})
	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(func() { server.Stop() })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

	return conn
}

// --- tests ---

func TestSecretTextGetGRPCFacade_Get(t *testing.T) {
	conn := setupGRPCServerSecretText(t)
	client := pb.NewSecretTextServiceClient(conn)
	facade := NewSecretTextGetGRPCFacade(client)

	secret, err := facade.Get(context.Background(), "test-token", "my-secret")
	assert.NoError(t, err)
	assert.Equal(t, "my-secret", secret.SecretName)
	assert.Equal(t, "my secret", secret.Content)
	assert.NotNil(t, secret.Meta)
	assert.Equal(t, "meta-text", *secret.Meta)
	assert.WithinDuration(t, time.Now(), secret.UpdatedAt, 2*time.Second)
}

func TestSecretTextSaveGRPCFacade_Save(t *testing.T) {
	conn := setupGRPCServerSecretText(t)
	client := pb.NewSecretTextServiceClient(conn)
	facade := NewSecretTextSaveGRPCFacade(client)

	updatedAt := time.Now()

	// Без meta
	secret := models.SecretTextClient{
		SecretName: "my-secret",
		Content:    "top secret",
		UpdatedAt:  updatedAt,
	}
	err := facade.Save(context.Background(), "test-token", secret)
	assert.NoError(t, err)

	// С meta
	metaStr := "info"
	secret.Meta = &metaStr
	err = facade.Save(context.Background(), "test-token", secret)
	assert.NoError(t, err)
}
