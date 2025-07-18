package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateBankCardExp(t *testing.T) {
	now := time.Now()

	// Helper to format future or past dates
	futureMonth := now.AddDate(0, 1, 0).Format("01/06")  // MM/YY, 1 month ahead
	futureYear := now.AddDate(0, 2, 0).Format("01/2006") // MM/YYYY, 2 months ahead
	pastMonth := now.AddDate(0, -1, 0).Format("01/06")   // MM/YY, 1 month before
	pastYear := now.AddDate(0, -2, 0).Format("01/2006")  // MM/YYYY, 2 months before

	tests := []struct {
		name    string
		exp     string
		wantErr bool
		errMsg  string
	}{
		{"Valid MM/YY future", futureMonth, false, ""},
		{"Valid MM/YYYY future", futureYear, false, ""},
		{"Empty string", "", true, "expiration date must not be empty"},
		{"Invalid format", "13/25", true, "expiration date must be in MM/YY or MM/YYYY format"},
		{"Invalid format no slash", "1225", true, "expiration date must be in MM/YY or MM/YYYY format"},
		{"Invalid month", "00/25", true, "expiration date must be in MM/YY or MM/YYYY format"},
		{"Invalid month long year", "00/2025", true, "expiration date must be in MM/YY or MM/YYYY format"},
		{"Invalid year short", "12/2a", true, "expiration date must be in MM/YY or MM/YYYY format"},
		{"Invalid year long", "12/20b5", true, "expiration date must be in MM/YY or MM/YYYY format"},
		{"Expired MM/YY", pastMonth, true, "expiration date must be in the future"},
		{"Expired MM/YYYY", pastYear, true, "expiration date must be in the future"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBankCardExp(tt.exp)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
