# Stack Research

**Domain:** CLI subcommands in a Go/Wails unified desktop binary
**Researched:** 2026-03-26
**Confidence:** HIGH

---

## Context: What Stays, What Changes

This document covers **only the additions needed for v2.1.0 CLI features**. The existing stack (Go 1.23, Wails v2.10.2, React 18, TypeScript 5, Ant Design 5) is validated and unchanged. See the v2.0.0 STACK.md history for those decisions.

The new question: what libraries handle CLI argument parsing, table output, and color in a binary that must also launch a GUI?

---

## Recommended Stack — New Additions Only

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| `github.com/spf13/cobra` | v1.10.2 | CLI subcommand framework | De-facto standard for Go CLIs with subcommands. Used by kubectl, Hugo, GitHub CLI. Provides automatic `--help` generation, POSIX-compliant flags, shell completion, and command hierarchy. stdlib `flag` cannot model subcommands cleanly — it requires manual dispatch and reimplementing help text. For 6 subcommands (`create`, `search`, `list`, `show`, `open`, `version`), cobra is the right tool. |
| `github.com/olekukonko/tablewriter` | v1.1.4 | Terminal table output | Best established Go table library. Renders ASCII/Unicode tables with column alignment, padding, borders. Used by `storcat list` (catalog metadata table). Already used in the ecosystem (Kubernetes, Helm, many CLIs). Do not use `text/tabwriter` — it only handles spacing, not borders or alignment. |
| `github.com/fatih/color` | v1.19.0 | ANSI color output | Simplest, most widely adopted Go color library. 26,000+ dependents. Auto-detects TTY and disables color when piped (via `mattn/go-isatty` and `mattn/go-colorable`, which are **already in go.mod** as Wails indirect dependencies — zero net dependency cost). No unnecessary heavy deps. |

### Supporting Pattern: Main Entry Point

| Concern | Approach | Why |
|---------|---------|-----|
| GUI vs CLI dispatch | Check `len(os.Args) > 1` before `wails.Run()` in `main.go` | Wails v2 runs the binary twice during `wails dev` (once for binding generation without args). The safe pattern: if `os.Args[1]` is a known subcommand, run CLI mode and `os.Exit(0)` before `wails.Run()` is ever called. GUI mode is the zero-args fallback. |
| Build tag concern | No build tag needed | The `os.Args` check happens before `wails.Run()`. The binding generation pass (no args) safely falls through to GUI mode. No `//go:build production` split is required for this use case. |
| CLI mode isolation | CLI commands in `cmd/` package, called directly from `main.go` dispatch | Keeps `main.go` thin. Each subcommand (`cmd/create.go`, `cmd/search.go`, etc.) reuses existing `internal/catalog` and `internal/search` packages unchanged. No new business logic needed. |

---

### Development Tools — No Changes

| Tool | Notes |
|------|-------|
| `wails dev` | Works unchanged. CLI subcommands only execute in the production binary (`wails build` output) or direct `go run`. |
| `wails build` | Build command unchanged. The unified binary contains embedded frontend + CLI dispatch. |

---

## Installation

```bash
# CLI framework
go get github.com/spf13/cobra@v1.10.2

# Table output
go get github.com/olekukonko/tablewriter@v1.1.4

# Color output (mattn/go-isatty and mattn/go-colorable already in go.mod via Wails)
go get github.com/fatih/color@v1.19.0
```

