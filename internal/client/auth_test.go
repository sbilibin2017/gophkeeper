package client

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	_ "modernc.org/sqlite"
)

// --- HTTP server for register/login/logout tests ---

func authRunTestHTTPServer(t *testing.T) (*http.Server, string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		if req.Username == "error" {
			http.Error(w, "registration error", http.StatusBadRequest)
			return
		}

		resp := models.RegisterResponse{Token: "test-token-for-" + req.Username}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		if req.Username == "error" {
			http.Error(w, "login error", http.StatusUnauthorized)
			return
		}

		resp := models.LoginResponse{Token: "test-token-for-" + req.Username}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &http.Server{Handler: mux}

	go func() {
		err := srv.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("HTTP server error: %v", err)
		}
	}()

	return srv, listener.Addr().String()
}

// --- gRPC server for register/login/logout tests ---

const testBufSize = 1024 * 1024

type testAuthServer struct {
	pb.UnimplementedRegisterServiceServer
	pb.UnimplementedLoginServiceServer
	pb.UnimplementedLogoutServiceServer
}

func (s *testAuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "error" {
		return nil, status.Errorf(codes.InvalidArgument, "registration error")
	}
	return &pb.RegisterResponse{Token: "test-token-for-" + req.Username}, nil
}

func (s *testAuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Username == "error" {
		return nil, status.Errorf(codes.Unauthenticated, "login error")
	}
	return &pb.LoginResponse{Token: "test-token-for-" + req.Username}, nil
}

func (s *testAuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	if req.Token == "" {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}
	return &emptypb.Empty{}, nil
}

func runTestGRPCServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	listener := bufconn.Listen(testBufSize)
	grpcServer := grpc.NewServer()
	srv := &testAuthServer{}
	pb.RegisterRegisterServiceServer(grpcServer, srv)
	pb.RegisterLoginServiceServer(grpcServer, srv)
	pb.RegisterLogoutServiceServer(grpcServer, srv)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Errorf("gRPC server error: %v", err)
		}
	}()

	return grpcServer, listener
}

func bufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}
}

// --- Tests ---

func TestRegisterHTTP(t *testing.T) {
	srv, addr := authRunTestHTTPServer(t)
	defer srv.Close()

	client := resty.New().SetBaseURL("http://" + addr)
	ctx := context.Background()

	// Success
	req := &models.RegisterRequest{Username: "user1", Password: "pass"}
	resp, err := RegisterHTTP(ctx, client, req)
	require.NoError(t, err)
	assert.Equal(t, "test-token-for-user1", resp.Token)

	// Failure case (bad status)
	reqFail := &models.RegisterRequest{Username: "error", Password: "pass"}
	_, err = RegisterHTTP(ctx, client, reqFail)
	require.Error(t, err)
}

func TestRegisterGRPC(t *testing.T) {
	grpcServer, listener := runTestGRPCServer(t)
	defer grpcServer.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(listener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewRegisterServiceClient(conn)
	req := &models.RegisterRequest{Username: "user1", Password: "pass"}

	resp, err := RegisterGRPC(ctx, client, req)
	require.NoError(t, err)
	assert.Equal(t, "test-token-for-user1", resp.Token)

	// Error case
	reqFail := &models.RegisterRequest{Username: "error", Password: "pass"}
	_, err = RegisterGRPC(ctx, client, reqFail)
	require.Error(t, err)
}

func TestLoginHTTP(t *testing.T) {
	srv, addr := authRunTestHTTPServer(t)
	defer srv.Close()

	client := resty.New().SetBaseURL("http://" + addr)
	ctx := context.Background()

	req := &models.LoginRequest{Username: "user1", Password: "pass"}
	resp, err := LoginHTTP(ctx, client, req)
	require.NoError(t, err)
	assert.Equal(t, "test-token-for-user1", resp.Token)

	// Failure case
	reqFail := &models.LoginRequest{Username: "error", Password: "pass"}
	_, err = LoginHTTP(ctx, client, reqFail)
	require.Error(t, err)
}

func TestLoginGRPC(t *testing.T) {
	grpcServer, listener := runTestGRPCServer(t)
	defer grpcServer.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(listener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewLoginServiceClient(conn)
	req := &models.LoginRequest{Username: "user1", Password: "pass"}

	resp, err := LoginGRPC(ctx, client, req)
	require.NoError(t, err)
	assert.Equal(t, "test-token-for-user1", resp.Token)

	// Error case
	reqFail := &models.LoginRequest{Username: "error", Password: "pass"}
	_, err = LoginGRPC(ctx, client, reqFail)
	require.Error(t, err)
}

func TestLogoutHTTP(t *testing.T) {
	srv, addr := authRunTestHTTPServer(t)
	defer srv.Close()

	client := resty.New().SetBaseURL("http://" + addr)
	ctx := context.Background()

	// Success
	err := LogoutHTTP(ctx, client, &models.LogoutRequest{Token: "valid-token"})
	require.NoError(t, err)

	// Failure case: missing token
	err = LogoutHTTP(ctx, client, &models.LogoutRequest{Token: ""})
	require.Error(t, err)
}

func TestLogoutGRPC(t *testing.T) {
	grpcServer, listener := runTestGRPCServer(t)
	defer grpcServer.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(listener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewLogoutServiceClient(conn)

	// Success
	err = LogoutGRPC(ctx, client, &models.LogoutRequest{Token: "valid-token"})
	require.NoError(t, err)

	// Failure: empty token triggers error
	err = LogoutGRPC(ctx, client, &models.LogoutRequest{Token: ""})
	require.Error(t, err)
}

func TestValidateRegisterUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid username", "user_123", false},
		{"Too short", "ab", true},
		{"Too long", "a_very_long_username_exceeding_30_chars", true},
		{"Invalid chars", "user!name", true},
		{"Empty username", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegisterUsername(tt.username)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateRegisterPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "Abcdef1g", false},
		{"Too short", "Ab1", true},
		{"No uppercase", "abcdefg1", true},
		{"No lowercase", "ABCDEFG1", true},
		{"No digit", "Abcdefgh", true},
		{"Empty password", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegisterPassword(tt.password)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateLoginUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid username", "user", false},
		{"Empty username", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLoginUsername(tt.username)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateLoginPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "password123", false},
		{"Empty password", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLoginPassword(tt.password)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDB(t *testing.T) {
	// test DB opens and closes without error
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()
}

func TestCreateBinaryRequestTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := CreateBinaryRequestTable(db)
	require.NoError(t, err)

}

func TestCreateTextRequestTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := CreateTextRequestTable(db)
	require.NoError(t, err)

}

func TestCreateUsernamePasswordRequestTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := CreateUsernamePasswordRequestTable(db)
	require.NoError(t, err)

}

func TestCreateBankCardRequestTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := CreateBankCardRequestTable(db)
	require.NoError(t, err)

}

// helper funcs

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)
	return db
}
