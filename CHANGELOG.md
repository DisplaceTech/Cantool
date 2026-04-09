# Changelog

All notable changes to Cantool are documented in this file.

## [0.1.1] - Unreleased

### Changed

- Convenience commands (`build`, `test`, `clean`, `dev`, `doctor`) extracted behind
  the built-in `convenience` plugin. These commands are disabled by default and can
  be enabled by setting `plugins.convenience.enabled: true` in `cantool.yaml`.
  No functionality changes when enabled.
- When convenience commands delegate to an external tool (dpm/daml), a single
  attribution line is printed to stderr (e.g., `-> delegating to dpm build`).
- `plugins` config field changed from a string list to a structured map.

### Fixed

- Binary install command in README now matches actual GoReleaser archive naming.
- Clarified dependency messaging: core CLI is zero-dependency; convenience commands
  delegate to dpm and require its prerequisites (JDK 17+, etc.).

## [0.1.0] - 2025-06-01

Initial release.
