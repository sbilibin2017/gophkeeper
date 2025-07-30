package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// helper to build context with metadata "authorization: Bearer <token>"
func contextWithAuthToken(token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestSecretWriteServer_Save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockParser := NewMockJWTParser(ctrl)
	srv := NewSecretWriteServer(mockWriter, mockParser)

	// Common request used in tests
	req := &pb.SecretSaveRequest{
		SecretName: "secret1",
		SecretType: "type1",
		Ciphertext: []byte("ciphertext"),
		AesKeyEnc:  []byte("aeskey"),
	}

	tests := []struct {
		name        string
		ctx         context.Context
		req         *pb.SecretSaveRequest
		wantErr     bool
		errContains string
		mockSetup   func()
	}{
		{
			name:    "successful save",
			ctx:     contextWithAuthToken("validtoken"),
			req:     req,
			wantErr: false,
			mockSetup: func() {
				mockParser.EXPECT().Parse("validtoken").Return("user1", nil).Times(1)
				mockWriter.EXPECT().Save(gomock.Any(), "user1", req.SecretName, req.SecretType, req.Ciphertext, req.AesKeyEnc).Return(nil).Times(1)
			},
		},
		{
			name:        "missing metadata",
			ctx:         context.Background(),
			req:         req,
			wantErr:     true,
			errContains: "missing metadata",
			mockSetup:   func() {},
		},
		{
			name:        "missing authorization header",
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.Pairs()),
			req:         req,
			wantErr:     true,
			errContains: "missing authorization token",
			mockSetup:   func() {},
		},
		{
			name:        "invalid authorization header format",
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "InvalidFormat")),
			req:         req,
			wantErr:     true,
			errContains: "invalid authorization token format",
			mockSetup:   func() {},
		},
		{
			name:        "token parse error",
			ctx:         contextWithAuthToken("badtoken"),
			req:         req,
			wantErr:     true,
			errContains: "parse error",
			mockSetup: func() {
				mockParser.EXPECT().Parse("badtoken").Return("", errors.New("parse error")).Times(1)
			},
		},
		{
			name:        "writer save error",
			ctx:         contextWithAuthToken("validtoken"),
			req:         req,
			wantErr:     true,
			errContains: "save error",
			mockSetup: func() {
				mockParser.EXPECT().Parse("validtoken").Return("user1", nil).Times(1)
				mockWriter.EXPECT().Save(gomock.Any(), "user1", req.SecretName, req.SecretType, req.Ciphertext, req.AesKeyEnc).Return(errors.New("save error")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			resp, err := srv.Save(tt.ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.IsType(t, &emptypb.Empty{}, resp)
			}
		})
	}
}

func TestSecretReadServer_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)
	srv := NewSecretReadServer(mockReader, mockParser)

	now := time.Now()

	type mockSetupFunc func()

	tests := []struct {
		name        string
		ctx         context.Context
		req         *pb.SecretGetRequest
		wantErr     bool
		errContains string
		mockSetup   mockSetupFunc
	}{
		{
			name:    "successful get",
			ctx:     contextWithAuthToken("validtoken"),
			req:     &pb.SecretGetRequest{SecretName: "secret1", SecretType: "type1"},
			wantErr: false,
			mockSetup: func() {
				mockParser.EXPECT().Parse("validtoken").Return("user1", nil).Times(1)
				mockReader.EXPECT().Get(gomock.Any(), "user1", "type1", "secret1").Return(&models.Secret{
					SecretName:  "secret1",
					SecretType:  "type1",
					SecretOwner: "user1",
					Ciphertext:  []byte("ciphertext"),
					AESKeyEnc:   []byte("aeskey"),
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil).Times(1)
			},
		},
		{
			name:        "missing metadata",
			ctx:         context.Background(),
			req:         &pb.SecretGetRequest{},
			wantErr:     true,
			errContains: "missing metadata",
			mockSetup:   func() {},
		},
		{
			name:        "missing authorization header",
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.Pairs()),
			req:         &pb.SecretGetRequest{},
			wantErr:     true,
			errContains: "missing authorization",
			mockSetup:   func() {},
		},
		{
			name:        "invalid authorization header format",
			ctx:         metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "InvalidFormat")),
			req:         &pb.SecretGetRequest{},
			wantErr:     true,
			errContains: "invalid authorization token format",
			mockSetup:   func() {},
		},
		{
			name:        "token parse error",
			ctx:         contextWithAuthToken("badtoken"),
			req:         &pb.SecretGetRequest{},
			wantErr:     true,
			errContains: "parse error",
			mockSetup: func() {
				mockParser.EXPECT().Parse("badtoken").Return("", errors.New("parse error")).Times(1)
			},
		},
		{
			name:        "secret get error",
			ctx:         contextWithAuthToken("validtoken"),
			req:         &pb.SecretGetRequest{SecretName: "secret1", SecretType: "type1"},
			wantErr:     true,
			errContains: "not found",
			mockSetup: func() {
				mockParser.EXPECT().Parse("validtoken").Return("user1", nil).Times(1)
				mockReader.EXPECT().Get(gomock.Any(), "user1", "type1", "secret1").Return(nil, errors.New("not found")).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			secret, err := srv.Get(tt.ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, secret)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, secret)
				assert.Equal(t, tt.req.GetSecretName(), secret.SecretName)
				assert.Equal(t, tt.req.GetSecretType(), secret.SecretType)
				assert.Equal(t, "user1", secret.SecretOwner)
				// Check timestamps correctly converted
				assert.True(t, secret.CreatedAt.AsTime().Equal(now))
				assert.True(t, secret.UpdatedAt.AsTime().Equal(now))
			}
		})
	}
}

