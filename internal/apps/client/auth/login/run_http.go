package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"
)

// NewRunHTTP returns a closure function that performs user authentication
// via an HTTP client using the specified authentication server URL and TLS certificate files.
//
// Parameters:
//   - authURL: The full URL of the authentication HTTP server (e.g., "https://auth.example.com").
//   - tlsCertFile: Path to the TLS certificate file for the client.
//   - tlsKeyFile: Path to the TLS key file for the client.
//
// The returned function accepts:
//   - ctx: Context for managing request lifecycle (e.g., timeouts, cancellations).
//   - username: The username of the user attempting to authenticate.
//   - password: The user's password.
//
// Returns:
//   - *models.AuthResponse: Contains access and refresh tokens if authentication is successful.
//   - error: Any error encountered during database setup, HTTP client creation, or login process.
func NewRunHTTP(authURL, tlsCertFile, tlsKeyFile string) func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
	return func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
		conn, err := db.NewDB("sqlite", "client.db")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer conn.Close()

		if err := bankcard.CreateClientTable(ctx, conn); err != nil {
			return nil, err
		}
		if err := text.CreateClientTable(ctx, conn); err != nil {
			return nil, err
		}
		if err := binary.CreateClientTable(ctx, conn); err != nil {
			return nil, err
		}
		if err := user.CreateClientTable(ctx, conn); err != nil {
			return nil, err
		}

		client, err := http.New(
			authURL,
			http.WithTLSCert(http.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
			http.WithRetryPolicy(http.RetryPolicy{Count: 3, Wait: 2 * time.Second}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP client: %w", err)
		}

		facade := auth.NewLoginHTTPFacade(client)

		authReq := &models.AuthRequest{
			Username: username,
			Password: password,
		}

		resp, err := facade.Login(ctx, authReq)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}
