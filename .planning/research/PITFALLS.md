# Pitfalls Research

**Domain:** Adding CLI subcommands to an existing Go/Wails desktop app
**Project:** StorCat v2.1.0
**Researched:** 2026-03-26
**Confidence:** MEDIUM-HIGH (Wails-specific sources + Go/Windows platform docs; some gaps noted)

---

## Critical Pitfalls

---

### Pitfall 1: Windows GUI Subsystem Swallows All CLI Output

**What goes wrong:**
`wails build` compiles with `-H windowsgui` in ldflags. On Windows this sets the PE subsystem to WINDOWS, which means the OS does not attach stdout/stderr to the parent console. Every `fmt.Println`, `fmt.Fprintf(os.Stderr, ...)`, and `os.Exit` in CLI mode produces zero output. The terminal hangs and the user sees nothing. This is not a code bug — the binary is running correctly, its output is simply discarded by the OS.

**Why it happens:**
Wails hardcodes the GUI subsystem flag to prevent a console window from flashing when users double-click the app. The flag is a binary linker-time decision; there is no runtime way to "re-attach" cleanly. Go developers coming from macOS/Linux are unaware of this distinction because it does not exist on those platforms.

**How to avoid:**
Two options, each with trade-offs:

Option A — Build with `-windowsconsole` Wails flag:
```bash
wails build -windowsconsole
```
This removes `-H windowsgui`, giving CLI output. Cost: a console window briefly flashes when the app is double-clicked on Windows for GUI mode.

Option B — Use Windows `AttachConsole` API at runtime (GUI subsystem preserved):
```go
//go:build windows

package main

import (
    "os"
    "syscall"
    "unsafe"
)

func attachWindowsConsole() {
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    attachConsole := kernel32.NewProc("AttachConsole")
    attachConsole.Call(uintptr(^uint32(0))) // ATTACH_PARENT_PROCESS = -1
    // Reopen stdout/stderr to CONOUT$
    conout, _ := os.OpenFile("CONOUT$", os.O_RDWR, 0)
    os.Stdout = conout
    os.Stderr = conout
}
```
Call this at the top of `main()` before argument dispatch. Cost: stdin is unreliable; extra Enter keypress may be needed to release the terminal prompt after CLI subcommands complete. Does not work for piped output.

**Recommendation for StorCat:** Option A with `-windowsconsole`. The flash is a known acceptable trade-off. StorCat CLI users are technically-oriented; they will not be confused by the brief console. Document it.

**Warning signs:**
- `storcat create .` on Windows produces no output at all
- No error code from the terminal either — silence
- Testing only on macOS misses this entirely

**Phase to address:** Phase 1 (CLI dispatch skeleton). Must validate on Windows before any CLI implementation is considered done.

---

### Pitfall 2: macOS Gatekeeper Injects `-psn_*` Argument When Launching via Finder

**What goes wrong:**
When a macOS `.app` bundle is launched by Finder or the `open` command (as opposed to running the binary directly from terminal), macOS Gatekeeper injects a process serial number argument like `-psn_0_8423432`. If the CLI argument parser (Cobra, `flag`, or manual `os.Args` check) encounters this unrecognized flag, it will either crash, print a usage error, or misfire a CLI command. This only happens on the first launch of downloaded apps that carry the `com.apple.quarantine` extended attribute.

**Why it happens:**
It is a macOS system behavior, not a Wails bug. The same process that handles Gatekeeper quarantine checks passes the PSN to identify the process. Standard CLI parsers have no reason to know about PSN and will treat it as an unknown flag.

**How to avoid:**
Filter `-psn_*` arguments before any CLI dispatch:
```go
func filterMacOSArgs(args []string) []string {
    filtered := make([]string, 0, len(args))
    for _, arg := range args {
        if !strings.HasPrefix(arg, "-psn_") {
            filtered = append(filtered, arg)
        }
    }
    return filtered
}

// In main(), before dispatch:
args := filterMacOSArgs(os.Args[1:])
```
Apply this filter on all platforms (it is a no-op on Windows/Linux) or guard it with `//go:build darwin`.

**Warning signs:**
- App works fine when tested from terminal but crashes or shows usage errors on fresh first launch from Finder/Downloads
- Error message mentions `-psn_0_...` as an unknown flag
- Only reproducible on macOS with a freshly downloaded/unsigned build

**Phase to address:** Phase 1 (CLI dispatch skeleton). Add filter as part of the args-parsing setup.

