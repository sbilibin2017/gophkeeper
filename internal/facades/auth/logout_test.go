package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// --- HTTP Test Server for Logout with error case ---

func startLogoutHTTPTestServer(t *testing.T) (string, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Force-Error") == "true" {
			http.Error(w, "logout failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := &http.Server{Handler: mux}

	go func() {
		_ = srv.Serve(ln)
	}()

	return "http://" + ln.Addr().String(), func() {
		_ = srv.Shutdown(context.Background())
	}
}

// --- gRPC Test Server for Logout with error case ---

type logoutAuthServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *logoutAuthServer) Logout(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if md, ok := grpcMDFromContext(ctx); ok {
		if val, exists := md["force-error"]; exists && len(val) > 0 && val[0] == "true" {
			return nil, status.Error(http.StatusInternalServerError, "logout failed")
		}
	}
	return &emptypb.Empty{}, nil
}

func grpcMDFromContext(ctx context.Context) (map[string][]string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	return md, ok
}

func startLogoutGRPCTestServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, &logoutAuthServer{})

	go func() {
		_ = grpcSrv.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		grpcSrv.GracefulStop()
	}
}

// --- Tests ---

func TestLogoutHTTPFacade_Logout(t *testing.T) {
	url, cleanup := startLogoutHTTPTestServer(t)
	defer cleanup()

	client := resty.New()
	client.SetBaseURL(url)

	facade := NewLogoutHTTPFacade(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Success case
	err := facade.Logout(ctx)
	require.NoError(t, err)

	// Error case
	clientOnErr := resty.New()
	clientOnErr.SetBaseURL(url)
	// Add header to trigger error in server
	clientOnErr.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("X-Force-Error", "true")
		return nil
	})

	facadeErr := NewLogoutHTTPFacade(clientOnErr)
	err = facadeErr.Logout(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "logout failed with status")
}

func TestLogoutGRPCFacade_Logout(t *testing.T) {
	addr, cleanup := startLogoutGRPCTestServer(t)
	defer cleanup()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	facade := NewLogoutGRPCFacade(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Success case
	err = facade.Logout(ctx)
	require.NoError(t, err)

	// Error case: pass metadata to force error
	mdCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("force-error", "true"))
	err = facade.Logout(mdCtx)
	require.Error(t, err)

	var grpcErr interface{ GRPCStatus() *status.Status }
	require.True(t, errors.As(err, &grpcErr))
}
