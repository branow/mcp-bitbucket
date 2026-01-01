# Project Aurora

Project Aurora is a small experimental service designed to explore clean API design, testing strategies, and integration patterns. The repository serves as a sandbox for trying ideas quickly without over-engineering.

## Features

- Simple HTTP API with JSON responses
- Configurable via environment variables
- Built-in health check endpoint
- Basic OAuth-aware request handling (for experimentation)
- End-to-end and integration test examples

## Getting Started

### Prerequisites

- Go 1.22 or newer
- Make (optional)
- Docker (optional, for local testing)

### Installation

Clone the repository and download dependencies:

```bash
git clone https://example.com/project-aurora.git
cd project-aurora
go mod download
```

### Running the Service

```bash
go run ./cmd/server
```

By default, the server listens on `http://127.0.0.1:8080`.

### Configuration

Configuration is done via environment variables:

| Variable | Description | Default |
|--------|-------------|---------|
| `SERVER_PORT` | Port to bind the HTTP server | `8080` |
| `LOG_LEVEL` | Log verbosity (`debug`, `info`, `warn`) | `info` |
| `OAUTH_ENABLED` | Enable OAuth handling | `false` |

## Testing

Run all tests:

```bash
go test ./...
```

Run only end-to-end tests:

```bash
go test ./internal/e2e -v
```

## Project Structure

```
.
├── cmd/            # Application entry points
├── internal/       # Private application code
│   ├── api/        # HTTP handlers
│   ├── auth/       # Authentication helpers
│   └── service/    # Core business logic
├── pkg/            # Reusable libraries
└── README.md
```

## Design Notes

- The project favors clarity over abstraction.
- Public interfaces are kept small and explicit.
- Tests are written close to the code they verify.

## Roadmap

- Add structured logging
- Improve configuration validation
- Expand test coverage
- Provide example client implementations

## Contributing

Contributions are welcome. Please open an issue before submitting large changes to discuss the approach.

## License

This project is licensed under the MIT License.