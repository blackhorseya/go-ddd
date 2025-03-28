BIN := bin

all: help

help: ## show help
	@grep -hE '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## Clean up the build artifacts
	rm -rf $(BIN) coverage.txt dist .env

.PHONY: install
install: install-tools

.PHONY: install-tools
install-tools: ## Install tools
	awk -F'"' '/_/ {print $$2}' tools.go | xargs -tI % go install %

.PHONY: build
build: ## Build the application
	go build -o $(BIN)/ ./...

.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit: ## Run unit tests
	go test -v -race --tags=!integration,!external -coverprofile=coverage.txt ./...

.PHONY: lint
lint: ## Run linters
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...

.PHONY: gen
gen: gen-swagger

.PHONY: gen-swagger
gen-swagger: ## Generate Swagger documentation
	go run github.com/swaggo/swag/cmd/swag init -g ./cmd/restful/impl.go --output ./docs

.PHONY: docker-run
docker-run: ## Run the application in a Docker container
	docker compose up --build --force-recreate --remove-orphans