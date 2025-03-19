# Component Communication in CodeHawk

This document describes how different components of the CodeHawk platform communicate with each other, including data formats, authentication flow, and interaction patterns.

## System Components

CodeHawk consists of several key components that need to communicate:

1. **VS Code Extension**: The client-side component running in the user's editor
2. **API Server**: The central backend service that processes requests
3. **Linting Engine**: Backend component that performs code analysis
4. **AI Service**: Component that generates intelligent code suggestions
5. **Database**: Stores analysis history, user data, and configurations

## Communication Flows

### 1. VS Code Extension to API Server

The VS Code extension communicates with the API server using RESTful HTTP requests.

#### Authentication Flow

1. **Initial Authentication**:
   - User enters API key in extension settings
   - Extension stores the API key securely in the VS Code secrets storage
   - All subsequent requests include the API key in the `X-API-Key` header

2. **JWT Authentication** (Enterprise version):
   - User logs in via the extension
   - Extension receives a JWT token
   - Token is stored securely and included in the `Authorization` header
   - Token is refreshed automatically before expiration

#### Data Formats

- **Code Analysis Request**:
  ```json
  {
    "code": "function example() { console.log('test'); }",
    "language": "javascript",
    "context": "Function in utils.js",
    "options": {
      "rules": {
        "no-console": "warning"
      }
    }
  }
  ```

- **Analysis Response**:
  ```json
  {
    "id": "analysis-123",
    "status": "success",
    "language": "javascript",
    "timestamp": "2023-06-15T10:30:00Z",
    "issues": [
      {
        "line": 1,
        "column": 20,
        "message": "Unexpected console statement",
        "severity": "warning",
        "ruleId": "no-console",
        "fix": {
          "description": "Remove console.log statement",
          "replacement": "function example() { }"
        }
      }
    ],
    "suggestions": [
      {
        "line": 1,
        "message": "Consider adding a function description",
        "severity": "suggestion",
        "fix": {
          "description": "Add JSDoc comment",
          "replacement": "/**\n * Example function\n */\nfunction example() { console.log('test'); }"
        }
      }
    ]
  }
  ```

#### Error Handling

Errors follow a standard format:
```json
{
  "status": "error",
  "message": "Invalid language specified",
  "code": "INVALID_LANGUAGE"
}
```

Common error codes include:
- `AUTHENTICATION_FAILED`: Invalid or missing API key
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INVALID_LANGUAGE`: Unsupported language
- `SERVER_ERROR`: Internal server error

#### Retry Logic

The extension implements exponential backoff retry logic for transient errors:
- Initial retry after 500ms
- Double the wait time for each retry (max 5 retries)
- Don't retry for authentication or validation errors

### 2. API Server to Linting Engine

The API server communicates with the linting engine via internal gRPC calls or direct function calls, depending on the deployment architecture.

#### Message Formats

- **Linting Request**:
  ```protobuf
  message LintingRequest {
    string code = 1;
    string language = 2;
    string context = 3;
    map<string, RuleConfig> rules = 4;
  }
  
  message RuleConfig {
    string severity = 1;
    map<string, string> options = 2;
  }
  ```

- **Linting Response**:
  ```protobuf
  message LintingResponse {
    repeated Issue issues = 1;
    string linter_version = 2;
    int32 elapsed_ms = 3;
  }
  
  message Issue {
    int32 line = 1;
    int32 column = 2;
    string message = 3;
    string severity = 4;
    string rule_id = 5;
    string context = 6;
    Fix fix = 7;
  }
  ```

#### Performance Optimization

- Linting requests are processed in parallel for different languages
- Results are streamed as they become available
- Language-specific linters run in isolated containers for security

### 3. API Server to AI Service

The API server communicates with the AI service through a message queue (Redis) for asynchronous processing and direct HTTP for synchronous requests.

#### Asynchronous Processing

For AI suggestions that may take longer to generate:

1. API server places a request in the queue
2. AI service processes the request
3. AI service publishes results to a results queue
4. API server retrieves results and stores them
5. Client can poll for suggestion status

#### Message Format

- **AI Request**:
  ```json
  {
    "analysis_id": "analysis-123",
    "code": "function example() { console.log('test'); }",
    "language": "javascript",
    "issues": [...],  // Issues found by linting
    "context": "Function in utils.js",
    "model": "gpt-4",
    "max_suggestions": 5
  }
  ```

- **AI Response**:
  ```json
  {
    "analysis_id": "analysis-123",
    "suggestions": [
      {
        "line": 1,
        "message": "Consider adding a function description",
        "severity": "suggestion",
        "confidence": 0.92,
        "fix": {
          "description": "Add JSDoc comment",
          "replacement": "/**\n * Example function\n */\nfunction example() { console.log('test'); }"
        },
        "reasoning": "Adding documentation improves code readability and maintainability."
      }
    ],
    "model_used": "gpt-4",
    "processing_time_ms": 1250
  }
  ```

### 4. Database Communication

All components communicate with the database through a data access layer that abstracts the underlying database technology.

#### Data Models

The key data models include:

- **Analysis**: Represents a code analysis request and result
- **Issue**: Represents a specific issue found in the code
- **Suggestion**: Represents an AI-generated suggestion
- **User**: Represents a user of the system
- **Organization**: Represents a group of users
- **Rule**: Represents a linting or suggestion rule

#### Access Patterns

Common access patterns include:

- Get analysis by ID
- Get analyses for a user or organization
- Save analysis results
- Get rule configurations for a language
- Get user or organization data

## Security Considerations

### Data in Transit

- All communication uses HTTPS/TLS 1.3
- Internal services use mutual TLS authentication
- API keys and tokens are never logged or stored unencrypted

### Request Validation

- All requests are validated against schemas
- Input sanitization is performed on all user input
- Rate limiting is applied based on API key or IP address

### Authentication and Authorization

- API keys are validated for every request
- JWT tokens include scopes for fine-grained permissions
- Database queries include tenant isolation

## Monitoring and Tracing

All component communications include:

- Correlation IDs to trace requests across services
- Timing information for performance monitoring
- Structured logging for error tracking
- Health check endpoints for service monitoring

## Scalability

The communication architecture is designed to scale:

- Stateless API servers can be horizontally scaled
- Linting engines run as separate containers that can be scaled independently
- Queue-based architecture allows for distributing AI processing
- Database access uses connection pooling for efficient scaling

## Offline Mode (VS Code Extension)

The VS Code extension supports limited offline functionality:

1. Basic linting using bundled linters
2. Previously cached suggestions
3. Local rule configurations
4. Queuing of analysis requests for when connectivity is restored

## Cross-Version Compatibility

Communication protocols include version information to ensure backward compatibility:

- API endpoints include version in the URL path
- Request/response schemas include version fields
- New fields are added in a backward-compatible way
- Deprecated fields are marked but still supported for a transition period