package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
)

// --- Setup test SQLite DB with binary_client table ---

func setupBinaryTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE binary_client (
		secret_name TEXT PRIMARY KEY,
		data BLOB,
		meta TEXT,
		updated_at DATETIME
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

// --- gRPC test server implementation ---

type binaryServer struct {
	pb.UnimplementedBinaryServiceServer
	storage map[string]*pb.BinaryDB
}

func newBinaryServer() *binaryServer {
	return &binaryServer{storage: make(map[string]*pb.BinaryDB)}
}

func (s *binaryServer) Get(ctx context.Context, req *pb.BinaryFilterRequest) (*pb.BinaryDB, error) {
	bin, ok := s.storage[req.SecretName]
	if !ok {
		return nil, grpc.Errorf(5, "not found") // NOT_FOUND
	}
	return bin, nil
}

func (s *binaryServer) Add(ctx context.Context, req *pb.BinaryAddRequest) (*emptypb.Empty, error) {
	bin := &pb.BinaryDB{
		SecretName:  req.SecretName,
		SecretOwner: "owner_from_grpc", // test-only value
		Data:        req.Data,
		Meta:        req.Meta,
		UpdatedAt:   timestamppb.Now(),
	}
	s.storage[req.SecretName] = bin
	return &emptypb.Empty{}, nil
}

// --- HTTP test server ---

func startBinaryHTTPServer(t *testing.T, storage map[string]*models.BinaryDB) (string, func()) {
	mux := http.NewServeMux()

	mux.HandleFunc("/binary/", func(w http.ResponseWriter, r *http.Request) {
		secretName := r.URL.Path[len("/binary/"):]
		bin, ok := storage[secretName]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(bin)
		require.NoError(t, err)
	})

	mux.HandleFunc("/binary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req models.BinaryAddRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		storage[req.SecretName] = &models.BinaryDB{
			SecretName: req.SecretName,
			Data:       req.Data,
			Meta:       req.Meta,
			UpdatedAt:  time.Now().UTC(),
		}
		w.WriteHeader(http.StatusOK)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)

	return "http://" + ln.Addr().String(), func() {
		srv.Close()
		ln.Close()
	}
}

// --- Tests ---

func TestBinaryClientDB(t *testing.T) {
	db := setupBinaryTestDB(t)
	ctx := context.Background()

	req := &models.BinaryAddRequest{
		SecretName: "bin1",
		Data:       []byte{1, 2, 3, 4},
		Meta:       &fields.StringMap{Map: map[string]string{"foo": "bar"}},
	}

	// Add binary
	err := BinaryAddClient(ctx, db, req)
	require.NoError(t, err)

	// Get binary by secret name
	bin, err := BinaryGetClient(ctx, db, "bin1")
	require.NoError(t, err)
	assert.Equal(t, "bin1", bin.SecretName)
	assert.Equal(t, []byte{1, 2, 3, 4}, bin.Data)

	// List binaries
	bins, err := BinaryListClient(ctx, db)
	require.NoError(t, err)
	assert.Len(t, bins, 1)
}

func TestBinaryHTTP(t *testing.T) {
	storage := make(map[string]*models.BinaryDB)
	url, shutdown := startBinaryHTTPServer(t, storage)
	defer shutdown()

	client := resty.New().SetHostURL(url)
	ctx := context.Background()

	req := &models.BinaryAddRequest{
		SecretName: "bin-http",
		Data:       []byte{9, 8, 7, 6},
		Meta:       &fields.StringMap{Map: map[string]string{"a": "b"}},
	}

	// AddHTTP
	err := BinaryAddHTTP(ctx, client, req)
	require.NoError(t, err)

	// GetHTTP
	bin, err := BinaryGetHTTP(ctx, client, "bin-http")
	require.NoError(t, err)
	assert.Equal(t, "bin-http", bin.SecretName)
	assert.Equal(t, []byte{9, 8, 7, 6}, bin.Data)
}

func TestBinaryGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	svc := newBinaryServer()
	pb.RegisterBinaryServiceServer(server, svc)
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewBinaryServiceClient(conn)
	ctx := context.Background()

	req := &models.BinaryAddRequest{
		SecretName: "bin-grpc",
		Data:       []byte{5, 5, 5, 5},
		Meta:       &fields.StringMap{Map: map[string]string{"x": "y"}},
	}

	// AddGRPC
	err = BinaryAddGRPC(ctx, client, req)
	require.NoError(t, err)

	// GetGRPC
	bin, err := BinaryGetGRPC(ctx, client, "bin-grpc")
	require.NoError(t, err)
	assert.Equal(t, "bin-grpc", bin.SecretName)
	assert.Equal(t, []byte{5, 5, 5, 5}, bin.Data)
}