After adding: `go mod tidy`

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| `cobra` | stdlib `flag` | Only for single-command tools with simple flags. For 6+ subcommands with help generation, flag is painful — manual dispatch, no help inheritance, no shell completion. |
| `cobra` | `urfave/cli` | cli/v2 is a reasonable alternative. Cobra is better here because StorCat's subcommand surface is small and well-defined — cobra's command tree model is cleaner for this. |
| `cobra` | `urfave/cli v3` | v3 was released late 2024. Would work, but cobra has more ecosystem examples and is the clear community default. No advantage for this project size. |
| `olekukonko/tablewriter` | `charmbracelet/lipgloss` table | lipgloss is excellent for TUI apps. For a simple catalog listing table it is over-engineered. Adds heavy bubbletea dependency chain. tablewriter is the right complexity level. |
| `olekukonko/tablewriter` | `jedib0t/go-pretty` | go-pretty supports more output formats (TSV, CSV, HTML). Would work, but StorCat only needs terminal text. tablewriter is simpler and more focused. |
| `olekukonko/tablewriter` | `text/tabwriter` (stdlib) | tabwriter only aligns columns via whitespace; no borders, no padding control, no header separator. Too bare for a catalog listing. |
| `fatih/color` | stdlib ANSI codes directly | Fragile. Requires manual TTY detection per-platform (especially Windows). fatih/color handles this correctly using dependencies already in go.mod. |
| `fatih/color` | `charmbracelet/lipgloss` for color | lipgloss is a full layout/style system. Using it just for color in CLI output is overkill for this project. |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `github.com/spf13/viper` | Config file management for CLI flags. StorCat CLI commands are one-shot invocations, not long-running services needing config files. Adds complexity with zero benefit. | Cobra's `Flags()` + `PersistentFlags()` |
| `github.com/charmbracelet/bubbletea` | Interactive TUI framework (spinner, progress bar). CLI commands are fast (catalog creation is sub-second). Progress UI is over-engineered for this scope. | Plain `fmt.Fprintf` for progress hints |
| `github.com/charmbracelet/lipgloss` | Full terminal layout/style system. Correct tool for TUIs, not for simple CLI table output. | tablewriter + fatih/color |
| `github.com/mitchellh/go-homedir` | Home directory resolution. `os.UserHomeDir()` (stdlib, Go 1.12+) is sufficient and simpler. | stdlib `os.UserHomeDir()` |
| Wails v3 migration | v3 is still alpha as of March 2026. The CLI pattern works cleanly on v2. Do not migrate mid-milestone. | Stay on Wails v2.10.2 |

---

## Stack Patterns by Variant

**If the subcommand needs to output JSON (e.g., `storcat list --json`):**
- Use stdlib `encoding/json` + `json.NewEncoder(os.Stdout)`
- No additional library needed — already in use for catalog files

**If the terminal is piped (e.g., `storcat list | grep foo`):**
- `fatih/color` auto-disables color via `go-isatty`
- `tablewriter` renders plain ASCII (no Unicode box chars) by default — pipe-safe
- No special handling needed

**If a subcommand needs to open HTML in the browser (storcat open):**
- Use `github.com/pkg/browser` — already in go.mod as a Wails indirect dependency
- Zero new dependency: `browser.OpenURL(htmlPath)`

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| `cobra v1.10.2` | Go 1.23 | Requires Go 1.22+. v1.10.2 published December 2025. |
| `tablewriter v1.1.4` | Go 1.23 | v1.0.0 is broken — avoid. Use v1.1.4. Published March 2026. |
| `fatih/color v1.19.0` | Go 1.23 | Uses `mattn/go-isatty` and `mattn/go-colorable` already in go.mod via Wails. No version conflict. Published March 2026. |

---

## Integration Notes

### Wails Build Does Not Change

`wails build` compiles the whole `main` package including the CLI dispatch. The resulting binary:
- Zero args → launches GUI (Wails behavior unchanged)
- Known subcommand as first arg → runs CLI, exits before `wails.Run()`

The embedded frontend (`//go:embed all:frontend/dist`) is still bundled but never loaded in CLI mode. Binary size is unchanged — the embedded assets are inert in CLI mode.

### Existing Business Logic Is Reused As-Is

`internal/catalog/service.go` and `internal/search/service.go` contain all the catalog creation and search logic. CLI commands are thin wrappers that call these packages directly — no duplication.

### `storcat version` Is Already Implemented

`version.go` handles version string injection via `//go:embed wails.json`. The `version` subcommand is a one-liner that prints `app.Version`.

---

## Sources

- cobra GitHub (v1.10.2 latest): https://github.com/spf13/cobra — HIGH confidence
- cobra pkg.go.dev: https://pkg.go.dev/github.com/spf13/cobra — HIGH confidence
- tablewriter pkg.go.dev (v1.1.4): https://pkg.go.dev/github.com/olekukonko/tablewriter — HIGH confidence
- fatih/color pkg.go.dev (v1.19.0): https://pkg.go.dev/github.com/fatih/color — HIGH confidence
- Wails discussion #4175 (os.Args in production builds): https://github.com/wailsapp/wails/discussions/4175 — MEDIUM confidence (community discussion, not official docs)
- Wails issue #2353 (CLI args pattern): https://github.com/wailsapp/wails/issues/2353 — MEDIUM confidence
- Existing go.mod (indirect deps: pkg/browser, go-isatty, go-colorable): verified from codebase — HIGH confidence

---
*Stack research for: CLI subcommands in Go/Wails unified binary (StorCat v2.1.0)*
*Researched: 2026-03-26*
