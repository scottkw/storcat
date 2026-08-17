# Phase 23: Rail + Virtualized Tree - Pattern Map

**Mapped:** 2026-08-13
**Files analyzed:** 16
**Analogs found:** 14 / 16

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/search/flatten.go` (NEW `LoadCatalogFlat`) | service | transform (DFS flatten) | `internal/search/service.go` (`LoadCatalog`, `BrowseCatalogs`) | exact (same package, same service) |
| `pkg/models/catalog.go` (add `FlatNode`, `FlatCatalog`, extend `CatalogMetadata`) | model | CRUD (struct definition) | `pkg/models/catalog.go` itself (`CatalogItem`, `CatalogMetadata`) | exact — additive to existing file |
| `internal/search/service.go` (`BrowseCatalogs` gains `parseError` via `json.Valid` fast path) | service | request-response + file-I/O | `internal/search/service.go:111-130` (`LoadCatalog`'s own dual-format parse + error wrapping) | exact |
| `internal/config/counts_cache.go` (NEW sidecar cache) | service/config | CRUD (cache read/mutate/save) with concurrent access | `internal/config/config.go` (`Manager`) | role-match, **with an explicit caveat — see Shared Patterns** |
| `internal/osutil/reveal_darwin.go` / `_windows.go` / `_linux.go` (NEW `RevealInFileManager`) | utility | event-driven (OS process spawn) | `app.go`'s `OpenExternal` (thin wrapper over `runtime.BrowserOpenURL`) | role-match (no existing `os/exec` analog in repo — new OS-integration surface) |
| `app.go` (register `LoadCatalogFlat`, `RevealInFileManager` bindings) | controller (Wails-bound methods) | request-response | `app.go`'s `LoadCatalog`, `BrowseCatalogs`, `OpenExternal`, `SelectDirectory` bindings | exact |
| `internal/search/flatten_test.go` (NEW) | test | — | `internal/search/service_test.go` | exact |
| `internal/config/counts_cache_test.go` (NEW) | test | — | `internal/search/service_test.go` (table-driven, `t.TempDir()`) — no existing `config` test file to copy from | role-match |
| `internal/osutil/reveal_test.go` (NEW, build-tag scoped) | test | — | `internal/search/service_test.go` | role-match |
| `frontend/src/contexts/AppContext.tsx` (extend: `currentCatalogId`, `expanded`, `selected`, `SELECT_CATALOG` action) | store/provider | event-driven (reducer) | `AppContext.tsx` itself (existing `SET_DENSITY`/`SET_RAIL_SIDE`/`SET_DETAIL_OVERLAY` pattern) | exact — extend in place |
| `frontend/src/components/workspace/CatalogRail.tsx` (fill) | component | CRUD (list/filter/select) | itself (Phase 22 skeleton) + `frontend/src/services/wailsAPI.ts` for the data call | exact (skeleton→fill) |
| `frontend/src/components/workspace/TreePane.tsx` (fill, virtualized) | component | streaming/CRUD (windowed render over flat array) | itself (Phase 22 skeleton) | exact (skeleton→fill); **no in-repo virtualization analog** |
| `frontend/src/components/workspace/DetailsPanel.tsx` (fill) | component | request-response (derived from in-memory state) | itself (Phase 22 skeleton, `DetailsPanelProps` already declared) | exact (skeleton→fill) |
| `frontend/src/components/workspace/StatusBar.tsx` (fill) | component | transform (derived/summed) | itself (Phase 22 skeleton) | exact (skeleton→fill) |
| `frontend/src/hooks/useVisibleRows.ts` (NEW, optional per RESEARCH structure) | hook | transform (memoized derivation) | `frontend/src/hooks/useMediaQuery.ts` | role-match (only existing hook in the codebase — different shape, but same file/naming/export convention) |
| `frontend/src/services/wailsAPI.ts` (extend: `loadCatalogFlat`, `revealInFileManager`) | service (IPC wrapper) | request-response | `wailsAPI.ts` itself (`loadCatalog`, `browseCatalogs`, `selectDirectory`) | exact — extend in place |

## Pattern Assignments

### `internal/search/flatten.go` (service, transform)

**Analog:** `internal/search/service.go`

**Imports pattern** (service.go lines 1-18, inferred from file header — package `search`, imports `encoding/json`, `fmt`, `os`, `path/filepath`, `strings`, `time`, `storcat-wails/pkg/models`, `github.com/djherbis/times`):
```go
package search

