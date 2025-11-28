# GitHub Auto Delete

A CLI tool to automatically delete merged branches in GitHub repositories.

## Features

- Automatically identifies and deletes merged branches
- Configurable branch protection rules
- Supports multiple repositories
- Token-based GitHub authentication

## Installation

```bash
go install github.com/josejulio/ghautodelete/cmd/ghautodelete@latest
```

## Usage

```bash
ghautodelete [command] [flags]
```

## Development

### Prerequisites

- Go 1.21 or later
- golangci-lint (for linting)

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

### Coverage

```bash
make coverage
```

## License

TBD