---

### Pitfall 3: Wails `wails dev` Double-Executes main() During Binding Generation

**What goes wrong:**
During development, `wails dev` runs the binary twice: once without arguments (to generate binding information) and once with the actual arguments. Any code in `main()` that validates argument presence or calls `os.Exit(1)` when no args are provided will cause `wails dev` to exit before the app launches. The developer sees a confusing immediate exit and cannot use the dev hot-reload workflow.

**Why it happens:**
Wails' development mode uses reflection-based binding generation, which requires a fresh process start. This is a `wails dev`-specific behavior; production builds run once. Developers writing "require subcommand or show usage and exit" logic in `main()` do not account for this double-execution.

**How to avoid:**
Use the `production` build tag to guard strict argument validation:
```go
//go:build production

package main

func isProduction() bool { return true }
```
```go
//go:build !production

package main

func isProduction() bool { return false }
```
```go
// In main():
if isProduction() && len(args) == 0 {
    // launch GUI
}
```
Alternatively: never call `os.Exit` on missing args during the pre-dispatch phase in dev mode. Reserve strict validation for within subcommand handlers.

The safest pattern: `os.Args` check dispatch should always fall through to GUI launch when no subcommand is recognized, not exit with an error. "Unknown args → launch GUI" is safer than "unknown args → exit 1".

**Warning signs:**
- `wails dev` exits immediately with a usage/help message
- The app never opens the GUI window in dev mode after adding CLI dispatch
- Works fine with `wails build` but not with `wails dev`

**Phase to address:** Phase 1 (CLI dispatch skeleton). Test `wails dev` still works after dispatch code is added.

---

### Pitfall 4: `embed.FS` Fails to Compile If Frontend Assets Are Not Pre-Built

**What goes wrong:**
`main.go` has `//go:embed all:frontend/dist`. If `frontend/dist` does not exist when `go build` (or `wails build`) runs, the build fails with:
```
pattern all:frontend/dist: directory prefix frontend/dist does not exist
```
This is guaranteed to happen on a fresh clone, in CI without a frontend build step, or when running `go build` directly instead of `wails build`.

**Why it happens:**
`//go:embed` is a compile-time directive. The directory must exist at compile time. `wails build` handles this by building the frontend first. Running `go build` directly skips the frontend build step.

**How to avoid:**
- Never use `go build` directly for the full binary. Always use `wails build`.
- For CI and cross-platform builds, ensure the frontend build step runs before the Go build: `cd frontend && npm install && npm run build`.
- For CLI-only development, consider a build tag to stub out the embedded assets:
  ```go
  //go:build !production

  package main

  import "embed"

  //go:embed all:frontend/dist
  var assets embed.FS  // Will fail if dist missing — use wails dev for GUI
  ```
  But only if you never need to run the binary directly in non-production mode.

**Warning signs:**
- `go build` fails on fresh clone with `directory prefix does not exist`
- CI pipeline fails on the Go build step without an obvious Go error
- `wails build` succeeds but direct `go build` fails

**Phase to address:** Phase 1. Document in the build guide that `wails build` is required, not `go build`.

---

### Pitfall 5: CLI Mode Loads Embedded Frontend Assets Into Memory Unnecessarily

**What goes wrong:**
The `//go:embed all:frontend/dist` directive causes the entire compiled React app (~2-5MB) to be memory-mapped into the process when the binary starts. In CLI mode (`storcat create ...`), this overhead is paid even though no GUI is ever opened. For a fast CLI tool, this is wasted startup overhead and memory.

**Why it happens:**
Go's `embed.FS` is initialized at binary load time. The data is part of the binary's read-only data segment and is mapped into memory when the OS loads the process, before `main()` executes.

**How to avoid:**
This is a structural limitation of the unified binary approach and cannot be fully eliminated. The embedded assets are part of the binary's data segment regardless of execution mode.

Mitigation: The assets are in read-only memory and do not contribute to heap usage. The OS may page them out if they are never accessed. For a desktop app binary, ~3-5MB of mapped-but-unused data is an acceptable trade-off. The alternative (build tag splitting) would require maintaining two separate binaries.

Acceptable cost: Accept it. StorCat's binary is already ~10MB for a Go/Wails app. Adding the CLI does not change binary size. If startup time becomes measurable, revisit.

