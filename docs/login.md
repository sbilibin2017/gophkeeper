| №  | Действие                                             | Компонент (структура, метод или функция)           |
|----|-----------------------------------------------------|--------------------------------------------------|
| 1  | Клиент отправляет запрос на авторизацию            | Client POST /login {username, password, device_id} |
| 2  | Проверка существования пользователя и верификация пароля | UserReadRepository.Get(username), Server logic   |
| 3  | Проверка устройства по device_id                   | DeviceReadRepository.Get(device_id, user_id)     |
| 4  | Генерация JWT-токена с данными пользователя и устройства | JWT.GenerateToken({user_id, device_id})          |
| 5  | Отправка токена клиенту                            | Server --> Client {token}                        |
