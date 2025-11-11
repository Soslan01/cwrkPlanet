SHELL := /bin/bash

GO   ?= go
BUF  ?= buf
PSQL ?= psql
CONFIG_PATH ?= ./config/config.yaml
DSN ?= postgres://auth:auth@127.0.0.1:5432/auth?sslmode=disable
PROTO_DIR := ./proto
MIGRATIONS_DIR := ./migrations

.PHONY: tools
tools:
	@echo "==> Installing Buf (psql установи через пакетный менеджер: brew install libpq && brew link --force libpq)"
	$(GO) install github.com/bufbuild/buf/cmd/buf@latest

.PHONY: proto
proto:
	@echo "==> buf generate (from ./proto)"
	@cd proto && $(BUF) generate

# todo: когда нибудь будет :)
.PHONY: lint
lint:
	@echo "==> buf lint (from ./proto)"
	@cd proto && $(BUF) lint

.PHONY: migrate
migrate:
	@if ! command -v $(PSQL) >/dev/null 2>&1; then echo "ERROR: psql not found. Install: brew install libpq && brew link --force libpq"; exit 1; fi
	@echo "==> applying migrations to $(DSN)"
	@set -e; \
	for f in $$(ls -1 $(MIGRATIONS_DIR)/*.sql | sort); do \
		echo "   -> $$f"; \
		$(PSQL) -v ON_ERROR_STOP=1 -d "$(DSN)" -f "$$f"; \
	done
	@echo "==> migrations applied"

.PHONY: build
build:
	@echo "==> building room-service"
	$(GO) build -o ./bin/room-service ./...

.PHONY: run
run:
	@echo "==> checking proto generation"
	@if [ ! -f proto/gen/room/v1/room.pb.go ]; then \
		echo "==> generating proto files"; \
		$(MAKE) proto; \
	else \
		echo "==> proto files exist"; \
	fi
	@echo "==> applying migrations if needed"
	$(MAKE) migrate
	@echo "==> running room-service"
	CONFIG_PATH=$(CONFIG_PATH) $(GO) run ./...
.PHONY: dev
dev: proto migrate run
	@true
