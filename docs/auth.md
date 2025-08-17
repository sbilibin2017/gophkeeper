## Регистрация (Register)

| №  | Действие                                               | Компонент (структура, метод или функция)                          |
|----|-------------------------------------------------------|------------------------------------------------------------------|
| 1  | Клиент отправляет запрос на регистрацию нового пользователя | `Client POST /register {username, password}`                     |
| 2  | Валидация данных пользователя                         | Логика на сервере: `validateUsername()` и `validatePassword()`   |
| 3  | Проверка существования пользователя                  | `UserReadRepository.Get(username)`                                |
| 4  | Генерация хэша пароля                                 | `PasswordHasher.Hash(password)`                                   |
| 5  | Генерация пары ключей RSA для устройства             | `RSAGenerator.GenerateKeyPair()`                                  |
| 6  | Создание нового пользователя                          | `UserWriteRepository.Save(user_id, username, hashed_password)`    |
| 7  | Создание новой записи устройства                      | `DeviceWriteRepository.Save(user_id, device_id, public_key)`      |
| 8  | Генерация JWT токена с данными пользователя и устройства | `Tokener.Generate(user_id, device_id)`                            |
| 9  | Отправка клиенту `user_id`, `device_id`, токена и приватного ключа | `Server --> Client {user_id, device_id, token, PrivateKey}`       |

## Логин (Login)

| №  | Действие                                               | Компонент (структура, метод или функция)                          |
|----|-------------------------------------------------------|------------------------------------------------------------------|
| 1  | Клиент отправляет запрос на авторизацию              | `Client POST /login {username, password, device_id}`             |
| 2  | Проверка существования пользователя                 | `UserReadRepository.Get(username)`                                |
| 3  | Верификация пароля                                   | `PasswordComparer.Compare(hash, password)`                        |
| 4  | Проверка устройства по `device_id`                  | `DeviceReadRepository.Get(user_id, device_id)`                    |
| 5  | Генерация JWT-токена с данными пользователя и устройства | `Tokener.Generate(user_id, device_id)`                            |
| 6  | Отправка токена клиенту                              | `Server --> Client {token}`                                       |
