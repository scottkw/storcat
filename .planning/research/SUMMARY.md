# Project Research Summary

**Project:** StorCat v2.1.0 — CLI Subcommands
**Domain:** Adding CLI subcommands to a unified Go/Wails desktop binary
**Researched:** 2026-03-26
**Confidence:** HIGH

## Executive Summary

StorCat v2.1.0 adds a CLI layer to an existing, shipping Go/Wails desktop application. This is not a new product — all business logic (catalog creation, search, browse) already exists in `internal/catalog` and `internal/search` packages that have no Wails runtime dependencies. The CLI work is thin command wrappers over those services, plus a dispatch mechanism in `main.go` that routes known subcommands away from `wails.Run()` before the GUI ever starts. The pattern is well-understood: check `os.Args` before calling `wails.Run()`, route to a `cli.Run()` function that uses stdlib `flag.FlagSet` per subcommand, and `os.Exit()` before the Wails framework initializes. Total implementation is estimated at 2-3 hours of focused Go work.

The recommended approach avoids the two most tempting over-engineering traps: don't add Cobra (6 subcommands with simple flags don't justify a ~2MB dependency in a binary where size is a selling point), and don't use the `App` struct from CLI mode (it requires a live Wails context and panics without one). Instead, instantiate `catalog.NewService()` and `search.NewService()` directly in CLI commands — exactly what `NewApp()` does, minus the Wails binding layer. A new `cli/` package encapsulates all subcommands and exposes a single `Run(args []string, version string) int` entry point.

The primary risks are platform-specific and must be front-loaded: Windows GUI subsystem silencing all CLI output (requires `-windowsconsole` build flag or `AttachConsole` syscall), macOS Gatekeeper injecting `-psn_*` arguments on first Finder launch, and `wails dev` double-execution breaking if the dispatch logic is too strict. All three are well-documented patterns with known fixes that must be addressed in the first implementation phase before any per-subcommand work begins. The macOS `.app` bundle PATH issue is a distribution concern: users who install by dragging to `/Applications` won't have `storcat` on their PATH without a symlink, which requires a documented install script shipped with the release.

## Key Findings

### Recommended Stack

The existing stack (Go 1.23, Wails v2.10.2, React 18, TypeScript 5, Ant Design 5) is unchanged. The v2.1.0 additions are two targeted Go libraries: `github.com/olekukonko/tablewriter v1.1.4` for terminal table output in `list` and `search` commands, and `github.com/fatih/color v1.19.0` for optional colorized tree output in `show`. Both are lightweight and well-established. `github.com/pkg/browser` (already a Wails transitive dependency at zero new cost) handles cross-platform browser opening for `storcat open`. The architecture research explicitly recommends against adding Cobra — stdlib `flag.FlagSet` is sufficient for 6 subcommands and keeps binary size lean.

**Core technologies:**
- `stdlib flag.FlagSet`: per-subcommand flag parsing — no external dependency, adequate for 6 well-defined subcommands with simple flag sets
- `github.com/olekukonko/tablewriter v1.1.4`: ASCII table output for `list` and `search` — established library, pipe-safe; note v1.0.0 is broken, use v1.1.4
- `github.com/fatih/color v1.19.0`: ANSI color for `show` tree — TTY auto-detection via `go-isatty` already in go.mod via Wails, zero net dependency cost
- `github.com/pkg/browser`: cross-platform browser open for `storcat open` — already in go.sum as a Wails transitive dep, zero new dependency

### Expected Features

**Must have (table stakes) — v2.1.0 launch:**
- `storcat create <dir>` with `--output`, `--name`, `--title` flags — core catalog creation via existing service
- `storcat search <term> <dir>` with `--json` flag — scriptable search
- `storcat list <dir>` with `--json` flag — scriptable catalog discovery
- `storcat show <catalog.json>` with `--depth N` flag — terminal tree visualization (requires new `GenerateTextTree` method on catalog.Service)
- `storcat open <catalog.json>` — cross-platform browser launch via `pkg/browser`
- `storcat version` — version string to stdout
- `storcat` (no args) — GUI launches unchanged; preserving this behavior is a hard requirement, not a nice-to-have
- Exit codes 0/1/2, errors to stderr, output to stdout — POSIX contract required for scripting
- `--help` on all commands — stdlib `flag` provides this automatically

