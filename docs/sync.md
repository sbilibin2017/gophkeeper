| №  | Действие                                             | Компонент (структура, метод или функция)                               |
|----|-----------------------------------------------------|------------------------------------------------------------------------|
| 1  | Пользователь запускает синхронизацию               | "GophKeeper CLI"                                                        |
| 2  | Получение списка серверных секретов                | SecretHTTPFacade.List(ctx context.Context, userID string) ([]*models.SecretDB, error) |
| 3  | Получение зашифрованного AES-ключа для каждого секрета | SecretKeyHTTPFacade.Get(ctx context.Context, secretID, deviceID string) (*models.SecretKeyDB, error) |
| 4  | Расшифровка AES-ключа                               | RSAEncryptor.DecryptAESKey(EncryptedAESKey)                             |
| 5  | Расшифровка payload                                 | AESGCM.Decrypt(EncryptedPayload, Nonce, AESKey)                         |
| 6  | Сравнение локальной и серверной версии             | Встроенная логика CLI                                                   |
| 7  | Разрешение конфликта                                | interactive режим: выбор пользователя / server/client режим: выбор по правилу |
| 8  | Завершение синхронизации и возврат результата      | "GophKeeper CLI"                                                        |
