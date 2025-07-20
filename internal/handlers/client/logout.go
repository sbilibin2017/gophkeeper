package client

import (
	"context"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

func LogoutHTTP(
	ctx context.Context,
	token, authURL, tlsCertFile, tlsKeyFile string,
) error {
	config, err := configs.NewClientConfig(
		configs.WithHTTPClient(authURL, tlsCertFile, tlsKeyFile, token),
	)
	if err != nil {
		return err
	}
	if config.HTTPClient == nil {
		return fmt.Errorf("HTTP client is not configured for URL: %s", authURL)
	}

	facade := auth.NewLogoutHTTPFacade(config.HTTPClient)

	err = facade.Logout(ctx)
	if err != nil {
		return err
	}

	bankcard.DropClientTable(ctx, config.DB)
	text.DropClientTable(ctx, config.DB)
	binary.DropClientTable(ctx, config.DB)
	user.DropClientTable(ctx, config.DB)

	return nil
}

func LogoutGRPC(
	ctx context.Context,
	token, authURL, tlsCertFile, tlsKeyFile string,
) error {
	config, err := configs.NewClientConfig(
		configs.WithGRPCClient(authURL, tlsCertFile, tlsKeyFile, token),
	)
	if err != nil {
		return err
	}
	if config.GRPCClient == nil {
		return fmt.Errorf("gRPC client is not configured for URL: %s", authURL)
	}

	grpcClient := pb.NewAuthServiceClient(config.GRPCClient)
	facade := auth.NewLogoutGRPCFacade(grpcClient)

	err = facade.Logout(ctx)
	if err != nil {
		return err
	}

	bankcard.DropClientTable(ctx, config.DB)
	text.DropClientTable(ctx, config.DB)
	binary.DropClientTable(ctx, config.DB)
	user.DropClientTable(ctx, config.DB)

	return nil
}
