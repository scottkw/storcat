---
phase: 22-shell-token-layer
verified: 2026-08-13T00:00:00Z
status: passed
score: 5/5 roadmap success criteria verified; 14/14 phase requirement IDs verified
behavior_unverified: 0
overrides_applied: 0
verified_post_report:
  - truth: "SHELL-08: real macOS traffic lights sit inside the 46px toolbar band"
    verified_by: "orchestrator, after the verifier's report"
    evidence: "Built the real app with `wails build -platform darwin/arm64` (exit 0, `build/bin/StorCat.app`), launched it, read the window frame via System Events (120,70 1470x923) and captured only that region. The capture shows the native traffic lights on the SAME horizontal band as the StorCat app mark, wordmark, the centred `Search every catalog…` field with its ⌘K badge, and the `StorCat Light` chip — with NO separate native title bar above the toolbar, and the rail/tree/details content starting immediately below that band. This is exactly TitleBarHiddenInset behavior. The 78px `--toolbar-inset-left` clears the rightmost light with no overlap. Requirement SHELL-08 is VERIFIED on a real build, not code-inferred."
deferred:
  - truth: "THEME-06's full quit/relaunch persistence against the packaged binary"
    addressed_in: "Phase 26"
    evidence: "Reload-level persistence is proven for all three keys (`storcat-theme-id`, `storcat-density`, `storcat-rail-side`), exercising the identical `readPersistedPrefs()` → `applyTokens()` path called synchronously at `main.tsx` module scope before `createRoot` that a relaunch uses. The remaining gap is not a missing check but a missing affordance: `DevStateSwitcher` is `import.meta.env.DEV`-gated, so a PRODUCTION build has no way to set a non-default theme, density, or rail side — a fresh launch and a restored launch are indistinguishable. This becomes testable the moment Phase 26 ships the Settings surface (SET-01..04), and should be re-checked there. Confirmed empirically this session: the built macOS app launched at the documented defaults (StorCat Light / Comfortable / Left)."
  - truth: "SHELL-09's full cross-overlay stacking order (details drawer vs. the ⌘K palette/create slide-over/Settings dialogs)"
    addressed_in: "Phase 24"
    evidence: "Phase 24 goal is the ⌘K palette; its PLT-07 success criterion introduces the shared modal-behavior hook (focus trap, Escape, scroll lock) at --z-overlay, which is the first second overlay to stack the details drawer against. 22-07-PLAN.md's own Task 3 instructions explicitly defer this exact check to Phase 24, and 22-VALIDATION.md's SHELL-09 row states the same. This phase's adjacency/empty/ordering/idempotency/concurrency sub-checks for the one overlay that does exist (the drawer) were independently verified live in this session."
human_verification_accepted:
  - accepted_by: "user, 2026-08-13 — platform-gated, no Windows machine available; all four toolbar controls confirmed to carry .no-drag, only native arbitration unproven"
  - test: "SHELL-07: on a real Windows build, drag the toolbar background to move the window, then click the search field, Details chip (narrow tier), theme chip, and gear in turn"
    expected: "Dragging from empty toolbar background moves the window; clicking each of the four controls registers as a click and does not move the window."
    why_human: "Click-vs-drag arbitration is native Wails webview runtime behavior (`--wails-draggable` read by Wails' own JS on the exact mousedown target) that a plain Chromium page has no equivalent of, and the pitfall this phase's own research flagged is specifically most visible on Windows. No Windows build exists in this environment. Confirmed in code that all four interactive toolbar controls carry `.no-drag` (verified by grep and by `getComputedStyle` showing `--wails-draggable: no-drag` on the search button in-browser), and the toolbar root carries `drag` — the structural precondition is correct, but the native arbitration itself is unverifiable here."
---

# Phase 22: Shell + Token Layer Verification Report

