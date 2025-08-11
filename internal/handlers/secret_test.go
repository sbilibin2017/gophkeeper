package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Helper to add token into context
func NewContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, "token", token)
}

// --- HTTP Handler Tests ---

func TestSecretReadHTTPHandler_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockJWT := NewMockUsernameGetter(ctrl)

	handler := NewSecretReadHTTPHandler(mockReader, mockJWT)

	sampleSecrets := []*models.SecretDB{
		{
			SecretName:  "secret1",
			SecretType:  "password",
			SecretOwner: "user1",
			Ciphertext:  []byte("encrypted1"),
			AESKeyEnc:   []byte("key1"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	tests := []struct {
		name           string
		tokenHeader    string
		getUsernameRes string
		getUsernameErr error
		listSecrets    []*models.SecretDB
		listErr        error
		wantStatusCode int
		wantBodyJSON   bool
	}{
		{"success", "validtoken", "user1", nil, sampleSecrets, nil, http.StatusOK, true},
		{"missing token", "", "", nil, nil, nil, http.StatusUnauthorized, false},
		{"invalid token", "invalidtoken", "", errors.New("invalid token"), nil, nil, http.StatusUnauthorized, false},
		{"list error", "validtoken", "user1", nil, nil, errors.New("db error"), http.StatusInternalServerError, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
			if tc.tokenHeader != "" {
				req.Header.Set("Authorization", "Bearer "+tc.tokenHeader)
			}
			w := httptest.NewRecorder()

			if tc.tokenHeader != "" {
				mockJWT.EXPECT().GetUsername(tc.tokenHeader).Return(tc.getUsernameRes, tc.getUsernameErr).MaxTimes(1)
			}

			if tc.wantStatusCode == http.StatusOK || tc.listErr != nil {
				mockReader.EXPECT().List(gomock.Any(), tc.getUsernameRes).Return(tc.listSecrets, tc.listErr).MaxTimes(1)
			}

			handler.List(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatusCode, resp.StatusCode)

			if tc.wantBodyJSON {
				var gotResp SecretListResponse
				err := json.NewDecoder(resp.Body).Decode(&gotResp)
				assert.NoError(t, err)
				assert.Len(t, gotResp.Secrets, len(tc.listSecrets))
			}
		})
	}
}

func TestSecretReadHTTPHandler_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockJWT := NewMockUsernameGetter(ctrl)

	handler := NewSecretReadHTTPHandler(mockReader, mockJWT)

	sampleSecret := &models.SecretDB{
		SecretName:  "secret1",
		SecretType:  "password",
		SecretOwner: "user1",
		Ciphertext:  []byte("encrypted1"),
		AESKeyEnc:   []byte("key1"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name           string
		tokenHeader    string
		getUsernameRes string
		getUsernameErr error
		queryParams    string
		getSecret      *models.SecretDB
		getErr         error
		wantStatusCode int
		wantBodyJSON   bool
	}{
		{"success", "validtoken", "user1", nil, "?name=secret1&type=password", sampleSecret, nil, http.StatusOK, true},
		{"missing token", "", "", nil, "", nil, nil, http.StatusUnauthorized, false},
		{"invalid token", "invalidtoken", "", errors.New("invalid token"), "", nil, nil, http.StatusUnauthorized, false},
		{"missing query params", "validtoken", "user1", nil, "?name=secret1", nil, nil, http.StatusBadRequest, false},
		{"secret not found", "validtoken", "user1", nil, "?name=secret1&type=password", nil, errors.New("not found"), http.StatusNotFound, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "/secrets/get" + tc.queryParams
			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tc.tokenHeader != "" {
				req.Header.Set("Authorization", "Bearer "+tc.tokenHeader)
			}
			w := httptest.NewRecorder()

			if tc.tokenHeader != "" {
				mockJWT.EXPECT().GetUsername(tc.tokenHeader).Return(tc.getUsernameRes, tc.getUsernameErr).MaxTimes(1)
			}

			if tc.getSecret != nil || tc.getErr != nil {
				mockReader.EXPECT().Get(gomock.Any(), tc.getUsernameRes, "secret1", "password").Return(tc.getSecret, tc.getErr).MaxTimes(1)
			}

			handler.Get(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatusCode, resp.StatusCode)

			if tc.wantBodyJSON {
				var gotResp SecretResponse
				err := json.NewDecoder(resp.Body).Decode(&gotResp)
				assert.NoError(t, err)
				assert.Equal(t, tc.getSecret.SecretName, gotResp.SecretName)
			}
		})
	}
}

