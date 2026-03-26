# Phase 8: CLI Foundation and Platform Compatibility - Research

**Researched:** 2026-03-26
**Domain:** Go/Wails unified binary — CLI dispatch, stdlib flag.FlagSet subcommands, cross-platform console output, macOS app bundle PATH access
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CLIP-01 | User can run `storcat` with no args to launch GUI (existing behavior preserved) | os.Args dispatch — fall-through to runGUI() when no known subcommand detected |
| CLIP-02 | User can run `storcat <command>` to execute CLI commands from the same binary | cli/ package with Run(args, version) int entry point; dispatch in main.go before wails.Run() |
| CLIP-03 | CLI dispatch uses stdlib flag.FlagSet (no Cobra dependency) | flag.NewFlagSet per subcommand; manual switch dispatch in cli/cli.go |
| CLIP-04 | CLI commands output errors to stderr and results to stdout | fmt.Fprintln(os.Stderr, ...) for errors; fmt.Fprintln(os.Stdout, ...) for results — explicit per command |
| CLIP-05 | CLI commands exit with code 0 on success, non-zero on error | Run() returns int exit code; main.go calls os.Exit(cli.Run(...)) |
| CLIP-06 | All commands support --help / -h flag | flag.FlagSet with flag.ContinueOnError; manual usage() func per subcommand |
| CLCM-06 | User can run `storcat version` to print version string | Version var from version.go passed as parameter to cli.Run(); one-liner print in cli/version.go |
| CLPC-01 | CLI output works in Windows terminals (console attachment for GUI subsystem binary) | Decision: use -windowsconsole wails build flag OR AttachConsole — must decide in this phase |
| CLPC-02 | macOS Gatekeeper -psn_* argument injection is filtered before CLI dispatch | filterMacOSArgs() strips -psn_ prefixed args before any dispatch |
| CLPC-03 | wails dev hot-reload still works after CLI dispatch changes | Dispatch must fall through to GUI when no known subcommand; never exit on unrecognized args before runGUI() |
| CLPC-04 | macOS install script creates /usr/local/bin/storcat symlink to .app bundle binary | scripts/install-cli.sh — ln -sf /Applications/StorCat.app/Contents/MacOS/StorCat /usr/local/bin/storcat |
</phase_requirements>

---

## Summary

Phase 8 establishes the CLI dispatch skeleton — the structural change that intercepts `os.Args` in `main.go` before `wails.Run()` is called. When a known subcommand is detected, the binary routes to `cli.Run(args, version)` and exits without starting the Wails GUI. When no subcommand is detected, GUI launch proceeds unchanged. This is the critical path for all v2.1.0 CLI features.

The existing service layer (`internal/catalog`, `internal/search`) is already CLI-compatible — both services are plain Go structs with no Wails runtime imports. No changes to business logic are required. The phase delivers: dispatch wiring, `cli/` package scaffolding, `storcat version` command, `--help` on all stubs, exit code contract, `-psn_*` filtering, Windows console compatibility, and macOS install script.

**Key locked decision from STATE.md:** Use `stdlib flag.FlagSet`, NOT Cobra. This is a project constraint, not a recommendation — the research previously considered Cobra but the STATE.md records the final decision as `flag.FlagSet`. Binary size is a project selling point (93% smaller than Electron); Cobra's ~2MB overhead is explicitly rejected.

**Primary recommendation:** Implement `cli.Run(args []string, version string) int` in a new `cli/` package. Dispatch via switch in main.go. For Windows: use `-windowsconsole` wails build flag (simpler, no console flash concern for CLI users). For macOS: filter `-psn_*` before dispatch. Provide `scripts/install-cli.sh` for PATH access.

---

## Project Constraints (from CLAUDE.md)

- Go: use `go fmt`, `golangci-lint`, context-aware functions with `ctx context.Context` where applicable
- Package managers: Go uses `go mod` — no adding packages outside `go get`
- No class-based patterns — Go idiomatic: functions, structs, interfaces
- Error handling: never silent fallbacks that swallow errors; let errors surface
- Testing: Go `testing` package; 80%+ coverage in critical components
- No Cobra dependency — explicitly locked in STATE.md as `stdlib flag.FlagSet` only
- GSD workflow enforcement: all file changes via GSD commands

