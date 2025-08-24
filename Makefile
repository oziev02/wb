SHELL := /bin/bash
COMPOSE := docker compose -f deployments/docker-compose.yml
GO ?= go
PRODUCE_N ?= 20
ENV_FILE := .env

.PHONY: up down ps topic-create migrate-up migrate-down run producer health last-id get mocks tidy test lint

# --- infra ---
up:
	$(COMPOSE) up -d postgres zookeeper kafka kafka-ui

down:
	$(COMPOSE) down -v

ps:
	$(COMPOSE) ps

topic-create:
	$(COMPOSE) exec kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic orders --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 || true
	$(COMPOSE) exec kafka /opt/bitnami/kafka/bin/kafka-topics.sh --describe --topic orders --bootstrap-server localhost:9092

# --- migrations (migrate должен быть установлен) ---
migrate-up:
	. $(ENV_FILE); migrate -path migrations -database "$$DB_URL" up

migrate-down:
	. $(ENV_FILE); migrate -path migrations -database "$$DB_URL" down 1

# --- app ---
run:
	set -a; . $(ENV_FILE); set +a; \
	HTTP_ADDR=$${HTTP_ADDR:-:8081} \
	KAFKA_BROKERS=$${KAFKA_BROKERS:-localhost:9092} \
	KAFKA_TOPIC=$${KAFKA_TOPIC:-orders} \
	KAFKA_GROUP=$${KAFKA_GROUP:-orders-consumer} \
	CACHE_CAP=$${CACHE_CAP:-10000} \
	CACHE_TTL=$${CACHE_TTL:-30m} \
	CACHE_RESTORE_LIMIT=$${CACHE_RESTORE_LIMIT:-10000} \
	$(GO) run ./cmd/app

producer:
	. $(ENV_FILE); \
	KAFKA_BROKERS=$${KAFKA_BROKERS:-localhost:9092} \
	KAFKA_TOPIC=$${KAFKA_TOPIC:-orders} \
	PRODUCE_N=$(PRODUCE_N) $(GO) run ./cmd/producer

health:
	@curl -sS -v http://localhost:$${HTTP_PORT:-8081}/healthz || true

last-id:
	@$(COMPOSE) exec -T postgres psql -U wb -d wb -t -A -c "select order_uid from orders order by date_created desc limit 1;"

# usage: make get ORDER_UID=<id>
get:
	@test -n "$(ORDER_UID)" || (echo "Usage: make get ORDER_UID=<order_uid>"; exit 1)
	@curl -sS "http://localhost:$${HTTP_PORT:-8081}/order/$(ORDER_UID)" | jq .

# --- dev utils ---
mocks:
	@command -v mockery >/dev/null 2>&1 || { echo "mockery не установлен. Установи: go install github.com/vektra/mockery/v2@latest"; exit 1; }
	mockery --name=OrderRepository --dir=internal/domain --output=internal/mocks --outpkg=mocks

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint не найден (optional). Install: https://golangci-lint.run/usage/install/"; exit 0; }
	golangci-lint run
