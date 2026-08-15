---
phase: 25-create-slide-over-progress-cancellation-partial-catalog
plan: 05
subsystem: ui
tags: [react, wails, volume-picker, form-validation, toggle-switch, localStorage]

# Dependency graph
requires:
  - phase: 25-01
    provides: "CreateSlideOver.tsx shell, lifted AppContext.scan state machine, StartScan/wailsAPI.startScan wiring this plan extends in place"
  - phase: 25-04
    provides: "App.ListVolumes binding + wailsAPI.listVolumes, ScanOptions.TotalBytesHint seam this plan fills from the selected volume card"
provides:
  - "VolumePicker.tsx -- volume cards (name/mount path/size/status tag), chosen-folder pseudo-card, staleness-guarded enumeration, zero-volumes empty state, first-volume preselection"
  - "CreateForm.tsx -- independent title/filename-root fields via native-placeholder-only defaulting, and the live, deterministically-ordered WILL WRITE preview with an already-exists qualifier"
  - "OptionsToggles.tsx -- hand-built toggle-switch control (write-HTML/copy-to-secondary/include-hidden), the secondary-location bootstrap (open-dialog-once, reuse-persisted, off-never-clears)"
  - "ScanSource type + basename/sourcePathOf/sourceDisplayNameOf (types/scan.ts) -- the shared volume-or-folder shape every create sub-component derives from"
  - "scanFormat.ts: slugifyRoot(), willWritePaths() -- pure, order-fixed preview-path derivation"
affects: [25-06, 25-07]

# Actuals (#2632)
actuals:
  tokens: 8500
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Native HTML input placeholder as the sole default-application mechanism (never a written-then-possibly-overwritten value) -- implements CRT-04's 'defaults apply only while untouched' contract with zero extra touched-flag state, since an empty field's placeholder always reflects the *current* source by construction"
    - "Component-owned localStorage read/write for a persisted GUI-only setting (OptionsToggles' secondary-copy directory), reported upward via an onChange callback rather than the parent owning the read/write itself -- the parent's own initial useState(() => safeGetItem(...)) reads the same key once at mount so both copies start in agreement"
    - "Staleness-guarded fetch-on-mount (requestIdRef), copied from CommandPalette's precedent, applied to VolumePicker's own volume enumeration"

key-files:
  created:
    - frontend/src/components/workspace/create/VolumePicker.tsx
    - frontend/src/components/workspace/create/CreateForm.tsx
    - frontend/src/components/workspace/create/OptionsToggles.tsx
  modified:
    - frontend/src/components/workspace/CreateSlideOver.tsx
    - frontend/src/lib/scanFormat.ts
    - frontend/src/types/scan.ts
    - frontend/src/services/wailsAPI.ts
    - frontend/src/workspace.css

key-decisions:
  - "No separate 'touched' boolean per field: CRT-04's independence/defaulting contract is fully satisfied by native HTML placeholder semantics alone (placeholder shows only while the field is empty; typed text is never overwritten; a cleared field re-shows the *current* source's default). A literal touched-flag state was judged an unrequested abstraction over behavior the platform already provides for free."
  - "OptionsToggles disabled prop is wired to isScanning and passed correctly, but under this plan's current architecture the whole form step (including the toggles) unmounts the instant a scan starts (isForm gates the body), so the visibly-disabled-while-scanning state is unreachable today. Implemented anyway as the correct component-level contract, since 25-UI-SPEC's Background Handoff Contract (a later plan) will reopen this same form into a live scanning state."
  - "willWritePaths' secondary-copy rows are gated by the caller passing an empty string when the copy-to-secondary toggle is off, not by a second boolean parameter on the helper itself -- keeps the pure helper's contract to exactly 'a secondary directory string, present or not', matching its single already-declared optional parameter."

patterns-established:
  - "Placeholder-only default application for independent, never-required form fields (see tech-stack.patterns) -- reusable for any future GSD form pair needing the same 'derived default, no overwrite once edited' behavior."

requirements-completed: [CRT-02, CRT-03, CRT-04, CRT-05, CRT-06]

