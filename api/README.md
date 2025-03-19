# CodeHawk API Documentation

This directory contains documentation for the CodeHawk API, which powers the CodeHawk platform for code analysis, linting, and AI-powered code suggestions.

## Overview

The CodeHawk API is a RESTful API that allows developers to:

- Submit code for analysis
- Retrieve linting issues and suggestions
- Manage organizations and users
- Configure custom rules and integrations

## API Specification

The API is documented using the OpenAPI (Swagger) specification in the [openapi.yaml](openapi.yaml) file. You can view this specification in any OpenAPI viewer or import it into tools like Postman for API testing.

## Base URL

- Production: `https://api.codehawk.dev/api/v1`
- Development: `http://localhost:8080/api/v1`

## Authentication

The API supports two authentication methods:

1. **API Key**: Pass your API key in the `X-API-Key` header
2. **Bearer Token**: Use a JWT token in the `Authorization` header with the format `Bearer {token}`

## Getting Started

1. Register an account at [codehawk.dev](https://codehawk.dev)
2. Generate an API key in your account settings
3. Use the API key to make requests to the API

## Examples

### Analyzing Code

```bash
curl -X POST https://api.codehawk.dev/api/v1/analyze \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "def hello_world():\n    print(\"Hello, World!\")",
    "language": "python",
    "context": "Example function"
  }'
```

### Getting Analysis Results

```bash
curl -X GET https://api.codehawk.dev/api/v1/analysis/analysis-123 \
  -H "X-API-Key: your-api-key"
```

## Rate Limits

- Free tier: 100 requests per day
- Pro tier: 1,000 requests per day
- Enterprise tier: Custom limits

## SDKs

We provide official SDKs for the following languages:

- JavaScript/TypeScript: `npm install codehawk-sdk`
- Python: `pip install codehawk-sdk`
- Go: `go get github.com/codehawk/codehawk-sdk-go`

## Support

If you have any questions or issues with the API, please contact [support@codehawk.dev](mailto:support@codehawk.dev).

## Changelog

See the [CHANGELOG.md](../CHANGELOG.md) file for a list of changes to the API.

## License

The CodeHawk API is licensed under the MIT License. See the [LICENSE](../LICENSE) file for details.