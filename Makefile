# CWRK - Local development (no Docker except database)

.PHONY: db-up db-down dev-frontend dev-backend dev

# Start only the database container
db-up:
	docker compose up -d db

# Stop the database
db-down:
	docker compose stop db

# Run auth-service (backend) - requires db running
dev-backend:
	cd auth-service && make run

# Run api-gateway (part of frontend) - requires auth-service running
dev-gateway:
	cd api-gateway && make run

# Run client dev server - requires api-gateway running
dev-client:
	cd client && npm run dev

# Full Docker stack
up:
	docker compose up -d

down:
	docker compose down
