## scry.quest tooling

BINARY_NAME=scry

all: help

.PHONY: help
help: ## Display available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[35m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build binary
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./...

.PHONY: clean
clean: ## Remove build artifacts
	go clean && rm -f $(BINARY_NAME) coverage.*

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: coverage
coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	gofmt -s -w .

.PHONY: tidy
tidy: ## Tidy modules
	go mod tidy

.PHONY: run
run: ## Run application
	go run ./...

.PHONY: vuln
vuln: ## Scan vulnerabilities
	govulncheck ./...

.PHONY: pre-commit
pre-commit: fmt tidy lint test ## Run all checks
