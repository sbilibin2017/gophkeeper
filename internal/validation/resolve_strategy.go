package validation

import (
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// ValidateResolveStrategy checks that resolveStrategy is one of the allowed constants.
func ValidateResolveStrategy(resolveStrategy string) error {
	switch resolveStrategy {
	case models.ResolveStrategyServer,
		models.ResolveStrategyClient,
		models.ResolveStrategyInteractive:
		return nil
	default:
		return fmt.Errorf("unsupported resolve strategy: %q, must be one of [%q, %q, %q]",
			resolveStrategy,
			models.ResolveStrategyServer,
			models.ResolveStrategyClient,
			models.ResolveStrategyInteractive)
	}
}
