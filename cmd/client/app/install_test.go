package app

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// --- Моки ---

type testInstallServer struct {
	pb.UnimplementedClientInstallServiceServer
}

func (s *testInstallServer) DownloadClient(ctx context.Context, req *pb.InstallRequest) (*pb.InstallResponse, error) {
	return &pb.InstallResponse{
		BinaryData: []byte("data"),
		FileName:   "client.bin",
	}, nil
}

func TestRunInstallApp_HTTP_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake client binary"))
	}))
	defer ts.Close()

	err := runInstallApp(ts.URL)
	assert.NoError(t, err)

	// Удаляем бинарник после теста
	err = os.Remove("client-linux-amd64")
	if err != nil && !os.IsNotExist(err) {
		t.Logf("failed to remove binary: %v", err)
	}
}

func TestRunInstallApp_GRPC_Success(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	pb.RegisterClientInstallServiceServer(grpcServer, &testInstallServer{})
	go grpcServer.Serve(lis)
	defer grpcServer.Stop()

	serverURL := "grpc://" + lis.Addr().String()
	err = runInstallApp(serverURL)
	assert.NoError(t, err)

	// Удаляем бинарник после теста
	err = os.Remove("client-linux-amd64")
	if err != nil && !os.IsNotExist(err) {
		t.Logf("failed to remove binary: %v", err)
	}
}

func TestRunInstallApp_NoClientsAvailable(t *testing.T) {
	err := runInstallApp("invalid-url-without-scheme")

	require.Error(t, err)

}

func TestRunInstallApp_InvalidURL_ErrorCreatingConfig(t *testing.T) {
	// URL с синтаксической ошибкой
	url := "://bad-url"

	err := runInstallApp(url)
	assert.Error(t, err)
}

func TestParseInstallFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("server-url", "", "Base URL")

	t.Run("missing server-url flag", func(t *testing.T) {
		_, err := parseInstallFlags(cmd)
		require.Error(t, err)

	})

	t.Run("valid server-url flag", func(t *testing.T) {
		err := cmd.Flags().Set("server-url", "http://localhost:8080")
		require.NoError(t, err)

		_, err = parseInstallFlags(cmd)
		require.NoError(t, err)

	})
}
