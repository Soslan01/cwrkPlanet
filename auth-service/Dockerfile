# Build stage
FROM golang:1.24.4-alpine AS builder

# Install protoc
RUN apk add --no-cache protobuf

# Install protoc-gen-go and protoc-gen-go-grpc via go install
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add Go bin to PATH so protoc can find the plugins
ENV PATH=$PATH:/go/bin

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate proto files
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/auth/auth.proto

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth-service ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/auth-service .

# Copy config file
COPY config/config.yaml /app/config/config.yaml

# Create keys directory (keys should be mounted or generated)
RUN mkdir -p /app/config/keys

EXPOSE 50051 8081

CMD ["./auth-service"]
