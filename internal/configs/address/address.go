package address

import (
	"errors"
	"strings"
)

// Поддерживаемые схемы адресов.
const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
	SchemeGRPC  = "grpc"
)

// ErrUnsupportedScheme возвращается, если адрес использует неизвестную или неподдерживаемую схему.
var ErrUnsupportedScheme = errors.New("unsupported address scheme")

// Address хранит схему и сам сетевой адрес.
type Address struct {
	Scheme  string // Схема адреса (http, https, grpc)
	Address string // Сетевой адрес (host:port или :port)
}

// New разбирает полный входной адрес и возвращает структуру Address с разделёнными схемой и адресом.
// Если схема не указана, по умолчанию используется "http".
// Примеры:
//
//	New(":8080")          -> Scheme: "http", Address: ":8080"
//	New("localhost:8080") -> Scheme: "http", Address: "localhost:8080"
//	New("https://example.com:443") -> Scheme: "https", Address: "example.com:443"
func New(input string) Address {
	scheme := SchemeHTTP
	addr := input

	// Проверяем наличие схемы в формате "scheme://"
	if idx := strings.Index(input, "://"); idx != -1 {
		prefix := input[:idx]
		switch prefix {
		case SchemeHTTP, SchemeHTTPS, SchemeGRPC:
			scheme = prefix
		default:
			// Для неизвестных схем оставляем как есть
			scheme = prefix
		}
		addr = input[idx+3:] // убираем "://"
	}

	return Address{
		Scheme:  scheme,
		Address: addr,
	}
}

// String реализует интерфейс fmt.Stringer и возвращает полный адрес в формате "scheme://address".
func (a Address) String() string {
	return a.Scheme + "://" + a.Address
}
