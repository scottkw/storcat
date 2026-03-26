# StorCat v2.1.0 Release Notes

**Release Date:** March 26, 2026
**Build Type:** Feature Release — CLI Commands

## Summary

StorCat v2.1.0 adds a full command-line interface to the unified binary. Run `storcat` with no arguments for the GUI, or use CLI subcommands for scripting and terminal workflows. The same binary serves both modes — no separate CLI install needed.

## What's New

### CLI Subcommands

| Command | Description |
|---------|-------------|
| `storcat create <dir>` | Create a catalog from a directory |
| `storcat search <term>` | Search catalogs for a filename pattern |
| `storcat list [dir]` | List catalogs with metadata |
| `storcat show <catalog>` | Display a catalog's tree structure |
| `storcat open <catalog>` | Open a catalog's HTML in the default browser |
| `storcat version` | Print the version |

### CLI Features
- `--json` flag for machine-readable output on create, search, list, and show
- `--depth N` flag to limit tree depth on show
- Colorized tree output with `--no-color` flag and `NO_COLOR` environment variable support
- `--title` and `--name` flags for create
- `--output` flag to specify output directory for create
- Cross-platform browser launch for open (macOS, Windows, Linux)
- Standard exit codes: 0 = success, 1 = runtime error, 2 = usage error
- Per-command `--help` flags

### Platform Integration
- macOS `-psn_*` Gatekeeper argument filtering (silently stripped before CLI dispatch)
- `scripts/install-cli.sh` creates `/usr/local/bin/storcat` symlink for macOS CLI access
- Windows console output support

### Architecture
- CLI dispatch in `main.go` — known subcommands route to CLI, no args falls through to GUI
- `cli/` package with per-command files (create.go, search.go, list.go, show.go, open.go, version.go)
- Shared `output.go` helpers for consistent formatting
- Full test coverage across all subcommands

### Tech Debt Cleanup
- Fixed help output to use stdout (was stderr)
- Updated version test to match embedded wails.json
- Removed orphaned `FormatBytes` utility
- Added `NO_COLOR` environment variable test

## Build Artifacts

Same platforms as v2.0.0 — the CLI is built into the existing binary:

- **macOS**: `StorCat.app` — Universal binary (Intel + Apple Silicon)
- **Windows**: `StorCat.exe` — x64 and ARM64
- **Linux**: `StorCat-linux-amd64` / `StorCat-linux-arm64`

## New Dependencies

- **fatih/color** v1.18.0 — Colorized CLI output
- **tablewriter** v1.1.4 — Tabular CLI output formatting
- **pkg/browser** — Cross-platform browser launch

## Upgrade from v2.0.0

Drop-in replacement. No configuration changes needed. The GUI is unchanged — the CLI is additive.

## Support

- **GitHub Issues**: [StorCat Issues](https://github.com/scottkw/storcat/issues)
- **Developer**: Ken Scott

---

**StorCat v2.1.0** — Fast, Native, Cross-Platform Storage Media Cataloging (Desktop + CLI)
