package validation

import (
	"fmt"
	"testing"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateResolveStrategy(t *testing.T) {
	tests := []struct {
		name           string
		strategy       string
		wantErr        bool
		expectedErrMsg string
	}{
		{"Valid: server", models.ResolveStrategyServer, false, ""},
		{"Valid: client", models.ResolveStrategyClient, false, ""},
		{"Valid: interactive", models.ResolveStrategyInteractive, false, ""},
		{"Invalid strategy", "invalid", true,
			fmt.Sprintf("unsupported resolve strategy: %q, must be one of [%q, %q, %q]",
				"invalid",
				models.ResolveStrategyServer,
				models.ResolveStrategyClient,
				models.ResolveStrategyInteractive)},
		{"Empty string", "", true,
			fmt.Sprintf("unsupported resolve strategy: %q, must be one of [%q, %q, %q]",
				"",
				models.ResolveStrategyServer,
				models.ResolveStrategyClient,
				models.ResolveStrategyInteractive)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResolveStrategy(tt.strategy)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
