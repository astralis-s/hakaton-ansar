.PHONY: run build test tidy fmt vet

run:
	go run ./cmd/bot

build:
	go build -o bin/amana-bot ./cmd/bot

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...
