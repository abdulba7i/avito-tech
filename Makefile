.PHONY: build run test lint clean docker-build docker-up docker-down migrate

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