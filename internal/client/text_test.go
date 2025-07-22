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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
)

// --- Setup test SQLite DB with text_client table ---

func setupTextTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE text_client (
		secret_name TEXT PRIMARY KEY,
		content TEXT,
		meta TEXT,
		updated_at DATETIME
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

// --- gRPC test server implementation ---

type textServer struct {
	pb.UnimplementedTextServiceServer
	storage map[string]*pb.TextDB
}

func newTextServer() *textServer {
	return &textServer{storage: make(map[string]*pb.TextDB)}
}

func (s *textServer) Get(ctx context.Context, req *pb.TextFilterRequest) (*pb.TextDB, error) {
	text, ok := s.storage[req.SecretName]
	if !ok {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return text, nil
}

func (s *textServer) Add(ctx context.Context, req *pb.TextAddRequest) (*emptypb.Empty, error) {
	text := &pb.TextDB{
		SecretName: req.SecretName,
		Content:    req.Content,
		Meta:       req.Meta,
		UpdatedAt:  timestamppb.Now(),
	}
	s.storage[req.SecretName] = text
	return &emptypb.Empty{}, nil
}

// --- HTTP test server ---

func startTextHTTPServer(t *testing.T, storage map[string]*models.TextDB) (string, func()) {
	mux := http.NewServeMux()

	mux.HandleFunc("/text/", func(w http.ResponseWriter, r *http.Request) {
		secretName := r.URL.Path[len("/text/"):]
		text, ok := storage[secretName]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(text)
		require.NoError(t, err)
	})

	mux.HandleFunc("/text", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req models.TextAddRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		storage[req.SecretName] = &models.TextDB{
			SecretName: req.SecretName,
			Content:    req.Content,
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

func TestTextClientDB(t *testing.T) {
	db := setupTextTestDB(t)
	ctx := context.Background()

	req := &models.TextAddRequest{
		SecretName: "text1",
		Content:    "hello world",
		Meta:       &fields.StringMap{Map: map[string]string{"foo": "bar"}},
	}

	// Add text
	err := TextAddClient(ctx, db, req)
	require.NoError(t, err)

	// Get text by secret name
	text, err := TextGetClient(ctx, db, "text1")
	require.NoError(t, err)
	assert.Equal(t, "text1", text.SecretName)
	assert.Equal(t, "hello world", text.Content)

	// List texts
	texts, err := TextListClient(ctx, db)
	require.NoError(t, err)
	assert.Len(t, texts, 1)
}

func TestTextHTTP(t *testing.T) {
	storage := make(map[string]*models.TextDB)
	url, shutdown := startTextHTTPServer(t, storage)
	defer shutdown()

	client := resty.New().SetHostURL(url)
	ctx := context.Background()

	req := &models.TextAddRequest{
		SecretName: "text-http",
		Content:    "resty test",
		Meta:       &fields.StringMap{Map: map[string]string{"a": "b"}},
	}

	// AddHTTP
	err := TextAddHTTP(ctx, client, req)
	require.NoError(t, err)

	// GetHTTP
	text, err := TextGetHTTP(ctx, client, "text-http")
	require.NoError(t, err)
	assert.Equal(t, "text-http", text.SecretName)
	assert.Equal(t, "resty test", text.Content)
}

func TestTextGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	svc := newTextServer()
	pb.RegisterTextServiceServer(server, svc)
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewTextServiceClient(conn)
	ctx := context.Background()

	req := &models.TextAddRequest{
		SecretName: "text-grpc",
		Content:    "grpc test",
		Meta:       &fields.StringMap{Map: map[string]string{"x": "y"}},
	}

	// AddGRPC
	err = TextAddGRPC(ctx, client, req)
	require.NoError(t, err)

	// GetGRPC
	text, err := TextGetGRPC(ctx, client, "text-grpc")
	require.NoError(t, err)
	assert.Equal(t, "text-grpc", text.SecretName)
	assert.Equal(t, "grpc test", text.Content)
}
