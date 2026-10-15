.PHONY: test test-race vet build sqlc compose-validate compose-validate-observability

test:
	cd server && go test ./...

test-race:
	cd server && go test -race ./...

vet:
	cd server && go vet ./...

build:
	cd server && go build ./...

sqlc:
	cd server && sqlc generate

compose-validate:
	APP_ENV_FILE=../../server/.env.example docker compose --env-file .env.example -f deploy/compose/docker.yaml config --quiet

compose-validate-observability:
	GRAFANA_ADMIN_PASSWORD=ci-only-placeholder APP_ENV_FILE=../../server/.env.example docker compose --env-file .env.example -f deploy/compose/docker.yaml -f deploy/compose/observability.yaml config --quiet
