package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/configs/transport/http"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
)

// RunClientRegisterHTTP подключается к базе данных пользователя, инициализирует HTTP-клиент
// и выполняет регистрацию нового пользователя и устройства через AuthHTTPFacade.
func RunClientRegisterHTTP(
	ctx context.Context,
	serverURL string,
	databaseMigrationsDir string,
	username string,
	password string,
	deviceID string,
) ([]byte, string, error) {
	dbConn, err := db.New(
		"sqlite",
		fmt.Sprintf("%s.db", username),
	)
	if err != nil {
		return nil, "", err
	}
	defer dbConn.Close()

	client := http.New(serverURL).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	authFacade := facades.NewAuthHTTPFacade(client, jwt.GetTokenFromRestyResponse)

	privKey, token, err := authFacade.Register(ctx, username, password, deviceID)
	if err != nil {
		return nil, "", err
	}

	return privKey, token, nil
}
