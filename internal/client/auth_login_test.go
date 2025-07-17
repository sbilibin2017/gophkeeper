package client

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	_ "modernc.org/sqlite"
)

const loginBufSize = 1024 * 1024

var loginLis *bufconn.Listener

type loginTestServer struct {
	pb.UnimplementedLoginServiceServer
	ReceivedReq *pb.LoginRequest
}

func (s *loginTestServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.ReceivedReq = req
	return &pb.LoginResponse{Token: "dummy-token"}, nil
}

func loginBufDialer(context.Context, string) (net.Conn, error) {
	return loginLis.Dial()
}

func loginSetupGRPCServer(t *testing.T) (pb.LoginServiceClient, *loginTestServer, func()) {
	loginLis = bufconn.Listen(loginBufSize)
	s := grpc.NewServer()
	serverImpl := &loginTestServer{}
	pb.RegisterLoginServiceServer(s, serverImpl)

	go func() {
		if err := s.Serve(loginLis); err != nil {
			t.Logf("gRPC server exited: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(loginBufDialer), grpc.WithInsecure())
	require.NoError(t, err)

	client := pb.NewLoginServiceClient(conn)
	cleanup := func() {
		conn.Close()
		s.Stop()
	}

	return client, serverImpl, cleanup
}

func TestLoginHTTP_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/login", r.URL.Path)
		w.Header().Set("Content-Type", "application/json") // Fix: add content-type header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"token":"dummy-token"}`))
	}))
	defer ts.Close()

	restyClient := resty.New().SetBaseURL(ts.URL)

	req := &models.LoginRequest{
		Username: "validuser",
		Password: "StrongPass1",
	}

	resp, err := LoginHTTP(context.Background(), restyClient, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "dummy-token", resp.Token)
}

func TestLoginHTTP_FailureStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer ts.Close()

	restyClient := resty.New().SetBaseURL(ts.URL)

	req := &models.LoginRequest{
		Username: "validuser",
		Password: "StrongPass1",
	}

	resp, err := LoginHTTP(context.Background(), restyClient, req)
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "login failed with status 401")
}

func TestLoginHTTP_RequestError(t *testing.T) {
	restyClient := resty.New().SetBaseURL("http://invalid.localhost")

	req := &models.LoginRequest{
		Username: "validuser",
		Password: "StrongPass1",
	}

	resp, err := LoginHTTP(context.Background(), restyClient, req)
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send HTTP login request")
}

func TestLoginGRPC_Success(t *testing.T) {
	client, serverImpl, cleanup := loginSetupGRPCServer(t)
	defer cleanup()

	req := &models.LoginRequest{
		Username: "grpcuser",
		Password: "StrongPass1",
	}

	resp, err := LoginGRPC(context.Background(), client, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "grpcuser", serverImpl.ReceivedReq.Username)
	assert.Equal(t, "StrongPass1", serverImpl.ReceivedReq.Password)
	assert.Equal(t, "dummy-token", resp.Token)
}

func TestValidateLoginUsername(t *testing.T) {
	tests := []struct {
		username string
		wantErr  bool
	}{
		{"user", false},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateLoginUsername(tt.username)
		if tt.wantErr {
			assert.Error(t, err, "username: %q", tt.username)
		} else {
			assert.NoError(t, err, "username: %q", tt.username)
		}
	}
}

func TestValidateLoginPassword(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"pass", false},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateLoginPassword(tt.password)
		if tt.wantErr {
			assert.Error(t, err, "password: %q", tt.password)
		} else {
			assert.NoError(t, err, "password: %q", tt.password)
		}
	}
}

func openTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite DB: %v", err)
	}
	return db
}

func tableExists(t *testing.T, db *sqlx.DB, tableName string) bool {
	var count int
	err := db.Get(&count, `SELECT count(name) FROM sqlite_master WHERE type='table' AND name=?;`, tableName)
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	return count > 0
}

func TestCreateBinaryRequestTable(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer db.Close()

	err := CreateBinaryRequestTable(ctx, db)
	assert.NoError(t, err)
	assert.True(t, tableExists(t, db, "secret_binary_request"))
}

func TestCreateTextRequestTable(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer db.Close()

	err := CreateTextRequestTable(ctx, db)
	assert.NoError(t, err)
	assert.True(t, tableExists(t, db, "secret_text_request"))
}

func TestCreateUsernamePasswordRequestTable(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer db.Close()

	err := CreateUsernamePasswordRequestTable(ctx, db)
	assert.NoError(t, err)
	assert.True(t, tableExists(t, db, "secret_username_password_request"))
}

func TestCreateBankCardRequestTable(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer db.Close()

	err := CreateBankCardRequestTable(ctx, db)
	assert.NoError(t, err)
	assert.True(t, tableExists(t, db, "secret_bank_card_request"))
}