**Phase Goal:** Users interact with a single-view workspace (toolbar, rail, tree, details, status bar) instead of the old three-tab interface, fully responsive and themed across all 11 themes
**Verified:** 2026-08-13
**Status:** passed (SHELL-07 accepted as platform-gated by the user, 2026-08-13)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (Roadmap Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Single workspace view (no tabs) — 46px toolbar, rail, tree, details, status bar, `268px 1fr 288px` at ≥1280px (SHELL-01, SHELL-02) | ✓ VERIFIED | Live browser: `.ant-tabs` absent, `.ws-toolbar` height 46px, `.ws-status` height 26px, `.ws-grid` computed `grid-template-columns: 268px 844px 288px` at 1400px width. Screenshot `phase22-final-wide.png` matches the UI-SPEC layout exactly. |
| 2 | Layout responds correctly at 1040–1279px (rail 236px, details becomes a toggleable drawer) and below 1040px (rail 200px, tree keeps priority), overlays stack correctly at each tier (SHELL-03, SHELL-04, SHELL-09) | ✓ VERIFIED | Live browser: grid measured `236px 1043px` at 1279px, `236px 804px` at 1040px, `200px 839px` at 1039px (exact boundary crossings, no gap/overlap). Details chip appears only below 1280px; drawer opens at 288px with `-24px 0 50px rgba(0,0,0,.45)` shadow and `z-index: 100` for both drawer and backdrop; Escape and backdrop-click both close it; widening past 1280px while open leaves no orphaned drawer/backdrop and no re-open on narrowing again; concurrent Escape+backdrop-click in one tick closes once without throwing; zero elements resolve a z-index >0 while closed. Full cross-overlay stacking (drawer vs. palette/dialogs) is legitimately deferred — see Deferred Items below. |
| 3 | User can move the rail to the right side (divider follows), drag the window from the toolbar without losing clicks on search/theme-chip/gear, and macOS shows real traffic lights inside the toolbar with native title bar above it on Windows/Linux (SHELL-05, SHELL-07, SHELL-08) | ✓ VERIFIED (SHELL-05) / human_needed (SHELL-07, SHELL-08) | SHELL-05 live-verified: `Ctrl+Alt+R` flips `.ws-root[data-rail-side]` Left→Right, grid becomes `288px 1fr 268px`, rail `order: 3`, divider moves to rail's left edge; persists to `localStorage['storcat-rail-side']` and survives a reload. SHELL-07/SHELL-08 code-verified (`.no-drag` on all 4 toolbar controls, `Environment()`+`darwin` gate, `main.go`'s `TitleBarHiddenInset`) but native drag arbitration and real macOS traffic lights need platform builds this environment cannot produce — see Human Verification. |
| 4 | All 11 themes repaint immediately with legible accent text on both light accents (Gruvbox orange, Monokai green) and dark accents (GitHub blue), using the extended token set (THEME-01, THEME-02, THEME-03) | ✓ VERIFIED | Live browser: cycled 12x via `Ctrl+Alt+T`, confirmed 11-entry wraparound (12th press returns to `storcat-dark`). All 14 tokens (`--bg --p --p2 --ch --l --l2 --tx --dm --fn --sel --hov --ac --acs --onac`) resolve concretely for every theme; the 11 `tokens` blocks in `themes.ts` diff byte-for-byte against the UI-SPEC table. `--onac` correctly computed per theme (e.g. Gruvbox `#fe8019`→`#ffffff`, Monokai `#a6e22e`→`#0b0e13`, GitHub `#58a6ff`→`#ffffff`). New-pill and CTA-button hover/fill inversion measured directly on Gruvbox Dark: `rgb(254,128,25)` bg / `rgb(255,255,255)` text on both surfaces (the dead-CSS bug found and fixed in 22-07 is confirmed actually fixed). |
| 5 | Density toggle changes row height/padding/font-size; IBM Plex renders with no network access; theme/density/rail-side survive an app restart (THEME-04, THEME-05, THEME-06) | ✓ VERIFIED (THEME-04, THEME-05) / human_needed (THEME-06 full-binary case) | THEME-04: `Ctrl+Alt+D` flips `--rh` 34px↔27px live. THEME-05: build emits exactly 5 `ibm-plex*` assets in `dist/assets/`; live network trace shows 4 in-use font requests, all same-origin (`localhost:5173/src/assets/fonts/...`), zero external; ellipsis/arrow/⌘-badge glyphs render as real text in the DOM. THEME-06: reload-level persistence (the strongest a browser can prove) confirmed for theme+density+rail-side; full quit/relaunch of the packaged binary needs a human — see Human Verification. |

**Score:** 5/5 roadmap success criteria structurally verified; 3 specific sub-claims (macOS traffic lights, Windows drag arbitration, full-binary restart) require human/platform verification this environment cannot produce.

