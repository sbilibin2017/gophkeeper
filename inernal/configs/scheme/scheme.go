package scheme

import "strings"

const (
	HTTP  = "http"
	HTTPS = "https"
	GRPC  = "grpc"
)

var schemeMap = map[string]string{
	"http://":  HTTP,
	"https://": HTTPS,
	"grpc://":  GRPC,
}

// GetSchemeFromURL determines the protocol type by checking the prefix of the given URL.
// It returns one of the predefined protocol constants ("http", "https", "grpc") if the prefix matches,
// or an empty string if no known prefix is found.
func GetSchemeFromURL(url string) string {
	for prefix, scheme := range schemeMap {
		if strings.HasPrefix(url, prefix) {
			return scheme
		}
	}
	return ""
}