**Should have (differentiators) — add after v2.1.0 ships:**
- Auto-TTY detection: table output on terminal, JSON when piped — matches `gh` CLI behavior, defers need to remember `--json`
- Colorized `show` tree output — directories bold/blue, matches `tree`/`eza` UX
- `--no-color` + `NO_COLOR` env var — only needed once colorized output ships

**Defer (v2.2+):**
- Watch mode (`--watch` on `create`) — requires `fsnotify`, daemon concerns
- Interactive TUI mode — pipe `storcat list` or `storcat search` to `fzf` instead
- YAML output — `--json | yq` covers the use case
- Shell completion via Cobra — only justified if subcommand count grows significantly

### Architecture Approach

The integration is a two-layer change to the existing codebase. First, `main.go` gains an `os.Args` dispatch block before `wails.Run()` — known subcommands route to `cli.Run(os.Args[1:], Version)` and exit; no-arg or unrecognized-arg invocations fall through to GUI launch unchanged. Second, a new `cli/` package provides one file per subcommand, each constructing services directly via `catalog.NewService()` or `search.NewService()`. All existing `internal/` packages and `app.go` remain unchanged except for one new exported method: `catalog.Service.GenerateTextTree()` for the `show` command. The version string is passed as a parameter to `cli.Run()` to keep the `//go:embed wails.json` directive in the `main` package where it belongs.

**Major components:**
1. `main.go` (modified) — adds `-psn_*` filter, `os.Args` dispatch, extracts GUI init to `runGUI()` function
2. `cli/` package (new) — `cli.go` root dispatcher + one file per subcommand (create, search, list, show, open, version) + `cli_windows.go` console attachment shim
3. `internal/catalog/service.go` (minor change) — adds exported `GenerateTextTree(item *models.CatalogItem) string` method for plain-text tree rendering; all other services unchanged
4. `pkg/cli/output.go` (optional) — shared table/tree/byte formatting helpers extracted only when real duplication emerges across commands

**Hard isolation rules that must not change:**
- CLI commands must call `internal/catalog` and `internal/search` directly — never through `App`
- `cli/` package must have zero imports from `wails/v2/pkg/runtime` — Wails runtime requires a live context
- `App` struct is the Wails binding layer only; it is not a general CLI service facade

### Critical Pitfalls

1. **Windows GUI subsystem silences all CLI output** — `wails build` uses `-H windowsgui` which discards stdout/stderr from Windows terminals. Fix: build with `wails build -windowsconsole` (accepted trade-off: brief console flash on GUI double-click) or use `AttachConsole(-1)` + `CONOUT$` redirect in `cli_windows.go`. Must be validated on real Windows before any CLI feature is considered done.

2. **macOS Gatekeeper injects `-psn_*` argument on first Finder launch** — Any arg parser encountering `-psn_0_8423432` will error or misfire a CLI command. Fix: filter args matching `-psn_*` prefix before dispatch. One function, apply on all platforms (no-op on non-macOS), add a test.

3. **`wails dev` double-executes `main()` during binding generation** — Strict "no args → exit 1" logic will break the dev hot-reload workflow. Fix: unknown args (or no args) must always fall through to GUI launch, never exit with an error. Reserve strict validation for inside subcommand handlers, not the dispatch layer.

4. **`embed.FS` compile failure on direct `go build`** — `//go:embed all:frontend/dist` requires pre-built frontend assets. Direct `go build` fails on fresh clones. Fix: document that `wails build` is the only supported build path; never use `go build` for the full binary.

5. **macOS `.app` bundle binary not on PATH** — Users who drag to `/Applications` cannot run `storcat create` from terminal without a symlink. Fix: provide `scripts/install-cli.sh` that creates `/usr/local/bin/storcat → storcat.app/Contents/MacOS/storcat`, document in release notes.

## Implications for Roadmap

