---
phase: 25-create-slide-over-progress-cancellation-partial-catalog
plan: 06
subsystem: ui
tags: [react, wails, live-progress, cancellation, status-bar, background-handoff]

# Dependency graph
requires:
  - phase: 25-01
    provides: "CreateSlideOver.tsx shell, lifted AppContext.scan state machine, scanPercent/formatEta (scanFormat.ts), classifyScanFailure/ScanFailure (types/scan.ts) this plan finally wires up"
  - phase: 25-05
    provides: "VolumePicker/CreateForm/OptionsToggles form sub-components this plan's ScanningBody sits alongside; sourceDisplayNameOf/sourcePathOf reused unchanged"
provides:
  - "ScanningBody.tsx -- the full live progress surface for both scanning sub-states (counting and percentage-known), sharing one component instance and state object"
  - "CreateSlideOver.tsx's handleCloseRequest(reason: CloseReason) -- the single state-aware close handler every close trigger (Escape, header x, scrim, Discard-and-close) routes through, cancelling the scan only when reason is 'cancel-the-scan' and the scan is actually running"
  - "AppContext's SCAN_PROGRESS reducer case extended with currentPath/log (9-line capped, monotonic-prepend) for the WALKING line and log box"
  - "StatusBar.tsx's fourth, conditionally-rendered, right-aligned .ws-status-scan segment -- clickable to reopen the panel at its live state"
  - "CreateSlideOver's scan:progress subscription un-gated from isOpen -- the background-handoff prerequisite that keeps AppContext's scan state (and therefore the status bar) live while the panel is closed"
affects: [25-07]

# Actuals (#2632)
actuals:
  tokens: 6439
  tasks: 3
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "One component/one state-object rendering of two visually-distinct sub-states (ScanningBody's counting vs percentage-known), derived from a single `pct === null` signal rather than a second explicit flag -- avoids the remount/reset-to-zero bug a component-swap approach would risk"
    - "Explicit close-reason argument (CloseReason: 'cancel-the-scan' | 'leave-it-running') on a single close handler, rather than inferring intent from which trigger fired or from state alone -- makes 'Run in background' deliberately NOT a cancel path by construction"
    - "Retention cap enforced at the point state is appended (AppContext's reducer), not only at render (ScanningBody's defensive re-slice) -- state itself never grows unbounded over a long scan (T-25-23)"
    - "An always-mounted overlay's own effect un-gated from its isOpen prop (CreateSlideOver's scan:progress subscription) as the mechanism for background-state liveness, rather than lifting the subscription into a separate always-mounted component"

key-files:
  created:
    - frontend/src/components/workspace/create/ScanningBody.tsx
  modified:
    - frontend/src/components/workspace/CreateSlideOver.tsx
    - frontend/src/components/workspace/StatusBar.tsx
    - frontend/src/contexts/AppContext.tsx
    - frontend/src/types/scan.ts
    - frontend/src/workspace.css

key-decisions:
  - "CreateSlideOver's scan:progress EventsOn subscription changed from isOpen-gated to always-on. CreateSlideOver is already permanently mounted by WorkspaceShell (established in 25-01), so un-gating this one effect was sufficient to make CRT-08's background handoff work -- no separate always-mounted subscriber component was needed, and StatusBar reads the same already-updated AppContext state rather than subscribing to the event itself."
  - "The scanning body's outer wrapper (`.ws-create-scan-body`) now owns its own flex/padding/gap (22px/22px per 25-UI-SPEC's Step 2 Scanning Contract) instead of being double-nested inside the generic `.ws-create-body` (18px/16px, correct for the form step only) -- CreateSlideOver's JSX was restructured so isForm/isScanning/isDone each own their correct top-level container rather than sharing one."
  - "handleCloseRequest keys cancellation on an explicit reason argument, not on which trigger called it or on state alone -- this is what keeps 'Run in background' a non-cancelling close by construction rather than by a state check a future trigger could get wrong."

patterns-established:
  - "Explicit close-reason argument for a single multi-purpose close handler -- reusable for any future overlay whose close gesture needs to carry more than one intent."

requirements-completed: [CRT-07, CRT-08, CRT-09]

