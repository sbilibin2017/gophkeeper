package address

import (
	"errors"
	"strings"
)

// Поддерживаемые схемы.
const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
	SchemeGRPC  = "grpc"
)

// ErrUnsupportedScheme возвращается, когда адрес использует неизвестную или неподдерживаемую схему.
var ErrUnsupportedScheme = errors.New("неподдерживаемая схема адреса")

// Address хранит схему и фактический сетевой адрес.
type Address struct {
	Scheme  string
	Address string
}

// New парсит полный входной адрес и возвращает разделённые схему и адрес.
// Если схема не указана, используется "http" по умолчанию.
func New(input string) (*Address, error) {
	addr := input
	scheme := SchemeHTTP

	switch {
	case strings.HasPrefix(addr, SchemeHTTP+"://"):
		scheme = SchemeHTTP
		addr = strings.TrimPrefix(addr, SchemeHTTP+"://")
	case strings.HasPrefix(addr, SchemeHTTPS+"://"):
		scheme = SchemeHTTPS
		addr = strings.TrimPrefix(addr, SchemeHTTPS+"://")
	case strings.HasPrefix(addr, SchemeGRPC+"://"):
		scheme = SchemeGRPC
		addr = strings.TrimPrefix(addr, SchemeGRPC+"://")
	case strings.Contains(addr, "://"):
		return nil, ErrUnsupportedScheme
	}

	if addr == "" || strings.HasPrefix(addr, ":") {
		addr = "0.0.0.0" + addr
	}

	return &Address{
		Scheme:  scheme,
		Address: addr,
	}, nil
}

// String реализует интерфейс fmt.Stringer и возвращает полный адрес.
func (a Address) String() string {
	return a.Scheme + "://" + a.Address
}
