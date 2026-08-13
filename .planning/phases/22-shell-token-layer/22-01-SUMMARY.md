---
phase: 22-shell-token-layer
plan: 01
subsystem: ui
tags: [react, typescript, wails, css-custom-properties, oklab, theming]

requires: []
provides:
  - "themeTokens.ts: lum/mixHex/mixAlpha/computeTokens/applyTokens/readPersistedPrefs/initThemeTokens"
  - "themes.ts tokens block (7-field ThemeTokens) on all 11 themes"
  - "workspace.css: 14-token + 5-density-var :root fallbacks, z-index scale, responsive grid ladder"
  - "WorkspaceShell + 5 region components (Toolbar/CatalogRail/TreePane/DetailsPanel/StatusBar) as static skeletons"
  - "App.tsx rewritten to render the workspace shell; CatalogModal kept functional but unreachable"
  - "main.go: TitleBarHiddenInset wired into the existing options.App literal"
  - "DevStateSwitcher: DEV-only Ctrl+Alt+T/D theme/density cycling, stripped from prod build"
affects: [22-02, 22-03, 22-04, 22-05, 22-06, 22-07]

actuals:
  tokens: 22800
  tasks: 2
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Theme tokens computed in TypeScript via OKLab blending (not CSS color-mix()) and applied synchronously at main.tsx module scope, before createRoot's render, to avoid launch-time theme flash"
    - "One code path (applyTokens) writes both the 14 extended tokens + 5 density vars and the 16 legacy ThemeColors properties, so index.css's antd overrides keep feeding CatalogModal through the migration"
    - "localStorage prefs (theme/density/rail-side) resolved via strict allowlist with fallback-and-rewrite, never trusting a raw stored string"
    - "Responsive grid ladder uses min-width-only media queries (200px 1fr / 236px 1fr / 268px 1fr 288px) to avoid the max-width/min-width gap at fractional viewport widths"

key-files:
  created:
    - frontend/src/themeTokens.ts
    - frontend/src/workspace.css
    - frontend/src/components/workspace/WorkspaceShell.tsx
    - frontend/src/components/workspace/Toolbar.tsx
    - frontend/src/components/workspace/CatalogRail.tsx
    - frontend/src/components/workspace/TreePane.tsx
    - frontend/src/components/workspace/DetailsPanel.tsx
    - frontend/src/components/workspace/StatusBar.tsx
    - frontend/src/components/dev/DevStateSwitcher.tsx
  modified:
    - frontend/src/themes.ts
    - frontend/src/main.tsx
    - frontend/src/App.tsx
    - frontend/src/style.css
    - frontend/src/index.css
    - main.go
  deleted:
    - frontend/src/components/Header.tsx
    - frontend/src/components/MainContent.tsx
    - frontend/src/components/WelcomeContent.tsx
    - frontend/src/components/ModernTable.tsx
    - frontend/src/components/tabs/CreateCatalogTab.tsx
    - frontend/src/components/tabs/SearchCatalogsTab.tsx
    - frontend/src/components/tabs/BrowseCatalogsTab.tsx

key-decisions:
  - "Implemented full OKLab conversion (Ottosson matrices) for mixHex/mixAlpha rather than sRGB channel averaging, per the locked CONTEXT.md decision -- exact fidelity with the prototype's color-mix(in oklab, ...) output"
  - "Dropped getThemeById/getDefaultTheme imports from the rewritten App.tsx (unused once readPersistedPrefs owns resolution) to keep tsc --noEmit clean under noUnusedLocals"
  - "IBM Plex @font-face declarations deferred to plan 22-02 (font files not vendored yet) -- workspace.css references the font-family names but falls through the stack until 22-02 lands the woff2 assets"
  - "Fixed a coordinator-flagged defect during the Task 1 checkpoint: html/body ground color was reading a hardcoded navy (style.css) and the legacy --app-bg token (index.css) instead of the new --bg token, causing a wrong ground color on all 6 light themes and a visible mismatch on StorCat Dark"

patterns-established:
  - "Dev-only keyboard affordances gated behind import.meta.env.DEV, verified stripped from dist/assets/*.js by grep after npm run build"
  - "No raw z-index literals in components -- always var(--z-details-drawer|--z-overlay|--z-dialog)"

requirements-completed: [SHELL-01, SHELL-02, SHELL-08, THEME-01, THEME-02, THEME-03, THEME-04, THEME-06]