---

## Standard Stack

### Core (Phase 8 only — no new dependencies needed)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `flag` (stdlib) | Go 1.23 built-in | Per-subcommand flag parsing | Locked decision (STATE.md): no Cobra |
| `os` (stdlib) | Go 1.23 built-in | Args inspection, stdout/stderr, exit codes | Standard — already used throughout |
| `strings` (stdlib) | Go 1.23 built-in | -psn_ prefix filtering | Standard |
| `fmt` (stdlib) | Go 1.23 built-in | Output formatting | Standard |
| `syscall` (stdlib, Windows only) | Go 1.23 built-in | AttachConsole fallback if -windowsconsole not used | Platform-specific stdlib |

### Already-in-go.mod Transitive Dependencies (Zero New Deps)

| Library | Version in go.mod | Purpose | Notes |
|---------|-------------------|---------|-------|
| `github.com/pkg/browser` | v0.0.0-20240102092130 | Cross-platform browser open (for Phase 10 `storcat open`) | Wails transitive dep — already available |
| `github.com/mattn/go-isatty` | v0.0.20 | TTY detection for color output | Wails transitive dep |
| `github.com/mattn/go-colorable` | v0.1.13 | Windows color output | Wails transitive dep |

### Future Dependencies (Phases 9–10, NOT Phase 8)

| Library | Version | Purpose | When |
|---------|---------|---------|------|
| `github.com/olekukonko/tablewriter` | v1.1.4 | Terminal table output for list/search | Phase 9 (list, search commands) |
| `github.com/fatih/color` | v1.19.0 | ANSI colorized tree output | Phase 10 (show colorization, CLOF-05) |

**Phase 8 adds ZERO new external dependencies.** The cli/ package uses only stdlib.

**Installation (Phase 8 — none required):**
```bash
# No new packages — stdlib only for Phase 8
go mod tidy  # after creating cli/ package to clean up
```

---

## Architecture Patterns

### Recommended Project Structure (Changes Only)

```
storcat-wails/
├── main.go                  MODIFIED — add os.Args dispatch + runGUI() extraction
├── app.go                   UNCHANGED
├── version.go               UNCHANGED — Version var stays in main package
├── internal/
│   ├── catalog/service.go   UNCHANGED for Phase 8 (GenerateTextTree added in Phase 10)
│   ├── search/service.go    UNCHANGED
│   └── config/config.go     UNCHANGED
├── pkg/models/catalog.go    UNCHANGED
├── cli/                     NEW directory
│   ├── cli.go               Run() entry point + subcommand dispatch + help
│   ├── version.go           storcat version subcommand (one-liner)
│   └── cli_windows.go       Windows AttachConsole (only if not using -windowsconsole)
└── scripts/
    └── install-cli.sh       NEW — macOS /usr/local/bin symlink installer
```

### Pattern 1: main.go os.Args Dispatch

**What:** Inspect `os.Args` before any Wails initialization. Known subcommands route to `cli.Run()`. Unknown or zero args fall through to `runGUI()`.

**When to use:** Always — this is the only correct integration point.

**Critical constraint:** Never call `os.Exit(1)` on unrecognized args at the dispatch level. Unknown args must fall through to GUI (preserves `wails dev` hot-reload compatibility).

```go
// main.go
func main() {
    // Filter macOS Gatekeeper -psn_* args before any inspection
    args := filterMacOSArgs(os.Args[1:])

    // CLI dispatch: known subcommand → CLI mode
    if len(args) > 0 {
        switch args[0] {
        case "version", "create", "search", "list", "show", "open", "help", "--help", "-h":
            os.Exit(cli.Run(args, Version))
        }
    }

    // GUI mode: no args, or unrecognized args fall through
    runGUI()
}

func runGUI() {
    app := NewApp()
    cfg := app.GetConfig()
    // ... wails.Run(...) as today — moved verbatim from current main()
}

func filterMacOSArgs(args []string) []string {
    filtered := make([]string, 0, len(args))
    for _, arg := range args {
        if !strings.HasPrefix(arg, "-psn_") {
            filtered = append(filtered, arg)
        }
    }
    return filtered
}
```

