APP=coffee-loyalty
BIN=bin/$(APP)

.PHONY: run build seed test fmt vet tidy

run:
	go run ./cmd/api

build:
	go build -o $(BIN) ./cmd/api

seed:
	go run ./cmd/seed

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy
