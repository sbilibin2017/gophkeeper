package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func startBufServer(t *testing.T) {
	lis = bufconn.Listen(bufSize)
	s := gogrpc.NewServer()
	go func() {
		require.NoError(t, s.Serve(lis))
	}()
}

func TestWithToken(t *testing.T) {
	opt := WithAuthToken("abc123")
	dialOpt, err := opt()
	require.NoError(t, err)
	require.NotNil(t, dialOpt)
}

func TestWithToken_Empty(t *testing.T) {
	opt := WithAuthToken("")
	dialOpt, err := opt()
	require.NoError(t, err)
	assert.Nil(t, dialOpt)
}

func TestWithRetryPolicy(t *testing.T) {
	opt := WithRetryPolicy(RetryPolicy{
		Count:   4,
		Wait:    100 * time.Millisecond,
		MaxWait: 300 * time.Millisecond,
	})
	dialOpt, err := opt()
	require.NoError(t, err)
	require.NotNil(t, dialOpt)
}

func TestWithRetryPolicy_Empty(t *testing.T) {
	opt := WithRetryPolicy(RetryPolicy{})
	dialOpt, err := opt()
	require.NoError(t, err)
	assert.Nil(t, dialOpt)
}

func TestNew_WithOptions(t *testing.T) {
	startBufServer(t)

	conn, err := New("bufnet",
		func() (gogrpc.DialOption, error) {
			return gogrpc.WithContextDialer(bufDialer), nil
		},
		WithAuthToken("abc123"),
		WithRetryPolicy(RetryPolicy{
			Count:   2,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, conn)
	conn.Close()
}

func TestNew_ErrorInOption(t *testing.T) {
	errOpt := func() (gogrpc.DialOption, error) {
		return nil, assert.AnError
	}
	conn, err := New("target", errOpt)
	require.Error(t, err)
	assert.Nil(t, conn)
}

func TestTokenAuth_GetRequestMetadata(t *testing.T) {
	token := "sometoken123"
	auth := tokenAuth{token: token}

	md, err := auth.GetRequestMetadata(context.Background(), "someURI")
	assert.NoError(t, err)
	assert.NotNil(t, md)
	assert.Equal(t, "Bearer "+token, md["authorization"])
}

func TestTokenAuth_RequireTransportSecurity(t *testing.T) {
	auth := tokenAuth{token: "anytoken"}

	requireTLS := auth.RequireTransportSecurity()
	assert.False(t, requireTLS) // Обновлено на false, т.к. TLS не нужен
}