**Source:** Verified from direct source analysis of main.go + Wails v2.10.2 behavior; PITFALLS.md Pitfall 3 (wails dev double-execution).

### Pattern 2: cli.Run() Signature

**What:** Single entry point in `cli/` package. Returns int exit code. Takes version as parameter (Version var is in `main` package and cannot be directly imported by `cli/` since `main` is not importable).

**Why pass version as parameter:** `version.go` uses `//go:embed wails.json` which is in package `main`. Package `main` is not importable. Passing `Version` as a parameter is the cleanest solution — no new package needed, no duplicate embed.

```go
// cli/cli.go
package cli

import (
    "flag"
    "fmt"
    "os"
)

// Run is the CLI entry point. args is os.Args[1:] (subcommand is args[0]).
// Returns exit code: 0 = success, 1 = runtime error, 2 = usage error.
func Run(args []string, version string) int {
    if len(args) == 0 {
        printUsage(version)
        return 0
    }

    switch args[0] {
    case "version":
        return runVersion(args[1:], version)
    case "create":
        return runCreate(args[1:])
    case "search":
        return runSearch(args[1:])
    case "list":
        return runList(args[1:])
    case "show":
        return runShow(args[1:])
    case "open":
        return runOpen(args[1:])
    case "help", "--help", "-h":
        printUsage(version)
        return 0
    default:
        fmt.Fprintf(os.Stderr, "storcat: unknown command %q\n", args[0])
        fmt.Fprintf(os.Stderr, "Run 'storcat --help' for usage.\n")
        return 2
    }
}

func printUsage(version string) {
    fmt.Printf(`storcat %s

Usage:
  storcat [command]

Commands:
  create    Create a catalog from a directory
  search    Search catalogs for a term
  list      List catalogs in a directory
  show      Display a catalog's tree structure
  open      Open a catalog's HTML in the default browser
  version   Print the version

Flags:
  -h, --help   Show help for a command

Run 'storcat <command> --help' for command-specific help.
`, version)
}
```

### Pattern 3: Per-Subcommand flag.FlagSet (Phase 8 stubs + version)

**What:** Each subcommand uses `flag.NewFlagSet(name, flag.ContinueOnError)` for its own flags. Prints usage on `-h`/`--help`.

**Why ContinueOnError:** With `flag.ExitOnError` (the default), `flag.Parse()` calls `os.Exit(2)` directly on bad input — bypassing our exit code management. `ContinueOnError` returns an error instead, letting us control the exit code and output format.

```go
// cli/version.go
package cli

import (
    "flag"
    "fmt"
    "os"
)

func runVersion(args []string, version string) int {
    fs := flag.NewFlagSet("version", flag.ContinueOnError)
    fs.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: storcat version\n\nPrint the version string.\n")
    }
    if err := fs.Parse(args); err != nil {
        if err == flag.ErrHelp {
            return 0
        }
        return 2
    }
    fmt.Printf("storcat %s\n", version)
    return 0
}
```

**Phase 8 stubs for create/search/list/show/open:** Each returns 1 with "not yet implemented" message. They must parse `--help` correctly (return 0) to satisfy CLIP-06.

### Pattern 4: Windows Console Attachment

**Recommended approach: `-windowsconsole` build flag** (simpler, more reliable than AttachConsole).

**What:** `wails build -windowsconsole` removes the `-H windowsgui` linker flag, giving the binary a console subsystem. CLI output is visible in Windows terminals.

**Trade-off:** A brief console window flashes when double-clicking the GUI app on Windows. StorCat's CLI users are technical; this is acceptable.

**Build command (Windows):**
```bash
wails build -windowsconsole
```

**Alternative — AttachConsole (only if -windowsconsole is ruled out):**
```go
// cli/cli_windows.go
//go:build windows

package cli

import (
    "os"
    "syscall"
)

func init() {
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    attachConsole := kernel32.NewProc("AttachConsole")
    // ATTACH_PARENT_PROCESS = 0xFFFFFFFF
    attachConsole.Call(uintptr(^uint32(0)))
    conout, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
    if err == nil {
        os.Stdout = conout
        os.Stderr = conout
    }
}
```

