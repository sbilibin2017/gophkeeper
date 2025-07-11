package main

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]string
	}{
		{
			name:     "no flags",
			args:     []string{},
			expected: map[string]string{},
		},
		{
			name: "server-url set",
			args: []string{"--server-url=http://localhost"},
			expected: map[string]string{
				"server-url": "http://localhost",
			},
		},
		{
			name: "interactive true",
			args: []string{"--interactive=true"},
			expected: map[string]string{
				"interactive": "true",
			},
		},
		{
			name: "both flags set",
			args: []string{"--server-url=http://localhost", "--interactive=true"},
			expected: map[string]string{
				"server-url":  "http://localhost",
				"interactive": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.String("server-url", "", "URL сервера")
			fs.Bool("interactive", false, "Интерактивный режим")

			err := fs.Parse(tt.args)
			assert.NoError(t, err)

			flags := parseFlags(fs)
			assert.Equal(t, tt.expected, flags)
		})
	}
}
