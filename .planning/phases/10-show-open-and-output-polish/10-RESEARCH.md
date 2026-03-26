# Phase 10: Show, Open, and Output Polish - Research

**Researched:** 2026-03-26
**Domain:** Go CLI tree rendering, cross-platform browser launch, ANSI color control
**Confidence:** HIGH

## Summary

Phase 10 completes the final two stub CLI commands (`show` and `open`) and adds cross-cutting output polish (color, depth limiting) to `show`. The codebase is already well-prepared: both `github.com/fatih/color` and `github.com/pkg/browser` are in `go.mod` as indirect dependencies (pulled in by Wails), `search.LoadCatalog` already handles both v1 and v2 JSON formats, and the HTML tree-drawing logic in `catalog.generateTreeStructure` provides the exact connector characters (`├── `, `└── `, `│   `) to replicate for terminal output.

The implementation pattern for `show` is: load JSON with `search.NewService().LoadCatalog(path)`, then walk the `*models.CatalogItem` tree recursively using the same connector logic already proven in `generateTreeStructure`. For color, promote `fatih/color` from indirect to direct, gate on `color.NoColor` (which already respects `NO_COLOR` env var and non-TTY automatically), and add a `--no-color` flag that sets `color.NoColor = true` before any output. For `open`, load the catalog, derive the HTML path by replacing `.json` with `.html`, and call `pkg/browser.OpenFile(htmlPath)` — this library handles `open`/`xdg-open`/`start` per-platform transparently.