**Why AttachConsole has drawbacks:** Terminal prompt does not return cleanly after the process exits (extra Enter required). Piped output does not work. The `-windowsconsole` approach is cleaner for scripting use cases.

**Decision required during planning:** Which approach to commit to for v2.1.0. Recommendation is `-windowsconsole`.

### Pattern 5: macOS Install Script

**What:** Shell script that creates a symlink from the binary inside the .app bundle to `/usr/local/bin/storcat`.

```bash
#!/bin/sh
# scripts/install-cli.sh
# Creates /usr/local/bin/storcat symlink for CLI access after macOS .app install.

APP_PATH="/Applications/StorCat.app/Contents/MacOS/StorCat"
LINK_PATH="/usr/local/bin/storcat"

if [ ! -f "$APP_PATH" ]; then
    echo "Error: StorCat.app not found in /Applications" >&2
    echo "Install StorCat.app to /Applications first." >&2
    exit 1
fi

ln -sf "$APP_PATH" "$LINK_PATH"
echo "Installed: $LINK_PATH -> $APP_PATH"
echo "Run 'storcat --help' to verify."
```

**Note on path casing:** `wails.json` has `"outputfilename": "StorCat"` (capital S, C). The binary inside the bundle is `StorCat`, not `storcat`. The symlink name on PATH is `storcat` (lowercase). Both casing variants must be accounted for.

**Verification:** `ls /Applications/StorCat.app/Contents/MacOS/` to confirm exact binary name before finalizing script.

### Anti-Patterns to Avoid

- **Never call `App` struct from CLI:** `App.startup()` expects a Wails context. `App.SelectDirectory()` panics with nil ctx. Use `catalog.NewService()` / `search.NewService()` directly.
- **Never call `wails.Run()` for CLI mode:** Initializes OS WebView, creates menu bar entry, blocks. ~300ms startup for no reason.
- **Never use Cobra:** Explicitly locked out in STATE.md. Adds ~2MB to binary.
- **Never call `flag.Parse()` globally:** Only call `fs.Parse(args)` inside each subcommand handler, after CLI mode is confirmed. Global `flag.Parse()` will interact with Wails internal flags.
- **Never import `wails/v2/pkg/runtime` in `cli/`:** `runtime.BrowserOpenURL` requires a live Wails context. Use `pkg/browser` (already in go.mod).
- **Never `os.Exit(1)` on unrecognized args in dispatch:** Unknown args must fall through to GUI to preserve `wails dev` hot-reload.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-platform browser open | `exec.Command("open"/"xdg-open"/"start"...)` with manual platform switch | `github.com/pkg/browser` (already in go.mod) | Handles all platforms + edge cases; zero new dependency |
| Byte formatting | Custom formatter | `catalog.Service.formatBytes()` (already exported-ready) | Already written and tested in catalog service |
| Catalog path → HTML path | Custom string manipulation | `app.GetCatalogHtmlPath()` logic (copy pattern, not import) | Already validated; replicate the `.json` → `.html` logic |

**Key insight:** The service layer already contains all business logic. Phase 8 is plumbing and wiring, not new logic.

---

## Common Pitfalls

### Pitfall 1: Windows GUI Subsystem Swallows CLI Output (CRITICAL)
**What goes wrong:** `wails build` uses `-H windowsgui` on Windows. All stdout/stderr is silently discarded. `storcat version` produces no output in a Windows terminal.
**Why it happens:** PE subsystem is a linker-time decision. Windows does not attach console to GUI subsystem processes.
**How to avoid:** Use `wails build -windowsconsole` for Windows builds. Document the brief console flash trade-off.
**Warning signs:** `storcat version` on Windows produces no output and no error — complete silence.

