# ========= Settings (override via env) =========
PROTO_DIR ?= proto
BUF_TEMPLATE ?= tools/buf.gen.yaml

SERVICE_PKG   ?= ./cmd
GO            ?= go
DOCKER        ?= docker
COMPOSE       ?= docker compose
DB_SERVICE    ?= db
DB_CONTAINER  ?= auth_pg

# DB creds must match docker-compose.yml
DB_USER       ?= auth
DB_PASS       ?= auth
DB_NAME       ?= auth
DB_HOST       ?= 127.0.0.1
DB_PORT       ?= 5432
PG_DSN        ?= postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# ========= PHONY =========
.PHONY: help tools gen db-up db-wait db-migrate run stop down clean fmt tidy

help:
	@echo "Targets:"
	@echo "  make run        - gen + up DB (docker) + wait + migrate + go run"
	@echo "  make gen        - generate stubs from $(CONTRACTS_DIR) into $(API_OUT)"
	@echo "  make stop       - stop only the app (Ctrl+C), DB keeps running"
	@echo "  make down       - stop DB (docker compose down)"
	@echo "  make clean      - remove DB volume (_pgdata)"
	@echo "  make tools      - install buf/protoc plugins"
	@echo "  make fmt, tidy  - go fmt / go mod tidy"

# ========= Tools =========
tools:
	@echo "==> Installing buf and protoc plugins"
	$(GO) install github.com/bufbuild/buf/cmd/buf@latest
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@echo "OK"

gen: $(BUF_TEMPLATE)
	@command -v buf >/dev/null || { echo "Error: buf not found. Run 'make tools'"; exit 1; }
	@echo "==> Fetching deps (proto/buf.yaml)"
	@( cd proto && buf dep update )
	@echo "==> Generating stubs from proto → proto/gen"
	buf generate "proto" --template "tools/buf.gen.yaml"



# ========= DB lifecycle =========
db-up:
	@echo "==> Starting PostgreSQL via docker compose"
	$(COMPOSE) up -d $(DB_SERVICE)

db-wait:
	@echo "==> Waiting for PostgreSQL to become healthy..."
	@for i in $$(seq 1 60); do \
		status=$$($(DOCKER) inspect -f '{{.State.Health.Status}}' $(DB_CONTAINER) 2>/dev/null || echo "starting"); \
		if [ "$$status" = "healthy" ]; then echo "DB is healthy"; exit 0; fi; \
		sleep 1; \
	done; \
	echo "DB failed to become healthy in time" && exit 1

db-migrate:
	@echo "==> Applying SQL migrations (psql inside container)"
	@set -e; \
	for f in $$(ls -1 migrations/*.sql 2>/dev/null | sort); do \
		echo "  -> $$f"; \
		$(DOCKER) exec -i $(DB_CONTAINER) psql "postgres://$(DB_USER):$(DB_PASS)@localhost:5432/$(DB_NAME)?sslmode=disable" -v ON_ERROR_STOP=1 -f - < $$f ; \
	done; \
	echo "Migrations applied"

# ========= App run =========
run: gen db-up db-wait db-migrate
	@echo "==> Running service"
	PG_DSN='$(PG_DSN)' SEC_JWT_SECRET="$${SEC_JWT_SECRET:-supersecret}" $(GO) run $(SERVICE_PKG)

# Convenience
stop:
	@echo "==> Stop: use Ctrl+C to stop the app. DB keeps running."

down:
	@echo "==> Stopping docker compose (DB)"
	$(COMPOSE) down

clean: down
	@echo "==> Removing DB volume"
	@rm -rf _pgdata

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy
