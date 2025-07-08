package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAddLoginPassword(t *testing.T) {
	ctx := context.Background()

	testEncoder := func(data []byte) ([]byte, error) {
		return append(data, []byte("-enc")...), nil
	}
	errorEncoder := func(data []byte) ([]byte, error) {
		return nil, errors.New("encode error")
	}

	tests := []struct {
		name         string
		encoders     []func([]byte) ([]byte, error)
		secret       *models.LoginPassword
		mockDBErr    error
		expectErr    bool
		encoderError bool
	}{
		{
			name:     "success with encoder",
			encoders: []func([]byte) ([]byte, error){testEncoder},
			secret: &models.LoginPassword{
				SecretID: "id1",
				Login:    "login1",
				Password: "pass1",
				Meta:     map[string]string{"foo": "bar"},
			},
			expectErr: false,
		},
		{
			name:     "success without encoder",
			encoders: nil,
			secret: &models.LoginPassword{
				SecretID: "id2",
				Login:    "login2",
				Password: "pass2",
				Meta:     map[string]string{"k": "v"},
			},
			expectErr: false,
		},
		{
			name:         "encoder returns error",
			encoders:     []func([]byte) ([]byte, error){errorEncoder},
			secret:       &models.LoginPassword{SecretID: "id3", Login: "login3", Password: "pass3"},
			expectErr:    true,
			encoderError: true,
		},
		{
			name:     "db exec error",
			encoders: []func([]byte) ([]byte, error){testEncoder},
			secret: &models.LoginPassword{
				SecretID: "id4",
				Login:    "login4",
				Password: "pass4",
				Meta:     map[string]string{},
			},
			mockDBErr: errors.New("db error"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")

			if !tt.encoderError {
				if tt.mockDBErr != nil {
					mock.ExpectExec(`INSERT INTO login_passwords`).
						WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnError(tt.mockDBErr)
				} else {
					mock.ExpectExec(`INSERT INTO login_passwords`).
						WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			err = AddLoginPassword(ctx, tt.secret,
				WithAddLoginPasswordEncoders(tt.encoders),
				WithAddLoginPasswordDB(sqlxDB),
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddText(t *testing.T) {
	ctx := context.Background()

	testEncoder := func(data []byte) ([]byte, error) {
		return append(data, []byte("-enc")...), nil
	}
	errorEncoder := func(data []byte) ([]byte, error) {
		return nil, errors.New("encode error")
	}

	tests := []struct {
		name         string
		encoders     []func([]byte) ([]byte, error)
		secret       *models.Text
		mockDBErr    error
		expectErr    bool
		encoderError bool
	}{
		{
			name:     "success with encoder",
			encoders: []func([]byte) ([]byte, error){testEncoder},
			secret: &models.Text{
				SecretID:  "id1",
				Content:   "content1",
				Meta:      map[string]string{"foo": "bar"},
				UpdatedAt: time.Now(),
			},
			expectErr: false,
		},
		{
			name:     "success without encoder",
			encoders: nil,
			secret: &models.Text{
				SecretID:  "id2",
				Content:   "content2",
				Meta:      map[string]string{"k": "v"},
				UpdatedAt: time.Now(),
			},
			expectErr: false,
		},
		{
			name:         "encoder error",
			encoders:     []func([]byte) ([]byte, error){errorEncoder},
			secret:       &models.Text{SecretID: "id3", Content: "content3"},
			expectErr:    true,
			encoderError: true,
		},
		{
			name:     "db exec error",
			encoders: []func([]byte) ([]byte, error){testEncoder},
			secret: &models.Text{
				SecretID:  "id4",
				Content:   "content4",
				Meta:      map[string]string{},
				UpdatedAt: time.Now(),
			},
			mockDBErr: errors.New("db error"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")

			if !tt.encoderError {
				if tt.mockDBErr != nil {
					mock.ExpectExec(`INSERT INTO texts`).
						WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnError(tt.mockDBErr)
				} else {
					mock.ExpectExec(`INSERT INTO texts`).
						WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			err = AddText(ctx, tt.secret,
				WithAddTextEncoders(tt.encoders),
				WithAddTextDB(sqlxDB),
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddCard(t *testing.T) {
	ctx := context.Background()

	testEncoder := func(data []byte) ([]byte, error) {
		return append(data, []byte("-enc")...), nil
	}
	errorEncoder := func(data []byte) ([]byte, error) {
		return nil, errors.New("encode error")
	}

	tests := []struct {
		name         string
		encoders     []func([]byte) ([]byte, error)
		secret       *models.Card
		mockDBErr    error
		expectErr    bool
		encoderError bool
	}{
		{
			name:     "success with encoder",
			encoders: []func([]byte) ([]byte, error){testEncoder},
			secret: &models.Card{
				SecretID:  "id1",
				Number:    "1234",
				Holder:    "holder1",
				ExpMonth:  12,
				ExpYear:   2025,
				CVV:       "123",
				Meta:      map[string]string{"foo": "bar"},
				UpdatedAt: time.Now(),
			},
			expectErr: false,
		},
		{
			name:     "success without encoder",
			encoders: nil,
			secret: &models.Card{
				SecretID:  "id2",
				Number:    "5678",
				Holder:    "holder2",
				ExpMonth:  6,
				ExpYear:   2024,
				CVV:       "456",
				Meta:      map[string]string{"k": "v"},
				UpdatedAt: time.Now(),
			},
			expectErr: false,
		},
		{
			name:         "encoder error",
			encoders:     []func([]byte) ([]byte, error){errorEncoder},
			secret:       &models.Card{SecretID: "id3"},
			expectErr:    true,
			encoderError: true,
		},
		{
			name:     "db exec error",
			encoders: []func([]byte) ([]byte, error){testEncoder},
			secret: &models.Card{
				SecretID:  "id4",
				Number:    "0000",
				Holder:    "holder4",
				ExpMonth:  1,
				ExpYear:   2023,
				CVV:       "000",
				Meta:      map[string]string{},
				UpdatedAt: time.Now(),
			},
			mockDBErr: errors.New("db error"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")

			if !tt.encoderError {
				if tt.mockDBErr != nil {
					mock.ExpectExec(`INSERT INTO cards`).
						WithArgs(
							sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
							sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
							sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnError(tt.mockDBErr)
				} else {
					mock.ExpectExec(`INSERT INTO cards`).
						WithArgs(
							sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
							sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
							sqlmock.AnyArg(), sqlmock.AnyArg()).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
			}

			err = AddCard(ctx, tt.secret,
				WithAddCardEncoders(tt.encoders),
				WithAddCardDB(sqlxDB),
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddBinary(t *testing.T) {
	ctx := context.Background()

	testEncoder := func(data []byte) ([]byte, error) {
		return append(data, []byte("-enc")...), nil
	}
	errorEncoder := func(data []byte) ([]byte, error) {
		return nil, errors.New("encode error")
	}

	now := time.Now()

	tests := []struct {
		name         string
		encoders     []func([]byte) ([]byte, error)
		secret       *models.Binary
		mockDBErr    error
		expectErr    bool
		encoderError bool
	}{
		{
			name:     "success with encoder",
			encoders: []func([]byte) ([]byte, error){testEncoder},
			secret: &models.Binary{
				SecretID:  "id1",
				Data:      []byte("some binary data"),
				Meta:      map[string]string{"foo": "bar"},
				UpdatedAt: now,
			},
			expectErr: false,
		},
		{
			name:     "success without encoder",
			encoders: nil,
			secret: &models.Binary{
				SecretID:  "id2",
				Data:      []byte("raw data"),
				Meta:      map[string]string{"k": "v"},
				UpdatedAt: now,
			},
			expectErr: false,
		},
		{
			name:         "encoder returns error",
			encoders:     []func([]byte) ([]byte, error){errorEncoder},
			secret:       &models.Binary{SecretID: "id3", Data: []byte("data")},
			expectErr:    true,
			encoderError: true,
		},
		{
			name:     "db exec error",
			encoders: []func([]byte) ([]byte, error){testEncoder},
			secret: &models.Binary{
				SecretID:  "id4",
				Data:      []byte("data4"),
				Meta:      map[string]string{},
				UpdatedAt: now,
			},
			mockDBErr: errors.New("db error"),
			expectErr: true,
		},
		{
			name:      "db not configured",
			secret:    &models.Binary{SecretID: "id5", Data: []byte("data")},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sqlxDB *sqlx.DB
			var mock sqlmock.Sqlmock
			var err error

			// Setup mock DB only if DB is required (not "db not configured" test)
			if tt.name != "db not configured" {
				var db *sql.DB
				db, mock, err = sqlmock.New()
				assert.NoError(t, err)
				defer db.Close()
				sqlxDB = sqlx.NewDb(db, "sqlmock")
			}

			if tt.mockDBErr != nil {
				mock.ExpectExec(`INSERT INTO binaries`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(tt.mockDBErr)
			} else if !tt.encoderError && tt.name != "db not configured" {
				mock.ExpectExec(`INSERT INTO binaries`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err = AddBinary(ctx, tt.secret,
				WithAddBinaryEncoders(tt.encoders),
				WithAddBinaryDB(sqlxDB),
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if mock != nil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
