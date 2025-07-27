package grpc_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/sbilibin2017/gophkeeper/internal/handlers/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func TestSecretWriteServiceServer_Save(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := grpc.NewMockSecretWriter(ctrl)
	mockJWT := grpc.NewMockJWTParser(ctrl)
	server := grpc.NewSecretWriteServiceServer(mockWriter, mockJWT)

	ctx := context.Background()
	token := "valid-token"
	owner := "user123"
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx = metadata.NewIncomingContext(ctx, md)

	req := &pb.SecretSaveRequest{
		SecretName: "name",
		SecretType: "type",
		Ciphertext: []byte("ciphertext"),
		AesKeyEnc:  []byte("aeskey"),
	}

	t.Run("success", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return(owner, nil)
		mockWriter.EXPECT().
			Save(ctx, owner, req.SecretName, req.SecretType, req.Ciphertext, req.AesKeyEnc).
			Return(nil)

		resp, err := server.Save(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("missing metadata", func(t *testing.T) {
		emptyCtx := context.Background()
		_, err := server.Save(emptyCtx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing metadata")
	})

	t.Run("missing authorization header", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())
		_, err := server.Save(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing authorization token")
	})

	t.Run("invalid authorization header", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "InvalidToken"))
		_, err := server.Save(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authorization token format")
	})

	t.Run("jwt parse error", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return("", errors.New("invalid token"))
		_, err := server.Save(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid token")
	})

	t.Run("writer save error", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return(owner, nil)
		mockWriter.EXPECT().
			Save(ctx, owner, req.SecretName, req.SecretType, req.Ciphertext, req.AesKeyEnc).
			Return(errors.New("save failed"))

		_, err := server.Save(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "save failed")
	})
}

func TestSecretReadServiceServer_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := grpc.NewMockSecretReader(ctrl)
	mockJWT := grpc.NewMockJWTParser(ctrl)
	server := grpc.NewSecretReadServiceServer(mockReader, mockJWT)

	ctx := context.Background()
	token := "valid-token"
	owner := "user123"
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx = metadata.NewIncomingContext(ctx, md)

	req := &pb.SecretGetRequest{
		SecretName: "name",
		SecretType: "type",
	}

	secretModel := &models.Secret{
		SecretOwner: owner,
		SecretName:  req.SecretName,
		SecretType:  req.SecretType,
		Ciphertext:  []byte("ciphertext"),
		AESKeyEnc:   []byte("aeskey"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return(owner, nil)
		mockReader.EXPECT().Get(ctx, owner, req.SecretType, req.SecretName).Return(secretModel, nil)

		resp, err := server.Get(ctx, req)
		require.NoError(t, err)
		require.Equal(t, secretModel.SecretName, resp.SecretName)
		require.Equal(t, secretModel.SecretType, resp.SecretType)
		require.Equal(t, secretModel.SecretOwner, resp.SecretOwner)
		require.Equal(t, secretModel.Ciphertext, resp.Ciphertext)
		require.Equal(t, secretModel.AESKeyEnc, resp.AesKeyEnc)
		require.WithinDuration(t, secretModel.CreatedAt, resp.CreatedAt.AsTime(), time.Second)
		require.WithinDuration(t, secretModel.UpdatedAt, resp.UpdatedAt.AsTime(), time.Second)
	})

	t.Run("missing metadata", func(t *testing.T) {
		_, err := server.Get(context.Background(), req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing metadata")
	})

	t.Run("missing authorization token", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())
		_, err := server.Get(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing authorization token")
	})

	t.Run("invalid authorization header", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "BadToken"))
		_, err := server.Get(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authorization token format")
	})

	t.Run("jwt parse error", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return("", errors.New("invalid token"))
		_, err := server.Get(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid token")
	})

	t.Run("reader get error", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return(owner, nil)
		mockReader.EXPECT().Get(ctx, owner, req.SecretType, req.SecretName).Return(nil, errors.New("get failed"))
		_, err := server.Get(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "get failed")
	})
}

