# Build stage
FROM golang:1.24.4-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /app

# Copy auth-service first (needed for local dependency)
COPY auth-service /tmp/auth-service

# Copy api-gateway source code (includes go.mod and go.sum)
COPY api-gateway .

# Use replace directive for local auth-service
RUN go mod edit -replace github.com/cwrk-planet/auth-service=/tmp/auth-service

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-gateway ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/api-gateway .

# Copy config file
COPY api-gateway/config/config.yaml /app/config/config.yaml

EXPOSE 8080

CMD ["./api-gateway"]
