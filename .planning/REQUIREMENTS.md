# Requirements: StorCat

**Defined:** 2026-03-26
**Core Value:** Fast, lightweight directory catalog management — Go/Wails delivers 93% smaller binaries and 5x faster search than the original Electron version, with full feature parity.

## v2.1.0 Requirements

Requirements for CLI Commands milestone. Each maps to roadmap phases.

### CLI Foundation

- [x] **CLIP-01**: User can run `storcat` with no args to launch GUI (existing behavior preserved)
- [x] **CLIP-02**: User can run `storcat <command>` to execute CLI commands from the same binary
- [x] **CLIP-03**: CLI dispatch uses stdlib `flag.FlagSet` (no Cobra dependency)
- [x] **CLIP-04**: CLI commands output errors to stderr and results to stdout
- [x] **CLIP-05**: CLI commands exit with code 0 on success, non-zero on error
- [x] **CLIP-06**: All commands support `--help` / `-h` flag

### CLI Commands

- [x] **CLCM-01**: User can run `storcat create <dir>` with `--title`, `--name`, `--output` flags to create a catalog
- [x] **CLCM-02**: User can run `storcat search <term> <dir>` to search catalogs for a term
- [x] **CLCM-03**: User can run `storcat list <dir>` to list catalogs with metadata
- [x] **CLCM-04**: User can run `storcat show <catalog.json>` to display catalog tree structure
- [x] **CLCM-05**: User can run `storcat open <catalog.json>` to open catalog HTML in default browser
- [x] **CLCM-06**: User can run `storcat version` to print version string

### Output Formatting

- [x] **CLOF-01**: `list` and `search` commands support `--json` flag for machine-readable output
- [x] **CLOF-02**: `show` command supports `--json` flag to output raw catalog JSON
- [x] **CLOF-03**: `create` command supports `--json` flag for structured result output
- [x] **CLOF-04**: `show` command supports `--depth N` flag to limit tree depth
- [x] **CLOF-05**: `show` command displays colorized tree output (directories bold/blue) on TTY
- [x] **CLOF-06**: All commands respect `--no-color` flag and `NO_COLOR` env var

### Platform Compatibility

- [x] **CLPC-01**: CLI output works in Windows terminals (console attachment for GUI subsystem binary)
- [x] **CLPC-02**: macOS Gatekeeper `-psn_*` argument injection is filtered before CLI dispatch
- [x] **CLPC-03**: `wails dev` hot-reload still works after CLI dispatch changes
- [x] **CLPC-04**: macOS install script creates `/usr/local/bin/storcat` symlink to `.app` bundle binary
- [x] **CLPC-05**: `storcat open` works cross-platform (macOS `open`, Linux `xdg-open`, Windows `start`)

## Future Requirements

Deferred to v2.1.x or later. Tracked but not in current roadmap.

### CLI Polish

- **CLPL-01**: Auto-TTY detection (table on terminal, JSON when piped) for `list` and `search`
- **CLPL-02**: Shell completion scripts (bash/zsh/fish/powershell) — requires Cobra or manual scripts

### CLI Advanced

- **CLAD-01**: Watch mode (`--watch` on `create`) for auto-recatalog on directory changes
- **CLAD-02**: Interactive TUI mode for catalog browsing

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Separate `storcat-cli` binary | Unified binary is the project goal |
| YAML output format | `--json \| yq` covers the use case |
| stdin piping for search | Ambiguous semantics; explicit positional args are clearer |
| Progress bar on create | TTY-only, breaks pipes, Go traversal is fast enough |
| Config file for CLI defaults (`~/.storcatrc`) | Hidden state; shell aliases suffice at this scale |
| Cobra CLI framework | 6 subcommands don't justify ~2MB binary size increase |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CLIP-01 | Phase 8 | Complete |
| CLIP-02 | Phase 8 | Complete |
| CLIP-03 | Phase 8 | Complete |
| CLIP-04 | Phase 8 | Complete |
| CLIP-05 | Phase 8 | Complete |
| CLIP-06 | Phase 8 | Complete |
| CLCM-01 | Phase 9 | Complete |
| CLCM-02 | Phase 9 | Complete |
| CLCM-03 | Phase 9 | Complete |
| CLCM-04 | Phase 10 | Complete |
| CLCM-05 | Phase 10 | Complete |
| CLCM-06 | Phase 8 | Complete |
| CLOF-01 | Phase 9 | Complete |
| CLOF-02 | Phase 10 | Complete |
| CLOF-03 | Phase 9 | Complete |
| CLOF-04 | Phase 10 | Complete |
| CLOF-05 | Phase 10 | Complete |
| CLOF-06 | Phase 10 | Complete |
| CLPC-01 | Phase 8 | Complete |
| CLPC-02 | Phase 8 | Complete |
| CLPC-03 | Phase 8 | Complete |
| CLPC-04 | Phase 8 | Complete |
| CLPC-05 | Phase 10 | Complete |

**Coverage:**
- v2.1.0 requirements: 23 total
- Mapped to phases: 23
- Unmapped: 0 — complete coverage

---
*Requirements defined: 2026-03-26*
*Last updated: 2026-03-26 after roadmap creation — traceability complete*