### Pitfall 2: macOS -psn_* Crash on Finder Launch
**What goes wrong:** macOS injects `-psn_0_XXXXXXX` arg when launching via Finder/Downloads quarantine. CLI dispatch treats it as an unknown subcommand and either falls through (OK) or crashes a stricter parser.
**Why it happens:** Gatekeeper process serial number injection — macOS system behavior.
**How to avoid:** `filterMacOSArgs()` strips `-psn_` prefixed args before dispatch. Apply unconditionally on all platforms (no-op on Windows/Linux).
**Warning signs:** App launches fine from terminal but crashes or shows CLI error on first Finder launch from Downloads.

### Pitfall 3: wails dev Double-Execution Breaks Hot-Reload
**What goes wrong:** `wails dev` runs the binary twice. If the dispatch exits with error on no-args, the second invocation fails and the dev GUI never opens.
**Why it happens:** Wails dev binding generation pre-run.
**How to avoid:** Zero-arg and unrecognized-arg paths ALWAYS fall through to `runGUI()`. Never `os.Exit` in the pre-dispatch phase for unknown input.
**Warning signs:** `wails dev` exits immediately after adding CLI dispatch code.

### Pitfall 4: Version String Not Accessible from cli/ Package
**What goes wrong:** `version.go` declares `var Version` in `package main`. `cli/` is `package cli`. `main` is not importable. Direct reference fails to compile.
**Why it happens:** Go `main` package is never importable by other packages.
**How to avoid:** Pass `Version` as a parameter: `cli.Run(args, Version)`. Documented in Pattern 2.
**Warning signs:** Compile error "cannot import package main".

### Pitfall 5: macOS .app Bundle — CLI Not on PATH
**What goes wrong:** Users install via DMG drag to /Applications. The binary is at `StorCat.app/Contents/MacOS/StorCat`. No `storcat` command in PATH.
**Why it happens:** macOS .app bundles are not on PATH by design.
**How to avoid:** Provide `scripts/install-cli.sh` with the symlink command. Document in README.
**Warning signs:** `storcat --help` returns "command not found" despite GUI working fine.

### Pitfall 6: flag.ContinueOnError vs flag.ExitOnError
**What goes wrong:** Using default `flag.ExitOnError` means `-h` triggers `os.Exit(2)` inside the flag package, bypassing our exit code management and printing to stderr instead of stdout.
**Why it happens:** `flag.ExitOnError` is the stdlib default — convenient for simple tools, wrong here.
**How to avoid:** Always use `flag.NewFlagSet(name, flag.ContinueOnError)`. Check for `flag.ErrHelp` and return 0.
**Warning signs:** `storcat version --help` exits with code 2 instead of 0; help goes to stderr.