import (
	"path/filepath"

	"storcat-wails/pkg/models"
)
```

**Core pattern — reuse `LoadCatalog` unmodified, then DFS** (verified `internal/search/service.go:109-130`):
```go
// LoadCatalog reads and parses a catalog JSON file, returning the root CatalogItem.
// Supports both bare-object format (v2.0.0) and array-wrapped format (v1.0 bash script).
func (s *Service) LoadCatalog(filePath string) (*models.CatalogItem, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog file: %w", err)
	}
	// Try array format first (v1.0 bash script compatibility)
	var catalogArray []*models.CatalogItem
	if err := json.Unmarshal(data, &catalogArray); err == nil && len(catalogArray) > 0 {
		return catalogArray[0], nil
	}
	// Try bare object format (v2.0.0)
	var catalogObj models.CatalogItem
	if err := json.Unmarshal(data, &catalogObj); err != nil {
		return nil, fmt.Errorf("failed to parse catalog JSON: %w", err)
	}
	return &catalogObj, nil
}
```
`LoadCatalogFlat` calls this verbatim and then flattens — do not duplicate the dual-format parse logic. See RESEARCH.md Pattern 1 for the full flattener sketch (DFS `walk` closure producing `FlatNode{Name: filepath.Base(item.Name), Path: item.Name, ...}`).

**CRITICAL:** `CatalogItem.Name` is a full relative display path (`"./sub/dir/file.txt"`), NOT a basename — confirmed at `internal/catalog/service.go:86-90`. Every existing consumer computes the basename explicitly via `filepath.Base(item.Name)` (see `cli/show.go:107`, `internal/catalog/service.go:244`). `FlatNode.Name` must do the same split.

**Error handling pattern** (from `LoadCatalog`): wrap with `fmt.Errorf("...: %w", err)`, never a bare `panic` or silent nil.

**Existing count helpers to reuse for the sidecar cache fill** (`internal/catalog/service.go:307-337`):
```go
func (s *Service) countFiles(catalog *models.CatalogItem) int { ... }
func (s *Service) countDirectories(catalog *models.CatalogItem) int { ... }
func (s *Service) formatBytes(bytes int64) string { ... }        // "271M", "3.4M"
func (s *Service) formatBytesForDisplay(bytes int64) string { ... } // "[271M]"
```
These live on `catalog.Service`, not `search.Service` — either sum bytes/count during `LoadCatalogFlat`'s own DFS (free, per RESEARCH Pattern 4) or call these directly if cross-package access is wired up. Do not reimplement byte-formatting; `formatBytes` already produces the rail/status-bar display format.

---

### `internal/search/service.go` — `BrowseCatalogs` `parseError` extension (service, request-response)

**Analog:** `BrowseCatalogs` itself, `internal/search/service.go:133-201`

**Core pattern** (existing loop structure to extend, verified):
```go
func (s *Service) BrowseCatalogs(catalogDirectory string) ([]*models.CatalogMetadata, error) {
	var catalogs []*models.CatalogMetadata
	entries, err := os.ReadDir(catalogDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(catalogDirectory, entry.Name())
		// ... existing title/size/created/modified/hasHtml logic ...
		catalogs = append(catalogs, &models.CatalogMetadata{
			Title: title, Name: entry.Name(), Filename: entry.Name(),
			Size: info.Size(), Created: createdTime.Format(time.RFC3339),
			Modified: info.ModTime().Format(time.RFC3339),
			FilePath: filePath, HasHtml: hasHtml,
			// NEW: FileCount, TotalBytes (cache-backed), ParseError
		})
	}
	return catalogs, nil
}
```

**New parse-error fast-path** (RESEARCH.md Pattern 6, mirrors `LoadCatalog`'s own dual-format attempt so the `Parser` field can report which format was tried):
```go
data, err := os.ReadFile(filePath)
var parseErr string
if err != nil {
	parseErr = err.Error()
} else if !json.Valid(data) {
	var arr []*models.CatalogItem
	if uerr := json.Unmarshal(data, &arr); uerr != nil {
		if syn, ok := uerr.(*json.SyntaxError); ok {
			parseErr = fmt.Sprintf("byte %d: %s", syn.Offset, syn.Error())
		} else {
			parseErr = uerr.Error()
		}
	}
}
```
`json.Valid()` first avoids struct-allocation cost in the common (valid) case; only the rare broken catalog pays for a full `Unmarshal`.

---

### `pkg/models/catalog.go` (model, additive)

**Analog:** the file itself — existing `CatalogItem`/`CatalogMetadata`/`CreateCatalogResult` struct style (verified, full file read):
```go
package models

