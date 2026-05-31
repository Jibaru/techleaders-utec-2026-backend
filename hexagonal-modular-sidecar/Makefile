APP=coffee-loyalty
BIN=bin/$(APP)

.PHONY: run mailworker build seed test fmt vet tidy mocks install-mockery up down logs

run:
	go run ./cmd/api

mailworker:
	go run ./cmd/mailworker

build:
	go build -o $(BIN) ./cmd/api
	go build -o bin/mailworker ./cmd/mailworker

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

# Install mockery v3 (one-time, into $GOBIN).
install-mockery:
	go install github.com/vektra/mockery/v3@latest

# Regenerate every mock listed in .mockery.yml.
mocks:
	mockery

# ---- docker-compose convenience ------------------------------------------
up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f api mailworker
