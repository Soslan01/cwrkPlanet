# Stage 1: Build client (React/Vite)
FROM node:20-alpine AS client-builder

WORKDIR /app

COPY client/package.json ./
RUN npm install

COPY client/ ./
RUN npm run build

# Stage 2: Build api-gateway (Go)
FROM golang:1.24.4-alpine AS gateway-builder

RUN apk add --no-cache git

WORKDIR /app

COPY auth-service /tmp/auth-service
COPY api-gateway .
RUN go mod edit -replace github.com/cwrk-planet/auth-service=/tmp/auth-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-gateway ./cmd/server

# Stage 3: Runtime - nginx + api-gateway
FROM alpine:latest

RUN apk --no-cache add ca-certificates nginx wget

WORKDIR /app

# Copy client static files
COPY --from=client-builder /app/dist /usr/share/nginx/html

# Copy api-gateway binary and config
COPY --from=gateway-builder /app/api-gateway .
COPY api-gateway/config/config.yaml /app/config/config.yaml

# Copy nginx config
COPY docker/frontend/nginx.conf /etc/nginx/http.d/default.conf

# Entrypoint: run api-gateway in background, nginx in foreground
COPY docker/frontend/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/entrypoint.sh"]
