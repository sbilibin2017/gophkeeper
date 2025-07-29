package grpc

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Opt — функция, возвращающая grpc.DialOption и ошибку.
type Opt func() (grpc.DialOption, error)

func New(target string, opts ...Opt) (*grpc.ClientConn, error) {
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	for _, opt := range opts {
		dialOpt, err := opt()
		if err != nil {
			return nil, err
		}
		if dialOpt != nil {
			dialOpts = append(dialOpts, dialOpt)
		}
	}

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// RetryPolicy конфигурирует политику повторных попыток.
type RetryPolicy struct {
	Count   int
	Wait    time.Duration
	MaxWait time.Duration
}

// WithRetryPolicy возвращает Opt для настройки политики повторов.
func WithRetryPolicy(rp RetryPolicy) Opt {
	return func() (grpc.DialOption, error) {
		if rp.Count <= 0 && rp.Wait <= 0 && rp.MaxWait <= 0 {
			return nil, nil
		}

		initialBackoff := fmt.Sprintf("%.3fs", rp.Wait.Seconds())
		maxBackoff := fmt.Sprintf("%.3fs", rp.MaxWait.Seconds())

		cfg := fmt.Sprintf(`{
			"methodConfig": [{
				"name": [{"service": ".*"}],
				"retryPolicy": {
					"maxAttempts": %d,
					"initialBackoff": "%s",
					"maxBackoff": "%s",
					"backoffMultiplier": 2,
					"retryableStatusCodes": ["UNAVAILABLE"]
				}
			}]
		}`, rp.Count, initialBackoff, maxBackoff)

		return grpc.WithDefaultServiceConfig(cfg), nil
	}
}
