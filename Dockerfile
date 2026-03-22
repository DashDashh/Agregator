# ─── Stage 1: build ──────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /aggregator ./cmd/main.go

# ─── Stage 2: minimal runtime ────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /aggregator /aggregator
# Миграции должны быть доступны при запуске сервиса
COPY --from=builder /app/migrations /migrations
# Страница фронтенда
COPY --from=builder /app/frontend /frontend

# Конфигурация передаётся через переменные окружения
ENV KAFKA_BROKER=kafka:9092 \
    KAFKA_REQUEST_TOPIC=aggregator.requests \
    KAFKA_RESPONSE_TOPIC=aggregator.responses \
    KAFKA_CONSUMER_GROUP=aggregator-group \
    KAFKA_DLT_TOPIC=aggregator.dead-letter

ENTRYPOINT ["/aggregator"]