type CatalogItem struct {
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	Size     int64          `json:"size"`
	Contents []*CatalogItem `json:"contents"`
}

type CatalogMetadata struct {
	Title    string `json:"title"`
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
	FilePath string `json:"path"`
	HasHtml  bool   `json:"hasHtml"`
}
```
New fields append to `CatalogMetadata` in the same style (`FileCount int`, `TotalBytes int64`, `ParseError string`); new `FlatNode`/`FlatCatalog` structs follow the identical flat-field, no-nested-pointer, `json:"..."` tag convention. **Do not modify `CatalogItem`** — `cli/show.go` and the on-disk format depend on its exact current shape (COMPAT-01).

---

### `internal/config/counts_cache.go` (NEW sidecar cache — role-match with caveat)

**Analog:** `internal/config/config.go` (`Manager`) — directory resolution and load/save shape ONLY.

**Reusable directory-resolution pattern** (verified, `internal/config/config.go:39-58`):
```go
configDir, err := os.UserConfigDir()
if err != nil {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configDir = filepath.Join(homeDir, ".config")
}
storcatConfigDir := filepath.Join(configDir, "storcat")
if err := os.MkdirAll(storcatConfigDir, 0755); err != nil {
	return nil, err
}
```
Cache file: `filepath.Join(storcatConfigDir, "counts-cache.json")` — sibling to `config.json`.

**Load/Save shape to mirror** (verified, `internal/config/config.go:74-102`):
```go
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.config = DefaultConfig()
			return nil
		}
		return err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	m.config = &config
	return nil
}

