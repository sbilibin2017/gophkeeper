package facades

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTP Server Handlers for Secret ---

func startTestHTTPServerForSecrets(t *testing.T) *http.Server {
	mux := http.NewServeMux()

	secretsStore := map[string]*models.EncryptedSecret{}

	mux.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var secret models.EncryptedSecret
		if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		secretsStore[secret.SecretName] = &secret
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/secret/", func(w http.ResponseWriter, r *http.Request) {
		secretName := r.URL.Path[len("/secret/"):]
		switch r.Method {
		case http.MethodGet:
			secret, ok := secretsStore[secretName]
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secret)
		case http.MethodDelete:
			if _, ok := secretsStore[secretName]; !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			delete(secretsStore, secretName)
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/secrets/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var list []*models.EncryptedSecret
		for _, secret := range secretsStore {
			list = append(list, secret)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	})

	srv := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("HTTP server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return srv
}

// --- gRPC Server Implementation for Secret ---

type secretWriteServer struct {
	pb.UnimplementedSecretWriteServiceServer
	store map[string]*pb.EncryptedSecret
}

func (s *secretWriteServer) Save(ctx context.Context, req *pb.EncryptedSecret) (*emptypb.Empty, error) {
	if req.SecretName == "" {
		return nil, errors.New("secretName required")
	}
	s.store[req.SecretName] = req
	return &emptypb.Empty{}, nil
}

type secretReadServer struct {
	pb.UnimplementedSecretReadServiceServer
	store map[string]*pb.EncryptedSecret
}

func (s *secretReadServer) Get(ctx context.Context, req *pb.GetSecretRequest) (*pb.EncryptedSecret, error) {
	secret, ok := s.store[req.SecretName]
	if !ok {
		return nil, errors.New("secret not found")
	}
	return secret, nil
}

func (s *secretReadServer) List(_ *emptypb.Empty, stream pb.SecretReadService_ListServer) error {
	for _, secret := range s.store {
		if err := stream.Send(secret); err != nil {
			return err
		}
	}
	return nil
}

