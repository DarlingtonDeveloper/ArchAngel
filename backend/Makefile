.PHONY: all build test lint clean deps run docker-build deploy

# Variables
BINARY_NAME=codehawk-api
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/yourusername/codehawk/backend/internal/version.Version=$(VERSION) -X github.com/yourusername/codehawk/backend/internal/version.BuildTime=$(BUILD_TIME)"

# Go environment
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOLINT=golangci-lint

# Directories
CMD_DIR=./cmd/server
PKG_DIR=./pkg
INTERNAL_DIR=./internal
CONFIG_DIR=./config

# Docker
DOCKER_IMAGE=codehawk/api
DOCKER_TAG=$(VERSION)

# Default target
all: deps lint test build

# Build the application
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) $(CMD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./... -coverprofile=coverage.out
	@echo "Coverage report:"
	$(GOCMD) tool cover -func=coverage.out

# Run code linting
lint:
	@echo "Linting code..."
	$(GOFMT) ./...
	$(GOLINT) run

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	fi

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GORUN) $(LDFLAGS) $(CMD_DIR)

# Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -f Dockerfile .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

# Deploy
deploy: docker-build
	@echo "Deploying $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	# Example deployment commands (modify as needed)
	# docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	# docker push $(DOCKER_IMAGE):latest
	# kubectl apply -f kubernetes/deployment.yaml

# Helper targets
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. ./...

coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	open coverage.html

# Help
help:
	@echo "Available targets:"
	@echo "  all           Build, test, and lint (default)"
	@echo "  build         Build the application"
	@echo "  test          Run tests"
	@echo "  lint          Run linters"
	@echo "  clean         Clean build artifacts"
	@echo "  deps          Install dependencies"
	@echo "  run           Run the application locally"
	@echo "  docker-build  Build Docker image"
	@echo "  deploy        Deploy the application"
	@echo "  vet           Run go vet"
	@echo "  fmt           Format code"
	@echo "  bench         Run benchmarks"
	@echo "  coverage      Generate coverage report"