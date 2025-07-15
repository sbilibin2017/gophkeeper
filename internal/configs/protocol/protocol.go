package protocol

import "strings"

const (
	HTTP  = "http"
	HTTPS = "https"
	GRPC  = "grpc"
)

var protocolMap = map[string]string{
	"http://":  HTTP,
	"https://": HTTPS,
	"grpc://":  GRPC,
}

// GetProtocolFromURL определяет тип протокола по префиксу URL.
func GetProtocolFromURL(url string) string {
	for prefix, proto := range protocolMap {
		if strings.HasPrefix(url, prefix) {
			return proto
		}
	}
	return ""
}
