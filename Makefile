.PHONY: help build run test fmt cert docker-dev docker-up docker-down migrate

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS=":.*?## "} {printf "\t\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go binary
	@go build -o bin/crmcore cmd/api/main.go

run: ## Run the application
	@go run cmd/api/main.go

test: ## Run all tests
	@go test ./... -cover

fmt: ## Format code
	@go fmt ./...

cert: ## Generate self-signed SAML cert and key (saml.crt, saml.key)
	@openssl req -x509 -newkey rsa:2048 -nodes -keyout saml.key -out saml.crt -days 365 -subj "/CN=localhost"

docker-dev: ## Build and run Docker compose for development
	@docker compose up --build

docker-up: ## Start Docker compose services
	@docker compose -f docker-compose.mock.yml  up -d # Start IdP
	@docker compose up -d

docker-down: ## Stop docker compose services
	@docker compose down

docker-up-watch: ## Start docker compose services and watch core logs
	@make docker-up
	@docker logs -f api -n30

idp-mock: ## Start and idp mock server using keycloack
	@docker compose -f docker-compose.mock.yml up

docker-migrate: ## Run migrations
	@docker exec api go run cmd/setup/main.go

secret=""
docker-secret: ## Prints the encoded secret to insert on identity_providers: (docker-secret secret=<client-secret>)
	@docker exec api go run cmd/secret/main.go -secret=${secret}