func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.configPath, data, 0644)
}
```

**⚠️ CAVEAT — DO NOT COPY VERBATIM:** `config.Manager` has **no mutex anywhere** — confirmed by reading its full import block (`encoding/json`, `os`, `path/filepath` only, no `sync`). That is tolerable for `config.Manager` because its setters are only ever called from discrete, user-triggered UI actions that serialize naturally in practice. The sidecar count cache does **not** have that guarantee: a background fill goroutine (populating counts for multiple catalogs after `BrowseCatalogs` returns) can run concurrently with an opportunistic fill from `LoadCatalogFlat` when the user clicks a catalog mid-background-fill, and Wails does not guarantee bound methods are invoked serially from the frontend. **The counts-cache `Manager`-equivalent must add a `sync.Mutex` guarding its load-mutate-save cycle** — this is the one place in this phase where copying the nearest analog exactly would introduce a real concurrency bug. Corrupt/unreadable cache file must degrade to "every entry is a miss," never an error surfaced to the rail (never block `BrowseCatalogs`).

---

### `internal/osutil/reveal_{darwin,windows,linux}.go` (utility, event-driven — no in-repo analog)

**No existing `os/exec` usage anywhere in this codebase** — `OpenExternal` (`app.go`) uses `runtime.BrowserOpenURL`, not `os/exec`. This is genuinely new OS-integration surface; there is no closer analog than `OpenExternal`'s general shape (thin per-purpose function, error-wrapped, called directly from `app.go`).

**Existing `app.go` binding shape to mirror for error style:**
```go
// GetCatalogHtmlPath returns the HTML file path for a catalog
func (a *App) GetCatalogHtmlPath(catalogPath string) (string, error) {
	...
	if _, err := os.Stat(htmlPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("HTML file not found: %s", htmlPath)
		}
		return "", fmt.Errorf("cannot access HTML file: %w", err)
	}
	return htmlPath, nil
}
```

**New pattern (from RESEARCH.md Pattern 5) — argv-only, never a shell string:**
```go
// internal/osutil/reveal_darwin.go
func RevealInFileManager(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("path not accessible: %w", err)
	}
	return exec.Command("open", "-R", absPath).Run()
}
```
Windows: `exec.Command("explorer", "/select,"+absPath).Run()` (single argv element — unverified on real Windows, flag `checkpoint:human-verify`). Linux: `exec.Command("xdg-open", filepath.Dir(absPath)).Run()`. **Never** `exec.Command("sh", "-c", ...)` or string-concatenate into a shell command — argv-slice form is the standard, sufficient injection mitigation.

---

### `internal/*_test.go` (Go table-driven tests)

**Analog:** `internal/search/service_test.go` (full file read this session)

**Pattern — `t.TempDir()` fixture helper + individual `Test*` functions per behavior** (verified structure, lines 1-48):
```go
package search

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestCatalog(t *testing.T) (dir string, filePath string, fileSize int64) {
	t.Helper()
	dir = t.TempDir()
	content := []byte(`{"type":"directory","name":"./","size":0,"contents":[]}`)
	filePath = filepath.Join(dir, "test-catalog.json")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat test catalog: %v", err)
	}
	fileSize = info.Size()
	return dir, filePath, fileSize
}

func TestLoadCatalogArrayFormat(t *testing.T) {
	s := NewService()
	dir := t.TempDir()
	content := []byte(`[{"type":"directory","name":"root","size":100,"contents":[]}]`)
	filePath := filepath.Join(dir, "array-catalog.json")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test catalog: %v", err)
	}
	item, err := s.LoadCatalog(filePath)
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}
	if item.Name != "root" {
		t.Errorf("expected Name='root', got %q", item.Name)
	}
}
```
Not table-driven in the classic `[]struct{name,input,want}` sense in this specific file, but each `Test*` is self-contained with its own fixture — follow this shape for `TestLoadCatalogFlat_DualFormat`, `TestLoadCatalogFlat_Structure`, `TestBrowseCatalogs_ParseError`, `TestCountsCache_*`, `TestRevealInFileManager` (build-tag scoped per OS).

---

### `frontend/src/contexts/AppContext.tsx` (store/provider, extend in place)

**Analog:** the file itself (full file read, 64 lines)

**Existing reducer pattern to extend, verbatim:**
```typescript
export interface AppState {
  density: Density;
  railSide: RailSide;
  detailOverlay: boolean;
}

type AppAction =
  | { type: 'SET_DENSITY'; payload: Density }
  | { type: 'SET_RAIL_SIDE'; payload: RailSide }
  | { type: 'SET_DETAIL_OVERLAY'; payload: boolean };

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case 'SET_DENSITY':
      return { ...state, density: action.payload };
    case 'SET_RAIL_SIDE':
      return { ...state, railSide: action.payload };
    case 'SET_DETAIL_OVERLAY':
      return { ...state, detailOverlay: action.payload };
    default:
      return state;
  }
}
```
**⚠️ These three fields (`currentCatalogId`, `expanded`, `selected`) do NOT exist yet** — CONTEXT.md's phrase "existing AppContext reducer" refers to the reducer *pattern*, not these specific fields (RESEARCH.md Pitfall N1). Add them following the exact same `{...state, field: action.payload}` spread style, plus one atomic `SELECT_CATALOG` action that sets `currentCatalogId` AND clears `expanded`/`selected` in a single reducer case (TREE-06 — must not be two dispatches):
```typescript
case 'SELECT_CATALOG':
  return { ...state, currentCatalogId: action.payload, expanded: {}, selected: null };
```

---

### `frontend/src/services/wailsAPI.ts` (service/IPC wrapper, extend in place)

**Analog:** the file itself (full file read, 210 lines) — the `{success, ...}` envelope every binding must follow.

**Core pattern, verbatim, to copy for `loadCatalogFlat` and `revealInFileManager`:**
```typescript
loadCatalog: async (filePath: string) => {
  try {
    const catalog = await LoadCatalog(filePath);
    return { success: true, catalog };
  } catch (error: any) {
    return { success: false, error: error.message || 'Unknown error' };
  }
},

selectDirectory: async () => {
  try {
    const path = await SelectDirectory();
    return { success: true as const, path };
  } catch (error: any) {
    return { success: false as const, error: error.message || 'Unknown error' };
  }
},
```
Import the new generated bindings (`LoadCatalogFlat`, `RevealInFileManager`) from `'../../wailsjs/go/main/App'` alongside the existing ones (line 1-17 import block) — **never hand-edit `frontend/wailsjs/`**, it is regenerated by the Wails CLI after `app.go` changes.

---

### `frontend/src/components/workspace/{CatalogRail,TreePane,DetailsPanel,StatusBar}.tsx` (components, fill skeletons)

**Analog:** each file itself — Phase 22 built these as literal zero-prop, zero-state function components with correct dimensions/tokens/empty states (verified, all four read in full).

**Current shape (e.g. `CatalogRail.tsx:1`):**
```typescript
function CatalogRail() {
  return (
    <div className="ws-rail">
      {/* static empty-state markup, no props, no useAppContext() call */}
    </div>
  );
}
export default CatalogRail;
```
**`DetailsPanel.tsx` is the only one with a props interface today** (`DetailsPanelProps { variant?: 'pane' | 'drawer' }`) — the other three take zero props and are rendered with no props passed by `WorkspaceShell.tsx` (`<CatalogRail />`, `<TreePane />`). Per RESEARCH.md Pitfall N2, the established convention to extend is: call `useAppContext()` directly inside each component (matching `WorkspaceShell`'s own pattern), not prop-drilling — this phase should pick that convention rather than inventing new prop interfaces for `CatalogRail`/`TreePane`/`StatusBar`.

**Function-declaration + default-export-at-bottom convention** (per `CLAUDE.md` root conventions, confirmed by every file read this session) — do not switch to arrow-function components or named exports.

The existing empty-state markup blocks (STATE-01 treatment, the "Nothing catalogued yet" TreePane block, the rail's "No catalogs here yet" block) are the literal STATE-01 UI already built — this phase gates them behind `rail.length === 0` / no-directory-configured conditionals rather than rebuilding them (per 23-UI-SPEC.md's "Copywriting Contract" — STATE-01 copy is "unchanged verbatim from `22-UI-SPEC.md`").

---

### `frontend/src/hooks/useVisibleRows.ts` (hook, NEW)

**Analog:** `frontend/src/hooks/useMediaQuery.ts` — the only existing hook in the codebase (Phase 22's first hook).

**Convention to copy** (verified, full file, 27 lines): file name `camelCase.ts` matching the hook name, JSDoc-style block comment above the export explaining *why* the memoization/effect boundary exists (not just what it does), single default focus (one concern per hook), typed return value:
```typescript
import { useEffect, useState } from 'react';

