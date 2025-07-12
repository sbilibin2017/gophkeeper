package services

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// --- HTTP tests ---

func TestListTextHTTP(t *testing.T) {
	// Создаём тестовый HTTP сервер, который отдаёт JSON с двумя объектами
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"content": "text1"},
			{"content": "text2"}
		]`))
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	result, err := ListTextHTTP(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "text1", result[0].Content)
	require.Equal(t, "text2", result[1].Content)
}

func TestListBinaryHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"data":"AQID"},{"data":"BAUG"}]`)) // base64: [1,2,3] и [4,5,6]
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	result, err := ListBinaryHTTP(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, []byte{1, 2, 3}, result[0].Data)
	require.Equal(t, []byte{4, 5, 6}, result[1].Data)
}

func TestListBankCardHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"number":"1111-2222-3333-4444"},{"number":"5555-6666-7777-8888"}]`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	result, err := ListBankCardHTTP(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "1111-2222-3333-4444", result[0].Number)
	require.Equal(t, "5555-6666-7777-8888", result[1].Number)
}

// --- gRPC tests ---

// Тестовый сервер заглушка для ListService
type testListServer struct {
	pb.UnimplementedListServiceServer
}

func (s *testListServer) ListText(ctx context.Context, req *pb.ListTextRequest) (*pb.ListTextResponse, error) {
	return &pb.ListTextResponse{
		Items: []*pb.TextItem{
			{Content: "text1"},
			{Content: "text2"},
		},
	}, nil
}

func (s *testListServer) ListBinary(ctx context.Context, req *pb.ListBinaryRequest) (*pb.ListBinaryResponse, error) {
	return &pb.ListBinaryResponse{
		Items: []*pb.BinaryItem{
			{Data: []byte{1, 2, 3}},
			{Data: []byte{4, 5, 6}},
		},
	}, nil
}

func (s *testListServer) ListBankCard(ctx context.Context, req *pb.ListBankCardRequest) (*pb.ListBankCardResponse, error) {
	return &pb.ListBankCardResponse{
		Items: []*pb.BankCardItem{
			{Number: "1111-2222-3333-4444"},
			{Number: "5555-6666-7777-8888"},
		},
	}, nil
}

func TestListTextGRPC(t *testing.T) {
	server := grpc.NewServer()
	pb.RegisterListServiceServer(server, &testListServer{})
	lis := newLocalListener(t)
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewListServiceClient(conn)
	items, err := ListTextGRPC(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "text1", items[0].Content)
	require.Equal(t, "text2", items[1].Content)
}

func TestListBinaryGRPC(t *testing.T) {
	server := grpc.NewServer()
	pb.RegisterListServiceServer(server, &testListServer{})
	lis := newLocalListener(t)
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewListServiceClient(conn)
	items, err := ListBinaryGRPC(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, []byte{1, 2, 3}, items[0].Data)
	require.Equal(t, []byte{4, 5, 6}, items[1].Data)
}

func TestListBankCardGRPC(t *testing.T) {
	server := grpc.NewServer()
	pb.RegisterListServiceServer(server, &testListServer{})
	lis := newLocalListener(t)
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewListServiceClient(conn)
	items, err := ListBankCardGRPC(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "1111-2222-3333-4444", items[0].Number)
	require.Equal(t, "5555-6666-7777-8888", items[1].Number)
}

// Вспомогательная функция для создания локального net.Listener
func newLocalListener(t *testing.T) net.Listener {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return lis
}

// Тест для ListUsernamePasswordHTTP
func TestListUsernamePasswordHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Пример данных JSON с двумя элементами
		w.Write([]byte(`[
			{"username":"user1","password":"pass1"},
			{"username":"user2","password":"pass2"}
		]`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	result, err := ListUsernamePasswordHTTP(context.Background(), client)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "user1", result[0].Username)
	require.Equal(t, "pass1", result[0].Password)
	require.Equal(t, "user2", result[1].Username)
	require.Equal(t, "pass2", result[1].Password)
}

// Мок сервер gRPC для ListUsernamePasswordGRPC
type mockListServiceServer struct {
	pb.UnimplementedListServiceServer
}

func (s *mockListServiceServer) ListUsernamePassword(ctx context.Context, req *pb.ListUsernamePasswordRequest) (*pb.ListUsernamePasswordResponse, error) {
	return &pb.ListUsernamePasswordResponse{
		Items: []*pb.UsernamePasswordItem{
			{Username: "user1", Password: "pass1"},
			{Username: "user2", Password: "pass2"},
		},
		Error: "",
	}, nil
}

func TestListUsernamePasswordGRPC(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterListServiceServer(s, &mockListServiceServer{})

	go func() {
		_ = s.Serve(lis)
	}()
	defer s.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithInsecure(),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewListServiceClient(conn)

	result, err := ListUsernamePasswordGRPC(ctx, client)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "user1", result[0].Username)
	require.Equal(t, "pass1", result[0].Password)
	require.Equal(t, "user2", result[1].Username)
	require.Equal(t, "pass2", result[1].Password)
}