### Pitfall 7: Wails dev -appargs Cannot Pass Flag-Style CLI Args
**What goes wrong:** `wails dev -appargs "--output /tmp"` fails because Wails CLI parses `--output` as its own flag.
**Why it happens:** Known Wails v2 bug (#1533). `-appargs` flag parsing does not properly separate app args from wails args.
**How to avoid:** Test CLI commands using the built binary directly: `./build/bin/StorCat version`. Add a Makefile target for CLI testing. Do not rely on `wails dev -appargs` for flag-style CLI testing.
**Warning signs:** Flag-style args work in built binary but fail in `wails dev`.

---

## Code Examples

### Verified: os.Args-before-wails Pattern (from ARCHITECTURE.md research)

```go
// main.go — confirmed pattern from Wails community discussion #4175
// Source: verified via direct source analysis + Wails v2 behavior
func main() {
    args := filterMacOSArgs(os.Args[1:])

    if len(args) > 0 {
        switch args[0] {
        case "version", "create", "search", "list", "show", "open", "help", "--help", "-h":
            os.Exit(cli.Run(args, Version))
        }
    }

    runGUI()
}
```

### Verified: flag.FlagSet with ContinueOnError

```go
// cli/version.go — stdlib pattern
// Source: Go stdlib flag package documentation
func runVersion(args []string, version string) int {
    fs := flag.NewFlagSet("version", flag.ContinueOnError)
    fs.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: storcat version\n\nPrint the StorCat version.\n")
    }
    if err := fs.Parse(args); err != nil {
        if err == flag.ErrHelp {
            return 0  // --help is success, not error
        }
        fmt.Fprintf(os.Stderr, "storcat version: %v\n", err)
        return 2
    }
    fmt.Printf("storcat %s\n", version)
    return 0
}
```

### Verified: -psn_ Filter (from PITFALLS.md research, confirmed via Godot PR #37719)

```go
// Applies to all platforms — no-op on Windows/Linux
func filterMacOSArgs(args []string) []string {
    filtered := make([]string, 0, len(args))
    for _, arg := range args {
        if !strings.HasPrefix(arg, "-psn_") {
            filtered = append(filtered, arg)
        }
    }
    return filtered
}
```

### Verified: Service Instantiation from CLI Context

```go
// Correct — bypasses App struct entirely
// Source: direct analysis of internal/catalog/service.go (no Wails imports)
svc := catalog.NewService()
result, err := svc.CreateCatalog(title, dir, output, "", nil)

// WRONG — App.ctx is nil in CLI mode, panics if any runtime method is called
app := NewApp()  // Do NOT do this in CLI code
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Separate CLI binary | Unified binary with os.Args dispatch | Wails community pattern circa 2022+ | Single binary for GUI + CLI; no separate distribution |
| Cobra for all Go CLIs | stdlib flag.FlagSet for small CLIs | Ongoing ecosystem maturation | Justified for <10 subcommands with binary size constraints |
| Manual platform browser open | `pkg/browser` | ~2020 | Cross-platform handling without platform switch boilerplate |

**Deprecated/outdated:**
- `runtime.BrowserOpenURL`: Wails-runtime-only function — cannot be used in CLI mode (requires active ctx).
- `flag.Parse()` globally: Pattern from single-command tools — wrong when mixing with framework that injects its own flags.

---

## Open Questions

1. **Windows build: `-windowsconsole` vs `AttachConsole`**
   - What we know: Both approaches work; `-windowsconsole` is simpler and more reliable for piping; `AttachConsole` preserves "no console flash on double-click" UX
   - What's unclear: Whether the console flash on double-click is acceptable to the project owner for Windows users
   - Recommendation: Decide during Phase 8 planning. Lean toward `-windowsconsole` — StorCat's CLI users are technical and the console flash is a minor UX cost. Document the trade-off in release notes.

2. **macOS binary name casing**
   - What we know: `wails.json` has `"outputfilename": "StorCat"` (capital S, C); the binary is likely `StorCat` not `storcat`
   - What's unclear: Need to verify exact binary name after wails build to ensure install script uses correct path
   - Recommendation: Verify with `ls build/bin/` after first wails build; finalize install-cli.sh path accordingly

3. **Phase 8 scope: stub commands vs. no stubs**
   - What we know: CLIP-06 requires `--help` on all commands; CLCM-06 is version only; create/search/list/show/open are Phase 9–10
   - What's unclear: Should Phase 8 include stub subcommand handlers for commands not yet implemented?
   - Recommendation: Yes — implement stubs that parse `--help` correctly and return `fmt.Fprintf(os.Stderr, "not yet implemented\n")` with exit 1. This satisfies CLIP-02 (binary dispatches to them) and CLIP-06 (--help works) without Phase 9 leaking into Phase 8.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | CLI package compilation | Yes | go1.26.1 darwin/arm64 | — |
| Wails CLI | `wails build`, `wails dev` | Yes | v2.10.2 | — |
| `go test` | Test validation | Yes | go1.26.1 | — |
| Windows machine | CLPC-01 validation | Unknown | — | Document as manual verification item; cannot test on macOS |

**Missing dependencies with no fallback:**
- Windows terminal for CLPC-01 validation — Windows console output behavior MUST be verified on real Windows before shipping. This cannot be simulated on macOS. Flag as manual verification step in Phase 8 acceptance criteria.

**Missing dependencies with fallback:**
- None.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` package (built-in) |
| Config file | none — `go test ./...` auto-discovers |
| Quick run command | `go test ./cli/... -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CLIP-01 | No-arg invocation falls through to GUI (dispatch logic) | unit | `go test ./cli/... -run TestDispatch` | No — Wave 0 |
| CLIP-02 | Known subcommand routes to cli.Run() | unit | `go test ./cli/... -run TestRun` | No — Wave 0 |
| CLIP-03 | Dispatch uses flag.FlagSet not Cobra | unit (import check) | `go test ./cli/... -run TestNoCobra` (check imports) | No — Wave 0 |
| CLIP-04 | Errors go to stderr, results to stdout | unit | `go test ./cli/... -run TestOutputRouting` | No — Wave 0 |
| CLIP-05 | Exit code 0 on success, non-zero on error | unit | `go test ./cli/... -run TestExitCodes` | No — Wave 0 |
| CLIP-06 | --help returns exit 0 on all commands | unit | `go test ./cli/... -run TestHelp` | No — Wave 0 |
| CLCM-06 | `storcat version` prints version and exits 0 | unit | `go test ./cli/... -run TestVersion` | No — Wave 0 |
| CLPC-02 | -psn_ args filtered before dispatch | unit | `go test -run TestFilterMacOSArgs` | No — Wave 0 |
| CLPC-03 | wails dev still opens GUI after dispatch changes | manual | `wails dev` — verify window opens | — |
| CLPC-01 | CLI output visible in Windows terminal | manual | Windows terminal: `storcat version` | — |
| CLPC-04 | install-cli.sh creates working symlink | manual | Run script on macOS, verify `storcat --help` | — |

### Sampling Rate
- **Per task commit:** `go test ./cli/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** `go test ./...` green + manual verification of CLPC-01/03/04 before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `cli/cli_test.go` — covers CLIP-01 through CLIP-06, CLPC-02
- [ ] `cli/version_test.go` — covers CLCM-06

*(No new test framework needed — `go test` is already the project standard with existing passing tests in `internal/catalog`, `internal/search`, `internal/config`)*

---

## Sources

### Primary (HIGH confidence)
- Direct source analysis: `main.go`, `app.go`, `version.go`, `internal/catalog/service.go`, `internal/search/service.go` — confirmed no Wails runtime imports in service layer; confirmed `Version` var is in `package main`
- `go list -m -json` — verified `pkg/browser` v0.0.0-20240102092130, `go-isatty` v0.0.20, `go-colorable` v0.1.13 already in go.mod as indirect Wails deps
- `go list -m -json cobra/tablewriter/fatih@latest` — confirmed cobra v1.10.2 (Dec 2025), tablewriter v1.1.4 (Mar 2026), fatih/color v1.19.0 (Mar 2026)
- Go stdlib `flag` package documentation — `flag.ContinueOnError`, `flag.ErrHelp` behavior
- `.planning/STATE.md` — locked decision: `stdlib flag.FlagSet`, not Cobra; `cli.Run(args []string, version string) int`

### Secondary (MEDIUM confidence)
- Wails v2 community discussion #4175 — os.Args in production builds; `wails dev` double-execution pattern
- Wails GitHub issue #544 — Windows console output limitation confirmed by maintainer
- Wails GitHub issue #1533 — `-appargs` flag passthrough bug confirmed
- Windows AttachConsole API: https://learn.microsoft.com/en-us/windows/console/attachconsole
- macOS -psn_ Gatekeeper injection documented in Godot PR #37719

### Tertiary (LOW confidence)
- `.planning/research/PITFALLS.md`, `ARCHITECTURE.md`, `STACK.md`, `FEATURES.md` — prior project research (2026-03-26) that fed the STATE.md decisions; cross-referenced against source code

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified via `go list -m`, confirmed no new deps needed for Phase 8
- Architecture: HIGH — verified against actual source files; service layer confirmed Wails-free
- Pitfalls: HIGH — Windows pitfall confirmed via Wails issue #544; macOS pitfall via Godot PR; wails dev pattern from discussion #4175
- Test infrastructure: HIGH — existing `go test ./...` passes; test files for `cli/` are Wave 0 gaps

**Research date:** 2026-03-26
**Valid until:** 2026-09-26 (stable domain — Go stdlib, Wails v2 behavior; 6-month validity)
