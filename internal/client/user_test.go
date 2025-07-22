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

// --- Setup test SQLite DB with user_client table ---

func setupUserTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE user_client (
		secret_name TEXT PRIMARY KEY,
		username TEXT,
		password TEXT,
		meta TEXT,
		updated_at DATETIME
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

// --- gRPC test server implementation for User ---

type userServer struct {
	pb.UnimplementedUserServiceServer
	storage map[string]*pb.UserDB
}

func newUserServer() *userServer {
	return &userServer{storage: make(map[string]*pb.UserDB)}
}

func (s *userServer) Get(ctx context.Context, req *pb.UserFilterRequest) (*pb.UserDB, error) {
	user, ok := s.storage[req.SecretName]
	if !ok {
		return nil, grpc.Errorf(5, "not found") // NOT_FOUND
	}
	return user, nil
}

func (s *userServer) Add(ctx context.Context, req *pb.UserAddRequest) (*emptypb.Empty, error) {
	user := &pb.UserDB{
		SecretName: req.SecretName,
		Username:   req.Username,
		Password:   req.Password,
		Meta:       req.Meta,
		UpdatedAt:  timestamppb.Now(),
	}
	s.storage[req.SecretName] = user
	return &emptypb.Empty{}, nil
}

// --- HTTP test server for User ---

func startUserHTTPServer(t *testing.T, storage map[string]*models.UserDB) (string, func()) {
	mux := http.NewServeMux()

	mux.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		secretName := r.URL.Path[len("/user/"):]
		user, ok := storage[secretName]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(user)
		require.NoError(t, err)
	})

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req models.UserAddRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		storage[req.SecretName] = &models.UserDB{
			SecretName: req.SecretName,
			Username:   req.Username,
			Password:   req.Password,
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

func TestUserClientDB(t *testing.T) {
	db := setupUserTestDB(t)
	ctx := context.Background()

	req := &models.UserAddRequest{
		SecretName: "user1",
		Username:   "alice",
		Password:   "pass123",
		Meta:       &fields.StringMap{Map: map[string]string{"role": "admin"}},
	}

	// Add user
	err := UserAddClient(ctx, db, req)
	require.NoError(t, err)

	// Get user by secret name
	user, err := UserGetClient(ctx, db, "user1")
	require.NoError(t, err)
	assert.Equal(t, "user1", user.SecretName)
	assert.Equal(t, "alice", user.Username)
	assert.Equal(t, "pass123", user.Password)

	// List users
	users, err := UserListClient(ctx, db)
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestUserHTTP(t *testing.T) {
	storage := make(map[string]*models.UserDB)
	url, shutdown := startUserHTTPServer(t, storage)
	defer shutdown()

	client := resty.New().SetHostURL(url)
	ctx := context.Background()

	req := &models.UserAddRequest{
		SecretName: "user-http",
		Username:   "bob",
		Password:   "secret",
		Meta:       &fields.StringMap{Map: map[string]string{"team": "dev"}},
	}

	// AddHTTP
	err := UserAddHTTP(ctx, client, req)
	require.NoError(t, err)

	// GetHTTP
	user, err := UserGetHTTP(ctx, client, "user-http")
	require.NoError(t, err)
	assert.Equal(t, "user-http", user.SecretName)
	assert.Equal(t, "bob", user.Username)
	assert.Equal(t, "secret", user.Password)
}

func TestUserGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	svc := newUserServer()
	pb.RegisterUserServiceServer(server, svc)
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	ctx := context.Background()

	req := &models.UserAddRequest{
		SecretName: "user-grpc",
		Username:   "charlie",
		Password:   "pwd",
		Meta:       &fields.StringMap{Map: map[string]string{"dept": "sales"}},
	}

	// AddGRPC
	err = UserAddGRPC(ctx, client, req)
	require.NoError(t, err)

	// GetGRPC
	user, err := UserGetGRPC(ctx, client, "user-grpc")
	require.NoError(t, err)
	assert.Equal(t, "user-grpc", user.SecretName)
	assert.Equal(t, "charlie", user.Username)
	assert.Equal(t, "pwd", user.Password)
}
