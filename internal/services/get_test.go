package services

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// ======== Embedded HTTP server for testing ========

func startGetTestHTTPServer(t *testing.T) *http.Server {
	mux := http.NewServeMux()

	// Handlers for different secret types
	mux.HandleFunc("/login-password/existing_secret_id", func(w http.ResponseWriter, r *http.Request) {
		lp := models.LoginPassword{
			SecretID:  "existing_secret_id",
			Login:     "user1",
			Password:  "pass1",
			Meta:      map[string]string{"env": "test"},
			UpdatedAt: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(lp)
		require.NoError(t, err)
	})

	mux.HandleFunc("/text/existing_secret_id", func(w http.ResponseWriter, r *http.Request) {
		txt := models.Text{
			SecretID:  "existing_secret_id",
			Content:   "some text content",
			Meta:      map[string]string{"env": "test"},
			UpdatedAt: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(txt)
		require.NoError(t, err)
	})

	mux.HandleFunc("/binary/existing_secret_id", func(w http.ResponseWriter, r *http.Request) {
		bin := models.Binary{
			SecretID:  "existing_secret_id",
			Data:      []byte{0x01, 0x02, 0x03},
			Meta:      map[string]string{"env": "test"},
			UpdatedAt: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(bin)
		require.NoError(t, err)
	})

	mux.HandleFunc("/card/existing_secret_id", func(w http.ResponseWriter, r *http.Request) {
		card := models.Card{
			SecretID:  "existing_secret_id",
			Number:    "1234567812345678",
			Holder:    "John Doe",
			ExpMonth:  12,
			ExpYear:   2030,
			CVV:       "123",
			Meta:      map[string]string{"env": "test"},
			UpdatedAt: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(card)
		require.NoError(t, err)
	})

	srv := &http.Server{
		Addr:    "127.0.0.1:8081",
		Handler: mux,
	}

	ln, err := net.Listen("tcp", srv.Addr)
	require.NoError(t, err)

	ready := make(chan struct{})
	go func() {
		close(ready)
		err := srv.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("HTTP test server failed: %v", err)
		}
	}()

	<-ready
	return srv
}

// ======== Embedded gRPC servers for testing ========

type testLoginPasswordServer struct {
	pb.UnimplementedGetLoginPasswordServiceServer
}

func (s *testLoginPasswordServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetLoginPasswordResponse, error) {
	if req.SecretId != "existing_secret_id" || req.Token == "" {
		return &pb.GetLoginPasswordResponse{Error: "not found or unauthorized"}, nil
	}
	return &pb.GetLoginPasswordResponse{
		LoginPassword: &pb.LoginPassword{
			SecretId:  "existing_secret_id",
			Login:     "user1",
			Password:  "pass1",
			Meta:      map[string]string{"env": "test"},
			UpdatedAt: time.Now().Unix(),
		},
	}, nil
}

type testTextServer struct {
	pb.UnimplementedGetTextServiceServer
}

func (s *testTextServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetTextResponse, error) {
	if req.SecretId != "existing_secret_id" || req.Token == "" {
		return &pb.GetTextResponse{Error: "not found or unauthorized"}, nil
	}
	return &pb.GetTextResponse{
		Text: &pb.Text{
			SecretId:  "existing_secret_id",
			Content:   "some text content",
			Meta:      map[string]string{"env": "test"},
			UpdatedAt: time.Now().Unix(),
		},
	}, nil
}

type testBinaryServer struct {
	pb.UnimplementedGetBinaryServiceServer
}

func (s *testBinaryServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetBinaryResponse, error) {
	if req.SecretId != "existing_secret_id" || req.Token == "" {
		return &pb.GetBinaryResponse{Error: "not found or unauthorized"}, nil
	}
	return &pb.GetBinaryResponse{
		Binary: &pb.Binary{
			SecretId:  "existing_secret_id",
			Data:      []byte{0x01, 0x02, 0x03},
			Meta:      map[string]string{"env": "test"},
			UpdatedAt: time.Now().Unix(),
		},
	}, nil
}

type testCardServer struct {
	pb.UnimplementedGetCardServiceServer
}

func (s *testCardServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetCardResponse, error) {
	if req.SecretId != "existing_secret_id" || req.Token == "" {
		return &pb.GetCardResponse{Error: "not found or unauthorized"}, nil
	}
	return &pb.GetCardResponse{
		Card: &pb.Card{
			SecretId:  "existing_secret_id",
			Number:    "1234567812345678",
			Holder:    "John Doe",
			ExpMonth:  12,
			ExpYear:   2030,
			Cvv:       "123",
			Meta:      map[string]string{"env": "test"},
			UpdatedAt: time.Now().Unix(),
		},
	}, nil
}

func startGetTestGRPCServers(t *testing.T) (lpServer *grpc.Server, lpLis net.Listener, txtServer *grpc.Server, txtLis net.Listener, binServer *grpc.Server, binLis net.Listener, cardServer *grpc.Server, cardLis net.Listener) {
	var err error

	lpLis, err = net.Listen("tcp", "127.0.0.1:0") // 0 means auto-assign free port
	require.NoError(t, err)
	lpServer = grpc.NewServer()
	pb.RegisterGetLoginPasswordServiceServer(lpServer, &testLoginPasswordServer{})
	reflection.Register(lpServer)
	go func() {
		if err := lpServer.Serve(lpLis); err != nil {
			t.Fatalf("LoginPassword gRPC test server failed: %v", err)
		}
	}()

	txtLis, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	txtServer = grpc.NewServer()
	pb.RegisterGetTextServiceServer(txtServer, &testTextServer{})
	reflection.Register(txtServer)
	go func() {
		if err := txtServer.Serve(txtLis); err != nil {
			t.Fatalf("Text gRPC test server failed: %v", err)
		}
	}()

	binLis, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	binServer = grpc.NewServer()
	pb.RegisterGetBinaryServiceServer(binServer, &testBinaryServer{})
	reflection.Register(binServer)
	go func() {
		if err := binServer.Serve(binLis); err != nil {
			t.Fatalf("Binary gRPC test server failed: %v", err)
		}
	}()

	cardLis, err = net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	cardServer = grpc.NewServer()
	pb.RegisterGetCardServiceServer(cardServer, &testCardServer{})
	reflection.Register(cardServer)
	go func() {
		if err := cardServer.Serve(cardLis); err != nil {
			t.Fatalf("Card gRPC test server failed: %v", err)
		}
	}()

	return
}

// ======== HTTP Tests ========

func TestGetHTTPLoginPassword(t *testing.T) {
	srv := startGetTestHTTPServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, srv.Shutdown(ctx))
	}()

	client := resty.New().SetBaseURL("http://127.0.0.1:8081")
	ctx := context.Background()
	token := "test-token"
	secretID := "existing_secret_id"

	lp, err := GetHTTPLoginPassword(ctx, client, token, secretID)
	require.NoError(t, err)
	require.NotNil(t, lp)

	assert.Equal(t, secretID, lp.SecretID)
	assert.Equal(t, "user1", lp.Login)
	assert.Equal(t, "pass1", lp.Password)
	assert.NotNil(t, lp.Meta)
	assert.WithinDuration(t, time.Now(), lp.UpdatedAt, time.Minute)
}

func TestGetHTTPText(t *testing.T) {
	srv := startGetTestHTTPServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, srv.Shutdown(ctx))
	}()

	client := resty.New().SetBaseURL("http://127.0.0.1:8081")
	ctx := context.Background()
	token := "test-token"
	secretID := "existing_secret_id"

	txt, err := GetHTTPText(ctx, client, token, secretID)
	require.NoError(t, err)
	require.NotNil(t, txt)

	assert.Equal(t, secretID, txt.SecretID)
	assert.Equal(t, "some text content", txt.Content)
	assert.NotNil(t, txt.Meta)
	assert.WithinDuration(t, time.Now(), txt.UpdatedAt, time.Minute)
}

