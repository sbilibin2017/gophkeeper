package fields

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringMap struct {
	Map map[string]string
}

// Scan implements sql.Scanner for reading from DB
func (s *StringMap) Scan(value any) error {
	if value == nil {
		s.Map = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan StringMap: expected []byte, got %T", value)
	}
	m := make(map[string]string)
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}
	s.Map = m
	return nil
}

// Value implements driver.Valuer for writing to DB
func (s StringMap) Value() (driver.Value, error) {
	if s.Map == nil {
		return nil, nil
	}
	return json.Marshal(s.Map)
}

// helper to parse meta
func ParseMeta(meta string) (*StringMap, error) {
	if meta == "" {
		return nil, nil
	}
	var metaMap map[string]string
	if err := json.Unmarshal([]byte(meta), &metaMap); err != nil {
		return nil, fmt.Errorf("failed to parse meta JSON: %w", err)
	}
	sm := StringMap{Map: metaMap}
	return &sm, nil
}
