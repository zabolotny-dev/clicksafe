COMPOSE_FILE := zarf/compose/docker-compose.yml
ENV_FILE     := .env

.PHONY: up down logs ps build rebuild migrate

## Запуск всего стека
up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up

## Запуск в фоне
up-d:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d

## Остановка и удаление контейнеров
down:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down

## Остановка + удаление volumes (сброс БД)
down-v:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v

## Логи
logs:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) logs -f

## Статус контейнеров
ps:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) ps

## Сборка образов
build:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) build

## Пересборка без кэша
rebuild:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) build --no-cache

## Только миграции
migrate:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) run --rm api-migrator
