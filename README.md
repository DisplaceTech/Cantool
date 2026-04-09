# Cantool

[![CI](https://github.com/DisplaceTech/Cantool/actions/workflows/ci.yaml/badge.svg)](https://github.com/DisplaceTech/Cantool/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/displacetech/cantool)](https://goreportcard.com/report/github.com/displacetech/cantool)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

Canton application development CLI — project scaffolding, DAML compilation, testing, local sandbox management, and MCP server for AI-assisted development.

## Why Cantool

- **Single binary, zero runtime dependencies for core commands.** Download one file, `chmod +x`, run. The core CLI (scaffolding, MCP server, environments, plugins) requires no JVM, Node.js, or container runtime.
- **Cross-compilation.** `GOOS=linux GOARCH=amd64 go build` produces a Linux binary from macOS. CI builds for all platforms trivially.
- **Static linking.** No shared library conflicts, no version mismatches, no container image bloat. Copy the binary into an air-gapped environment and it works.
- **Plugin system.** Extensible via JSON-RPC plugins. Core commands prove the plugin contract before any external plugins ship.

## Installation

```bash
# Homebrew (macOS)
brew install displacetech/tap/cantool

# Binary download (preferred for institutions)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH="amd64"
curl -sSL "https://github.com/DisplaceTech/Cantool/releases/latest/download/cantool_${OS}_${ARCH}.tar.gz" \
  | tar xz cantool
sudo mv cantool /usr/local/bin/

# Go install (for Go developers)
go install github.com/displacetech/cantool@latest
```

## Quick Start

```bash
# Create a new project
cantool init my-app --template basic
cd my-app

# Enable convenience commands (build, test, dev, clean, doctor)
# Either set globally once: see Configuration > Global Config
# Or per-project in cantool.yaml: plugins.convenience.enabled: true

# Check prerequisites
cantool doctor

# Build and test
cantool build
cantool test

# Start local sandbox with hot-reload
cantool dev
```

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `cantool init [name]` | Create a new Canton project from a template (`basic` or `token`) |
| `cantool env {add,use,list,remove}` | Manage named environment profiles |
| `cantool status` | Show Canton node health, version, and connected parties |
| `cantool mcp serve` | Start MCP server for AI assistant integration (stdio) |
| `cantool plugin list` | List installed plugins |

### Optional Convenience Commands (plugin)

These commands are provided by the built-in `convenience` plugin and must be enabled in `cantool.yaml` (see [Plugins](#plugins)). They delegate to dpm/daml and print an attribution line to stderr (e.g., `-> delegating to dpm build`).

| Command | Description |
|---------|-------------|
| `cantool build [--watch]` | Compile DAML sources. `--watch` rebuilds on file changes |
| `cantool test [--verbose]` | Run DAML Script tests with structured output |
| `cantool dev [--port P]` | Start local sandbox with hot-reload and party provisioning |
| `cantool clean [--all]` | Remove build artifacts |
| `cantool doctor` | Check prerequisites (Java, SDK, Docker, ports, Go) |

### Global Flags

| Flag | Description |
|------|-------------|
| `--format` | Output format: `human` (default), `json`, `quiet` |
| `--version` | Print version |

## Configuration

Cantool supports two levels of configuration:

1. **Global config** — checked in order: `$XDG_CONFIG_HOME/cantool/config.yaml` (default `~/.config/cantool/`), `~/Library/Application Support/cantool/config.yaml` (macOS only), `~/.cantool/config.yaml` (fallback). Applied to all projects. Useful for enabling plugins once across all projects.
2. **Project config** — `cantool.yaml` in the project root. Project settings override global settings.

### Global Config

Create a global config to set defaults for all projects:

```bash
# Linux
mkdir -p ~/.config/cantool
cat > ~/.config/cantool/config.yaml << 'EOF'
version: "1"
plugins:
  convenience:
    enabled: true
EOF

# macOS
mkdir -p ~/Library/Application\ Support/cantool
cat > ~/Library/Application\ Support/cantool/config.yaml << 'EOF'
version: "1"
plugins:
  convenience:
    enabled: true
EOF
```

This enables convenience commands everywhere without editing each project's `cantool.yaml`. A project can still override the global setting by explicitly setting `plugins.convenience.enabled: false` in its own `cantool.yaml`.

### Project Config

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

plugins:
  convenience:
    enabled: true
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

## Plugins

Cantool supports built-in and external plugins. Plugins are configured in the `plugins` section of `cantool.yaml` or the global config file (see [Configuration](#configuration)).

### Built-in: Convenience Plugin

The `convenience` plugin ships with Cantool and provides wrapper commands for common dpm/daml operations. It is **disabled by default**.

To enable it globally (all projects):

```bash
# Linux
mkdir -p ~/.config/cantool
cat > ~/.config/cantool/config.yaml << 'EOF'
version: "1"
plugins:
  convenience:
    enabled: true
EOF

# macOS
mkdir -p ~/Library/Application\ Support/cantool
cat > ~/Library/Application\ Support/cantool/config.yaml << 'EOF'
version: "1"
plugins:
  convenience:
    enabled: true
EOF
```

Or enable it per-project in `cantool.yaml`:

```yaml
plugins:
  convenience:
    enabled: true
```

To disable it in a specific project (overriding a global enable):

```yaml
plugins:
  convenience:
    enabled: false
```

When enabled, the following commands become available:

| Command | Delegates to |
|---------|-------------|
| `cantool build` | `dpm build` / `daml build` |
| `cantool test` | `dpm test` / `daml test` |
| `cantool dev` | `dpm sandbox` / `daml sandbox` |
| `cantool clean` | Removes `.daml/` build artifacts |
| `cantool doctor` | Checks for Java, SDK, Docker, ports, Go |

When a convenience command delegates to an external tool, it prints a single attribution line to stderr (e.g., `-> delegating to dpm build`).

Use `cantool plugin list` to see installed plugins.

### External Plugins

External plugins are standalone binaries that communicate via JSON-RPC 2.0 over stdin/stdout. Plugins are discovered in `~/.cantool/plugins/` and must respond to `--metadata` with a JSON payload:

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
