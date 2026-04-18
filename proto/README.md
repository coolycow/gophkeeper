# Каталог `proto/`

Здесь лежит **исходный контракт** gRPC API приложения GophKeeper: файл **`gophkeeper.proto`**.

- Это **единственный ручной** источник правды для сервисов, сообщений и RPC. По нему компилятор `protoc` с плагинами генерирует Go-код.
- Сгенерированные файлы **не редактируют вручную** — они пересоздаются после каждого изменения `.proto`.
- Итог генерации попадает в **`internal/proto/gophkeeperpb/`** (`*.pb.go`, `*_grpc.pb.go`).

Если в `proto/` добавлена копия **`google/protobuf/timestamp.proto`**, импорты well-known разрешаются через `-I proto`. Если её нет — используйте каталог **`include`** из установки `protoc` (см. вариант B ниже).

---

## Что установить один раз

1. **`protoc`** — компилятор Protocol Buffers: [релизы protobuf](https://github.com/protocolbuffers/protobuf/releases) (нужен исполняемый файл и каталог **`include`** с `google/protobuf/*.proto`, если не используете копии из этого репозитория).

2. **Плагины для Go** (в `$GOBIN` или `$GOPATH/bin` должны быть в `PATH`):

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

---

## Как сгенерировать Go-код

Команды выполняются **из корня репозитория** (там, где `go.mod`), модуль: `github.com/coolycow/gophkeeper`.

Если в `proto/` **нет** каталога `google/protobuf/`, для `import "google/protobuf/timestamp.proto"` нужен **вариант B** (системный `include` от `protoc`).

### Вариант A: есть локальная копия `proto/google/protobuf/timestamp.proto`

```bash
protoc -I proto -I . ^
  --go_out=. --go_opt=module=github.com/coolycow/gophkeeper ^
  --go-grpc_out=. --go-grpc_opt=module=github.com/coolycow/gophkeeper ^
  proto/gophkeeper.proto
```

В **bash** строки можно склеить без `^`, одной строкой:

```bash
protoc -I proto -I . --go_out=. --go_opt=module=github.com/coolycow/gophkeeper --go-grpc_out=. --go-grpc_opt=module=github.com/coolycow/gophkeeper proto/gophkeeper.proto
```

### Вариант B: только системный `protoc`

Подставьте путь к **`include`** своей установки (на Windows часто `C:\protoc-XX\include`):

```bash
protoc -I <ПУТЬ_К_INCLUDE> -I . --go_out=. --go_opt=module=github.com/coolycow/gophkeeper --go-grpc_out=. --go-grpc_opt=module=github.com/coolycow/gophkeeper proto/gophkeeper.proto
```

---

## Что получится на выходе

- `internal/proto/gophkeeperpb/gophkeeper.pb.go` — типы сообщений и сериализация.
- `internal/proto/gophkeeperpb/gophkeeper_grpc.pb.go` — интерфейс сервера, клиентские заглушки, регистрация сервиса.

После генерации имеет смысл проверить сборку:

```bash
go build ./...
```

---

## Кратко: когда что делать

| Действие | Когда |
|----------|--------|
| Правите только `.proto` | Запустите `protoc` (команда выше), закоммитьте обновлённые `*.pb.go`. |
| Ошибка `google/protobuf/... not found` | Добавьте `-I` на каталог с well-known типами (`proto` или `include` из установки `protoc`). |
| Ошибка `protoc-gen-go` not found | Повторите `go install ...` и проверьте `PATH`. |