func TestSecretWriteHTTPHandler_Save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockJWT := NewMockUsernameGetter(ctrl)

	handler := NewSecretWriteHTTPHandler(mockWriter, mockJWT)

	type saveReq struct {
		Name       string `json:"name"`
		Type       string `json:"type"`
		Ciphertext string `json:"ciphertext"`  // base64 encoded
		AESKeyEnc  string `json:"aes_key_enc"` // base64 encoded
	}

	validReq := saveReq{
		Name:       "secret1",
		Type:       "password",
		Ciphertext: "ZW5jcnlwdGVkLWRhdGE=", // base64 for "encrypted-data"
		AESKeyEnc:  "ZW5jcnlwdGVkLWtleQ==", // base64 for "encrypted-key"
	}

	tests := []struct {
		name           string
		token          string
		getUsernameRes string
		getUsernameErr error
		requestBody    interface{}
		saveErr        error
		wantStatusCode int
	}{
		{
			name:           "success",
			token:          "validtoken",
			getUsernameRes: "user1",
			getUsernameErr: nil,
			requestBody:    validReq,
			saveErr:        nil,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing token",
			token:          "",
			getUsernameRes: "",
			getUsernameErr: nil,
			requestBody:    validReq,
			saveErr:        nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			token:          "invalidtoken",
			getUsernameRes: "",
			getUsernameErr: errors.New("invalid token"),
			requestBody:    validReq,
			saveErr:        nil,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "bad json body",
			token:          "validtoken",
			getUsernameRes: "user1",
			getUsernameErr: nil,
			requestBody:    "not a json",
			saveErr:        nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "save error",
			token:          "validtoken",
			getUsernameRes: "user1",
			getUsernameErr: nil,
			requestBody:    validReq,
			saveErr:        errors.New("db error"),
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			var err error

			switch v := tc.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, err = json.Marshal(v)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(bodyBytes))
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}

			w := httptest.NewRecorder()

			if tc.token != "" {
				mockJWT.EXPECT().GetUsername(tc.token).Return(tc.getUsernameRes, tc.getUsernameErr).MaxTimes(1)
			}

			if tc.wantStatusCode == http.StatusOK || tc.wantStatusCode == http.StatusInternalServerError {
				mockWriter.EXPECT().Save(gomock.Any(), tc.getUsernameRes, validReq.Name, validReq.Type,
					[]byte("encrypted-data"), []byte("encrypted-key")).Return(tc.saveErr).MaxTimes(1)
			}

			handler.Save(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatusCode, resp.StatusCode)
		})
	}
}

// --- Mocks ---

type mockWriter struct {
	saveErr                       error
	lastCtx                       context.Context
	lastOwner, lastName, lastType string
	lastCiphertext, lastAesKeyEnc []byte
}

func (m *mockWriter) Save(ctx context.Context, secretOwner, secretName, secretType string, ciphertext, aesKeyEnc []byte) error {
	m.lastCtx = ctx
	m.lastOwner = secretOwner
	m.lastName = secretName
	m.lastType = secretType
	m.lastCiphertext = ciphertext
	m.lastAesKeyEnc = aesKeyEnc
	return m.saveErr
}

type mockReader struct {
	listSecrets []*models.SecretDB
	listErr     error
	getSecret   *models.SecretDB
	getErr      error
}

func (m *mockReader) List(ctx context.Context, secretOwner string) ([]*models.SecretDB, error) {
	return m.listSecrets, m.listErr
}

func (m *mockReader) Get(ctx context.Context, secretOwner, secretName, secretType string) (*models.SecretDB, error) {
	return m.getSecret, m.getErr
}

type mockJWT struct {
	username string
	err      error
}

func (m *mockJWT) GetUsername(tokenStr string) (string, error) {
	return m.username, m.err
}

// --- Context helper: inject token in gRPC metadata ---

