# GophKeeper

Клиент-серверное приложение для безопасного хранения приватных данных (логины и пароли, текст, бинарные данные, данные банковских карт и др.). Полезная нагрузка на сервере хранится в зашифрованном виде; формат шифрования и метаданные задаёт клиент.

Сервер поднимает **HTTP**-интерфейс (например, служебные маршруты и статистику) и **gRPC** API (`GophKeeperService`) на отдельном порту. Конфигурация задаётся флагами, переменными окружения и JSON-файлом.

---

## Настройка сервера

### Приоритет значений

Итоговая конфигурация собирается в таком порядке (последние источники **перекрывают** предыдущие):

1. **Значения по умолчанию** в коде.
2. **JSON-файл** — путь задаётся флагом `-c` / `--config` или переменной `CONFIG`. Если путь не указан явно и файла по умолчанию нет, шаг пропускается.
3. **Переменные окружения** (см. таблицу ниже).
4. **Флаги командной строки** — имеют наивысший приоритет; учитываются только **переданные** флаги (`Changed`).

Пример шаблона JSON: [`config_server.json.example`](config_server.json.example). Скопируйте его в `config_server.json` и подставьте свой DSN и секреты.

### Переменные окружения

| Переменная | Назначение |
|------------|------------|
| `CONFIG` | Путь к JSON-конфигу (если не переопределён флагом `-c`). |
| `HOST` | Адрес прослушивания (HTTP и gRPC). |
| `PORT` | Порт HTTP. |
| `GRPC_PORT` | Порт gRPC (не должен совпадать с `PORT`). |
| `LOG_LEVEL` | Уровень логов. |
| `DATABASE_DSN` | Строка подключения к PostgreSQL. |
| `RUN_MIGRATIONS` | Запуск миграций при старте (`true`/`false`). |
| `ENABLE_HTTPS` | TLS для HTTP (`true`/`false`). |
| `TLS_CERT_FILE`, `TLS_KEY_FILE` | PEM сертификат и ключ для HTTP/gRPC TLS. |
| `TRUSTED_SUBNET` | CIDR для доступа к внутренней статистике (если используется). |
| `SECRET_KEY` | Секрет для токенов/шифрования на сервере (**ровно 32 символа**). |
| `SECRET_VERSION_COUNT` | Лимит версий секрета (0 — без лимита в допустимых границах). |
| `SALT_LENGTH` | Длина соли (параметры пользователя/пароля). |
| `MIN_PASSWORD_LENGTH`, `MAX_PASSWORD_LENGTH` | Ограничения длины пароля. |
| `AUDIT_FILE`, `AUDIT_URL` | Файл и URL для аудита. |

### Основные флаги

| Флаг | Кратко | Описание |
|------|--------|----------|
| `-h`, `--host` | хост | По умолчанию `127.0.0.1`. |
| `-p`, `--port` | порт HTTP | По умолчанию `8080`. |
| `-g`, `--grpc-port` | порт gRPC | `0` — выбрать свободный начиная с `PORT+1`. |
| `-l`, `--log-level` | уровень логов | Например `info`. |
| `-d`, `--database-dsn` | DSN БД | |
| `-r`, `--run-migrations` | миграции | |
| `-s`, `--enable-https` | HTTPS | |
| `-t`, `--tls-cert-file` | сертификат TLS | |
| `-k`, `--tls-key-file` | ключ TLS | |
| `-c`, `--config` | JSON-конфиг | По умолчанию имя `config_server.json`. |
| `--trusted-subnet` | доверенная подсеть | CIDR. |
| `-x`, `--secret-key` | секретный ключ | 32 символа. |
| `-v`, `--secret-version-count` | лимит версий | |
| `-m`, `--salt-length` | длина соли | |
| `-o`, `--min-password-length` | мин. длина пароля | |
| `--max-password-length` | макс. длина пароля | |
| `-a`, `--audit-file` | файл аудита | |
| `-b`, `--audit-url` | URL аудита | |

Путь к JSON и сами поля в файле используют **snake_case** (`host`, `database_dsn`, `grpc_port`, …). Ключ `config` внутри JSON для загрузки не используется — путь к файлу задаётся только `-c` / `CONFIG`.

