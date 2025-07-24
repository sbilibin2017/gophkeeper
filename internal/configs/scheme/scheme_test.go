package scheme

import (
	"testing"
)

func TestGetSchemeFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"HTTP prefix", "http://example.com", HTTP},
		{"HTTPS prefix", "https://example.com", HTTPS},
		{"GRPC prefix", "grpc://service.local", GRPC},
		{"No prefix", "ftp://example.com", ""},
		{"Empty string", "", ""},
		{"Partial match http", "htt://example.com", ""},
		{"HTTPS but uppercase", "HTTPS://example.com", ""}, // case-sensitive
		{"HTTP in the middle", "example.com/http://", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSchemeFromURL(tt.url)
			if got != tt.want {
				t.Errorf("GetSchemeFromURL(%q) = %q; want %q", tt.url, got, tt.want)
			}
		})
	}
}