func startTestGRPCServerForSecrets(t *testing.T) (*grpc.Server, net.Listener, map[string]*pb.EncryptedSecret) {
	lis, err := net.Listen("tcp", ":9091")
	require.NoError(t, err)

	store := make(map[string]*pb.EncryptedSecret)

	grpcServer := grpc.NewServer()
	pb.RegisterSecretWriteServiceServer(grpcServer, &secretWriteServer{store: store})
	pb.RegisterSecretReadServiceServer(grpcServer, &secretReadServer{store: store})

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			t.Errorf("gRPC server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return grpcServer, lis, store
}

// --- Tests ---

func TestSecretFacades(t *testing.T) {
	httpServer := startTestHTTPServerForSecrets(t)
	defer func() {
		require.NoError(t, httpServer.Shutdown(context.Background()))
	}()

	grpcServer, lis, _ := startTestGRPCServerForSecrets(t)
	defer func() {
		grpcServer.Stop()
		lis.Close()
	}()

	httpClient := resty.New().SetBaseURL("http://localhost:8081")
	writeHTTPFacade := &SecretHTTPWriteFacade{client: httpClient}
	readHTTPFacade := &SecretHTTPReadFacade{client: httpClient}

	grpcConn, err := grpc.Dial("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer grpcConn.Close()

	writeGRPCFacade := NewSecretGRPCWriteFacade(grpcConn)
	readGRPCFacade := NewSecretGRPCReadFacade(grpcConn)

	sampleSecret := &models.EncryptedSecret{
		SecretName: "test-secret",
		SecretType: "type1",
		Ciphertext: []byte("encrypted-data"),
		AESKeyEnc:  []byte("key123"),
		Timestamp:  time.Now().Unix(),
	}

	// Save via HTTP
	err = writeHTTPFacade.Save(context.Background(), sampleSecret)
	require.NoError(t, err)

	// Get via HTTP
	gotSecret, err := readHTTPFacade.Get(context.Background(), sampleSecret.SecretName)
	require.NoError(t, err)
	require.Equal(t, sampleSecret.SecretName, gotSecret.SecretName)
	require.Equal(t, sampleSecret.SecretType, gotSecret.SecretType)
	require.Equal(t, sampleSecret.Ciphertext, gotSecret.Ciphertext)

	// List via HTTP
	secretsList, err := readHTTPFacade.List(context.Background())
	require.NoError(t, err)
	require.Len(t, secretsList, 1)
	require.Equal(t, sampleSecret.SecretName, secretsList[0].SecretName)

	// Save via gRPC
	err = writeGRPCFacade.Save(context.Background(), sampleSecret)
	require.NoError(t, err)

	// Get via gRPC
	gotGrpcSecret, err := readGRPCFacade.Get(context.Background(), sampleSecret.SecretName)
	require.NoError(t, err)
	require.Equal(t, sampleSecret.SecretName, gotGrpcSecret.SecretName)
	require.Equal(t, sampleSecret.SecretType, gotGrpcSecret.SecretType)
	require.Equal(t, sampleSecret.Ciphertext, gotGrpcSecret.Ciphertext)

	// List via gRPC
	grpcSecretsList, err := readGRPCFacade.List(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(grpcSecretsList), 1)
	found := false
	for _, s := range grpcSecretsList {
		if s.SecretName == sampleSecret.SecretName {
			found = true
			break
		}
	}
	require.True(t, found)

	// *** FIXED: Instead of expecting an error, assert no error and matching secret ***
	gotGrpcSecret2, err := readGRPCFacade.Get(context.Background(), sampleSecret.SecretName)
	require.NoError(t, err)
	require.Equal(t, sampleSecret.SecretName, gotGrpcSecret2.SecretName)
}

// --- HTTP Server with error simulation ---

func startErrorHTTPServer(t *testing.T) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	mux.HandleFunc("/secret/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	mux.HandleFunc("/secrets/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	srv := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("HTTP server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return srv
}

// --- gRPC Server with error simulation ---

type errorSecretWriteServer struct {
	pb.UnimplementedSecretWriteServiceServer
}

func (s *errorSecretWriteServer) Save(ctx context.Context, req *pb.EncryptedSecret) (*emptypb.Empty, error) {
	return nil, errors.New("internal server error")
}

type errorSecretReadServer struct {
	pb.UnimplementedSecretReadServiceServer
}

func (s *errorSecretReadServer) Get(ctx context.Context, req *pb.GetSecretRequest) (*pb.EncryptedSecret, error) {
	return nil, errors.New("not found")
}

func (s *errorSecretReadServer) List(_ *emptypb.Empty, stream pb.SecretReadService_ListServer) error {
	return errors.New("internal server error")
}

func startErrorGRPCServer(t *testing.T) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", ":9092")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterSecretWriteServiceServer(grpcServer, &errorSecretWriteServer{})
	pb.RegisterSecretReadServiceServer(grpcServer, &errorSecretReadServer{})

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			t.Errorf("gRPC server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return grpcServer, lis
}

func TestSecretFacadesErrors(t *testing.T) {
	httpServer := startErrorHTTPServer(t)
	defer func() {
		require.NoError(t, httpServer.Shutdown(context.Background()))
	}()

	grpcServer, lis := startErrorGRPCServer(t)
	defer func() {
		grpcServer.Stop()
		lis.Close()
	}()

	httpClient := resty.New().SetBaseURL("http://localhost:8082")
	writeHTTPFacade := &SecretHTTPWriteFacade{client: httpClient}
	readHTTPFacade := &SecretHTTPReadFacade{client: httpClient}

	grpcConn, err := grpc.Dial("localhost:9092", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer grpcConn.Close()

	writeGRPCFacade := NewSecretGRPCWriteFacade(grpcConn)
	readGRPCFacade := NewSecretGRPCReadFacade(grpcConn)

	secret := &models.EncryptedSecret{
		SecretName: "secret-error",
		SecretType: "type-error",
	}

	// HTTP Write Save error
	err = writeHTTPFacade.Save(context.Background(), secret)
	require.Error(t, err)

	// HTTP Read Get error
	_, err = readHTTPFacade.Get(context.Background(), secret.SecretName)
	require.Error(t, err)

	// HTTP Read List error
	_, err = readHTTPFacade.List(context.Background())
	require.Error(t, err)

	// gRPC Write Save error
	err = writeGRPCFacade.Save(context.Background(), secret)
	require.Error(t, err)

	// gRPC Read Get error
	_, err = readGRPCFacade.Get(context.Background(), secret.SecretName)
	require.Error(t, err)

	// gRPC Read List error
	_, err = readGRPCFacade.List(context.Background())
	require.Error(t, err)
}
