# Krustron Makefile
# Author: Anubhav Gain <anubhavg@infopercept.com>

.PHONY: all build run test clean docker helm lint fmt vet swagger proto install deps dev

# Build variables
BINARY_NAME=krustron
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X github.com/anubhavg-icpl/krustron/pkg/version.Version=$(VERSION) \
                  -X github.com/anubhavg-icpl/krustron/pkg/version.GitCommit=$(GIT_COMMIT) \
                  -X github.com/anubhavg-icpl/krustron/pkg/version.BuildTime=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Docker parameters
DOCKER_IMAGE=ghcr.io/anubhavg-icpl/krustron
DOCKER_TAG?=$(VERSION)

# Directories
CMD_DIR=./cmd/krustron
BUILD_DIR=./build
COVERAGE_DIR=./coverage

all: deps lint test build

# Install dependencies
deps:
	@echo ">>> Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Build the binary
build:
	@echo ">>> Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo ">>> Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)

build-darwin:
	@echo ">>> Building for macOS..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)

build-windows:
	@echo ">>> Building for Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

# Run the application
run:
	@echo ">>> Running $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	$(BUILD_DIR)/$(BINARY_NAME) serve

# Development mode with hot reload (requires air)
dev:
	@echo ">>> Starting development server..."
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air

# Run tests
test:
	@echo ">>> Running tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html

# Run unit tests only
test-unit:
	@echo ">>> Running unit tests..."
	$(GOTEST) -v -short ./...

# Run integration tests
test-integration:
	@echo ">>> Running integration tests..."
	$(GOTEST) -v -run Integration ./tests/integration/...

# Run e2e tests
test-e2e:
	@echo ">>> Running e2e tests..."
	$(GOTEST) -v -run E2E ./tests/e2e/...

# Linting
lint:
	@echo ">>> Running linter..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Format code
fmt:
	@echo ">>> Formatting code..."
	$(GOFMT) -s -w .

# Vet code
vet:
	@echo ">>> Vetting code..."
	$(GOVET) ./...

# Generate Swagger docs
swagger:
	@echo ">>> Generating Swagger documentation..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/krustron/main.go -o ./api/docs

# Generate protobuf
proto:
	@echo ">>> Generating protobuf files..."
	protoc --go_out=. --go-grpc_out=. api/grpc/proto/*.proto

# Clean build artifacts
clean:
	@echo ">>> Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)

# Docker build
docker-build:
	@echo ">>> Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

# Docker push
docker-push:
	@echo ">>> Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

# Helm package
helm-package:
	@echo ">>> Packaging Helm chart..."
	helm package ./charts/krustron -d $(BUILD_DIR)

# Helm install (local)
helm-install:
	@echo ">>> Installing Helm chart..."
	helm upgrade --install krustron ./charts/krustron --namespace krustron --create-namespace

# Helm uninstall
helm-uninstall:
	@echo ">>> Uninstalling Helm chart..."
	helm uninstall krustron --namespace krustron

# Generate mocks for testing
mocks:
	@echo ">>> Generating mocks..."
	@which mockgen > /dev/null || go install github.com/golang/mock/mockgen@latest
	go generate ./...

# Security scan
security:
	@echo ">>> Running security scan..."
	@which gosec > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

# Database migrations
migrate-up:
	@echo ">>> Running database migrations..."
	$(BUILD_DIR)/$(BINARY_NAME) migrate up

migrate-down:
	@echo ">>> Rolling back database migrations..."
	$(BUILD_DIR)/$(BINARY_NAME) migrate down

# Install pre-commit hooks
install-hooks:
	@echo ">>> Installing pre-commit hooks..."
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

# Help
help:
	@echo "Krustron - Kubernetes Native Platform"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Install deps, lint, test, and build"
	@echo "  deps           - Install dependencies"
	@echo "  build          - Build the binary"
	@echo "  build-all      - Build for all platforms"
	@echo "  run            - Build and run the application"
	@echo "  dev            - Run with hot reload (requires air)"
	@echo "  test           - Run all tests with coverage"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run end-to-end tests"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo "  swagger        - Generate Swagger docs"
	@echo "  proto          - Generate protobuf files"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-push    - Push Docker image"
	@echo "  helm-package   - Package Helm chart"
	@echo "  helm-install   - Install Helm chart locally"
	@echo "  helm-uninstall - Uninstall Helm chart"
	@echo "  mocks          - Generate mocks"
	@echo "  security       - Run security scan"
	@echo "  help           - Show this help"
