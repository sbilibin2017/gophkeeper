package protocol

import (
	"errors"
	"net/url"
	"strings"
)

// Список поддерживаемых протоколов подключения.
const (
	HTTP  = "http"  // HTTP — протокол без шифрования.
	HTTPS = "https" // HTTPS — протокол с шифрованием (SSL/TLS).
	GRPC  = "grpc"  // GRPC — протокол на базе HTTP/2.
)

// GetProtocol разбирает протокол (схему) из переданного адреса.
//
// Принимает строку адреса (например, "https://example.com").
// Возвращает один из поддерживаемых протоколов ("http", "https", "grpc") или ошибку,
// если протокол отсутствует или не поддерживается.
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
