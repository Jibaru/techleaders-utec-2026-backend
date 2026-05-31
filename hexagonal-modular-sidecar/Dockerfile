# Multi-target Dockerfile. docker-compose picks the target per service.
#
#   docker compose build api          → uses the `api` target
#   docker compose build mailworker   → uses the `mailworker` target

# ---- builder --------------------------------------------------------------
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Cache dependencies separately from source code.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/mailworker ./cmd/mailworker

# ---- api ------------------------------------------------------------------
FROM alpine:3.20 AS api
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/api /usr/local/bin/api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]

# ---- mailworker -----------------------------------------------------------
FROM alpine:3.20 AS mailworker
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/mailworker /usr/local/bin/mailworker
ENTRYPOINT ["/usr/local/bin/mailworker"]
