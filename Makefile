GOCACHE ?= /tmp/go-build-cache

.PHONY: help build build-components test tests unit-test ci-unit-test integration-test ci-integration-test docker-up docker-up-dev docker-down docker-logs

help:
	@echo "make build       - build gateway binary"
	@echo "make test        - run unit tests"
	@echo "make tests       - run unit and integration tests"
	@echo "make unit-test   - run unit tests"
	@echo "make ci-unit-test - run unit tests"
	@echo "make integration-test - run integration tests in docker compose"
	@echo "make docker-up   - start postgres + aggregator via docker compose kafka profile"
	@echo "make docker-up-dev - start local dev stack with kafka"
	@echo "make docker-up-micro - start gateway + component services"
	@echo "make docker-down - stop docker compose services"
	@echo "make docker-logs - follow service logs"

build:
	go build -o bin/agregator ./src/gateway

build-components:
	go build ./cmd/registry ./cmd/orders ./cmd/contracts ./cmd/analytics

test: unit-test

tests: unit-test integration-test

unit-test:
	GOCACHE=$(GOCACHE) go test ./...

ci-unit-test: unit-test

integration-test:
	@docker network create $${DOCKER_NETWORK:-drones_net} >/dev/null 2>&1 || true
	@AUTH_REQUIRED=true docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile tests up -d --build aggregator postgres zookeeper kafka kafka-init
	@AUTH_REQUIRED=true docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile tests run --build --rm tests
	@AUTH_REQUIRED=true docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile tests down -v --remove-orphans

ci-integration-test: integration-test

docker-up:
	docker compose --profile kafka up -d --build

docker-up-dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka up -d --build

docker-up-micro:
	COMPONENT_DISPATCH_MODE=broker docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile kafka --profile microservices up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
