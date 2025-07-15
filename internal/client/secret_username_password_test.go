package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTP tests ---

func TestSaveSecretUsernamePasswordHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/save/secret-username-password" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if auth := r.Header.Get("Authorization"); auth != "Bearer testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		require.NotEmpty(t, body["secret_name"])
		require.NotEmpty(t, body["username"])
		require.NotEmpty(t, body["password"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)
	meta := "{\"info\":\"meta\"}"
	secret := models.SecretUsernamePasswordSaveRequest{
		SecretName: "login1",
		Username:   "user1",
		Password:   "pass1",
		Meta:       &meta,
	}

	err := SaveSecretUsernamePasswordHTTP(context.Background(), clientHTTP, "testtoken", secret)
	assert.NoError(t, err)
}

func TestGetSecretUsernamePasswordHTTP(t *testing.T) {
	meta := "{\"info\":\"meta\"}"
	updatedAt := time.Now().Format(time.RFC3339)

	secretResponse := models.SecretUsernamePasswordGetResponse{
		SecretName:  "login1",
		SecretOwner: "John",
		Username:    "user1",
		Password:    "pass1",
		Meta:        &meta,
		UpdatedAt:   func() *time.Time { t, _ := time.Parse(time.RFC3339, updatedAt); return &t }(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/get/secret-username-password/login1" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		respMap := map[string]interface{}{
			"secret_name":  secretResponse.SecretName,
			"secret_owner": secretResponse.SecretOwner,
			"username":     secretResponse.Username,
			"password":     secretResponse.Password,
			"meta":         *secretResponse.Meta,
			"updated_at":   updatedAt,
		}

		respBytes, _ := json.Marshal(respMap)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)

	secret, err := GetSecretUsernamePasswordHTTP(context.Background(), clientHTTP, "testtoken", "login1")
	require.NoError(t, err)
	assert.Equal(t, secretResponse.SecretName, secret.SecretName)
	assert.Equal(t, secretResponse.SecretOwner, secret.SecretOwner)
	assert.Equal(t, secretResponse.Username, secret.Username)
	assert.Equal(t, secretResponse.Password, secret.Password)
	assert.Equal(t, *secretResponse.Meta, *secret.Meta)
	assert.NotNil(t, secret.UpdatedAt)
}

