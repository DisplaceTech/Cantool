# Cantool

[![CI](https://github.com/DisplaceTech/Cantool/actions/workflows/ci.yaml/badge.svg)](https://github.com/DisplaceTech/Cantool/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/displacetech/cantool)](https://goreportcard.com/report/github.com/displacetech/cantool)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Canton application development CLI — project scaffolding, DAML compilation, testing, local sandbox management, and MCP server for AI-assisted development.

## Why Cantool

- **Single binary, zero dependencies.** No Node.js, no npm, no JVM for the CLI itself. Download, chmod, run. Institutional environments with locked-down package managers get a tool that works out of the box.
- **Cross-compilation.** `GOOS=linux GOARCH=amd64 go build` produces a Linux binary from macOS. CI builds for all platforms trivially.
- **Static linking.** No shared library conflicts, no version mismatches, no container image bloat. Copy the binary into an air-gapped environment and it works.
- **Plugin system.** Extensible via JSON-RPC plugins. Core commands prove the plugin contract before any external plugins ship.

## Installation

```bash
# Homebrew (macOS)
brew install displacetech/tap/cantool

# Binary download (preferred for institutions)
curl -sSL https://github.com/DisplaceTech/Cantool/releases/latest/download/cantool-$(uname -s)-$(uname -m) -o cantool
chmod +x cantool && sudo mv cantool /usr/local/bin/

# Go install (for Go developers)
go install github.com/DisplaceTech/Cantool@latest
```

## Quick Start

```bash
# Check prerequisites
cantool doctor

# Create a new project
cantool init my-app --template basic
cd my-app

# Build and test
cantool build
cantool test

# Start local sandbox with hot-reload
cantool dev
```

## Commands

| Command | Description |
|---------|-------------|
| `cantool init [name]` | Create a new Canton project from a template (`basic` or `token`) |
| `cantool build [--watch]` | Compile DAML sources. `--watch` rebuilds on file changes |
| `cantool test [--verbose]` | Run DAML Script tests with structured output |
| `cantool dev [--port P]` | Start local sandbox with hot-reload and party provisioning |
| `cantool env {add,use,list,remove}` | Manage named environment profiles |
| `cantool status` | Show Canton node health, version, and connected parties |
| `cantool doctor` | Check prerequisites (Java, SDK, Docker, ports, Go) |
| `cantool clean [--all]` | Remove build artifacts |
| `cantool mcp serve` | Start MCP server for AI assistant integration (stdio) |
| `cantool plugin list` | List installed plugins |

### Global Flags

| Flag | Description |
|------|-------------|
| `--format` | Output format: `human` (default), `json`, `quiet` |
| `--version` | Print version |

## Configuration

Cantool projects use a `cantool.yaml` file:

```yaml
version: "1"

project:
  name: my-app
  sdk-version: "3.4.11"

parties:
  - name: Alice
    display: "Alice Corp"
  - name: Bob
    display: "Bob Trading"

environments:
  local:
    host: localhost
    ledger-port: 5011
    json-api-port: 7575

dev:
  watch-paths:
    - "daml/"
  hot-reload: true
  sandbox-port: 5011

plugins: []
```

## MCP Integration

Configure Cantool as an MCP server for Claude Code, Cursor, or other MCP-aware tools:

```json
{
  "mcpServers": {
    "cantool": {
      "command": "cantool",
      "args": ["mcp", "serve"]
    }
  }
}
```

**Available tools:** `query_contracts`, `list_parties`, `ledger_status`, `allocate_party`, `upload_dar`, `list_packages`

**Available resources:** `canton://contracts`, `canton://parties`, `canton://packages`

## Plugin Development

Cantool plugins are standalone binaries that communicate via JSON-RPC 2.0 over stdin/stdout. Plugins are discovered in `~/.cantool/plugins/` and must respond to `--metadata` with a JSON payload:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "My custom plugin",
  "author": "Your Name",
  "min_host_version": "0.1.0"
}
```

See `internal/plugin/interface.go` for the full plugin interface definition.

## Error Codes

Cantool uses structured error codes by subsystem:

| Range | Subsystem |
|-------|-----------|
| CT1xxx | Configuration |
| CT2xxx | SDK/Tooling |
| CT3xxx | Sandbox/Runtime |
| CT4xxx | Build |
| CT5xxx | Test |
| CT6xxx | Ledger API |
| CT7xxx | MCP Server |
| CT8xxx | Plugin System |
| CT9xxx | Environment Management |

Every error includes a human-readable suggestion for how to fix it.

## License

[Apache 2.0](LICENSE)
