package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/stretchr/testify/require"
)

func TestAddLoginPassword(t *testing.T) {
	tests := []struct {
		name      string
		secret    *models.UsernamePassword
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "success",
			secret: &models.UsernamePassword{
				Username: "user1",
				Password: "pass1",
				Meta:     map[string]string{"foo": "bar"},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO username_passwords`).
					WithArgs("user1", "pass1", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "json marshal error",
			secret: &models.UsernamePassword{
				Username: "user1",
				Password: "pass1",
				Meta:     map[string]string{"foo": string([]byte{0xff, 0xfe, 0xfd})}, // invalid utf8 (simulate error)
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// marshal error will happen before DB call, so no mock expectations
			},
			wantErr: true,
		},
		{
			name: "db exec error",
			secret: &models.UsernamePassword{
				Username: "user1",
				Password: "pass1",
				Meta:     map[string]string{"foo": "bar"},
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO username_passwords`).
					WithArgs("user1", "pass1", sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}

			err = services.AddLoginPassword(context.Background(), db, tt.secret)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}

func TestAddText(t *testing.T) {
	tests := []struct {
		name    string
		text    *models.Text
		mockSet func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			text: &models.Text{
				Content: "hello world",
				Meta:    map[string]string{"lang": "en"},
			},
			mockSet: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO texts`).
					WithArgs("hello world", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "json marshal error",
			text: &models.Text{
				Content: "bad meta",
				Meta:    map[string]string{"bad": string([]byte{0xff, 0xfe})}, // invalid utf8
			},
			mockSet: nil,
			wantErr: true,
		},
		{
			name: "db exec error",
			text: &models.Text{
				Content: "fail insert",
				Meta:    map[string]string{"foo": "bar"},
			},
			mockSet: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO texts`).
					WithArgs("fail insert", sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			if tt.mockSet != nil {
				tt.mockSet(mock)
			}

			err = services.AddText(context.Background(), db, tt.text)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}

func TestAddBinary(t *testing.T) {
	tests := []struct {
		name    string
		bin     *models.Binary
		mockSet func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			bin: &models.Binary{
				Data: []byte{1, 2, 3},
				Meta: map[string]string{"type": "bin"},
			},
			mockSet: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO binaries`).
					WithArgs([]byte{1, 2, 3}, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "json marshal error",
			bin: &models.Binary{
				Data: []byte{1, 2, 3},
				Meta: map[string]string{"bad": string([]byte{0xff})},
			},
			mockSet: nil,
			wantErr: true,
		},
		{
			name: "db exec error",
			bin: &models.Binary{
				Data: []byte{1, 2, 3},
				Meta: map[string]string{"foo": "bar"},
			},
			mockSet: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO binaries`).
					WithArgs([]byte{1, 2, 3}, sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			if tt.mockSet != nil {
				tt.mockSet(mock)
			}

			err = services.AddBinary(context.Background(), db, tt.bin)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}

func TestAddBankCard(t *testing.T) {
	tests := []struct {
		name    string
		card    *models.BankCard
		mockSet func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			card: &models.BankCard{
				Number: "1234567890123456",
				Owner:  "John Doe",
				Expiry: "12/34",
				CVV:    "123",
				Meta:   map[string]string{"note": "personal"},
			},
			mockSet: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO bank_cards`).
					WithArgs("1234567890123456", "John Doe", "12/34", "123", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "json marshal error",
			card: &models.BankCard{
				Number: "1234",
				Owner:  "John",
				Expiry: "01/23",
				CVV:    "999",
				Meta:   map[string]string{"bad": string([]byte{0xff})},
			},
			mockSet: nil,
			wantErr: true,
		},
		{
			name: "db exec error",
			card: &models.BankCard{
				Number: "1234",
				Owner:  "John",
				Expiry: "01/23",
				CVV:    "999",
				Meta:   map[string]string{"foo": "bar"},
			},
			mockSet: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO bank_cards`).
					WithArgs("1234", "John", "01/23", "999", sqlmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			if tt.mockSet != nil {
				tt.mockSet(mock)
			}

			err = services.AddBankCard(context.Background(), db, tt.card)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}
