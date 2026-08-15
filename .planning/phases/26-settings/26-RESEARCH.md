# Phase 26: Settings - Research

**Researched:** 2026-08-15
**Domain:** Go/Wails config persistence + React/TS settings UI (no new dependencies)
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Settings persistence — resolving the two-store split**
- The Go config (`internal/config`) becomes the single source of truth, with a one-time `localStorage` → config migration on first run. Phase 22 explicitly deferred this ("theme stays local pending Phase 26's Settings-owned theme state"); this is that resolution. Two stores that can silently diverge is the actual defect being fixed here, not just a tidiness preference.
- Every change writes immediately — no debounce. The config file is a few hundred bytes and `Manager.Save()` already exists. A timer would add a lost-write window on quit for no measurable benefit, and SET-05's "no explicit save step" is satisfied either way.
- New config fields: `density`, `railSide`, `catalogDirectory`, `defaultFilenameRoot`, `writeHTML`, `copyToSecondary`, `secondaryDirectory`, `watchDirectory`. `rememberWindow` is NOT new — it already exists as `windowPersistence` with working `Set`/`Get` methods, so COMPAT-05 wires the toggle to the existing field rather than reinventing it.
- Phase 25's `storcat-catalog-directory` and `storcat-secondary-directory` localStorage keys migrate in the same pass. Phase 25 bootstrapped its own directory persistence ahead of Settings existing (following the RAIL-05 precedent); folding them in here was always the plan.

**The two carried security obligations (deferred four times — must be discharged)**
- T-22-05 — DELETE `CatalogModal.tsx` and its `App.tsx` wiring, rather than sanitizing it. Verified during discuss: the only `dispatchEvent` in the whole frontend is for `themeChange`, so nothing dispatches `openCatalogModal` — the listener at `App.tsx:37` is registered but unreachable. The component still calls the legacy `window.electronAPI` shim and antd's `message` (antd was removed this milestone). `DetailsPanel.tsx:121-123` already provides the real "Open HTML catalog" path via `getCatalogHtmlPath` + `openExternal`. Deleting removes the threat rather than mitigating it; sanitizing would mean hand-rolling HTML sanitization, since no new dependency is permitted.
- FU-23-A — `GetCatalogHtmlPath` gets `catalogDir` threaded through and reuses `internal/osutil`'s `containsPath`, exactly the treatment `RevealInFileManager` received in Phase 23's review finding WR-02.
- FU-23-A — `OpenExternal` is restricted to `file://` paths contained in the catalog directory. Its only live caller opens a catalog's own HTML file. Rejecting everything else is simpler and safer than allow-listing URL schemes.
- `SearchIndexed` is removed from the FU-23-A sweep list, with the reason recorded. Phase 25's security audit verified it takes no renderer-supplied path reaching a filesystem write and received its own containment at introduction. STATE.md's target list should be corrected rather than carrying a redundant item forward.

