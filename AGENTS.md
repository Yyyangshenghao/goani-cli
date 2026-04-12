# Repository Guidelines

## Project Snapshot
- Stack: Go 1.22+, Bubble Tea TUI, multi-platform CLI player launcher.
- Main entrypoint: `cmd/goani`.
- Debug-only helpers live under `cmd/goani-debug-*` and are not automated tests.
- Persistent user config is stored in `config.json`; source cache is stored separately in `sources_cache.json`.

## Important Directories
- `cmd/goani`: CLI program entrypoint.
- `internal/cli/commands`: command registration and thin command handlers.
- `internal/workflow`: cross-layer flows for search, playback, and config TUI.
- `internal/source`: source models, subscriptions, grouping, cache, and fetching.
- `internal/player`: player discovery, config, and HLS proxy compatibility layer.
- `internal/ui/tui`: Bubble Tea pages and selectors.
- `internal/ui/console`: classic CLI interaction helpers.
- `docs/`: user and maintainer docs. Treat `docs/dev/todo.md` as backlog, not current behavior spec.

## Working Rules
- Read code and command help before changing docs or behavior.
- Keep command handlers thin; move multi-step flow into `internal/workflow`.
- Prefer adapting existing TUI / console patterns over adding parallel interaction styles.
- Do not treat `cmd/goani-debug-*` as a substitute for `*_test.go`.
- Keep changes focused. Avoid renames, file moves, or new dependencies unless the task needs them.
- Repo-local build artifacts should live under `bin/`. Do not leave `goani.exe` or other build outputs in the repository root.
- User handoff builds should default to `bin/goani-test.exe` so the shared PowerShell workflow always points at one stable test binary.

## Validation
- Preferred automated check: `go test ./...`
- Preferred static check: `go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 run`
- Fast fallback static check: `go vet ./...`
- Preferred build check for fast local verification: `go build -o $env:TEMP\goani-check.exe .\cmd\goani`
- Preferred build target for user handoff / manual smoke testing: `go build -o .\bin\goani-test.exe .\cmd\goani`
- Root help can be inspected with: `go run .\cmd\goani --help`
- Command-specific help can be inspected with:
  - `go run .\cmd\goani config --help`
  - `go run .\cmd\goani source --help`
- `make build`, `make build-all`, `make fmt`, and `make lint` exist for GNU Make + Bash environments.
- `make test` should match `go test ./...`. If it drifts again, prefer the direct Go command.
- `make lint-fix` applies formatter-safe fixes via `golangci-lint --fix`.
- CI lives in `.github/workflows/ci.yml` and runs `golangci-lint`, `go test ./...`, and a CLI build on push / pull request.

## Done Means
- Code builds for the changed path.
- Relevant tests pass, or any skipped verification is stated explicitly.
- Docs are updated when command surface, workflow, or setup behavior changes.
- User-facing behavior claims are verified against code or command output, not only against old docs.
