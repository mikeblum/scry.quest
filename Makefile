## 🔮 scry.quest 🔮

MAKEFLAGS += --silent

BINARY_NAME=scry.quest

all: help

.PHONY: help
help: ## ❓ Makefile incantations
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[35m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## ⚒️ Build scry.quest
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./...

.PHONY: clean
clean: ## 🧹 Cleanup build artifects
	go clean && rm -f $(BINARY_NAME) coverage.*

.PHONY: lint
lint: ## 👁️ Run linter checks
	golangci-lint run

.PHONY: fmt
fmt: ## ✨ Format code
	go fmt ./...

.PHONY: tidy
tidy: ## 📚 Tidy modules
	go mod tidy

.PHONY: docs
docs: ## 📖 Godocs
	go doc -http

.PHONY: test
test: ## 🧪 Run all tests
	go test -test.v -race -covermode=atomic -coverprofile=coverage.out ./... && \
	if [ -z "$$SSH_CLIENT" ] && [ -z "$$SSH_TTY" ]; then \
		go tool cover -html=coverage.out; \
	else \
		go tool cover -html=coverage.out -o coverage.html && echo "Coverage report saved to coverage.html"; \
	fi && \
	rm -f coverage.out

.PHONY: test-perf
test-perf: ## ⚡ Run benchmark tests
	go test -test.v -benchmem -bench=. -coverprofile=coverage-bench.out ./... && \
	if [ -z "$$SSH_CLIENT" ] && [ -z "$$SSH_TTY" ]; then \
		go tool cover -html=coverage-bench.out; \
	else \
		go tool cover -html=coverage-bench.out -o coverage-bench.html && echo "Coverage report saved to coverage-bench.html"; \
	fi && \
	rm -f coverage-bench.out

.PHONY: vuln
vuln: ## 🛡️ Scan for vulnerabilities
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

.PHONY: sqlc-generate
sqlc-generate: ## 🐘 Generate Go code from SQL
	sqlc generate

.PHONY: sqlc-vet
sqlc-vet: ## 🐘 Vet SQL queries
	sqlc vet

.PHONY: docker-up
docker-up: ## 🐳 Start docker compose services
	docker compose up -d

.PHONY: docker-down
docker-down: ## 🐳 Teardown docker compose services
	docker compose down

.PHONY: embeddings
embeddings: ## 🔮 Run embeddings CLI
	go run ./cmd/embeddings $(filter-out $@,$(MAKECMDGOALS))

.PHONY: pre-commit
pre-commit: fmt tidy lint test sqlc-vet ## ✅ Run all checks