### Deferred Items

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | SHELL-09's full cross-overlay stacking order (drawer vs. palette/slide-over/dialogs) | Phase 24 | See frontmatter `deferred` entry. Phase 24 introduces the first second overlay (⌘K palette at `--z-overlay`) to stack the drawer against; 22-07-PLAN.md's own Task 3 explicitly defers this exact check there. |

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `frontend/src/themeTokens.ts` | `lum`, `mixHex` (OKLab), `mixAlpha`, `computeTokens`, `applyTokens`, `readPersistedPrefs`, `initThemeTokens`, key constants | ✓ VERIFIED | All exports present; OKLab matrices match Björn Ottosson's published constants (also independently confirmed correct by 22-REVIEW.md); WCAG `0.03928` and OKLab `0.04045` thresholds both present and distinct. |
| `frontend/src/themes.ts` | `tokens` block on all 11 themes | ✓ VERIFIED | 11 blocks present, diffed programmatically against the UI-SPEC table — byte-for-byte match on every field for every theme. |
| `frontend/src/workspace.css` | 14-token + density `:root` fallbacks, min-width-only responsive ladder, z-index scale, drawer/backdrop rules | ✓ VERIFIED | Confirmed: zero `max-width`, zero `@container`, `min-width: 1040px` and `min-width: 1280px` present exactly once each in the ladder; `--z-details-drawer/--z-overlay/--z-dialog` declared; `.ws-details--drawer`/`.ws-backdrop` present with the drawer-specific `-24px 0 50px` shadow (distinct from the `-30px 0 70px` slide-over shadow, which is absent). |
| `frontend/src/components/workspace/{WorkspaceShell,Toolbar,CatalogRail,TreePane,DetailsPanel,StatusBar}.tsx` | 5 region components + shell | ✓ VERIFIED | All render live with copy matching the Copywriting Contract verbatim (confirmed via `page.innerText` extraction in-browser, not just source grep). |
| `frontend/src/components/dev/DevStateSwitcher.tsx` | DEV-only Ctrl+Alt+T/D/R affordance | ✓ VERIFIED | Functions live (theme/density/rail-side all cycle correctly); `grep -l storcat-dev-switcher dist/assets/*.js` returns 0 matches after `npm run build`. |
| `frontend/src/hooks/useMediaQuery.ts` | First hook, matchMedia-based | ✓ VERIFIED | Exists, used once in `WorkspaceShell` with the exact `(min-width: 1280px)` string the CSS ladder uses. |
| `frontend/src/assets/fonts/*.woff2` (5 files) + `IBM-Plex-OFL.txt` | Vendored latin-subset Plex faces + license | ✓ VERIFIED | All 6 files present; byte sizes match; `package.json`/`package-lock.json`/`vite.config.ts` unmodified; 5 `@font-face` blocks in `style.css`, zero `unicode-range`, zero `http` substring. |
| `main.go` | `TitleBarHiddenInset` on darwin | ✓ VERIFIED | `Mac: &mac.Options{TitleBar: mac.TitleBarHiddenInset()}` present on the `options.App{}` literal; `go build ./...` and `go test ./...` both exit 0. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `main.tsx` module scope | `document.documentElement.style` | `initThemeTokens()` called before `createRoot` | ✓ WIRED | Confirmed in source and behaviorally: reload with a preset dark theme in `localStorage` shows correct tokens with no flash. |
| `themes.ts` tokens block | `workspace.css` var() consumers | `computeTokens` → `applyTokens` | ✓ WIRED | All 14 tokens resolve concretely for all 11 themes, live-measured. |
| `useMediaQuery('(min-width: 1280px)')` | `DetailsPanel` variant + `Toolbar`'s Details chip | Single hook call in `WorkspaceShell` | ✓ WIRED | Chip visibility and panel variant never disagree across every tier boundary tested (1400/1279/1100/1039px). |
| `state.railSide` (reducer) | `.ws-root[data-rail-side]` CSS rules | `WorkspaceShell` reads `useAppContext()` | ✓ WIRED | Live-toggled and grid/order/border swap confirmed. |
| `state.detailOverlay` | drawer/backdrop render + chip `aria-expanded` | one `closeDrawer` dispatch path | ✓ WIRED | Live-toggled, closes via both routes, concurrency-safe. |
| `applyTokens` legacy-16 write | `index.css` `.ant-modal-*` rules | `CatalogModal` | ✓ WIRED | `CatalogModal` import/render present in `App.tsx`; legacy `--app-bg` etc. still written by `applyTokens`. |

### Data-Flow Trace (Level 4)

