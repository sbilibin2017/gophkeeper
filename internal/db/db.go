package db

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Opt определяет тип функции, которая применяет конфигурацию к sqlx.DB.
type Opt func(*sqlx.DB)

// NewDB устанавливает соединение с базой данных и применяет переданные опции.
// driver — драйвер базы данных (например, "sqlite").
// dsn — строка подключения к базе данных.
// opts — опциональные функции для настройки соединения.
func New(driver string, dsn string, opts ...Opt) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(db)
	}

	return db, nil
}

// WithMaxOpenConns устанавливает максимальное количество открытых соединений.
func WithMaxOpenConns(opts ...int) Opt {
	return func(db *sqlx.DB) {
		for _, opt := range opts {
			if opt > 0 {
				db.SetMaxOpenConns(opt)
				break
			}
		}
	}
}

// WithMaxIdleConns устанавливает максимальное количество простаивающих соединений.
func WithMaxIdleConns(opts ...int) Opt {
	return func(db *sqlx.DB) {
		for _, opt := range opts {
			if opt > 0 {
				db.SetMaxIdleConns(opt)
				break
			}
		}
	}
}

// WithConnMaxLifetime устанавливает максимальное время жизни соединения.
func WithConnMaxLifetime(opts ...time.Duration) Opt {
	return func(db *sqlx.DB) {
		for _, opt := range opts {
			if opt != 0 {
				db.SetConnMaxLifetime(opt)
				break
			}
		}
	}
}
