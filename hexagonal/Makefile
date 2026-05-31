APP=coffee-loyalty
BIN=bin/$(APP)

.PHONY: run build seed test fmt vet tidy mocks install-mockery

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

# Install mockery v3 (one-time, into $GOBIN).
install-mockery:
	go install github.com/vektra/mockery/v3@latest

# Regenerate every mock listed in .mockery.yml.
mocks:
	mockery
