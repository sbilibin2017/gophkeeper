package flags

import (
	"fmt"
	"strings"
)

type MetaFlag map[string]string

func (m *MetaFlag) String() string {
	var parts []string
	for k, v := range *m {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ", ")
}

func (m *MetaFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid meta format: %s (expected key=value)", value)
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if key == "" {
		return fmt.Errorf("invalid meta format: key is empty")
	}
	if *m == nil {
		*m = make(map[string]string)
	}
	(*m)[key] = val
	return nil
}

func (m *MetaFlag) Type() string {
	return "meta"
}