type mockSecretReadServiceListServer struct {
	grpcpb pb.SecretReadService_ListServer
	ctx    context.Context
	sendFn func(*pb.Secret) error
}

func (m *mockSecretReadServiceListServer) Send(resp *pb.Secret) error {
	if m.sendFn != nil {
		return m.sendFn(resp)
	}
	return nil
}

func (m *mockSecretReadServiceListServer) SetHeader(metadata.MD) error  { return nil }
func (m *mockSecretReadServiceListServer) SendHeader(metadata.MD) error { return nil }
func (m *mockSecretReadServiceListServer) SetTrailer(metadata.MD)       {}
func (m *mockSecretReadServiceListServer) Context() context.Context     { return m.ctx }
func (m *mockSecretReadServiceListServer) SendMsg(mes interface{}) error {
	return nil
}
func (m *mockSecretReadServiceListServer) RecvMsg(mes interface{}) error {
	return nil
}

func TestSecretReadServiceServer_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := grpc.NewMockSecretReader(ctrl)
	mockJWT := grpc.NewMockJWTParser(ctrl)
	server := grpc.NewSecretReadServiceServer(mockReader, mockJWT)

	token := "valid-token"
	owner := "user123"
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))

	secret1 := &models.Secret{
		SecretOwner: owner,
		SecretName:  "secret1",
		SecretType:  "type1",
		Ciphertext:  []byte("cipher1"),
		AESKeyEnc:   []byte("key1"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	secret2 := &models.Secret{
		SecretOwner: owner,
		SecretName:  "secret2",
		SecretType:  "type2",
		Ciphertext:  []byte("cipher2"),
		AESKeyEnc:   []byte("key2"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return(owner, nil)
		mockReader.EXPECT().List(ctx, owner).Return([]*models.Secret{secret1, secret2}, nil)

		sendCount := 0
		stream := &mockSecretReadServiceListServer{
			ctx: ctx,
			sendFn: func(resp *pb.Secret) error {
				sendCount++
				require.True(t, strings.HasPrefix(resp.SecretName, "secret"))
				return nil
			},
		}

		err := server.List(&emptypb.Empty{}, stream)
		require.NoError(t, err)
		require.Equal(t, 2, sendCount)
	})

	t.Run("missing metadata", func(t *testing.T) {
		stream := &mockSecretReadServiceListServer{
			ctx: context.Background(),
		}
		err := server.List(&emptypb.Empty{}, stream)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing metadata")
	})

	t.Run("missing authorization", func(t *testing.T) {
		stream := &mockSecretReadServiceListServer{
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs()),
		}
		err := server.List(&emptypb.Empty{}, stream)
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing authorization token")
	})

	t.Run("invalid authorization format", func(t *testing.T) {
		stream := &mockSecretReadServiceListServer{
			ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "BadToken")),
		}
		err := server.List(&emptypb.Empty{}, stream)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authorization token format")
	})

	t.Run("jwt parse error", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return("", errors.New("invalid token"))
		stream := &mockSecretReadServiceListServer{
			ctx: ctx,
		}
		err := server.List(&emptypb.Empty{}, stream)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid token")
	})

	t.Run("reader list error", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return(owner, nil)
		mockReader.EXPECT().List(ctx, owner).Return(nil, errors.New("list failed"))

		stream := &mockSecretReadServiceListServer{
			ctx: ctx,
		}

		err := server.List(&emptypb.Empty{}, stream)
		require.Error(t, err)
		require.Contains(t, err.Error(), "list failed")
	})

	t.Run("send error", func(t *testing.T) {
		mockJWT.EXPECT().Parse(token).Return(owner, nil)
		mockReader.EXPECT().List(ctx, owner).Return([]*models.Secret{secret1}, nil)

		stream := &mockSecretReadServiceListServer{
			ctx: ctx,
			sendFn: func(resp *pb.Secret) error {
				return errors.New("send failed")
			},
		}

		err := server.List(&emptypb.Empty{}, stream)
		require.Error(t, err)
		require.Contains(t, err.Error(), "send failed")
	})
}
