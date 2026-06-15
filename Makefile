.DEFAULT_GOAL := help
.PHONY: help run build test tidy migrate sqlc swag lint docker-up docker-down clean

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'

## run: run the API locally (loads .env if present)
run:
	go run ./cmd/api

## build: compile the binary into ./bin/amana
build:
	go build -trimpath -o bin/amana ./cmd/api

## test: run all tests
test:
	go test ./...

## tidy: tidy go.mod/go.sum
tidy:
	go mod tidy

## migrate: apply database migrations and exit (uses DATABASE_URL)
migrate:
	go run ./cmd/api -migrate-only

## sqlc: generate type-safe SQL code (requires sqlc; used from Phase 3)
sqlc:
	sqlc generate

## swag: regenerate the public OpenAPI spec (json+yaml only, no docs.go) (requires swag; Phase 6)
swag:
	swag init -g cmd/api/main.go -o api/openapi --outputTypes json,yaml --parseDependency --parseInternal

## lint: static checks (go vet; swap for golangci-lint if installed)
lint:
	go vet ./...

## docker-up: build and start the full stack (app + postgres)
docker-up:
	docker compose up --build

## docker-down: stop the stack and remove volumes
docker-down:
	docker compose down -v

## clean: remove build artifacts
clean:
	rm -rf bin
