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

.PHONY: pre-commit
pre-commit: fmt tidy lint test ## ✅ Run all checks
