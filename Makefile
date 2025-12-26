.PHONY: help up down build restart logs test test-unit test-integration test-concurrency clean

help: ## Показать справку по командам
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Запустить все сервисы
	docker compose up -d

up-build: ## Запустить с пересборкой образов
	docker compose up -d --build

down: ## Остановить все сервисы
	docker compose down

down-volumes: ## Остановить и удалить volumes
	docker compose down -v

build: ## Собрать образы
	docker compose build

restart: ## Перезапустить сервисы
	docker compose restart

logs: ## Показать логи всех сервисов
	docker compose logs -f

logs-app: ## Показать логи приложения
	docker compose logs -f app

logs-db: ## Показать логи базы данных
	docker compose logs -f postgres

test: ## Запустить все тесты
	go test ./...

test-unit: ## Запустить unit-тесты
	go test ./internal/service/... -v

test-integration: ## Запустить интеграционные тесты (требует запущенной БД)
	docker compose up -d postgres
	sleep 3
	go test ./tests/... -v -run TestIntegration

test-concurrency: ## Запустить тесты на конкурентность (требует запущенной БД)
	docker compose up -d postgres
	sleep 3
	go test ./tests/... -v -run TestConcurrency

test-all: ## Запустить все тесты (unit + integration + concurrency)
	docker compose up -d postgres
	sleep 3
	go test ./... -v

clean: ## Очистить Docker (остановить контейнеры, удалить образы)
	docker compose down -v
	docker rmi $$(docker images -q wallet* 2>/dev/null) 2>/dev/null || true

ps: ## Показать статус контейнеров
	docker compose ps

shell-app: ## Войти в контейнер приложения
	docker compose exec app sh

shell-db: ## Войти в контейнер БД
	docker compose exec postgres psql -U postgres -d itk_wallet

dev: ## Запустить только БД для разработки
	docker compose up -d postgres

stop: ## Остановить все сервисы
	docker compose stop