Based on the research, the implementation is best structured in three phases. All platform gotchas must be resolved before per-subcommand work begins — the dispatch skeleton is the critical path dependency for every subsequent command.

### Phase 1: CLI Foundation and Platform Compatibility
**Rationale:** The three critical pitfalls (Windows console, macOS PSN, wails dev double-exec) all manifest at the dispatch layer. Building the CLI skeleton first lets you verify the integration is sound on all platforms before writing any command logic. This is the true critical path dependency — every subcommand shares the same dispatch mechanism.
**Delivers:** A working `storcat version` and `storcat --help` on all three platforms (macOS, Windows, Linux). `wails dev` still opens the GUI. macOS Finder launch does not crash. Windows terminal shows output.
**Addresses:** Table-stakes features: `storcat version`, exit codes, stderr/stdout split, `--help`; POSIX contract established
**Avoids:** Windows console silent failure (Pitfall 1), macOS PSN crash (Pitfall 2), wails dev regression (Pitfall 3), embed.FS documentation gap (Pitfall 4)
**Specific work:**
- Modify `main.go`: add `-psn_*` filter, extract `runGUI()`, add `os.Args` dispatch switch, pass `Version` to `cli.Run()`
- Create `cli/cli.go` with `Run(args []string, version string) int` entry point, help text, and subcommand routing
- Create `cli/version.go` — trivial, one line of output
- Create `cli/cli_windows.go` with `AttachConsole` shim
- Decide and document Windows build approach (`-windowsconsole` flag recommended)
- Create `scripts/install-cli.sh` for macOS symlink
- Document that `wails build` is required, not `go build`
- Verify `wails dev` still opens GUI after dispatch changes

### Phase 2: Core Subcommands (Create, List, Search)
**Rationale:** These three commands cover the primary scripting use cases and all call existing services with no new business logic — just flag parsing and output formatting. `list` and `search` share the `search.Service` constructor pattern; all three share the table output approach. Implementing them together is efficient.
**Delivers:** A fully scriptable CLI for the core catalog workflow: create a catalog, list what's cataloged, search for files. Human-readable table output on TTY, JSON output via `--json`.
**Uses:** `olekukonko/tablewriter` for tabular output, `encoding/json` for `--json` flag (existing model structs already have `json:` tags), `stdlib flag.FlagSet` for parsing
**Implements:** `cli/create.go`, `cli/list.go`, `cli/search.go`
**Avoids:** App struct / Wails ctx misuse (instantiate services directly), Windows path separator inconsistency in output

### Phase 3: Show and Open Subcommands
**Rationale:** These two commands have distinct implementation requirements. `show` requires a new exported method on `catalog.Service` (the text tree renderer, since the existing `generateTreeStructure` is HTML-specific with `&nbsp;` and `html.EscapeString`). `open` requires cross-platform browser exec logic. Grouping them keeps the one service-layer change isolated to a single phase.
**Delivers:** Complete CLI feature parity — users can inspect a catalog tree structure and open its HTML report without launching the GUI.
**Implements:** `catalog.Service.GenerateTextTree()` (new exported method), `cli/show.go`, `cli/open.go`
**Avoids:** Using `runtime.BrowserOpenURL` from Wails in CLI context (requires active Wails ctx, will panic) — use `pkg/browser` instead

### Phase Ordering Rationale

- Phase 1 must precede all others. The `main.go` dispatch is a prerequisite for any CLI command to be invokable, and platform validation cannot be deferred — discovering Windows console failure after all commands are implemented is a much more expensive fix than discovering it on a one-liner `version` command.
- Phases 2 and 3 are ordered by implementation complexity and service-layer change scope, not by relative user value (all commands are P1 features). Phase 2 commands are pure call-through with no new service methods; Phase 3 requires one service method addition.
- Shell completion (`storcat completion`) is free only with Cobra. Staying with stdlib flag means deferring completion to a later milestone. This is the right trade-off at this scope — document the decision explicitly in Phase 1.

### Research Flags

