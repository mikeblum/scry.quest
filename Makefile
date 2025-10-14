## 🔮 scry.quest 🔮
SHELL := /bin/bash
MAKEFLAGS += --silent

BINARY_NAME=scry.quest
TAILWIND_VERSION := 3.4.0
HEROICONS_VERSION := 2.0.18
ASSETS_DIR := cmd/web/static
# podman or docker
DOCKER := podman

all: help

.PHONY: help
help: ## ❓ Makefile incantations
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[35m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: assets
assets: $(ASSETS_DIR)/css/tailwind.min.css $(ASSETS_DIR)/heroicons/ ## 📦 Build frontend assets + generate templates
	templ generate

$(ASSETS_DIR):
	mkdir -p $(ASSETS_DIR)

# Tailwind CSS - build with standalone CLI
$(ASSETS_DIR)/css/tailwind.min.css: $(ASSETS_DIR) $(ASSETS_DIR)/css/styles.css tailwind.config.js
	mkdir -p $(ASSETS_DIR)/css
	curl -fsSL https://github.com/tailwindlabs/tailwindcss/releases/download/v$(TAILWIND_VERSION)/tailwindcss-linux-x64 -o /tmp/tailwindcss
	chmod +x /tmp/tailwindcss
	/tmp/tailwindcss -i $(ASSETS_DIR)/css/styles.css -o $(ASSETS_DIR)/css/tailwind.min.css --minify

# Download heroicons as SVG files
$(ASSETS_DIR)/heroicons/: $(ASSETS_DIR)
	mkdir -p $@
	curl -fsSL https://github.com/tailwindlabs/heroicons/archive/refs/tags/v$(HEROICONS_VERSION).tar.gz | \
		tar -xz --strip-components=3 -C $@ heroicons-$(HEROICONS_VERSION)/optimized/24/outline/
	mkdir -p $@/solid
	curl -fsSL https://github.com/tailwindlabs/heroicons/archive/refs/tags/v$(HEROICONS_VERSION).tar.gz | \
		tar -xz --strip-components=3 -C $@/solid heroicons-$(HEROICONS_VERSION)/optimized/24/solid/

.PHONY: install
install: ## Install dependencies
	go install github.com/a-h/templ/cmd/templ@latest

.PHONY: build
build: assets ## ⚒️ Build scry.quest
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/web

.PHONY: clean
clean: ## 🧹 Cleanup build artifacts
	go clean && rm -f $(BINARY_NAME) coverage.*
	rm -rf $(ASSETS_DIR)/css/tailwind.min.css $(ASSETS_DIR)/heroicons/

.PHONY: dev
dev: install assets ## 🚀 Start development server
	go run ./cmd/web

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
	go tool cover -html=coverage.out -o coverage.html && \
	echo "Coverage report saved to coverage.html" && \
	rm -f coverage.out

.PHONY: test-perf
test-perf: ## ⚡ Run benchmark tests
	go test -test.v -benchmem -bench=. -coverprofile=coverage-bench.out ./... && \
	go tool cover -html=coverage-bench.out -o coverage-bench.html && \
	echo "Coverage report saved to coverage-bench.html" && \
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
	@$(DOCKER) compose up -d

.PHONY: docker-down
docker-down: ## 🐳 Teardown docker compose services
	@$(DOCKER) compose down

.PHONY: psql
psql: ## 🐘 Connect to postgres dev
	@$(DOCKER) exec -it scry-quest-postgres psql -U scry_quest -d scry_quest_dev

.PHONY: embeddings
embeddings: ## 🔮 Run embeddings CLI
	go run ./cmd/embeddings $(filter-out $@,$(MAKECMDGOALS))

.PHONY: pre-commit
pre-commit: assets fmt tidy lint test sqlc-vet ## ✅ Run all checks

# pass through CLI flags to ./cmd/
%:
	@:
