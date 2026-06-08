## ─── e-Dossier API Makefile ───────────────────────────────────────────────────

.PHONY: help run build test lint tidy docker-up docker-down migrate seed

BINARY     := edossier
CMD        := ./cmd/server
BUILD_FLAGS := -ldflags="-w -s"

## help: Show available targets
help:
	@grep -E '^##' Makefile | sed 's/## //'

## run: Run the API locally (reads .env)
run:
	@if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi; \
	go run $(CMD)

## build: Compile the binary
build:
	go build $(BUILD_FLAGS) -o bin/$(BINARY) $(CMD)

## test: Run all tests with race detector
test:
	go test -race -count=1 ./...

## test/cover: Run tests with coverage report
test/cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run golangci-lint (install separately)
lint:
	golangci-lint run ./...

## tidy: Tidy and verify Go modules
tidy:
	go mod tidy
	go mod verify

## fmt: Format all Go source files
fmt:
	gofmt -w .

## vet: Run go vet
vet:
	go vet ./...

## docker-up: Start all services with Docker Compose
docker-up:
	docker compose up -d --build
	@echo "API running at http://localhost:8090"

## docker-down: Stop all Docker services and remove volumes
docker-down:
	docker compose down -v

## docker-logs: Tail API container logs
docker-logs:
	docker compose logs -f api

## docker-ps: Show running containers
docker-ps:
	docker compose ps

## clean: Remove build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

## check: Run fmt, vet, and test in sequence
check: fmt vet test
