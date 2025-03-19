# CodeHawk Development Setup

This guide will help you set up a local development environment for the CodeHawk project. It covers both the backend API service and the VS Code extension.

## Prerequisites

Before you begin, make sure you have the following installed:

- **Go** (version 1.18 or higher) - [Installation Guide](https://golang.org/doc/install)
- **Node.js** (version 16 or higher) - [Installation Guide](https://nodejs.org/)
- **npm** (comes with Node.js)
- **Docker** and **Docker Compose** - [Installation Guide](https://docs.docker.com/get-docker/)
- **Git** - [Installation Guide](https://git-scm.com/downloads)
- **Visual Studio Code** - [Download](https://code.visualstudio.com/)

### Optional Tools

- **golangci-lint** - `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **mockgen** - `go install github.com/golang/mock/mockgen@latest`
- **sqlc** - `go install github.com/kyleconroy/sqlc/cmd/sqlc@latest`
- **vsce** - `npm install -g vsce`

## Project Structure

```
codehawk/
├── api/                  # API documentation
├── backend/              # Backend service (Go)
│   ├── cmd/              # Application entry points
│   ├── internal/         # Private application code
│   ├── pkg/              # Public libraries
│   └── ...
├── docs/                 # Documentation
├── kubernetes/           # Kubernetes manifests
├── vscode-extension/     # VS Code extension (TypeScript)
│   ├── src/              # Source code
│   └── ...
└── ...
```

## Setting Up the Development Environment

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/codehawk.git
cd codehawk
```

### 2. Backend Setup

#### Starting with Docker Compose (Recommended for First Time)

The easiest way to get started is using Docker Compose, which sets up the backend API, PostgreSQL database, and Redis cache:

```bash
cd codehawk
docker-compose up -d
```

This will:
- Start the PostgreSQL database on port 5432
- Start the Redis cache on port 6379
- Start the CodeHawk API on port 8080

#### Local Backend Development

If you want to develop the backend service locally:

1. Start the database and Redis:

```bash
cd codehawk
docker-compose up -d postgres redis
```

2. Set up the environment variables:

```bash
cd backend
cp .env.example .env
# Edit .env with appropriate values
```

3. Run the backend service:

```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### Backend Development Tips

- **Run tests**: `go test ./...`
- **Run linter**: `golangci-lint run`
- **Generate mocks**: `go generate ./...`
- **Build binary**: `go build -o bin/codehawk-api cmd/server/main.go`

### 3. VS Code Extension Setup

1. Install dependencies:

```bash
cd codehawk/vscode-extension
npm install
```

2. Build the extension:

```bash
npm run compile
```

3. Launch the extension in debug mode:
   - Open the project in VS Code
   - Press `F5` to start debugging
   - A new VS Code window will open with the extension running

#### VS Code Extension Development Tips

- **Watch mode**: `npm run watch` - automatically recompiles on changes
- **Run tests**: `npm run test`
- **Lint code**: `npm run lint`
- **Package extension**: `npm run package`

### 4. Configuring the Extension to Use Your Local Backend

When developing, you'll want the VS Code extension to communicate with your local backend:

1. In the VS Code instance running your extension:
   - Open Settings (`Ctrl+,`)
   - Search for "CodeHawk"
   - Set `codehawk.apiUrl` to `http://localhost:8080/api/v1`
   - Set `codehawk.apiKey` to your local development API key (default is `demo_api_key_123`)

## Dependency Management

### Backend Dependencies

- We use Go modules for dependency management
- Add new dependencies with `go get <package>`
- Update dependencies with `go get -u` or `go get -u=patch`
- Run `go mod tidy` to clean up the module file

### Frontend Dependencies

- We use npm for dependency management
- Add new dependencies with `npm install <package>`
- Add development dependencies with `npm install --save-dev <package>`
- Update dependencies with `npm update`

## Database Migrations

Database migrations are handled with golang-migrate:

1. Install the migrate tool:
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

2. Create a new migration:
   ```bash
   cd backend
   migrate create -ext sql -dir db/migrations -seq new_migration_name
   ```

3. Apply migrations:
   ```bash
   migrate -database "postgres://postgres:postgres@localhost:5432/codehawk?sslmode=disable" -path db/migrations up
   ```

## Testing

### Backend Testing

Run all tests:
```bash
cd backend
go test ./...
```

Run tests with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### VS Code Extension Testing

Run tests:
```bash
cd vscode-extension
npm run test
```

## Debugging

### Debugging the Backend

1. **Using VS Code**:
   - Open the backend folder in VS Code
   - Click on the Run and Debug tab
   - Select "Launch Backend" configuration
   - Press F5 to start debugging

2. **Using Delve**:
   ```bash
   cd backend
   dlv debug cmd/server/main.go
   ```

### Debugging the VS Code Extension

1. Press F5 in VS Code with the extension project open
2. A new VS Code window will open with the extension running in debug mode
3. Set breakpoints in the TypeScript code
4. Interact with the extension to trigger your breakpoints

## Common Issues and Solutions

### Backend

#### "Permission denied" when running the backend

Make sure you have the correct permissions:
```bash
chmod +x scripts/*.sh
```

#### Database connection issues

Check your database connection parameters in `.env` file or verify that PostgreSQL is running:
```bash
docker ps | grep postgres
```

### VS Code Extension

#### "Cannot find module" errors

Try cleaning and reinstalling node modules:
```bash
cd vscode-extension
rm -rf node_modules
npm install
```

#### Extension not showing up in VS Code

Make sure you're running VS Code with the correct extension host:
```bash
# From vscode-extension directory
code --extensionDevelopmentPath=$(pwd)
```

## Code Style and Guidelines

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Run `gofmt -s -w .` to format code
- Use `golangci-lint` to check for issues

### TypeScript Code Style

- We follow the [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)
- Run `npm run lint` to check for issues
- Run `npm run lint -- --fix` to automatically fix issues

## Pull Request Process

1. Create a new branch from `develop` with a descriptive name
2. Make your changes, with appropriate tests
3. Ensure all tests pass and linting issues are fixed
4. Update documentation if necessary
5. Submit a pull request to the `develop` branch
6. Request review from at least one team member

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [VS Code Extension API](https://code.visualstudio.com/api)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [TypeScript Documentation](https://www.typescriptlang.org/docs/)