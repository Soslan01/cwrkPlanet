# CWRK

## Architecture

| Container | Services | Description |
|-----------|----------|-------------|
| **db** | PostgreSQL | Database only |
| **frontend** | API Gateway + Client (nginx) | Single entry point: serves SPA and proxies /api to api-gateway |
| **backend** | Auth-service (+ future services) | Backend microservices |

## Request Flow

```
Client (browser) → nginx → api-gateway → backend services (gRPC)
```

All API requests use `/api` prefix and are proxied through nginx.

## Docker

```bash
docker compose up
```

- App: http://localhost:3000
- Database: localhost:5433 (postgres/postgres)

## Local Development (no Docker except DB)

1. Start database:
   ```bash
   docker compose up -d db
   ```

2. Start auth-service:
   ```bash
   cd auth-service && make run
   ```

3. Start api-gateway:
   ```bash
   cd api-gateway && make run
   ```

4. Start client (Vite dev server with /api proxy):
   ```bash
   cd client && npm run dev
   ```

Configs: `config.local.yaml` is used when `CONFIG_PATH` is not set (services default to localhost).
