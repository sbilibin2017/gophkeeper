package facades

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const bufSize2 = 1024 * 1024

// ---- Setup gRPC Server for testing ----

func setupGRPCServer2(t *testing.T) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize2)
	s := grpc.NewServer()

	pb.RegisterSecretWriteServiceServer(s, &mockSecretWriteServer{})
	pb.RegisterSecretReadServiceServer(s, &mockSecretReadServer{})

	go func() {
		err := s.Serve(lis)
		require.NoError(t, err)
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	require.NoError(t, err)

	return conn, func() {
		conn.Close()
		s.Stop()
	}
}

type mockSecretWriteServer struct {
	pb.UnimplementedSecretWriteServiceServer
}

func (m *mockSecretWriteServer) Save(ctx context.Context, req *pb.SecretSaveRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

type mockSecretReadServer struct {
	pb.UnimplementedSecretReadServiceServer
}

func (m *mockSecretReadServer) Get(ctx context.Context, req *pb.SecretGetRequest) (*pb.SecretDB, error) {
	return &pb.SecretDB{
		SecretName: req.SecretName,
		SecretType: req.SecretType,
		Ciphertext: []byte("ciphertext"),
		AesKeyEnc:  []byte("aeskeyenc"),
		CreatedAt:  timestamppb.New(time.Date(2023, 7, 26, 12, 0, 0, 0, time.UTC)),
		UpdatedAt:  timestamppb.New(time.Date(2023, 7, 26, 13, 0, 0, 0, time.UTC)),
	}, nil
}

func (m *mockSecretReadServer) List(req *pb.SecretListRequest, stream pb.SecretReadService_ListServer) error {
	secret1 := &pb.SecretDB{
		SecretName: "secret1",
		SecretType: "type1",
		Ciphertext: []byte("ciphertext1"),
		AesKeyEnc:  []byte("aeskeyenc1"),
		CreatedAt:  timestamppb.New(time.Date(2023, 7, 25, 10, 0, 0, 0, time.UTC)),
		UpdatedAt:  timestamppb.New(time.Date(2023, 7, 25, 11, 0, 0, 0, time.UTC)),
	}
	secret2 := &pb.SecretDB{
		SecretName: "secret2",
		SecretType: "type2",
		Ciphertext: []byte("ciphertext2"),
		AesKeyEnc:  []byte("aeskeyenc2"),
		CreatedAt:  timestamppb.New(time.Date(2023, 7, 26, 9, 0, 0, 0, time.UTC)),
		UpdatedAt:  timestamppb.New(time.Date(2023, 7, 26, 10, 0, 0, 0, time.UTC)),
	}

	if err := stream.Send(secret1); err != nil {
		return err
	}
	if err := stream.Send(secret2); err != nil {
		return err
	}
	return nil
}

func TestSaveSecretHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/save/", r.URL.Path)
		require.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := resty.New().SetBaseURL(srv.URL)
	facade := NewSecretWriteHTTP(client)

	err := facade.Save(context.Background(), &models.SecretSaveRequest{
		Token: "mytoken",
	})
	require.NoError(t, err)
}

func TestGetSecretHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/get/type1/name1", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"secret_name": "name1",
			"secret_type": "type1",
			"ciphertext": "Y2lwaGVydGV4dA==",
			"aes_key_enc": "YWVza2V5ZW5j",
			"created_at": "2023-07-26T12:00:00Z",
			"updated_at": "2023-07-26T13:00:00Z"
		}`))
	}))
	defer srv.Close()

	client := resty.New().SetBaseURL(srv.URL)
	facade := NewSecretReadHTTP(client)

	req := &models.SecretGetRequest{
		Token:      "token",
		SecretType: "type1",
		SecretName: "name1",
	}

	secret, err := facade.Get(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "name1", secret.SecretName)
	require.Equal(t, "type1", secret.SecretType)
	require.Equal(t, []byte("ciphertext"), secret.Ciphertext)
	require.Equal(t, []byte("aeskeyenc"), secret.AESKeyEnc)
}

func TestListSecretsHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/list/", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{
				"secret_name": "name1",
				"secret_type": "type1",
				"ciphertext": "Y2lwaGVydGV4dA==",
				"aes_key_enc": "YWVza2V5ZW5j",
				"created_at": "2023-07-26T12:00:00Z",
				"updated_at": "2023-07-26T13:00:00Z"
			},
			{
				"secret_name": "name2",
				"secret_type": "type2",
				"ciphertext": "Y2lwaGVydGV4dDI=",
				"aes_key_enc": "YWVza2V5ZW5jMg==",
				"created_at": "2023-07-25T10:00:00Z",
				"updated_at": "2023-07-25T11:00:00Z"
			}
		]`))
	}))
	defer srv.Close()

	client := resty.New().SetBaseURL(srv.URL)
	facade := NewSecretReadHTTP(client)

	req := &models.SecretListRequest{
		Token: "token",
	}

	secrets, err := facade.List(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, secrets, 2)
	require.Equal(t, "name1", secrets[0].SecretName)
	require.Equal(t, "name2", secrets[1].SecretName)
}

func TestSaveSecretGRPC(t *testing.T) {
	conn, cleanup := setupGRPCServer2(t)
	defer cleanup()

	facade := NewSecretWriteGRPC(conn)

	err := facade.Save(context.Background(), &models.SecretSaveRequest{
		SecretName: "sname",
		SecretType: "stype",
		Ciphertext: []byte("ciphertext"),
		AESKeyEnc:  []byte("aeskeyenc"),
		Token:      "token",
	})
	require.NoError(t, err)
}

func TestGetSecretGRPC(t *testing.T) {
	conn, cleanup := setupGRPCServer2(t)
	defer cleanup()

	facade := NewSecretReadGRPC(conn)

	req := &models.SecretGetRequest{
		SecretName: "sname",
		SecretType: "stype",
		Token:      "token",
	}

	secret, err := facade.Get(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "sname", secret.SecretName)
	require.Equal(t, "stype", secret.SecretType)
	require.Equal(t, []byte("ciphertext"), secret.Ciphertext)
	require.Equal(t, []byte("aeskeyenc"), secret.AESKeyEnc)
	require.Equal(t, time.Date(2023, 7, 26, 12, 0, 0, 0, time.UTC), secret.CreatedAt)
	require.Equal(t, time.Date(2023, 7, 26, 13, 0, 0, 0, time.UTC), secret.UpdatedAt)
}

func TestListSecretsGRPC(t *testing.T) {
	conn, cleanup := setupGRPCServer2(t)
	defer cleanup()

	facade := NewSecretReadGRPC(conn)

	req := &models.SecretListRequest{
		Token: "token",
	}

	secrets, err := facade.List(context.Background(), req)
	require.NoError(t, err)
	require.Len(t, secrets, 2)

	require.Equal(t, "secret1", secrets[0].SecretName)
	require.Equal(t, "type1", secrets[0].SecretType)
	require.Equal(t, []byte("ciphertext1"), secrets[0].Ciphertext)

	require.Equal(t, "secret2", secrets[1].SecretName)
	require.Equal(t, "type2", secrets[1].SecretType)
	require.Equal(t, []byte("ciphertext2"), secrets[1].Ciphertext)
}
