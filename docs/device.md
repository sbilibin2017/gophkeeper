| №  | Действие                                           | Компонент (структура, метод или функция)              |
|----|---------------------------------------------------|------------------------------------------------------|
| 1  | Клиент отправляет GET-запрос с токеном           | `Client GET /get-device (Authorization: Bearer <token>)` |
| 2  | Извлечение токена из запроса                      | `TokenDecoder.GetFromRequest(r)`                     |
| 3  | Парсинг токена и получение userID и deviceID     | `TokenDecoder.Parse(tokenString)`                    |
| 4  | Проверка существования устройства                | `DeviceGetter.Get(ctx, userID, deviceID)`           |
| 5  | Формирование JSON-ответа с данными устройства    | `DeviceResponse`                                     |
| 6  | Отправка ответа клиенту                          | `w.WriteHeader(http.StatusOK) + json.NewEncoder(w).Encode(resp)` |
| 7  | Обработка ошибок: токен неверный                 | `400 Bad Request`                                    |
| 8  | Обработка ошибок: токен невалидный               | `401 Unauthorized`                                   |
| 9  | Обработка ошибок: устройство не найдено          | `404 Not Found`                                      |
| 10 | Обработка ошибок: внутренняя ошибка сервера      | `500 Internal Server Error`                           |