Phases likely needing deeper investigation during planning:
- **Phase 1 (Windows console):** The `-windowsconsole` trade-off (console flash on GUI launch) vs. `AttachConsole` (stdin unreliability, piped output limitations) needs a Windows environment decision before committing. Recommend testing both on a real Windows build before choosing.
- **Phase 3 (cross-platform browser exec):** `pkg/browser` handles macOS and Linux cleanly; Windows `cmd /c start` with paths containing spaces has known edge cases that need explicit testing.

Phases with standard patterns (skip additional research):
- **Phase 2:** `catalog.NewService()` and `search.NewService()` are already proven in the GUI context. Flag parsing and table output are standard Go patterns. JSON output is `json.Marshal` on existing structs with `json:` tags — nearly free.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Existing stack validated and unchanged. Library choices (tablewriter, fatih/color) are well-established with high adoption. The Cobra vs stdlib flag recommendation is well-reasoned and reversible if scope grows. |
| Features | HIGH | All business logic exists in proven services. CLI conventions are well-established via CLIG and reference tools (tree, fd, eza, rg). Feature set is tightly scoped. |
| Architecture | HIGH | Based on direct source code analysis confirming service layer has zero Wails runtime imports. Dispatch pattern is confirmed via Wails community discussions. |
| Pitfalls | MEDIUM-HIGH | Windows and macOS gotchas documented from multiple Wails community sources and platform docs. Some platform edge cases (piped output on Windows after AttachConsole, multi-monitor behavior) may surface during testing. |

**Overall confidence:** HIGH

### Gaps to Address

- **Windows `-windowsconsole` vs `AttachConsole` choice:** Both approaches are documented with known trade-offs, but the right choice for StorCat's user base needs a real Windows environment test in Phase 1 before the rest of the implementation is built around it.
- **Shell completion without Cobra:** `storcat completion` is listed as P1 in FEATURES.md but is only free with Cobra. This is a scope decision: accept the loss of completion for v2.1.0 (document it), adopt Cobra for completion support, or implement manual completion scripts. Must be decided explicitly in Phase 1 before implementation begins.
- **`storcat list` and `storcat search` default directory behavior:** When no `<dir>` argument is provided, what is the default? Current directory? Last-used directory from config? Research recommends reading from `config.Manager` (same defaults as GUI) rather than hardcoding, but the exact UX decision needs confirmation during Phase 2 planning.

## Sources

### Primary (HIGH confidence)
- Direct source analysis: `main.go`, `app.go`, `internal/catalog/service.go`, `internal/search/service.go`, `pkg/models/catalog.go` — confirmed no Wails runtime imports in service layer
- `github.com/olekukonko/tablewriter` pkg.go.dev (v1.1.4) — table output library
- `github.com/fatih/color` pkg.go.dev (v1.19.0) — color output library
- [Command Line Interface Guidelines (clig.dev)](https://clig.dev/) — exit codes, stdout/stderr separation, `--json` flag, piping conventions, color/TTY detection
- Wails v2 build system reference: https://wails.io/docs/reference/cli/

### Secondary (MEDIUM confidence)
- Wails issue #544 — Windows console output limitation with `-H windowsgui` linker flag
- Wails issue #3008 — `-windowsconsole` debug logging build flag
- Wails discussion #4175 — `os.Args` behavior in production builds and `wails dev` double-execution
- Wails issue #1533 — `-appargs` flag passthrough bug in dev mode
- Wails discussion #3098 — CLI integration patterns with Wails apps
- macOS `-psn_*` Gatekeeper injection documented via Godot engine fix: https://github.com/godotengine/godot/pull/37719
- Windows `AttachConsole` technique: https://www.tillett.info/2013/05/13/how-to-create-a-windows-program-that-works-as-both-as-a-gui-and-console-application/

### Tertiary (reference / needs validation)
- `wails build -windowsconsole` flag — referenced in issue #3008 but not in official docs; validate on real Windows build before committing
- macOS app bundle CLI access: https://www.jviotti.com/2022/11/28/launching-macos-applications-from-the-command-line.html
- Cobra alternatives analysis: https://www.glukhov.org/post/2025/11/go-cli-applications-with-cobra-and-viper/

---
*Research completed: 2026-03-26*
*Ready for roadmap: yes*
