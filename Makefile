.PHONY: run build docker-up docker-down swagger lint fmt help

help:
	@echo "Available targets:"
	@echo "  make run         - Run the application"
	@echo "  make build       - Build the application"
	@echo "  make docker-up   - Start Docker containers"
	@echo "  make docker-down - Stop Docker containers"
	@echo "  make swagger     - Generate Swagger documentation"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"

run:
	go run ./cmd/server

build:
	go build -o server ./cmd/server

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

swagger:
	swag init -g cmd/server/main.go

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