**Primary recommendation:** Implement both commands in new files `cli/show.go` and `cli/open.go`, replacing `stubs.go` entries. Use `fatih/color` for colorization and `pkg/browser` for browser launch — both are already in the module cache.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CLCM-04 | `storcat show <catalog.json>` displays catalog tree structure | `search.LoadCatalog` loads JSON; custom recursive walker prints tree with `├──`/`└──` connectors |
| CLCM-05 | `storcat open <catalog.json>` opens catalog HTML in default browser | `pkg/browser.OpenFile(htmlPath)` handles macOS/Linux/Windows; HTML path derived by replacing `.json` with `.html` |
| CLOF-02 | `show` supports `--json` flag to output raw catalog JSON | Load `*models.CatalogItem` then call existing `printJSON(root)` helper from `output.go` |
| CLOF-04 | `show` supports `--depth N` flag to limit tree depth | Pass `maxDepth int` parameter to recursive walker; stop recursing when `currentDepth >= maxDepth`; -1 = unlimited |
| CLOF-05 | `show` displays colorized tree (directories bold/blue) on TTY | `color.New(color.FgBlue, color.Bold).Sprint(name)` for directories; color auto-disabled on non-TTY |
| CLOF-06 | All commands respect `--no-color` flag and `NO_COLOR` env var | `fatih/color.NoColor` already checks `NO_COLOR` env var; `--no-color` flag sets `color.NoColor = true` |
| CLPC-05 | `storcat open` works cross-platform | `pkg/browser.OpenFile` calls `open`/`xdg-open`/`cmd /c start` per platform |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/fatih/color` | v1.18.0 | ANSI color output with TTY detection | Already in go.mod (indirect via Wails); handles NO_COLOR, isatty, Windows colorable automatically |
| `github.com/pkg/browser` | v0.0.0-20240102092130-5ac0b6a4141c | Cross-platform browser/file launcher | Already in go.mod (indirect); covers macOS `open`, Linux `xdg-open`, Windows `start` |
| `github.com/mattn/go-isatty` | v0.0.20 | TTY detection | Already in go.mod (indirect, used by fatih/color internally) |
| `storcat-wails/internal/search` | (internal) | `LoadCatalog` for JSON parsing with v1/v2 compat | Already handles both array-wrapped (v1) and bare-object (v2) formats |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `storcat-wails/pkg/models` | (internal) | `CatalogItem` struct | Tree walker input type |
| `storcat-wails/cli.printJSON` | (internal) | Indented JSON to stdout | `show --json` path |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `fatih/color` | raw ANSI escape codes | fatih/color already in module; handles Windows VT mode, NO_COLOR, isatty automatically — don't hand-roll |
| `pkg/browser` | `os/exec` + platform switch | pkg/browser already in module; handles edge cases per platform — don't hand-roll |

**Installation:** No new `go get` required — both libraries are already in `go.mod` as indirect dependencies. Promoting them to direct dependencies is done by importing them in non-test production code and running `go mod tidy`.

**Version verification:** Confirmed via `go.mod` — `github.com/fatih/color v1.18.0` and `github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c` are present in the module graph.

## Architecture Patterns

### Recommended Project Structure

Two new files replace the two stub functions currently in `stubs.go`:

```
cli/
├── cli.go          # Run() dispatcher (no changes)
├── output.go       # printJSON, formatBytes (no changes)
├── create.go       # runCreate (no changes)
├── list.go         # runList (no changes)
├── search.go       # runSearch (no changes)
├── version.go      # runVersion (no changes)
├── show.go         # NEW: runShow, printTree, color helper
├── open.go         # NEW: runOpen
├── stubs.go        # REMOVE both functions after implementing
├── show_test.go    # NEW: tests for runShow
└── open_test.go    # NEW: tests for runOpen (mocked browser)
```

### Pattern 1: Flag Parsing for `show`

**What:** `show` has three flags (`--json`, `--depth N`, `--no-color`) and one positional arg (the catalog file path). The interspersed flag separation pattern used by `list` and `search` applies here.

**When to use:** Whenever positional args and flags need to coexist in any order.

```go
// Source: established pattern in cli/list.go and cli/search.go
func runShow(args []string) int {
    fs := flag.NewFlagSet("show", flag.ContinueOnError)
    jsonFlag := fs.Bool("json", false, "Output raw catalog JSON")
    depthFlag := fs.Int("depth", -1, "Limit tree depth (-1 = unlimited)")
    noColorFlag := fs.Bool("no-color", false, "Disable color output")
    fs.Usage = func() { ... }

    var positional []string
    var flagArgs []string
    for _, a := range args {
        if strings.HasPrefix(a, "-") {
            flagArgs = append(flagArgs, a)
        } else {
            positional = append(positional, a)
        }
    }
    if err := fs.Parse(flagArgs); err != nil {
        if err == flag.ErrHelp { return 0 }
        return 2
    }
    if len(positional) < 1 {
        fmt.Fprintf(os.Stderr, "storcat show: catalog file argument required\n...")
        return 2
    }
    // Apply --no-color before any output
    if *noColorFlag {
        color.NoColor = true
    }
    ...
}
```

### Pattern 2: Tree Walker with Depth Limiting

**What:** Recursive function that mirrors `catalog.generateTreeStructure` but writes to stdout with optional ANSI color. Depth limiting is a `currentDepth int` counter against a `maxDepth int` limit.

**When to use:** `show` command tree rendering.

```go
// Source: mirrors logic in internal/catalog/service.go generateTreeStructure
func printTree(w io.Writer, item *models.CatalogItem, isLast bool, prefix string, currentDepth, maxDepth int) {
    connector := "├── "
    if isLast {
        connector = "└── "
    }

    name := filepath.Base(item.Name)
    if item.Type == "directory" {
        dirColor := color.New(color.FgBlue, color.Bold)
        name = dirColor.Sprint(name)
    }

    fmt.Fprintf(w, "%s%s%s\n", prefix, connector, name)

    if item.Type == "directory" && len(item.Contents) > 0 {
        if maxDepth >= 0 && currentDepth >= maxDepth {
            return // depth limit reached
        }
        newPrefix := prefix
        if isLast {
            newPrefix += "    "
        } else {
            newPrefix += "│   "
        }
        for i, child := range item.Contents {
            childIsLast := i == len(item.Contents)-1
            printTree(w, child, childIsLast, newPrefix, currentDepth+1, maxDepth)
        }
    }
}
```

**Root node:** The root item has `Name: "./"` and `Type: "directory"`. Print it unconditionally (it is the tree root, not a child), then recurse into its `Contents`.

### Pattern 3: `--no-color` Global Disable

**What:** `fatih/color.NoColor` is a package-level variable. Setting it to `true` before any color formatting call disables all color output from that point forward (within the process). `fatih/color` also checks `NO_COLOR` env var at init time — no additional env var handling is needed in application code.

**When to use:** Apply when `--no-color` flag is present, BEFORE any `color.New(...).Sprint(...)` calls.

```go
// Source: fatih/color docs — NoColor is package-level
import "github.com/fatih/color"

