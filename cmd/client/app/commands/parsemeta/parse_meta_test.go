package parsemeta

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMeta(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected map[string]string
	}{
		{
			name:     "empty input",
			input:    []string{},
			expected: map[string]string{},
		},
		{
			name:  "single key-value",
			input: []string{"key=value"},
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			name:  "multiple key-values",
			input: []string{"k1=v1", "k2 = v2", "k3= v3"},
			expected: map[string]string{
				"k1": "v1",
				"k2": "v2",
				"k3": "v3",
			},
		},
		{
			name:  "line without equal sign",
			input: []string{"key=value", "invalidline", "another=val"},
			expected: map[string]string{
				"key":     "value",
				"another": "val",
			},
		},
		{
			name:  "empty key",
			input: []string{"=value", "key2=val2"},
			expected: map[string]string{
				"key2": "val2",
			},
		},
		{
			name:  "spaces around keys and values",
			input: []string{"  key1  =  val1  ", "  key2=val2"},
			expected: map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
		},
		{
			name:  "empty value",
			input: []string{"k1="},
			expected: map[string]string{
				"k1": "",
			},
		},
		{
			name:  "multiple equals",
			input: []string{"k1=v1=extra"},
			expected: map[string]string{
				"k1": "v1=extra",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMeta(tt.input)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestParseMetaInteractive(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMeta    map[string]string
		expectError bool
	}{
		{
			name:     "empty input",
			input:    "\n",
			wantMeta: map[string]string{},
		},
		{
			name:  "single key-value",
			input: "key=value\n\n",
			wantMeta: map[string]string{
				"key": "value",
			},
		},
		{
			name:  "multiple key-values",
			input: "k1=v1\nk2 = v2\nk3= v3\n\n",
			wantMeta: map[string]string{
				"k1": "v1",
				"k2": "v2",
				"k3": "v3",
			},
		},
		{
			name:  "line without equal sign",
			input: "key=value\ninvalidline\nanother=val\n\n",
			wantMeta: map[string]string{
				"key":     "value",
				"another": "val",
			},
		},
		{
			name:  "empty key",
			input: "=value\nkey2=val2\n\n",
			wantMeta: map[string]string{
				"key2": "val2",
			},
		},
		{
			name:  "spaces around keys and values",
			input: "  key1  =  val1  \n  key2=val2\n\n",
			wantMeta: map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
		},
		{
			name:  "input with EOF instead of newline",
			input: "foo=bar",
			wantMeta: map[string]string{
				"foo": "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			gotMeta, err := ParseMetaInteractive(reader)

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantMeta, gotMeta)
		})
	}
}
