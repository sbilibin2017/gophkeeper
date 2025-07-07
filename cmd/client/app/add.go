package app

import (
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

// newAddCommand creates a cobra command for adding new data/secrets
func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add --server-url <url> --type <type> --file <path> --interactive --hmac-key <key> --rsa-public-key <path>",
		Short: "Add new secret to the client from a file or interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _, err := parseAddFlags(cmd)
			return err
		},
	}

	cmd.Flags().StringP("server-url", "s", "", "Server URL")
	cmd.Flags().StringP("type", "t", "", "Secret type")
	cmd.Flags().StringP("file", "f", "", "Input file path")
	cmd.Flags().BoolP("interactive", "i", false, "Enable interactive input mode")
	cmd.Flags().String("hmac-key", "", "HMAC encryption key")
	cmd.Flags().String("rsa-public-key", "", "Path to RSA public key")

	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

func parseAddFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.SecretAddRequest, error) {
	serverURL, _ := cmd.Flags().GetString("server-url")
	stype, _ := cmd.Flags().GetString("type")
	file, _ := cmd.Flags().GetString("file")
	interactive, _ := cmd.Flags().GetBool("interactive")
	hmacKey, _ := cmd.Flags().GetString("hmac-key")
	rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa-public-key")

	if file == "" && !interactive {
		return nil, nil, errors.New("either --file or --interactive must be specified")
	}
	if file != "" && interactive {
		return nil, nil, errors.New("--file and --interactive cannot be used together")
	}

	config, err := configs.NewClientConfig(
		configs.WithClient(serverURL),
		configs.WithHMACEncoder(hmacKey),
		configs.WithRSAEncoder(rsaPublicKeyPath),
	)
	if err != nil {
		return nil, nil, err
	}

	req, err := models.NewSecretAddRequest(
		models.WithServerURL(serverURL),
		models.WithSType(stype),
		models.WithFile(file),
		models.WithInteractive(interactive),
		models.WithHMACKey(hmacKey),
		models.WithRSAPublicKeyPath(rsaPublicKeyPath),
	)
	if err != nil {
		return nil, nil, err
	}

	return config, req, nil
}