func ctxWithToken(token string) context.Context {
	md := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
	return metadata.NewIncomingContext(context.Background(), md)
}

// --- Tests ---

func TestSecretWriteGRPCHandler_Save_Success(t *testing.T) {
	writer := &mockWriter{}
	jwtHandler := &mockJWT{username: "testuser"}

	handler := NewSecretWriteGRPCHandler(writer, jwtHandler)

	req := &pb.SecretSaveRequest{
		SecretName: "secret1",
		SecretType: "type1",
		Ciphertext: []byte("encrypted-data"),
		AesKeyEnc:  []byte("encrypted-key"),
	}

	ctx := ctxWithToken("valid-token")

	resp, err := handler.Save(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "testuser", writer.lastOwner)
	assert.Equal(t, req.SecretName, writer.lastName)
	assert.Equal(t, req.SecretType, writer.lastType)
	assert.Equal(t, req.Ciphertext, writer.lastCiphertext)
	assert.Equal(t, req.AesKeyEnc, writer.lastAesKeyEnc)
}

func TestSecretWriteGRPCHandler_Save_Unauthenticated(t *testing.T) {
	writer := &mockWriter{}
	jwtHandler := &mockJWT{}

	handler := NewSecretWriteGRPCHandler(writer, jwtHandler)

	req := &pb.SecretSaveRequest{
		SecretName: "secret1",
		SecretType: "type1",
		Ciphertext: []byte("encrypted-data"),
		AesKeyEnc:  []byte("encrypted-key"),
	}

	// Context without token
	ctx := context.Background()

	_, err := handler.Save(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	// JWT handler returns error
	ctx = ctxWithToken("bad-token")
	jwtHandler.err = errors.New("invalid token")
	_, err = handler.Save(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestSecretWriteGRPCHandler_Save_SaveFails(t *testing.T) {
	writer := &mockWriter{saveErr: errors.New("db error")}
	jwtHandler := &mockJWT{username: "testuser"}

	handler := NewSecretWriteGRPCHandler(writer, jwtHandler)

	req := &pb.SecretSaveRequest{
		SecretName: "secret1",
		SecretType: "type1",
		Ciphertext: []byte("encrypted-data"),
		AesKeyEnc:  []byte("encrypted-key"),
	}

	ctx := ctxWithToken("valid-token")

	_, err := handler.Save(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// --- SecretReadGRPCHandler.Get tests ---

func TestSecretReadGRPCHandler_Get_Success(t *testing.T) {
	now := time.Now()
	secret := &models.SecretDB{
		SecretName:  "secret1",
		SecretType:  "type1",
		SecretOwner: "testuser",
		Ciphertext:  []byte("encrypted-data"),
		AESKeyEnc:   []byte("encrypted-key"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	reader := &mockReader{
		listSecrets: []*models.SecretDB{secret},
	}

	jwtHandler := &mockJWT{username: "testuser"}

	handler := NewSecretReadGRPCHandler(reader, jwtHandler)

	ctx := ctxWithToken("valid-token")

	req := &pb.SecretGetRequest{
		SecretName: "secret1",
		SecretType: "type1",
	}

	resp, err := handler.Get(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, secret.SecretName, resp.SecretName)
	assert.Equal(t, secret.SecretType, resp.SecretType)
	assert.Equal(t, secret.SecretOwner, resp.SecretOwner)
	assert.Equal(t, secret.Ciphertext, resp.Ciphertext)
	assert.Equal(t, secret.AESKeyEnc, resp.AesKeyEnc)
	assert.Equal(t, timestamppb.New(secret.CreatedAt), resp.CreatedAt)
	assert.Equal(t, timestamppb.New(secret.UpdatedAt), resp.UpdatedAt)
}

func TestSecretReadGRPCHandler_Get_Unauthenticated(t *testing.T) {
	reader := &mockReader{}
	jwtHandler := &mockJWT{}

	handler := NewSecretReadGRPCHandler(reader, jwtHandler)

	req := &pb.SecretGetRequest{
		SecretName: "secret1",
		SecretType: "type1",
	}

	ctx := context.Background()

	_, err := handler.Get(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	ctx = ctxWithToken("bad-token")
	jwtHandler.err = errors.New("invalid token")

	_, err = handler.Get(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestSecretReadGRPCHandler_Get_SecretNotFound(t *testing.T) {
	reader := &mockReader{
		listSecrets: []*models.SecretDB{},
	}
	jwtHandler := &mockJWT{username: "testuser"}

	handler := NewSecretReadGRPCHandler(reader, jwtHandler)

	req := &pb.SecretGetRequest{
		SecretName: "notfound",
		SecretType: "type1",
	}

	ctx := ctxWithToken("valid-token")

	_, err := handler.Get(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestSecretReadGRPCHandler_Get_ListFails(t *testing.T) {
	reader := &mockReader{
		listErr: errors.New("db error"),
	}
	jwtHandler := &mockJWT{username: "testuser"}

	handler := NewSecretReadGRPCHandler(reader, jwtHandler)

	req := &pb.SecretGetRequest{
		SecretName: "secret1",
		SecretType: "type1",
	}

	ctx := ctxWithToken("valid-token")

	_, err := handler.Get(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

// --- SecretReadGRPCHandler.List tests ---

type mockSecretReadServiceListServer struct {
	grpc.ServerStream
	sent []*pb.Secret
	ctx  context.Context
}

func (m *mockSecretReadServiceListServer) Send(secret *pb.Secret) error {
	m.sent = append(m.sent, secret)
	return nil
}

func (m *mockSecretReadServiceListServer) Context() context.Context {
	return m.ctx
}

func TestSecretReadGRPCHandler_List_Success(t *testing.T) {
	now := time.Now()
	secrets := []*models.SecretDB{
		{
			SecretName:  "secret1",
			SecretType:  "type1",
			SecretOwner: "testuser",
			Ciphertext:  []byte("data1"),
			AESKeyEnc:   []byte("key1"),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			SecretName:  "secret2",
			SecretType:  "type2",
			SecretOwner: "testuser",
			Ciphertext:  []byte("data2"),
			AESKeyEnc:   []byte("key2"),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	reader := &mockReader{
		listSecrets: secrets,
	}

	jwtHandler := &mockJWT{username: "testuser"}

	handler := NewSecretReadGRPCHandler(reader, jwtHandler)

	ctx := ctxWithToken("valid-token")

	stream := &mockSecretReadServiceListServer{ctx: ctx}

	err := handler.List(&emptypb.Empty{}, stream)
	assert.NoError(t, err)
	assert.Len(t, stream.sent, 2)

	for i, secret := range secrets {
		assert.Equal(t, secret.SecretName, stream.sent[i].SecretName)
		assert.Equal(t, secret.SecretType, stream.sent[i].SecretType)
		assert.Equal(t, secret.SecretOwner, stream.sent[i].SecretOwner)
		assert.Equal(t, secret.Ciphertext, stream.sent[i].Ciphertext)
		assert.Equal(t, secret.AESKeyEnc, stream.sent[i].AesKeyEnc)
		assert.Equal(t, timestamppb.New(secret.CreatedAt), stream.sent[i].CreatedAt)
		assert.Equal(t, timestamppb.New(secret.UpdatedAt), stream.sent[i].UpdatedAt)
	}
}

func TestSecretReadGRPCHandler_List_Unauthenticated(t *testing.T) {
	reader := &mockReader{}
	jwtHandler := &mockJWT{}

	handler := NewSecretReadGRPCHandler(reader, jwtHandler)

	stream := &mockSecretReadServiceListServer{ctx: context.Background()}

	err := handler.List(&emptypb.Empty{}, stream)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	stream.ctx = ctxWithToken("bad-token")
	jwtHandler.err = errors.New("invalid token")

	err = handler.List(&emptypb.Empty{}, stream)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestSecretReadGRPCHandler_List_ListFails(t *testing.T) {
	reader := &mockReader{
		listErr: errors.New("db error"),
	}
	jwtHandler := &mockJWT{username: "testuser"}

	handler := NewSecretReadGRPCHandler(reader, jwtHandler)

	stream := &mockSecretReadServiceListServer{ctx: ctxWithToken("valid-token")}

	err := handler.List(&emptypb.Empty{}, stream)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}
