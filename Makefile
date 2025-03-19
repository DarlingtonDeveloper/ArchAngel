.PHONY: all build test clean backend extension docs docker-up docker-down

# Default target
all: build

# Build everything
build: backend extension

# Run tests
test: test-backend test-extension

# Clean all build artifacts
clean: clean-backend clean-extension

# Backend targets
backend:
	@echo "Building backend..."
	@cd codehawk/backend && $(MAKE) build

test-backend:
	@echo "Testing backend..."
	@cd codehawk/backend && $(MAKE) test

clean-backend:
	@echo "Cleaning backend..."
	@cd codehawk/backend && $(MAKE) clean

# VS Code extension targets
extension:
	@echo "Building VS Code extension..."
	@cd codehawk/vscode-extension && npm run compile

test-extension:
	@echo "Testing VS Code extension..."
	@cd codehawk/vscode-extension && npm run test

clean-extension:
	@echo "Cleaning VS Code extension..."
	@rm -rf codehawk/vscode-extension/out
	@rm -rf codehawk/vscode-extension/*.vsix

package-extension:
	@echo "Packaging VS Code extension..."
	@cd codehawk/vscode-extension && npm run package

# Documentation
docs:
	@echo "Building documentation..."
	@cd docs && mkdocs build

# Docker
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

# CI/CD
ci-build: backend test-backend package-extension
	@echo "CI build completed successfully"

# Deployment
deploy-backend:
	@echo "Deploying backend..."
	@cd codehawk/backend && $(MAKE) deploy

deploy-extension:
	@echo "Publishing VS Code extension..."
	@cd codehawk/vscode-extension && npm run publish

# Help
help:
	@echo "Available targets:"
	@echo "  all              Build everything (default)"
	@echo "  build            Build backend and extension"
	@echo "  test             Run all tests"
	@echo "  clean            Clean all build artifacts"
	@echo "  backend          Build backend only"
	@echo "  test-backend     Test backend only"
	@echo "  clean-backend    Clean backend artifacts"
	@echo "  extension        Build VS Code extension only"
	@echo "  test-extension   Test VS Code extension only"
	@echo "  clean-extension  Clean VS Code extension artifacts"
	@echo "  package-extension Package VS Code extension for distribution"
	@echo "  docs             Build documentation"
	@echo "  docker-up        Start Docker services"
	@echo "  docker-down      Stop Docker services"
	@echo "  ci-build         Run CI build process"
	@echo "  deploy-backend   Deploy backend to production"
	@echo "  deploy-extension Publish VS Code extension to marketplace"