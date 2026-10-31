.PHONY: help up down deploy logs ps migrate restart generate

COMPOSE = docker compose --env-file .env --env-file server/.env -f deploy/compose/docker.yaml
APP_SERVICES = server backoffice worker media

help:
	@echo "Leamout deployment commands:"
	@echo "  make up       Build and start the stack"
	@echo "  make down     Stop the stack"
	@echo "  make deploy   Pull and deploy the latest main branch"
	@echo "  make logs     Follow service logs (optional: SERVICE=server)"
	@echo "  make ps       Show service status"
	@echo "  make migrate  Run database migrations"
	@echo "  make restart  Restart application services"
	@echo "  make generate Regenerate generated code"

up:
	python3 tests/ops/preflight.py --infra-env .env --app-env server/.env
	$(COMPOSE) config --quiet
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

deploy:
	@test -z "$$(git status --porcelain)" || { echo "Refusing to deploy with a dirty working tree." >&2; exit 1; }
	git fetch origin main
	git switch main
	git pull --ff-only origin main
	$(MAKE) up

logs:
	$(COMPOSE) logs --follow --tail=100 $(SERVICE)

ps:
	$(COMPOSE) ps

migrate:
	$(COMPOSE) run --rm migrate

restart:
	$(COMPOSE) restart $(APP_SERVICES)

generate:
	cd server && go run github.com/a-h/templ/cmd/templ@v0.3.1100 generate
	cd server && sqlc generate
