---
phase: 22-shell-token-layer
plan: 02
subsystem: ui
tags: [fonts, css, vite, wails, ibm-plex, licensing]

requires:
  - phase: 22-01
    provides: "workspace.css font-family references to 'IBM Plex Sans'/'IBM Plex Mono' (declared but unbacked by any @font-face until this plan)"
provides:
  - "5 latin-subset IBM Plex woff2 files under frontend/src/assets/fonts/ (Sans 400/500/600, Mono 400/500)"
  - "5 @font-face declarations in frontend/src/style.css, font-display: swap, no unicode-range"
  - "frontend/src/assets/fonts/IBM-Plex-OFL.txt — IBM's own OFL-1.1 text, separate from Nunito's OFL.txt"
affects: [22-03, 22-04, 22-05, 22-06, 22-07]

actuals:
  tokens: 340
  tasks: 2
  commits: 1

tech-stack:
  added: []
  patterns:
    - "Font vendoring via one-time npm pack + tarball extraction, never added to package.json — identical to the pre-existing Nunito precedent in this repo"

key-files:
  created:
    - frontend/src/assets/fonts/ibm-plex-sans-latin-400-normal.woff2
    - frontend/src/assets/fonts/ibm-plex-sans-latin-500-normal.woff2
    - frontend/src/assets/fonts/ibm-plex-sans-latin-600-normal.woff2
    - frontend/src/assets/fonts/ibm-plex-mono-latin-400-normal.woff2
    - frontend/src/assets/fonts/ibm-plex-mono-latin-500-normal.woff2
    - frontend/src/assets/fonts/IBM-Plex-OFL.txt
  modified:
    - frontend/src/style.css

key-decisions:
  - "Checkpoint (Task 1, package-legitimacy gate) resolved on the orchestrator's advance, standing approval per its checkpoint_authority block — not re-raised during execution."
  - "@font-face src paths declared without a leading './' to exactly match the existing Nunito rule's form (url(\"assets/fonts/...\")), per the plan's explicit instruction to mirror that rule's shape."

requirements-completed: [THEME-05]

coverage:
  - id: D1
    description: "5 latin-subset IBM Plex woff2 files (Sans 400/500/600, Mono 400/500) vendored via npm pack, never added to package.json/package-lock.json"
    requirement: "THEME-05"
    verification:
      - kind: unit
        ref: "ls -l + file(1) magic-byte check on all 5 files; byte sizes 22588/24184/24252/14708/14888 matched 22-RESEARCH.md's measured table exactly"
        status: pass
      - kind: unit
        ref: "grep -c fontsource frontend/package.json == 0; git diff --stat frontend/package.json frontend/package-lock.json frontend/vite.config.ts == empty"
        status: pass
    human_judgment: false
  - id: D2
    description: "5 @font-face blocks declared in style.css: font-display swap, no unicode-range, no http substring"
    requirement: "THEME-05"
    verification:
      - kind: unit
        ref: "grep -c 'IBM Plex' style.css == 5; grep -c 'font-display: *swap' style.css == 5; grep -c unicode-range style.css == 0; grep -ci http style.css == 0"
        status: pass
    human_judgment: false
  - id: D3
    description: "IBM's own OFL-1.1 copyright/license text shipped as a separate file from Nunito's OFL.txt"
    requirement: "THEME-05"
    verification:
      - kind: unit
        ref: "frontend/src/assets/fonts/IBM-Plex-OFL.txt created; head -5 confirms 'Copyright 2019 IBM Corp.' text distinct from the existing OFL.txt's Nunito copyright block, which is untouched"
        status: pass
    human_judgment: false
  - id: D4
    description: "Vite's default asset pipeline fingerprints the 5 new woff2 files into dist/assets/ with zero vite.config.ts/main.go changes, proving the //go:embed all:frontend/dist chain bakes them into the binary"
    requirement: "THEME-05"
    verification:
      - kind: unit
        ref: "cd frontend && npx tsc --noEmit && npm run build (exit 0); ls dist/assets/ | grep -c ibm-plex == 5; built CSS's url() values confirmed same-origin relative paths, no external host"
        status: pass
    human_judgment: false
  - id: D5
    description: "Fonts render correctly in the actual running Wails app (not just the build output) with zero external font requests on a normal load, and Plex still renders — not a system fallback — with the network disconnected"
    requirement: "THEME-05"
    verification: []
    human_judgment: true
    rationale: "Requires a real `wails dev` window (DevTools Network tab inspection + Offline-mode reload). A plain `vite dev` server cannot substitute: Toolbar.tsx's Environment() call throws synchronously against `window.runtime` outside an actual Wails webview (pre-existing from 22-01, not introduced here), so the React tree never mounts under a bare browser. dev-browser (CDP) cannot attach to Wails' native WKWebView from this environment. Structural evidence (built CSS uses only same-origin relative asset URLs, confirmed by direct inspection of dist/assets/index.*.css) strongly supports the offline claim but does not substitute for the actual manual check called out in 22-VALIDATION.md's Sampling Rate."
---

# Phase 22 Plan 02: IBM Plex Font Vendoring Summary

**Vendored 5 latin-subset IBM Plex woff2 files (Sans 400/500/600, Mono 400/500) via one-time `npm pack` extraction — never added as a `package.json` dependency — and declared them as `@font-face` in `style.css`, closing THEME-05's self-hosted, no-network typography requirement.**

## Performance

