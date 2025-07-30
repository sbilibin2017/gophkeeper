package facades

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func newTestClientWithAuth(serverURL string) *resty.Client {
	client := resty.New().
		SetBaseURL(serverURL).
		SetHeader("Authorization", "Bearer dummy-token")
	return client
}

func TestSecretWriterHTTP_Save(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		require.NotEmpty(t, auth)

		var secret models.Secret
		err := json.NewDecoder(r.Body).Decode(&secret)
		require.NoError(t, err)

		assert.NotEmpty(t, secret.SecretName)
		assert.NotEmpty(t, secret.SecretType)
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewSecretWriterHTTP(newTestClientWithAuth(server.URL))

	err := client.Save(
		context.Background(),
		"dummy-token",
		"name1",
		"type1",
		[]byte("ciphertext"),
		[]byte("key"),
	)
	assert.NoError(t, err)
}

func TestSecretReaderHTTP_Get(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/get/type1/name1", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		require.NotEmpty(t, auth)

		secret := models.Secret{
			SecretName: "name1",
			SecretType: "type1",
			Ciphertext: []byte("ciphertext"),
			AESKeyEnc:  []byte("key"),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secret)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewSecretReaderHTTP(newTestClientWithAuth(server.URL))

	secret, err := client.Get(context.Background(), "dummy-token", "type1", "name1")
	require.NoError(t, err)
	assert.Equal(t, "name1", secret.SecretName)
	assert.Equal(t, "type1", secret.SecretType)
	assert.Equal(t, []byte("ciphertext"), secret.Ciphertext)
	assert.Equal(t, []byte("key"), secret.AESKeyEnc)
}

func TestSecretReaderHTTP_List(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		require.NotEmpty(t, auth)

		secrets := []*models.Secret{
			{
				SecretName: "name1",
				SecretType: "type1",
			},
			{
				SecretName: "name2",
				SecretType: "type2",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewSecretReaderHTTP(newTestClientWithAuth(server.URL))

	secrets, err := client.List(context.Background(), "dummy-token")
	require.NoError(t, err)
	require.Len(t, secrets, 2)
	assert.Equal(t, "name1", secrets[0].SecretName)
	assert.Equal(t, "type1", secrets[0].SecretType)
	assert.Equal(t, "name2", secrets[1].SecretName)
	assert.Equal(t, "type2", secrets[1].SecretType)
}

// testSecretService implements SecretWriteService and SecretReadService from your proto.
type testSecretService struct {
	pb.UnimplementedSecretWriteServiceServer
	pb.UnimplementedSecretReadServiceServer

	store map[string]*pb.Secret
}

func newTestSecretService() *testSecretService {
	return &testSecretService{
		store: make(map[string]*pb.Secret),
	}
}

func (s *testSecretService) Save(ctx context.Context, req *pb.SecretSaveRequest) (*emptypb.Empty, error) {
	key := req.SecretType + "/" + req.SecretName
	now := timestamppb.Now()

	s.store[key] = &pb.Secret{
		SecretName:  req.SecretName,
		SecretType:  req.SecretType,
		SecretOwner: "test-owner",
		Ciphertext:  req.Ciphertext,
		AesKeyEnc:   req.AesKeyEnc,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return &emptypb.Empty{}, nil
}

func (s *testSecretService) Get(ctx context.Context, req *pb.SecretGetRequest) (*pb.Secret, error) {
	key := req.SecretType + "/" + req.SecretName
	secret, ok := s.store[key]
	if !ok {
		return nil, grpc.Errorf(grpc.Code(grpc.ErrClientConnClosing), "secret not found")
	}
	return secret, nil
}

func (s *testSecretService) List(_ *emptypb.Empty, stream pb.SecretReadService_ListServer) error {
	for _, secret := range s.store {
		if err := stream.Send(secret); err != nil {
			return err
		}
	}
	return nil
}

func startTestGRPCServer(t *testing.T) (addr string, stopFunc func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	svc := newTestSecretService()
	pb.RegisterSecretWriteServiceServer(server, svc)
	pb.RegisterSecretReadServiceServer(server, svc)

	go func() {
		_ = server.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		server.Stop()
		lis.Close()
	}
}

func TestSecretWriterGRPC_Save_and_SecretReaderGRPC_Get_List(t *testing.T) {
	addr, stop := startTestGRPCServer(t)
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	writer := NewSecretWriterGRPC(conn)
	reader := NewSecretReaderGRPC(conn)

	secret := &models.Secret{
		SecretName: "name1",
		SecretType: "type1",
		Ciphertext: []byte("ciphertext"),
		AESKeyEnc:  []byte("aeskey"),
	}

	// Save the secret
	err = writer.Save(
		context.Background(),
		"test-owner",
		secret.SecretName,
		secret.SecretType,
		secret.Ciphertext,
		secret.AESKeyEnc,
	)
	require.NoError(t, err)

	// Get the secret
	got, err := reader.Get(context.Background(), "test-owner", secret.SecretType, secret.SecretName)
	require.NoError(t, err)

	assert.Equal(t, secret.SecretName, got.SecretName)
	assert.Equal(t, secret.SecretType, got.SecretType)
	assert.Equal(t, secret.Ciphertext, got.Ciphertext)
	assert.Equal(t, secret.AESKeyEnc, got.AESKeyEnc)

	// List secrets
	secrets, err := reader.List(context.Background(), "test-owner")
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	assert.Equal(t, secret.SecretName, secrets[0].SecretName)
	assert.Equal(t, secret.SecretType, secrets[0].SecretType)
}
