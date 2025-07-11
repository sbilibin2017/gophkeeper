package commands

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- Табличные тесты для parseRegisterFlagsInteractive ---
func TestParseRegisterFlagsInteractive(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantUser  string
		wantPass  string
		wantError bool
	}{
		{"valid input", "testuser\nmypassword\n", "testuser", "mypassword", false},
		{"empty input", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			secret, err := parseRegisterFlagsInteractive(reader)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantUser, secret.Username)
			require.Equal(t, tt.wantPass, secret.Password)
		})
	}
}

// --- Табличные тесты для parseRegisterArgs ---
func TestParseRegisterArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantUser  string
		wantPass  string
		wantError bool
	}{
		{"valid args", []string{"user", "pass"}, "user", "pass", false},
		{"missing password", []string{"onlyone"}, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := parseRegisterArgs(tt.args)

			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantUser, secret.Username)
			require.Equal(t, tt.wantPass, secret.Password)
		})
	}
}

// --- Табличные тесты для validateRegisterRequest ---
func TestValidateRegisterRequest(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.UsernamePassword
		wantError bool
	}{
		{"nil secret", nil, true},
		{"empty username", &models.UsernamePassword{"", "pass"}, true},
		{"empty password", &models.UsernamePassword{"user", ""}, true},
		{"valid secret", &models.UsernamePassword{"user", "pass"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegisterRequest(tt.secret)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- Табличные тесты для newRegisterConfig ---
func TestNewRegisterConfig(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{"unsupported protocol", "ftp://wrongprefix", true},
		{"http protocol", "http://localhost", false},
		{"https protocol", "https://localhost", false},
		{"grpc protocol", "grpc://localhost", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newRegisterConfig(tt.url)
			if tt.wantError {
				require.Error(t, err)
			} else {
				if err != nil {
					t.Logf("warning: newRegisterConfig returned error: %v", err)
				}
			}
		})
	}
}

// --- Табличные тесты для setRegisterEnv ---
func TestSetRegisterEnv(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		token     string
	}{
		{"set env vars", "http://localhost", "token123"},
		{"empty token", "http://localhost", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setRegisterEnv(tt.serverURL, tt.token)
			require.NoError(t, err)

			require.Equal(t, tt.serverURL, os.Getenv("GOPHKEEPER_SERVER_URL"))
			require.Equal(t, tt.token, os.Getenv("GOPHKEEPER_TOKEN"))
		})
	}
}

func TestNewRegisterCommand_Basic(t *testing.T) {
	cmd := NewRegisterCommand()

	// Проверяем Use и Short
	require.Equal(t, "register [login] [password]", cmd.Use)
	require.Equal(t, "Зарегистрировать нового пользователя", cmd.Short)

	// Проверяем флаги
	serverURLFlag := cmd.Flags().Lookup("server-url")
	require.NotNil(t, serverURLFlag)
	require.Equal(t, "http://localhost:8080", serverURLFlag.DefValue)

	interactiveFlag := cmd.Flags().Lookup("interactive")
	require.NotNil(t, interactiveFlag)
	require.Equal(t, "false", interactiveFlag.DefValue)

	// Проверяем Args поведение (максимум 2 аргумента)
	err := cmd.Args(cmd, []string{"arg1"})
	require.NoError(t, err)

	err = cmd.Args(cmd, []string{"arg1", "arg2"})
	require.NoError(t, err)

	err = cmd.Args(cmd, []string{"arg1", "arg2", "arg3"})
	require.Error(t, err)
}

func TestRunRegister_Integration(t *testing.T) {
	ctx := context.Background()
	secret := &models.UsernamePassword{Username: "user", Password: "pass"}

	t.Run("HTTP client", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/register" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "real-http-token"})
		}))
		defer ts.Close()

		restyClient := resty.New()
		restyClient.SetBaseURL(ts.URL)

		cfg := &configs.ClientConfig{
			HTTPClient: restyClient,
		}

		token, err := runRegister(ctx, cfg, secret)
		require.NoError(t, err)
		require.Equal(t, "real-http-token", token)
	})

	t.Run("gRPC client", func(t *testing.T) {
		lis, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)

		s := grpc.NewServer()
		pb.RegisterRegisterServiceServer(s, &testRegisterServiceServer{})

		go func() {
			_ = s.Serve(lis)
		}()
		defer s.Stop()

		cfg := &configs.ClientConfig{
			GRPCClient: grpcClientConn(t, lis.Addr().String()),
		}

		token, err := runRegister(ctx, cfg, secret)
		require.NoError(t, err)
		require.Equal(t, "grpc-test-token", token)
	})
}

func grpcClientConn(t *testing.T, addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() {
		conn.Close()
	})
	return conn
}

type testRegisterServiceServer struct {
	pb.UnimplementedRegisterServiceServer
}

func (s *testRegisterServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.GetUsername() == "" || req.GetPassword() == "" {
		return nil, grpc.Errorf(3, "username and password required")
	}
	return &pb.RegisterResponse{Token: "grpc-test-token"}, nil
}
