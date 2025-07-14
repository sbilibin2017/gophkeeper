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

// --- HTTP тест ---

func TestSecretUsernamePasswordListHTTPFacade_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/list/secret-username-password", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		secrets := []models.SecretUsernamePasswordClient{
			{
				SecretName: "login1",
				Username:   "user1",
				Password:   "pass1",
				Meta:       nil,
				UpdatedAt:  time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
	facade := NewSecretUsernamePasswordListHTTPFacade(client) // правильное имя конструктора

	secrets, err := facade.List(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "login1", secrets[0].SecretName)
	assert.Equal(t, "user1", secrets[0].Username)
	assert.Equal(t, "pass1", secrets[0].Password)
}

// --- gRPC мок сервер ---

type mockUsernamePasswordServer struct {
	pb.UnimplementedSecretUsernamePasswordServiceServer
}

func (s *mockUsernamePasswordServer) List(ctx context.Context, req *pb.SecretUsernamePasswordListRequest) (*pb.SecretUsernamePasswordListResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}
	auth := md.Get("authorization")
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretUsernamePasswordListResponse{
		Items: []*pb.SecretUsernamePassword{
			{
				SecretName: "login1",
				Username:   "user1",
				Password:   "pass1",
				Meta:       "",
				UpdatedAt:  time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func TestSecretUsernamePasswordListGRPCFacade_List(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterSecretUsernamePasswordServiceServer(server, &mockUsernamePasswordServer{})

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

	client := pb.NewSecretUsernamePasswordServiceClient(conn)
	facade := NewSecretUsernamePasswordListGRPCFacade(client)

	secrets, err := facade.List(ctx, "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "login1", secrets[0].SecretName)
	assert.Equal(t, "user1", secrets[0].Username)
	assert.Equal(t, "pass1", secrets[0].Password)
}

// --- gRPC мок сервер с ошибочным UpdatedAt ---

type badUsernamePasswordServer struct {
	pb.UnimplementedSecretUsernamePasswordServiceServer
}

func (s *badUsernamePasswordServer) List(ctx context.Context, req *pb.SecretUsernamePasswordListRequest) (*pb.SecretUsernamePasswordListResponse, error) {
	return &pb.SecretUsernamePasswordListResponse{
		Items: []*pb.SecretUsernamePassword{
			{
				SecretName: "login1",
				Username:   "user1",
				Password:   "pass1",
				Meta:       "",
				UpdatedAt:  "bad-format",
			},
		},
	}, nil
}

func (s *mockUsernamePasswordServer) Save(ctx context.Context, req *pb.SecretUsernamePasswordSaveRequest) (*pb.SecretUsernamePasswordSaveResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}
	auth := md.Get("authorization")
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	// Минимальная проверка для демонстрации
	if req.Secret.SecretName == "" || req.Secret.Username == "" {
		return nil, errors.New("missing required fields")
	}

	return &pb.SecretUsernamePasswordSaveResponse{}, nil
}

func (s *mockUsernamePasswordServer) Get(ctx context.Context, req *pb.SecretUsernamePasswordGetRequest) (*pb.SecretUsernamePasswordGetResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}
	auth := md.Get("authorization")
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretUsernamePasswordGetResponse{
		Secret: &pb.SecretUsernamePassword{
			SecretName: req.SecretName,
			Username:   "admin",
			Password:   "pass123",
			Meta:       "example-meta",
			UpdatedAt:  time.Now().Format(time.RFC3339),
		},
	}, nil
}

func TestSecretUsernamePasswordListGRPCFacade_List_InvalidUpdatedAt(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterSecretUsernamePasswordServiceServer(server, &badUsernamePasswordServer{})

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

	client := pb.NewSecretUsernamePasswordServiceClient(conn)
	facade := NewSecretUsernamePasswordListGRPCFacade(client)

	_, err = facade.List(ctx, "test-token")
	assert.Error(t, err)
	assert.Equal(t, "invalid updated_at format in response", err.Error())
}

// --- Setup gRPC ---
func setupGRPCUsernamePassword(t *testing.T) *grpc.ClientConn {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretUsernamePasswordServiceServer(server, &mockUsernamePasswordServer{})
	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(func() { server.Stop() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return conn
}

// --- Тест на сохранение через gRPC ---
func TestSecretUsernamePasswordSaveGRPCFacade_Save(t *testing.T) {
	conn := setupGRPCUsernamePassword(t)
	client := pb.NewSecretUsernamePasswordServiceClient(conn)
	facade := NewSecretUsernamePasswordSaveGRPCFacade(client)

	now := time.Now()
	meta := "example-meta"

	err := facade.Save(context.Background(), "test-token", models.SecretUsernamePasswordClient{
		SecretName: "my-secret",
		Username:   "admin",
		Password:   "pass123",
		Meta:       &meta,
		UpdatedAt:  now,
	})
	assert.NoError(t, err)
}

// --- Тест на получение через gRPC ---
func TestSecretUsernamePasswordGetGRPCFacade_Get(t *testing.T) {
	conn := setupGRPCUsernamePassword(t)
	client := pb.NewSecretUsernamePasswordServiceClient(conn)
	facade := NewSecretUsernamePasswordGetGRPCFacade(client)

	result, err := facade.Get(context.Background(), "test-token", "my-secret")
	assert.NoError(t, err)
	assert.Equal(t, "my-secret", result.SecretName)
	assert.Equal(t, "admin", result.Username)
	assert.Equal(t, "pass123", result.Password)
	assert.NotNil(t, result.Meta)
	assert.Equal(t, "example-meta", *result.Meta)
}

func TestSecretUsernamePasswordSaveHTTPFacade_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/save/secret-username-password", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			var body map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&body)
			assert.NoError(t, err)

			assert.Equal(t, "my-secret", body["secret_name"])
			assert.Equal(t, "admin", body["username"])
			assert.Equal(t, "pass123", body["password"])
			assert.Equal(t, "example-meta", body["meta"])
			assert.NotEmpty(t, body["updated_at"])

			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
		facade := NewSecretUsernamePasswordSaveHTTPFacade(client)

		meta := "example-meta"
		err := facade.Save(context.Background(), "test-token", models.SecretUsernamePasswordClient{
			SecretName: "my-secret",
			Username:   "admin",
			Password:   "pass123",
			Meta:       &meta,
			UpdatedAt:  time.Now(),
		})
		assert.NoError(t, err)
	})

	t.Run("server error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}))
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL)
		facade := NewSecretUsernamePasswordSaveHTTPFacade(client)

		err := facade.Save(context.Background(), "token", models.SecretUsernamePasswordClient{
			SecretName: "fail-secret",
			Username:   "fail",
			Password:   "123",
			UpdatedAt:  time.Now(),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server error")
	})
}

// Вспомогательная функция для указателя на строку
func strPtr(s string) *string {
	return &s
}

func TestSecretUsernamePasswordGetHTTPFacade_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := models.SecretUsernamePasswordClient{
			SecretName: "test-secret",
			Username:   "admin",
			Password:   "password123",
			Meta:       strPtr("meta-info"),
			UpdatedAt:  time.Now(),
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/get/secret-username-password/test-secret", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(expected)
		}))
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL)
		facade := NewSecretUsernamePasswordGetHTTPFacade(client)

		result, err := facade.Get(context.Background(), "test-token", "test-secret")
		assert.NoError(t, err)
		assert.Equal(t, expected.SecretName, result.SecretName)
		assert.Equal(t, expected.Username, result.Username)
		assert.Equal(t, expected.Password, result.Password)
		assert.NotNil(t, result.Meta)
		assert.Equal(t, *expected.Meta, *result.Meta)
	})

	t.Run("not found", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL)
		facade := NewSecretUsernamePasswordGetHTTPFacade(client)

		_, err := facade.Get(context.Background(), "test-token", "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error fetching username-password secret")
	})

	t.Run("connection error", func(t *testing.T) {
		client := resty.New().SetBaseURL("http://127.0.0.1:0") // Некорректный порт
		facade := NewSecretUsernamePasswordGetHTTPFacade(client)

		_, err := facade.Get(context.Background(), "any-token", "any-name")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server unavailable")
	})
}