func TestGetHTTPBinary(t *testing.T) {
	srv := startGetTestHTTPServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, srv.Shutdown(ctx))
	}()

	client := resty.New().SetBaseURL("http://127.0.0.1:8081")
	ctx := context.Background()
	token := "test-token"
	secretID := "existing_secret_id"

	bin, err := GetHTTPBinary(ctx, client, token, secretID)
	require.NoError(t, err)
	require.NotNil(t, bin)

	assert.Equal(t, secretID, bin.SecretID)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, bin.Data)
	assert.NotNil(t, bin.Meta)
	assert.WithinDuration(t, time.Now(), bin.UpdatedAt, time.Minute)
}

func TestGetHTTPCard(t *testing.T) {
	srv := startGetTestHTTPServer(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		require.NoError(t, srv.Shutdown(ctx))
	}()

	client := resty.New().SetBaseURL("http://127.0.0.1:8081")
	ctx := context.Background()
	token := "test-token"
	secretID := "existing_secret_id"

	card, err := GetHTTPCard(ctx, client, token, secretID)
	require.NoError(t, err)
	require.NotNil(t, card)

	assert.Equal(t, secretID, card.SecretID)
	assert.Equal(t, "1234567812345678", card.Number)
	assert.Equal(t, "John Doe", card.Holder)
	assert.Equal(t, 12, card.ExpMonth)
	assert.Equal(t, 2030, card.ExpYear)
	assert.Equal(t, "123", card.CVV)
	assert.NotNil(t, card.Meta)
	assert.WithinDuration(t, time.Now(), card.UpdatedAt, time.Minute)
}

