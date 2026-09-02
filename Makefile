.DEFAULT_GOAL := help

ENV_FILE ?= .env
COMPOSE_FILE ?= deploy/compose.yaml
COMPOSE := docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE)
CERT_DIR ?= deploy/certs

.PHONY: help
help:
	@echo "Leamout deployment commands:"
	@echo "  make certs             Generate self-signed TLS certificates for local/CI use"
	@echo "  make certs-production  Request/install a Let's Encrypt production certificate"
	@echo "  make certs-renew       Renew/sync the Let's Encrypt certificate and restart OpenSIPS"
	@echo "  make check-certs       Validate required TLS certificate files"
	@echo "  make up                Build and start the stack"
	@echo "  make down              Stop the stack"
	@echo "  make deploy            Pull and deploy the latest main branch"
	@echo "  make logs              Follow logs"
	@echo "  make ps                Show service status"
	@echo "  make migrate           Run database migrations"
	@echo "  make restart           Restart application services"

.PHONY: certs
certs:
	CERT_DIR=$(CERT_DIR) sh scripts/certs/generate-self-signed.sh

.PHONY: certs-production
certs-production:
	CERT_DIR=$(CERT_DIR) sh scripts/certs/provision-letsencrypt.sh

.PHONY: certs-renew
certs-renew:
	CERT_DIR=$(CERT_DIR) ENV_FILE=$(ENV_FILE) COMPOSE_FILE=$(COMPOSE_FILE) sh scripts/certs/renew-letsencrypt.sh

.PHONY: check-certs
check-certs:
	CERT_DIR=$(CERT_DIR) sh scripts/certs/check-certs.sh

.PHONY: up
up: check-certs
	$(COMPOSE) up -d --build

.PHONY: down
down:
	$(COMPOSE) down

.PHONY: deploy
deploy: check-certs
	git pull --ff-only origin main
	$(COMPOSE) up -d --build
	$(COMPOSE) ps

.PHONY: logs
logs:
	$(COMPOSE) logs -f --tail=200

.PHONY: ps
ps:
	$(COMPOSE) ps

.PHONY: migrate
migrate:
	$(COMPOSE) run --rm migrate

.PHONY: restart
restart:
	$(COMPOSE) restart server worker
