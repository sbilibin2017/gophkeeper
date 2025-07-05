package protocol

import (
	"errors"
	"net/url"
	"strings"
)

// List of supported connection protocols.
const (
	HTTP  = "http"  // HTTP — protocol without encryption.
	HTTPS = "https" // HTTPS — protocol with encryption (SSL/TLS).
	GRPC  = "grpc"  // GRPC — protocol based on HTTP/2.
)

// GetProtocol parses the protocol (scheme) from the given address.
//
// Accepts an address string (for example, "https://example.com").
// Returns one of the supported protocols ("http", "https", "grpc") or an error
// if the protocol is missing or not supported.
func GetProtocol(address string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(address))
	if err != nil {
		return "", err
	}

	switch strings.ToLower(parsed.Scheme) {
	case HTTP:
		return HTTP, nil
	case HTTPS:
		return HTTPS, nil
	case GRPC:
		return GRPC, nil
	default:
		return "", errors.New("unsupported or missing protocol in URL")
	}
}
