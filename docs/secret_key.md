## Сохранение секретного ключа (Save Secret Key)

| №  | Действие                                           | Компонент (структура, метод или функция)              |
|----|---------------------------------------------------|------------------------------------------------------|
| 1  | Клиент отправляет POST-запрос с токеном и телом   | `Client POST /save-secret-key {SecretKeyResponse, Authorization: Bearer <token>}` |
| 2  | Извлечение токена из запроса                      | `SecretKeyTokenDecoder.GetFromRequest(r)`           |
| 3  | Парсинг токена для получения secretID и deviceID | `SecretKeyTokenDecoder.Parse(tokenString)`          |
| 4  | Декодирование тела запроса                        | `json.NewDecoder(r.Body).Decode(&req)`              |
| 5  | Сохранение секретного ключа                       | `SecretKeyWriter.Save(ctx, secretKeyID, secretID, deviceID, encryptedAESKey)` |
| 6  | Отправка успешного ответа                         | `w.WriteHeader(http.StatusOK)`                      |
| 7  | Обработка ошибок: токен неверный                  | `400 Bad Request`                                   |
| 8  | Обработка ошибок: токен невалидный                | `401 Unauthorized`                                  |
| 9  | Обработка ошибок: внутренняя ошибка сервера       | `500 Internal Server Error`                          |

## Получение секретного ключа (Get Secret Key)

| №  | Действие                                           | Компонент (структура, метод или функция)              |
|----|---------------------------------------------------|------------------------------------------------------|
| 1  | Клиент отправляет GET-запрос с токеном           | `Client GET /get-secret-key (Authorization: Bearer <token>)` |
| 2  | Извлечение токена из запроса                      | `SecretKeyTokenDecoder.GetFromRequest(r)`           |
| 3  | Парсинг токена для получения secretID и deviceID | `SecretKeyTokenDecoder.Parse(tokenString)`          |
| 4  | Получение секретного ключа                        | `SecretKeyGetter.Get(ctx, secretID, deviceID)`      |
| 5  | Формирование JSON-ответа                          | `SecretKeyResponse`                                 |
| 6  | Отправка ответа клиенту                           | `w.WriteHeader(http.StatusOK) + json.NewEncoder(w).Encode(resp)` |
| 7  | Обработка ошибок: токен неверный                  | `400 Bad Request`                                   |
| 8  | Обработка ошибок: токен невалидный                | `401 Unauthorized`                                  |
| 9  | Обработка ошибок: секретный ключ не найден       | `404 Not Found`                                     |
| 10 | Обработка ошибок: внутренняя ошибка сервера      | `500 Internal Server Error`                          |
