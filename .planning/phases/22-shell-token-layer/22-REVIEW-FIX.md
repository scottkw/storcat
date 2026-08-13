---
phase: 22-shell-token-layer
fixed_at: 2026-08-13T00:00:00Z
review_path: .planning/phases/22-shell-token-layer/22-REVIEW.md
iteration: 1
findings_in_scope: 5
fixed: 4
skipped: 1
status: partial
---

# Phase 22: Code Review Fix Report

**Fixed at:** 2026-08-13
**Source review:** .planning/phases/22-shell-token-layer/22-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 5 (0 Critical / 2 Warning / 3 Info -- full scope)
- Fixed: 4
- Skipped: 1 (not-needed per the finding's own recommendation, not a fix failure)

## Fixed Issues

### WR-01: Dead reducer actions/state added for the catalog modal — duplicates an already-working mechanism

**Files modified:** `frontend/src/contexts/AppContext.tsx`
**Commit:** `8f1d51e1`
**Applied fix:** Confirmed via repo-wide grep (`catalogModalOpen|OPEN_CATALOG_MODAL|CLOSE_CATALOG_MODAL|catalogModalTitle|catalogModalHtmlPath`) that only the definitions inside `AppContext.tsx` itself matched -- zero consumers, in this phase or before it. Deleted the four `AppState` fields, the two `AppAction` variants, their `initialState` entries, and their two reducer cases exactly as the review's suggested diff specified. `CatalogModal` continues to be driven by `App.tsx`'s local `useState` + `openCatalogModal` `CustomEvent`, unchanged; migrating it into the reducer stays out of scope until Phase 26 per `22-CONTEXT.md`.

### WR-02: Toolbar's theme-name chip is not reactive — goes stale after a theme change

**Files modified:** `frontend/src/App.tsx`, `frontend/src/components/workspace/WorkspaceShell.tsx`, `frontend/src/components/workspace/Toolbar.tsx`, `frontend/src/components/dev/DevStateSwitcher.tsx`
**Commit:** `e9471c0e`
**Applied fix:** Adapted the review's suggested lift-as-prop fix, extended to also collapse `DevStateSwitcher`'s competing write path (the review flagged this as a secondary option: "should update the same lifted state (or dispatch to it)"):
- `App.tsx`'s `currentTheme` state is now the single reactive source, threaded as a prop (`themeName` / `currentTheme`) through `WorkspaceShell` into `Toolbar` and into `DevStateSwitcher`.
- `Toolbar` no longer calls `readPersistedPrefs()` in its render body; it renders the `themeName` prop.
- `App.tsx`'s `handleThemeChange` now writes `THEME_KEY` to `localStorage` (previously it only called `applyTokens` + updated local state, silently diverging from storage).
- `DevStateSwitcher`'s `Ctrl+Alt+T` handler no longer owns a parallel `useState` copy of theme or calls `applyTokens`/`localStorage.setItem` directly -- it dispatches the same `window.dispatchEvent(new CustomEvent('themeChange', ...))` that `App.tsx` already listens for (the same established cross-component mechanism `22-CONTEXT.md` documents for theme, alongside `openCatalogModal`), so there is exactly one path that applies tokens, updates React state, and persists to storage.

No new context or state library was introduced -- theme stays outside the `AppState` reducer per `22-CONTEXT.md`'s Phase 26 boundary, propagated via existing props/CustomEvent patterns only.

**Verified live in a browser** (Vite dev server, `Ctrl+Alt+T` cycling Light → Dark):
```json
{
 "before": "StorCat Light",
 "after": {
  "chip": "StorCat Dark",
  "bg": "#0b0e13",
  "stored": "storcat-dark"
 }
}
```
Chip label, `--bg` token, and persisted `storcat-theme-id` all agree after the switch.

### IN-02: macOS toolbar inset is applied one tick after first paint, risking a one-frame layout shift

**Files modified:** `frontend/src/components/workspace/Toolbar.tsx`
**Commit:** `24108009`
**Applied fix:** Swapped `useEffect` for `useLayoutEffect` per the review's suggested tightening, closing React's own commit-to-paint gap. The `Promise.resolve().then(() => Environment())` guard is unchanged and still required -- the raw `Environment()` call throws synchronously outside a Wails webview, which would unmount the whole tree; deferring through the resolved promise keeps that a catchable rejection so the shell still renders in a plain browser. As the review itself noted, this narrows but does not eliminate the window, since `Environment()`'s async fetch delay is unaffected.

### IN-03: Empty CSS rule left as a placeholder

**Files modified:** `frontend/src/workspace.css`
**Commit:** `dd869b25`
**Applied fix:** Removed the empty `.ws-details--pane { }` rule; moved its explanatory comment to sit above `.ws-details--drawer`, the block that actually performs the positioning override.

## Skipped Issues

### IN-01: `readPersistedPrefs()` (has read+write side effects) is called redundantly from four separate places at startup

**File:** `frontend/src/contexts/AppContext.tsx:24`, `frontend/src/main.tsx:11`, `frontend/src/App.tsx:12`, `frontend/src/components/workspace/Toolbar.tsx:13`
**Reason:** Not a fix failure -- skipped per the finding's own **Fix:** text: "No urgent action needed; when theme moves into `AppState` in Phase 26, collapse these call sites to the single `initThemeTokens()` call plus reducer state reads." All four call sites are idempotent after the first call (no functional bug), and consolidating them now would be premature given theme's Phase 26 reducer migration is the natural point to collapse them. Correctly left as a tracked Phase 26 cleanup item.
**Original issue:** `readPersistedPrefs()` reads three `localStorage` keys and conditionally rewrites them on every call; it's invoked once at `AppContext.tsx` module scope, once inside `initThemeTokens()` in `main.tsx`, once in `App.tsx`'s lazy initializer, and once per-render in `Toolbar` (this last site was already eliminated as a side effect of the WR-02 fix, which replaced `Toolbar`'s render-body read with a prop).

## Verification

All four project gates run clean after all four fixes, in the main checkout (worktrees disabled via `workflow.use_worktrees: false` in `.planning/config.json` -- fixes were applied and committed directly on `main`, not in an isolated worktree, so these results are reproducible from this tree as-is):
- `cd frontend && npx tsc --noEmit` — exit 0, no errors
- `cd frontend && npm run build` — exit 0, 1446 modules transformed, `dist/` produced
- `go build ./...` — exit 0
- `go test ./...` — exit 0, all packages `ok`

Live browser verification (Vite dev server at `http://localhost:5173`, `dev-browser`, `Ctrl+Alt+T` theme cycle) confirmed the WR-02 fix: theme chip label, `--bg` computed CSS custom property, and the persisted `storcat-theme-id` `localStorage` key all agree after a theme switch (see WR-02 above for the captured evidence).

`22-REVIEW.md` was updated in place with a **Disposition:** line under each of the 5 findings recording fixed/not-needed status and commit hashes, and committed separately.

---

_Fixed: 2026-08-13_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
