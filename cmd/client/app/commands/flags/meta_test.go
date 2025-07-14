package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareMetaJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      *string
		expectErr bool
	}{
		{
			name:      "empty input returns nil",
			input:     "",
			want:      nil,
			expectErr: false,
		},
		{
			name:      "valid JSON returns formatted JSON string",
			input:     `{"key":"value"}`,
			want:      ptr(`{"key":"value"}`),
			expectErr: false,
		},
		{
			name:      "valid JSON with multiple keys",
			input:     `{"a":"1","b":"2"}`,
			want:      ptr(`{"a":"1","b":"2"}`),
			expectErr: false,
		},
		{
			name:      "invalid JSON returns error",
			input:     `{"key":}`,
			want:      nil,
			expectErr: true,
		},
		{
			name:      "valid JSON with extra whitespace",
			input:     ` {  "key" : "value" } `,
			want:      ptr(`{"key":"value"}`),
			expectErr: false,
		},
		{
			name:      "not a JSON object",
			input:     `"just a string"`,
			want:      nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrepareMetaJSON(tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

// helper to get *string
func ptr(s string) *string {
	return &s
}
