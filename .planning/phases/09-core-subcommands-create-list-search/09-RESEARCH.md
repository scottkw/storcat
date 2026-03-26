# Phase 9: Core Subcommands — Create, List, Search - Research

**Researched:** 2026-03-26
**Domain:** Go CLI subcommand implementation — table output, JSON output, stdlib flag.FlagSet, service layer integration
**Confidence:** HIGH

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CLCM-01 | User can run `storcat create <dir>` with `--title`, `--name`, `--output` flags to create a catalog | `catalog.Service.CreateCatalog()` is ready; CLI wrapper maps flags to service params — impedance mismatch on `--output` documented below |
| CLCM-02 | User can run `storcat search <term> <dir>` to search catalogs for a term | `search.Service.SearchCatalogs(term, dir)` is ready; CLI wrapper adds positional arg validation + table output |
| CLCM-03 | User can run `storcat list <dir>` to list catalogs with metadata | `search.Service.BrowseCatalogs(dir)` is ready; CLI wrapper adds table output |
| CLOF-01 | `list` and `search` commands support `--json` flag for machine-readable output | `encoding/json` stdlib; marshal slice to stdout — no new deps |
| CLOF-03 | `create` command supports `--json` flag for structured result output | `models.CreateCatalogResult` already has JSON tags; marshal to stdout |
</phase_requirements>

---

## Summary

Phase 9 replaces the five stub functions in `cli/stubs.go` with real implementations for `create`, `list`, and `search`. The business logic already exists — `internal/catalog` and `internal/search` are production-ready Go services with no Wails dependencies. Phase 9's work is entirely in the `cli/` package: argument validation, flag wiring, calling existing services, and formatting output.

There is one impedance mismatch to resolve: `catalog.Service.CreateCatalog()` always writes output files into `directoryPath` (the catalogued directory itself), using `outputRoot` as the filename stem. The CLI's `--output <dir>` implies writing to a different directory. The service's `copyToDirectory` parameter copies files to a secondary location, but doesn't set a primary output dir. The plan must decide: either (a) use `copyToDirectory` to write to `--output` and skip the copy-to-source pattern, or (b) modify the service to accept an explicit output directory. Option (a) avoids service changes; option (b) is cleaner for CLI semantics. This decision is left for the planner — both are viable with minimal risk.

The table output library (`github.com/olekukonko/tablewriter` v1.1.4) is Phase 8's identified solution. It is NOT yet in `go.mod` and must be added in Phase 9's Wave 0. JSON output uses `encoding/json` from stdlib — no new dependency.

**Primary recommendation:** Replace stubs in `cli/stubs.go` with real implementations. Add `tablewriter` as a direct dependency. Keep all output formatting logic inside `cli/` — services return data structs, CLI formats them. For `--output`, use `copyToDirectory` approach (no service changes needed).

---

## Project Constraints (from CLAUDE.md)

- Go: use `go fmt`, `golangci-lint`, context-aware where applicable
- `go mod` only — `pnpm` for JS; `go get` for Go packages
- No Cobra — stdlib `flag.FlagSet` per subcommand (locked in STATE.md)
- No Wails imports in `cli/` package (enforced by `TestRun_NoCobra` test in `cli_test.go`)
- Error handling: no silent fallbacks; errors surface to stderr with non-zero exit
- Testing: Go `testing` package; `captureOutput` helper already exists in `cli_test.go`
- GSD workflow enforcement: all file changes via GSD commands

---

## Standard Stack

### Core (Phase 9 — adds one new dependency)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `flag` (stdlib) | Go 1.23 built-in | Per-subcommand flag parsing | Locked decision — no Cobra |
| `encoding/json` (stdlib) | Go 1.23 built-in | `--json` output marshaling | stdlib; `models.*` structs already have json tags |
| `fmt`, `os`, `path/filepath` (stdlib) | Go 1.23 built-in | Output, args, path resolution | Already in use throughout |
| `github.com/olekukonko/tablewriter` | **v1.1.4** | Terminal table output for `list` and `search` | Identified in Phase 8 research; confirmed current via `go get` |