coverage:
  - id: D1
    description: "The counting sub-state shows a live incrementing file count with no progress track and no fabricated percentage; the percentage-known sub-state shows a real progress bar, byte total, and computed ETA; the transition between them happens in place with no remount and no reset to zero (CRT-07)"
    requirement: CRT-07
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- a folder scan (frontend/node_modules, 24171 files) showed 'Counting...' in muted color with no progress track and a live incrementing '1409 files found so far' before resolving; a volume-shaped source with a real totalBytesHint (a synthetic ListVolumes override pointing at a real, large, readable directory since no true external volume with substantial readable content exists on this machine) showed the percentage-known layout immediately at 0%, then progressed 0% -> 30% -> 34% -> 48% -> 57% -> 66% -> done with real counters ('92916 files - 7.2G - about 15s left') and a pulsing WALKING path; percentage never regressed and never showed 100% before the write actually completed"
        status: pass
    human_judgment: false
  - id: D2
    description: "The log box retains at most 9 newest-first lines inside a fixed max-height, never scrolls, and renders empty with no placeholder text when no lines exist yet; the WALKING path and log lines both break anywhere"
    requirement: CRT-07
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- live scan against /Users/ken/Library observed exactly 8-9 '+ path' log lines at any sampled instant, newest-first, with long nested absolute paths wrapping via word-break rather than overflowing the box"
        status: pass
    human_judgment: false
  - id: D3
    description: "Escape, the header x, and a scrim click each cancel a running scan (foreground) -- the underlying context is cancelled, nothing is written, and the panel exit-animates closed exactly like any other close; no confirmation dialog anywhere in the flow (CRT-09)"
    requirement: CRT-09
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- three separate live scans against /Users/ken/Library (99GB, guaranteed not to finish in the test window) were cancelled one each via Escape, the header x button, and a scrim click; in every case the panel exit-animated closed (ws-create-panel-exit observed mid-transition), the status-bar segment disappeared, and `ls /tmp/storcat-catalogs` after each test showed no new/partial library.json or library.html -- confirmed by diffing directory contents before and after"
        status: pass
    human_judgment: false
  - id: D4
    description: "'Run in background' is not a cancel path: it closes the panel via the same exit while the scan keeps running in Go, updating the status-bar segment live, and the scan reaches a real completed done state with real written files (CRT-08)"
    requirement: CRT-08
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- clicked 'Run in background' on a real ~260MB node_modules scan; the panel closed while '1 catalogs' -> new catalog eventually appeared in the rail once the scan actually completed and wrote 24171 files / 184.9M to disk; a second, larger synthetic-volume scan was backgrounded, watched to completion (0% through 66% then straight to a real done state, 443525 files / 15.9G written), confirming the walk is genuinely unaffected by the panel's mount state"
        status: pass
    human_judgment: false
  - id: D5
    description: "With no scan running the status bar shows exactly its pre-phase 3 segments (no 4th, not an empty reserved slot); once a scan is backgrounded a 4th right-aligned segment appears reading the locked '● scanning {name} - {pct}%' copy (or '- counting...' during the indeterminate sub-state), computed through the same scanPercent helper the panel uses so the two numbers can never disagree; clicking the segment reopens the panel directly into its live scanning state, never the form"
    requirement: CRT-08
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- DOM query confirmed `.ws-status` has exactly one child (`.ws-status-left`) with no scan running; backgrounding a counting-substate scan produced the exact text '● dev - counting...'; backgrounding a percentage-known scan produced '● synthetic-volume-test - 34%' matching the panel's own last-shown '34%' exactly at the moment of backgrounding; clicking the segment reopened the panel with headerTitle 'Cataloguing volume' / headerStep 'step 2 of 2' / percentage '48%' (not the form's 'New catalog' / 'step 1 of 2') and the isForm-only volume-list DOM node absent"
        status: pass
    human_judgment: false
  - id: D6
    description: "go build/test, tsc --noEmit and npm run build all green; cli/create.go and go.mod/go.sum/frontend/package(-lock).json are untouched"
    verification:
      - kind: other
        ref: "go build ./... && go test ./... -race -count=1 -- all packages pass; cd frontend && npx tsc --noEmit && npm run build -- both exit 0; git diff --stat -- cli/create.go go.mod go.sum frontend/package.json frontend/package-lock.json -- empty; git status --short clean after the two task commits"
        status: pass
    human_judgment: false

duration: 35min
completed: 2026-08-14
status: complete
---

# Phase 25 Plan 6: Live Progress, Cancellation, and the Background Handoff Summary