**Warning signs:**
- CLI commands feel noticeably slow to start (>200ms for simple operations)
- Memory usage of CLI invocation is surprisingly high
- This is not expected to be a real problem for StorCat's scale

**Phase to address:** Not a blocker. Flag as accepted trade-off in architecture decisions.

---

## Moderate Pitfalls

---

### Pitfall 6: Cobra/flag Argument Parsing Conflicts With Wails Internal Flags

**What goes wrong:**
Wails injects its own internal flags into the process under certain conditions (primarily during `wails dev`). Using Go's `flag.Parse()` globally at startup will consume these flags and may produce errors or unintended side-effects. Additionally, `cobra` uses `pflag` which calls `flag.CommandLine.Parse()` — if both Cobra and any library that uses the standard `flag` package are in play, you can get "flag provided but not defined" panics.

**Why it happens:**
Go's `flag` package uses a global `CommandLine` FlagSet. Any library that registers flags to `flag.CommandLine` (including `wails dev` infrastructure) will be seen by any code that calls `flag.Parse()`. Cobra's `pflag` is a separate implementation but if both are active, interactions are possible.

**How to avoid:**
- Use `cobra` for CLI dispatch and do NOT call `flag.Parse()` globally in `main()`.
- Cobra handles its own parsing via `rootCmd.Execute()`.
- Only call CLI parsing after the dispatch decision is made (i.e., after confirming we are in CLI mode, not GUI mode).
- Avoid registering global flags to `flag.CommandLine`.

**Warning signs:**
- `wails dev` produces "flag provided but not defined: -some-internal-flag"
- Cobra panics with "use of uninitialized flag set"
- CLI help output includes internal Wails flags not intended for users

**Phase to address:** Phase 1 (CLI dispatch skeleton) when choosing the argument parsing approach.

---

### Pitfall 7: macOS `.app` Bundle Makes the Binary Inaccessible From PATH

**What goes wrong:**
`wails build` on macOS produces `build/bin/storcat.app/Contents/MacOS/storcat` — a binary inside an `.app` bundle. Users who install by dragging to `/Applications` get the GUI app but the `storcat` CLI command is not on their PATH. Running `storcat create ...` in a terminal fails with "command not found". This defeats the purpose of adding CLI subcommands.

**Why it happens:**
macOS `.app` bundles are not on PATH by default. The binary is buried inside the bundle's directory structure. This is intentional for GUI apps (users interact via Finder/dock) but is the wrong UX for CLI usage.

**How to avoid:**
Three approaches:

Option A — Symlink documentation: Document `ln -s /Applications/storcat.app/Contents/MacOS/storcat /usr/local/bin/storcat`. Simple but requires user action.

Option B — Companion install script: A `scripts/install-cli.sh` that creates the symlink automatically. Can be included in the release and run post-install.

Option C — Separate CLI binary in release: Build both `storcat.app` (GUI) and a standalone `storcat` CLI binary. The CLI binary has no embedded assets (use build tags). Distribute both in the DMG. More complex build pipeline but cleanest user experience.

**Recommendation:** Option A+B for v2.1.0 MVP. Document the symlink in the release notes and include a `scripts/install-cli.sh`. Option C is a v2.2.0 concern if users find the symlink approach burdensome.

**Warning signs:**
- CLI feature ships but no installation path is documented
- `storcat --help` only works if the user already knows about the `.app` bundle path
- macOS users report "command not found" despite having the app installed

**Phase to address:** Phase 1 (CLI dispatch skeleton) — settle the installation approach before implementing subcommands. Implement install script in same phase.

---

### Pitfall 8: Windows PATH and File Path Separator Gotchas in CLI Output

