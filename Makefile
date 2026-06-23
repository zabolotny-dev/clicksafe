COMPOSE_FILE           := zarf/compose/docker-compose.yml
VOICE_GPU_COMPOSE_FILE := zarf/compose/docker-compose.voice-gpu.yml
ENV_FILE               := .env

.PHONY: up up-d down down-v logs ps build rebuild max-build max-rebuild max-up max-up-d max-logs max-ps voice-build voice-rebuild voice-up voice-up-d voice-logs voice-ps voice-gpu-build voice-gpu-rebuild voice-gpu-up voice-gpu-up-d voice-gpu-logs voice-gpu-ps migrate seed seed-minimal migrate-seed load-fixture cli test test-unit test-frontend test-integration load-test


## Все тесты: unit (Go + frontend)
test: test-unit test-frontend

## Unit-тесты Go (без Docker)
test-unit:
	go test ./business/... -count=1

## Тесты фронтенда (Vitest)
test-frontend:
	cd api/frontend && npm test

## Интеграционные тесты (нужен Docker, запускаются последовательно; нагрузочный тест исключён)
test-integration:
	go test $(shell go list ./api/service/tests/... | grep -v loadtest) -count=1 -p 1 -timeout 300s

## Нагрузочный тест: поднимает изолированный Postgres, сидает и тестирует (нужен Docker)
load-test:
	go test ./api/service/tests/loadtest/... -v -run TestLoadScenarios -count=1 -timeout 40m

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

## Сборка max-adapter
max-build:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) build max-adapter

## Пересборка max-adapter без кэша
max-rebuild:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) build --no-cache max-adapter

## Запуск max-adapter
max-up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up max-adapter

## Запуск max-adapter в фоне
max-up-d:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d max-adapter

## Логи max-adapter
max-logs:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) logs -f max-adapter

## Статус max-adapter
max-ps:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) ps max-adapter

## Сборка voice-adapter для CPU
voice-build:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice build voice-adapter

## Пересборка voice-adapter для CPU без кэша
voice-rebuild:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice build --no-cache voice-adapter

## Запуск voice-adapter на CPU
voice-up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice up voice-adapter

## Запуск voice-adapter на CPU в фоне
voice-up-d:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice up -d voice-adapter

## Логи voice-adapter
voice-logs:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice logs -f voice-adapter

## Статус voice-adapter
voice-ps:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice ps voice-adapter

## Сборка voice-adapter для NVIDIA GPU/CUDA
voice-gpu-build:
	docker compose -f $(COMPOSE_FILE) -f $(VOICE_GPU_COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice build voice-adapter

## Пересборка voice-adapter для NVIDIA GPU/CUDA без кэша
voice-gpu-rebuild:
	docker compose -f $(COMPOSE_FILE) -f $(VOICE_GPU_COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice build --no-cache voice-adapter

## Запуск voice-adapter на NVIDIA GPU/CUDA
voice-gpu-up:
	docker compose -f $(COMPOSE_FILE) -f $(VOICE_GPU_COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice up voice-adapter

## Запуск voice-adapter на NVIDIA GPU/CUDA в фоне
voice-gpu-up-d:
	docker compose -f $(COMPOSE_FILE) -f $(VOICE_GPU_COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice up -d voice-adapter

## Логи voice-adapter на NVIDIA GPU/CUDA
voice-gpu-logs:
	docker compose -f $(COMPOSE_FILE) -f $(VOICE_GPU_COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice logs -f voice-adapter

## Статус voice-adapter на NVIDIA GPU/CUDA
voice-gpu-ps:
	docker compose -f $(COMPOSE_FILE) -f $(VOICE_GPU_COMPOSE_FILE) --env-file $(ENV_FILE) --profile voice ps voice-adapter

## Только миграции
migrate:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) run --rm api-tool

## Загрузка demo seed-данных через tooling
seed:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) run --rm api-tool /clicksafe_tool seed demo

## Минимальный seed: организация
seed-minimal:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) run --rm api-tool /clicksafe_tool seed minimal

## Миграции + demo seed-данные
migrate-seed:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) run --rm api-tool /clicksafe_tool migrate-seed demo

## Старое имя для demo seed
load-fixture: seed

## Запуск CLI-утилиты (пример: make cli CMD="admins" или make cli CMD="revoke <id>")
cli:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) run --rm api-tool /clicksafe_tool $(CMD)
