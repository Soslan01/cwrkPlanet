SHELL := /bin/bash

# ==== Paths ====
ROOT_DIR    := $(abspath .)
LOG_DIR     := $(ROOT_DIR)/logs
PIDS_DIR    := $(ROOT_DIR)/.pids

AUTH_DIR    := $(ROOT_DIR)/auth-service
ROOM_DIR    := $(ROOT_DIR)/room-service
GATEWAY_DIR := $(ROOT_DIR)/api-gateway
CLIENT_DIR  := $(ROOT_DIR)/client

# ==== Tools ====
GO  ?= go
NPM ?= npm

# ==== Phony ====
.PHONY: help bootstrap db-up migrate start start-auth start-room start-gateway start-client stop stop-soft down clean status logs tail-% rebuild

help:
	@echo "Targets:"
	@echo "  make bootstrap   - установить инструменты/генерацию (buf, proto, npm ci)"
	@echo "  make db-up       - поднять PostgreSQL через auth-service (docker compose) и дождаться health"
	@echo "  make migrate     - применить миграции auth + room"
	@echo "  make start       - запустить все сервисы фоном (auth, room, gateway, client)"
	@echo "  make stop        - остановить все сервисы (SIGTERM по PID’ам)"
	@echo "  make down        - остановить БД (docker compose down в auth-service)"
	@echo "  make clean       - down + очистка логов и PID"
	@echo "  make status      - показать PID’ы запущенных процессов"
	@echo "  make logs        - показать, какие лог-файлы есть"
	@echo "  make tail-<svc>  - tail -f для сервиса (auth|room|gateway|client)"

# ==== Bootstrap ====
bootstrap:
	@mkdir -p "$(LOG_DIR)" "$(PIDS_DIR)"
	@echo "==> buf/tools for auth-service"
	@$(MAKE) -C "$(AUTH_DIR)" tools || true
	@echo "==> proto gen: auth-service"
	@$(MAKE) -C "$(AUTH_DIR)" gen || true
	@echo "==> buf/tools for room-service"
	@$(MAKE) -C "$(ROOM_DIR)" tools || true
	@echo "==> proto gen: room-service"
	@$(MAKE) -C "$(ROOM_DIR)" proto || true
	@echo "==> go mod tidy: api-gateway"
	@$(MAKE) -C "$(GATEWAY_DIR)" tidy || true
	@echo "==> npm ci: client"
	@if [ -f "$(CLIENT_DIR)/package.json" ]; then (cd "$(CLIENT_DIR)" && $(NPM) ci); fi
	@echo "Bootstrap done."

# ==== DB auth-service ====
db-up:
	@mkdir -p "$(LOG_DIR)" "$(PIDS_DIR)"
	@echo "==> Starting DB via auth-service/docker compose"
	@$(MAKE) -C "$(AUTH_DIR)" db-up
	@$(MAKE) -C "$(AUTH_DIR)" db-wait

migrate:
	@echo "==> Migrate auth-service (psql in container)"
	@$(MAKE) -C "$(AUTH_DIR)" db-migrate
	@echo "==> Migrate room-service (psql к той же БД)"
	@$(MAKE) -C "$(ROOM_DIR)" migrate

# ==== Start services ====

run: db-up migrate start-auth start-room start-gateway start-client
	@echo "==> All services started"
	@$(MAKE) status

# Запуск auth-service без блокировки текущей сессии
start-auth:
	@mkdir -p "$(LOG_DIR)" "$(PIDS_DIR)"
	@echo "==> Starting auth-service (background)"
	@bash -lc 'cd "$(AUTH_DIR)" && PG_DSN="$${PG_DSN:-postgres://auth:auth@127.0.0.1:5432/auth?sslmode=disable}" \
		SEC_JWT_SECRET="$${SEC_JWT_SECRET:-supersecret}" \
		$(GO) run ./cmd >> "$(LOG_DIR)/auth.log" 2>&1 & echo $$! > "$(PIDS_DIR)/auth.pid"'
	@echo "auth-service PID: $$(cat "$(PIDS_DIR)/auth.pid")"