coverage:
  - id: D1
    description: "Workspace shell renders (46px toolbar, three-column responsive grid, 26px status bar) with no tab UI, painted from computed theme tokens"
    requirement: "SHELL-01"
    verification:
      - kind: automated_ui
        ref: "coordinator UAT: .ant-tabs absent, .ws-toolbar 46px, .ws-status 26px confirmed via Vite dev server DOM inspection"
        status: pass
    human_judgment: false
  - id: D2
    description: "Grid computes to 268px 1fr 288px at >=1280px"
    requirement: "SHELL-02"
    verification:
      - kind: automated_ui
        ref: "coordinator UAT: .ws-grid computed grid-template-columns = 268px 839px 288px at 1395px viewport"
        status: pass
    human_judgment: false
  - id: D3
    description: "macOS TitleBarHiddenInset wired into main.go's options.App literal"
    requirement: "SHELL-08"
    verification:
      - kind: unit
        ref: "go build ./... && go test ./... (exit 0)"
        status: pass
    human_judgment: true
    rationale: "Real traffic-light inset/clearance can only be confirmed visually on an actual macOS build, per 22-RESEARCH.md's own Environment Availability note -- not verifiable from this session"
  - id: D4
    description: "All 14 extended tokens resolve to concrete values on :root for every theme; no unresolved color-mix() or empty value"
    requirement: "THEME-03"
    verification:
      - kind: automated_ui
        ref: "coordinator UAT: :root computed styles listed all 14 tokens with concrete rgb()/hex/rgba() values"
        status: pass
    human_judgment: false
  - id: D5
    description: "Theme repaints fully on switch (verified this phase only via DevStateSwitcher, no Settings UI yet); --onac luminance contrast correct"
    requirement: "THEME-01, THEME-02"
    verification: []
    human_judgment: true
    rationale: "Cycling all 11 themes via Ctrl+Alt+T and confirming --onac legibility on light-accent themes (Gruvbox orange, Monokai green) requires a human/visual pass through DevStateSwitcher in wails dev -- not exercised by this session's automated checks"
  - id: D6
    description: "Density toggle changes --rh/--rp/--mp/--hp/--fs between Compact and Comfortable"
    requirement: "THEME-04"
    verification:
      - kind: automated_ui
        ref: "coordinator UAT: :root density vars confirmed present at Comfortable defaults (--rh 34px etc.)"
        status: pass
    human_judgment: true
    rationale: "Compact-value swap (27px etc.) via Ctrl+Alt+D was not exercised in the coordinator's UAT pass, only the Comfortable default was confirmed"
  - id: D7
    description: "Theme applied synchronously before first paint; missing/corrupt localStorage values fall back to locked defaults and rewrite"
    requirement: "THEME-06"
    verification:
      - kind: automated_ui
        ref: "coordinator UAT: presetting storcat-theme-id=storcat-dark then reloading shows 35 inline custom properties already set on document.documentElement.style before paint, --bg: #0b0e13"
        status: pass
    human_judgment: false
  - id: D8
    description: "DEV-only theme/density switcher stripped entirely from production build"
    verification:
      - kind: other
        ref: "npm run build && grep -l storcat-dev-switcher dist/assets/*.js (0 matches)"
        status: pass
    human_judgment: false

duration: 11min
completed: 2026-08-13
status: complete
---

# Phase 22 Plan 01: Shell + Token Layer Summary

**Workspace shell tracer slice: OKLab-computed 14-token theme layer applied pre-paint, three-tab UI replaced by a 46px-toolbar/rail/tree/details/26px-status-bar grid, and macOS TitleBarHiddenInset wired in Go.**

## Performance

- **Duration:** ~11 min (commit-to-commit; excludes checkpoint wait for coordinator UAT)
- **Started:** 2026-08-13T15:28:50-05:00
- **Completed:** 2026-08-13T15:40:13-05:00
- **Tasks:** 2
- **Files modified:** 20 (13 created, 6 modified, 7 deleted)

## Accomplishments
- All 11 themes in `themes.ts` carry a 7-field `tokens` block sourced verbatim from the design handoff's `THEMES` array
- `themeTokens.ts` ports `lum()`/`varsFor()`/density-branch verbatim, implementing full OKLab conversion (not RGB averaging) for `mixHex`/`mixAlpha` per the locked CONTEXT.md decision
- Theme/density/rail-side resolved via strict allowlist validation with fallback-and-rewrite (`readPersistedPrefs`), applied synchronously in `main.tsx` before `createRoot`'s render — no theme flash
- Three-tab Ant Design UI fully removed; `App.tsx` renders the new `WorkspaceShell`, `CatalogModal` stays functional but unreachable
- macOS `TitleBarHiddenInset` wired into `main.go`'s existing `options.App` literal, no build tag needed
- `DevStateSwitcher` gives the only way to exercise all 11 themes and both densities before a real Settings UI ships; confirmed stripped from the production bundle

## Task Commits

Each task was committed atomically:

1. **Task 1: End-to-end "the workspace frame paints itself from computed theme tokens"** - `9e4bc1b4` (feat)
2. **Fix (checkpoint deviation): use `--bg` token for html/body ground color** - `7c4b20f4` (fix)
3. **Task 2: DEV-only state-exercise affordance for theme and density** - `c18c8b66` (feat)

