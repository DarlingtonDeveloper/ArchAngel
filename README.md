# CodeHawk

CodeHawk is a smart linting and automated code review platform with VS Code integration, designed for developers of all skill levels. It helps maintain high-quality code by providing real-time feedback, suggestions, and automated fixes directly in your editor.

![CodeHawk Logo](codehawk/vscode-extension/media/codehawk-icon.svg)

## Features

- **Intelligent Linting**: Get real-time feedback on your code quality with language-specific linting rules
- **AI-Powered Suggestions**: Receive smart recommendations to improve your code
- **VS Code Integration**: Seamless integration with Visual Studio Code for a smooth development experience
- **Quick Fixes**: Apply automated fixes with a single click
- **Multi-language Support**: Works with JavaScript, TypeScript, Python, Go, Java, C#, PHP, and Ruby

## Project Structure

The CodeHawk project consists of two main components:

1. **VS Code Extension**: Integrates with Visual Studio Code to provide code analysis features directly in your editor
2. **Backend API**: Analyzes code and provides feedback, suggestions, and fixes

```
codehawk/
├── backend/              # Go backend for code analysis
│   ├── cmd/              # Application entry points
│   ├── internal/         # Private application code
│   ├── pkg/              # Public libraries
│   └── config/           # Configuration files
└── vscode-extension/     # VS Code extension
    ├── src/              # TypeScript source code
    │   ├── commands/     # Command implementations
    │   ├── providers/    # VS Code providers
    │   ├── utils/        # Utility functions
    │   └── views/        # UI components
    └── media/            # Icons and images
```

## Getting Started

### Prerequisites

- Node.js 14.x or higher
- Go 1.17 or higher
- Visual Studio Code
- Docker (optional, for containerized deployment)

### Setting Up the Backend

1. Navigate to the backend directory:
   ```bash
   cd codehawk/backend
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

   The API server will be available at http://localhost:8080.

4. (Optional) Using Docker:
   ```bash
   docker-compose up -d
   ```

### Setting Up the VS Code Extension

1. Navigate to the VS Code extension directory:
   ```bash
   cd codehawk/vscode-extension
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Build the extension:
   ```bash
   npm run compile
   ```

4. To launch the extension in development mode:
   - Open the directory in VS Code
   - Press F5 to start debugging
   - A new VS Code window will open with the extension loaded

## Using CodeHawk

Once installed, CodeHawk automatically provides linting and suggestions for supported languages. Here's how to use it:

1. **Analyze Current File**:
   - Click the "CodeHawk" icon in the status bar
   - Or use the command palette (`Ctrl+Shift+P`) and select "CodeHawk: Analyze Current File"

2. **Analyze Selected Code**:
   - Select a portion of code
   - Right-click and choose "CodeHawk: Analyze Selected Code"

3. **View Results**:
   - Issues appear as underlines in your code
   - Click the CodeHawk icon in the activity bar to see detailed analysis
   - The panel shows Errors, Warnings, and Suggestions

4. **Apply Fixes**:
   - Hover over an underlined issue to see details
   - Click the quick fix lightbulb to see available fixes
   - Select a fix to apply it automatically

## Configuration

You can configure CodeHawk in VS Code settings:

1. Open Settings (`Ctrl+,`)
2. Search for "CodeHawk"
3. Configure settings like:
   - API URL
   - API Key
   - Auto-analyze on save
   - Show inline issues

## Development Workflow

### Backend Development

1. Make changes to the Go code
2. Run tests:
   ```bash
   go test ./...
   ```
3. Start the server to test changes:
   ```bash
   go run cmd/server/main.go
   ```

### VS Code Extension Development

1. Make changes to the TypeScript code
2. Compile the extension:
   ```bash
   npm run compile
   ```
   or use watch mode:
   ```bash
   npm run watch
   ```
3. Press F5 in VS Code to launch with the extension

## Roadmap

- **Enhanced AI Suggestions**: More context-aware, intelligent suggestions
- **Custom Rules**: Allow users to define custom linting rules
- **Team Collaboration**: Share configurations across teams
- **CI/CD Integration**: Integrate with popular CI/CD tools
- **Additional Languages**: Support for more programming languages
- **Performance Metrics**: Code quality trends and improvement tracking

## Contributing

We welcome contributions to CodeHawk! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
