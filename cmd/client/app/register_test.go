package app

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRegisterCommand_ValidFlags(t *testing.T) {
	cmd := newRegisterCommand()
	cmd.SetArgs([]string{
		"--server-url", "https://localhost:8000",
		"--username", "testuser",
		"--password", "testpass",
		"--rsa-public-key-path", "", // важно: флаг должен быть зарегистрирован
	})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		flags       map[string]string
		expectError bool
		errorSubstr string
		username    string
		password    string
		serverURL   string
	}{
		{
			name: "valid flags",
			flags: map[string]string{
				"server-url":          "https://localhost:8000",
				"username":            "testuser",
				"password":            "testpass",
				"rsa-public-key-path": "",
			},
			expectError: false,
			username:    "testuser",
			password:    "testpass",
			serverURL:   "https://localhost:8000",
		},
		{
			name: "missing server-url",
			flags: map[string]string{
				"username": "user",
				"password": "pass",
			},
			expectError: true,
			errorSubstr: "server-url is required",
		},
		{
			name: "missing username",
			flags: map[string]string{
				"server-url": "https://localhost:8000",
				"password":   "pass",
			},
			expectError: true,
			errorSubstr: "username is required",
		},
		{
			name: "missing password",
			flags: map[string]string{
				"server-url": "https://localhost:8000",
				"username":   "user",
			},
			expectError: true,
			errorSubstr: "password is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("server-url", "", "")
			cmd.Flags().String("username", "", "")
			cmd.Flags().String("password", "", "")
			cmd.Flags().String("rsa-public-key-path", "", "") // ПРАВИЛЬНОЕ имя
			cmd.Flags().String("hmac-key", "", "")

			for k, v := range tc.flags {
				_ = cmd.Flags().Set(k, v)
			}

			cfg, creds, err := parseFlags(cmd)

			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorSubstr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.username, creds.Username)
				require.Equal(t, tc.password, creds.Password)
				require.Equal(t, tc.serverURL, cfg.ServerURL)
			}
		})
	}
}

func TestNewRegisterService(t *testing.T) {
	tests := []struct {
		name         string
		serverURL    string
		expectError  bool
		expectedType string
	}{
		{
			name:         "http protocol",
			serverURL:    "http://localhost:8000",
			expectError:  false,
			expectedType: "*services.RegisterHTTPService",
		},
		{
			name:         "https protocol",
			serverURL:    "https://localhost:8000",
			expectError:  false,
			expectedType: "*services.RegisterHTTPService",
		},
		{
			name:         "grpc protocol",
			serverURL:    "grpc://localhost:8000",
			expectError:  false,
			expectedType: "*services.RegisterGRPCService",
		},
		{
			name:        "unknown protocol",
			serverURL:   "ftp://localhost:8000",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := newRegisterService(tc.serverURL, "", "")

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, svc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, svc)
				actual := getPrivateContext(svc)
				require.Equal(t, tc.expectedType, typeName(actual))
			}
		})
	}
}

// getPrivateContext извлекает приватное поле 'context' из RegisterService.
func getPrivateContext(s *services.RegisterService) services.Registerer {
	val := reflect.ValueOf(s).Elem().FieldByName("context")
	ptr := unsafe.Pointer(val.UnsafeAddr())
	realVal := reflect.NewAt(val.Type(), ptr).Elem()
	return realVal.Interface().(services.Registerer)
}

// typeName возвращает строковое представление типа объекта.
func typeName(v interface{}) string {
	return reflect.TypeOf(v).String()
}