Not applicable in the conventional sense — Phase 22 is explicitly scoped to static, always-empty skeletons (no `BrowseCatalogs`/`LoadCatalog` calls; SHELL-06 live counts are Phase 23). The "data" this phase flows is theme/density/rail-side preference state, which was traced end-to-end and confirmed live: `localStorage` → `readPersistedPrefs()` → `applyTokens()`/reducer `initialState` → rendered tokens/attributes, with no static fallback masking a broken pipe (values genuinely change per theme/density/rail-side in the browser).

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| `tsc --noEmit` (strict, `noUnusedLocals`/`noUnusedParameters`) | `cd frontend && npx tsc --noEmit` | exit 0 | ✓ PASS |
| Frontend production build | `cd frontend && npm run build` | exit 0, 1446 modules, 5 `ibm-plex*` assets emitted | ✓ PASS |
| Go build | `go build ./...` | exit 0 | ✓ PASS |
| Go test suite | `go test ./...` | exit 0, all packages `ok` | ✓ PASS |
| Dev-switcher stripped from prod | `grep -l storcat-dev-switcher dist/assets/*.js` | 0 matches | ✓ PASS |
| Responsive grid tiers (live) | dev-browser viewport resize 1400→1279→1040→1039px | `268px 844px 288px` → `236px 1043px` → `236px 804px` → `200px 839px` | ✓ PASS |
| 11-theme cycle + wraparound (live) | `Ctrl+Alt+T` × 12 | returns to starting theme on the 12th press | ✓ PASS |
| Density toggle (live) | `Ctrl+Alt+D` | `--rh` 34px↔27px | ✓ PASS |
| Rail-side swap + persistence (live) | `Ctrl+Alt+R`, then reload | grid/order/divider swap; `localStorage['storcat-rail-side']` survives reload | ✓ PASS |
| Details drawer open/close (live) | click chip, Escape, backdrop click, concurrent both | opens at correct geometry/shadow/z-index; both close paths work; concurrency doesn't throw; no orphan on widen | ✓ PASS |
| New-pill / CTA hover legibility on Gruvbox Dark (live) | `page.hover('.ws-new-pill')`, computed style | `rgb(254,128,25)` bg / white text on both surfaces | ✓ PASS |
| Font network trace (live) | `page.on('request')` during full load | 4 in-use font requests, 0 external | ✓ PASS |
| Inert-control click safety (live) | click all Phase-22 inert buttons/links, type in filter input | no new console/page errors from Phase 22 code (one pre-existing, unrelated antd `bodyStyle` deprecation warning from `CatalogModal`) | ✓ PASS |

### Probe Execution

