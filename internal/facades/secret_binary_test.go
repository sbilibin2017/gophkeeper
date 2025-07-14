package facades

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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

func TestSecretBinaryListFacade_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/list/secret-binary", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		secrets := []models.SecretBinaryClient{
			{
				SecretName: "binary1",
				Data:       []byte{0x01, 0x02, 0x03},
				Meta:       nil,
				UpdatedAt:  time.Now(),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetTimeout(5 * time.Second)
	facade := NewSecretBinaryListFacade(client)

	secrets, err := facade.List(context.Background(), "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "binary1", secrets[0].SecretName)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, secrets[0].Data)
}

// --- gRPC mock server setup ---

// --- gRPC mock server implementing SecretBinaryServiceServer ---

type mockBinaryServiceServer struct {
	pb.UnimplementedSecretBinaryServiceServer
}

func (s *mockBinaryServiceServer) List(ctx context.Context, req *pb.SecretBinaryListRequest) (*pb.SecretBinaryListResponse, error) {
	// Authorization token should be in metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 || authHeaders[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretBinaryListResponse{
		Items: []*pb.SecretBinary{
			{
				SecretName: "binary1",
				Data:       []byte{0x01, 0x02, 0x03},
				Meta:       "",
				UpdatedAt:  time.Now().Format(time.RFC3339),
			},
		},
	}, nil
}

// --- Test for SecretBinaryListGRPCFacade.List ---

func TestSecretBinaryListGRPCFacade_List(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretBinaryServiceServer(server, &mockBinaryServiceServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretBinaryServiceClient(conn)
	facade := NewSecretBinaryListGRPCFacade(client)

	secrets, err := facade.List(ctx, "test-token")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "binary1", secrets[0].SecretName)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, secrets[0].Data)
}

// --- gRPC mock server with bad UpdatedAt ---

type badBinaryServer struct {
	pb.UnimplementedSecretBinaryServiceServer
}

func (s *badBinaryServer) List(ctx context.Context, req *pb.SecretBinaryListRequest) (*pb.SecretBinaryListResponse, error) {
	return &pb.SecretBinaryListResponse{
		Items: []*pb.SecretBinary{
			{
				SecretName: "binary1",
				Data:       []byte{0x01, 0x02, 0x03},
				Meta:       "",
				UpdatedAt:  "bad-format",
			},
		},
	}, nil
}

func TestSecretBinaryListGRPCFacade_List_InvalidUpdatedAt(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretBinaryServiceServer(server, &badBinaryServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewSecretBinaryServiceClient(conn)
	facade := NewSecretBinaryListGRPCFacade(client)

	_, err = facade.List(ctx, "test-token")
	assert.Error(t, err)
	assert.Equal(t, "invalid updated_at format in response", err.Error())
}

// --- HTTP tests ---

func TestSecretBinarySaveHTTPFacade_Save(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/save/secret-binary", r.URL.Path)
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "secret1", body["secret_name"])

		expectedEncodedData := base64.StdEncoding.EncodeToString([]byte("binarydata"))
		assert.Equal(t, expectedEncodedData, body["data"])
		assert.NotEmpty(t, body["updated_at"])

		// Optional meta
		if meta, ok := body["meta"]; ok {
			assert.Equal(t, "some-meta", meta)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretBinarySaveHTTPFacade(client)

	updatedAt := time.Now()
	secret := models.SecretBinaryClient{
		SecretName: "secret1",
		Data:       []byte("binarydata"),
		Meta:       nil,
		UpdatedAt:  updatedAt,
	}

	err := facade.Save(context.Background(), "token", secret)
	assert.NoError(t, err)

	// Test with Meta set
	metaStr := "some-meta"
	secret.Meta = &metaStr
	err = facade.Save(context.Background(), "token", secret)
	assert.NoError(t, err)
}

func TestSecretBinaryGetHTTPFacade_Get(t *testing.T) {
	expected := models.SecretBinaryClient{
		SecretName: "secret1",
		Data:       []byte("binarydata"),
		UpdatedAt:  time.Now(),
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/get/secret-binary/secret1", r.URL.Path)
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretBinaryGetHTTPFacade(client)

	secret, err := facade.Get(context.Background(), "token", "secret1")
	assert.NoError(t, err)
	assert.Equal(t, expected.SecretName, secret.SecretName)
	assert.Equal(t, expected.Data, secret.Data)
}

// --- gRPC mocks and tests ---

type mockSecretBinaryServer struct {
	pb.UnimplementedSecretBinaryServiceServer
}

func (s *mockSecretBinaryServer) Get(ctx context.Context, req *pb.SecretBinaryGetRequest) (*pb.SecretBinaryGetResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata in context")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	return &pb.SecretBinaryGetResponse{
		Secret: &pb.SecretBinary{
			SecretName: req.SecretName,
			Data:       []byte("binarydata"), // тут всё ок
			Meta:       "meta-info",
			UpdatedAt:  time.Now().Format(time.RFC3339),
		},
	}, nil
}

func (s *mockSecretBinaryServer) Save(ctx context.Context, req *pb.SecretBinarySaveRequest) (*pb.SecretBinarySaveResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata in context")
	}
	auth := md["authorization"]
	if len(auth) == 0 || auth[0] != "Bearer test-token" {
		return nil, errors.New("unauthorized")
	}

	if req.Secret.SecretName == "" || len(req.Secret.Data) == 0 {
		return nil, errors.New("missing data")
	}

	return &pb.SecretBinarySaveResponse{}, nil
}

func setupGRPCServerSecretBinary(t *testing.T) *grpc.ClientConn {
	lis = bufconn.Listen(bufSize)
	server := grpc.NewServer()
	pb.RegisterSecretBinaryServiceServer(server, &mockSecretBinaryServer{})
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

func TestSecretBinaryGetGRPCFacade_Get(t *testing.T) {
	conn := setupGRPCServerSecretBinary(t)
	client := pb.NewSecretBinaryServiceClient(conn)
	facade := NewSecretBinaryGetGRPCFacade(client)

	secret, err := facade.Get(context.Background(), "test-token", "secret1")
	assert.NoError(t, err)
	assert.Equal(t, "secret1", secret.SecretName)
	assert.Equal(t, []byte("binarydata"), secret.Data) // ← Исправлено
	assert.NotNil(t, secret.Meta)
	assert.Equal(t, "meta-info", *secret.Meta)
}

func TestSecretBinarySaveGRPCFacade_Save(t *testing.T) {
	conn := setupGRPCServerSecretBinary(t)
	client := pb.NewSecretBinaryServiceClient(conn)
	facade := NewSecretBinarySaveGRPCFacade(client)

	secret := models.SecretBinaryClient{
		SecretName: "secret1",
		Data:       []byte("binarydata"),
		UpdatedAt:  time.Now(),
	}

	err := facade.Save(context.Background(), "test-token", secret)
	assert.NoError(t, err)

	metaStr := "meta-info"
	secret.Meta = &metaStr
	err = facade.Save(context.Background(), "test-token", secret)
	assert.NoError(t, err)
}
