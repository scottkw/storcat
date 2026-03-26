# Feature Research

**Domain:** CLI subcommands for a directory catalog tool (StorCat v2.1.0)
**Researched:** 2026-03-26
**Confidence:** HIGH (CLI conventions are well-established; all business logic already exists in Go services)

---

## Context

This research covers CLI-specific features only. The GUI (v2.0.0) already ships all core functionality.
The v2.1.0 CLI layer exposes those same operations as subcommands of the unified `storcat` binary.

**Existing Go service layer** (`internal/catalog`, `internal/search`) provides all business logic.
The CLI needs thin command wrappers, output formatting, and POSIX conventions — not new business logic.

**Reference tools studied:** `tree`, `fd`, `eza`, `rg` (ripgrep), Cobra (Go CLI framework), CLIG (https://clig.dev/).

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels broken or non-scriptable.

| Feature | Why Expected | Complexity | Depends On |
|---------|--------------|------------|------------|
| `storcat create <dir>` — create catalog from directory | Core operation; without it the CLI has no value | LOW | `internal/catalog.Service.CreateCatalog` (existing) |
| `storcat search <term> <dir>` — search catalogs for a term | Core operation; mirrors GUI search tab | LOW | `internal/search.Service.SearchCatalogs` (existing) |
| `storcat list <dir>` — list catalogs with metadata | Users need to discover what's cataloged before using other commands | LOW | `internal/search.Service.BrowseCatalogs` (existing) |
| `storcat show <catalog.json>` — print catalog tree | Mirror of GUI tree view; enables inspection without launching GUI | MEDIUM | `internal/search.Service.LoadCatalog` (existing) + new tree printer |
| `storcat open <catalog.json>` — open HTML in default browser | Power-user shortcut matching the GUI "Open HTML" button | LOW | `os/exec` per platform + `GetCatalogHtmlPath` (existing) |
| `storcat version` — print version string to stdout | Expected by every CLI tool; needed for scripting/CI | LOW | `version.go` (existing) |
| `storcat` (no args) — launch GUI as today | Current behavior must not change; users who double-click the app must get GUI | LOW | No change to Wails startup path |
| Exit code 0 on success, non-zero on error | POSIX contract; scripts and CI break without it | LOW | Cobra default behavior |
| Errors to stderr, output to stdout | POSIX contract; piping breaks without this separation | LOW | `fmt.Fprintln(os.Stderr, ...)` pattern |
| `--help` / `-h` on all commands | Users expect self-documenting CLI; flag required for discoverability | LOW | Cobra built-in (automatic) |
| Human-readable default output on TTY | Tools like `eza`, `fd`, `gh` default to readable text when writing to a terminal | LOW | Detect TTY via `os.Stdout.Stat()` |
| `--json` flag on output commands (list, search, show) | Scripts and agents need structured output for piping to `jq` or other tools | LOW | `encoding/json` marshal of existing model structs |

### Differentiators (Competitive Advantage)

Features that make StorCat CLI stand out. Not required for initial launch but add real value.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| `storcat completion [bash\|zsh\|fish\|powershell]` | Tab completion for subcommands and flags; nearly free via Cobra | LOW | Cobra provides `completion` command generator out of the box |
| Auto-detect TTY: table format on terminal, JSON when piped | Eliminates the need to remember `--json` in scripts; matches `gh` CLI behavior | MEDIUM | `os.Stdout.Stat()` + `syscall.S_IFCHR`; worth doing for `list` and `search` |
| `--depth N` flag on `show` | Control tree verbosity; matches `tree -L N` and `eza --level N` conventions | LOW | Pass int to recursive tree printer; minimal code |
| `--no-color` flag + `NO_COLOR` env var | CI/CD compatibility; respects https://no-color.org standard | LOW | Only needed if colorized output is added; check env var first |
| `--output <dir>` flag on `create` | Scriptable output path control; mirrors GUI "output directory" field | LOW | Pass through to `CreateCatalog` `copyToDirectory` param |
| `--name <filename>` flag on `create` | Control output filename in scripts | LOW | Pass through `outputName` param (already a `CreateCatalog` arg) |
| `--title <string>` flag on `create` | Set catalog display name; mirrors GUI "Catalog Title" input | LOW | Pass through `title` param (already a `CreateCatalog` arg) |
| Colorized `show` tree output | Directories in bold/blue, files plain; matches `tree`/`eza` UX | MEDIUM | ANSI escape codes; must respect `--no-color` and non-TTY detection |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Interactive TUI mode (fuzzy finder, cursor nav) | "Like fzf or lazygit" | Scope explosion; conflicts with pipe-friendliness; separate tool concern | Pipe `storcat list` or `storcat search` output to `fzf` |
| Config file for CLI defaults (`~/.storcatrc`) | Avoid retyping `--json` or common directories | Adds hidden state; debugging pain; not needed at this scale | Shell aliases or shell functions |
| Watch mode / `--watch` on `create` | Auto-recatalog when directory changes | Requires `fsnotify`, daemon-like behavior; out of scope for v2.1.0 | `watchexec storcat create ...` or cron |
| Progress bar on `create` | Visual feedback for large directories | TTY-only; breaks in pipes; Go traversal is fast enough (~100ms for most dirs) | `--verbose` printing files as scanned; or nothing |
| Glob/regex filter on `list` | Show only catalogs matching a pattern | Pipe composability principle: `storcat list <dir> | grep pattern` is correct | Pipe to `grep`; don't embed filtering into `list` |
| YAML output format | Some users prefer YAML over JSON | JSON is the universal machine-readable format; YAML adds a dependency | `--json` + `yq` for YAML conversion |
| `storcat search` reading from stdin | Pipe catalog paths in | Ambiguous: are you piping a list of dirs, or content? No clear convention | Explicit positional arg `storcat search <term> <dir>` |
| Separate `storcat-cli` binary | Clean separation of GUI and CLI | Defeats the "unified binary" goal; complicates distribution and PATH | Detect subcommands before Wails startup in `main.go` |

---

## Feature Dependencies

```
storcat create <dir>
    requires  --> internal/catalog.Service.CreateCatalog (existing)
    requires  --> Cobra subcommand structure (new, shared across all commands)
    optional  --> --output <dir> flag
    optional  --> --name <filename> flag
    optional  --> --title <string> flag

storcat search <term> <dir>
    requires  --> internal/search.Service.SearchCatalogs (existing)
    requires  --> Cobra subcommand structure
    optional  --> --json flag
    optional  --> auto-TTY detection (enhances default output)

storcat list <dir>
    requires  --> internal/search.Service.BrowseCatalogs (existing)
    requires  --> Cobra subcommand structure
    optional  --> --json flag
    optional  --> auto-TTY detection

storcat show <catalog.json>
    requires  --> internal/search.Service.LoadCatalog (existing)
    requires  --> Cobra subcommand structure
    requires  --> tree printer function (new; ~30 LOC recursive)
    optional  --> --depth N flag
    optional  --> colorized output
    optional  --> --no-color flag (only needed if colorized output is added)

storcat open <catalog.json>
    requires  --> GetCatalogHtmlPath (existing in app.go)
    requires  --> cross-platform browser exec (new; os/exec open/xdg-open/start)
    requires  --> Cobra subcommand structure

storcat version
    requires  --> version string from version.go (existing)
    requires  --> Cobra subcommand structure

storcat completion [shell]
    requires  --> Cobra completion built-in (free; zero implementation work)
    enhances  --> all subcommands (tab-complete args and flags)

Cobra subcommand structure (CRITICAL PATH)
    required by  --> ALL CLI commands above
    requires     --> os.Args inspection before Wails startup in main.go
    note: main.go currently passes ALL args to Wails; CLI detection must
          intercept subcommands (len(os.Args) > 1 && os.Args[1] != "") before
          calling wails.Run()

--no-color / NO_COLOR env var
    enhances  --> storcat show (colorized tree)
    enhances  --> storcat list (table formatting)
    conflicts --> progress bar (both are TTY-only; simplify: omit progress bar entirely)
```

### Dependency Notes

- **Cobra subcommand structure is the critical path dependency for all CLI features.** The current `main.go` passes all args to Wails unconditionally. The CLI integration requires intercepting `os.Args` before calling `wails.Run()`. This is the single structural change that unlocks everything else.
- **`show` requires a tree printer.** `LoadCatalog` returns `*models.CatalogItem` (a recursive struct). A small `printTree(item, prefix, depth, maxDepth)` function (~30 LOC) is the only new business logic needed.
- **`open` requires cross-platform browser launch.** Wails's `runtime.BrowserOpenURL` is only available inside a Wails GUI context. CLI `open` needs `os/exec` with `open` (macOS), `xdg-open` (Linux), `start` (Windows) — standard Go pattern.
- **Auto-TTY detection enhances but does not block.** `list` and `search` can default to plain text and add TTY detection as a polish step. Do not block shipping on it.
- **All output commands should emit valid JSON with `--json`.** The existing model structs (`SearchResult`, `CatalogMetadata`, `CreateCatalogResult`) already have `json:` tags; JSON output is `json.Marshal` on the existing slice/struct — nearly free.

---

## MVP Definition

This is a subsequent milestone on a shipping product. "MVP" means the minimum that makes the CLI genuinely useful for scripting and power users.

### Launch With (v2.1.0)

- [ ] Cobra subcommand structure with `os.Args` dispatch before Wails — enables all other CLI features
- [ ] `storcat create <dir>` with `--output`, `--name`, `--title` flags — core catalog creation
- [ ] `storcat search <term> <dir>` with `--json` flag — scriptable search
- [ ] `storcat list <dir>` with `--json` flag — scriptable catalog discovery
- [ ] `storcat show <catalog.json>` with `--depth N` flag — tree visualization
- [ ] `storcat open <catalog.json>` — cross-platform browser launch
- [ ] `storcat version` — version string to stdout
- [ ] `storcat completion [bash|zsh|fish|powershell]` — shell completion (Cobra provides for free)
- [ ] Exit code 0 / non-zero, errors to stderr, output to stdout — POSIX contract
- [ ] `--help` on all commands — Cobra provides automatically

### Add After Validation (v2.1.x)

- [ ] Auto-TTY detection (table on terminal, JSON when piped) for `list` and `search` — once v2.1.0 ships and usage patterns are observed
- [ ] Colorized tree output in `show` — quality-of-life polish
- [ ] `--no-color` + `NO_COLOR` env var — only needed once colorized output is added

### Future Consideration (v2.2+)

- [ ] Watch mode (`--watch` on `create`) — requires `fsnotify`, daemon concerns, separate scope
- [ ] Interactive TUI mode — use `fzf` as an integration target; not storcat's responsibility
- [ ] `--stats` aggregate output — show totals across all matched results

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Cobra subcommand structure + os.Args dispatch | HIGH | LOW | P1 |
| `storcat create` | HIGH | LOW | P1 |
| `storcat search` | HIGH | LOW | P1 |
| `storcat list` | HIGH | LOW | P1 |
| `storcat show` + tree printer | HIGH | LOW | P1 |
| `storcat open` | MEDIUM | LOW | P1 |
| `storcat version` | HIGH | LOW | P1 |
| `--json` flag on list/search/show | HIGH | LOW | P1 |
| Exit codes + stderr/stdout split | HIGH | LOW (POSIX plumbing) | P1 |
| `storcat completion` | MEDIUM | LOW (Cobra built-in) | P1 |
| `--depth N` on show | MEDIUM | LOW | P2 |
| Auto-TTY detection (table vs JSON) | MEDIUM | MEDIUM | P2 |
| Colorized `show` tree output | LOW | MEDIUM | P3 |
| `--no-color` + `NO_COLOR` | LOW | LOW (only if P3 added) | P3 |

**Priority key:** P1 = must have for v2.1.0 launch; P2 = should have, add when P1 done; P3 = nice to have, v2.1.x+

---

## Competitor Feature Analysis

| Feature | `tree` | `fd` / `rg` | `eza` | StorCat CLI approach |
|---------|--------|-------------|-------|----------------------|
| Output format | ASCII tree; `-J` for JSON | Plain lines; `--json` | Color table; `--json` | Human table default; `--json` flag on all output commands |
| Color | Default on TTY; `-n` to disable | Default on TTY; `--no-color` | Default on TTY; `--no-color` | Default on TTY; `--no-color` + `NO_COLOR` env var |
| Exit codes | 0 success, 1+ error | 0 match, 1 no-match, 2 error | 0/1/2 | 0 success, 1 runtime error, 2 usage/invalid args |
| Shell completion | None | Built-in | Built-in | Via Cobra `completion` subcommand (bash/zsh/fish/pwsh) |
| Depth control | `-L N` | N/A | `--level N` | `--depth N` on `storcat show` |
| Piping | Plain text | Plain text | Plain text / JSON | JSON with `--json`; plain text default |
| Stats summary | `--du` disk usage | N/A | `--total-size` | `storcat create` always prints fileCount + totalSize |
| Help | `-h` / `--help` | `-h` / `--help` | `-h` / `--help` | Cobra auto-generates `-h` / `--help` per subcommand |

---

## Output Format Specifications

Concrete output shapes for each command — informs implementation directly.

### `storcat create` — human output (TTY default)
```
Created: My Project
  JSON:  /path/to/my-project.json
  HTML:  /path/to/my-project.html
  Files: 1,234
  Size:  45.2 MB
```

### `storcat create` — `--json` output (stdout)
Marshals existing `CreateCatalogResult` struct directly:
```json
{
  "jsonPath": "/path/to/my-project.json",
  "htmlPath": "/path/to/my-project.html",
  "fileCount": 1234,
  "totalSize": 47399936
}
```

### `storcat list <dir>` — human output (TTY default)
```
TITLE              FILENAME              MODIFIED
My Project         my-project.json       2026-03-25
Another Catalog    another.json          2026-03-20
```

### `storcat list <dir>` — `--json` output
Marshals `[]*CatalogMetadata` directly — all fields already have `json:` tags.

### `storcat search <term> <dir>` — human output (TTY default)
```
CATALOG          TYPE   PATH
my-project       file   src/main.go
my-project       file   src/app.go
another          dir    build/output
```

### `storcat search <term> <dir>` — `--json` output
Marshals `[]*SearchResult` directly — all fields already have `json:` tags.

### `storcat show <catalog.json>` — tree output (default)
```
My Project  (1,234 files, 45.2 MB)
├── src/
│   ├── main.go  (12 KB)
│   └── app.go   (8 KB)
└── README.md    (2 KB)
```

### `storcat version` — output
```
storcat 2.1.0
```

---

## Sources

- [Command Line Interface Guidelines (clig.dev)](https://clig.dev/) — exit codes, stdout/stderr separation, --json flag, color/TTY detection, piping conventions
- [Cobra Shell Completion docs](https://cobra.dev/docs/how-to-guides/shell-completion/) — built-in bash/zsh/fish/powershell completion generation
- [Modern Rust CLI Tools: eza, bat, fd, zoxide (32blog)](https://32blog.com/en/cli/cli-modern-rust-tools) — output format and flag conventions from reference tools
- [Building CLI Apps in Go with Cobra & Viper (2025)](https://www.glukhov.org/post/2025/11/go-cli-applications-with-cobra-and-viper/) — Go-specific persistent flag patterns
- [ripgrep manpage (Debian)](https://manpages.debian.org/testing/ripgrep/rg.1.en.html) — flag design patterns (--no-color, --json, --count)
- [Understanding stdin/stdout: Building CLI Tools Like a Pro](https://dev.to/sudiip__17/understanding-stdinstdout-building-cli-tools-like-a-pro-2njk) — pipe conventions and isatty behavior
- [tree manpage (mankier)](https://www.mankier.com/1/tree) — -J JSON flag, -L depth flag, tree output conventions
- [no-color.org](https://no-color.org/) — NO_COLOR env var standard

---

*Feature research for: StorCat v2.1.0 CLI subcommands*
*Researched: 2026-03-26*
