package apps

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunHTTP(t *testing.T) {
	ctx := context.Background()

	// Мокаем сервер
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/register" {
			http.Error(w, "не найдено", http.StatusNotFound)
			return
		}

		w.Header().Set("Authorization", "Bearer mocked_token")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("mocked_private_key"))
	}))
	defer mockServer.Close()

	// Удаляем БД после теста
	defer func() {
		os.Remove("testuser.db")
	}()

	privKey, token, err := RunClientRegisterHTTP(ctx, mockServer.URL, "", "testuser", "testpass", "device123")
	require.NoError(t, err, "ожидалась успешная регистрация")
	require.Equal(t, []byte("mocked_private_key"), privKey, "приватный ключ должен совпадать с моковым")
	require.Equal(t, "mocked_token", token, "токен должен совпадать с моковым")
}

func TestRunHTTP_DeviceIDError(t *testing.T) {
	ctx := context.Background()

	// Передаем пустой deviceID для проверки ошибки
	privKey, token, err := RunClientRegisterHTTP(ctx, "http://localhost:8080", "", "user", "pass", "")
	require.Error(t, err)
	require.Nil(t, privKey)
	require.Equal(t, "", token)

	// Удаляем БД после теста
	defer func() {
		os.Remove("user.db")
	}()
}
