PROJECT ?= avito-test-applicant
COMPOSE := docker compose -p $(PROJECT)
FILES := -f compose.yml

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help



up: ### Run docker-compose
	$(COMPOSE) $(FILES) up -d --build --remove-orphans
.PHONY: up

down: ### Down docker-compose
	$(COMPOSE) $(FILES) down --remove-orphans
.PHONY: down

logs: ### Show docker-compose logs
	$(COMPOSE) $(FILES) logs -f
.PHONY: logs


up-postgres: ### Run only Postgres service
	$(COMPOSE) $(FILES) up -d postgres
.PHONY: up-postgres

down-postgres: ### Stop and remove Postgres service
	$(COMPOSE) $(FILES) down --remove-orphans postgres
.PHONY: down-postgres

logs-postgres: ### Follow logs for Postgres service
	$(COMPOSE) $(FILES) logs -f postgres
.PHONY: logs-postgres

clean-postgres: down-postgres ### Delete postgres-data volume
	docker volume rm $(PROJECT)_postgres-data
.PHONY: clean-postgres



logs-app: ### Follow logs for App service
	$(COMPOSE) $(FILES) logs -f app
.PHONY: logs-app

rebuild: ### Rebuild app service (stop/remove then up --build)
	$(COMPOSE) $(FILES) stop app
	$(COMPOSE) $(FILES) rm -f app
	$(COMPOSE) $(FILES) up -d --build app
.PHONY: rebuild



test: ### run test
	go test -v ./...
.PHONY: test

cover: ### run test with coverage
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out
	rm coverage.out
.PHONY: cover



generate-api: ## Generate API code from OpenAPI spec
	oapi-codegen --config=docs/oapi-codegen.yml docs/openapi.yml
.PHONY: generate-api

migrate-create:  ### create new migration
	migrate create -ext sql -dir migrations $(name)
.PHONY: migrate-create
