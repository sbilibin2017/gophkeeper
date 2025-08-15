package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDB(t *testing.T) {
	tests := []struct {
		name      string
		driver    string
		dsn       string
		opts      []Opt
		wantError bool
	}{
		{
			name:      "valid sqlite in-memory",
			driver:    "sqlite",
			dsn:       ":memory:",
			opts:      nil,
			wantError: false,
		},
		{
			name:      "invalid driver",
			driver:    "invalid",
			dsn:       ":memory:",
			opts:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := New(tt.driver, tt.dsn, tt.opts...)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				_ = db.Close()
			}
		})
	}
}

func TestDBOptions(t *testing.T) {
	db, err := New("sqlite", ":memory:",
		WithMaxOpenConns(5),
		WithMaxIdleConns(3),
		WithConnMaxLifetime(10*time.Second),
	)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Проверяем только MaxOpenConns
	assert.Equal(t, 5, db.Stats().MaxOpenConnections)

	// ConnMaxLifetime и MaxIdleConns напрямую проверить через Stats() нельзя, поэтому только что установка не паниковала
	_ = db.Close()
}

func TestDBOptions_Defaults(t *testing.T) {
	db, err := New("sqlite", ":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	_ = db.Close()
}
