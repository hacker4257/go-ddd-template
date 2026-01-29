# =========================
# Project
# =========================
APP_NAME := go-ddd-template
SERVER_BIN := server
WORKER_BIN := worker

GO := go
GOFLAGS := -trimpath

# Docker
DOCKER := docker
DOCKER_COMPOSE := docker compose

# Images
SERVER_IMAGE := $(APP_NAME)-server
WORKER_IMAGE := $(APP_NAME)-worker

# =========================
# Default
# =========================
.PHONY: help
help:
	@echo ""
	@echo "Usage:"
	@echo "  make server          Run HTTP server locally"
	@echo "  make worker          Run worker locally"
	@echo ""
	@echo "  make build           Build server & worker binaries"
	@echo "  make clean           Remove local binaries"
	@echo ""
	@echo "  make deps-up         Start dependencies (mysql/redis/kafka)"
	@echo "  make deps-down       Stop dependencies"
	@echo "  make deps-reset      Stop dependencies and remove volumes"
	@echo "  make deps-logs       Tail dependency logs"
	@echo ""
	@echo "  make docker-build    Build docker images (server & worker)"
	@echo "  make docker-up       Run full stack with docker-compose"
	@echo "  make docker-down     Stop docker-compose"
	@echo ""

# =========================
# Local development
# =========================
.PHONY: server
server:
	$(GO) run ./cmd/server

.PHONY: worker
worker:
	$(GO) run ./cmd/worker

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o bin/$(SERVER_BIN) ./cmd/server
	$(GO) build $(GOFLAGS) -o bin/$(WORKER_BIN) ./cmd/worker

.PHONY: clean
clean:
	rm -rf bin

# =========================
# Dependencies (docker-compose)
# =========================
.PHONY: deps-up
deps-up:
	$(DOCKER_COMPOSE) up -d mysql redis kafka kafka-init

.PHONY: deps-down
deps-down:
	$(DOCKER_COMPOSE) down

.PHONY: deps-reset
deps-reset:
	$(DOCKER_COMPOSE) down -v

.PHONY: deps-logs
deps-logs:
	$(DOCKER_COMPOSE) logs -f

# =========================
# Docker build & run
# =========================
.PHONY: docker-build
docker-build:
	$(DOCKER) build -f build/Dockerfile.server -t $(SERVER_IMAGE) .
	$(DOCKER) build -f build/Dockerfile.worker -t $(WORKER_IMAGE) .

.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) up -d

.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

# =========================
# Lint / Test (optional hooks)
# =========================
.PHONY: test
test:
	$(GO) test ./...