### Already Available (No Addition Needed)

| Library | Source | Purpose |
|---------|--------|---------|
| `storcat-wails/internal/catalog` | in-repo | `catalog.NewService()`, `CreateCatalog()` |
| `storcat-wails/internal/search` | in-repo | `search.NewService()`, `SearchCatalogs()`, `BrowseCatalogs()` |
| `storcat-wails/pkg/models` | in-repo | `CatalogMetadata`, `SearchResult`, `CreateCatalogResult` |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| tablewriter | `text/tabwriter` (stdlib) | tabwriter requires manual column alignment; tablewriter handles borders, padding, column wrapping automatically — worth ~40KB binary increase for readable CLI output |
| tablewriter | `github.com/jedib0t/go-pretty/table` | go-pretty is heavier (200KB+); tablewriter is the standard lightweight choice |
| tablewriter JSON output | Direct `json.Marshal` to stdout | JSON mode should bypass tablewriter entirely — just marshal and print |

**Installation (Wave 0):**
```bash
go get github.com/olekukonko/tablewriter@v1.1.4
go mod tidy
```

**Version verified:** `go get github.com/olekukonko/tablewriter@latest` resolves to v1.1.4 (confirmed 2026-03-26).

---

## Architecture Patterns

### File Layout — Phase 9 Changes

```
cli/
├── cli.go          # unchanged (dispatch switch)
├── cli_test.go     # add Phase 9 integration tests; keep existing
├── stubs.go        # REPLACED — runCreate, runSearch, runList move to dedicated files
├── create.go       # NEW — runCreate() implementation
├── list.go         # NEW — runList() implementation
├── search.go       # NEW — runSearch() implementation
├── output.go       # NEW — shared printTable(), printJSON() helpers
├── version.go      # unchanged
└── version_test.go # unchanged
```

The stubs for `runShow` and `runOpen` stay in `stubs.go` since those are Phase 10 work. After splitting, `stubs.go` shrinks to just those two functions.

### Pattern 1: Command Implementation Structure

Each command file follows the same pattern — flag setup, positional arg validation, service call, output dispatch:

```go
// Source: consistent with Phase 8 version.go pattern + flag.FlagSet docs
func runList(args []string) int {
    fs := flag.NewFlagSet("list", flag.ContinueOnError)
    jsonFlag := fs.Bool("json", false, "Output as JSON")
    fs.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: storcat list <directory> [flags]\n\n...")
    }
    if err := fs.Parse(args); err != nil {
        if err == flag.ErrHelp {
            return 0
        }
        return 2
    }
    if fs.NArg() < 1 {
        fmt.Fprintf(os.Stderr, "storcat list: directory argument required\n")
        return 2
    }
    dir, err := filepath.Abs(fs.Arg(0))
    if err != nil {
        fmt.Fprintf(os.Stderr, "storcat list: invalid path: %v\n", err)
        return 1
    }
    svc := search.NewService()
    catalogs, err := svc.BrowseCatalogs(dir)
    if err != nil {
        fmt.Fprintf(os.Stderr, "storcat list: %v\n", err)
        return 1
    }
    if *jsonFlag {
        return printJSON(catalogs)
    }
    return printListTable(catalogs)
}
```

### Pattern 2: Shared Output Helpers in output.go

```go
// printJSON marshals v to os.Stdout; returns exit code
func printJSON(v any) int {
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    if err := enc.Encode(v); err != nil {
        fmt.Fprintf(os.Stderr, "storcat: failed to encode JSON: %v\n", err)
        return 1
    }
    return 0
}
```

Table rendering uses `tablewriter.NewWriter(os.Stdout)` with `SetBorder(false)` and `SetColumnSeparator("  ")` for clean terminal output. Column headers and row data are set per command.

### Pattern 3: create Command Flag-to-Service Mapping