if *noColorFlag {
    color.NoColor = true
}
// From this point, color.New(color.FgBlue, color.Bold).Sprint(x) == x (no escape codes)
```

**Important:** `color.NoColor` defaults to `true` when stdout is not a TTY (when `go-isatty` returns false). This means color is automatically suppressed when piped — no explicit TTY check is needed in application code.

### Pattern 4: `open` HTML Path Derivation

**What:** The catalog JSON file and HTML file are co-located with the same basename. Given `foo.json`, the HTML is `foo.html`. `CatalogMetadata.HasHtml` confirms HTML existence at browse time, but `show` and `open` work with direct file paths, so derive HTML path via string replacement.

```go
// Source: search/service.go BrowseCatalogs pattern
htmlPath := strings.TrimSuffix(jsonPath, ".json") + ".html"
if _, err := os.Stat(htmlPath); err != nil {
    fmt.Fprintf(os.Stderr, "storcat open: HTML file not found: %s\n", htmlPath)
    return 1
}
if err := browser.OpenFile(htmlPath); err != nil {
    fmt.Fprintf(os.Stderr, "storcat open: failed to open browser: %v\n", err)
    return 1
}
return 0
```

### Anti-Patterns to Avoid

- **Hand-rolling platform detection for browser launch:** `pkg/browser` is already a dependency. Do not write `switch runtime.GOOS { case "darwin": exec.Command("open", ...) }`.
- **Calling `color.New()` inside the recursion hot path:** Create the `dirColor` object once outside the walker or at the start of `runShow`, not inside `printTree` per-call.
- **Forgetting the root node:** The JSON root item (`"./"`) must be printed as the first line of the tree. The HTML generator skips this (it's the `<h1>` title), but terminal output starts from the root.
- **Depth semantics ambiguity:** Depth 0 means root only (no children). Depth 1 means root + immediate children. -1 means unlimited. Document this in `--help` text.
- **Not handling `--depth 0` for `--json` mode:** `--json` outputs the full catalog JSON regardless of `--depth` — depth only applies to the tree display path.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-platform browser open | `os/exec` switch on `runtime.GOOS` | `pkg/browser.OpenFile` | Already in module; handles edge cases (no browser, WSL, etc.) |
| ANSI color with TTY detection | Manual escape codes + isatty check | `fatih/color` | Already in module; handles Windows VT mode, NO_COLOR, colorable stdout |
| JSON loading with v1/v2 compat | New JSON parser | `search.NewService().LoadCatalog(path)` | Already handles both formats; tested |

**Key insight:** Both browser launch and color have deceptively complex edge cases (Windows console vs. VT mode, WSL, NO_COLOR standard, Cygwin terminals). The libraries that handle them are already compiled into the binary via indirect deps.

## Common Pitfalls

### Pitfall 1: `color.NoColor` Thread Safety
**What goes wrong:** If tests run in parallel and one test sets `color.NoColor = true`, subsequent tests see color disabled.
**Why it happens:** `color.NoColor` is a package-level variable with no mutex.
**How to avoid:** In tests that set `color.NoColor`, save and restore the original value with `defer`. Alternatively, avoid setting the package-level flag in tests; instead test `--no-color` by verifying no ANSI escape sequences appear in captured stdout.
**Warning signs:** Test flakiness when running `go test -parallel N` in the `cli` package.

### Pitfall 2: `pkg/browser` Output Suppression
**What goes wrong:** `pkg/browser` may write diagnostic output to `browser.Stderr` (which defaults to `os.Stderr`). In tests this produces noise.
**Why it happens:** The library logs browser launch failures to its `Stderr` writer.
**How to avoid:** In `open` tests, don't actually invoke `browser.OpenFile`. Instead, test that the HTML path derivation logic is correct (file found / not found) and that the correct error codes are returned. The cross-platform launch itself is not unit-testable without a browser present.
**Warning signs:** Tests that call `runOpen` with a real catalog file failing on CI where no browser is available.

### Pitfall 3: Root Node Depth Counting
**What goes wrong:** Off-by-one in depth: passing `--depth 1` shows only the root and nothing else, or shows two levels when one was expected.
**Why it happens:** Ambiguity between "depth of the root" vs. "depth of children".
**How to avoid:** Define depth as "number of levels below root". Root is always shown. `--depth 1` shows root + its direct children. Implementation: start `currentDepth = 0` at root's children call, stop recursing when `currentDepth >= maxDepth` (and `maxDepth >= 0`).

### Pitfall 4: Non-`.json` Input to `show` and `open`
**What goes wrong:** User passes a directory or non-JSON file; `LoadCatalog` returns a JSON parse error that looks confusing.
**Why it happens:** No input validation before attempting to parse.
**How to avoid:** Validate that the path ends in `.json` before calling `LoadCatalog`. Return a clear error: `"storcat show: expected a .json catalog file"`.

### Pitfall 5: HTML File Missing for `open`
**What goes wrong:** User passes a `catalog.json` that has no corresponding `.html` file (e.g., manually created JSON, or HTML was deleted).
**Why it happens:** `open` derives HTML path by replacing `.json` with `.html` — the file may not exist.
**How to avoid:** Stat the HTML path before calling `browser.OpenFile`. Return exit code 1 with a clear message.

## Code Examples

Verified patterns from official sources:

### fatih/color: Conditional Bold/Blue
```go
// Source: fatih/color v1.18.0 — go doc github.com/fatih/color
import "github.com/fatih/color"

dirColor := color.New(color.FgBlue, color.Bold)
// When color.NoColor is true (NO_COLOR env, non-TTY, or --no-color flag), Sprint returns undecorated string
coloredName := dirColor.Sprint(name)
```

### pkg/browser: Open file in system browser
```go
// Source: pkg/browser — go doc github.com/pkg/browser
import "github.com/pkg/browser"

