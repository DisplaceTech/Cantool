# Contributing to Cantool

## Development Setup

```bash
git clone https://github.com/DisplaceTech/Cantool.git
cd Cantool
go mod download
make test
```

### Prerequisites

- Go 1.25+
- golangci-lint (for `make lint`)

## Code Style

- Run `gofmt` on all Go files (enforced by CI)
- Run `golangci-lint run ./...` before submitting
- No external mocking frameworks — use Go interfaces and hand-written mocks in `testutil/`

## Testing

- **Table-driven tests** for all functions with multiple scenarios
- **>99% statement coverage** on `internal/` packages (enforced by CI)
- **No flaky tests** — no `time.Sleep` in unit tests, no real network calls
- **No real SDK required** — all external commands mocked via `exec.CommandRunner`
- Integration tests use build tag `//go:build integration`

Run tests:

```bash
make test          # All tests with race detector
make coverage      # Generate coverage report
```

## Pull Request Process

1. Create a feature branch from `main`
2. Write tests first, then implementation
3. Ensure `make test` and `make lint` pass
4. Open a PR with a clear description of changes
5. Coverage must remain above 99% on `internal/` packages

## Project Structure

```
cantool/
├── main.go              # Entry point
├── cmd/                 # Cobra command definitions
├── internal/            # Private implementation packages
│   ├── config/          # cantool.yaml schema and loading
│   ├── exec/            # Process execution abstraction
│   ├── output/          # Output formatting (human/json/quiet)
│   ├── sdk/             # DAML SDK detection and interaction
│   ├── ledger/          # Canton Ledger API client
│   ├── scaffold/        # Project scaffolding with templates
│   ├── env/             # Environment profile management
│   ├── devserver/       # Local dev server orchestration
│   ├── doctor/          # Prerequisite checking
│   ├── mcp/             # MCP server for AI assistants
│   └── plugin/          # Plugin system
└── testutil/            # Shared test helpers
```
