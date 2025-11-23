# Reviewer Service
Сервис назначения ревьюеров для Pull Request'ов.

---

## Требования

- Docker
- Docker Compose

## Запуск

Команды `Makefile`:
```
# Сборка Docker образа
docker-build:
	docker-compose build

# Запуск через docker-compose
docker-up:
	docker-compose up -d

# Остановка docker-compose
docker-down:
	docker-compose down

# Полная пересборка и запуск
rebuild: docker-down docker-build docker-up

# Запуск тестов
test:
	go test -v ./internal/tests/...
```

Сервис будет доступен на `http://localhost:8080`

## Структура проекта

- `cmd/service/main.go` - точка входа приложения
- `cmd/inits/` - инициализация БД, сервисов, маршрутов, миграции, конфиг и .env файл
- `internal/handlers/` - HTTP handlers для всех эндпоинтов
- `internal/services/` - бизнес-логика
- `internal/storage/` - работа с базой данных
- `internal/models/` - модели данных
- `migrations/` - SQL миграции
- `api/openapi.yaml` - OpenAPI спецификация

## API Endpoints

- `POST /team/add` - Создать команду с участниками
- `GET /team/get` - Получить команду
- `POST /users/setIsActive` - Установить флаг активности пользователя
- `GET /users/getReview?user_id=<id>` - Получить PR'ы пользователя
- `POST /pullRequest/create` - Создать PR и назначить ревьюверов
- `POST /pullRequest/merge` - Пометить PR как MERGED
- `POST /pullRequest/reassign` - Переназначить ревьювера
- `GET /health` - Health check
- `GET /statistics` - статистика

## Особенности реализации

- Автоматическое назначение до 2 активных ревьюверов из команды автора при создании PR
- Переназначение ревьювера из команды заменяемого ревьювера
- Запрет изменения ревьюверов после MERGED
- Идемпотентная операция merge
- Миграции применяются автоматически при старте приложения  (файл `cmd/inits/migrations.go`)
- Эндпоинт статистики по назначениям и PR

## Дополнительные возможности

### Статистика
Эндпоинт `/statistics` предоставляет:
- Количество назначений по каждому пользователю
- Общую статистику по PR (всего, открытых, мерженных, назначений)

### Массовая деактивация
Эндпоинт `/users/bulkDeactivateTeam` позволяет:
- Деактивировать всех пользователей команды одной операцией
- Автоматически переназначить ревьюверов в открытых PR
- Оптимизирован для работы < 100 мс при средних объемах данных

## Тестирование

Интеграционные тесты находятся в `internal/tests/integration_test.go`.

Запуск тестов:
```bash
make test
```


## Переменные окружения

Все переменные окружения настраиваются в файле `.env`:

- `DB_HOST` - хост базы данных (по умолчанию `db`)
- `DB_PORT` - порт базы данных (по умолчанию `5432`)
- `DB_USER` - пользователь БД
- `DB_PASSWORD` - пароль БД
- `DB_NAME` - имя БД
- `HTTP_ADDRESS` - адрес HTTP сервера (по умолчанию `:8080`)
- `HTTP_TIMEOUT` - таймаут HTTP запросов
- `HTTP_IDLE_TIMEOUT` - таймаут простоя соединений

## Остановка

Для остановки сервисов используйте:
```bash
docker-compose down
```

Для остановки с удалением данных БД:
```bash
docker-compose down -v
```

---
### Примечание:
При разработке проекта использовал чистую архитекуру, которой пользуется в командах авито. Ниже прикрепил файл:

https://drive.google.com/drive/folders/10WHW-ncrVZfct4Y0OkV5LUTIKXTvosOB?usp=drive_link