start-room:
	@mkdir -p "$(LOG_DIR)" "$(PIDS_DIR)"
	@echo "==> Starting room-service (background)"
	@bash -lc 'cd "$(ROOM_DIR)" && CONFIG_PATH="$${CONFIG_PATH:-./config/config.yaml}" \
		$(GO) run ./... >> "$(LOG_DIR)/room.log" 2>&1 & echo $$! > "$(PIDS_DIR)/room.pid"'
	@echo "room-service PID: $$(cat "$(PIDS_DIR)/room.pid")"

start-gateway:
	@mkdir -p "$(LOG_DIR)" "$(PIDS_DIR)"
	@echo "==> Starting api-gateway (background)"
	@bash -lc 'cd "$(GATEWAY_DIR)" && $(MAKE) build >/dev/null && CONFIG_PATH="internal/config/config.yaml" \
		./bin/api-gateway >> "$(LOG_DIR)/gateway.log" 2>&1 & echo $$! > "$(PIDS_DIR)/gateway.pid"'
	@echo "api-gateway PID: $$(cat "$(PIDS_DIR)/gateway.pid")"

start-client:
	@mkdir -p "$(LOG_DIR)" "$(PIDS_DIR)"
	@echo "==> Starting client (background: npm run dev)"
	@bash -lc 'cd "$(CLIENT_DIR)" && if [ ! -d node_modules ]; then $(NPM) ci; fi; \
		$(NPM) run dev >> "$(LOG_DIR)/client.log" 2>&1 & echo $$! > "$(PIDS_DIR)/client.pid"'
	@echo "client PID: $$(cat "$(PIDS_DIR)/client.pid")"

# ==== Stop / Down / Clean ====

stop:
	@echo "==> Stopping background services"
	@for svc in auth room gateway client; do \
		pf="$(PIDS_DIR)/$$svc.pid"; \
		if [ -f "$$pf" ]; then \
			pid=$$(cat "$$pf"); \
			echo "Stopping $$svc (PID $$pid)"; \
			kill -TERM "$$pid" 2>/dev/null || true; \
			for i in $$(seq 1 10); do \
				if ! kill -0 "$$pid" 2>/dev/null; then \
					echo "$$svc stopped"; \
					break; \
				fi; \
				sleep 1; \
			done; \
			kill -KILL "$$pid" 2>/dev/null || true; \
			rm -f "$$pf"; \
		else \
			echo "$$svc not running"; \
		fi; \
	done

stop-soft: stop

down: stop
	@echo "==> Docker compose down (auth-service)"
	@$(MAKE) -C "$(AUTH_DIR)" down

clean: down
	@echo "==> Cleaning logs and PIDs"
	@rm -rf "$(LOG_DIR)" "$(PIDS_DIR)"

status:
	@echo "==> Status"
	@for svc in auth room gateway client; do \
		pf="$(PIDS_DIR)/$$svc.pid"; \
		if [ -f "$$pf" ]; then \
			pid=$$(cat "$$pf"); \
			if kill -0 "$$pid" 2>/dev/null; then \
				echo "  $$svc: RUNNING (PID $$pid)"; \
			else \
				echo "  $$svc: STALE PID ($$pid)"; \
			fi; \
		else \
			echo "  $$svc: STOPPED"; \
		fi \
	done

logs:
	@echo "==> Available logs in $(LOG_DIR):"
	@ls -1 "$(LOG_DIR)" 2>/dev/null || echo "(no logs)"

tail-%:
	@svc="$*"; \
	logf="$(LOG_DIR)/$$svc.log"; \
	if [ -f "$$logf" ]; then \
		echo "==> tail -f $$logf"; \
		tail -f "$$logf"; \
	else \
		echo "No log file: $$logf"; \
	fi