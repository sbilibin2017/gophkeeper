## Сохранение секрета (Save Secret)

| №  | Действие                                           | Компонент (структура, метод или функция)              |
|----|---------------------------------------------------|------------------------------------------------------|
| 1  | Клиент отправляет POST-запрос с токеном и телом   | `Client POST /save-secret {SecretResponse, Authorization: Bearer <token>}` |
| 2  | Извлечение токена из запроса                      | `SecretTokenDecoder.GetFromRequest(r)`              |
| 3  | Парсинг токена для получения secretID и userID   | `SecretTokenDecoder.Parse(tokenString)`             |
| 4  | Декодирование тела запроса                        | `json.NewDecoder(r.Body).Decode(&req)`              |
| 5  | Сохранение секрета                                | `SecretWriter.Save(ctx, secretID, req.UserID, req.SecretName, req.SecretType, req.EncryptedPayload, req.Nonce, req.Meta)` |
| 6  | Отправка успешного ответа                         | `w.WriteHeader(http.StatusOK)`                      |
| 7  | Обработка ошибок: токен отсутствует или некорректный | `400 Bad Request`                                   |
| 8  | Обработка ошибок: токен недействителен           | `401 Unauthorized`                                  |
| 9  | Обработка ошибок: внутренняя ошибка сервера      | `500 Internal Server Error`                          |


## Получение секрета по имени (Get Secret)

| №  | Действие                                           | Компонент (структура, метод или функция)              |
|----|---------------------------------------------------|------------------------------------------------------|
| 1  | Клиент отправляет GET-запрос с токеном и именем секрета | `Client GET /get-secret?secret_name=<name> (Authorization: Bearer <token>)` |
| 2  | Извлечение токена из запроса                      | `SecretTokenDecoder.GetFromRequest(r)`              |
| 3  | Парсинг токена для получения userID              | `SecretTokenDecoder.Parse(tokenString)`             |
| 4  | Получение имени секрета из параметров запроса    | `r.URL.Query().Get("secret_name")`                  |
| 5  | Получение секрета из хранилища                   | `SecretReader.Get(ctx, userID, secretName)`         |
| 6  | Формирование JSON-ответа                          | `SecretResponse`                                    |
| 7  | Отправка ответа клиенту                           | `w.Header().Set("Content-Type","application/json") + json.NewEncoder(w).Encode(resp)` |
| 8  | Обработка ошибок: токен отсутствует или некорректный | `400 Bad Request`                                   |
| 9  | Обработка ошибок: токен недействителен           | `401 Unauthorized`                                  |
| 10 | Обработка ошибок: секрет не найден               | `404 Not Found`                                     |
| 11 | Обработка ошибок: внутренняя ошибка сервера      | `500 Internal Server Error`                          |


## Получение списка всех секретов пользователя (List Secrets)

| №  | Действие                                           | Компонент (структура, метод или функция)              |
|----|---------------------------------------------------|------------------------------------------------------|
| 1  | Клиент отправляет GET-запрос с токеном           | `Client GET /list-secrets (Authorization: Bearer <token>)` |
| 2  | Извлечение токена из запроса                      | `SecretTokenDecoder.GetFromRequest(r)`              |
| 3  | Парсинг токена для получения userID              | `SecretTokenDecoder.Parse(tokenString)`             |
| 4  | Получение списка секретов пользователя           | `SecretReader.List(ctx, userID)`                    |
| 5  | Преобразование секретов в JSON-ответ             | `[]SecretResponse`                                  |
| 6  | Отправка ответа клиенту                           | `w.Header().Set("Content-Type","application/json") + json.NewEncoder(w).Encode(resp)` |
| 7  | Обработка ошибок: токен отсутствует или некорректный | `400 Bad Request`                                   |
| 8  | Обработка ошибок: токен недействителен           | `401 Unauthorized`                                  |
| 9  | Обработка ошибок: внутренняя ошибка сервера      | `500 Internal Server Error`                          |
