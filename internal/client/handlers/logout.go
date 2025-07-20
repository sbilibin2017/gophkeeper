package handlers

import (
	"context"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/client/config"
	"github.com/sbilibin2017/gophkeeper/internal/client/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

func LogoutHTTP(
	ctx context.Context,
	token, authURL, tlsCertFile, tlsKeyFile string,
) error {
	config, err := config.NewConfig(
		config.WithHTTPClient(authURL, tlsCertFile, tlsKeyFile, token),
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
	config, err := config.NewConfig(
		config.WithGRPCClient(authURL, tlsCertFile, tlsKeyFile, token),
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
