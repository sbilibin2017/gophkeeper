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

func TestSaveSecretTextHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/save/secret-text" || r.Method != http.MethodPost {
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
		require.NotEmpty(t, body["content"])
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)
	meta := "meta info"
	secret := models.SecretTextSaveRequest{
		SecretName: "text1",
		Content:    "some secret content",
		Meta:       &meta,
	}

	err := SaveSecretTextHTTP(context.Background(), clientHTTP, "testtoken", secret)
	assert.NoError(t, err)
}

func TestGetSecretTextHTTP(t *testing.T) {
	meta := "meta info"
	secretResponse := models.SecretTextGetResponse{
		SecretName:  "text1",
		SecretOwner: "John",
		Content:     "secret content",
		Meta:        &meta,
		UpdatedAt:   nil,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/get/secret-text/text1" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		respBytes, _ := json.Marshal(secretResponse)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)

	secret, err := GetSecretTextHTTP(context.Background(), clientHTTP, "testtoken", "text1")
	require.NoError(t, err)
	assert.Equal(t, secretResponse.SecretName, secret.SecretName)
	assert.Equal(t, secretResponse.SecretOwner, secret.SecretOwner)
	assert.Equal(t, secretResponse.Content, secret.Content)
	assert.Equal(t, secretResponse.Meta, secret.Meta)
}

func TestListSecretTextHTTP(t *testing.T) {
	meta := "meta info"
	listResponse := []models.SecretTextGetResponse{
		{SecretName: "text1", SecretOwner: "John", Content: "content1", Meta: &meta},
		{SecretName: "text2", SecretOwner: "Jane", Content: "content2", Meta: &meta},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/secret-text" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		respBytes, _ := json.Marshal(listResponse)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)

	secrets, err := ListSecretTextHTTP(context.Background(), clientHTTP, "testtoken")
	require.NoError(t, err)
	require.Len(t, secrets, 2)
	assert.Equal(t, "text1", secrets[0].SecretName)
	assert.Equal(t, "text2", secrets[1].SecretName)
}

// --- gRPC mock service ---

type mockSecretTextService struct {
	pb.UnimplementedSecretTextServiceServer
}

func (m *mockSecretTextService) Save(ctx context.Context, req *pb.SecretTextSaveRequest) (*pb.SecretTextSaveResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	auth := md.Get("authorization")
	if len(auth) == 0 || auth[0] != "Bearer testtoken" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return &pb.SecretTextSaveResponse{}, nil
}

func (m *mockSecretTextService) Get(ctx context.Context, req *pb.SecretTextGetRequest) (*pb.SecretTextGetResponse, error) {
	return &pb.SecretTextGetResponse{
		SecretName:  req.SecretName,
		SecretOwner: "John Doe",
		Content:     "some secret content",
		Meta:        "meta info",
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func (m *mockSecretTextService) List(ctx context.Context, req *pb.SecretTextListRequest) (*pb.SecretTextListResponse, error) {
	return &pb.SecretTextListResponse{
		Items: []*pb.SecretTextGetResponse{
			{
				SecretName:  "text1",
				SecretOwner: "John Doe",
				Content:     "some secret content",
				Meta:        "meta info",
				UpdatedAt:   time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func startMockGRPCTextServer(t *testing.T) (pb.SecretTextServiceClient, func()) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pb.RegisterSecretTextServiceServer(server, &mockSecretTextService{})

	go server.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)

	client := pb.NewSecretTextServiceClient(conn)

	cleanup := func() {
		server.Stop()
		conn.Close()
	}

	return client, cleanup
}

// --- gRPC client tests ---

func TestSaveSecretTextGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCTextServer(t)
	defer cleanup()

	meta := "meta info"
	secret := models.SecretTextSaveRequest{
		SecretName: "text1",
		Content:    "some secret content",
		Meta:       &meta,
	}

	err := SaveSecretTextGRPC(context.Background(), clientGRPC, "testtoken", secret)
	assert.NoError(t, err)
}

func TestGetSecretTextGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCTextServer(t)
	defer cleanup()

	secret, err := GetSecretTextGRPC(context.Background(), clientGRPC, "testtoken", "text1")
	assert.NoError(t, err)
	assert.Equal(t, "text1", secret.SecretName)
	assert.Equal(t, "John Doe", secret.SecretOwner)
	assert.Equal(t, "some secret content", secret.Content)
	assert.Equal(t, "meta info", *secret.Meta)
}

func TestListSecretTextGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCTextServer(t)
	defer cleanup()

	secrets, err := ListSecretTextGRPC(context.Background(), clientGRPC, "testtoken")
	assert.NoError(t, err)
	assert.NotEmpty(t, secrets)
	assert.Equal(t, "text1", secrets[0].SecretName)
}

func secretTextDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_text_request (
		secret_name TEXT PRIMARY KEY,
		content TEXT,
		meta TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSaveAndGetSecretTextRequest(t *testing.T) {
	ctx := context.Background()
	db := secretTextDB(t)
	defer db.Close()

	metaJSON := `{"some":"metadata"}`
	secret := models.SecretTextSaveRequest{
		SecretName: "text1",
		Content:    "Hello, world!",
		Meta:       &metaJSON,
	}

	// Сохраняем текстовый секрет
	err := SaveSecretTextRequest(ctx, db, secret)
	require.NoError(t, err)

	// Получаем список всех текстовых секретов (только имена)
	secrets, err := GetAllSecretsTextRequest(ctx, db)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "text1", secrets[0].SecretName)

	// Получаем текстовый секрет по имени
	secretResp, err := GetSecretTextByNameRequest(ctx, db, "text1")
	require.NoError(t, err)
	assert.Equal(t, secret.SecretName, secretResp.SecretName)
	assert.Equal(t, secret.Content, secretResp.Content)
	assert.NotNil(t, secretResp.Meta)
	assert.Equal(t, metaJSON, *secretResp.Meta)

	// Обновляем текстовый секрет с другим контентом и meta
	newMetaJSON := `{"updated":"data"}`
	secret.Content = "Updated content"
	secret.Meta = &newMetaJSON

	err = SaveSecretTextRequest(ctx, db, secret)
	require.NoError(t, err)

	updatedSecretResp, err := GetSecretTextByNameRequest(ctx, db, "text1")
	require.NoError(t, err)
	assert.Equal(t, "Updated content", updatedSecretResp.Content)
	assert.NotNil(t, updatedSecretResp.Meta)
	assert.Equal(t, newMetaJSON, *updatedSecretResp.Meta)
}

func TestGetSecretTextByNameRequest_NotFound(t *testing.T) {
	ctx := context.Background()
	db := secretTextDB(t)
	defer db.Close()

	secretResp, err := GetSecretTextByNameRequest(ctx, db, "missing_text")
	assert.Nil(t, secretResp)
	assert.Error(t, err)
	assert.Equal(t, "secret not found or error fetching", err.Error())
}