/**
 * Subscribes to a media query via matchMedia's change event -- fires only on
 * threshold crossings, never on every pixel of a resize like a window
 * 'resize' listener would.
 */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() => window.matchMedia(query).matches);
  useEffect(() => {
    const mql = window.matchMedia(query);
    setMatches(mql.matches);
    const handleChange = (event: MediaQueryListEvent) => setMatches(event.matches);
    mql.addEventListener('change', handleChange);
    return () => mql.removeEventListener('change', handleChange);
  }, [query]);
  return matches;
}
```
For `useVisibleRows`/`computeVisible`, the equivalent shape is a `useMemo` (not `useEffect`) keyed on `[nodes, expanded]` — same "single focused concern, explanatory comment, typed signature" convention, different React primitive since this is a pure derivation, not a subscription. See RESEARCH.md Pattern 3 for the `computeVisible` O(n) algorithm itself.

---

## Shared Patterns

### `{success, ...}` envelope (all new Wails bindings)
**Source:** `frontend/src/services/wailsAPI.ts` (every method in the file)
**Apply to:** `loadCatalogFlat`, `revealInFileManager` wrappers.
Go side never throws across the IPC boundary implicitly — the Go method returns `(*T, error)`; the TS wrapper is solely responsible for the try/catch → `{success, error}` conversion. Go methods themselves should NOT pre-wrap into an envelope shape (unlike the frontend, Go bindings return plain `(*T, error)`, matching `LoadCatalog`/`BrowseCatalogs`/`GetCatalogHtmlPath` today).

### Dual-format catalog parsing (COMPAT-01)
**Source:** `internal/search/service.go:111-130` (`LoadCatalog`)
**Apply to:** `LoadCatalogFlat` (calls it verbatim, does not reimplement), `BrowseCatalogs`'s new parse-error fallback path (mirrors the same array-then-object attempt order so the `Parser` field in STATE-02 reports which format was tried).

### Go error wrapping
**Source:** `internal/search/service.go` and `app.go` throughout
**Apply to:** all new Go functions — `fmt.Errorf("...: %w", err)`, never a bare error string reconstruction, never `panic`.

### Directory-resolution + atomic-write discipline for local JSON files
**Source:** `internal/config/config.go:39-58` (directory resolution — safe to copy), Save() pattern (load-then-marshal-then-write — copy the SHAPE, not the concurrency posture)
**Apply to:** `internal/config/counts_cache.go` — **plus a `sync.Mutex` this analog does not have** (see caveat above). Corrupted/partial cache files degrade to "cache miss," never crash `BrowseCatalogs` — same "Silent Fallbacks" discipline flagged in `~/CLAUDE.md`'s Core Principles table is explicitly OVERRIDDEN here by the project's own security-domain note (a half-written cache is convenience data, not a hard failure to surface) — this is a deliberate, scoped exception, not blanket license to swallow errors elsewhere.

### Function-declaration components, default export at bottom
**Source:** every existing `frontend/src/components/**/*.tsx` file, and Root `CLAUDE.md`'s documented convention
**Apply to:** all four filled-in workspace components.

### `useAppContext()` called directly inside components (not prop-drilled)
**Source:** `frontend/src/components/workspace/WorkspaceShell.tsx` (per RESEARCH.md Pitfall N2's cited pattern)
**Apply to:** `CatalogRail.tsx`, `TreePane.tsx`, `StatusBar.tsx` (all three currently zero-prop).

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `@tanstack/react-virtual` usage inside `TreePane.tsx` | component (virtualization) | streaming/windowed render | No virtualization of any kind exists anywhere in this codebase — Phase 22's `TreePane` skeleton renders a single static empty-state block, not a list. This is genuinely new construction; RESEARCH.md's `useVirtualizer` sketch (Pattern 2) is a design sketch cross-referenced against actual `AppContext`/`themeTokens` state, not a copy-paste from an official doc (no authoritative-docs fetch tool was available during research) — treat it as a starting point to adapt, not a verified drop-in. |
| `internal/osutil/reveal_*.go` (`os/exec` usage) | utility | event-driven (OS process spawn) | No existing `os/exec` call anywhere in the Go codebase (confirmed by reading `app.go`, `internal/catalog/service.go`, `internal/search/service.go` in full) — `OpenExternal` uses Wails' own `runtime.BrowserOpenURL`, a materially different mechanism. RESEARCH.md Pattern 5 is the design sketch to follow. |

## Metadata

**Analog search scope:** `app.go`, `pkg/models/catalog.go`, `internal/search/service.go`, `internal/search/service_test.go`, `internal/catalog/service.go`, `internal/config/config.go`, `frontend/src/contexts/AppContext.tsx`, `frontend/src/services/wailsAPI.ts`, `frontend/src/hooks/useMediaQuery.ts`, `frontend/src/components/workspace/{CatalogRail,TreePane,DetailsPanel,StatusBar}.tsx`
**Files scanned:** 13 read in full + `internal/catalog/service.go`/`internal/search/service.go` function signatures grepped
**Pattern extraction date:** 2026-08-13
