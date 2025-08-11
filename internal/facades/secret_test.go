package facades_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	// assuming your facades package is imported like this:
	"github.com/sbilibin2017/gophkeeper/internal/facades"
)

// --- HTTP facade integration tests ---

func TestSecretHTTPFacades_SaveGetList(t *testing.T) {
	// Storage for testing
	secretsStorage := make(map[string]models.SecretDB)

	// Setup httptest.Server simulating your HTTP API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		secretOwner := strings.TrimPrefix(auth, "Bearer ")

		switch r.URL.Path {
		case "/secret/save":
			var reqBody struct {
				SecretName string `json:"secret_name"`
				SecretType string `json:"secret_type"`
				Ciphertext []byte `json:"ciphertext"`
				AesKeyEnc  []byte `json:"aes_key_enc"`
			}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			key := secretOwner + ":" + reqBody.SecretType + ":" + reqBody.SecretName
			now := time.Now()
			secretsStorage[key] = models.SecretDB{
				SecretOwner: secretOwner,
				SecretName:  reqBody.SecretName,
				SecretType:  reqBody.SecretType,
				Ciphertext:  reqBody.Ciphertext,
				AESKeyEnc:   reqBody.AesKeyEnc,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			w.WriteHeader(http.StatusOK)

		case "/secret/get":
			var reqBody struct {
				SecretName string `json:"secret_name"`
				SecretType string `json:"secret_type"`
			}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			key := secretOwner + ":" + reqBody.SecretType + ":" + reqBody.SecretName
			secret, ok := secretsStorage[key]
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(secret)

		case "/secret/list":
			// Return all secrets owned by secretOwner
			var list []models.SecretDB
			for k, v := range secretsStorage {
				if strings.HasPrefix(k, secretOwner+":") {
					list = append(list, v)
				}
			}
			json.NewEncoder(w).Encode(list)

		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	writeFacade := facades.NewSecretWriteHTTPFacade(client)
	readFacade := facades.NewSecretReadHTTPFacade(client)

	ctx := context.Background()

	// Save secret
	err := writeFacade.Save(ctx, "owner1", "name1", "type1", []byte("cipher1"), []byte("key1"))
	assert.NoError(t, err)

	// Get secret
	secret, err := readFacade.Get(ctx, "owner1", "name1", "type1")
	assert.NoError(t, err)
	assert.Equal(t, "owner1", secret.SecretOwner)
	assert.Equal(t, "name1", secret.SecretName)
	assert.Equal(t, "type1", secret.SecretType)
	assert.Equal(t, []byte("cipher1"), secret.Ciphertext)
	assert.Equal(t, []byte("key1"), secret.AESKeyEnc)

	// List secrets
	secrets, err := readFacade.List(ctx, "owner1")
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "name1", secrets[0].SecretName)
}

// --- gRPC facade integration tests ---

type testSecretWriteServer struct {
	pb.UnimplementedSecretWriteServiceServer
	secrets map[string]*pb.Secret
}

func (s *testSecretWriteServer) Save(ctx context.Context, req *pb.SecretSaveRequest) (*emptypb.Empty, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	auth := md["authorization"]
	if len(auth) == 0 || !strings.HasPrefix(auth[0], "Bearer ") {
		return nil, grpc.Errorf(grpc.Code(grpc.ErrClientConnClosing), "unauthorized")
	}
	owner := strings.TrimPrefix(auth[0], "Bearer ")

	key := owner + ":" + req.SecretType + ":" + req.SecretName
	now := timestamppb.Now()
	s.secrets[key] = &pb.Secret{
		SecretName:  req.SecretName,
		SecretType:  req.SecretType,
		SecretOwner: owner,
		Ciphertext:  req.Ciphertext,
		AesKeyEnc:   req.AesKeyEnc,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return &emptypb.Empty{}, nil
}

type testSecretReadServer struct {
	pb.UnimplementedSecretReadServiceServer
	secrets map[string]*pb.Secret
}

func (s *testSecretReadServer) Get(ctx context.Context, req *pb.SecretGetRequest) (*pb.Secret, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	auth := md["authorization"]
	if len(auth) == 0 || !strings.HasPrefix(auth[0], "Bearer ") {
		return nil, grpc.Errorf(grpc.Code(grpc.ErrClientConnClosing), "unauthorized")
	}
	owner := strings.TrimPrefix(auth[0], "Bearer ")

	key := owner + ":" + req.SecretType + ":" + req.SecretName
	secret, ok := s.secrets[key]
	if !ok {
		return nil, grpc.Errorf(grpc.Code(grpc.ErrClientConnClosing), "not found")
	}

	return secret, nil
}

func (s *testSecretReadServer) List(_ *emptypb.Empty, stream pb.SecretReadService_ListServer) error {
	md, _ := metadata.FromIncomingContext(stream.Context())
	auth := md["authorization"]
	if len(auth) == 0 || !strings.HasPrefix(auth[0], "Bearer ") {
		return grpc.Errorf(grpc.Code(grpc.ErrClientConnClosing), "unauthorized")
	}
	owner := strings.TrimPrefix(auth[0], "Bearer ")

	for k, secret := range s.secrets {
		if strings.HasPrefix(k, owner+":") {
			if err := stream.Send(secret); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestSecretGRPCFacades_SaveGetList(t *testing.T) {
	secrets := make(map[string]*pb.Secret)

	grpcServer := grpc.NewServer()
	writeServer := &testSecretWriteServer{secrets: secrets}
	readServer := &testSecretReadServer{secrets: secrets}
	pb.RegisterSecretWriteServiceServer(grpcServer, writeServer)
	pb.RegisterSecretReadServiceServer(grpcServer, readServer)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer conn.Close()

	writeFacade := facades.NewSecretWriteGRPCFacade(conn)
	readFacade := facades.NewSecretReadGRPCFacade(conn)

	ctx := context.Background()

	// Save secret
	err = writeFacade.Save(ctx, "owner1", "name1", "type1", []byte("cipher1"), []byte("key1"))
	assert.NoError(t, err)

	// Get secret
	secret, err := readFacade.Get(ctx, "owner1", "name1", "type1")
	assert.NoError(t, err)
	assert.Equal(t, "owner1", secret.SecretOwner)
	assert.Equal(t, "name1", secret.SecretName)
	assert.Equal(t, "type1", secret.SecretType)
	assert.Equal(t, []byte("cipher1"), secret.Ciphertext)
	assert.Equal(t, []byte("key1"), secret.AESKeyEnc)

	// List secrets
	secretsList, err := readFacade.List(ctx, "owner1")
	assert.NoError(t, err)
	assert.Len(t, secretsList, 1)
	assert.Equal(t, "name1", secretsList[0].SecretName)
}
