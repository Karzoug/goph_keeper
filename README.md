# Менеджер паролей GophKeeper

GophKeeper представляет собой клиент-серверную систему, позволяющую пользователю надёжно и безопасно хранить логины, пароли, бинарные данные и прочую приватную информацию.

[![asciicast](https://asciinema.org/a/601014.svg)](https://asciinema.org/a/601014?speed=3)

[![asciicast](https://asciinema.org/a/601016.svg)](https://asciinema.org/a/601016?speed=2.5)

# Запуск
Для локального запуска следует использовать Docker и предложенный Makefile:
- up-server:
  - генерирует ключ и сертификат для TLS,
  - генерирует случайный секретный ключ, используемый для подписи токенов,
  - запускает redis, postgres, mailhog (для перехвата писем от сервера) и сервер GophKeeper;
  - на порту 8025 размещает веб-интерфейс mailhog.
- build-client: создает клиент GophKeeper в папке client/cmd/.