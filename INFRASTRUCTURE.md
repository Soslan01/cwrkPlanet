# Документация инфраструктуры

## Обзор архитектуры

Проект CWRK — микросервисное приложение с трёхуровневой Docker-архитектурой:

```
┌─────────────────────────────────────────────────────────────────┐
│                         КЛИЕНТ (браузер)                         │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│  КОНТЕЙНЕР 2: FRONTEND                                          │
│  ┌──────────────┐    ┌──────────────────┐                       │
│  │    nginx     │───▶│   api-gateway    │                       │
│  │ (статика +   │    │ (HTTP → gRPC)    │                       │
│  │  /api proxy) │    │   :8080          │                       │
│  └──────────────┘    └────────┬─────────┘                       │
│         :80                   │                                  │
└───────────────────────────────┼─────────────────────────────────┘
                                │ gRPC :50051
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│  КОНТЕЙНЕР 3: BACKEND                                           │
│  ┌──────────────────────────────────┐                           │
│  │         auth-service             │                           │
│  │  (gRPC + бизнес-логика)          │                           │
│  └──────────────┬───────────────────┘                           │
└─────────────────┼───────────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────────┐
│  КОНТЕЙНЕР 1: DB                                                │
│  ┌──────────────────────────────────┐                           │
│  │       PostgreSQL 16              │                           │
│  │         :5432                    │                           │
│  └──────────────────────────────────┘                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Контейнеры

### 1. db (PostgreSQL)

- **Образ**: `postgres:16-alpine`
- **Порт**: 5433:5432
- **База**: `auth_db`
- **Учётные данные**: postgres / postgres
- **Healthcheck**: `pg_isready -U postgres`
- **Volume**: `postgres_data` для хранения данных

### 2. frontend (API Gateway + Client)

- **Сборка**: multi-stage Dockerfile (`docker/frontend/Dockerfile`)
- **Порт**: 3000:80
- **Содержимое**:
  - nginx — статические файлы SPA, проксирование `/api/` на api-gateway
  - api-gateway (Go) — HTTP REST → gRPC, слушает 127.0.0.1:8080
- **Зависимости**: backend (auth-service)
- **Конфиг**: монтируется `api-gateway/config/config.yaml`

### 3. backend (auth-service)

- **Сборка**: `auth-service/Dockerfile`
- **Порты**: gRPC 50051, HTTP 8081 (внутренние)
- **Содержимое**: auth-service — аутентификация, JWT, сессии
- **Зависимости**: db (после healthcheck)
- **Конфиг**: монтируется `auth-service/config/config.yaml`
- **Ключи JWT**: монтируется `auth-service/config/keys/`

---

## Маршрутизация запросов

1. **Клиент** → `/api/v1/auth/*` (относительный путь, base URL = `/api`)
2. **nginx** → `location /api/` → `proxy_pass http://127.0.0.1:8080/` (префикс `/api` отбрасывается)
3. **api-gateway** → получает `/v1/auth/*`, вызывает auth-service по gRPC
4. **auth-service** → выполняет запрос к PostgreSQL

---

## Структура каталогов

```
cwrk/
├── api-gateway/           # HTTP API Gateway (Go, Chi)
│   ├── cmd/server/        # точка входа
│   ├── config/            # config.yaml (Docker), config.local.yaml (локально)
│   └── internal/
│       ├── client/auth/   # gRPC-клиент auth-service
│       ├── config/
│       ├── domain/        # DTO, ошибки
│       ├── handler/http/  # HTTP-хендлеры
│       ├── logger/
│       ├── service/auth/  # оркестрация
│       └── transport/http/
│
├── auth-service/          # Сервис аутентификации (Go, gRPC)
│   ├── cmd/server/
│   ├── config/            # config.yaml (Docker), config.local.yaml (локально)
│   ├── proto/auth/        # Protobuf-схема
│   └── internal/
│       ├── database/
│       ├── handler/       # gRPC-хендлеры
│       ├── repository/    # миграции, репозитории
│       ├── service/
│       └── ...
│
├── client/                # SPA (React, Vite, TypeScript)
│   ├── src/
│   │   ├── components/
│   │   ├── context/
│   │   ├── pages/
│   │   └── services/api.ts
│   └── vite.config.ts     # proxy /api → localhost:8080
│
├── docker/
│   └── frontend/
│       ├── Dockerfile     # client + api-gateway + nginx
│       ├── nginx.conf
│       └── entrypoint.sh  # запуск api-gateway, затем nginx
│
├── docker-compose.yaml
├── Makefile               # команды для локальной разработки
├── API.md                 # документация API
└── INFRASTRUCTURE.md      # эта документация
```

---

## Конфигурация

### Переменные окружения

| Сервис    | Переменная  | Описание                                      |
|-----------|-------------|-----------------------------------------------|
| frontend  | CONFIG_PATH | Путь к config api-gateway (Docker: `/app/config/config.yaml`) |
| backend   | CONFIG_PATH | Путь к config auth-service (Docker: `/app/config/config.yaml`) |

### Docker (config.yaml)

**api-gateway** (`api-gateway/config/config.yaml`):
- `clients.auth.address`: `backend:50051`
- HTTP: `:8080`

**auth-service** (`auth-service/config/config.yaml`):
- `postgres.dsn`: `postgres://...@db:5432/auth_db`
- gRPC: `:50051`
- JWT: RS256, ключи в `/app/config/keys/`

### Локальная разработка (config.local.yaml)

При запуске без Docker (кроме БД) сервисы по умолчанию используют `config.local.yaml`, если `CONFIG_PATH` не задан.

**auth-service** (`auth-service/config/config.local.yaml`):
- `postgres.dsn`: `localhost:5433` (Docker-порт БД)

**api-gateway** — нужен `api-gateway/config/config.local.yaml`:
```yaml
clients:
  auth:
    address: "localhost:50051"
```

---

## Сеть Docker

- **Сеть**: `app-network` (bridge)
- Все сервисы в одной сети, обращаются друг к другу по имени сервиса (db, backend, frontend).

---

## Запуск

### Полный стек (Docker)

```bash
docker compose up -d
```

Приложение: http://localhost:3000

### Локальная разработка (БД в Docker)

1. `docker compose up -d db`
2. `cd auth-service && make run`
3. `cd api-gateway && make run`
4. `cd client && npm run dev`

Клиент: http://localhost:3000 (Vite проксирует `/api` на api-gateway).

---

## База данных

- **Миграции**: выполняются при старте auth-service (`internal/repository/migrations/`)
- **Таблицы**: `users`, `sessions`
- **Индексы**: по email, username, session token, refresh token, expires_at

---

## Безопасность

- **JWT**: RS256, асимметричные ключи (auth-service подписывает, api-gateway проверяет через gRPC)
- **Пароли**: bcrypt (cost 10)
- **Сессии**: refresh token хранится в БД с хэшем, при logout инвалидируется
