.PHONY: help build test docker-up docker-down docker-logs

help:
	@echo "make build       - build gateway binary"
	@echo "make test        - run Go tests"
	@echo "make docker-up   - start the system via docker compose"
	@echo "make docker-down - stop docker compose services"
	@echo "make docker-logs - follow service logs"

build:
	go build -o bin/agregator ./src/gateway

test:
	go test ./...

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