**The Settings surface**
- A centred dialog at `--z-dialog` (300) — Phase 22's locked z-index scale places Settings there — consuming Phase 24's `useModalBehavior` for focus trap, Escape, scroll lock and focus restore. No bespoke reimplementation (PLT-07's standing constraint on Phases 25–27).
- 11 theme cards, each a 4-swatch strip plus a light/dark tag, rendered from the existing `THEMES` array, which PROJECT.md already records as authoritative.
- Theme applies immediately on click — SET-05 specifies no save step and `applyTokens()` already runs synchronously. No apply-on-close.
- Density and rail position use segmented controls, a new shared control introduced by this phase in the same way Phase 25 introduced the toggle switch.

### Claude's Discretion
- Config field names/JSON keys and the migration's exact detection mechanism.
- Settings dialog component decomposition and section ordering.
- Segmented-control markup and its ARIA pattern.
- Whether the theme chip opens Settings scrolled to the theme section or at the top.

### Deferred Ideas (OUT OF SCOPE)
- The fsnotify watcher itself — Phase 27 (WATCH-01/02/03). This phase persists the toggle only.
- Catalog rename/duplicate/delete — Phase 27.
- Re-scan and diff — Phase 28.
- Frontend unit tests for the Settings surface — TEST-01, deferred at v3.0.0 requirements definition.
- CLI subcommands for settings — new capabilities are GUI-only (FUT-03).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SET-01 | User can open Settings with ⌘,, the gear, or the theme chip | `WorkspaceShell.tsx`'s existing ⌘K listener is the template for the ⌘, listener; `Toolbar.tsx` gear/theme chip already render, need only an `onClick`. See Architecture Patterns → Dialog Mount & Entry Points. |
| SET-02 | User can pick a theme from 11 cards, each with a 4-swatch strip and light/dark tag | `themes.ts`'s 11-entry `themes` array verified read directly (id order, `tokens.{bg,p2,ac,tx}` fields exist). `themeChange` CustomEvent is the existing, load-bearing apply path. |
| SET-03 | User can set row density and catalog rail position from segmented controls | `AppContext.tsx` already has `SET_DENSITY`/`SET_RAIL_SIDE` reducer actions and `WorkspaceShell.tsx` already re-applies tokens reactively on `state.density`; `data-rail-side` attribute already reads `state.railSide`. New UI dispatches into existing reducer state — no new state shape needed. |
| SET-04 | User can set the catalog directory, default filename root, and four catalog toggles | `CatalogRail.tsx`'s `storcat-catalog-directory` and `OptionsToggles.tsx`'s `storcat-secondary-directory` are the two localStorage keys to fold into config, verified read directly. `wailsAPI.selectDirectory()` is the existing native-picker binding. |
| SET-05 | Settings save as changed, no explicit save step | `config.Manager.Save()` already exists and is synchronous-on-call; every existing `Set*` binding already writes-then-returns with no batching. Pattern to replicate for the eight new fields. |
| COMPAT-05 | Window size/position persistence continues to work, controlled by the Settings toggle | `WindowPersistenceEnabled`/`SetWindowPersistence`/`GetWindowPersistence` verified to already exist end-to-end (`internal/config/config.go:17,148-157`, `app.go:574-587`, consumed by `domReady`/`beforeClose`). Settings toggle is a UI surface over this unchanged field — do not add a second field. |
</phase_requirements>

## Summary

This phase has almost no external-library research surface — no new npm or Go dependency is introduced (confirmed by the UI-SPEC and cross-checked against `frontend/package.json`). The real work is entirely in-repo: (1) extending `internal/config.Config` with eight new fields and wiring `Set*`/`Get*` methods + `App` bindings for each, following the exact shape the four existing `Set*` methods already establish; (2) building a one-time `localStorage` → config migration for six keys; (3) building the Settings dialog UI from already-existing primitives (`useModalBehavior`, `ToggleRow`, the reducer's `SET_DENSITY`/`SET_RAIL_SIDE`); and (4) discharging two carried security findings — deleting `CatalogModal.tsx` and threading `catalogDir` containment through `GetCatalogHtmlPath`/`OpenExternal` using the exact `ContainsPath` helper `RevealInFileManager` already calls.

The single highest-risk technical detail this research surfaced, which CONTEXT.md's decisions do not address, is a **synchronous-boot-vs-async-config conflict**: `main.tsx` calls `initThemeTokens()` synchronously, before React's first render, specifically to avoid a themed-then-repainted launch flash (the code comment is explicit: "a post-mount effect fires after first paint and reintroduces the launch flash"). That function reads theme/density/rail-side from `localStorage` synchronously. Wails bindings (including a would-be `GetConfig()` call) are inherently asynchronous — there is no synchronous JS path to the Go config at module-load time. "Go config becomes the single source of truth" therefore cannot mean "delete the synchronous localStorage read" without reintroducing the flash THEME-06 already fixed. The resolution this research recommends (detailed in Common Pitfalls #1) is: keep `localStorage` as a write-through boot cache that every config-writing code path updates in the same synchronous tick as the Go config call, so the two stores cannot diverge (satisfying CONTEXT.md's actual concern) while the boot-time read stays synchronous and local.

A second, independently verified finding: the codebase already has an existing, unused `Config.SidebarPosition` field with a live but never-called `SetSidebarPosition` binding — a leftover that predates this milestone's `density`/`railSide` concept and is unrelated to it (`railSide` is a new value in CONTEXT.md's own field list, not a rename of `SidebarPosition`). The planner should decide explicitly whether to leave this dead field alone or remove it — it is out of this phase's stated scope either way, but silently adding a second, unrelated "position" field next to an already-unused one is worth a one-line acknowledgment rather than compounding confusion.

**Primary recommendation:** Extend `internal/config.Config` with the eight new fields and mechanical `Set*`/`Get*` pairs (mirroring the four that exist today exactly), thread `catalogDir` through `GetCatalogHtmlPath`/`OpenExternal` reusing `osutil.ContainsPath` verbatim as `RevealInFileManager` does, delete `CatalogModal.tsx` and its three `App.tsx` wiring points outright, and build the Settings dialog from `useModalBehavior` + the reducer's already-existing `SET_DENSITY`/`SET_RAIL_SIDE` actions + a new shared `ToggleRow`-alongside `SegmentedControl` component — writing every change through both `localStorage` (synchronous boot cache) and the Go config (durable source of truth) in the same call.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Settings dialog shell, entry points, focus/scroll behavior | Browser / Client (Frontend) | — | Pure React overlay state (`isOpen` boolean), reuses `useModalBehavior` unchanged |
| Theme/density/rail-side selection + immediate repaint | Browser / Client (Frontend) | API / Backend (persistence) | `applyTokens()` is a synchronous DOM/CSS-variable operation; the frontend is the only tier that can apply it without a flash |
| Catalog directory / filename root / toggle values | Browser / Client (Frontend, read+write) | API / Backend (persistence) | Frontend owns `AppContext.state.catalogDir` (already the case since Phase 23); Settings is a second UI surface over the same reducer field |
| Settings persistence (single source of truth) | API / Backend (Go, `internal/config`) | Database / Storage (`config.json` on disk) | `config.Manager` already owns disk I/O; extending it keeps the write path centralized, matching every existing `Set*` binding |
| Path containment / sanitization for `GetCatalogHtmlPath`/`OpenExternal` | API / Backend (Go, `internal/osutil`) | — | Containment must be enforced server-side (Go) since these bindings are callable from any renderer JS — client-side checks are not a security boundary |
| `localStorage` boot cache (theme/density/rail-side) | Browser / Client (Frontend) | — | Exists solely to satisfy the synchronous-read-before-first-paint requirement; not a second source of truth once every write path is unified |

## Standard Stack

No new dependency of any kind is introduced this phase — confirmed against `frontend/package.json` (no npm install needed) and `go.mod` (no new Go module needed; `internal/osutil` and `internal/config` are already in-repo). This phase is 100% composition of already-present primitives.

### Core (existing, reused)
| Component | Location | Purpose | Why reused, not rebuilt |
|-----------|----------|---------|--------------------------|
| `config.Manager` | `internal/config/config.go` | Load/Save/Get, four existing `Set*` methods | Adding 8 fields + `Set*`/`Get*` pairs is mechanical; a new persistence layer would duplicate this exactly |
| `osutil.ContainsPath` | `internal/osutil/reveal.go:93-110` [VERIFIED: internal/osutil/reveal.go:93] | Symlink-resolved, separator-safe containment check | Exported specifically for this reuse — its own doc comment says "Exported: it is now also the write-path containment gate `App.StartScan` uses ... not only the reveal read gate" |
| `useModalBehavior` | `frontend/src/hooks/useModalBehavior.ts` | Focus trap, Escape, scroll lock, focus restore | Explicitly written "for Phases 25, 26 and 27" per its own doc comment (line 9) |
| `ToggleRow` (from `OptionsToggles.tsx`) | `frontend/src/components/workspace/create/OptionsToggles.tsx:41-76` | Toggle-switch markup/geometry | UI-SPEC locks this as the shared component Settings must reuse, not a second toggle implementation |
| `SET_DENSITY`/`SET_RAIL_SIDE` reducer actions | `frontend/src/contexts/AppContext.tsx:49-50,127-130` | Already-wired density/rail-side state | Settings dispatches into existing state; `WorkspaceShell.tsx`'s reactive `useEffect` on `state.density` already re-applies tokens |
| `themeChange` CustomEvent | `frontend/src/App.tsx:20-27,29`, `DevStateSwitcher.tsx:33` | The one path that applies tokens, updates state, and persists theme | Settings should dispatch this same event rather than inventing a second theme-apply path — DevStateSwitcher's own comment confirms this is "the same 'themeChange' CustomEvent the future Settings surface will use" |

### Alternatives Considered
| Instead of | Could use | Tradeoff (why rejected) |
|------------|-----------|--------------------------|
| Reusing `themeChange` CustomEvent for theme apply | Lifting theme into `AppContext` reducer state (a `SET_THEME` action, matching density/rail-side's shape) | Cleaner long-term, but out of this phase's stated scope (CONTEXT.md's discretion list does not mention it) and touches more call sites (`App.tsx`, `DevStateSwitcher.tsx`) than necessary to satisfy SET-02. Flagged as an open question below, not decided here. |
| Sanitizing `CatalogModal.tsx`'s `srcDoc` | A hand-rolled HTML sanitizer, or an allowlist-based DOM sanitizer library | CONTEXT.md explicitly rejects both: no new dependency permitted, and hand-rolling sanitization is worse than deleting genuinely-dead code. Verified dead via `grep -rn dispatchEvent` (only `themeChange` found). |
| Native `<input>`-based ARIA radiogroup for segmented controls | `role="tablist"`/`role="tab"` | UI-SPEC already resolved this: a segmented control is a mutually-exclusive single-choice input, which a radio group models correctly; a tablist is for content panels, which nothing here has. |

**Installation:** none required.

**Version verification:** not applicable — no packages added. `antd` remains at `^5.27.4` in `frontend/package.json` [VERIFIED: frontend/package.json:14] — see Common Pitfalls #3 for why this matters.

## Package Legitimacy Audit

**Not applicable.** This phase installs no new npm, Go, or other ecosystem packages. Confirmed by reading `frontend/package.json` and `go.mod` directly, and by the UI-SPEC's own explicit statement: "No new npm dependency this phase."

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│ Frontend (React, WorkspaceShell.tsx)                             │
│                                                                    │
│  Entry points: ⌘, keydown | Toolbar gear onClick | theme chip    │
│  onClick  ──────────────────────────►  SettingsDialog isOpen=true│
│                                              │                     │
│                                              ▼                     │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ SettingsDialog (mounted always, gated by isOpen — same      │ │
│  │ pattern as CommandPalette)                                   │ │
│  │                                                                │
│  │  Theme card click ──► applyTokens() [sync, immediate paint]  │ │
│  │                   └─► dispatchEvent('themeChange')           │ │
│  │                   └─► wailsAPI.setTheme(id)     [async]  ─┐  │ │
│  │                                                             │  │
│  │  Segment click ──► dispatch(SET_DENSITY/SET_RAIL_SIDE)      │  │
│  │                └─► safeSetItem(localStorage)  [sync cache]  │  │
│  │                └─► wailsAPI.setDensity/setRailSide [async]─┤  │ │
│  │                                                             │  │
│  │  Directory/toggle change ──► dispatch(AppContext)            │  │
│  │                          └─► wailsAPI.set*         [async]─┤  │ │
│  └────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────┬───────────────────────┘
                                            │ Wails IPC (async, JSON)
                                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ Go backend (app.go bindings ──► internal/config.Manager)         │
│                                                                    │
│  SetTheme/SetDensity/SetRailSide/SetCatalogDirectory/            │
│  SetDefaultFilenameRoot/SetWriteHTML/SetCopyToSecondary/          │
│  SetSecondaryDirectory/SetWatchDirectory/SetWindowPersistence     │
│         │                                                          │
│         ▼                                                          │
│  Manager.Save() ──► config.json (single source of truth on disk) │
└─────────────────────────────────────────────────────────────────┘

Separately — the security sweep (no data flow relationship to the above):

  DetailsPanel "Open HTML catalog" click
        │
        ▼
  wailsAPI.getCatalogHtmlPath(catalog.path, catalogDir)  [NEW 2nd arg]
        │
        ▼
  App.GetCatalogHtmlPath(catalogPath, catalogDir)
        │  osutil.ContainsPath(catalogDir, resolved) gate  [NEW, reused]
        ▼
  App.OpenExternal(htmlPath, catalogDir)  [NEW 2nd arg, same gate]
        │
        ▼
  runtime.BrowserOpenURL — now only ever a file:// path inside catalogDir
```

### Recommended Project Structure
```
frontend/src/components/workspace/
├── SettingsDialog.tsx          # New — dialog shell, sections, entry-point wiring
├── settings/
│   ├── ThemeGrid.tsx           # New — 11 theme cards (optional decomposition)
│   ├── SegmentedControl.tsx    # New — shared control (density + rail-position rows)
│   └── CatalogSettingsSection.tsx  # New — directory, filename root, 4 toggles (optional)
├── create/OptionsToggles.tsx   # Existing — ToggleRow extracted/exported for reuse
├── Toolbar.tsx                 # Modified — gear/theme chip onClick wiring
└── WorkspaceShell.tsx          # Modified — mounts SettingsDialog, ⌘, listener, overlay mutual-exclusion

internal/config/
└── config.go                   # Modified — 8 new fields, Set*/Get* pairs, migration hook

app.go                          # Modified — new bindings, GetCatalogHtmlPath/OpenExternal containment
frontend/src/services/wailsAPI.ts  # Modified — new wrapper functions, updated call signatures
frontend/wailsjs/**              # Regenerated (wails dev auto-regenerates; restart if a new method doesn't appear)
```

### Pattern 1: Mechanical config field addition
**What:** Every existing config field follows an identical four-part shape: a struct field with a JSON tag, inclusion in `DefaultConfig()`, a `Set*` method that mutates-then-`Save()`s, and (where the frontend reads it back) a `Get*` method or reliance on `GetConfig()`'s full struct return.
**When to use:** For every one of the 8 new fields (`density`, `railSide`, `catalogDirectory`, `defaultFilenameRoot`, `writeHTML`, `copyToSecondary`, `secondaryDirectory`, `watchDirectory`).
**Example (verified pattern, not invented):**
```go
// Source: internal/config/config.go:122-132 (read this session)
// SetTheme updates theme setting
func (m *Manager) SetTheme(theme string) error {
	m.config.Theme = theme
	return m.Save()
}

// SetSidebarPosition updates sidebar position
func (m *Manager) SetSidebarPosition(position string) error {
	m.config.SidebarPosition = position
	return m.Save()
}
```
Every one of the 8 new setters is this same three-line shape. `App`'s bindings wrap them identically to `app.go:550-555`'s `SetTheme` binding — a nil-`configManager` guard, then delegate.

### Pattern 2: Containment-gated file binding (FU-23-A discharge)
**What:** `RevealInFileManager` already establishes the exact shape `GetCatalogHtmlPath`/`OpenExternal` must adopt: resolve the path, check containment via `osutil.ContainsPath`, reject before doing anything with the path.
**When to use:** Both `GetCatalogHtmlPath` and `OpenExternal` in `app.go`.
**Example:**
```go
// Source: internal/osutil/reveal.go:93-110 (read this session, verbatim)
func ContainsPath(catalogDir, resolved string) (bool, error) {
	absCatalogDir, err := filepath.Abs(catalogDir)
	if err != nil {
		return false, err
	}
	resolvedCatalogDir, err := filepath.EvalSymlinks(absCatalogDir)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(resolvedCatalogDir, resolved)
	if err != nil {
		return false, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}
```
```go
// Current signature, unguarded (app.go:675-689, read this session):
func (a *App) GetCatalogHtmlPath(catalogPath string) (string, error) { /* ... */ }
func (a *App) OpenExternal(url string) { /* ... */ }

// Target shape (mirrors RevealInFileManager's own call, app.go:703-705):
func (a *App) GetCatalogHtmlPath(catalogPath string, catalogDir string) (string, error) {
	// ... existing extension/Stat logic ...
	ok, err := osutil.ContainsPath(catalogDir, htmlPath) // or the resolved/abs form
	if err != nil || !ok {
		return "", fmt.Errorf("...: outside configured catalog directory")
	}
	return htmlPath, nil
}
```
The frontend call site (`DetailsPanel.tsx:117-128`, `Footer` component) already has `catalogDir` in scope as a prop (it's already threaded to `handleReveal` two functions below) — `handleOpenHtml` needs the same fail-closed check `handleReveal` already has (`if (!catalogDir) { setError(...); return; }`) before calling the two updated bindings.

### Pattern 3: Dialog mount matching the palette, not the slide-over
**What:** `WorkspaceShell.tsx` already mounts `CommandPalette` unconditionally and gates it with an `isOpen` boolean local state (`paletteOpen`), never conditionally rendering/unmounting it.
**When to use:** `SettingsDialog`.
**Example:**
```tsx
// Source: frontend/src/components/workspace/WorkspaceShell.tsx:104 (read this session)
<CommandPalette isOpen={paletteOpen} onClose={() => setPaletteOpen(false)} />
// SettingsDialog follows the identical shape:
// <SettingsDialog isOpen={settingsOpen} onClose={() => setSettingsOpen(false)} />
```
No `closing`-flag/animated-exit state is needed (unlike `CreateSlideOver`), per UI-SPEC's explicit call-out.

### Anti-Patterns to Avoid
- **Reading Go config synchronously at module load:** There is no such API in Wails v2 — `GetConfig()` is a Promise-returning binding. Do not attempt to replace `main.tsx`'s synchronous `initThemeTokens()` localStorage read with an awaited config call; this reintroduces the launch flash THEME-06 fixed. See Common Pitfalls #1.
- **Debouncing config writes:** CONTEXT.md explicitly rejects this ("Every change writes immediately — no debounce"). Do not add a `setTimeout`/`lodash.debounce`-style batching layer around any of the 8 new `Set*` calls.
- **A second toggle implementation:** UI-SPEC found the demo's own markup internally inconsistent (Settings toggles at `32×18px`/`14px` vs. the create-form's `30×17px`/`13px`) and locked the resolution as *one* shared `ToggleRow`. Do not build a visually-near-identical second component.
- **Sanitizing instead of deleting `CatalogModal.tsx`:** Explicitly rejected by CONTEXT.md — the component is unreachable dead code (verified: no `dispatchEvent('openCatalogModal', ...)` exists anywhere in the frontend), so sanitizing its `srcDoc` would be defending an execution path that cannot occur, at the cost of a hand-rolled sanitizer.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Focus trap / Escape / scroll lock for the dialog | A bespoke `useEffect` keydown+focus handler | `useModalBehavior` (`frontend/src/hooks/useModalBehavior.ts`) | Already handles the exact four behaviors, written explicitly to be reused by Phase 26; a second implementation risks diverging edge-case behavior (e.g., the Tab-key trap's `container.contains(active)` check) |
| Toggle switch | New CSS/markup for 4 Settings toggle rows | `ToggleRow` extracted from `OptionsToggles.tsx` | UI-SPEC found and resolved a real inconsistency in the design handoff itself by mandating exactly one shared component |
| HTML sanitization for `CatalogModal` | A hand-rolled tag/attribute allowlist sanitizer | Delete the component | No new dependency permitted, the code path is unreachable, and hand-rolled HTML sanitizers are a well-known source of bypass bugs — deleting the dead code removes the surface entirely rather than defending it |
| Path containment checks | A second `strings.HasPrefix`-based containment function for the two new bindings | `osutil.ContainsPath` | Already exported specifically for reuse beyond `RevealInFileManager`; a second implementation risks missing the symlink-resolution or separator-boundary edge cases this one already handles (its own doc comment explains exactly why `strings.HasPrefix` alone is wrong) |

**Key insight:** Every non-trivial primitive this phase needs — modal behavior, toggle geometry, containment checking, config persistence shape — already exists in this codebase from Phases 22–25. The engineering risk in this phase is almost entirely in *wiring these together correctly* (especially the sync-boot-vs-async-config tension below), not in building anything new.

## Runtime State Inventory

This phase performs a **data migration** — six `localStorage` keys folding into the Go config file — which triggers this section per the "any phase involving... migration" rule.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | Six `localStorage` keys, all verified present by direct grep + read this session: `storcat-theme-id` (`themeTokens.ts:6`), `storcat-density` (`themeTokens.ts:7`), `storcat-rail-side` (`themeTokens.ts:8`), `storcat-catalog-directory` (`CatalogRail.tsx:7`), `storcat-secondary-directory` (`OptionsToggles.tsx:7`). `storcat-dev-switcher` (`DevStateSwitcher.tsx:53`) is a DOM element `id`, not a storage key — it is not part of the migration. | Data migration: on first run after this phase ships, read each existing localStorage value (if present) and write it into the corresponding new Go config field, then leave localStorage in place as the boot cache (do not delete the keys — see Common Pitfalls #1) |
| Live service config | None — this app has no external service with UI-only configuration (no n8n/Datadog/Tailscale-style split store). | None |
| OS-registered state | None found. No Task Scheduler/pm2/launchd/systemd registration exists for this desktop app; window state is already handled entirely through `internal/config` (`WindowWidth/Height/X/Y`), which this phase does not touch. | None |
| Secrets/env vars | None — `config.json` holds no secret material, and no SOPS/env-var key references any of the fields this phase adds or renames. | None |
| Build artifacts / installed packages | `frontend/wailsjs/go/main/App.d.ts`/`App.js` are checked-in generated files that hardcode today's single-argument signatures for `GetCatalogHtmlPath(arg1)` and `OpenExternal(arg1)` [VERIFIED: frontend/wailsjs/go/main/App.d.ts:14,28]. These do not auto-update from a source change alone — they are regenerated by the Wails build tooling. | Code edit + regeneration: run `wails dev` (verified this session to auto-regenerate bindings; restart the dev server if a changed signature doesn't appear — [CITED: wails.io "Application Development" / community troubleshooting guidance, WebSearch this session]) or `wails build` before relying on the two-argument call sites compiling in the frontend. |

**The canonical question, answered:** After this phase's Go/frontend code ships, the six localStorage keys remain on disk (not deleted) but are no longer the value a fresh migration reads from on a second run — a `migrated` marker (Claude's Discretion: exact detection mechanism) prevents the migration from re-running and clobbering a value the user has since changed in Settings. No runtime system outside this app's own two stores (localStorage + `config.json`) holds any of this state.

## Common Pitfalls

### Pitfall 1: The synchronous-boot-flash requirement conflicts with "Go config is the single source of truth" read literally
**What goes wrong:** If theme/density/rail-side are read from `GetConfig()` (a Wails binding, inherently async) instead of `localStorage` at boot, `initThemeTokens()` can no longer run synchronously before `createRoot(...).render(...)`, and the app briefly paints with default tokens before repainting with the real theme — exactly the flash `main.tsx`'s own comment says was fixed by making this call synchronous.
**Why it happens:** "Single source of truth" is a true statement about *where a value durably lives and where the last-write-wins* — it does not mean "the only place a value may ever be read from." Wails v2 has no synchronous JS-to-Go call path.
**How to avoid:** Keep `localStorage` as a write-through boot cache: every code path that calls a new `wailsAPI.set*` config binding also calls `safeSetItem` on the matching `storcat-*` key, in the same synchronous handler (not after the async call resolves). Both writes happen in the same tick from the same event, so they cannot diverge under normal operation — this satisfies CONTEXT.md's actual concern (silent divergence) without breaking the synchronous boot read. `main.tsx`/`initThemeTokens()` keeps reading `localStorage` exactly as it does today.
**Warning signs:** A visible flash of the default (StorCat Light, Comfortable, Left) theme on launch before the user's real theme paints; a failing or newly-flaky manual check of THEME-06 ("theme, density, and rail position survive an app restart").

### Pitfall 2: `SidebarPosition`/`SetSidebarPosition` is dead code that predates and is unrelated to `railSide`
**What goes wrong:** A planner or executor skimming `config.go` may assume `SidebarPosition` is what `railSide` should extend or rename, and either silently repurpose it (breaking nothing today since it's unused, but conflating two different concepts) or leave a second, confusingly-named, still-unused field sitting beside the new `railSide` field with no explanation.
**Why it happens:** `Config.SidebarPosition` (`config.go:12`), `Manager.SetSidebarPosition` (`config.go:129-132`), `App.SetSidebarPosition` (`app.go:558-563`), and `wailsAPI.setSidebarPosition` (`wailsAPI.ts:218-224`) all exist and compile — but grep confirms **zero** call sites of `wailsAPI.setSidebarPosition` anywhere in `frontend/src/` outside its own definition, and no component reads `config.sidebarPosition`. It is leftover, never wired to this milestone's `Density`/`RailSide` concept (`themeTokens.ts:3-4`), which uses `'Left'`/`'Right'` values distinct from whatever `SidebarPosition` originally held.
**How to avoid:** Treat `railSide` as genuinely new, per CONTEXT.md's own field list. Decide explicitly (plan-time or Claude's-discretion) whether to leave `SidebarPosition` alone (out of scope, harmless) or remove it as dead code in the same pass — either is defensible, but the plan should say which, rather than silently doing neither.
**Warning signs:** A code reviewer finding two fields that both sound like "which side is X on."

### Pitfall 3: `CatalogModal.tsx`'s deletion does not remove antd from the codebase
**What goes wrong:** CONTEXT.md's stated rationale for why `CatalogModal.tsx` cannot easily be "fixed" instead of deleted is "antd's `message` (antd was removed this milestone)" — but `antd` is still a live dependency: `frontend/package.json:14` lists `"antd": "^5.27.4"` [VERIFIED: frontend/package.json:14], and `App.tsx:2,51-58` still imports and actively uses antd's `ConfigProvider` and `theme as antdTheme` for algorithm/token wiring [VERIFIED: frontend/src/App.tsx:2,51-58]. Deleting `CatalogModal.tsx` removes its two antd imports (`Modal`, `message`) but leaves `antd` in `package.json` and leaves `App.tsx`'s `ConfigProvider` wrapper untouched.
**Why it happens:** "antd was removed this milestone" is directionally true (most component usage was replaced across Phases 22–25) but not literally true of the whole codebase — it is easy to conflate "mostly gone" with "gone."
**How to avoid:** Do not treat this phase as an opportunity to fully remove `antd` — that is out of this phase's stated scope (SET-01 through SET-05, COMPAT-05) and would touch `App.tsx`'s `ConfigProvider`/theming wrapper, a change with its own blast radius. Simply note, in the plan or its summary, that `antd` remains a dependency after this phase and `CatalogModal.tsx`'s deletion only removes its two usages inside that one file.
**Warning signs:** A reviewer expecting `npm ls antd` to come back empty after this phase and being surprised it doesn't.

### Pitfall 4: The footer's literal "StorCat 3.0.0" copy does not match `GetVersion()`'s current return value
**What goes wrong:** UI-SPEC's Copywriting Contract locks the footer status line as `"StorCat 3.0.0 · settings save as you change them"` verbatim. `App.GetVersion()` already exists and returns `Version`, which is parsed from `wails.json`'s `info.productVersion` at build time [VERIFIED: version.go:12-22, wails.json:16]. `wails.json`'s `productVersion` is currently `"2.3.0"` [VERIFIED: wails.json:16], not `"3.0.0"` — and `frontend/package.json`'s own `"version"` field is `"2.0.0"` [VERIFIED: frontend/package.json:4]. If the footer calls `wailsAPI.getVersion()` dynamically (the more correct, DRY approach — reusing an existing binding rather than a hardcoded literal), it will currently render `"StorCat 2.3.0"`, not matching the UI-SPEC's locked copy, until a separate version-bump task lands.
**Why it happens:** The UI-SPEC's copy was written against the target v3.0.0 milestone name, not against the current build-time version string, and no phase in this roadmap visibly includes a `wails.json` version bump.
**How to avoid:** Flag this explicitly to the planner as a decision point: either (a) hardcode the literal string per the UI-SPEC's copy contract (simplest, matches the locked design, drifts from the real binding until a version bump happens elsewhere), or (b) call `wailsAPI.getVersion()` and accept that it will show `2.3.0` until `wails.json` is bumped, or (c) bump `wails.json`'s `productVersion` to `3.0.0` as part of this phase or flag it as a gap for `COMPAT-06`'s pre-ship sweep. This research does not resolve which — it is a genuine open question, not Claude's discretion per CONTEXT.md's list.
**Warning signs:** A UI review flagging the footer version string as wrong regardless of which approach is chosen, unless the discrepancy is understood and intentional.

### Pitfall 5: `handleOpenHtml` in `DetailsPanel.tsx` does not yet have the fail-closed `catalogDir` guard its sibling `handleReveal` already has
**What goes wrong:** After `GetCatalogHtmlPath`/`OpenExternal` gain a required `catalogDir` parameter, `handleOpenHtml` (`DetailsPanel.tsx:117-128`) must supply it — but unlike `handleReveal` (`DetailsPanel.tsx:130-147`), it currently has no early-return guard for a missing/empty `catalogDir`. Calling the binding with an empty string will correctly fail server-side (Go's `ContainsPath`/empty-dir checks reject it, following `RevealInFileManager`'s own `if catalogDir == "" { return fmt.Errorf(...) }` precedent), but the frontend will show a generic Go-error string rather than the same clear, purpose-written message `handleReveal` already has ("No catalog directory configured.").
**Why it happens:** `handleOpenHtml` was written before `catalogDir` was a required parameter anywhere in this file; `handleReveal`'s guard was added specifically for WR-02 in Phase 23, after `handleOpenHtml` already existed.
**How to avoid:** Add the identical fail-closed guard to `handleOpenHtml` that `handleReveal` already has, for the same reason (fail closed on the client before sending an empty directory the backend will reject anyway) and for UX consistency between the two buttons in the same `Footer` component.
**Warning signs:** A live-verification pass finding `handleOpenHtml`'s error message reads like a raw Go error string instead of the app's own house style.

## Code Examples

### Extending `internal/config.Config` (mechanical pattern, verified against current file)
```go
// Source: internal/config/config.go (read this session, current struct)
type Config struct {
	Theme                    string `json:"theme"`
	SidebarPosition          string `json:"sidebarPosition"` // existing, unused — see Pitfall 2
	WindowWidth              int    `json:"windowWidth"`
	WindowHeight             int    `json:"windowHeight"`
	WindowX                  int    `json:"windowX"`
	WindowY                  int    `json:"windowY"`
	WindowPersistenceEnabled bool   `json:"windowPersistenceEnabled"`
	// New fields this phase adds (exact JSON keys are Claude's Discretion
	// per CONTEXT.md — this is an illustrative, not locked, shape):
	// Density              string `json:"density"`
	// RailSide             string `json:"railSide"`
	// CatalogDirectory     string `json:"catalogDirectory"`
	// DefaultFilenameRoot  string `json:"defaultFilenameRoot"`
	// WriteHTML            bool   `json:"writeHTML"`
	// CopyToSecondary      bool   `json:"copyToSecondary"`
	// SecondaryDirectory   string `json:"secondaryDirectory"`
	// WatchDirectory       bool   `json:"watchDirectory"`
}
```

### `wailsAPI.ts` wrapper pattern to replicate for each new binding
```typescript
// Source: frontend/src/services/wailsAPI.ts:209-216 (read this session, verbatim)
setTheme: async (theme: string) => {
  try {
    await SetTheme(theme);
    return { success: true };
  } catch (error: any) {
    return wailsError(error);
  }
},
```
Every new `set*` wrapper (density, railSide, catalogDirectory, defaultFilenameRoot, writeHTML, copyToSecondary, secondaryDirectory, watchDirectory) follows this identical three-line try/catch shape, matching the project's established `extractErrorMessage()`/`wailsError()` convention (Wails rejects with a plain string, not an `Error` instance — `wailsError` already handles this).

### `RevealInFileManager` call-site pattern for the two updated bindings
```typescript
// Source: frontend/src/components/workspace/DetailsPanel.tsx:130-147 (read this session)
async function handleReveal() {
  if (revealBusy) return;
  setRevealBusy(true);
  setError(null);
  if (!catalogDir) {
    setError('No catalog directory configured.');
    setRevealBusy(false);
    return;
  }
  const result = await wailsAPI.revealInFileManager(catalog.path, catalogDir);
  if (!result.success) {
    setError(result.error);
  }
  setRevealBusy(false);
}
```
`handleOpenHtml` (same file, lines 117-128) needs the identical guard added, then both `wailsAPI.getCatalogHtmlPath(catalog.path, catalogDir)` and `wailsAPI.openExternal(htmlPathResult.htmlPath, catalogDir)` calls updated to pass the second argument.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| Theme/density/rail-side persisted only to `localStorage`; catalog directory persisted only to `localStorage` (Phase 23/25 bootstrap) | Go `config.json` as single source of truth, `localStorage` retained only as a synchronous boot cache | This phase | Resolves the two-store divergence risk Phase 22 explicitly deferred; requires care to preserve THEME-06's flash-free boot (see Pitfall 1) |
| `CatalogModal.tsx` + `openCatalogModal` CustomEvent (unreachable since Phase 23's real IPC landed) | Deleted; `DetailsPanel.tsx`'s "Open HTML catalog" button is the sole path to viewing a catalog's HTML | This phase | Removes an unsanitized `srcDoc` attack surface entirely rather than mitigating it |
| `GetCatalogHtmlPath`/`OpenExternal` reachable from any renderer JS with no containment check | Both gated by `osutil.ContainsPath`, matching `RevealInFileManager`'s Phase 23 WR-02 fix | This phase | Closes FU-23-A, the last of the two carried findings |

**Deprecated/outdated:** `Config.SidebarPosition`/`SetSidebarPosition` — dead since before this milestone's `railSide` concept existed; not touched by this phase but worth a documented decision (Pitfall 2).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|----------------|
| A1 | `wails dev` auto-regenerates `frontend/wailsjs/go/main/App.{d.ts,js}` when a Go method's signature changes, restarting the dev server if a change doesn't appear | Runtime State Inventory, Anti-Patterns | If regeneration silently fails, `tsc --noEmit` will show stale two-vs-one-argument mismatches; low risk since this is a compile-time-catchable failure, not a silent one |
| A2 | Exact config JSON field names for the 8 new fields (illustrative names shown in Code Examples are not locked) | Code Examples | None — CONTEXT.md explicitly marks field-name choice as Claude's Discretion; any reasonable, consistent naming satisfies the requirement |
| A3 | The migration's "already migrated" detection mechanism is unspecified (a version marker field, presence-of-new-field-implies-migrated, or an explicit `migrated: true` flag) | Runtime State Inventory | If unspecified and a naive "field is zero-value implies not-yet-migrated" check is used, a user who deliberately sets a field back to its zero value (e.g., clears the filename root) could have it look "unmigrated" and be silently overwritten by a stale localStorage value on next launch — worth an explicit marker field to avoid this |

**If this table is empty:** N/A — see rows above. Every other factual claim in this document was verified this session via direct `Read`/`Bash grep` against the live source tree, not training-data recall.

## Open Questions

1. **Should theme become `AppContext` reducer state, or stay as `App.tsx`-local state + `themeChange` CustomEvent?**
   - What we know: Density/rail-side are already reducer state (`SET_DENSITY`/`SET_RAIL_SIDE`); theme is not — it lives in `App.tsx`'s local `useState`, applied via a `themeChange` window CustomEvent that `DevStateSwitcher.tsx`'s own comment says "the future Settings surface will use."
   - What's unclear: Whether "will use" means "dispatch the same event" (minimal change, matches the Alternatives Considered table above) or whether Settings should be the trigger to finally lift theme into the reducer (bigger change, touches `App.tsx` and `DevStateSwitcher.tsx`).
   - Recommendation: Reuse the existing `themeChange` event (minimal, matches CONTEXT.md's "Theme applies immediately on click ... same code path themeTokens.ts already exposes" language) unless the planner has a specific reason to do the larger refactor. Not blocking — either satisfies SET-02.

2. **Footer version string: hardcode "3.0.0" or call `GetVersion()` (currently returns "2.3.0")?**
   - See Pitfall 4 in full. Not resolved by this research — genuinely underspecified between CONTEXT.md and the current build-time version value.

3. **Does `SidebarPosition`/`SetSidebarPosition` get removed, left alone, or documented as intentionally-orphaned in this phase?**
   - See Pitfall 2. Low-stakes either way, but should be a stated decision rather than an accidental omission.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|-------------|-----------|---------|----------|
| Go toolchain | `internal/config`/`app.go`/`internal/osutil` changes | ✓ | go1.26.6 darwin/arm64 (module requires go 1.23) | — |
| Node.js | Frontend build | ✓ | v24.14.1 | — |
| npm | Frontend build/regeneration | ✓ | 11.18.0 | — |
| Wails CLI | `wails dev`/`wails build`, binding regeneration | ✓ | v2.10.2 | — |

No new external dependency is required by this phase; all four tools were already exercised successfully by Phases 22–25 per STATE.md's own per-plan history.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework (Go) | stdlib `testing`, table-driven, files verified present: `internal/config/config_test.go`, `internal/osutil/reveal_test.go` |
| Framework (frontend) | None — TEST-01 explicitly deferred (v2 requirement). Proof is `tsc --noEmit` + `vite build` + live `dev-browser` against `:34115`, per STATE.md's standing convention for Phases 22–25 |
| Config file | none (stdlib `go test ./...`, no config needed) |
| Quick run command | `go test ./internal/config/... ./internal/osutil/... -race -count=1` |
| Full suite command | `go build ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build)` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|---------------------|--------------|
| SET-01 | ⌘,/gear/theme chip open Settings | manual-only (live dev-browser) | dev-browser click + keyboard event against `:34115` | N/A — no frontend test framework |
| SET-02 | 11 theme cards apply immediately | manual-only + Go-side `TestSetTheme_*` for persistence | `go test ./internal/config/... -run TestSetTheme` | ❌ Wave 0 (new test) |
| SET-03 | Density/rail-position segmented controls | manual-only + Go-side persistence test | `go test ./internal/config/... -run TestSetDensity|TestSetRailSide` | ❌ Wave 0 |
| SET-04 | Catalog directory, filename root, 4 toggles | manual-only + Go-side persistence tests | `go test ./internal/config/... -run TestSetCatalog|TestSetWriteHTML|TestSetCopyToSecondary|TestSetWatchDirectory` | ❌ Wave 0 |
| SET-05 | Save-as-you-change, no debounce | unit (Go): assert `Save()` called synchronously by each `Set*` | `go test ./internal/config/...` | ❌ Wave 0 |
| COMPAT-05 | Window persistence toggle continuity | unit (Go, existing) + manual restart check | `go test ./internal/config/... -run TestWindowPersistence` | Existing coverage likely present — verify at plan time |
| T-22-05 | `CatalogModal.tsx` deleted, no dead listener | unit/static: `grep -rn openCatalogModal frontend/src` returns nothing | shell grep, not a test file | N/A |
| FU-23-A | `GetCatalogHtmlPath`/`OpenExternal` containment | unit (Go): table-driven, mirroring `reveal_test.go`'s existing containment cases | `go test ./internal/... -run TestGetCatalogHtmlPath|TestOpenExternal` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go build ./... && go test ./internal/config/... ./internal/osutil/... -race -count=1` (fast, scoped)
- **Per wave merge:** Full suite command above, plus a live dev-browser pass against `:34115` (per STATE.md's standing note: `curl` liveness is not binding freshness — verify `Object.keys(window.go.main.App)` includes the new/changed methods before recording evidence)
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] New table-driven cases in `internal/config/config_test.go` for the 8 new `Set*`/`Get*` pairs, following `newTestManager(t)`'s existing temp-dir fixture pattern (`config_test.go:11-30`, read this session)
- [ ] New table-driven cases in `internal/osutil` (new file or extend `reveal_test.go`) for `GetCatalogHtmlPath`/`OpenExternal`'s containment behavior, following `reveal_test.go`'s existing structure for `ContainsPath`
- [ ] No new frontend test file — none needed, consistent with TEST-01's deferral and every prior phase's precedent

## Security Domain

`security_enforcement` is not set to `false` in `.planning/config.json` (absent key defaults to enabled), and this phase carries two explicitly-named carried security findings, so this section is mandatory, not optional.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|----------------|---------|--------------------|
| V1 Architecture | yes | Config as single source of truth reduces the two-store-divergence class of bug this phase exists partly to fix |
| V4 Access Control | yes | `osutil.ContainsPath` gate on `GetCatalogHtmlPath`/`OpenExternal`, matching the existing `RevealInFileManager` pattern |
| V5 Input Validation | yes | Config field values (density/railSide are closed enums; directory paths go through `filepath.Abs`/`EvalSymlinks` before use, matching the existing `SearchCatalogs`/`BrowseCatalogs` pattern) |
| V12 File & Resources | yes | Both `GetCatalogHtmlPath` and `OpenExternal` are file-path-accepting bindings reachable from any renderer JS — the exact class of surface FU-23-A addresses |
| V2/V3 Authentication/Session | no | Single-user local desktop app, no session concept (consistent with `25-SECURITY.md`'s T-25-07 acceptance rationale) |
| V6 Cryptography | no | No cryptographic operation in this phase's scope |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|------------------------|
| Unsanitized `iframe srcDoc` from attacker-influenceable HTML | Tampering / XSS | Delete the unreachable component (T-22-05) rather than sanitize — verified dead via `dispatchEvent` grep |
| Arbitrary-path file-open binding reachable from any renderer JS | Elevation of Privilege / Tampering | `osutil.ContainsPath` containment gate before any Stat/open/exec, exactly the pattern `RevealInFileManager` already established for the same threat class (T-23-01/T-23-02) |
| `../` / symlink escape past a configured directory boundary | Tampering | `filepath.EvalSymlinks` before `filepath.Rel`-based containment check (not `strings.HasPrefix`, which a same-prefix sibling directory name would defeat) — already implemented in `ContainsPath`, reused not reinvented |
| Config-write race under concurrent Settings interaction | Tampering (low severity, single-user desktop) | Each `Set*` call is independently synchronous (`mutate field, then Save()`); no concurrent-write guard exists today and none is added by this phase — acceptable for a single-window, single-user desktop app where Settings dialog interactions are inherently serialized by the UI thread |

## Sources

### Primary (HIGH confidence — read directly this session)
- `internal/config/config.go` — full file read; `Config` struct, `Manager`, all four existing `Set*`/`Get*` methods
- `internal/osutil/reveal.go` — full file read; `ContainsPath` (exported, capitalized — not `containsPath` as CONTEXT.md's prose spells it), `RevealInFileManager`
- `app.go` — bindings list via grep + targeted reads of `GetCatalogHtmlPath`, `OpenExternal`, `RevealInFileManager`, `GetConfig`/`SetTheme`/`SetSidebarPosition`/`SetWindowSize`/`GetWindowPersistence`/`SetWindowPersistence`/`SetWindowPosition`, `domReady`, `beforeClose`
- `frontend/src/App.tsx` — full file read; confirmed antd `ConfigProvider` usage, `openCatalogModal` listener, `themeChange` listener
- `frontend/src/components/CatalogModal.tsx` — full file read; confirmed antd `Modal`/`message` imports, `window.electronAPI` legacy shim usage
- `frontend/src/contexts/AppContext.tsx` — full file read; `SET_DENSITY`/`SET_RAIL_SIDE` reducer actions, state shape
- `frontend/src/themeTokens.ts` — full file read; `initThemeTokens()`/`applyTokens()`/`readPersistedPrefs()`, the synchronous-boot-read mechanism
- `frontend/src/main.tsx` — full file read; confirmed synchronous `initThemeTokens()` call before `createRoot(...).render(...)`
- `frontend/src/hooks/useModalBehavior.ts` — full file read
- `frontend/src/components/workspace/create/OptionsToggles.tsx` — full file read; `ToggleRow` component
- `frontend/src/components/workspace/CatalogRail.tsx` (partial, lines 1-70) — `storcat-catalog-directory` persistence pattern
- `frontend/src/components/workspace/DetailsPanel.tsx` (targeted reads, lines 95-170) — `handleOpenHtml`/`handleReveal`, `catalogDir` prop threading
- `frontend/src/components/workspace/WorkspaceShell.tsx` (lines 1-90, plus grep) — ⌘K listener pattern, `CommandPalette` mount pattern, density re-apply effect
- `frontend/src/components/dev/DevStateSwitcher.tsx` — full file read; confirmed `themeChange` dispatch and reducer-dispatch pattern for density/rail-side
- `frontend/src/themes.ts` (lines 1-105 + grep for all `id:` entries) — confirmed 11-theme array, order, `ThemeTokens` shape (`bg`, `p`, `p2`, `ch`, `l`, `tx`, `ac`)
- `frontend/wailsjs/go/main/App.d.ts`/`App.js` (grep) — confirmed current single-argument generated signatures for `GetCatalogHtmlPath`/`OpenExternal`
- `frontend/package.json` (grep) — confirmed `antd ^5.27.4` still present, `"version": "2.0.0"`
- `wails.json` — full file read; confirmed `productVersion: "2.3.0"`
- `version.go` — full file read; confirmed `GetVersion()`'s source
- `internal/config/config_test.go` (lines 1-40) — existing table-driven test fixture pattern
- `.planning/phases/25-.../25-SECURITY.md` (targeted grep + read) — confirmed FU-23-A's current scope is exactly `GetCatalogHtmlPath`/`OpenExternal`, `SearchIndexed` already excluded with cited rationale
- `.planning/phases/25-.../25-*-PLAN.md` (grep) — confirmed the FU-23-A sweep's assignment history to Phase 26
- `.planning/config.json` — full file read; confirmed `nyquist_validation: true`, no `security_enforcement: false` override
- Live environment probes this session: `go version` (go1.26.6), `node --version` (v24.14.1), `npm --version` (11.18.0), `wails version` (v2.10.2)

### Secondary (MEDIUM confidence)
- Wails v2 binding auto-regeneration behavior on `wails dev` — [CITED: WebSearch this session against wails.io official docs and community sources; no single canonical page fetched in full, cross-referenced against this project's own working `wails dev` history in STATE.md's per-phase notes]

### Tertiary (LOW confidence)
- None — every claim in this document is either a direct source read this session or an explicit `[ASSUMED]`-free open question.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependency; every reused primitive verified by direct file read
- Architecture: HIGH — mount pattern, reducer shape, config shape all verified against live source; the one genuine architectural risk (sync boot vs. async config) is clearly flagged with a recommended resolution, not left as a guess
- Pitfalls: HIGH — all five pitfalls are grounded in direct source reads this session (line-cited), not inferred from CONTEXT.md's prose alone; two of the five (Pitfall 3's antd claim, Pitfall 4's version string) are corrections to CONTEXT.md's own stated rationale, verified independently

**Research date:** 2026-08-15
**Valid until:** 30 days (stable in-repo domain; no external library version drift risk since no new dependency was added)