**The running scan is now fully legible in both sub-states with real numbers throughout (no spinner, no fabricated percentage), every applicable close path actually cancels the walk and writes nothing, and a backgrounded scan stays visible and retrievable from the status bar's first-ever right-aligned segment.**

## Performance

- **Duration:** 35 min
- **Started:** 2026-08-14T20:10:00-05:00 (approximate, immediately following plan 25-05's completion)
- **Completed:** 2026-08-14T20:45:00-05:00 (Task 3 commit)
- **Tasks:** 3
- **Files modified:** 6 (1 created, 5 modified)

## Accomplishments

- `ScanningBody.tsx` renders both scanning sub-states from one component/state object: counting shows a live incrementing file count with no progress track and muted "Counting…" copy; percentage-known shows the real progress bar, byte total, and computed ETA in accent color; the transition happens in place with no remount and no reset to zero
- The log box (`.ws-create-log`) retains at most 9 newest-first `+ /path` lines, capped both in `AppContext`'s reducer (state itself never grows unbounded, T-25-23) and defensively at render — verified live against a scan producing hundreds of thousands of files without ever exceeding 9 visible lines
- `CreateSlideOver`'s five close paths route through one `handleCloseRequest(reason: CloseReason)`: Escape, the header ×, a scrim click, and "Discard and close" cancel the scan via `wailsAPI.cancelScan()` before running the standard exit when a scan is actually running; "Run in background" closes without cancelling — modeled as an explicit two-value argument, never inferred from state or trigger
- `handleCreate`'s failure branch now classifies a `StartScan` rejection via `classifyScanFailure`: a cancellation resets to idle with no error UI at all; only a source loss transitions to the `error` member (rendered fully by plan 25-07)
- `StatusBar.tsx` gained its first-ever right-aligned segment, rendered only while a scan is running (either sub-state), reading the locked `● scanning {name} · {pct}%` copy (or `· counting…`) through the same `scanPercent` helper the panel uses so the two can never disagree; clicking it dispatches `SET_CREATE_OPEN: true`, reopening the panel directly into its live scanning state
- Fixed the prerequisite for all of the above: `CreateSlideOver`'s `scan:progress` subscription was gated on `isOpen`, meaning a backgrounded scan's live updates stopped flowing into `AppContext` (and therefore the status bar) the instant the panel closed. Un-gated it — `CreateSlideOver` is already permanently mounted by `WorkspaceShell`, so this one-line change was the whole fix; no separate always-mounted subscriber was needed (documented as a Rule 2 auto-add: CRT-08 is unimplementable without it)

## Task Commits

Each task was committed atomically, with Task 1 and Task 2 landing together (see Deviations):

1. **Tasks 1+2: Scanning body (both sub-states) + cancellation wiring** - `d3297dd6` (feat)
2. **Task 3: Status bar background-scan segment** - `fcc6d219` (feat)

## Files Created/Modified

- `frontend/src/components/workspace/create/ScanningBody.tsx` (new) — both scanning sub-states, capped log, run-in-background footer
- `frontend/src/components/workspace/CreateSlideOver.tsx` — `handleCloseRequest`, `classifyScanFailure` wiring, always-on progress subscription, per-state body containers (isForm/isScanning/isDone each own their correct top-level wrapper instead of sharing one)
- `frontend/src/components/workspace/StatusBar.tsx` — `.ws-status-left` grouping + the conditional `.ws-status-scan` segment
- `frontend/src/contexts/AppContext.tsx` — `currentPath`/`log` added to the `counting`/`scanning` state variants and their reducer cases, with a 9-line retention cap
- `frontend/src/types/scan.ts` — `currentPath`/`log` fields on `ScanState`; new `CloseReason` union
- `frontend/src/workspace.css` — `.ws-create-scan-body` (now the full 22px/22px scanning-body container), `.ws-create-scan-pct-counting`, `.ws-create-walking*`, `.ws-create-log*`, `.ws-create-scan-footer`, `.ws-create-btn-outline`, `ws-create-pulse` keyframe + reduced-motion entry, `.ws-status` distribution + `.ws-status-left`/`.ws-status-scan*`

## Decisions Made

- CreateSlideOver's `scan:progress` subscription changed from `isOpen`-gated to always-on rather than adding a second always-mounted subscriber — the component was already permanently mounted, so this was the minimal correct fix (see key-decisions in frontmatter)
- The scanning body's outer container now owns the full 22px padding / 22px gap the UI-SPEC specifies for Step 2, instead of being double-nested inside the form step's 18px/16px `.ws-create-body` wrapper — JSX restructured so each of isForm/isScanning/isDone renders its own correctly-styled top-level container
- `handleCloseRequest`'s cancellation keys on an explicit `reason` argument, not on state or trigger identity — the only way "Run in background" can be guaranteed to never accidentally cancel

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] `scan:progress` subscription un-gated from `isOpen`**
- **Found during:** Task 3 (implementing the background handoff)
- **Issue:** `CreateSlideOver`'s progress subscription only ran `if (isOpen)`, so a scan sent to the background via "Run in background" stopped updating `AppContext`'s lifted scan state the instant the panel closed — CRT-08's status-bar segment would have frozen at whatever percentage was last shown, never reaching 100% or agreeing with reality.
- **Fix:** Removed the `isOpen` gate; the effect now subscribes unconditionally, relying on `CreateSlideOver` already being permanently mounted (25-01's established pattern).
- **Files modified:** `frontend/src/components/workspace/CreateSlideOver.tsx`
- **Verification:** Live dev-browser test: backgrounded a real multi-minute scan and watched the status-bar segment's percentage advance from 0% through 66% before the scan completed, matching the panel's own last-shown value exactly at every checkpoint.
- **Committed in:** `d3297dd6` (Task 1+2 commit)

---

**Total deviations:** 1 auto-fixed (Rule 2, required for CRT-08 to function at all).

**Process note (not a deviation, a commit-boundary slip):** Task 3's `.ws-status-left`/`.ws-status-scan*` CSS rules were intended for a separate commit alongside `StatusBar.tsx`, but a `git stash pop` merge during the split-commit staging process combined them into `d3297dd6` (the Task 1+2 commit) instead. The code itself is complete and correctly scoped to its task; only the commit boundary is imprecise. `fcc6d219` (Task 3) contains only `StatusBar.tsx`.

## Issues Encountered

**Selecting a plain folder via "…or choose any folder" cannot be automated through CDP** — same limitation `25-01-SUMMARY.md` and `25-05-SUMMARY.md` already documented (a native macOS `NSOpenPanel` is invisible to Chrome DevTools Protocol). Per the standing prohibition on host-OS GUI automation, this plan instead overrode `window.electronAPI.selectDirectory`/`listVolumes` in-page (staying entirely inside the webview, matching the standing prohibition's own guidance to "call the binding directly in the live webview") to exercise the counting sub-state, the volume-preselected percentage-known sub-state, and the background handoff against real, large, local directories (`frontend/node_modules`, `~/dev`, `~/Library`) rather than against a real external volume or a real user-driven folder pick. No true external volume with substantial readable content exists on this machine (`pi-downloader`/`software` are both `d--x--x--x` and produced instant `0 files · 0B` catalogs rather than a sustained percentage-known scan) — this is recorded as expected, out-of-scope hardware unavailability, not a defect.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- CRT-07 (live progress, both sub-states), CRT-08 (background handoff + status bar), and CRT-09 (cancellation, no confirmation) are implemented and live-verified against real, large-scale scans on this machine, including full completion in two separate end-to-end runs (24171 files/184.9M and 443525 files/15.9G, both correctly written and closed out to a real `done` state)
- Waves 1–5 re-confirmed unbroken live: the 340ms/260ms animations, `useModalBehavior` scroll-lock/focus-restore, volume cards with real `/Volumes` data, the folder alternative, title/root independence, the WILL WRITE preview, and the three toggles all still work
- `error` status continues to render as the existing form-step placeholder (an inline banner atop the volume picker) established in `25-01-SUMMARY.md`'s decision — the full CRT-10 error body (round badge, retry, write-partial-catalog) is plan 25-07's job and was intentionally not built here; a backgrounded source-loss failure does correctly clear the status-bar segment and land the reopened panel on `scan.status === 'error'`, but plan 25-07's richer body is what will make that landing visually distinct from the form
- `ScanningBody`'s `scan` prop type (`Extract<ScanState, {status:'counting'}|{status:'scanning'}>`), `CloseReason`, and `classifyScanFailure` are stable exported surfaces ready for plan 25-07 to extend without re-deriving scan-state narrowing

---
*Phase: 25-create-slide-over-progress-cancellation-partial-catalog*
*Completed: 2026-08-14*

## Self-Check: PASSED
