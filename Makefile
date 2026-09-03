.DEFAULT_GOAL := help

ENV_FILE ?= .env
COMPOSE_FILE ?= deploy/compose.yaml
COMPOSE := docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE)
CERT_DIR ?= deploy/certs
DEPLOY_ENV := ENV_FILE=$(ENV_FILE) COMPOSE_FILE=$(COMPOSE_FILE) CERT_DIR=$(CERT_DIR)

.PHONY: help
help:
	@echo "Leamout deployment commands:"
	@echo "  make certs             Generate self-signed TLS certificates for local/CI use"
	@echo "  make certs-production  Request/install a Let's Encrypt production certificate"
	@echo "  make certs-renew       Renew/sync the Let's Encrypt certificate through safe activation"
	@echo "  make certs-auto-renew  Install Certbot deploy hook and enable renewal timer when available"
	@echo "  make rotate-certs      Activate SOURCE_CERT_DIR after validation and graceful drain"
	@echo "  make rotate-turn-secret     Begin old/new TURN shared-secret overlap"
	@echo "  make finalize-turn-secret   Remove the previous TURN secret after the overlap window"
	@echo "  make rotate-esl-password    Rotate FreeSWITCH ESL password after graceful drain"
	@echo "  make rotate-carrier-key     Transactionally re-encrypt stored carrier credentials"
	@echo "  make check-certs       Validate required TLS certificate files"
	@echo "  make preflight         Validate environment, certificates, and Compose configuration"
	@echo "  make up                Build and start the stack"
	@echo "  make down              Stop the stack"
	@echo "  make deploy            Pull and deploy the latest main branch, then verify it"
	@echo "  make verify            Verify deployment service health"
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

.PHONY: certs-auto-renew
certs-auto-renew:
	CERT_DIR=$(CERT_DIR) ENV_FILE=$(ENV_FILE) COMPOSE_FILE=$(COMPOSE_FILE) sh scripts/certs/install-certbot-renewal.sh

.PHONY: rotate-certs
rotate-certs:
	CERT_DIR=$(CERT_DIR) ENV_FILE=$(ENV_FILE) COMPOSE_FILE=$(COMPOSE_FILE) sh scripts/certs/activate-runtime.sh

.PHONY: rotate-turn-secret
rotate-turn-secret:
	ENV_FILE=$(ENV_FILE) COMPOSE_FILE=$(COMPOSE_FILE) sh scripts/secrets/rotate-turn-secret.sh begin

.PHONY: finalize-turn-secret
finalize-turn-secret:
	ENV_FILE=$(ENV_FILE) COMPOSE_FILE=$(COMPOSE_FILE) sh scripts/secrets/rotate-turn-secret.sh finalize

.PHONY: rotate-esl-password
rotate-esl-password:
	ENV_FILE=$(ENV_FILE) COMPOSE_FILE=$(COMPOSE_FILE) sh scripts/secrets/rotate-esl-password.sh

.PHONY: rotate-carrier-key
rotate-carrier-key:
	ENV_FILE=$(ENV_FILE) COMPOSE_FILE=$(COMPOSE_FILE) sh scripts/secrets/rotate-carrier-key.sh

.PHONY: check-certs
check-certs:
	CERT_DIR=$(CERT_DIR) sh scripts/certs/check-certs.sh

.PHONY: preflight
preflight:
	$(DEPLOY_ENV) sh scripts/deploy/preflight.sh

.PHONY: up
up:
	$(DEPLOY_ENV) sh scripts/deploy/up.sh

.PHONY: down
down:
	$(COMPOSE) down

.PHONY: deploy
deploy:
	$(DEPLOY_ENV) sh scripts/deploy/deploy.sh

.PHONY: verify
verify:
	$(DEPLOY_ENV) sh scripts/deploy/verify.sh

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
	$(DEPLOY_ENV) sh scripts/deploy/restart.sh
