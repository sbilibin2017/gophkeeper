package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUsernamePasswordFromInteractive(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantUser   string
		wantPass   string
		wantErr    bool
		errMessage string
	}{
		{
			name:     "valid input",
			input:    "user1\npass1\n",
			wantUser: "user1",
			wantPass: "pass1",
			wantErr:  false,
		},
		{
			name:       "error on reading username",
			input:      "",
			wantErr:    true,
			errMessage: "EOF",
		},
		{
			name:       "error on reading password",
			input:      "user1\n",
			wantErr:    true,
			errMessage: "EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			up, err := NewUsernamePasswordFromInteractive(r)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, up)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUser, up.Username)
				assert.Equal(t, tt.wantPass, up.Password)
			}
		})
	}
}

func TestNewUsernamePasswordFromArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantUser   string
		wantPass   string
		wantErr    bool
		errMessage string
	}{
		{
			name:     "valid args",
			args:     []string{"user", "pass"},
			wantUser: "user",
			wantPass: "pass",
			wantErr:  false,
		},
		{
			name:       "no args",
			args:       []string{},
			wantErr:    true,
			errMessage: "expected exactly 2 arguments",
		},
		{
			name:       "one arg",
			args:       []string{"user"},
			wantErr:    true,
			errMessage: "expected exactly 2 arguments",
		},
		{
			name:       "three args",
			args:       []string{"user", "pass", "extra"},
			wantErr:    true,
			errMessage: "expected exactly 2 arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			up, err := NewUsernamePasswordFromArgs(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, up)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUser, up.Username)
				assert.Equal(t, tt.wantPass, up.Password)
			}
		})
	}
}

func TestValidateUsernamePassword(t *testing.T) {
	tests := []struct {
		name       string
		input      *UsernamePassword
		wantErr    bool
		errMessage string
	}{
		{
			name: "valid input",
			input: &UsernamePassword{
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "empty username",
			input: &UsernamePassword{
				Username: "",
				Password: "pass",
			},
			wantErr:    true,
			errMessage: "username cannot be empty",
		},
		{
			name: "empty password",
			input: &UsernamePassword{
				Username: "user",
				Password: "",
			},
			wantErr:    true,
			errMessage: "password cannot be empty",
		},
		{
			name: "empty username and password",
			input: &UsernamePassword{
				Username: "",
				Password: "",
			},
			wantErr:    true,
			errMessage: "username cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsernamePassword(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
