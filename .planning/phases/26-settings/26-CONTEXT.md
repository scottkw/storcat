# Phase 26: Settings - Context

**Gathered:** 2026-08-15
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous) — all three grey areas accepted as recommended

<domain>
## Phase Boundary

Users configure theme, density, rail position, and catalog defaults from one settings surface that saves as they go.

**In scope:** the Settings dialog and its three entry points (⌘, / gear / theme chip); 11 theme cards; density and rail-position segmented controls; catalog directory, default filename root, and the four catalog toggles; save-as-you-change persistence; **resolving the two-store split** between `localStorage` and the Go config; and **discharging the two carried security obligations that have now been deferred four times.**

**Out of scope (later phases):** catalog actions and fsnotify watch (Phase 27 — this phase only adds the *watch directory* toggle and persists it; the watcher itself is Phase 27), re-scan and diff (Phase 28).

</domain>

<decisions>
## Implementation Decisions

### Settings persistence — resolving the two-store split
- **The Go config (`internal/config`) becomes the single source of truth**, with a **one-time `localStorage` → config migration on first run**. Phase 22 explicitly deferred this ("theme stays local pending Phase 26's Settings-owned theme state"); this is that resolution. Two stores that can silently diverge is the actual defect being fixed here, not just a tidiness preference.
- **Every change writes immediately — no debounce.** The config file is a few hundred bytes and `Manager.Save()` already exists. A timer would add a lost-write window on quit for no measurable benefit, and SET-05's "no explicit save step" is satisfied either way.
- **New config fields:** `density`, `railSide`, `catalogDirectory`, `defaultFilenameRoot`, `writeHTML`, `copyToSecondary`, `secondaryDirectory`, `watchDirectory`. **`rememberWindow` is NOT new** — it already exists as `windowPersistence` with working `Set`/`Get` methods, so COMPAT-05 wires the toggle to the existing field rather than reinventing it.
- **Phase 25's `storcat-catalog-directory` and `storcat-secondary-directory` localStorage keys migrate in the same pass.** Phase 25 bootstrapped its own directory persistence ahead of Settings existing (following the RAIL-05 precedent); folding them in here was always the plan.

### The two carried security obligations (deferred four times — must be discharged)
- **T-22-05 — DELETE `CatalogModal.tsx` and its `App.tsx` wiring**, rather than sanitizing it. Verified during discuss: the only `dispatchEvent` in the whole frontend is for `themeChange`, so **nothing dispatches `openCatalogModal`** — the listener at `App.tsx:37` is registered but unreachable. The component still calls the legacy `window.electronAPI` shim and antd's `message` (antd was removed this milestone), and `DetailsPanel.tsx:121-123` already provides the real "Open HTML catalog" path via `getCatalogHtmlPath` + `openExternal`. Deleting **removes** the threat rather than mitigating it; sanitizing would mean hand-rolling HTML sanitization, since no new dependency is permitted.
- **FU-23-A — `GetCatalogHtmlPath` gets `catalogDir` threaded through and reuses `internal/osutil`'s `containsPath`**, exactly the treatment `RevealInFileManager` received in Phase 23's review finding WR-02.
- **FU-23-A — `OpenExternal` is restricted to `file://` paths contained in the catalog directory.** Its only live caller opens a catalog's own HTML file. Rejecting everything else is simpler and safer than allow-listing URL schemes.
- **`SearchIndexed` is removed from the FU-23-A sweep list, with the reason recorded.** Phase 25's security audit verified it takes no renderer-supplied path reaching a filesystem write and received its own containment at introduction. STATE.md's target list should be corrected rather than carrying a redundant item forward.

