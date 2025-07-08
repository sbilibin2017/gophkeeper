package app

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setFlags(cmd *cobra.Command, flags map[string]string) error {
	for k, v := range flags {
		if err := cmd.Flags().Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

// ==== Тесты для parseAddLoginPasswordFlags ====

func TestParseAddLoginPasswordFlags_Flags(t *testing.T) {
	cmd := newAddLoginPasswordCommand()
	err := setFlags(cmd, map[string]string{
		"secret_id":   "id123",
		"login":       "user1",
		"password":    "pass1",
		"meta":        "k1=v1,k2=v2",
		"interactive": "false",
	})
	require.NoError(t, err)

	config, req, err := parseAddLoginPasswordFlags(cmd)
	require.NoError(t, err)
	assert.Equal(t, "id123", req.SecretID)
	assert.Equal(t, "user1", req.Login)
	assert.Equal(t, "pass1", req.Password)
	assert.Equal(t, map[string]string{"k1": "v1", "k2": "v2"}, req.Meta)
	assert.NotNil(t, config)
}

func TestParseAddLoginPasswordFlags_MissingRequired(t *testing.T) {
	// Нет secret_id
	{
		cmd := newAddLoginPasswordCommand()
		err := setFlags(cmd, map[string]string{
			"login":    "login",
			"password": "pass",
		})
		require.NoError(t, err)
		_, _, err = parseAddLoginPasswordFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret_id required")
	}

	// Нет login
	{
		cmd := newAddLoginPasswordCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
			"password":  "pass",
		})
		require.NoError(t, err)
		_, _, err = parseAddLoginPasswordFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "login required")
	}

	// Нет password
	{
		cmd := newAddLoginPasswordCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
			"login":     "login",
		})
		require.NoError(t, err)
		_, _, err = parseAddLoginPasswordFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password required")
	}
}

// ==== Тесты для parseAddTextFlags ====

func TestParseAddTextFlags_Flags(t *testing.T) {
	cmd := newAddTextSecretCommand()
	err := setFlags(cmd, map[string]string{
		"secret_id":   "textid",
		"content":     "some content",
		"meta":        "a=b,c=d",
		"interactive": "false",
	})
	require.NoError(t, err)

	config, req, err := parseAddTextFlags(cmd)
	require.NoError(t, err)
	assert.Equal(t, "textid", req.SecretID)
	assert.Equal(t, "some content", req.Content)
	assert.Equal(t, map[string]string{"a": "b", "c": "d"}, req.Meta)
	assert.NotNil(t, config)
}

func TestParseAddTextFlags_MissingRequired(t *testing.T) {
	// Нет secret_id
	{
		cmd := newAddTextSecretCommand()
		err := setFlags(cmd, map[string]string{
			"content": "text",
		})
		require.NoError(t, err)
		_, _, err = parseAddTextFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret_id required")
	}

	// Нет content
	{
		cmd := newAddTextSecretCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
		})
		require.NoError(t, err)
		_, _, err = parseAddTextFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content required")
	}
}

// ==== Тесты для parseAddBinaryFlags ====

func TestParseAddBinaryFlags_Flags(t *testing.T) {
	// Создаём временный файл с данными
	content := []byte("binary data test")
	tmpfile, err := os.CreateTemp("", "testbinary")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	cmd := newAddBinarySecretCommand()
	err = setFlags(cmd, map[string]string{
		"secret_id":   "binid",
		"file":        tmpfile.Name(),
		"meta":        "x=y",
		"interactive": "false",
	})
	require.NoError(t, err)

	config, req, err := parseAddBinaryFlags(cmd)
	require.NoError(t, err)
	assert.Equal(t, "binid", req.SecretID)
	assert.Equal(t, content, req.Data)
	assert.Equal(t, map[string]string{"x": "y"}, req.Meta)
	assert.NotNil(t, config)
}

func TestParseAddBinaryFlags_MissingRequired(t *testing.T) {
	// Нет secret_id
	{
		cmd := newAddBinarySecretCommand()
		err := setFlags(cmd, map[string]string{
			"file": "somefile",
		})
		require.NoError(t, err)
		_, _, err = parseAddBinaryFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret_id required")
	}

	// Нет file
	{
		cmd := newAddBinarySecretCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
		})
		require.NoError(t, err)
		_, _, err = parseAddBinaryFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file path required")
	}

	// Ошибка чтения файла
	{
		cmd := newAddBinarySecretCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
			"file":      "nonexistent.file",
		})
		require.NoError(t, err)
		_, _, err = parseAddBinaryFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file read error")
	}
}

// ==== Тесты для parseAddCardFlags ====

