---
phase: 22-shell-token-layer
reviewed: 2026-08-13T00:00:00Z
depth: standard
files_reviewed: 17
files_reviewed_list:
  - frontend/src/themeTokens.ts
  - frontend/src/themes.ts
  - frontend/src/workspace.css
  - frontend/src/main.tsx
  - frontend/src/App.tsx
  - frontend/src/contexts/AppContext.tsx
  - frontend/src/hooks/useMediaQuery.ts
  - frontend/src/components/workspace/WorkspaceShell.tsx
  - frontend/src/components/workspace/Toolbar.tsx
  - frontend/src/components/workspace/CatalogRail.tsx
  - frontend/src/components/workspace/TreePane.tsx
  - frontend/src/components/workspace/DetailsPanel.tsx
  - frontend/src/components/workspace/StatusBar.tsx
  - frontend/src/components/dev/DevStateSwitcher.tsx
  - frontend/src/style.css
  - frontend/src/index.css
  - main.go
findings:
  critical: 0
  warning: 2
  info: 3
  total: 5
status: issues_found
---

# Phase 22: Code Review Report

**Reviewed:** 2026-08-13
**Depth:** standard
**Files Reviewed:** 17 (plus verification that `Header.tsx`, `MainContent.tsx`, `WelcomeContent.tsx`, `ModernTable.tsx`, `components/tabs/*` were fully deleted with no dangling references)
**Status:** issues_found

## Summary

This is a well-executed phase. `tsc --noEmit`, `npm run build`, and `go build ./...` all pass clean. The OKLab conversion matrices in `themeTokens.ts` were checked term-by-term against Björn Ottosson's published constants and are correct; `mixHex`/`mixAlpha` direction (`pct=100` → `a`, `pct=0` → `b`) matches the spec's `mix(a, pct%, b)` semantics for every one of the seven derived-token formulas; and all 11 themes' `tokens` blocks were diffed against the UI-SPEC's authoritative table byte-for-byte with no mismatches. `readPersistedPrefs()` correctly allowlists `storcat-density`/`storcat-rail-side` to exact strings, falls back to the locked defaults on any missing/corrupt/unknown value, never throws (all storage access wrapped in try/catch), and rewrites a normalized value. The responsive grid in `workspace.css` is a genuine mobile-first `min-width`-only ladder (200px 1fr → 236px 1fr @1040px → 268px 1fr 288px @1280px) with zero `max-width` pairing anywhere in the diff, so the fractional-width gap the plan called out cannot reoccur. `useMediaQuery`'s add/remove of the `change` listener is symmetric, uses the modern `addEventListener` API only, and has no stale-closure risk. The drawer's Escape/backdrop close paths, and the "widen past 1280px clears `detailOverlay`" reset, are all correctly implemented in `WorkspaceShell.tsx`. No dangling imports/references to any of the five deleted tab-era files were found anywhere in the tree.

Two Warnings were found, both concentrated in the theme layer's state-management seams introduced by this phase (not in the token math itself): a set of newly-added but entirely unreachable reducer actions/state for the catalog modal, and a real, reproducible staleness bug in the toolbar's theme-name chip that undermines this phase's own manual THEME-01 verification step. Three Info items round out minor code-quality observations.

## Warnings

### WR-01: Dead reducer actions/state added for the catalog modal — duplicates an already-working mechanism

**File:** `frontend/src/contexts/AppContext.tsx:9-11, 18-19, 30-32, 43-56`
**Issue:** This phase's `AppContext` refactor pruned 15 dead tab-era fields but added a new, entirely unused slice: `catalogModalOpen`, `catalogModalTitle`, `catalogModalHtmlPath` state fields and the `OPEN_CATALOG_MODAL` / `CLOSE_CATALOG_MODAL` action types plus their reducer cases. Nothing in the codebase dispatches these actions or reads this state — `grep -rn "catalogModalOpen\|OPEN_CATALOG_MODAL\|CLOSE_CATALOG_MODAL"` matches only the definitions inside `AppContext.tsx` itself. The catalog modal is actually driven, unchanged, by `App.tsx`'s own local `useState` (`catalogModalVisible`/`catalogModalPath`) plus the pre-existing `openCatalogModal` `CustomEvent` listener (`App.tsx:29-34`). This is a second, competing state mechanism for the same concern that was never wired up — exactly the "config/state for values that never change/get used" pattern the project's over-engineering guardrail exists to catch.
**Fix:** Delete the four unused lines of `AppState`, the two unused `AppAction` variants, and their two reducer cases. If the intent is to migrate `CatalogModal`'s visibility into the reducer as part of this phase, finish the wiring (have `App.tsx` dispatch `OPEN_CATALOG_MODAL`/`CLOSE_CATALOG_MODAL` and read `state.catalogModalOpen` instead of its local `useState`) — but per `22-CONTEXT.md`, `CatalogModal` migration is explicitly out of scope until Phase 26, so deleting the dead branch is the correct minimal fix now:
```ts
export interface AppState {
  density: Density;
  railSide: RailSide;
  detailOverlay: boolean;
}
type AppAction =
  | { type: 'SET_DENSITY'; payload: Density }
  | { type: 'SET_RAIL_SIDE'; payload: RailSide }
  | { type: 'SET_DETAIL_OVERLAY'; payload: boolean };
```

### WR-02: Toolbar's theme-name chip is not reactive — goes stale after a theme change, breaking THEME-01's own verification method

