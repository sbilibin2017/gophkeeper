package validation

import (
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func ValidateSecretType(t string) error {
	switch t {
	case models.SecretTypeBankCard, models.SecretTypeBinary, models.SecretTypeText, models.SecretTypeUsernamePassword:
		return nil
	default:
		return fmt.Errorf("unknown secret type: %q", t)
	}
}