coverage:
  - id: D1
    description: "Volume cards render name/mount path/size/mounted-or-read-errors status per detected volume, the first is preselected, a read-errors volume stays selectable, and zero volumes renders no fabricated row while the choose-any-folder link stays the answer (CRT-02)"
    requirement: CRT-02
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- cards matched this machine's real /Volumes minus the boot symlink (pi-downloader, software, both correctly tagged 'read errors' per their d--x--x--x permissions); first card preselected; clicking the second card swapped selection; window.go.main.App.ListVolumes stubbed to return [] rendered zero cards with the choose-any-folder link still present and Create still disabled"
        status: pass
    human_judgment: false
  - id: D2
    description: "Title and filename-root fields are independent (editing one never overwrites the other), both may be left blank with no validation error, and both populate synchronously via placeholder the instant a source is picked (CRT-04 adjacency/empty/loading)"
    requirement: CRT-04
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- typing 'My Custom Title' left the root field's placeholder at 'pi-downloader' and its value empty; typing into root afterward left title's typed value untouched; Create stayed enabled throughout"
        status: pass
    human_judgment: false
  - id: D3
    description: "The WILL WRITE preview lists paths in the fixed json-then-html-then-secondary order, recomputes on every keystroke/toggle change, and flags a row whose file already exists in the configured catalog directory (CRT-04 ordering/partial/populated)"
    requirement: CRT-04
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- typing into root live-updated both preview rows on every keystroke; a real on-disk my-root.json was flagged '-- already exists' while the sibling my-root.html row (not on disk) was not; enabling copy-to-secondary appended two more rows in the same json-then-html order at the end"
        status: pass
    human_judgment: false
  - id: D4
    description: "All three toggles default correctly (write-HTML on, copy-to-secondary off, include-hidden off), flip and flip back, and the secondary-location value survives an off-then-on cycle without re-opening the dialog when already persisted (CRT-05 empty/idempotency)"
    requirement: CRT-05
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- initial toggle states matched the documented defaults with the write-HTML note reading the live '{root}.html'; write-HTML and include-hidden each flipped and flipped back; with storcat-secondary-directory pre-seeded, clicking the secondary toggle went straight to checked=true with the real path as its note (no dialog call, confirming the reuse-without-dialog branch); toggling off then on again restored the identical path"
        status: pass
    human_judgment: false
  - id: D5
    description: "A long secondary-copy path ellipsizes from the left inside the toggle row (direction:rtl technique) and never widens the panel (CRT-05/E4 long-text)"
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- computed style on .ws-create-note-path with a 100+ char path showed direction:rtl, overflow:hidden, textOverflow:ellipsis, whiteSpace:nowrap"
        status: pass
    human_judgment: false
  - id: D6
    description: "Create and the ⌘↵ shortcut share exactly one handler; a second activation while a scan is already running/submitting is a no-op (CRT-06 idempotency/concurrency)"
    requirement: CRT-06
    verification:
      - kind: automated_ui
        ref: "dev-browser :34115 -- pressed Cmd+Enter twice in immediate succession against a real preselected volume; exactly one scan ran through to a real done state ('pi-downloader catalogued'), never a second concurrent scan or a duplicated dispatch"
        status: pass
    human_judgment: false
  - id: D7
    description: "The two native-OS-dialog code paths (choose-any-folder, and the secondary-location toggle's first-activation dialog + cancel-reverts-to-off) are correct by inspection and by the reuse-without-dialog branch's proven behavior, but were not clicked through end to end"
    verification: []
    human_judgment: true
    rationale: "Chrome DevTools Protocol has no visibility into a native macOS NSOpenPanel (same limitation 25-01-SUMMARY.md documented and this plan's phase-specific instructions name explicitly). The guard logic (call selectDirectory only when no path is persisted; on cancel, result.success/result.path is falsy and the function returns without dispatching a state change) was verified by code inspection and indirectly proven by D4's reuse-without-dialog test (which confirms the *skip* branch fires correctly when a path already exists) -- but a human has not clicked 'choose any folder' or the secondary toggle's first activation and cancelled the resulting native dialog to confirm the revert-to-off path live."
  - id: D8
    description: "Volume names/mount paths containing non-ASCII characters, spaces, or emoji render correctly and reach the scanner unmangled (CRT-02 encoding)"
    verification: []
    human_judgment: true
    rationale: "No volume with a non-ASCII/emoji name exists on this development machine to test against. VolumePicker renders vol.name/vol.mountPath as plain React text nodes with no string transformation, encoding, or truncation applied before they reach wailsAPI.startScan's sourcePath argument -- correct by construction/inspection, not empirically exercised."
  - id: D9
    description: "go build/test, tsc --noEmit and npm run build all green; cli/create.go and go.mod/go.sum/frontend/package(-lock).json are untouched"
    verification:
      - kind: other
        ref: "go build ./... && go test ./... -race -count=1 -- all packages pass; cd frontend && npx tsc --noEmit && npm run build -- both exit 0; git diff --stat -- cli/create.go go.mod go.sum frontend/package.json frontend/package-lock.json -- empty"
        status: pass
    human_judgment: false

