# configaudit

`configaudit` — CLI-утилита на Go для аудита YAML- и JSON-конфигураций приложений на потенциально опасные настройки.

Основной режим работы — CLI. Утилита анализирует содержимое конфигурации, находит проблемы безопасности, показывает уровень критичности и рекомендации по исправлению, поддерживает рекурсивное сканирование директорий и может опционально проверять права доступа к файлам. HTTP и gRPC доступны как дополнительные транспорты поверх того же движка аудита.

## Возможности

- CLI — основной способ использования утилиты.
- Сканирует конфигурационные файлы в форматах JSON и YAML.
- Читает входные данные из файла или из `stdin`.
- Рекурсивно обходит директории и анализирует файлы `.json`, `.yaml` и `.yml`.
- Для каждой проблемы показывает уровень критичности, путь в конфигурации, сообщение, рекомендацию и путь к файлу, если это уместно.
- Возвращает ненулевой код выхода при обнаружении проблем или ошибок входных данных.
- Опционально проверяет права доступа к конфигурационным файлам через `os.Stat`.
- HTTP и gRPC используют тот же движок аудита, что и CLI.
- Построена на расширяемом наборе правил: новые проверки можно добавлять без переписывания сканера.

## Сборка

```bash
go build -o configaudit ./cmd/configaudit
```

## Make commands

```bash
make fmt
make tidy
make test
make build
make run
make docker-build
```

## Docker

Сборка Docker-образа:

```bash
docker build -t configaudit .
```

Сканирование файла из текущей директории через контейнер:

```bash
docker run --rm \
  -v "$(pwd):/work" \
  configaudit testdata/debug-log.json
```

Сканирование из `stdin` через контейнер:

```bash
cat testdata/weak-algorithm.yaml | docker run --rm -i configaudit --stdin
```

## Использование CLI

Сканирование файла:

```bash
./configaudit testdata/debug-log.json
```

Сканирование из `stdin`:

```bash
cat testdata/weak-algorithm.yaml | ./configaudit --stdin
```

Явное указание формата входных данных:

```bash
cat config.txt | ./configaudit --stdin --format yaml
```

Сохранять код выхода `0`, даже если проблемы найдены:

```bash
./configaudit --silent testdata/debug-log.json
```

Рекурсивное сканирование директории:

```bash
./configaudit --recursive ./configs
```

Рекурсивное сканирование с проверкой прав доступа к файлам:

```bash
./configaudit --recursive --check-permissions ./configs
```

## Пример вывода

Ниже показан фактический вывод CLI:

```text
Found 2 problem(s)

[HIGH] storage.digest-algorithm
Weak algorithm MD5 detected.
Recommendation: Replace it with a modern secure alternative such as SHA-256, SHA-512, bcrypt, scrypt, Argon2, AES-GCM, depending on the use case.

[LOW] log.level
Debug or trace logging is enabled (DEBUG).
Recommendation: Do not use debug/trace logging in production. Use info or a more restrictive level.
```

При рекурсивном сканировании в вывод также включается путь к файлу:

```text
[HIGH] configs/prod.yaml:storage.digest-algorithm
Weak algorithm MD5 detected.
Recommendation: Replace it with a modern secure alternative such as SHA-256, SHA-512, bcrypt, scrypt, Argon2, AES-GCM, depending on the use case.
```

## Optional transports

CLI остаётся основным интерфейсом. HTTP и gRPC — это опциональные транспорты, которые переиспользуют тот же движок аудита и те же правила.

### HTTP

Запуск сервера:

```bash
./configaudit --http :8080
```

или:

```bash
./configaudit server --http :8080
```

Проверка `/_info`:

```bash
curl -s http://localhost:8080/_info
```

Сканирование сырого YAML:

```bash
curl -s \
  -X POST http://localhost:8080/scan \
  -H 'Content-Type: application/yaml' \
  --data-binary $'storage:\n  digest-algorithm: MD5\n'
```

Если `Content-Type` неоднозначен, можно явно передать `?format=json` или `?format=yaml`.
`POST /scan` принимает сырой текст конфигурации в теле запроса. Формат определяется по `Content-Type` или через query-параметр `format`.

Запуск HTTP-сервера в контейнере:

```bash
docker run --rm -p 8080:8080 configaudit --http :8080
```

### gRPC

Proto-файл:

`api/proto/configaudit.proto`

Запуск сервера:

```bash
./configaudit --grpc :9090
```

или:

```bash
./configaudit server --grpc :9090
```

Пример запроса через `grpcurl`:

```bash
grpcurl -plaintext \
  -d '{"content":"log:\n  level: debug\n","format":"yaml"}' \
  localhost:9090 \
  configaudit.v1.ConfigAudit/Scan
```

В репозиторий уже включены сгенерированные Go-файлы для protobuf/gRPC, поэтому проект можно собрать без отдельного запуска генерации кода.

Если нужно пересоздать их заново:

```bash
PATH="$HOME/go/bin:$PATH" \
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  api/proto/configaudit.proto
```

Для этого локально должны быть установлены `protoc`, `protoc-gen-go` и `protoc-gen-go-grpc`.

Запуск gRPC-сервера в контейнере:

```bash
docker run --rm -p 9090:9090 configaudit --grpc :9090
```

## Реализованные правила

1. Включён `debug` или `trace` уровень логирования.
2. Пароли и секреты хранятся в открытом виде в конфигурации.
3. Сервис слушает `0.0.0.0`.
4. TLS или проверка сертификатов отключены.
5. Используются слабые или устаревшие алгоритмы, например `MD5`, `SHA1`, `DES`, `3DES`, `RC4`, `Blowfish`, `none` или `plaintext`.
6. Конфигурационный файл доступен на запись группе или всем пользователям.

## Коды выхода

- `0`: проблем не найдено, либо найденные проблемы были проигнорированы по флагу `--silent`.
- `1`: найдена одна или более проблем.
- `2`: некорректное использование, отсутствующий входной файл, ошибки чтения/парсинга или внутренняя ошибка, включая ошибку запуска сервера.

## Запуск тестов

```bash
go test ./...
```
## Быстрая проверка

```bash
gofmt -w .
go mod tidy
go test ./...
go build -o configaudit ./cmd/configaudit

./configaudit testdata/debug-log.json
./configaudit --silent testdata/debug-log.json
cat testdata/weak-algorithm.yaml | ./configaudit --stdin

docker build -t configaudit .
docker run --rm -v "$(pwd):/work" configaudit testdata/debug-log.json
```