type mockSecretReadService_ListServer struct {
	ctx     context.Context
	sendFn  func(*pb.Secret) error
	sent    []*pb.Secret
	sendErr error
}

func (m *mockSecretReadService_ListServer) Send(secret *pb.Secret) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, secret)
	if m.sendFn != nil {
		return m.sendFn(secret)
	}
	return nil
}

func (m *mockSecretReadService_ListServer) Context() context.Context {
	return m.ctx
}

func (m *mockSecretReadService_ListServer) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockSecretReadService_ListServer) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockSecretReadService_ListServer) SetTrailer(md metadata.MD) {
	// no-op
}

func (m *mockSecretReadService_ListServer) RecvMsg(interface{}) error {
	return nil
}

func (m *mockSecretReadService_ListServer) SendMsg(interface{}) error {
	return nil
}

func TestSecretReadServer_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)
	srv := NewSecretReadServer(mockReader, mockParser)

	now := time.Now()

	type mockSetupFunc func(stream *mockSecretReadService_ListServer)

	tests := []struct {
		name        string
		stream      *mockSecretReadService_ListServer
		wantErr     bool
		errContains string
		mockSetup   mockSetupFunc
		wantSent    int
	}{
		{
			name: "successful list",
			stream: &mockSecretReadService_ListServer{
				ctx: contextWithAuthToken("validtoken"),
			},
			wantErr:  false,
			wantSent: 2,
			mockSetup: func(stream *mockSecretReadService_ListServer) {
				mockParser.EXPECT().Parse("validtoken").Return("user1", nil).Times(1)
				mockReader.EXPECT().List(gomock.Any(), "user1").Return([]*models.Secret{
					{
						SecretName:  "secret1",
						SecretType:  "type1",
						SecretOwner: "user1",
						Ciphertext:  []byte("ciphertext1"),
						AESKeyEnc:   []byte("aeskey1"),
						CreatedAt:   now,
						UpdatedAt:   now,
					},
					{
						SecretName:  "secret2",
						SecretType:  "type2",
						SecretOwner: "user1",
						Ciphertext:  []byte("ciphertext2"),
						AESKeyEnc:   []byte("aeskey2"),
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				}, nil).Times(1)
			},
		},
		{
			name: "missing metadata",
			stream: &mockSecretReadService_ListServer{
				ctx: context.Background(),
			},
			wantErr:     true,
			errContains: "missing metadata",
			mockSetup:   func(stream *mockSecretReadService_ListServer) {},
		},
		{
			name: "missing authorization",
			stream: &mockSecretReadService_ListServer{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs()),
			},
			wantErr:     true,
			errContains: "missing authorization",
			mockSetup:   func(stream *mockSecretReadService_ListServer) {},
		},
		{
			name: "invalid authorization format",
			stream: &mockSecretReadService_ListServer{
				ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "InvalidFormat")),
			},
			wantErr:     true,
			errContains: "invalid authorization token format",
			mockSetup:   func(stream *mockSecretReadService_ListServer) {},
		},
		{
			name: "parse token error",
			stream: &mockSecretReadService_ListServer{
				ctx: contextWithAuthToken("badtoken"),
			},
			wantErr:     true,
			errContains: "parse error",
			mockSetup: func(stream *mockSecretReadService_ListServer) {
				mockParser.EXPECT().Parse("badtoken").Return("", errors.New("parse error")).Times(1)
			},
		},
		{
			name: "reader list error",
			stream: &mockSecretReadService_ListServer{
				ctx: contextWithAuthToken("validtoken"),
			},
			wantErr:     true,
			errContains: "list error",
			mockSetup: func(stream *mockSecretReadService_ListServer) {
				mockParser.EXPECT().Parse("validtoken").Return("user1", nil).Times(1)
				mockReader.EXPECT().List(gomock.Any(), "user1").Return(nil, errors.New("list error")).Times(1)
			},
		},
		{
			name: "stream send error",
			stream: &mockSecretReadService_ListServer{
				ctx:     contextWithAuthToken("validtoken"),
				sendErr: errors.New("send error"),
			},
			wantErr:     true,
			errContains: "send error",
			mockSetup: func(stream *mockSecretReadService_ListServer) {
				mockParser.EXPECT().Parse("validtoken").Return("user1", nil).Times(1)
				mockReader.EXPECT().List(gomock.Any(), "user1").Return([]*models.Secret{
					{
						SecretName:  "secret1",
						SecretType:  "type1",
						SecretOwner: "user1",
						Ciphertext:  []byte("ciphertext1"),
						AESKeyEnc:   []byte("aeskey1"),
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				}, nil).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(tt.stream)
			err := srv.List(&emptypb.Empty{}, tt.stream)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tt.stream.sent, tt.wantSent)
				// Optional: verify content of sent secrets
				for _, secret := range tt.stream.sent {
					assert.NotEmpty(t, secret.SecretName)
					assert.NotEmpty(t, secret.SecretType)
					assert.Equal(t, "user1", secret.SecretOwner)
					assert.True(t, secret.CreatedAt.AsTime().Equal(now))
					assert.True(t, secret.UpdatedAt.AsTime().Equal(now))
				}
			}
		})
	}
}
