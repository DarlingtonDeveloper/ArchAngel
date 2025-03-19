# CodeHawk Quick Start Guide

Welcome to CodeHawk! This guide will help you get started quickly with analyzing your code and leveraging AI-powered suggestions to improve your programming.

## Installation

### VS Code Extension

1. Open Visual Studio Code
2. Go to the Extensions view (Ctrl+Shift+X)
3. Search for "CodeHawk"
4. Click "Install"

### Setting Up the API Connection

To use CodeHawk, you need to connect to the CodeHawk API:

1. Get an API key from [codehawk.dev/account](https://codehawk.dev/account) (or set up your own backend)
2. In VS Code, open the Command Palette (Ctrl+Shift+P)
3. Type "CodeHawk: Configure Settings" and press Enter
4. Enter your API URL and API key in the configuration panel

## First Analysis

Let's analyze your first file:

1. Open a code file in one of the supported languages (JavaScript, TypeScript, Python, Go, etc.)
2. Click the CodeHawk icon in the status bar or run the "CodeHawk: Analyze Current File" command
3. Wait a few seconds for the analysis to complete
4. You'll see issues underlined in the editor and listed in the Problems panel

## Understanding the Results

CodeHawk provides several types of feedback:

### Issues

Issues are potential problems in your code, categorized by severity:

- **Errors**: Serious problems that need to be fixed
- **Warnings**: Potential issues that could lead to bugs
- **Suggestions**: Style and best practice recommendations

### AI Suggestions

AI suggestions are intelligent recommendations to improve your code:

1. Click the CodeHawk icon in the Activity Bar (left sidebar)
2. Select the "AI Suggestions" view
3. Browse through the suggestions
4. Click "Apply" to automatically implement a suggestion

## Quick Fixes

Many issues come with quick fixes:

1. Hover over an underlined issue in your code
2. Click the lightbulb icon or press (Ctrl+.)
3. Select a fix from the menu to apply it

## Managing Configurations

### Auto-Analysis

You can configure CodeHawk to analyze your code automatically:

1. Open the CodeHawk settings
2. Enable "Auto Analyze on Save"
3. Your code will be analyzed each time you save a file

### Ignoring Rules

To ignore specific rules:

1. Open the CodeHawk settings panel
2. Navigate to the "Ignored Rules" section
3. Add the rule IDs you want to ignore
4. Or right-click on an issue and select "Ignore Rule"

## Supported Languages

CodeHawk works with multiple languages:

- **JavaScript/TypeScript**: Uses ESLint and TSLint rules
- **Python**: Analyzes with Pylint
- **Go**: Uses golangci-lint and staticcheck
- **Java**: Based on CheckStyle
- **C#**: Leverages .NET Analyzer
- **PHP**: Uses PHP_CodeSniffer
- **Ruby**: Implements RuboCop rules

## Getting Help

If you encounter any issues:

- Check the [Documentation](https://codehawk.dev/docs)
- Visit the [FAQs](https://codehawk.dev/faq)
- Submit an issue on [GitHub](https://github.com/yourusername/codehawk/issues)
- Contact [support@codehawk.dev](mailto:support@codehawk.dev)

## Next Steps

Now that you're up and running, try these next steps:

1. **Analyze an entire project**: Open a folder in VS Code and run analyses on multiple files
2. **Customize your settings**: Adjust severity levels, ignored rules, and other preferences
3. **Learn the shortcuts**: Set up key bindings for common CodeHawk commands
4. **Explore detailed reports**: Use the CodeHawk panel for in-depth analysis

Happy coding with CodeHawk!