_Note: commit 2 is a Rule 1 auto-fix surfaced by the coordinator's own UAT pass during the tracer's checkpoint, not a separately-planned task._

## Files Created/Modified

- `frontend/src/themes.ts` - Added `ThemeTokens` interface + `tokens` field on all 11 theme objects
- `frontend/src/themeTokens.ts` - `lum`, OKLab `mixHex`/`mixAlpha`, `computeTokens`, `applyTokens`, `readPersistedPrefs`, `initThemeTokens`
- `frontend/src/workspace.css` - Token/density/z-index `:root` fallbacks, responsive grid ladder, shell layout classes, drag-region and density-contract utility classes
- `frontend/src/main.tsx` - Calls `initThemeTokens()` synchronously at module scope before `createRoot`
- `frontend/src/App.tsx` - Rewritten to render `WorkspaceShell` + `CatalogModal` (+ DEV-only `DevStateSwitcher`); dropped the tab-era `applyTheme`/theme-load `useEffect`
- `frontend/src/style.css` / `frontend/src/index.css` - `html`/`body` ground color now reads `var(--bg)` instead of a hardcoded navy / the legacy `--app-bg` token
- `frontend/src/components/workspace/{WorkspaceShell,Toolbar,CatalogRail,TreePane,DetailsPanel,StatusBar}.tsx` - New shell + 5 skeleton region components
- `frontend/src/components/dev/DevStateSwitcher.tsx` - Ctrl+Alt+T/D keyboard cycling, DEV-only
- `main.go` - `Mac: &mac.Options{TitleBar: mac.TitleBarHiddenInset()}` added to the existing `options.App{}` literal
- Deleted: `Header.tsx`, `MainContent.tsx`, `WelcomeContent.tsx`, `ModernTable.tsx`, `components/tabs/*`

## Decisions Made
- Full OKLab conversion (Ottosson matrices) implemented for `mixHex`/`mixAlpha` rather than the simpler sRGB-average approximation RESEARCH.md flagged as an open question — matches the milestone's pixel-final fidelity bar.
- Dropped now-unused `getThemeById`/`getDefaultTheme` imports from `App.tsx` (PATTERNS.md's "keep verbatim" guidance for these two symbols was superseded by `readPersistedPrefs` owning theme resolution) to satisfy `tsc --noEmit` under `noUnusedLocals`.
- IBM Plex `@font-face` declarations deferred to plan 22-02 (its own task, alongside vendoring the woff2 files) — `workspace.css` references the `'IBM Plex Sans'`/`'IBM Plex Mono'` family names now so no later plan needs to touch the shell layout rules, but the actual `@font-face` blocks and asset files land in 22-02.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `html`/`body` ground color didn't use the new `--bg` token**
- **Found during:** Coordinator's UAT pass on Task 1's tracer checkpoint
- **Issue:** `style.css`'s `html` rule still carried the Wails scaffold's hardcoded `rgba(27, 38, 54, 1)` (a fixed dark navy), wrong on all 6 light themes now that a token-driven ground exists. `index.css`'s `body` rule read the legacy `--app-bg` token instead of the new `--bg`, causing a visible mismatch in themes where the two diverge (e.g. StorCat Dark: `#0f172a` vs `#0b0e13`).
- **Fix:** `style.css`: `html { background-color: var(--bg); }`. `index.css`: `body { background: var(--bg); }` (left `color: var(--app-text)` and the rest of that legacy block untouched — it still feeds `CatalogModal`'s antd overrides).
- **Files modified:** `frontend/src/style.css`, `frontend/src/index.css`
- **Verification:** `npx tsc --noEmit` and `npm run build` both still exit 0 after the change; coordinator confirmed both light and dark launches render correctly pre-paint.
- **Committed in:** `7c4b20f4`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Necessary correctness fix surfaced by the tracer feedback gate exactly as designed — caught before Task 2 built on top of it. No scope creep.

## Issues Encountered
- `frontend/node_modules` was not yet installed in this environment; ran `npm install` once before the first `tsc`/`build` verification. No `package.json`/`package-lock.json` changes resulted (dependencies were already pinned correctly).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- The token pipeline, shell frame, and macOS window-chrome option are proven end to end — plans 22-02 (fonts/typography), 22-03 (toolbar), 22-04 (rail), 22-05 (tree/details), 22-06 (responsive/AppContext), and 22-07 (details drawer) can now expand outward from real, committed skeletons rather than a design draft.
- **Manual verification still outstanding** (flagged in `coverage` above, `human_judgment: true`): full 11-theme cycle + `--onac` legibility check via `DevStateSwitcher`, Compact-density value confirmation, and a real macOS build for the `TitleBarHiddenInset` traffic-light inset — none of these block downstream plans, but should be swept before the phase gate per `22-VALIDATION.md`'s Sampling Rate.
- No blockers for 22-02 onward.

---
*Phase: 22-shell-token-layer*
*Completed: 2026-08-13*