**What goes wrong:**
On Windows, directory paths use backslash (`\`) but many Go standard library functions return forward-slash paths or mixed paths. When the CLI outputs paths (e.g., `storcat create C:\Users\ken\docs` → output: `Catalog saved to C:/Users/ken/docs/catalog.json`), the path format inconsistency looks wrong to Windows users. Additionally, when users supply paths with backslashes in CLI arguments, Go's `filepath` functions need to normalize them.

Separate issue: On Windows, `os.Args[0]` may include the `.exe` extension or the full path to the binary. Any code that inspects `os.Args[0]` to determine the binary name (e.g., for help text) must strip the extension and path.

**Why it happens:**
Go uses OS-native path separators in `filepath` functions but uses forward slashes in URL contexts and some standard library outputs. Mixed-slash output is a common Go-on-Windows bug. `os.Args[0]` differences are a Windows-only quirk.

**How to avoid:**
- Always use `filepath.ToSlash()` and `filepath.FromSlash()` explicitly when outputting paths to users vs. passing paths to OS APIs.
- For CLI help text that shows the binary name: `filepath.Base(strings.TrimSuffix(os.Args[0], ".exe"))`.
- Test CLI output paths on Windows with directories that include spaces and backslashes.

**Warning signs:**
- CLI output on Windows shows forward slashes in file paths
- Help text on Windows shows the full path to the binary or includes `.exe` extension
- Path arguments with backslashes cause "file not found" errors

**Phase to address:** Per-command implementation phases. Add Windows path normalization to each command's output logic.

---

### Pitfall 9: wails dev `-appargs` Does Not Properly Pass Flag-Style Arguments

**What goes wrong:**
When testing CLI commands during development using `wails dev -appargs "create /some/path"`, if any argument starts with a `-` (flag-style), Wails dev mode intercepts and fails to pass it through. Example: `wails dev -appargs "--output /tmp"` will fail because `--output` is parsed by the Wails CLI, not passed to the app binary.

**Why it happens:**
This is a known Wails v2 bug (GitHub issue #1533). The `-appargs` parsing does not properly quote or escape flag-style arguments from the application's perspective.

**How to avoid:**
- During development, test CLI behavior using the built binary directly: `./build/bin/storcat create /some/path`.
- For `wails dev`, use only positional arguments (no flags) in `-appargs` during development testing.
- Add a Makefile target: `make cli-test` that runs `wails build` + exercises the binary directly.

**Warning signs:**
- `wails dev -appargs "--flag value"` exits immediately with Wails CLI usage
- Arguments that are flags work in the built binary but not in dev mode
- Only affects development workflow, not production behavior

**Phase to address:** Phase 1 — document the limitation in the development workflow. Do not block implementation on it.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Single `main.go` dispatch with `if/else` instead of Cobra | Less code, no new dependency | Help text, completions, flag handling all manual; grows unmanageably | Never for more than 2-3 subcommands |
| Testing CLI only on macOS (dev machine) | Fast iteration | Windows console subsystem bug ships; macOS `.app` PATH issue undiscovered | Never |
| Using `fmt.Println` for structured CLI output | Simple to write | No `--json` output flag possible later; harder to test output programmatically | Acceptable for v2.1.0 MVP if `--json` is out of scope |
| Hardcoding output directory defaults in CLI flags | Simpler UX | Conflicts with GUI defaults stored in config; users confused by two separate defaults | Never — read defaults from the same config.Manager the GUI uses |
| Sharing `App` struct between GUI and CLI modes | Code reuse | `App.startup(ctx)` requires a Wails context; CLI mode has no context; panics if any runtime call is made | Never — create a separate CLI service layer that calls internal packages directly |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Wails build system | Running `go build` directly for full binary | Always use `wails build`; `go build` fails without pre-built frontend assets |
| Cobra + Wails | Calling `flag.Parse()` globally before dispatch | Dispatch to Cobra only after CLI mode is confirmed; never call `flag.Parse()` globally |
| Windows console | Assuming stdout works like macOS/Linux | Build with `-windowsconsole` or implement `AttachConsole`; test on real Windows |
| macOS app bundle | Testing CLI via `./storcat` inside bundle path | Test via `storcat.app/Contents/MacOS/storcat`; provide symlink install script |
| Internal services | Calling `App` methods from CLI (they use Wails ctx) | Call `internal/catalog` and `internal/search` packages directly; bypass the `App` struct |
| Config manager | CLI mode using different defaults than GUI | CLI reads from the same `config.Manager` using `config.NewManager()`; no separate config |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| `storcat search` scanning entire disk | Long hangs with no output; looks frozen | Always require explicit `-dir` flag or default to current directory; never default to `~` | Any directory with >10k files |
| CLI output buffering | Final lines missing when piped | Use `os.Stdout.Sync()` or flush before `os.Exit`; Cobra handles this on `RunE` return | Any piped output |
| Loading full catalog JSON for `storcat show` large catalogs | Slow `show` command for 1M-file catalogs | Stream or limit depth; add `--depth N` flag | Catalogs >100k entries |

---

## "Looks Done But Isn't" Checklist

- [ ] **Windows CLI output:** `storcat version` prints something on Windows (not silent). Test with a real Windows terminal, not just macOS.
- [ ] **macOS app bundle CLI:** Users can call `storcat create` from a terminal after a fresh macOS install — symlink or installer is documented and tested.
- [ ] **`-psn_` filtering:** App launched from macOS Finder after fresh download does not crash or print a usage error.
- [ ] **`wails dev` still works:** After adding CLI dispatch code, `wails dev` still opens the GUI window (double-execution regression check).
- [ ] **GUI launch still works:** Running `storcat` with no arguments (double-click or bare terminal invocation) still opens the GUI, not a CLI help screen.
- [ ] **Cobra root command:** Running `storcat --help` shows subcommands, not the Wails app. Running `storcat` with no args still launches GUI.
- [ ] **Exit codes:** `storcat create /nonexistent` exits with code 1, not 0. `storcat version` exits with code 0.
- [ ] **Path output on Windows:** `storcat create` output paths use Windows-native separators.
- [ ] **Internal package isolation:** CLI commands call `internal/catalog` and `internal/search` directly — zero Wails runtime imports in CLI code paths.

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Windows console output silent | LOW | Add `-windowsconsole` to `wails build` flags; rebuild |
| `wails dev` broken by CLI dispatch | LOW | Add `!isProduction()` guard around strict arg validation; test |
| `-psn_` crash on macOS Finder launch | LOW | Add filter before dispatch; one function, one test |
| GUI no longer launches (dispatch bug) | MEDIUM | Revert dispatch logic to explicit subcommand check; nil/empty args → GUI |
| macOS PATH issue discovered post-release | MEDIUM | Add install script to patch release; document in GitHub README |
| Cobra conflicts with Wails internal flags | MEDIUM | Move `rootCmd.Execute()` inside CLI branch only; never call at startup unconditionally |
| `App` struct used in CLI (Wails ctx nil panic) | HIGH | Refactor — extract business logic into standalone functions; requires service layer |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Windows console output silent | Phase 1: CLI dispatch skeleton | `storcat version` on Windows prints version string |
| macOS `-psn_` injection | Phase 1: CLI dispatch skeleton | Launch from Finder after fresh download; no crash |
| `wails dev` double-execution | Phase 1: CLI dispatch skeleton | `wails dev` still opens GUI window |
| `embed.FS` compile failure | Phase 1: build documentation | Fresh clone + `wails build` succeeds; `go build` documented as unsupported |
| macOS `.app` bundle PATH | Phase 1: CLI dispatch skeleton | Install script exists; README documents symlink |
| Cobra/flag conflicts | Phase 1: CLI dispatch skeleton | `wails dev` produces no flag-parse errors |
| Wails dev `-appargs` flag issue | Phase 1: development workflow | Makefile `cli-test` target using built binary |
| Windows path separators | Per-command phases (create, search, etc.) | CLI output paths on Windows use `\` |
| App struct / Wails ctx in CLI | Phase 1: service layer design | CLI commands have zero imports from `wails/v2/pkg/runtime` |

---

## Sources

- Wails console output issue #544: https://github.com/wailsapp/wails/issues/544
- Wails debug logging with `-windowsconsole` issue #3008: https://github.com/wailsapp/wails/issues/3008
- Wails `-appargs` flag passthrough bug #1533: https://github.com/wailsapp/wails/issues/1533
- Wails CLI args after build discussion #4175: https://github.com/wailsapp/wails/discussions/4175
- Wails CLI with app discussion #3098: https://github.com/wailsapp/wails/discussions/3098
- Wails build tags issue #1610: https://github.com/wailsapp/wails/issues/1610
- Wails pass command line args to main.go issue #2353: https://github.com/wailsapp/wails/issues/2353
- Windows AttachConsole API: https://learn.microsoft.com/en-us/windows/console/attachconsole
- Windows dual-mode app technique: https://www.tillett.info/2013/05/13/how-to-create-a-windows-program-that-works-as-both-as-a-gui-and-console-application/
- macOS `-psn_` Gatekeeper injection (Godot fix): https://github.com/godotengine/godot/pull/37719
- macOS app bundle CLI access: https://www.jviotti.com/2022/11/28/launching-macos-applications-from-the-command-line.html
- Go Windows no println output with windowsgui: https://forum.golangbridge.org/t/no-println-output-with-go-build-ldflags-h-windowsgui/7633

---
*Pitfalls research for: Adding CLI subcommands to Go/Wails desktop app*
*Researched: 2026-03-26*
