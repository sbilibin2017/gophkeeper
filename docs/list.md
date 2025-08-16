### Offline

| №  | Действие                                             | Компонент (структура, метод или функция)           |
|----|-----------------------------------------------------|--------------------------------------------------|
| 1  | Пользователь запрашивает список секретов           | "GophKeeper CLI"                                 |
| 2  | Получение зашифрованных секретов из локальной БД   | SecretReadRepository.List(userID)               |
| 3  | Получение зашифрованного AES-ключа устройства      | SecretKeyReadRepository.Get(secretID, deviceID) |
| 4  | Расшифровка AES-ключа                               | RSAEncryptor.DecryptAESKey(EncryptedAESKey)     |
| 5  | Расшифровка payload                                 | AESGCM.Decrypt(EncryptedPayload, Nonce, AESKey) |
| 6  | Преобразование секрета по типу                      | Встроенная логика CLI: LoginPassword / TextNote / BinaryData / BankCard / Unknown |
| 7  | Возврат списка пользователю                         | "GophKeeper CLI"                                 |

### Online

| №  | Действие                                             | Компонент (структура, метод или функция)           |
|----|-----------------------------------------------------|--------------------------------------------------|
| 1  | Пользователь запрашивает список секретов           | "GophKeeper CLI"                                 |
| 2  | Получение зашифрованных секретов с сервера         | SecretHTTPFacade.List(ctx, token)                |
| 3  | Получение зашифрованного AES-ключа устройства      | SecretKeyHTTPFacade.Get(secretID, deviceID)      |
| 4  | Расшифровка AES-ключа                               | RSAEncryptor.DecryptAESKey(EncryptedAESKey)     |
| 5  | Расшифровка payload                                 | AESGCM.Decrypt(EncryptedPayload, Nonce, AESKey) |
| 6  | Преобразование секрета по типу                      | Встроенная логика CLI: LoginPassword / TextNote / BinaryData / BankCard / Unknown |
| 7  | Возврат списка пользователю                         | "GophKeeper CLI"                                 |
