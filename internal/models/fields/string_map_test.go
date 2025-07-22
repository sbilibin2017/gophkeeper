package fields

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringMap_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "valid JSON map",
			input:    []byte(`{"key1":"val1","key2":"val2"}`),
			expected: map[string]string{"key1": "val1", "key2": "val2"},
			wantErr:  false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{invalid json}`),
			wantErr: true,
		},
		{
			name:    "non-byte input",
			input:   "not bytes",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sm StringMap
			err := sm.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, sm.Map)
			}
		})
	}
}

func TestStringMap_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    StringMap
		expected driver.Value
		wantErr  bool
	}{
		{
			name:     "nil map",
			input:    StringMap{Map: nil},
			expected: nil,
			wantErr:  false,
		},
		{
			name:  "non-empty map",
			input: StringMap{Map: map[string]string{"a": "1", "b": "2"}},
			expected: func() driver.Value {
				b, _ := json.Marshal(map[string]string{"a": "1", "b": "2"})
				return b
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.input.Value()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, val)
			}
		})
	}
}