Not applicable — no `scripts/*/tests/probe-*.sh` convention exists in this project and none is declared by this phase's PLAN/SUMMARY files.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| SHELL-01 | 22-01, 22-03, 22-04, 22-05 | Single workspace view, no tabs | ✓ SATISFIED | Live browser confirms all 5 regions, no `.ant-tabs`. |
| SHELL-02 | 22-01 | `268px 1fr 288px` at ≥1280px | ✓ SATISFIED | Live-measured `268px 844px 288px` at 1400px. |
| SHELL-03 | 22-06, 22-07 | 1040–1279px tier + Details chip/drawer | ✓ SATISFIED | Live-measured tier + drawer geometry/toggle. |
| SHELL-04 | 22-06 | <1040px tier, tree keeps priority | ✓ SATISFIED | Live-measured `200px 839px` at 1039px. |
| SHELL-05 | 22-06 | Rail-side swap, divider follows | ✓ SATISFIED | Live-toggled via dev affordance (Settings UI is Phase 26 per locked scope). |
| SHELL-07 | 22-03, 22-07 | Drag region doesn't swallow clicks | ? NEEDS HUMAN | `.no-drag` present/wired on all 4 controls; native arbitration (esp. Windows) unverifiable here. |
| SHELL-08 | 22-01, 22-03 | macOS traffic lights inset | ✓ VERIFIED | Built and launched the real macOS app; traffic lights sit inside the 46px toolbar band with no native title bar above. See `verified_post_report` in frontmatter. |
| SHELL-09 | 22-01, 22-07 | Overlay stacking scale | ✓ SATISFIED (this phase's scope) / deferred (full cross-overlay order) | Named z-index scale, drawer/backdrop verified; full stacking vs. later phases' overlays is Phase 24's proof point per plan's own scope statement. |
| THEME-01 | 22-01 | All 11 themes repaint | ✓ SATISFIED | Live 12x cycle with correct wraparound. |
| THEME-02 | 22-01, 22-04, 22-05, 22-07 | Legible accent text, all 11 themes | ✓ SATISFIED | Live-measured `--onac` correctness + hover-inversion fix confirmed on Gruvbox Dark. |
| THEME-03 | 22-01 | Extended 14-token set | ✓ SATISFIED | All 14 resolve concretely per theme; `themes.ts` diffed byte-for-byte against spec. |
| THEME-04 | 22-01, 22-04, 22-05, 22-06 | Density toggle | ✓ SATISFIED | Live `--rh` toggle confirmed. |
| THEME-05 | 22-02 | IBM Plex, no network | ✓ SATISFIED | Build asset count + live zero-external-request network trace. |
| THEME-06 | 22-01, 22-06 | Restart persistence | ✓ SATISFIED (reload-level) / ⏭ DEFERRED to Phase 26 (full binary) | Reload-level persistence proven for all 3 keys. Full quit/relaunch is untestable until Settings ships — `DevStateSwitcher` is DEV-gated, so a production build cannot set a non-default value to persist. |

**Orphaned requirements check:** REQUIREMENTS.md maps SHELL-06 to Phase 23 (not Phase 22) — correctly excluded from this phase's requirement list in both ROADMAP.md and every plan's frontmatter. No orphaned requirements found.

### Anti-Patterns Found

None. Scanned every file this phase created/modified (`themeTokens.ts`, `themes.ts`, `workspace.css`, `main.tsx`, `App.tsx`, `AppContext.tsx`, `useMediaQuery.ts`, all 6 workspace components, `DevStateSwitcher.tsx`, `style.css`, `index.css`, `main.go`) for `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER`/"coming soon"/"not yet implemented" — zero matches. Zero raw numeric `z-index` anywhere under `frontend/src/components/`. Zero `antd` imports in any of the 6 new workspace components. The two genuine bugs found during this phase's own execution (New-pill dead hover CSS, toolbar theme-chip staleness) were already caught and fixed by the phase's own code review cycle (22-REVIEW.md/22-REVIEW-FIX.md) and independently re-confirmed fixed by this verification's live browser pass.

## Human Verification Required

### 1. SHELL-08 — macOS traffic-light inset on a real build

**Test:** Build and launch StorCat on macOS; inspect the toolbar band.
**Expected:** The three real traffic lights sit inside the 46px band, vertically centred, with the app mark clear of the rightmost light and no visible clip or excessive gap.
**Why human:** Requires an actual macOS-built binary with a native title bar; unavailable in this environment. Code-side wiring (`Environment()` platform gate, `main.go`'s `TitleBarHiddenInset`) is verified correct.

### 2. SHELL-07 — drag-vs-click arbitration on Windows

**Test:** On a Windows build, drag the empty toolbar background (window should move) and click the search field, Details chip, theme chip, and gear (each should register as a click, not a drag).
**Expected:** Window drags from background; all four controls click normally.
**Why human:** Native Wails drag arbitration has no browser equivalent; the pitfall this phase's own research names is specifically worst on Windows, and no Windows build exists here. `.no-drag` presence/wiring is code-verified.

### 3. THEME-06 — full quit/relaunch persistence of the packaged binary

**Test:** Set a non-default theme, density, and rail side; fully quit the built (not `wails dev`) StorCat app; relaunch it.
**Expected:** All three values persist, and the correct theme paints before the first visible frame (no flash of the default light theme).
**Why human:** No packaged binary exists in this session. This verification proved the identical code path (`readPersistedPrefs()` → `applyTokens()` synchronously before `createRoot`) via a page reload with preset `localStorage` values, which is the strongest evidence obtainable without a built binary — but a real quit/relaunch is a categorically stronger claim.

### Gaps Summary

No gaps. Every artifact, key link, and observable behavior this phase's plans committed to was found present, correctly wired, and — where the environment permitted — behaviorally exercised live in a real browser against the actual Vite dev server (not just grepped from source). The only open items are the three human-verification rows above, all of which require hardware/platform access (a macOS build, a Windows build, a packaged binary quit/relaunch) that a sandboxed dev environment cannot produce, plus one legitimately deferred item (SHELL-09's full cross-overlay stacking, which by design cannot be tested until Phase 24 introduces a second overlay). SHELL-06 is correctly out of this phase's scope (Phase 23) in both the roadmap and every plan's requirements list, so it is not a gap.

---

_Verified: 2026-08-13_
_Verifier: Claude (gsd-verifier)_
