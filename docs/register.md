| №  | Действие                                             | Компонент (структура, метод или функция)           |
|----|-----------------------------------------------------|--------------------------------------------------|
| 1  | Клиент отправляет запрос на регистрацию нового пользователя | Client POST /register {username, password}       |
| 2  | Валидация данных пользователя                       | Server logic: validate username и password       |
| 3  | Проверка существования пользователя                | UserReadRepository.Get(username)                 |
| 4  | Генерация пары ключей RSA для устройства           | RSA.GenerateRSAKeyPair()                         |
| 5  | Создание нового пользователя                        | UserWriteRepository.Save(username, hashed_password) |
| 6  | Создание новой записи устройства                    | DeviceWriteRepository.Save(user_id, public_key) |
| 7  | Генерация JWT токена с данными пользователя и устройства | JWT.GenerateToken({user_id, device_id})          |
| 8  | Отправка клиенту user_id, device_id, token и приватного ключа | Server --> Client {user_id, device_id, token, PrivateKey} |