if err := browser.OpenFile(htmlPath); err != nil {
    return fmt.Errorf("failed to open browser: %w", err)
}
```

### search.LoadCatalog: Load and parse catalog JSON
```go
// Source: internal/search/service.go — LoadCatalog handles v1 and v2 formats
import "storcat-wails/internal/search"

svc := search.NewService()
root, err := svc.LoadCatalog(filePath)
if err != nil {
    fmt.Fprintf(os.Stderr, "storcat show: %v\n", err)
    return 1
}
```

### Test helper: captureOutput (existing, reuse)
```go
// Source: cli/cli_test.go — captureOutput already in test package
// All new tests in show_test.go and open_test.go use the same captureOutput helper.
// It redirects os.Stdout and os.Stderr via os.Pipe().
```

### stubs.go cleanup
```go
// After implementing show.go and open.go, delete cli/stubs.go entirely.
// The stub tests in cli_test.go (TestRun_StubShow, TestRun_StubOpen) must be
// replaced with real tests. The test names can be kept but their expectations change.
```

## Environment Availability

Step 2.6: SKIPPED (no external dependencies beyond Go toolchain; `fatih/color` and `pkg/browser` are already in the module cache).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` stdlib |
| Config file | none (standard `go test`) |
| Quick run command | `go test ./cli/... -run TestRunShow\|TestRunOpen -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CLCM-04 | `show <catalog.json>` prints tree to stdout | unit | `go test ./cli/... -run TestRunShow` | ❌ Wave 0 |
| CLOF-02 | `show --json` outputs raw catalog JSON | unit | `go test ./cli/... -run TestRunShow_JSON` | ❌ Wave 0 |
| CLOF-04 | `show --depth N` limits tree depth | unit | `go test ./cli/... -run TestRunShow_Depth` | ❌ Wave 0 |
| CLOF-05 | `show` colorizes directories on TTY | unit | `go test ./cli/... -run TestRunShow_Color` | ❌ Wave 0 |
| CLOF-06 | `--no-color` suppresses ANSI codes | unit | `go test ./cli/... -run TestRunShow_NoColor` | ❌ Wave 0 |
| CLOF-06 | `NO_COLOR` env var suppresses ANSI codes | unit | `go test ./cli/... -run TestRunShow_NOCOLOREnv` | ❌ Wave 0 |
| CLCM-05 | `open` finds HTML and invokes browser | unit | `go test ./cli/... -run TestRunOpen` | ❌ Wave 0 |
| CLPC-05 | `open` works cross-platform | manual | manual verification on macOS (dev), CI on Linux | N/A |

**Existing stub tests to update:** `TestRun_StubShow` and `TestRun_StubOpen` in `cli_test.go` must be updated once stubs are replaced with implementations (they currently assert `"not yet implemented"` in stderr).

### Sampling Rate
- **Per task commit:** `go test ./cli/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `cli/show_test.go` — covers CLCM-04, CLOF-02, CLOF-04, CLOF-05, CLOF-06
- [ ] `cli/open_test.go` — covers CLCM-05

## Sources

### Primary (HIGH confidence)
- `go doc github.com/fatih/color` — NoColor semantics, New/Sprint API, NO_COLOR auto-detection
- `go doc github.com/pkg/browser` — OpenFile, OpenURL API
- `/Users/ken/dev/storcat/go.mod` — confirmed both libraries present as indirect deps at specific versions
- `/Users/ken/dev/storcat/internal/catalog/service.go` — `generateTreeStructure` connector characters and logic
- `/Users/ken/dev/storcat/internal/search/service.go` — `LoadCatalog` v1/v2 compat handling
- `/Users/ken/dev/storcat/cli/list.go` — interspersed flag parsing pattern
- `/Users/ken/dev/storcat/cli/output.go` — `printJSON` helper
- `/Users/ken/dev/storcat/cli/stubs.go` — existing stub signatures to replace
- `/Users/ken/dev/storcat/pkg/models/catalog.go` — `CatalogItem` struct

### Secondary (MEDIUM confidence)
- NO_COLOR specification (nocolor.org) — `NO_COLOR` env var is checked at `fatih/color` package init; no value required, mere presence disables color

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — both libraries confirmed present in go.mod and inspected via `go doc`
- Architecture: HIGH — patterns derived directly from existing codebase code (list.go, search.go, catalog/service.go)
- Pitfalls: HIGH — color.NoColor thread-safety and browser test isolation are known Go testing concerns
- Test design: HIGH — existing test helpers (captureOutput) and patterns (flag testing, tempdir catalogs) are established in the test suite

**Research date:** 2026-03-26
**Valid until:** 2026-04-26 (stable libraries; go.mod is pinned)