func TestParseAddCardFlags_Flags(t *testing.T) {
	cmd := newAddCardSecretCommand()
	err := setFlags(cmd, map[string]string{
		"secret_id":   "cardid",
		"number":      "1234567890123456",
		"holder":      "John Doe",
		"exp_month":   "12",
		"exp_year":    "2030",
		"cvv":         "123",
		"meta":        "key=val",
		"interactive": "false",
	})
	require.NoError(t, err)

	config, req, err := parseAddCardFlags(cmd)
	require.NoError(t, err)
	assert.Equal(t, "cardid", req.SecretID)
	assert.Equal(t, "1234567890123456", req.Number)
	assert.Equal(t, "John Doe", req.Holder)
	assert.Equal(t, 12, req.ExpMonth)
	assert.Equal(t, 2030, req.ExpYear)
	assert.Equal(t, "123", req.CVV)
	assert.Equal(t, map[string]string{"key": "val"}, req.Meta)
	assert.NotNil(t, config)
}

func TestParseAddCardFlags_MissingRequired(t *testing.T) {
	// helper для установки флагов
	setFlags := func(cmd *cobra.Command, flags map[string]string) error {
		for k, v := range flags {
			if err := cmd.Flags().Set(k, v); err != nil {
				return err
			}
		}
		return nil
	}

	// Проверка отсутствия secret_id
	{
		cmd := newAddCardSecretCommand()
		err := setFlags(cmd, map[string]string{
			"number":    "123",
			"holder":    "holder",
			"exp_month": "1",
			"exp_year":  "2024",
			"cvv":       "321",
		})
		require.NoError(t, err)

		_, _, err = parseAddCardFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret_id required")
	}

	// Проверка отсутствия number
	{
		cmd := newAddCardSecretCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
			"holder":    "holder",
			"exp_month": "1",
			"exp_year":  "2024",
			"cvv":       "321",
		})
		require.NoError(t, err)

		_, _, err = parseAddCardFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card number required")
	}

	// Проверка отсутствия holder
	{
		cmd := newAddCardSecretCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
			"number":    "123",
			"exp_month": "1",
			"exp_year":  "2024",
			"cvv":       "321",
		})
		require.NoError(t, err)

		_, _, err = parseAddCardFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cardholder required")
	}

	// Проверка отсутствия exp_month
	{
		cmd := newAddCardSecretCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
			"number":    "123",
			"holder":    "holder",
			"exp_year":  "2024",
			"cvv":       "321",
		})
		require.NoError(t, err)

		_, _, err = parseAddCardFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid exp_month")
	}

	// Проверка отсутствия exp_year
	{
		cmd := newAddCardSecretCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
			"number":    "123",
			"holder":    "holder",
			"exp_month": "1",
			"cvv":       "321",
		})
		require.NoError(t, err)

		_, _, err = parseAddCardFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid exp_year")
	}

	// Проверка отсутствия cvv
	{
		cmd := newAddCardSecretCommand()
		err := setFlags(cmd, map[string]string{
			"secret_id": "id",
			"number":    "123",
			"holder":    "holder",
			"exp_month": "1",
			"exp_year":  "2024",
		})
		require.NoError(t, err)

		_, _, err = parseAddCardFlags(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cvv required")
	}
}

func TestParseAddCardFlags_InvalidExpMonth(t *testing.T) {
	cmd := newAddCardSecretCommand()
	err := setFlags(cmd, map[string]string{
		"secret_id": "id",
		"number":    "123",
		"holder":    "holder",
		"exp_month": "13", // invalid month
		"exp_year":  "2025",
		"cvv":       "123",
	})
	require.NoError(t, err)
	_, _, err = parseAddCardFlags(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid exp_month")
}

func TestParseAddMetaString(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{
			input:    "",
			expected: map[string]string{},
		},
		{
			input:    "key1=value1",
			expected: map[string]string{"key1": "value1"},
		},
		{
			input:    "key1=value1,key2=value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			input:    " key1 = value1 , key2= value2 ",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			input:    "key1=value1,key2=",
			expected: map[string]string{"key1": "value1", "key2": ""},
		},
		{
			input:    "key1=value1,key2",
			expected: map[string]string{"key1": "value1"}, // ключ без = игнорируется
		},
		{
			input:    "key1=value=with=equals,key2=value2",
			expected: map[string]string{"key1": "value=with=equals", "key2": "value2"},
		},
		{
			input:    "key1=value1,,key2=value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
	}

	for _, tt := range tests {
		t.Run("input="+tt.input, func(t *testing.T) {
			result := parseAddMetaString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
