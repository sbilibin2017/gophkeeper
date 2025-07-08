package services_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

// Dummy encoders for testing
func passEncoder(b []byte) ([]byte, error) { return b, nil }
func failEncoder(b []byte) ([]byte, error) { return nil, errors.New("fail encoder") }

func TestAddLoginPasswordHTTP(t *testing.T) {
	// Setup test server to mock /secrets/login-password
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/secrets/login-password" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := resty.New()
	client.SetBaseURL(ts.URL)

	tests := []struct {
		name     string
		secret   *models.LoginPassword
		encoders []func([]byte) ([]byte, error)
		wantErr  bool
	}{
		{
			name: "valid login-password",
			secret: &models.LoginPassword{
				SecretID:  "lp1",
				Login:     "user123",
				Password:  "pass123",
				Meta:      map[string]string{"env": "test"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){passEncoder},
			wantErr:  false,
		},
		{
			name: "encoder failure on secretID",
			secret: &models.LoginPassword{
				SecretID:  "lp2",
				Login:     "user123",
				Password:  "pass123",
				Meta:      map[string]string{"env": "test"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){failEncoder},
			wantErr:  true,
		},
		{
			name: "encoder failure on meta key",
			secret: &models.LoginPassword{
				SecretID:  "lp3",
				Login:     "user123",
				Password:  "pass123",
				Meta:      map[string]string{"badkey": "val"},
				UpdatedAt: time.Now(),
			},
			// Custom encoder: pass first call, fail on second (which would be meta key or value)
			encoders: []func([]byte) ([]byte, error){
				func(data []byte) ([]byte, error) {
					// fail when encoding meta key "badkey"
					if string(data) == "badkey" {
						return nil, errors.New("fail on meta key")
					}
					return data, nil
				},
			},
			wantErr: true,
		},
		{
			name: "http client failure",
			secret: &models.LoginPassword{
				SecretID:  "lp4",
				Login:     "user123",
				Password:  "pass123",
				Meta:      map[string]string{"env": "test"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){passEncoder},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override client for http client failure test case
			testClient := client
			if tt.name == "http client failure" {
				// Use a client with invalid base URL to trigger failure
				testClient = resty.New()
				testClient.SetBaseURL("http://invalid-host")
			}

			err := services.AddLoginPasswordHTTP(
				context.Background(),
				tt.secret,
				services.WithAddLoginPasswordHTTPEncoders(tt.encoders),
				services.WithAddLoginPasswordHTTPClient(testClient),
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddLoginPasswordHTTP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type stubAddLoginPasswordClient struct {
	pb.AddLoginPasswordServiceClient
	returnError bool
}

func (s *stubAddLoginPasswordClient) AddLoginPassword(ctx context.Context, in *pb.LoginPassword, opts ...grpc.CallOption) (*pb.AddResponse, error) {
	if s.returnError {
		return nil, errors.New("stub error")
	}
	return &pb.AddResponse{}, nil
}

func TestAddLoginPasswordGRPC(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.LoginPassword
		encoders  []func([]byte) ([]byte, error)
		wantErr   bool
		returnErr bool // stub returns error
	}{
		{
			name: "valid data",
			secret: models.NewLoginPassword(
				models.WithLoginPasswordSecretID("grpc1"),
				models.WithLoginPasswordLogin("grpcUser"),
				models.WithLoginPasswordPassword("grpcPass"),
				models.WithLoginPasswordMeta(map[string]string{"grpc": "yes"}),
				models.WithLoginPasswordUpdatedAt(time.Now()),
			),
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   false,
			returnErr: false,
		},
		{
			name: "encoder failure",
			secret: models.NewLoginPassword(
				models.WithLoginPasswordSecretID("grpc2"),
				models.WithLoginPasswordLogin("grpcUser2"),
				models.WithLoginPasswordPassword("grpcPass2"),
				models.WithLoginPasswordMeta(map[string]string{"grpc": "no"}),
				models.WithLoginPasswordUpdatedAt(time.Now()),
			),
			encoders:  []func([]byte) ([]byte, error){failEncoder},
			wantErr:   true,
			returnErr: false,
		},
		{
			name: "grpc client error",
			secret: models.NewLoginPassword(
				models.WithLoginPasswordSecretID("grpc3"),
				models.WithLoginPasswordLogin("grpcUser3"),
				models.WithLoginPasswordPassword("grpcPass3"),
				models.WithLoginPasswordMeta(map[string]string{"grpc": "error"}),
				models.WithLoginPasswordUpdatedAt(time.Now()),
			),
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   true,
			returnErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubClient := &stubAddLoginPasswordClient{returnError: tt.returnErr}

			err := services.AddLoginPasswordGRPC(context.Background(), tt.secret,
				services.WithAddLoginPasswordGRPCEncoders(tt.encoders),
				services.WithAddLoginPasswordGRPCClient(stubClient),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddLoginPasswordGRPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ------------------- AddTextSecretHTTP -------------------

func TestAddTextSecretHTTP(t *testing.T) {
	// Create a test HTTP server that mocks the expected endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/secrets/text" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// You can validate request body here if needed

		// Return success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := resty.New()
	client.SetBaseURL(ts.URL) // Use the test server URL

	tests := []struct {
		name     string
		secret   *models.Text
		encoders []func([]byte) ([]byte, error)
		wantErr  bool
	}{
		{
			name: "valid text",
			secret: &models.Text{
				SecretID:  "text1",
				Content:   "some content",
				Meta:      map[string]string{"key": "value"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){passEncoder},
			wantErr:  false,
		},
		{
			name: "encoder failure",
			secret: &models.Text{
				SecretID:  "text2",
				Content:   "some content",
				Meta:      map[string]string{"key": "value"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){failEncoder},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.AddTextSecretHTTP(
				context.Background(),
				tt.secret,
				services.WithAddTextSecretHTTPEncoders(tt.encoders),
				services.WithAddTextSecretHTTPClient(client),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTextSecretHTTP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Stub for AddTextServiceClient
type stubAddTextClient struct {
	pb.AddTextServiceClient
	returnError bool
	respError   string
}

func (s *stubAddTextClient) AddText(ctx context.Context, in *pb.Text, opts ...grpc.CallOption) (*pb.AddResponse, error) {
	if s.returnError {
		return nil, errors.New("stub error")
	}
	return &pb.AddResponse{Error: s.respError}, nil
}

func TestAddTextSecretGRPC(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.Text
		encoders  []func([]byte) ([]byte, error)
		wantErr   bool
		returnErr bool
		respErr   string
	}{
		{
			name: "valid text",
			secret: &models.Text{
				SecretID:  "grpcText1",
				Content:   "grpc content",
				Meta:      map[string]string{"grpc": "yes"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   false,
			returnErr: false,
			respErr:   "",
		},
		{
			name: "encoder failure",
			secret: &models.Text{
				SecretID:  "grpcText2",
				Content:   "grpc content",
				Meta:      map[string]string{"grpc": "no"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){failEncoder},
			wantErr:   true,
			returnErr: false,
			respErr:   "",
		},
		{
			name: "grpc client error",
			secret: &models.Text{
				SecretID:  "grpcText3",
				Content:   "grpc content",
				Meta:      map[string]string{"grpc": "error"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   true,
			returnErr: true,
			respErr:   "",
		},
		{
			name: "grpc server returned error",
			secret: &models.Text{
				SecretID:  "grpcText4",
				Content:   "grpc content",
				Meta:      map[string]string{"grpc": "respErr"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   true,
			returnErr: false,
			respErr:   "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubClient := &stubAddTextClient{returnError: tt.returnErr, respError: tt.respErr}

			err := services.AddTextSecretGRPC(context.Background(), tt.secret,
				services.WithAddTextSecretGRPCEncoders(tt.encoders),
				services.WithAddTextSecretGRPCClient(stubClient),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTextSecretGRPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ------------------- AddBinarySecretHTTP -------------------

func TestAddBinarySecretHTTP(t *testing.T) {
	// Create test HTTP server to mock /secrets/binary endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/secrets/binary" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		// Optionally, validate request body here if needed

		// Return success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := resty.New()
	client.SetBaseURL(ts.URL) // Use the test server URL

	tests := []struct {
		name     string
		secret   *models.Binary
		encoders []func([]byte) ([]byte, error)
		wantErr  bool
	}{
		{
			name: "valid binary",
			secret: &models.Binary{
				SecretID:  "bin1",
				Data:      []byte{1, 2, 3},
				Meta:      map[string]string{"key": "val"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){passEncoder},
			wantErr:  false,
		},
		{
			name: "encoder failure",
			secret: &models.Binary{
				SecretID:  "bin2",
				Data:      []byte{1, 2, 3},
				Meta:      map[string]string{"key": "val"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){failEncoder},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.AddBinarySecretHTTP(
				context.Background(),
				tt.secret,
				services.WithAddBinarySecretHTTPEncoders(tt.encoders),
				services.WithAddBinarySecretHTTPClient(client),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddBinarySecretHTTP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Stub for AddBinaryServiceClient
type stubAddBinaryClient struct {
	pb.AddBinaryServiceClient
	returnError bool
	respError   string
}

func (s *stubAddBinaryClient) AddBinary(ctx context.Context, in *pb.Binary, opts ...grpc.CallOption) (*pb.AddResponse, error) {
	if s.returnError {
		return nil, errors.New("stub error")
	}
	return &pb.AddResponse{Error: s.respError}, nil
}

func TestAddBinarySecretGRPC(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.Binary
		encoders  []func([]byte) ([]byte, error)
		wantErr   bool
		returnErr bool
		respErr   string
	}{
		{
			name: "valid binary",
			secret: &models.Binary{
				SecretID:  "grpcBin1",
				Data:      []byte{10, 20, 30},
				Meta:      map[string]string{"grpc": "yes"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   false,
			returnErr: false,
			respErr:   "",
		},
		{
			name: "encoder failure",
			secret: &models.Binary{
				SecretID:  "grpcBin2",
				Data:      []byte{10, 20, 30},
				Meta:      map[string]string{"grpc": "no"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){failEncoder},
			wantErr:   true,
			returnErr: false,
			respErr:   "",
		},
		{
			name: "grpc client error",
			secret: &models.Binary{
				SecretID:  "grpcBin3",
				Data:      []byte{10, 20, 30},
				Meta:      map[string]string{"grpc": "error"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   true,
			returnErr: true,
			respErr:   "",
		},
		{
			name: "grpc server returned error",
			secret: &models.Binary{
				SecretID:  "grpcBin4",
				Data:      []byte{10, 20, 30},
				Meta:      map[string]string{"grpc": "respErr"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   true,
			returnErr: false,
			respErr:   "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubClient := &stubAddBinaryClient{returnError: tt.returnErr, respError: tt.respErr}

			err := services.AddBinarySecretGRPC(context.Background(), tt.secret,
				services.WithAddBinarySecretGRPCEncoders(tt.encoders),
				services.WithAddBinarySecretGRPCClient(stubClient),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddBinarySecretGRPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ------------------- AddCardSecretHTTP -------------------

func TestAddCardSecretHTTP(t *testing.T) {
	// Create test HTTP server to mock /secrets/card endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/secrets/card" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		// Optionally read body or validate request here

		// Return success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	client := resty.New()
	client.SetBaseURL(ts.URL) // Use test server URL

	tests := []struct {
		name     string
		secret   *models.Card
		encoders []func([]byte) ([]byte, error)
		wantErr  bool
	}{
		{
			name: "valid card",
			secret: &models.Card{
				SecretID:  "card1",
				Number:    "1234 5678 9012 3456",
				Holder:    "John Doe",
				ExpMonth:  12,
				ExpYear:   2025,
				CVV:       "123",
				Meta:      map[string]string{"key": "val"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){passEncoder},
			wantErr:  false,
		},
		{
			name: "encoder failure",
			secret: &models.Card{
				SecretID:  "card2",
				Number:    "1234 5678 9012 3456",
				Holder:    "John Doe",
				ExpMonth:  12,
				ExpYear:   2025,
				CVV:       "123",
				Meta:      map[string]string{"key": "val"},
				UpdatedAt: time.Now(),
			},
			encoders: []func([]byte) ([]byte, error){failEncoder},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := services.AddCardSecretHTTP(
				context.Background(),
				tt.secret,
				services.WithAddCardSecretHTTPEncoders(tt.encoders),
				services.WithAddCardSecretHTTPClient(client),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddCardSecretHTTP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Stub for AddCardServiceClient
type stubAddCardClient struct {
	pb.AddCardServiceClient
	returnError bool
	respError   string
}

func (s *stubAddCardClient) AddCard(ctx context.Context, in *pb.Card, opts ...grpc.CallOption) (*pb.AddResponse, error) {
	if s.returnError {
		return nil, errors.New("stub error")
	}
	return &pb.AddResponse{Error: s.respError}, nil
}

func TestAddCardSecretGRPC(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.Card
		encoders  []func([]byte) ([]byte, error)
		wantErr   bool
		returnErr bool
		respErr   string
	}{
		{
			name: "valid card",
			secret: &models.Card{
				SecretID:  "grpcCard1",
				Number:    "1111 2222 3333 4444",
				ExpMonth:  11,
				ExpYear:   2024,
				Holder:    "Alice Bob",
				CVV:       "999",
				Meta:      map[string]string{"grpc": "yes"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   false,
			returnErr: false,
			respErr:   "",
		},
		{
			name: "encoder failure",
			secret: &models.Card{
				SecretID:  "grpcCard2",
				Number:    "1111 2222 3333 4444",
				ExpMonth:  11,
				ExpYear:   2024,
				Holder:    "Alice Bob",
				CVV:       "999",
				Meta:      map[string]string{"grpc": "no"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){failEncoder},
			wantErr:   true,
			returnErr: false,
			respErr:   "",
		},
		{
			name: "grpc client error",
			secret: &models.Card{
				SecretID:  "grpcCard3",
				Number:    "1111 2222 3333 4444",
				ExpMonth:  11,
				ExpYear:   2024,
				Holder:    "Alice Bob",
				CVV:       "999",
				Meta:      map[string]string{"grpc": "error"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   true,
			returnErr: true,
			respErr:   "",
		},
		{
			name: "grpc server returned error",
			secret: &models.Card{
				SecretID:  "grpcCard4",
				Number:    "1111 2222 3333 4444",
				ExpMonth:  11,
				ExpYear:   2024,
				Holder:    "Alice Bob",
				CVV:       "999",
				Meta:      map[string]string{"grpc": "respErr"},
				UpdatedAt: time.Now(),
			},
			encoders:  []func([]byte) ([]byte, error){passEncoder},
			wantErr:   true,
			returnErr: false,
			respErr:   "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubClient := &stubAddCardClient{returnError: tt.returnErr, respError: tt.respErr}

			err := services.AddCardSecretGRPC(context.Background(), tt.secret,
				services.WithAddCardSecretGRPCEncoders(tt.encoders),
				services.WithAddCardSecretGRPCClient(stubClient),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddCardSecretGRPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
