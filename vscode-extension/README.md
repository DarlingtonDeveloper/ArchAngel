# CodeHawk - Smart Code Analysis & AI-Powered Suggestions

![CodeHawk Logo](media/codehawk-icon.svg)

CodeHawk is a powerful VS Code extension that provides smart linting, automated code review, and AI-powered suggestions to help you write cleaner, more efficient code.

## Features

- **Multi-language Support**: Analyze code in JavaScript, TypeScript, Python, Go, Java, C#, PHP, and Ruby
- **Real-time Feedback**: Get instant feedback on your code quality with inline diagnostics
- **AI-Powered Suggestions**: Receive intelligent suggestions to improve your code
- **Quick Fixes**: Apply automated fixes with a single click
- **Integration with Popular Linters**: Uses industry-standard linting tools under the hood

![CodeHawk in Action](media/screenshot.png)

## Getting Started

### Prerequisites

- Visual Studio Code 1.60.0 or higher
- An active CodeHawk API account (or a self-hosted backend)

### Installation

1. Install the extension from the VS Code Marketplace
2. Configure your API key in the extension settings
3. Start analyzing your code!

## Usage

### Analyzing Code

#### Analyze Current File
- Click the CodeHawk icon in the status bar
- Run the command `CodeHawk: Analyze Current File` from the Command Palette (`Ctrl+Shift+P`)
- Use the shortcut key binding (`Alt+Shift+A` by default)

#### Analyze Selected Code
- Select a portion of code
- Right-click and choose `CodeHawk: Analyze Selected Code`
- Or run the command from the Command Palette

### Viewing Results

CodeHawk shows analysis results in multiple ways:

1. **Inline in the Editor**: Issues are underlined with different colors based on severity
2. **Problems Panel**: Issues appear in the VS Code Problems panel
3. **CodeHawk Panel**: Click the CodeHawk icon in the Activity Bar to see detailed results

### Applying Fixes

1. Hover over an underlined issue to see details
2. Click the lightbulb icon or press (`Ctrl+.`) to see available fixes
3. Select a fix to apply it automatically

## Configuration

### Extension Settings

This extension contributes the following settings:

* `codehawk.apiUrl`: URL of the CodeHawk API server
* `codehawk.apiKey`: API key for authenticating with the CodeHawk service
* `codehawk.autoAnalyze`: Automatically analyze files when they are saved
* `codehawk.showInlineIssues`: Show issues inline in the editor
* `codehawk.telemetryEnabled`: Enable anonymous usage data collection
* `codehawk.ignoredRules`: Rules to ignore during analysis
* `codehawk.severityLevels`: Mapping of API severity levels to VS Code DiagnosticSeverity

### Custom Configuration

For more advanced configuration, click the gear icon in the CodeHawk panel or run the `CodeHawk: Configure Settings` command.

## Self-Hosting the Backend

CodeHawk can work with a self-hosted backend. Follow these steps to set up your own server:

1. Clone the CodeHawk repository
2. Navigate to the `backend` directory
3. Start the server using Docker:
   ```bash
   docker-compose up -d
   ```
4. Configure the extension to use your server's URL

See the [Backend Documentation](https://github.com/yourusername/codehawk/blob/main/backend/README.md) for more details.

## Language Support

CodeHawk supports the following languages with varying levels of analysis:

| Language   | Linter         | Style Checking | Best Practices | AI Suggestions |
|------------|----------------|----------------|----------------|----------------|
| JavaScript | ESLint         | ✅             | ✅             | ✅             |
| TypeScript | ESLint+TSLint  | ✅             | ✅             | ✅             |
| Python     | Pylint         | ✅             | ✅             | ✅             |
| Go         | golangci-lint  | ✅             | ✅             | ✅             |
| Java       | CheckStyle     | ✅             | ✅             | ⚠️ (Limited)   |
| C#         | .NET Analyzer  | ✅             | ✅             | ⚠️ (Limited)   |
| PHP        | PHP_CodeSniffer| ✅             | ✅             | ⚠️ (Limited)   |
| Ruby       | RuboCop        | ✅             | ✅             | ⚠️ (Limited)   |

## Contributing

We welcome contributions to CodeHawk! Please see our [Contributing Guide](https://github.com/yourusername/codehawk/blob/main/CONTRIBUTING.md) for details.

## Privacy

CodeHawk respects your privacy:

- No code is stored permanently
- Analysis is performed securely
- Telemetry is anonymous and can be disabled
- All data transmission is encrypted

See our [Privacy Policy](https://codehawk.dev/privacy) for more information.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- [Documentation](https://codehawk.dev/docs)
- [FAQs](https://codehawk.dev/faq)
- [GitHub Issues](https://github.com/yourusername/codehawk/issues)
- [Email Support](mailto:support@codehawk.dev)

---

**CodeHawk** - Write better code, faster.
