package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// --- HTTP тесты (исправленный ClientInstallHTTP тест) ---

func TestClientInstallHTTP(t *testing.T) {
	expectedContent := []byte("mock client binary")

	// HTTP сервер с содержимым бинарника
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := fmt.Sprintf("/clients/%s-%s", runtime.GOOS, runtime.GOARCH)
		require.Equal(t, expectedPath, r.URL.Path)

		w.WriteHeader(200)
		_, _ = w.Write(expectedContent)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(cwd)
	require.NoError(t, os.Chdir(tmpDir))

	client := resty.New().SetBaseURL(server.URL)

	err = ClientInstallHTTP(context.Background(), client)
	require.NoError(t, err)

	fileName, err := generateClientBinaryFileName(runtime.GOOS, runtime.GOARCH)
	require.NoError(t, err)

	filePath := filepath.Join(tmpDir, fileName)
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, expectedContent, data)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0755), info.Mode().Perm())
	}
}

// --- Тест generateClientBinaryFileName ---

func TestGenerateClientBinaryFileName(t *testing.T) {
	tests := []struct {
		goos       string
		goarch     string
		want       string
		shouldFail bool
	}{
		{"windows", "amd64", "client-windows-amd64.exe", false},
		{"linux", "amd64", "client-linux-amd64", false},
		{"darwin", "amd64", "client-darwin-amd64", false},
		{"darwin", "arm64", "client-darwin-arm64", false},
		{"plan9", "386", "", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s-%s", tt.goos, tt.goarch), func(t *testing.T) {
			got, err := generateClientBinaryFileName(tt.goos, tt.goarch)
			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// --- gRPC тесты ---

const bufSize = 1024 * 1024

func TestClientInstallGRPC(t *testing.T) {
	// Создаём буферный net.Listener для gRPC без реального TCP
	listener := bufconn.Listen(bufSize)

	// Создаём и запускаем тестовый gRPC сервер
	srv := grpc.NewServer()
	pb.RegisterClientInstallServiceServer(srv, &testClientInstallServer{
		t:            t,
		expectedOS:   runtime.GOOS,
		expectedArch: runtime.GOARCH,
	})
	go func() {
		_ = srv.Serve(listener)
	}()
	defer srv.Stop()

	// Создаём gRPC клиент с подключением через буфер
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewClientInstallServiceClient(conn)

	// Меняем рабочую директорию на временную
	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(cwd)
	require.NoError(t, os.Chdir(tmpDir))

	// Вызываем тестируемую функцию
	err = ClientInstallGRPC(ctx, client)
	require.NoError(t, err)

	fileName, err := generateClientBinaryFileName(runtime.GOOS, runtime.GOARCH)
	require.NoError(t, err)

	filePath := filepath.Join(tmpDir, fileName)
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, []byte("grpc client binary"), data)
}

// Тестовый gRPC сервер
type testClientInstallServer struct {
	pb.UnimplementedClientInstallServiceServer
	t            *testing.T
	expectedOS   string
	expectedArch string
}

func (s *testClientInstallServer) DownloadClient(ctx context.Context, req *pb.InstallRequest) (*pb.InstallResponse, error) {
	assert.Equal(s.t, s.expectedOS, req.Os)
	assert.Equal(s.t, s.expectedArch, req.Arch)

	fileName, err := generateClientBinaryFileName(req.Os, req.Arch)
	require.NoError(s.t, err)

	return &pb.InstallResponse{
		BinaryData: []byte("grpc client binary"),
		FileName:   fileName,
		Error:      "",
	}, nil
}