### The Settings surface
- **A centred dialog at `--z-dialog` (300)** — Phase 22's locked z-index scale places Settings there — **consuming Phase 24's `useModalBehavior`** for focus trap, Escape, scroll lock and focus restore. No bespoke reimplementation (PLT-07's standing constraint on Phases 25–27).
- **11 theme cards, each a 4-swatch strip plus a light/dark tag**, rendered from the existing `THEMES` array, which PROJECT.md already records as authoritative.
- **Theme applies immediately on click** — SET-05 specifies no save step and `applyTokens()` already runs synchronously. No apply-on-close.
- **Density and rail position use segmented controls**, a new shared control introduced by this phase in the same way Phase 25 introduced the toggle switch.
- **The footer version string is rendered from the existing `GetVersion()` binding, and `wails.json`'s `productVersion` is bumped `2.3.0` → `3.0.0` in this phase** (USER decision, resolving 26-RESEARCH.md Open Question 2). The UI-SPEC's locked copy `"StorCat 3.0.0 · settings save as you change them"` is therefore produced dynamically, not hardcoded — one source of truth, no hand-updated string at each release. Accepted side effect: build artifacts stamp `3.0.0` from this phase onward, ahead of the v3.0.0 release itself.

### Claude's Discretion
- Config field names/JSON keys and the migration's exact detection mechanism.
- Settings dialog component decomposition and section ordering.
- Segmented-control markup and its ARIA pattern.
- Whether the theme chip opens Settings scrolled to the theme section or at the top.
- **Theme state plumbing** (26-RESEARCH.md Open Question 1): default to reusing the existing `themeChange` CustomEvent that `DevStateSwitcher.tsx` already names as the future Settings path, rather than lifting theme into the `AppContext` reducer. Either satisfies SET-02; take the larger refactor only if the plan finds a concrete reason.
- **Dead `Config.SidebarPosition` / `SetSidebarPosition`** (26-RESEARCH.md Open Question 3): default to leaving both in place, untouched and unrepurposed, with a one-line note that they are intentionally orphaned v1 leftovers distinct from the new `railSide` field. They are not the new concept, and deleting a persisted config field is not this phase's job.
</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/config/config.go` — `Config` struct, `Manager` with `Load`/`Save`/`Get` and existing `SetTheme`, `SetSidebarPosition`, `SetWindowSize`, `SetWindowPosition`, `SetWindowPersistence`, `GetWindowPersistence`.
- `app.go:541-580` — existing `GetConfig`, `SetTheme`, `SetSidebarPosition`, window bindings.
- `frontend/src/hooks/useModalBehavior.ts` — Phase 24's shared hook; Phase 25 already extended it for an animated exit without breaking its `isOpen`-transition cleanup.
- `frontend/src/themes.ts` — the `THEMES` array (11 themes, each with colors + light/dark type).
- `frontend/src/themeTokens.ts` — `applyTokens()`, `readPersistedPrefs()`, `safeGetItem`/`safeSetItem`.
- `internal/osutil` — `containsPath`, the containment helper FU-23-A must reuse.
- `frontend/src/components/workspace/create/OptionsToggles.tsx` — Phase 25's toggle switch, reusable here.

### Established Patterns
- New GUI capability goes in a new method beside any CLI-shared one; the CLI path stays untouched.
- Binding calls route through `wailsAPI.ts`'s `extractErrorMessage()` (Wails rejects with a plain string).
- Go tests are table-driven `*_test.go` beside source; **no frontend test framework** (TEST-01 deferred) — proof is `tsc --noEmit` + `vite build` + live dev-browser at `:34115`.

### Integration Points
- `internal/config/config.go` → new fields + setters + migration.
- `app.go` → new/updated bindings; containment on `GetCatalogHtmlPath` and `OpenExternal`.
- `frontend/src/App.tsx` → remove `CatalogModal` import, state, listener and render.
- **DELETE** `frontend/src/components/CatalogModal.tsx`.
- `AppContext.tsx` → density/railSide now sourced from config rather than localStorage.
- `Toolbar.tsx` (gear, theme chip), `WorkspaceShell.tsx` (⌘, listener) → Settings entry points.
- `frontend/wailsjs/**` → regenerate after binding changes.
</code_context>

<specifics>
## Specific Ideas

- STATE.md is explicit that T-22-05 and FU-23-A **must not be re-accepted a third time**; they have in fact now been carried through Phases 22, 23, 24 and 25. This phase discharges both.
- `.planning/WINDOWS.md` currently holds 6 open entries (#1 Phase 23 reveal argv, #2 Phase 24 Ctrl+K, #3 Phase 25 CRT-13 — since closed live, #4 Windows `GetDiskFreeSpaceEx`, #5 Linux `/proc/mounts`, #6 atomic-write SIGKILL). Entry #3 should be marked fixed, since wave 7 of Phase 25 verified it live.
- COMPAT-05 is about *continuity*: window size/position persistence must keep working exactly as it did pre-milestone, now surfaced as a Settings toggle over the already-working `windowPersistence` field.
</specifics>

<deferred>
## Deferred Ideas

- The fsnotify watcher itself — Phase 27 (WATCH-01/02/03). This phase persists the toggle only.
- Catalog rename/duplicate/delete — Phase 27.
- Re-scan and diff — Phase 28.
- Frontend unit tests for the Settings surface — TEST-01, deferred at v3.0.0 requirements definition.
- CLI subcommands for settings — new capabilities are GUI-only (FUT-03).
</deferred>
