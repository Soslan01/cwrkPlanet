.PHONY: proto run proto-check

# Check if proto files exist, generate if missing
proto-check:
	@if [ ! -f "proto/auth/auth.pb.go" ] || [ ! -f "proto/auth/auth_grpc.pb.go" ]; then \
		echo "Proto files missing, generating..."; \
		$(MAKE) proto; \
	fi

# Generate proto files
proto:
	@echo "Generating proto files..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/auth/auth.proto

# Run the service
run: proto-check
	@echo "Starting auth-service..."
	@go run cmd/server/main.go
