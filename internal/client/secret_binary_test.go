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

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// --- HTTP tests ---

func TestSaveSecretBinaryHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/save/secret-binary" || r.Method != http.MethodPost {
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
		require.NotEmpty(t, body["data"])
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	clientHTTP := resty.New().SetBaseURL(server.URL)
	meta := "{\"info\":\"meta\"}"
	secret := models.SecretBinarySaveRequest{
		SecretName: "binary1",
		Data:       []byte{0x01, 0x02, 0x03},
		Meta:       &meta,
	}

	err := SaveSecretBinaryHTTP(context.Background(), clientHTTP, "testtoken", secret)
	assert.NoError(t, err)
}

func TestGetSecretBinaryHTTP(t *testing.T) {
	meta := "{\"info\":\"meta\"}"
	secretResponse := models.SecretBinaryGetResponse{
		SecretName:  "binary1",
		SecretOwner: "John",
		Data:        []byte{0x01, 0x02, 0x03},
		Meta:        &meta,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/get/secret-binary/binary1" || r.Method != http.MethodGet {
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

	secret, err := GetSecretBinaryHTTP(context.Background(), clientHTTP, "testtoken", "binary1")
	require.NoError(t, err)
	assert.Equal(t, secretResponse.SecretName, secret.SecretName)
	assert.Equal(t, secretResponse.SecretOwner, secret.SecretOwner)
	assert.Equal(t, secretResponse.Data, secret.Data)
	assert.Equal(t, secretResponse.Meta, secret.Meta)
}

func TestListSecretBinaryHTTP(t *testing.T) {
	meta := "{\"info\":\"meta\"}"
	listResponse := []models.SecretBinaryGetResponse{
		{SecretName: "binary1", SecretOwner: "John", Data: []byte{0x01}, Meta: &meta},
		{SecretName: "binary2", SecretOwner: "Jane", Data: []byte{0x02}, Meta: &meta},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list/secret-binary" || r.Method != http.MethodGet {
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

	secrets, err := ListSecretBinaryHTTP(context.Background(), clientHTTP, "testtoken")
	require.NoError(t, err)
	require.Len(t, secrets, 2)
	assert.Equal(t, "binary1", secrets[0].SecretName)
	assert.Equal(t, "binary2", secrets[1].SecretName)
}

// --- gRPC mock service ---

type mockSecretBinaryService struct {
	pb.UnimplementedSecretBinaryServiceServer
}

func (m *mockSecretBinaryService) Save(ctx context.Context, req *pb.SecretBinarySaveRequest) (*pb.SecretBinarySaveResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	auth := md.Get("authorization")
	if len(auth) == 0 || auth[0] != "Bearer testtoken" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return &pb.SecretBinarySaveResponse{}, nil
}

func (m *mockSecretBinaryService) Get(ctx context.Context, req *pb.SecretBinaryGetRequest) (*pb.SecretBinaryGetResponse, error) {
	return &pb.SecretBinaryGetResponse{
		SecretName:  req.SecretName,
		SecretOwner: "John Doe",
		Data:        []byte{0x01, 0x02, 0x03},
		Meta:        "{\"info\":\"meta\"}",
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func (m *mockSecretBinaryService) List(ctx context.Context, req *pb.SecretBinaryListRequest) (*pb.SecretBinaryListResponse, error) {
	return &pb.SecretBinaryListResponse{
		Items: []*pb.SecretBinaryGetResponse{
			{
				SecretName:  "binary1",
				SecretOwner: "John Doe",
				Data:        []byte{0x01, 0x02, 0x03},
				Meta:        "{\"info\":\"meta\"}",
				UpdatedAt:   time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

func startMockGRPCBinaryServer(t *testing.T) (pb.SecretBinaryServiceClient, func()) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pb.RegisterSecretBinaryServiceServer(server, &mockSecretBinaryService{})

	go server.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)

	client := pb.NewSecretBinaryServiceClient(conn)

	cleanup := func() {
		server.Stop()
		conn.Close()
	}

	return client, cleanup
}

// --- gRPC client tests ---

func TestSaveSecretBinaryGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCBinaryServer(t)
	defer cleanup()

	meta := "{\"info\":\"meta\"}"
	secret := models.SecretBinarySaveRequest{
		SecretName: "binary1",
		Data:       []byte{0x01, 0x02, 0x03},
		Meta:       &meta,
	}

	err := SaveSecretBinaryGRPC(context.Background(), clientGRPC, "testtoken", secret)
	assert.NoError(t, err)
}

func TestGetSecretBinaryGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCBinaryServer(t)
	defer cleanup()

	secret, err := GetSecretBinaryGRPC(context.Background(), clientGRPC, "testtoken", "binary1")
	assert.NoError(t, err)
	assert.Equal(t, "binary1", secret.SecretName)
	assert.Equal(t, "John Doe", secret.SecretOwner)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, secret.Data)
}

func TestListSecretBinaryGRPC(t *testing.T) {
	clientGRPC, cleanup := startMockGRPCBinaryServer(t)
	defer cleanup()

	secrets, err := ListSecretBinaryGRPC(context.Background(), clientGRPC, "testtoken")
	assert.NoError(t, err)
	assert.NotEmpty(t, secrets)
	assert.Equal(t, "binary1", secrets[0].SecretName)
}

func secretBinaryDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_binary_request (
		secret_name TEXT PRIMARY KEY,
		data BLOB,
		meta TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSaveAndGetSecretBinaryRequest(t *testing.T) {
	ctx := context.Background()
	db := secretBinaryDB(t)
	defer db.Close()

	metaJSON := `{"some":"metadata"}`
	data := []byte{1, 2, 3, 4, 5}
	secret := models.SecretBinarySaveRequest{
		SecretName: "binary1",
		Data:       data,
		Meta:       &metaJSON,
	}

	// Сохраняем бинарный секрет
	err := SaveSecretBinaryRequest(ctx, db, secret)
	require.NoError(t, err)

	// Получаем список всех бинарных секретов (только имена)
	secrets, err := GetAllSecretsBinaryRequest(ctx, db)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "binary1", secrets[0].SecretName)

	// Получаем бинарный секрет по имени
	secretResp, err := GetSecretBinaryByNameRequest(ctx, db, "binary1")
	require.NoError(t, err)
	assert.Equal(t, secret.SecretName, secretResp.SecretName)
	assert.Equal(t, secret.Data, secretResp.Data)
	assert.NotNil(t, secretResp.Meta)
	assert.Equal(t, metaJSON, *secretResp.Meta)

	// Обновляем бинарный секрет с другими данными и meta
	newMetaJSON := `{"updated":"data"}`
	newData := []byte{9, 8, 7, 6}
	secret.Data = newData
	secret.Meta = &newMetaJSON

	err = SaveSecretBinaryRequest(ctx, db, secret)
	require.NoError(t, err)

	updatedSecretResp, err := GetSecretBinaryByNameRequest(ctx, db, "binary1")
	require.NoError(t, err)
	assert.Equal(t, newData, updatedSecretResp.Data)
	assert.NotNil(t, updatedSecretResp.Meta)
	assert.Equal(t, newMetaJSON, *updatedSecretResp.Meta)
}

func TestGetSecretBinaryByNameRequest_NotFound(t *testing.T) {
	ctx := context.Background()
	db := secretBinaryDB(t)
	defer db.Close()

	secretResp, err := GetSecretBinaryByNameRequest(ctx, db, "missing_binary")
	assert.Nil(t, secretResp)
	assert.Error(t, err)
	assert.Equal(t, "secret not found or error fetching", err.Error())
}