// ======== gRPC Tests ========

func TestGetGRPCLoginPassword(t *testing.T) {
	lpServer, lpLis, _, _, _, _, _, _ := startGetTestGRPCServers(t)
	defer lpServer.Stop()

	conn, err := grpc.Dial(lpLis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewGetLoginPasswordServiceClient(conn)
	ctx := context.Background()
	token := "test-token"
	secretID := "existing_secret_id"

	lp, err := GetGRPCLoginPassword(ctx, client, token, secretID)
	require.NoError(t, err)
	require.NotNil(t, lp)

	assert.Equal(t, secretID, lp.SecretID)
	assert.Equal(t, "user1", lp.Login)
	assert.Equal(t, "pass1", lp.Password)
	assert.NotNil(t, lp.Meta)
	assert.WithinDuration(t, time.Now(), lp.UpdatedAt, time.Minute)
}

func TestGetGRPCText(t *testing.T) {
	_, _, txtServer, txtLis, _, _, _, _ := startGetTestGRPCServers(t)
	defer txtServer.Stop()

	conn, err := grpc.Dial(txtLis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewGetTextServiceClient(conn)
	ctx := context.Background()
	token := "test-token"
	secretID := "existing_secret_id"

	txt, err := GetGRPCText(ctx, client, token, secretID)
	require.NoError(t, err)
	require.NotNil(t, txt)

	assert.Equal(t, secretID, txt.SecretID)
	assert.Equal(t, "some text content", txt.Content)
	assert.NotNil(t, txt.Meta)
	assert.WithinDuration(t, time.Now(), txt.UpdatedAt, time.Minute)
}

func TestGetGRPCBinary(t *testing.T) {
	_, _, _, _, binServer, binLis, _, _ := startGetTestGRPCServers(t)
	defer binServer.Stop()

	conn, err := grpc.Dial(binLis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewGetBinaryServiceClient(conn)
	ctx := context.Background()
	token := "test-token"
	secretID := "existing_secret_id"

	bin, err := GetGRPCBinary(ctx, client, token, secretID)
	require.NoError(t, err)
	require.NotNil(t, bin)

	assert.Equal(t, secretID, bin.SecretID)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, bin.Data)
	assert.NotNil(t, bin.Meta)
	assert.WithinDuration(t, time.Now(), bin.UpdatedAt, time.Minute)
}

func TestGetGRPCCard(t *testing.T) {
	_, _, _, _, _, _, cardServer, cardLis := startGetTestGRPCServers(t)
	defer cardServer.Stop()

	conn, err := grpc.Dial(cardLis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewGetCardServiceClient(conn)
	ctx := context.Background()
	token := "test-token"
	secretID := "existing_secret_id"

	card, err := GetGRPCCard(ctx, client, token, secretID)
	require.NoError(t, err)
	require.NotNil(t, card)

	assert.Equal(t, secretID, card.SecretID)
	assert.Equal(t, "1234567812345678", card.Number)
	assert.Equal(t, "John Doe", card.Holder)
	assert.Equal(t, 12, card.ExpMonth)
	assert.Equal(t, 2030, card.ExpYear)
	assert.Equal(t, "123", card.CVV)
	assert.NotNil(t, card.Meta)
	assert.WithinDuration(t, time.Now(), card.UpdatedAt, time.Minute)
}