func TestListSecretUsernamePasswordHTTP(t *testing.T) {
	meta := "{\"info\":\"meta\"}"
	updatedAt := time.Now().Format(time.RFC3339)

	listResponse := []models.SecretUsernamePasswordGetResponse{
		{
			SecretName:  "login1",
			SecretOwner: "John",
			Username:    "user1",
			Password:    "pass1",
			Meta:        &meta,
			UpdatedAt:   func() *time.Time { t, _ := time.Parse(time.RFC3339, updatedAt); return &t }(),
		},
		{
			SecretName:  "login2",
			SecretOwner: "Jane",
			Username:    "user2",
			Password:    "pass2",
			Meta:        &meta,
			UpdatedAt:   func() *time.Time { t, _ := time.Parse(time.RFC3339, updatedAt); return &t }(),
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/secret-username-password" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var respItems []map[string]interface{}
		for _, item := range listResponse {
			respItems = append(respItems, map[string]interface{}{
				"secret_name":  item.SecretName,
				"secret_owner": item.SecretOwner,
				"username":     item.Username,
				"password":     item.Password,
				"meta":         *item.Meta,
				"updated_at":   updatedAt,
			})
		}

		respBytes, _ := json.Marshal(respItems)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)

	secrets, err := ListSecretUsernamePasswordHTTP(context.Background(), clientHTTP, "testtoken")
	require.NoError(t, err)
	require.Len(t, secrets, 2)
	assert.Equal(t, "login1", secrets[0].SecretName)
	assert.Equal(t, "login2", secrets[1].SecretName)
}

// --- gRPC mock service ---

type mockSecretUsernamePasswordService struct {
	pb.UnimplementedSecretUsernamePasswordServiceServer
}

func (m *mockSecretUsernamePasswordService) Save(ctx context.Context, req *pb.SecretUsernamePasswordSaveRequest) (*pb.SecretUsernamePasswordSaveResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	auth := md.Get("authorization")
	if len(auth) == 0 || auth[0] != "Bearer testtoken" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return &pb.SecretUsernamePasswordSaveResponse{}, nil
}

func (m *mockSecretUsernamePasswordService) Get(ctx context.Context, req *pb.SecretUsernamePasswordGetRequest) (*pb.SecretUsernamePasswordGetResponse, error) {
	return &pb.SecretUsernamePasswordGetResponse{
		SecretName:  req.SecretName,
		SecretOwner: "John Doe",
		Username:    "user1",
		Password:    "pass1",
		Meta:        "{\"info\":\"meta\"}",
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func (m *mockSecretUsernamePasswordService) List(ctx context.Context, req *pb.SecretUsernamePasswordListRequest) (*pb.SecretUsernamePasswordListResponse, error) {
	return &pb.SecretUsernamePasswordListResponse{
		Items: []*pb.SecretUsernamePasswordGetResponse{
			{
				SecretName:  "login1",
				SecretOwner: "John Doe",
				Username:    "user1",
				Password:    "pass1",
				Meta:        "{\"info\":\"meta\"}",
				UpdatedAt:   time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func startMockGRPCUsernamePasswordServer(t *testing.T) (pb.SecretUsernamePasswordServiceClient, func()) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pb.RegisterSecretUsernamePasswordServiceServer(server, &mockSecretUsernamePasswordService{})

	go server.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)

	client := pb.NewSecretUsernamePasswordServiceClient(conn)

	cleanup := func() {
		server.Stop()
		conn.Close()
	}

	return client, cleanup
}

// --- gRPC client tests ---

func TestSaveSecretUsernamePasswordGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCUsernamePasswordServer(t)
	defer cleanup()

	meta := "{\"info\":\"meta\"}"
	secret := models.SecretUsernamePasswordSaveRequest{
		SecretName: "login1",
		Username:   "user1",
		Password:   "pass1",
		Meta:       &meta,
	}

	err := SaveSecretUsernamePasswordGRPC(context.Background(), clientGRPC, "testtoken", secret)
	assert.NoError(t, err)
}

func TestGetSecretUsernamePasswordGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCUsernamePasswordServer(t)
	defer cleanup()

	secret, err := GetSecretUsernamePasswordGRPC(context.Background(), clientGRPC, "testtoken", "login1")
	assert.NoError(t, err)
	assert.Equal(t, "login1", secret.SecretName)
	assert.Equal(t, "John Doe", secret.SecretOwner)
	assert.Equal(t, "user1", secret.Username)
	assert.Equal(t, "pass1", secret.Password)
	assert.NotNil(t, secret.Meta)
}

func TestListSecretUsernamePasswordGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCUsernamePasswordServer(t)
	defer cleanup()

	secrets, err := ListSecretUsernamePasswordGRPC(context.Background(), clientGRPC, "testtoken")
	assert.NoError(t, err)
	assert.NotEmpty(t, secrets)
	assert.Equal(t, "login1", secrets[0].SecretName)
}

func secretUsernamePasswordDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_username_password_request (
		secret_name TEXT PRIMARY KEY,
		username TEXT,
		password TEXT,
		meta TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSaveAndGetSecretUsernamePasswordRequest(t *testing.T) {
	ctx := context.Background()
	db := secretUsernamePasswordDB(t)
	defer db.Close()

	metaJSON := `{"some":"metadata"}`
	secret := models.SecretUsernamePasswordSaveRequest{
		SecretName: "login1",
		Username:   "user123",
		Password:   "pass123",
		Meta:       &metaJSON,
	}

	// Сохраняем секрет с логином и паролем
	err := SaveSecretUsernamePasswordRequest(ctx, db, secret)
	require.NoError(t, err)

	// Получаем список всех секретов (только имена)
	secrets, err := GetAllSecretsUsernamePasswordRequest(ctx, db)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "login1", secrets[0].SecretName)

	// Получаем секрет по имени
	secretResp, err := GetSecretUsernamePasswordByNameRequest(ctx, db, "login1")
	require.NoError(t, err)
	assert.Equal(t, secret.SecretName, secretResp.SecretName)
	assert.Equal(t, secret.Username, secretResp.Username)
	assert.Equal(t, secret.Password, secretResp.Password)
	assert.NotNil(t, secretResp.Meta)
	assert.Equal(t, metaJSON, *secretResp.Meta)

	// Обновляем секрет с другим паролем и meta
	newMetaJSON := `{"updated":"data"}`
	secret.Password = "newpass456"
	secret.Meta = &newMetaJSON

	err = SaveSecretUsernamePasswordRequest(ctx, db, secret)
	require.NoError(t, err)

	updatedSecretResp, err := GetSecretUsernamePasswordByNameRequest(ctx, db, "login1")
	require.NoError(t, err)
	assert.Equal(t, "newpass456", updatedSecretResp.Password)
	assert.NotNil(t, updatedSecretResp.Meta)
	assert.Equal(t, newMetaJSON, *updatedSecretResp.Meta)
}

func TestGetSecretUsernamePasswordByNameRequest_NotFound(t *testing.T) {
	ctx := context.Background()
	db := secretUsernamePasswordDB(t)
	defer db.Close()

	secretResp, err := GetSecretUsernamePasswordByNameRequest(ctx, db, "missing_login")
	assert.Nil(t, secretResp)
	assert.Error(t, err)
	assert.Equal(t, "secret not found or error fetching", err.Error())
}
