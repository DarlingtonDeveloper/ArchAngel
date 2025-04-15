# ArchAngel: Smart Linting & Automated Code Review Platform

<p align="center">
  <img src="docs/images/ArchAngel-logo.png" alt="ArchAngel Logo" width="200"/>
  <br>
  <em>Write better code, faster.</em>
</p>

[![CI Status](https://github.com/yourusername/ArchAngel/actions/workflows/ci.yml/badge.svg)](https://github.com/yourusername/ArchAngel/actions/workflows/ci.yml)
[![VS Code Extension](https://img.shields.io/visual-studio-marketplace/v/ArchAngel.ArchAngel)](https://marketplace.visualstudio.com/items?itemName=ArchAngel.ArchAngel)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Overview

ArchAngel is a comprehensive code quality platform designed for modern development environments, particularly focused on helping teams who use AI-assisted coding ("vibe coding") to maintain high-quality standards. ArchAngel integrates directly into your development workflow through a VS Code extension and CI/CD integrations, providing real-time feedback, smart suggestions, and automated fixes.

### The Problem We Solve

As AI tools become more common in code generation, traditional code review processes are being challenged. ArchAngel addresses:

1. **Inconsistent Quality**: Ensuring AI-generated code follows best practices
2. **Scalability of Reviews**: Automating repetitive code quality checks
3. **Knowledge Gaps**: Providing context-aware suggestions to improve code
4. **Integration Challenges**: Seamlessly working across the development lifecycle

## System Architecture

ArchAngel follows a modern, distributed architecture:

<p align="center">
  <img src="docs/images/architecture-diagram.png" alt="ArchAngel Architecture" width="700"/>
</p>

### Core Components

- **VS Code Extension**: Delivers real-time feedback within your editor
- **API Server**: Processes code analysis requests and returns actionable feedback
- **Linting Engine**: Language-specific code analysis using best-in-class linters
- **AI Suggestion Service**: Generates intelligent recommendations for code improvements
- **Database**: Stores analysis history, user preferences, and team configurations

## Key Features

- **Multi-language Support**: JavaScript, TypeScript, Python, Go, Java, C#, PHP, Ruby
- **Real-time Analysis**: Immediate feedback as you code
- **AI-Powered Suggestions**: Context-aware recommendations that go beyond traditional linting
- **Quick Fixes**: One-click solutions for common issues
- **Team Collaboration**: Share configurations and best practices
- **CI/CD Integration**: Automated checks for your pipelines
- **Enterprise Controls**: Custom rules, compliance checking, and security scans

## Target Audience

ArchAngel is designed for:

- **Individual Developers**: Who want to improve their code quality
- **Development Teams**: Seeking consistent standards across the team
- **Organizations**: Implementing code quality at scale
- **AI Tool Users**: Ensuring AI-generated code meets quality standards

## Getting Started

- [Installation Guide](docs/installation.md)
- [Quick Start](docs/quickstart.md)
- [VS Code Extension](ArchAngel/vscode-extension/README.md)
- [API Documentation](api/README.md)
- [Enterprise Guide](docs/enterprise.md)

## Development

- [Development Setup](DEVELOPMENT.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [Architecture Decisions](docs/adr/README.md)
- [Testing Guidelines](docs/testing.md)

## Community & Support

- [Join Our Discord](https://discord.gg/ArchAngel)
- [Report an Issue](https://github.com/yourusername/ArchAngel/issues)
- [Request a Feature](https://github.com/yourusername/ArchAngel/issues/new?template=feature_request.md)
- [Email Support](mailto:support@ArchAngel.dev)

## License

ArchAngel is available under the [MIT License](LICENSE).
