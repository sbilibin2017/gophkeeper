package client

import (
	"net/url"

	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewHTTPClient creates a new HTTP client with the specified base URL.
func NewHTTPClient(baseURL string) (*resty.Client, error) {
	_, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	client := resty.New().SetBaseURL(baseURL)

	return client, nil
}

// NewGRPCClient creates a new gRPC connection to the server at baseURL.
func NewGRPCClient(baseURL string) (*grpc.ClientConn, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	client, err := grpc.NewClient(
		parsed.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