---

## gRPC API (для Insomnia и других клиентов)

### Подключение

- **Адрес:** `HOST:GRPC_PORT` (тот же хост, что и у HTTP; порт из конфигурации).
- **TLS:** если включён `ENABLE_HTTPS`, gRPC использует те же `TLS_CERT_FILE` / `TLS_KEY_FILE` — в Insomnia включите TLS и при self-signed укажите доверие к сертификату или отключите проверку только для разработки.
- **Схема:** импортируйте в Insomnia файл [`proto/gophkeeper.proto`](proto/gophkeeper.proto) (или включите server reflection, если добавите его в сервер).
- **Сервис в proto:** `gophkeeper.v1.GophKeeperService`.

### Авторизация (metadata)

Для всех методов, **кроме** перечисленных в блоке «без токена», нужен заголовок metadata:

| Ключ | Значение |
|------|----------|
| `authorization` | Токен доступа в том виде, как его выдаёт сервер после `Register` / `Login` / `RefreshToken` (строка в hex). Допустим префикс `Bearer ` перед токеном. |

В Insomnia: для запроса gRPC откройте раздел **Metadata** / **Headers** (в зависимости от версии) и добавьте пару `authorization` = ваш токен.

### Методы

Полный путь вызова в стиле gRPC (как в логах и в коде):

| Метод (RPC) | Полный путь | Нужен `authorization` | Кратко |
|-------------|-------------|-------------------------|--------|
| `Ping` | `/gophkeeper.v1.GophKeeperService/Ping` | Нет | Проверка живости; тело: `PingRequest` (пустое сообщение). |
| `Register` | `…/Register` | Нет | Регистрация; тело: `email`, `password`. |
| `Login` | `…/Login` | Нет | Вход; тело: `email`, `password`. |
| `RefreshToken` | `…/RefreshToken` | Нет | Обновление сессии; тело: `refresh_token`. |
| `ListSecrets` | `…/ListSecrets` | Да | Список секретов; тело: опционально `filter_kind` (`SecretKind`). |
| `GetSecret` | `…/GetSecret` | Да | Один секрет; тело: `secret_id`, `include_version_history`. |
| `CreateSecret` | `…/CreateSecret` | Да | Новый секрет; тело: `data_encrypted`, `data_format_version`, `kind`. |
| `UpdateSecret` | `…/UpdateSecret` | Да | Новая версия; тело: `secret_id`, `data_encrypted`, `data_format_version`. |
| `DeleteSecret` | `…/DeleteSecret` | Да | Удаление; тело: `secret_id`. |
| `ListSecretVersions` | `…/ListSecretVersions` | Да | История версий; тело: `secret_id`. |
| `GetSecretVersion` | `…/GetSecretVersion` | Да | Одна версия; тело: `secret_id`, `secret_version_id`. |
| `ListAttachments` | `…/ListAttachments` | Да | Список вложений; тело: **один из** `secret_id` **или** `secret_version_id` (`oneof`). |
| `GetAttachment` | `…/GetAttachment` | Да | Скачать вложение; тело: `attachment_id`. |
| `CreateAttachment` | `…/CreateAttachment` | Да | Загрузить вложение; тело: `secret_version_id`, форматы, `info_encrypted`, `data_encrypted`. |
| `DeleteAttachment` | `…/DeleteAttachment` | Да | Удалить вложение; тело: `attachment_id`. |

Типы полей и перечисление `SecretKind` описаны в [`proto/gophkeeper.proto`](proto/gophkeeper.proto).

---

## Запуск сервера

Из корня репозитория (после настройки `DATABASE_DSN` и при необходимости `config_server.json`):

```bash
go run ./cmd/server/...
```

С миграциями один раз:

```bash
go run ./cmd/server/... -r
```

(или `RUN_MIGRATIONS=true` в окружении — см. актуальное поведение в `internal/config`.)

ВНИМАНИЕ: при первом запуске сервере миграции запустятся автоматически если не будет найдена таблица `users`.
