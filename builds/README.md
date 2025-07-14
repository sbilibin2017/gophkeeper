# Build

В этой папке находятся собранные бинарники и скрипты установки для различных платформ.

## Содержимое

- `gophkeeper-client-linux-amd64` — бинарник клиента для Linux (amd64)
- `gophkeeper-client-macos-amd64` — бинарник клиента для macOS (amd64)
- `gophkeeper-client-windows-amd64.exe` — бинарник клиента для Windows (amd64)


## Установка на Linux/macOS

Скопируйте соответствующий бинарник в каталог, который входит в ваш PATH:

```
sudo cp gophkeeper-client-linux-amd64 /usr/local/bin/gophkeeper-client
sudo chmod +x /usr/local/bin/gophkeeper-client
```

## Установка на Windows

1. Скопируйте файл `gophkeeper-client-windows-amd64.exe` в директорию `%USERPROFILE%\bin`. Если папка не существует, создайте её:

```
mkdir $env:USERPROFILE\bin
copy .\gophkeeper-client-windows-amd64.exe $env:USERPROFILE\bin\gophkeeper-client.exe
```

* Добавьте %USERPROFILE%\bin в переменную окружения PATH:
* Откройте «Переменные среды» (Environment Variables) через свойства системы.
* В разделе «Переменные пользователя» выберите PATH и нажмите «Изменить».
* Добавьте новую заsпись: %USERPROFILE%\bin.
* Нажмите «ОК» и закройте все окна.
* Перезапустите терминал или компьютер, чтобы изменения PATH вступили в силу.
* Теперь клиент доступен из командной строки командой: