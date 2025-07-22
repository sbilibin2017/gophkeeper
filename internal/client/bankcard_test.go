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

// --- Setup test SQLite DB with bankcard_client table ---

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE bankcard_client (
		secret_name TEXT PRIMARY KEY,
		secret_owner TEXT,
		number TEXT,
		owner TEXT,
		exp TEXT,
		cvv TEXT,
		meta TEXT,
		updated_at DATETIME
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

// --- gRPC test server implementation ---

type bankCardServer struct {
	pb.UnimplementedBankCardServiceServer
	storage map[string]*pb.BankCard
}

func newBankCardServer() *bankCardServer {
	return &bankCardServer{storage: make(map[string]*pb.BankCard)}
}

func (s *bankCardServer) Get(ctx context.Context, req *pb.BankCardFilterRequest) (*pb.BankCard, error) {
	card, ok := s.storage[req.SecretName]
	if !ok {
		return nil, grpc.Errorf(5, "not found") // NOT_FOUND
	}
	return card, nil
}

func (s *bankCardServer) Add(ctx context.Context, req *pb.BankCardAddRequest) (*emptypb.Empty, error) {
	card := &pb.BankCard{
		SecretName:  req.SecretName,
		SecretOwner: "owner_from_grpc", // for test
		Number:      req.Number,
		Owner:       req.Owner,
		Exp:         req.Exp,
		Cvv:         req.Cvv,
		Meta:        req.Meta,
		UpdatedAt:   timestamppb.Now(),
	}
	s.storage[req.SecretName] = card
	return &emptypb.Empty{}, nil
}

// --- HTTP test server ---

func startBankCardHTTPServer(t *testing.T, storage map[string]*models.BankCardDB) (string, func()) {
	mux := http.NewServeMux()

	mux.HandleFunc("/bankcard/", func(w http.ResponseWriter, r *http.Request) {
		secretName := r.URL.Path[len("/bankcard/"):]
		card, ok := storage[secretName]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(card)
		require.NoError(t, err)
	})

	mux.HandleFunc("/bankcard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req models.BankCardAddRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		storage[req.SecretName] = &models.BankCardDB{
			SecretName: req.SecretName,
			Owner:      req.Owner,
			Number:     req.Number,
			Exp:        req.Exp,
			CVV:        req.CVV,
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

func TestBankCardClientDB(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	cardReq := &models.BankCardAddRequest{
		SecretName: "card1",
		Number:     "123456",
		Owner:      "Alice",
		Exp:        "12/34",
		CVV:        "999",
		Meta:       &fields.StringMap{Map: map[string]string{"foo": "bar"}},
	}

	// Add card
	err := BankCardAddClient(ctx, db, cardReq)
	require.NoError(t, err)

	// Get card by secret name
	card, err := BankCardGetClient(ctx, db, "card1")
	require.NoError(t, err)
	assert.Equal(t, "card1", card.SecretName)
	assert.Equal(t, "Alice", card.Owner)
	assert.Equal(t, "123456", card.Number)

	// List cards
	cards, err := BankCardListClient(ctx, db)
	require.NoError(t, err)
	assert.Len(t, cards, 1)
}

func TestBankCardHTTP(t *testing.T) {
	storage := make(map[string]*models.BankCardDB)
	url, shutdown := startBankCardHTTPServer(t, storage)
	defer shutdown()

	client := resty.New().SetHostURL(url)
	ctx := context.Background()

	// Test AddHTTP
	req := &models.BankCardAddRequest{
		SecretName: "card-http",
		Number:     "987654",
		Owner:      "Bob",
		Exp:        "01/25",
		CVV:        "123",
		Meta:       &fields.StringMap{Map: map[string]string{"a": "b"}},
	}
	err := BankCardAddHTTP(ctx, client, req)
	require.NoError(t, err)

	// Test GetHTTP (should return the card just added)
	card, err := BankCardGetHTTP(ctx, client, "card-http")
	require.NoError(t, err)
	assert.Equal(t, "card-http", card.SecretName)
	assert.Equal(t, "Bob", card.Owner)
	assert.Equal(t, "987654", card.Number)
}

func TestBankCardGRPC(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	svc := newBankCardServer()
	pb.RegisterBankCardServiceServer(server, svc)
	go server.Serve(lis)
	defer server.Stop()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewBankCardServiceClient(conn)
	ctx := context.Background()

	// Add card via gRPC
	req := &models.BankCardAddRequest{
		SecretName: "card-grpc",
		Number:     "555555",
		Owner:      "Charlie",
		Exp:        "11/26",
		CVV:        "321",
		Meta:       &fields.StringMap{Map: map[string]string{"x": "y"}},
	}
	err = BankCardAddGRPC(ctx, client, req)
	require.NoError(t, err)

	// Get card via gRPC
	card, err := BankCardGetGRPC(ctx, client, "card-grpc")
	require.NoError(t, err)
	assert.Equal(t, "card-grpc", card.SecretName)
	assert.Equal(t, "Charlie", card.Owner)
	assert.Equal(t, "555555", card.Number)
}
