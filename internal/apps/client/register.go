package client

import (
	"fmt"
	"strings"

	clientHandlers "github.com/sbilibin2017/gophkeeper/internal/handlers/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

// Define variables for mocking that point to the actual functions by default
var (
	registerHTTPFunc = clientHandlers.RegisterHTTP
	registerGRPCFunc = clientHandlers.RegisterGRPC
)

func RegisterRegisterCommand(root *cobra.Command) {
	var (
		username    string
		password    string
		authURL     string
		tlsCertFile string
		tlsKeyFile  string
	)

	cmd := &cobra.Command{
		Use:     "register",
		Short:   "Register a new user",
		Long:    "Register a new user by providing a username, password, and authentication details.",
		Example: `gophkeeper register --username alice --password "S3cr3tPass!" --auth-url https://example.com --tls-client-cert cert.pem --tls-client-key key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp *models.AuthResponse
			var err error

			if strings.HasPrefix(authURL, "grpc://") {
				resp, err = registerGRPCFunc(cmd.Context(), username, password, authURL, tlsCertFile, tlsKeyFile)
			} else if strings.HasPrefix(authURL, "http://") || strings.HasPrefix(authURL, "https://") {
				resp, err = registerHTTPFunc(cmd.Context(), username, password, authURL, tlsCertFile, tlsKeyFile)
			} else {
				return fmt.Errorf("unsupported auth URL scheme, must start with grpc://, http:// or https://")
			}

			if err != nil {
				return fmt.Errorf("registration failed: %w", err)
			}

			cmd.Println(resp.Token)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for registration")
	cmd.Flags().StringVar(&password, "password", "", "Password for registration")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication server URL")
	cmd.Flags().StringVar(&tlsCertFile, "tls-client-cert", "", "Path to TLS client certificate file")
	cmd.Flags().StringVar(&tlsKeyFile, "tls-client-key", "", "Path to TLS client key file")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("auth-url")
	cmd.MarkFlagRequired("tls-client-cert")
	cmd.MarkFlagRequired("tls-client-key")

	root.AddCommand(cmd)
}
