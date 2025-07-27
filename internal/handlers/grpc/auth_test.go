package grpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/handlers/grpc"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/require"
)

func TestAuthServer_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegisterer := grpc.NewMockRegisterer(ctrl)
	mockLoginer := grpc.NewMockLoginer(ctrl) // needed to create AuthServer, but not used here

	server := grpc.NewAuthServer(mockRegisterer, mockLoginer)
	ctx := context.Background()

	t.Run("success returns token", func(t *testing.T) {
		expectedToken := "token123"
		mockRegisterer.EXPECT().
			Register(ctx, "user", "pass").
			Return(&expectedToken, nil).
			Times(1)

		req := &pb.AuthRequest{
			Username: "user",
			Password: "pass",
		}

		resp, err := server.Register(ctx, req)
		require.NoError(t, err)
		require.Equal(t, expectedToken, resp.GetToken())
	})

	t.Run("success returns empty token when nil", func(t *testing.T) {
		mockRegisterer.EXPECT().
			Register(ctx, "user", "pass").
			Return(nil, nil).
			Times(1)

		req := &pb.AuthRequest{
			Username: "user",
			Password: "pass",
		}

		resp, err := server.Register(ctx, req)
		require.NoError(t, err)
		require.Equal(t, "", resp.GetToken())
	})

	t.Run("returns error from registerer", func(t *testing.T) {
		expectedErr := errors.New("register failed")
		mockRegisterer.EXPECT().
			Register(ctx, "user", "pass").
			Return(nil, expectedErr).
			Times(1)

		req := &pb.AuthRequest{
			Username: "user",
			Password: "pass",
		}

		resp, err := server.Register(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, expectedErr, err)
	})
}

func TestAuthServer_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegisterer := grpc.NewMockRegisterer(ctrl) // needed to create AuthServer, but not used here
	mockLoginer := grpc.NewMockLoginer(ctrl)

	server := grpc.NewAuthServer(mockRegisterer, mockLoginer)
	ctx := context.Background()

	t.Run("success returns token", func(t *testing.T) {
		expectedToken := "token456"
		mockLoginer.EXPECT().
			Login(ctx, "user", "pass").
			Return(&expectedToken, nil).
			Times(1)

		req := &pb.AuthRequest{
			Username: "user",
			Password: "pass",
		}

		resp, err := server.Login(ctx, req)
		require.NoError(t, err)
		require.Equal(t, expectedToken, resp.GetToken())
	})

	t.Run("success returns empty token when nil", func(t *testing.T) {
		mockLoginer.EXPECT().
			Login(ctx, "user", "pass").
			Return(nil, nil).
			Times(1)

		req := &pb.AuthRequest{
			Username: "user",
			Password: "pass",
		}

		resp, err := server.Login(ctx, req)
		require.NoError(t, err)
		require.Equal(t, "", resp.GetToken())
	})

	t.Run("returns error from loginer", func(t *testing.T) {
		expectedErr := errors.New("login failed")
		mockLoginer.EXPECT().
			Login(ctx, "user", "pass").
			Return(nil, expectedErr).
			Times(1)

		req := &pb.AuthRequest{
			Username: "user",
			Password: "pass",
		}

		resp, err := server.Login(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Equal(t, expectedErr, err)
	})
}
