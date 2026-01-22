.PHONY: build test clean fmt vet lint run help

build: ## Build the binaries
	go build -o git-commit-analysis ./cmd/git-commit-analysis
	go build -o mcp-server ./cmd/mcp-server

test: ## Run tests
	go test ./...

clean: ## Remove artifacts
	rm -f git-commit-analysis mcp-server

fmt: ## Format code
	go fmt ./...

vet: ## Vet code
	go vet ./...

lint: ## Lint code (requires golangci-lint)
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Skipping linting."; \
		echo "Install it via: brew install golangci-lint (or check https://golangci-lint.run/usage/install/)"; \
	fi

run: ## Run the tool (use ARGS="..." to pass arguments)
	go run ./cmd/git-commit-analysis $(ARGS)

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | sed 's/:.*## /    /'
