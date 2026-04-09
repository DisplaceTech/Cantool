# Changelog

All notable changes to Cantool are documented in this file.

## [0.1.3] - Unreleased

### Fixed

- Global config not detected on macOS. Added `~/Library/Application Support/cantool/config.yaml`
  to the config search path via `os.UserConfigDir()`.
- Honor `$XDG_CONFIG_HOME` environment variable per XDG Base Directory spec. Previously
  the XDG path was always hardcoded to `~/.config/`.
- `go install` instructions in README used wrong case (`DisplaceTech/Cantool` vs
  module path `displacetech/cantool`), causing install failures.

## [0.1.2] - 2026-04-09

### Added

- Global config support: `~/.config/cantool/config.yaml` (XDG) or
  `~/.cantool/config.yaml` (fallback). Enable plugins once for all projects.
  Project config overrides global settings.
- `cantool plugin list` now shows the built-in convenience plugin with
  enabled/disabled status.

### Fixed

- `cantool -h` and `cantool help` now show identical command lists.
- `plugin list` refers to README for plugin management instructions.

## [0.1.1] - 2026-04-09

### Changed

- Convenience commands (`build`, `test`, `clean`, `dev`, `doctor`) extracted behind
  the built-in `convenience` plugin. These commands are disabled by default and can
  be enabled by setting `plugins.convenience.enabled: true` in `cantool.yaml`
  or the global config file. No functionality changes when enabled.
- When convenience commands delegate to an external tool (dpm/daml), a single
  attribution line is printed to stderr (e.g., `-> delegating to dpm build`).
- `plugins` config field changed from a string list to a structured map.

### Fixed

- Binary install command in README now matches actual GoReleaser archive naming.
- Clarified dependency messaging: core CLI is zero-dependency; convenience commands
  delegate to dpm and require its prerequisites (JDK 17+, etc.).

## [0.1.0] - 2025-06-01

Initial release.