`catalog.Service.CreateCatalog(title, directoryPath, outputRoot, copyToDirectory, onProgress)` signature:

| CLI flag | Service param | Notes |
|----------|---------------|-------|
| positional `<dir>` | `directoryPath` | Must be `filepath.Abs()` first |
| `--title` | `title` | Default: `filepath.Base(dir)` |
| `--name` | `outputRoot` | Default: `filepath.Base(dir)` (filename stem without extension) |
| `--output` | **see note** | The service writes JSON/HTML into `directoryPath`, not an output dir |

**Impedance mismatch resolution:** The service currently writes output files into the catalogued directory. `--output` maps to `copyToDirectory` — files are written to source dir AND copied to `--output` dir. This is the least-invasive approach (no service changes). Alternatively, the service could be modified to accept an explicit primary output dir. The planner should choose; the research recommends modifying the service for clean CLI semantics (write only to `--output` when specified, not to source dir).

### Anti-Patterns to Avoid

- **Importing Wails in cli/**: `TestRun_NoCobra` test will catch `wailsapp` import; service layer is Wails-free so this is naturally safe
- **Instantiating App in cli/**: CLI commands call `catalog.NewService()` / `search.NewService()` directly, never through `App` struct (which requires Wails context)
- **fmt.Println for errors**: All errors go to `os.Stderr` via `fmt.Fprintf`; success output goes to `os.Stdout`
- **os.Exit in command functions**: `runXxx()` returns int exit code; `main.go` calls `os.Exit()`. Never call `os.Exit` inside `cli/` package functions (breaks test capture pattern)
- **Empty results as error**: `list` on a dir with no catalogs should print empty table (or "no catalogs found") and exit 0, not exit 1

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Terminal table formatting | Custom column-align string builder | `tablewriter` | Unicode width, multi-line cells, column separators, header borders — fiddly to get right |
| JSON pretty-print | `fmt.Sprintf` or manual indentation | `json.NewEncoder` + `SetIndent` | Handles escaping, large data, streaming |
| Path resolution | String manipulation | `filepath.Abs()` + `filepath.Base()` | Cross-platform separators, symlink resolution, `.` handling |
| Size formatting | Custom bytes-to-human | Reuse `catalog.Service.formatBytes()` — but it's unexported | Expose as package-level function OR duplicate in `cli/output.go` — planner must decide |

**Key insight on size formatting:** `catalog.Service.formatBytes()` is an unexported method. The CLI needs human-readable sizes for `list` table output. Options: (a) export as `catalog.FormatBytes(bytes int64) string`, (b) duplicate a simple version in `cli/output.go`. Option (a) is cleaner; option (b) avoids touching the service package. This is a planner decision.

---

## Common Pitfalls

### Pitfall 1: `create` Writes to Wrong Directory

**What goes wrong:** `catalog.Service.CreateCatalog()` writes JSON/HTML into `directoryPath` (the directory being catalogued), ignoring any `--output` flag the user passed.
**Why it happens:** The service was designed for GUI use where output dir == source dir. The `copyToDirectory` parameter was a secondary copy, not a primary output.
**How to avoid:** Either modify `CreateCatalog()` to accept an explicit `outputDir` parameter, or document that `--output` triggers a copy. The current service behavior will surprise users who expect `--output /tmp/catalogs` to put files only in `/tmp/catalogs`.
**Warning signs:** Integration test creates catalog with `--output /tmp/out` and finds file also in the source directory.

### Pitfall 2: `search` With Empty Results Returning Error

**What goes wrong:** `search.Service.SearchCatalogs()` returns `(nil, nil)` when no matches are found (not an error). CLI code that treats nil slice as error will exit 1 incorrectly.
**Why it happens:** Returning nil slice for "no results" is idiomatic Go; it's not an error.
**How to avoid:** Check `err != nil` for error; check `len(results) == 0` for "no results" message.

### Pitfall 3: `list` on Non-Existent Directory

**What goes wrong:** `BrowseCatalogs()` returns a wrapped error "failed to read catalog directory: ...". If the CLI just prints the error message and exits 1, it's correct — but if it panics or exits 2 (usage error), that's wrong.
**Why it happens:** Exit code 1 = runtime error (bad dir); exit code 2 = usage error (missing arg). A bad path is exit code 1, not 2.
**How to avoid:** Reserve exit code 2 only for missing/invalid arguments (before service call). All service errors are exit code 1.

### Pitfall 4: `create` Default Title/Name

**What goes wrong:** If `--title` and `--name` are omitted, the service receives empty strings, resulting in `.json` and `.html` output files (empty filename stem) and an empty `<title>` tag in HTML.
**Why it happens:** `catalog.Service.CreateCatalog()` takes whatever title/name it receives — no default logic.
**How to avoid:** CLI must default `title` to `filepath.Base(absDir)` and `outputRoot` to `filepath.Base(absDir)` when flags are omitted.

### Pitfall 5: tablewriter v1.x API Change

**What goes wrong:** tablewriter v1.x introduced breaking API changes from v0.x. Code copied from older tutorials may use removed methods.
**Why it happens:** v1.1.4 is the current stable release but some community examples still use v0.0.5.
**How to avoid:** Use tablewriter v1.x API from official docs. Key difference: `table.Rich()` and renderer API changed. For basic use (`Append`, `SetHeader`, `Render`), the core API is stable.

### Pitfall 6: Test Isolation — Service Calls Write Real Files

**What goes wrong:** Integration tests for `runCreate` call the real `catalog.Service.CreateCatalog()` which writes files to disk. Tests that don't clean up leave artifacts.
**Why it happens:** Phase 8 tests only tested stubs; Phase 9 tests call live service code.
**How to avoid:** Use `t.TempDir()` for all test directories — Go testing framework auto-cleans on test completion. Pass the temp dir as both source and destination.

---

## Code Examples

### tablewriter Basic Usage (list command pattern)

```go
// Source: github.com/olekukonko/tablewriter README + v1.1.4 API
import "github.com/olekukonko/tablewriter"

func printListTable(catalogs []*models.CatalogMetadata) int {
    if len(catalogs) == 0 {
        fmt.Fprintln(os.Stdout, "No catalogs found.")
        return 0
    }
    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"Name", "Files", "Size", "Modified"})
    table.SetBorder(false)
    for _, c := range catalogs {
        table.Append([]string{
            strings.TrimSuffix(c.Name, ".json"),
            // file count requires loading each catalog — see note below
            formatBytes(c.Size),
            c.Modified,
        })
    }
    table.Render()
    return 0
}
```

**Note on file count in `list`:** `CatalogMetadata.Size` is the JSON file size in bytes, not the file count in the catalogued directory. To show file count, the CLI must either (a) load each catalog JSON and count entries, or (b) add file count to `CatalogMetadata` during `BrowseCatalogs()`. Loading all catalogs for `list` could be slow on large directories. The planner should decide: skip file count in table, add lazy loading, or add it to `BrowseCatalogs()` return value.

### JSON Output Pattern

```go
// Source: encoding/json stdlib docs
func printJSON(v any) int {
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    if err := enc.Encode(v); err != nil {
        fmt.Fprintf(os.Stderr, "storcat: failed to encode JSON: %v\n", err)
        return 1
    }
    return 0
}
```

### create Command — Flag Mapping

```go
func runCreate(args []string) int {
    fs := flag.NewFlagSet("create", flag.ContinueOnError)
    title  := fs.String("title", "", "Catalog title (default: directory name)")
    name   := fs.String("name", "", "Output filename stem (default: directory name)")
    output := fs.String("output", "", "Output directory (default: source directory)")
    asJSON := fs.Bool("json", false, "Output result as JSON")
    fs.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: storcat create <directory> [flags]\n\n...")
    }
    if err := fs.Parse(args); err != nil {
        if err == flag.ErrHelp { return 0 }
        return 2
    }
    if fs.NArg() < 1 {
        fmt.Fprintf(os.Stderr, "storcat create: directory argument required\n")
        return 2
    }
    dir, err := filepath.Abs(fs.Arg(0))
    if err != nil {
        fmt.Fprintf(os.Stderr, "storcat create: %v\n", err)
        return 1
    }
    if *title == "" { *title = filepath.Base(dir) }
    if *name == "" { *name = filepath.Base(dir) }
    // ...
}
```

### Test Pattern Using Existing captureOutput Helper

```go
// captureOutput already defined in cli_test.go — reuse across new test files
func TestRunCreate_MissingArg(t *testing.T) {
    _, stderr := captureOutput(func() {
        code := cli.Run([]string{"create"}, "2.0.0")
        if code != 2 {
            t.Errorf("expected exit code 2, got %d", code)
        }
    })
    if !strings.Contains(stderr, "directory argument required") {
        t.Errorf("expected usage error in stderr, got %q", stderr)
    }
}

func TestRunCreate_Success(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
    stdout, _ := captureOutput(func() {
        code := cli.Run([]string{"create", dir}, "2.0.0")
        if code != 0 {
            t.Errorf("expected exit code 0, got %d", code)
        }
    })
    if !strings.Contains(stdout, "Created") {
        t.Errorf("expected success message, got %q", stdout)
    }
}
```

---

## Open Questions

1. **`create --output` semantics: copy vs. primary output**
   - What we know: `CreateCatalog()` writes to source dir; `copyToDirectory` copies to a second location
   - What's unclear: Should `--output` write ONLY to output dir (not source dir), or in addition to it?
   - Recommendation: Modify service to accept `outputDir` parameter (write primary output to `outputDir` when specified); this is cleaner and matches user expectation. Accept the small service change.

2. **File count in `storcat list` output**
   - What we know: `CatalogMetadata` contains file size (JSON file bytes) but NOT the count of catalogued files
   - What's unclear: Is file count worth adding given the cost of loading every catalog JSON?
   - Recommendation: Add `FileCount int` to `CatalogMetadata` and populate it in `BrowseCatalogs()` by loading each JSON. The search service already parses JSON for search — loading metadata is acceptable overhead for `list`.

3. **`storcat list <dir>` — cwd default (STATE.md pending todo)**
   - What we know: STATE.md explicitly notes "Confirm `storcat list` / `storcat search` default directory behavior (cwd vs config last-used) during Phase 9 planning"
   - What's unclear: Whether omitting `<dir>` should default to cwd or last-used config dir
   - Recommendation: Default to cwd (`filepath.Abs(".")`) when `<dir>` is omitted. Config last-used is hidden state (explicitly called out as "out of scope" in REQUIREMENTS.md). Keep it simple: cwd default, no config reading.

4. **`formatBytes` — export or duplicate**
   - What we know: `catalog.Service.formatBytes()` is unexported; `list` table needs human-readable sizes
   - What's unclear: Whether to touch the catalog package or add a duplicate in `cli/output.go`
   - Recommendation: Export as `catalog.FormatBytes(n int64) string` — it's a pure utility with no side effects; exporting is the right Go pattern.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing` stdlib, v1.23 |
| Config file | none — `go test ./...` |
| Quick run command | `go test ./cli/... -run TestRun -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CLCM-01 | `create <dir>` writes JSON+HTML, prints confirmation | integration | `go test ./cli/... -run TestRunCreate -v` | ❌ Wave 0 |
| CLCM-01 | `create --title --name --output` flags wired | integration | `go test ./cli/... -run TestRunCreate_Flags -v` | ❌ Wave 0 |
| CLCM-01 | `create` missing arg → exit 2 + usage to stderr | unit | `go test ./cli/... -run TestRunCreate_MissingArg -v` | ❌ Wave 0 |
| CLCM-02 | `search <term> <dir>` returns matching files in table | integration | `go test ./cli/... -run TestRunSearch -v` | ❌ Wave 0 |
| CLCM-02 | `search` missing args → exit 2 | unit | `go test ./cli/... -run TestRunSearch_MissingArgs -v` | ❌ Wave 0 |
| CLCM-03 | `list <dir>` shows catalog table | integration | `go test ./cli/... -run TestRunList -v` | ❌ Wave 0 |
| CLCM-03 | `list` missing arg defaults to cwd (or requires arg) | unit | `go test ./cli/... -run TestRunList_NoArg -v` | ❌ Wave 0 |
| CLOF-01 | `list --json` outputs valid JSON array | integration | `go test ./cli/... -run TestRunList_JSON -v` | ❌ Wave 0 |
| CLOF-01 | `search --json` outputs valid JSON array | integration | `go test ./cli/... -run TestRunSearch_JSON -v` | ❌ Wave 0 |
| CLOF-03 | `create --json` outputs valid JSON object | integration | `go test ./cli/... -run TestRunCreate_JSON -v` | ❌ Wave 0 |
| All | Error to stderr, exit non-zero | unit | `go test ./cli/... -run TestRun.*Error -v` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./cli/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `cli/create_test.go` — covers CLCM-01, CLOF-03
- [ ] `cli/list_test.go` — covers CLCM-03, CLOF-01 (list)
- [ ] `cli/search_test.go` — covers CLCM-02, CLOF-01 (search)
- [ ] `go get github.com/olekukonko/tablewriter@v1.1.4 && go mod tidy` — must precede any import

*(Existing `captureOutput` helper in `cli_test.go` is reusable across new test files — no new test infrastructure needed)*

---

## Environment Availability

Step 2.6: No external service dependencies. This phase is pure Go code changes — no databases, daemons, or CLI tools beyond the Go toolchain.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All | ✓ | 1.23 | — |
| `github.com/olekukonko/tablewriter` | list/search table output | ✗ (not in go.mod yet) | v1.1.4 available | none needed — `go get` in Wave 0 |

**Missing dependencies with no fallback:**
- `tablewriter` — not yet in `go.mod`; add with `go get github.com/olekukonko/tablewriter@v1.1.4` as first Wave 0 task.

---

## Sources

### Primary (HIGH confidence)
- `/Users/ken/dev/storcat/internal/catalog/service.go` — exact signature and behavior of `CreateCatalog()`
- `/Users/ken/dev/storcat/internal/search/service.go` — exact signature of `SearchCatalogs()`, `BrowseCatalogs()`
- `/Users/ken/dev/storcat/pkg/models/*.go` — field names and JSON tags on all result structs
- `/Users/ken/dev/storcat/cli/cli.go` + `stubs.go` + `cli_test.go` — existing dispatch structure, stub signatures, test helper pattern
- `/Users/ken/dev/storcat/.planning/phases/08-cli-foundation-and-platform-compatibility/08-RESEARCH.md` — Phase 8 locked decisions and identified Phase 9 deps
- `go get github.com/olekukonko/tablewriter@latest` — confirmed v1.1.4 is current (executed 2026-03-26)
- `/Users/ken/dev/storcat/.planning/STATE.md` — locked architectural decisions, pending todos resolved here
- `go test ./...` — confirmed all tests passing as of research date

### Secondary (MEDIUM confidence)
- tablewriter v1.x API: inferred from installation output + Go module docs pattern; verified library resolves to v1.1.4

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — tablewriter version confirmed by live `go get`; all other deps are stdlib or in-repo
- Architecture: HIGH — stubs, service signatures, and test helpers are code-verified
- Pitfalls: HIGH — impedance mismatch on `create --output` is a concrete code-level finding, not speculation
- Open questions: HIGH confidence in identifying them; resolution is planner decision

**Research date:** 2026-03-26
**Valid until:** 90 days (stable stack; no fast-moving dependencies)
