package auth

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"google.golang.org/protobuf/types/known/emptypb"
)

// LogoutHTTPFacade handles HTTP-based logout.
type LogoutHTTPFacade struct {
	client *resty.Client
}

// NewLogoutHTTPFacade returns a new instance of LogoutHTTPFacade.
func NewLogoutHTTPFacade(client *resty.Client) *LogoutHTTPFacade {
	return &LogoutHTTPFacade{client: client}
}

// Logout performs logout via HTTP.
func (f *LogoutHTTPFacade) Logout(ctx context.Context) error {
	resp, err := f.client.R().
		SetContext(ctx).
		Post("/logout")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("logout failed with status: %s", resp.Status())
	}

	return nil
}

// LogoutGRPCFacade handles gRPC-based logout.
type LogoutGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewLogoutGRPCFacade returns a new instance of LogoutGRPCFacade.
func NewLogoutGRPCFacade(client pb.AuthServiceClient) *LogoutGRPCFacade {
	return &LogoutGRPCFacade{client: client}
}

// Logout performs logout via gRPC.
func (f *LogoutGRPCFacade) Logout(ctx context.Context) error {
	_, err := f.client.Logout(ctx, &emptypb.Empty{})
	return err
}
