# Repository Instructions

## Build and test commands

- Build the CLI: `go build ./cmd/whichport`
- Run the full test suite: `go test ./...`
- Run a single package: `go test ./internal/ports`
- Run a single test: `go test ./internal/ports -run TestParsePSMetadata`
- Another targeted test example: `go test ./internal/cli -run TestTruncateMiddle`

## High-level architecture

- `cmd/whichport/main.go` is intentionally thin. It only creates the Cobra command from `internal/cli` and exits on error.
- `internal/cli/root.go` owns the user-facing command shape: flags, JSON vs table output selection, and the handoff into discovery via `ports.Discover(...)`.
- `internal/cli/render.go` is the human-output layer. It uses Lip Gloss for styling, calculates column widths itself, and truncates long path/command-line fields in the middle rather than wrapping them.
- `internal/ports/discover.go` defines the shared domain model (`Listener`, `Query`, protocol handling) and sorts the final result set before it goes back to the CLI.
- `internal/ports/provider_darwin.go` is the macOS collector. Discovery is split into two phases:
  - collect listening sockets from `lsof` (`port`, `pid`, short command name)
  - enrich the unique PIDs with executable paths from `lsof -d txt` and full command lines from `ps -ww`
- `internal/ports/provider_linux.go` is the Linux collector. It uses `ss` to discover listening sockets and `/proc/<pid>` to enrich them with executable paths and full command lines.
- `internal/ports/parse.go` holds the string-parsing helpers for `lsof`, `ss`, `ps`, and `/proc`-derived metadata. Keep parsing logic here so it stays testable outside of command execution.
- `internal/ports/provider_unsupported.go` is the fallback stub for platforms other than macOS and Linux. OS-specific behavior should stay behind build tags rather than leaking conditional logic into shared files.

## Key conventions

- Keep the Cobra entrypoint thin; new behavior should usually land in `internal/ports` or `internal/cli/render.go`, not in `main.go`.
- Treat `internal/ports` as the boundary between shelling out to system tools and the rest of the app. Shared logic belongs in untagged files; platform-specific command execution belongs in build-tagged providers.
- When adding new process metadata, prefer batching by PID instead of per-listener shell calls. The current design intentionally deduplicates PIDs before enrichment.
- Preserve the current platform split: macOS uses `lsof`/`ps`, while Linux uses `ss` and `/proc`. Avoid forcing one platform's command shape onto the other.
- Preserve the distinction between:
  - `Command`: short executable/process name from the platform's socket collector (`lsof` on macOS, `ss` on Linux)
  - `CommandLine`: full invocation from process metadata (`ps` on macOS, `/proc/<pid>/cmdline` on Linux)
  - `Path`: executable path from platform-native metadata (`lsof -d txt` on macOS, `/proc/<pid>/exe` on Linux), with platform-appropriate fallbacks only when needed
- Renderer changes should preserve the current output style: width-aware columns, middle truncation for long values, and a plain JSON mode that reuses the same `Listener` struct.
- Tests in this repo are fixture-style parser/render tests. When changing command parsing or output shaping, add or update focused tests in `internal/ports/parse_test.go` or `internal/cli/render_test.go` instead of relying only on manual CLI runs.
