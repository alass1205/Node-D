# Benchy - Ethereum Network Benchmarking Tool

.PHONY: build clean setup test-all help

# Variables
BINARY_NAME=benchy
BUILD_DIR=.
GO_FILES=$(shell find . -name "*.go" -type f)

# Build commands
build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@echo "🔨 Building Benchy..."
	go build -o $(BINARY_NAME) cmd/benchy/main.go
	@echo "✅ Build completed: ./$(BINARY_NAME)"

# Development commands
clean:
	@echo "🧹 Cleaning up..."
	rm -f $(BINARY_NAME)
	./scripts/cleanup.sh 2>/dev/null || true
	@echo "✅ Cleanup completed"

setup:
	@echo "🏗️  Setting up Benchy..."
	go mod tidy
	chmod +x scripts/*.sh
	@echo "✅ Setup completed"

# Testing commands
test-network:
	@echo "🧪 Testing network launch..."
	./$(BINARY_NAME) launch-network

test-monitoring:
	@echo "🧪 Testing monitoring..."
	./$(BINARY_NAME) infos

test-scenarios:
	@echo "🧪 Testing all scenarios..."
	./$(BINARY_NAME) scenario init
	./$(BINARY_NAME) scenario transfers
	./$(BINARY_NAME) scenario erc20
	./$(BINARY_NAME) scenario replacement

test-all: build test-network test-monitoring test-scenarios
	@echo "🎉 All tests completed!"

# Audit commands
audit-quick:
	@echo "🎯 Quick audit check..."
	@echo "✅ README.md: $$(test -f README.md && echo "Present" || echo "Missing")"
	@echo "✅ Binary: $$(test -f $(BINARY_NAME) && echo "Built" || echo "Not built")"
	@echo "✅ Containers: $$(docker ps | grep benchy | wc -l)/5 running"

audit-full: clean build test-all
	@echo "🏆 Full audit simulation completed!"

# Docker commands
docker-status:
	@echo "🐳 Docker status:"
	@docker ps | grep benchy || echo "No benchy containers running"

docker-logs:
	@echo "📋 Container logs:"
	@for container in $$(docker ps -a | grep benchy | awk '{print $$NF}'); do \
		echo "--- $$container ---"; \
		docker logs $$container 2>&1 | tail -5; \
		echo ""; \
	done

# Help
help:
	@echo "🚀 Benchy Development Commands:"
	@echo ""
	@echo "Build commands:"
	@echo "  make build       - Build the benchy binary"
	@echo "  make clean       - Clean up binary and containers"
	@echo "  make setup       - Setup development environment"
	@echo ""
	@echo "Testing commands:"
	@echo "  make test-all    - Run all tests"
	@echo "  make test-network    - Test network launch"
	@echo "  make test-monitoring - Test monitoring"
	@echo "  make test-scenarios  - Test all scenarios"
	@echo ""
	@echo "Audit commands:"
	@echo "  make audit-quick - Quick audit check"
	@echo "  make audit-full  - Full audit simulation"
	@echo ""
	@echo "Docker commands:"
	@echo "  make docker-status - Show container status"
	@echo "  make docker-logs   - Show container logs"
	@echo ""
	@echo "  make help        - Show this help"