**File:** `frontend/src/components/workspace/Toolbar.tsx:13`, `frontend/src/components/dev/DevStateSwitcher.tsx:23-28`, `frontend/src/App.tsx:18-24`
**Issue:** `Toolbar` derives its displayed theme name with `const themeName = readPersistedPrefs().theme.name;` — a fresh `localStorage` read performed directly in the render body, not from any reactive state or prop. Nothing causes `Toolbar` to re-render when *only* the theme changes:
- `DevStateSwitcher`'s `Ctrl+Alt+T` handler (`DevStateSwitcher.tsx:23-28`) writes the new theme to `localStorage` and calls `applyTokens()` directly — this updates the CSS custom properties (so panel colors/accents repaint correctly) and updates `DevStateSwitcher`'s own local `theme` state (so its bottom-right debug readout updates), but `DevStateSwitcher` and `Toolbar` are unconnected sibling subtrees under `<AppProvider>` in `App.tsx` — no context value or prop connects them for theme. `WorkspaceShell`'s only density-triggered `applyTokens` effect (`WorkspaceShell.tsx:22-24`) depends solely on `state.density`, so a theme-only keystroke does not cause `WorkspaceShell`/`Toolbar` to re-render at all. The toolbar's theme-name label is left showing the *previous* theme until an unrelated re-render happens to occur (e.g. the tester also presses `Ctrl+Alt+D`).
- Separately, `App.tsx`'s `handleThemeChange` (`App.tsx:18-24`, the future Settings-surface hook for Phase 26) calls `applyTokens(newTheme, ...)` and updates `currentTheme` local state, but **never writes `THEME_KEY` to `localStorage`** — so even if something forced `Toolbar` to re-render, its `readPersistedPrefs()` read would still return the *old* theme id, permanently out of sync with the tokens actually painted on `:root` until the next full reload.

This is real, reproducible today via the shipped Wave-0 dev affordance, and it directly undermines `22-VALIDATION.md`'s own manual instruction for THEME-01/THEME-02: *"Cycle all 11 via the Wave 0 dev affordance; confirm every surface repaints with no stale colors."* The colors do repaint; the toolbar chip's text does not, which will read as a false negative (or false positive, if the tester doesn't notice the label lagging) during the mandated all-11-themes contrast pass.
**Fix:** Give `Toolbar` the current theme as a prop derived from a single reactive source instead of an ad-hoc `localStorage` read, e.g. lift `currentTheme` from `App.tsx` down through `WorkspaceShell` to `Toolbar`, and have `handleThemeChange` persist to `localStorage` as part of applying it:
```ts
// App.tsx
const handleThemeChange = (event: CustomEvent) => {
  const { theme: newTheme } = event.detail;
  if (newTheme) {
    setCurrentTheme(newTheme);
    localStorage.setItem(THEME_KEY, newTheme.id); // keep storage in sync
    applyTokens(newTheme, readPersistedPrefs().density);
  }
};
// ... <WorkspaceShell themeName={currentTheme.name} />
```
and thread `themeName` as a prop into `Toolbar` rather than calling `readPersistedPrefs()` inside its render body. `DevStateSwitcher` should update the same lifted state (or dispatch to it) rather than writing `localStorage` independently.

## Info

### IN-01: `readPersistedPrefs()` (has read+write side effects) is called redundantly from four separate places at startup

**File:** `frontend/src/contexts/AppContext.tsx:24`, `frontend/src/main.tsx:11` (via `initThemeTokens`), `frontend/src/App.tsx:12`, `frontend/src/components/workspace/Toolbar.tsx:13`
**Issue:** `readPersistedPrefs()` reads three `localStorage` keys and conditionally rewrites them on every call. It is invoked once at `AppContext.tsx` module scope, once inside `initThemeTokens()` in `main.tsx`, once in `App.tsx`'s `useState` lazy initializer, and once per-render inside `Toolbar`. All calls are idempotent after the first (no functional bug), but this is duplicated I/O and four independent read sites for what should be one source of truth — a symptom of theme not yet living in the reducer (tracked for Phase 26).
**Fix:** No urgent action needed; when theme moves into `AppState` in Phase 26, collapse these call sites to the single `initThemeTokens()` call plus reducer state reads.

### IN-02: macOS toolbar inset is applied one tick after first paint, risking a one-frame layout shift

**File:** `frontend/src/components/workspace/Toolbar.tsx:15-35`
**Issue:** `--toolbar-inset-left` (which reserves 78px for the macOS traffic lights) is set via `document.documentElement.style.setProperty(...)` inside a `Promise.resolve().then(() => Environment())` chain, i.e., after mount. On a real macOS build this resolves within a microtask/short delay, but it is not guaranteed to complete before the browser paints the first frame, so there is a narrow window where toolbar content can render flush-left before snapping over by 78px. This is a much smaller version of the "launch flash" problem the phase otherwise took care to eliminate for theme tokens (`main.tsx`'s comment about avoiding a post-mount effect).
**Fix:** Low priority given the magnitude (a few pixels, one frame, cosmetic only). If it needs tightening, the Wails runtime's platform can potentially be read synchronously via `window.runtime` metadata rather than the async `Environment()` call, or the inset could be applied via a `useLayoutEffect` to at least run before the browser paints (though `Environment()` itself is still a Promise, so this only removes React's own commit-to-paint gap, not the async fetch delay).

### IN-03: Empty CSS rule left as a placeholder

**File:** `frontend/src/workspace.css:189-191`
**Issue:** `.ws-details--pane { /* Inline grid child -- no positioning overrides needed. */ }` is a rule block containing only a comment — it compiles fine and is harmless, but it's dead CSS (a selector with nothing to select for).
**Fix:** Either remove the empty rule (the comment can move to sit above `.ws-details--drawer` instead, which is the block doing the actual overriding), or leave as-is if the team wants an explicit "no-op documented intentionally" marker — genuinely optional, not blocking.

---

_Reviewed: 2026-08-13_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