- **Duration:** ~15 min
- **Completed:** 2026-08-13
- **Tasks:** 2 (1 checkpoint, pre-approved by orchestrator authority; 1 auto)
- **Files modified:** 7 (6 created, 1 modified)

## Accomplishments
- Extracted exactly 5 pre-subsetted `.woff2` files from `@fontsource/ibm-plex-sans@5.3.0` / `@fontsource/ibm-plex-mono@5.3.0` tarballs in a scratch directory, verified byte-for-byte against 22-RESEARCH.md's measured table (22588 / 24184 / 24252 / 14708 / 14888 bytes) and confirmed genuine WOFF2 binaries via `file(1)` magic-byte inspection
- Copied IBM's own OFL-1.1 `LICENSE` text into a new `IBM-Plex-OFL.txt`, kept fully separate from the pre-existing Nunito `OFL.txt` (different copyright holders — IBM Corp. vs. The Nunito Project Authors)
- Appended 5 `@font-face` blocks to `frontend/src/style.css`, immediately after the existing Nunito block, matching its exact `url("assets/fonts/...")` path form (no leading `./`), each with `font-display: swap` and no `unicode-range`
- `frontend/package.json`, `frontend/package-lock.json`, and `frontend/vite.config.ts` are byte-for-byte unchanged — confirmed via `git diff --stat`
- `npm run build` fingerprints all 5 new files into `dist/assets/` with zero build-config changes, and the built CSS's `url()` values resolve to same-origin relative paths only (no `http`/external host) — the same zero-config `//go:embed all:frontend/dist` mechanism the existing Nunito font already uses

## Task Commits

Each task was committed atomically:

1. **Task 1: Confirm @fontsource package legitimacy before any registry fetch** — checkpoint, no commit (resolved via orchestrator's standing `checkpoint_authority` approval; not re-raised)
2. **Task 2: Vendor five IBM Plex woff2 files, IBM's OFL text, and five @font-face declarations** - `30f08a76` (feat)

## Files Created/Modified

- `frontend/src/assets/fonts/ibm-plex-sans-latin-400-normal.woff2` - 22,588 bytes
- `frontend/src/assets/fonts/ibm-plex-sans-latin-500-normal.woff2` - 24,184 bytes
- `frontend/src/assets/fonts/ibm-plex-sans-latin-600-normal.woff2` - 24,252 bytes
- `frontend/src/assets/fonts/ibm-plex-mono-latin-400-normal.woff2` - 14,708 bytes
- `frontend/src/assets/fonts/ibm-plex-mono-latin-500-normal.woff2` - 14,888 bytes
- `frontend/src/assets/fonts/IBM-Plex-OFL.txt` - IBM Corp.'s OFL-1.1 copyright/license text (new, separate from `OFL.txt`)
- `frontend/src/style.css` - 5 new `@font-face` blocks appended after the existing Nunito block

## Decisions Made
- Task 1's blocking package-legitimacy checkpoint was resolved on the orchestrator's own advance approval (delivered via this session's `<checkpoint_authority>` block), per its explicit "Proceed on this authority — do not re-raise the checkpoint" instruction. No interactive prompt was issued.
- `@font-face` `src` paths were written as `url("assets/fonts/<file>.woff2")` — matching the existing Nunito rule's exact form (no leading `./`) — rather than RESEARCH.md's own code-example sketch, which used a `./` prefix. The plan's action text explicitly calls out matching the existing rule's form as the authoritative shape.

## Deviations from Plan

None - plan executed exactly as written. All byte sizes matched the measured table with zero deviation, so no re-derivation of the artifact list was needed.

## Issues Encountered

- Attempted to visually verify font rendering and zero-network-request behavior via `dev-browser` against a plain `vite dev` server (bypassing the full Wails backend, since only static asset delivery was in question). The app's React tree failed to mount under a bare browser: `Toolbar.tsx`'s `Environment()` call (added in 22-01, unrelated to this plan) throws synchronously against `window.runtime`, which only exists inside an actual Wails webview. This is a pre-existing condition from 22-01, not a regression introduced here, and it means the `<human-check>`'s `wails dev` + DevTools Offline-mode verification cannot be substituted with a headless browser check in this environment — see `coverage` D5 above for the resulting `human_judgment: true` flag.
- As a partial substitute, inspected the built `dist/assets/index.*.css` directly and confirmed all 5 `@font-face` `url()` values are same-origin relative paths (`/assets/ibm-plex-*.woff2`) with no external host referenced anywhere in the stylesheet — structural evidence supporting (but not replacing) the offline-reload human check.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- The five Plex faces are vendored, declared, and proven to fingerprint into the Vite/Go build chain — 22-03 (toolbar), 22-04 (rail), 22-05 (tree/details) can now rely on `'IBM Plex Sans'`/`'IBM Plex Mono'` resolving to real files rather than falling through to the generic-family end of the stack.
- **Manual verification still outstanding** (flagged in `coverage` D5 above, `human_judgment: true`): a real `wails dev` session with DevTools Network-tab inspection (confirm same-origin-only font requests) and an Offline-mode reload (confirm Plex still renders, not a system fallback). Per 22-VALIDATION.md's Sampling Rate, this should be swept before the phase gate — does not block 22-03 onward.
- No blockers for 22-03 onward.

## Self-Check: PASSED

All 6 created/modified files verified present on disk; commit `30f08a76` verified present in `git log`.

---
*Phase: 22-shell-token-layer*
*Completed: 2026-08-13*
