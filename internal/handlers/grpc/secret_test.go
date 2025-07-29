package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func TestSecretWriteServiceServer_Save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	svc := NewSecretWriteServiceServer(mockWriter, mockParser)

	ctx := testContextWithAuthToken("Bearer token123")

	req := &pb.SecretSaveRequest{
		SecretName: "mysecret",
		SecretType: "password",
		Ciphertext: []byte("encrypted-data"),
		AesKeyEnc:  []byte("aes-key"),
	}

	mockParser.EXPECT().Parse("token123").Return("user1", nil)

	mockWriter.EXPECT().
		Save(gomock.Any(), "user1", "mysecret", "password", []byte("encrypted-data"), []byte("aes-key")).
		Return(nil)

	resp, err := svc.Save(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSecretWriteServiceServer_Save_MissingMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockParser := NewMockJWTParser(ctrl)
	svc := NewSecretWriteServiceServer(mockWriter, mockParser)

	ctx := context.Background()
	req := &pb.SecretSaveRequest{}

	resp, err := svc.Save(ctx, req)
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestSecretReadServiceServer_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	svc := NewSecretReadServiceServer(mockReader, mockParser)

	ctx := testContextWithAuthToken("Bearer token456")

	req := &pb.SecretGetRequest{
		SecretName: "mysecret",
		SecretType: "password",
	}

	secret := &models.Secret{
		SecretName:  "mysecret",
		SecretType:  "password",
		SecretOwner: "user1",
		Ciphertext:  []byte("cipher"),
		AESKeyEnc:   []byte("aes"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockParser.EXPECT().Parse("token456").Return("user1", nil)
	mockReader.EXPECT().Get(gomock.Any(), "user1", "password", "mysecret").Return(secret, nil)

	resp, err := svc.Get(ctx, req)
	require.NoError(t, err)
	require.Equal(t, "mysecret", resp.SecretName)
	require.Equal(t, "password", resp.SecretType)
	require.Equal(t, "user1", resp.SecretOwner)
	require.Equal(t, []byte("cipher"), resp.Ciphertext)
	require.Equal(t, []byte("aes"), resp.AesKeyEnc)
}

func TestSecretReadServiceServer_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	svc := NewSecretReadServiceServer(mockReader, mockParser)

	mockStream := &mockSecretReadServiceListServer{
		ctx: testContextWithAuthToken("Bearer token789"),
		sendFn: func(secret *pb.Secret) error {
			return nil
		},
	}

	secrets := []*models.Secret{
		{
			SecretName:  "secret1",
			SecretType:  "type1",
			SecretOwner: "user1",
			Ciphertext:  []byte("cipher1"),
			AESKeyEnc:   []byte("aes1"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			SecretName:  "secret2",
			SecretType:  "type2",
			SecretOwner: "user1",
			Ciphertext:  []byte("cipher2"),
			AESKeyEnc:   []byte("aes2"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockParser.EXPECT().Parse("token789").Return("user1", nil)
	mockReader.EXPECT().List(gomock.Any(), "user1").Return(secrets, nil)

	err := svc.List(&emptypb.Empty{}, mockStream)
	require.NoError(t, err)
}

// Helper: create context with metadata authorization token
func testContextWithAuthToken(token string) context.Context {
	md := metadata.New(map[string]string{"authorization": token})
	return metadata.NewIncomingContext(context.Background(), md)
}

// Mock stream for SecretReadService_ListServer
type mockSecretReadServiceListServer struct {
	pb.SecretReadService_ListServer
	ctx    context.Context
	sendFn func(*pb.Secret) error
}

func (m *mockSecretReadServiceListServer) Send(secret *pb.Secret) error {
	return m.sendFn(secret)
}

func (m *mockSecretReadServiceListServer) Context() context.Context {
	return m.ctx
}
