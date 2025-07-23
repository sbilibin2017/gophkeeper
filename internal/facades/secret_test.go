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

	// In-memory store for secrets
	secretsStore := map[string]*models.EncryptedSecret{}

	// Add secret handler - POST /secret
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

	// Get secret handler - GET /secret/{secretName}
	mux.HandleFunc("/secret/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Extract secretName from URL path
		secretName := r.URL.Path[len("/secret/"):]
		secret, ok := secretsStore[secretName]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secret)
	})

	// List secrets handler - GET /secrets/
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

	time.Sleep(100 * time.Millisecond) // wait for server to start
	return srv
}

// --- gRPC Server Implementation for Secret ---

type secretWriteServer struct {
	pb.UnimplementedSecretWriteServiceServer
	store map[string]*pb.EncryptedSecret
}

func (s *secretWriteServer) Add(ctx context.Context, req *pb.EncryptedSecret) (*emptypb.Empty, error) {
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

	time.Sleep(100 * time.Millisecond) // wait for server to start
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

	// Setup HTTP facades
	httpClient := resty.New().SetBaseURL("http://localhost:8081")
	writeHTTPFacade := &SecretHTTPWriteFacade{client: httpClient}
	readHTTPFacade := &SecretHTTPReadFacade{client: httpClient}

	// Setup gRPC facades
	grpcConn, err := grpc.Dial("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer grpcConn.Close()

	writeGRPCFacade := NewSecretGRPCWriteFacade(grpcConn)
	readGRPCFacade := NewSecretGRPCReadFacade(grpcConn)

	// Sample secret to use in tests
	sampleSecret := &models.EncryptedSecret{
		SecretName: "test-secret",
		SecretType: "type1",
		Ciphertext: []byte("encrypted-data"),
		HMAC:       []byte("hmac"),
		Nonce:      []byte("nonce"),
		AESKeyEnc:  []byte("key123"),
		Timestamp:  time.Now().Unix(),
	}

	// --- HTTP Add secret ---
	err = writeHTTPFacade.Add(context.Background(), sampleSecret)
	require.NoError(t, err)

	// --- HTTP Get secret ---
	gotSecret, err := readHTTPFacade.Get(context.Background(), sampleSecret.SecretName)
	require.NoError(t, err)
	require.Equal(t, sampleSecret.SecretName, gotSecret.SecretName)
	require.Equal(t, sampleSecret.SecretType, gotSecret.SecretType)
	require.Equal(t, sampleSecret.Ciphertext, gotSecret.Ciphertext)

	// --- HTTP List secrets ---
	secretsList, err := readHTTPFacade.List(context.Background())
	require.NoError(t, err)
	require.Len(t, secretsList, 1)
	require.Equal(t, sampleSecret.SecretName, secretsList[0].SecretName)

	// --- gRPC Add secret ---
	err = writeGRPCFacade.Add(context.Background(), sampleSecret)
	require.NoError(t, err)

	// --- gRPC Get secret ---
	gotGrpcSecret, err := readGRPCFacade.Get(context.Background(), sampleSecret.SecretName)
	require.NoError(t, err)
	require.Equal(t, sampleSecret.SecretName, gotGrpcSecret.SecretName)
	require.Equal(t, sampleSecret.SecretType, gotGrpcSecret.SecretType)
	require.Equal(t, sampleSecret.Ciphertext, gotGrpcSecret.Ciphertext)

	// --- gRPC List secrets ---
	grpcSecretsList, err := readGRPCFacade.List(context.Background())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(grpcSecretsList), 1) // at least one secret
	found := false
	for _, s := range grpcSecretsList {
		if s.SecretName == sampleSecret.SecretName {
			found = true
			break
		}
	}
	require.True(t, found)
}

// --- HTTP Server with error simulation ---

func startErrorHTTPServer(t *testing.T) *http.Server {
	mux := http.NewServeMux()

	// Add secret - always return 500 error
	mux.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	// Get secret - always return 404 not found
	mux.HandleFunc("/secret/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	// List secrets - always return 500 error
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

func (s *errorSecretWriteServer) Add(ctx context.Context, req *pb.EncryptedSecret) (*emptypb.Empty, error) {
	return nil, errors.New("write error")
}

type errorSecretReadServer struct {
	pb.UnimplementedSecretReadServiceServer
}

func (s *errorSecretReadServer) Get(ctx context.Context, req *pb.GetSecretRequest) (*pb.EncryptedSecret, error) {
	return nil, errors.New("not found")
}

func (s *errorSecretReadServer) List(_ *emptypb.Empty, stream pb.SecretReadService_ListServer) error {
	return errors.New("list error")
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

// --- Error Tests ---

func TestSecretFacades_ErrorCases(t *testing.T) {
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
		SecretName: "err-secret",
		SecretType: "type",
		Ciphertext: []byte("cipher"),
		HMAC:       []byte("hmac"),
		Nonce:      []byte("nonce"),
		AESKeyEnc:  []byte("key"),
		Timestamp:  time.Now().Unix(),
	}

	// HTTP Add error
	err = writeHTTPFacade.Add(context.Background(), secret)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to add secret")

	// HTTP Get error
	_, err = readHTTPFacade.Get(context.Background(), secret.SecretName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get secret")

	// HTTP List error
	_, err = readHTTPFacade.List(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list secrets")

	// gRPC Add error
	err = writeGRPCFacade.Add(context.Background(), secret)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write error")

	// gRPC Get error
	_, err = readGRPCFacade.Get(context.Background(), secret.SecretName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// gRPC List error
	_, err = readGRPCFacade.List(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "list error")
}
