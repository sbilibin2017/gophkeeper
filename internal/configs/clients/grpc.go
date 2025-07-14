package clients

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// NewGRPCClient создает gRPC подключение и возвращает *grpc.ClientConn
func NewGRPCClient(target string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // вместо WithInsecure()
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
