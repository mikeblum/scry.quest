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

.PHONY: test
test: ## 🧪 Run all tests
	go test -test.v -race -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -html=coverage.out && rm coverage.out

.PHONY: test-perf
test-perf: ## ⚡ Run benchmark tests
	go test -test.v -benchmem -bench=. -coverprofile=coverage-bench.out ./... && go tool cover -html=coverage-bench.out && rm coverage-bench.out

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

.PHONY: embeddings-generate
embeddings-generate: ## 🤖 Generate embeddings for all SRD content
	go run ./cmd/embeddings -command=generate -config=env/.env.local

.PHONY: embeddings-generate-spells
embeddings-generate-spells: ## 🔮 Generate embeddings for spells
	go run ./cmd/embeddings -command=generate -type=spell -config=env/.env.local

.PHONY: embeddings-generate-bestiary
embeddings-generate-bestiary: ## 🐉 Generate embeddings for bestiary
	go run ./cmd/embeddings -command=generate -type=bestiary -config=env/.env.local

.PHONY: embeddings-generate-classes
embeddings-generate-classes: ## ⚔️ Generate embeddings for classes
	go run ./cmd/embeddings -command=generate -type=class -config=env/.env.local

.PHONY: embeddings-generate-species
embeddings-generate-species: ## 🧝 Generate embeddings for species
	go run ./cmd/embeddings -command=generate -type=species -config=env/.env.local

.PHONY: embeddings-search
embeddings-search: ## 🔍 Search content using embeddings (requires QUERY)
	go run ./cmd/embeddings -command=search -query="$(QUERY)" -config=env/.env.local

.PHONY: embeddings-stats
embeddings-stats: ## 📊 Show embedding statistics
	go run ./cmd/embeddings -command=stats -config=env/.env.local

.PHONY: embeddings-clear
embeddings-clear: ## 🗑️ Clear embeddings for a model (requires MODEL)
	go run ./cmd/embeddings -command=clear -model="$(MODEL)" -config=env/.env.local

.PHONY: embeddings-clear-all
embeddings-clear-all: ## 🗑️ Clear all embeddings
	go run ./cmd/embeddings -command=clear -model=all -config=env/.env.local

.PHONY: pre-commit
pre-commit: fmt tidy lint test sqlc-vet ## ✅ Run all checks