duration: 18min
completed: 2026-08-14
status: complete
---

# Phase 25 Plan 5: Volume Cards, Title/Root/Preview, and the Three Toggles Summary

**Turned the tracer's bare folder-only form into the real one: selectable volume cards with live size/status, independent title and filename-root fields backed by a deterministically-ordered live WILL WRITE preview, and the three creation toggles including a working secondary-location bootstrap -- all with zero new npm/Go dependencies.**

## Performance

- **Duration:** 18 min
- **Started:** 2026-08-14T19:52:00-05:00 (approximate, immediately following plan 25-04's completion)
- **Completed:** 2026-08-14T20:10:05-05:00 (Task 3 commit)
- **Tasks:** 3
- **Files modified:** 8 (3 created, 5 modified)

## Accomplishments
- `VolumePicker.tsx` renders one card per detected volume (chip/name/mount-path/size/status tag), preselects the first, keeps a read-errors volume selectable, and the choose-any-folder link is always present -- verified live against this machine's real `/Volumes` (`pi-downloader`/`software`, both correctly tagged "read errors")
- `CreateForm.tsx`: title and filename-root are truly independent fields via native placeholder-only defaulting (no separate touched-flag state needed), and the WILL WRITE preview recomputes live in the fixed json/html/secondary order with an "already exists" qualifier sourced from the rail's already-loaded catalog listing (no new binding)
- `OptionsToggles.tsx`: hand-built toggle-switch control (no package added), correct diverged defaults (write-HTML on, copy-to-secondary off, include-hidden off), and a working secondary-location bootstrap that opens the native dialog only on first activation, reuses a persisted path silently thereafter, and never clears it on an off toggle
- `CreateSlideOver.tsx` now threads the selected source's real total-bytes hint, the real option values, and the real secondary directory into `StartScan`'s arguments; the `⌘↵` shortcut is gated to the form step and shares `handleCreate` exactly with the Create button

## Task Commits

Each task was committed atomically:

1. **Task 1: Volume cards, the folder alternative, and selection** - `6a15e216` (feat)
2. **Task 2: Title, filename root, and the live WILL WRITE preview** - `d78826a3` (feat)
3. **Task 3: The three creation toggles and the keyboard start path** - `c7c7804e` (feat)

**Plan metadata:** _(pending final commit)_

## Files Created/Modified
- `frontend/src/components/workspace/create/VolumePicker.tsx` - volume cards, chosen-folder pseudo-card, staleness-guarded enumeration, zero-volumes empty state
- `frontend/src/components/workspace/create/CreateForm.tsx` - title/root fields, WILL WRITE preview, already-exists qualifier
- `frontend/src/components/workspace/create/OptionsToggles.tsx` - three toggle rows, secondary-location bootstrap, disabled-while-scanning contract
- `frontend/src/components/workspace/CreateSlideOver.tsx` - composes the three sub-components; scan-start args now carry title/root/output-dir/options/totalBytesHint; `⌘↵` gated to the form step
- `frontend/src/lib/scanFormat.ts` - `slugifyRoot()`, `willWritePaths()`
- `frontend/src/types/scan.ts` - `ScanSource`, `basename`, `sourcePathOf`, `sourceDisplayNameOf`
- `frontend/src/services/wailsAPI.ts` - `startScan`'s opts gained `totalBytesHint` (deviation, see below)
- `frontend/src/workspace.css` - `.ws-create-vol-*`, `.ws-create-grid`, `.ws-create-willwrite*`, `.ws-create-toggle-*`, `.ws-create-note-path`

## Decisions Made
- No separate "touched" boolean per field -- native HTML placeholder semantics alone satisfy CRT-04's independence/defaulting contract (see key-decisions in frontmatter for the full reasoning)
- `OptionsToggles`' `disabled` prop is correctly wired to `isScanning` even though the current form-step gating makes that state unreachable today -- a forward-compatible component contract, not dead code, ahead of the later background-handoff plan that will reopen this same form into a live scan
- `willWritePaths`' secondary rows are gated by the caller passing an empty string (not a second boolean parameter) when the toggle is off, keeping the pure helper's signature to exactly "a directory string, present or not"

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `wailsAPI.ts`'s `startScan` opts type extended with `totalBytesHint`**
- **Found during:** Task 2
- **Issue:** Task 2's own instruction explicitly requires wiring "the selected source's total-bytes hint" into the scan-start arguments. `app.go`'s `ScanOptions.TotalBytesHint` field already existed (added in plan 25-04), but `wailsAPI.ts`'s `startScan` wrapper's `opts` parameter type never declared it, so passing it from `CreateSlideOver` would fail TypeScript's excess-property check on the call site. `wailsAPI.ts` is not in this task's `files_modified` list.
- **Fix:** Added `totalBytesHint: number` to `startScan`'s `opts` type, matching `main.ScanOptions`'s already-generated shape.
- **Files modified:** `frontend/src/services/wailsAPI.ts`
- **Verification:** `npx tsc --noEmit` and `npm run build` both green; live-verified the hint reaches the binding (a preselected volume's known `totalBytes` flows through unmodified)
- **Committed in:** `d78826a3` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (blocking, required by the task's own written instruction).
**Impact on plan:** No architectural change -- a type-signature extension matching a field the Go side already had since plan 25-04.

## Issues Encountered

**Two code paths could not be driven end to end through CDP.** The "choose any folder" link and the secondary-location toggle's first-activation dialog both invoke the existing `SelectDirectory` binding, which opens a native macOS `NSOpenPanel` -- Chrome DevTools Protocol has no visibility into it, the same limitation 25-01-SUMMARY.md documented and this plan's phase-specific instructions named explicitly up front. Per the standing prohibition on host-OS GUI automation, no attempt was made to drive these dialogs via `osascript` or similar. Both guard branches (skip the dialog when a path is already persisted; treat a cancelled/failed result as "do nothing, no error") were verified by code inspection and indirectly proven live: pre-seeding `storcat-secondary-directory` and toggling on went straight to `checked=true` with the persisted path as the note and no dialog call, which is the same code path that would also skip a *second* activation after a real dialog resolves. Recorded honestly as coverage `D7` (`human_judgment: true`), not silently asserted as proven.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Every requirement this plan owned (CRT-02 through CRT-06) is implemented and live-verified against real data on this machine (real `/Volumes`, real toggle round-trips, a real end-to-end scan through `⌘↵` to a real done state)
- `ScanSource`/`basename`/`sourcePathOf`/`sourceDisplayNameOf` (types/scan.ts) and the `willWritePaths`/`slugifyRoot` helpers (scanFormat.ts) are stable, exported surfaces ready for plan 25-06 (scanning body/cancellation UI) and 25-07 (error/partial-catalog UI) to reuse without re-deriving source-path or preview logic
- **Two items need a human click-through, not because the code is unproven but because CDP cannot reach a native OS dialog** (coverage D7): confirm "choose any folder" opens the real picker and that cancelling the secondary-location toggle's first-activation dialog reverts it to off with no error banner
- **One item needs a human's real hardware, not this plan's logic** (coverage D8): confirm a non-ASCII/emoji-named external volume renders correctly in the card list on a machine that has one attached
- Waves 1-4 (slide-over animation, five close paths, `useModalBehavior` consumption, folder-to-catalog path, "Open in workspace") were re-confirmed live and unbroken by this plan's changes

---
*Phase: 25-create-slide-over-progress-cancellation-partial-catalog*
*Completed: 2026-08-14*

## Self-Check: PASSED

All 8 created/modified files verified present on disk; all three task commit hashes (`6a15e216`, `d78826a3`, `c7c7804e`) verified present in git history.
