.PHONY: test test-race vet build sqlc compose-validate compose-validate-observability deployment-preflight deployment-up backoffice-assets backoffice-generate

test: backoffice-generate
	cd server && go test ./...

test-race: backoffice-generate
	cd server && go test -race ./...

vet: backoffice-generate
	cd server && go vet ./...

build: backoffice-generate
	cd server && go build ./...

sqlc:
	cd server && sqlc generate

compose-validate:
	APP_ENV_FILE=../../server/.env.example docker compose --env-file .env.example -f deploy/compose/docker.yaml config --quiet

compose-validate-observability:
	GRAFANA_ADMIN_PASSWORD=ci-only-placeholder APP_ENV_FILE=../../server/.env.example docker compose --env-file .env.example -f deploy/compose/docker.yaml -f deploy/compose/observability.yaml config --quiet

deployment-preflight:
	python3 tests/ops/preflight.py --infra-env .env --app-env server/.env

deployment-up: deployment-preflight
	docker compose --env-file .env -f deploy/compose/docker.yaml config --quiet
	docker compose --env-file .env -f deploy/compose/docker.yaml up -d --build

backoffice-assets:
	cd server/internal/backoffice && bun install --frozen-lockfile && bun run build

backoffice-generate:
	cd server && go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate
