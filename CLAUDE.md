# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
just build          # Build the binary to ./clippy
just run            # Build and run the application
just test           # Run all tests
just test-verbose   # Run tests with verbose output
just test-coverage  # Run tests with coverage report (outputs coverage.html)
just fmt            # Format all Go code
just lint           # Run golangci-lint (requires: brew install golangci-lint)
just clean          # Remove build artifacts
just demo           # Run the demo application
```

To run a single test package:
```bash
go test ./internal/history/...
go test ./internal/search/...
go test ./internal/ui/...
go test ./internal/db/...
```

## Architecture

Clippy is a terminal-based clipboard history manager. The data flow is:

1. **Clipboard polling** — `ui.Tick()` fires every 2 seconds, the `Model.Update()` handler reads the system clipboard via `atotto/clipboard` and calls `history.Manager.AddItem()`
2. **Persistence** — `internal/db` wraps an Automerge CRDT document (`~/.clippy/clippy.automerge`) using `automerge-go` (requires CGO). Items are stored with SHA-256 hash, content, timestamp, and pinned state. Pinned items sort to the top; ties broken by timestamp ascending.
3. **Deduplication** — `Manager` maintains an in-memory hash set; `AddItem` skips content already seen in this session or in the document.
4. **TUI** — Built with Bubble Tea. `ui.Model` is the top-level Bubble Tea model. It delegates table rendering to `ui/table.Manager` and fuzzy search to `internal/search.FuzzyMatcher`.

### Package layout

- `cmd/clippy/` — Entry point: wires `history.Manager` → `ui.Model` → `tea.Program`
- `internal/db/` — Automerge client (`db.AutomergeClient`): `Insert`, `Delete`, `LoadAll`, `SetPinned`, `Close`; also defines the `DBClient` interface and `ClipboardEntry` type
- `internal/history/` — `Manager`: in-memory item list + hash set backed by `db.DBClient`; `ClipboardHistory` type (Item, Hash, TimeStamp, Pinned)
- `internal/search/` — `FuzzyMatcher`: fzf-style scoring (consecutive match bonus, word boundary bonus, camelCase bonus)
- `internal/ui/` — Bubble Tea model with two view modes: `TableView` and `SearchView`
  - `ui/table/` — `table.Manager` wraps `charmbracelet/bubbles/table`; stable cursor selection by hash on every `UpdateRows` call
  - `ui/styles/` — Lipgloss themes (`Theme`, `TableTheme`)

### CGO requirement

`automerge-go` wraps a C library and requires CGO. Build with `CGO_ENABLED=1` (the default). Releases are cross-compiled from a single Linux runner using `zig cc` as the C compiler.

### Testing patterns

Tests use `history.NewManagerWithPath(dbPath)` with a temp directory for isolation — see `internal/ui/test_helpers.go` for the shared `setupTestHistoryManager` helper used across UI tests.
