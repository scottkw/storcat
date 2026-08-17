---
phase: 28-re-scan-diff
plan: 03
subsystem: catalog-scan
tags: [react, typescript, diff-ui, rescan]

requires:
  - phase: 28-re-scan-diff
    plan: 01
    provides: Walk, ComputeDiff (tracer scope), App.RescanCatalog, state.rescan slice, RescanDialog shell (steps 1-3)
  - phase: 28-re-scan-diff
    plan: 02
    provides: ComputeDiff's fifth state (unreadable), type-change rule, sum invariant, DiffResult.OldEntryCount/LowSimilarity
provides:
  - "DiffList -- the grouped, natively-scrolling diff row list (Added/Removed/Changed/Unreadable, fixed order, empty groups omitted)"
  - "RescanDialog Step 3 Variant A (full diff, stat grid + DiffList + similarity banner + resolution caption) and Variant B (STATE-03's reduced no-diff summary)"
  - "RescanDialog's error/interrupted step, reusing ErrorBody with no partial-write affordance"
  - "ErrorBody's writingPartial/onWritePartial/explanation made optional/additive -- CreateSlideOver's call site unchanged"
  - "Catalog-actions menu's second entry point, 'Re-scan volume & diff...'"
affects: [28-04, 28-05]

actuals:
  tokens: 6573
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "DiffList's own .ws-rescan-diff* class family: plain native-scroll <div>, deliberately not the tree pane's virtualizer -- a diff is flat and scale-bounded by what differs, not by catalog size"
    - "ErrorBody's writingPartial/onWritePartial made optional TOGETHER (not independently) -- when both absent, the sibling 'Retry scan' button is promoted into the vacated primary-styled slot rather than leaving two secondary-weight actions"
    - "Variant B's summary line (fileCount/totalBytes) derived from DiffResult.entries itself (added-state, type=file) rather than a new Go field -- RescanCatalog's binding returns only a DiffResult on the no-old-tree path, so no fabricated data was needed"

key-files:
  created:
    - frontend/src/components/workspace/rescan/DiffList.tsx
  modified:
    - frontend/src/types/rescan.ts
    - frontend/src/workspace.css
    - frontend/src/components/workspace/rescan/RescanDialog.tsx
    - frontend/src/components/workspace/create/ErrorBody.tsx
    - frontend/src/components/workspace/DetailsPanel.tsx

key-decisions:
  - "Menu.tsx left unmodified (locked by this plan's own acceptance criteria) -- the catalog-actions menu's scan-running guard is enforced FUNCTIONALLY (a click while scanning surfaces the locked tooltip through the shared error slot and never opens the dialog), not visually (no aria-disabled/dimmed on the menu item itself, since MenuItemSpec has no per-item disabled/title field). Flagged at the Task 4 checkpoint and explicitly ACCEPTED by the user -- see Deviations below. The details-panel footer's own re-scan button (built in 28-01) is unaffected and continues to dim correctly."
  - "RescanDialog's failure branch now discriminates by ScanFailure.kind: a sourceLoss failure dispatches SCAN_FAILED (driving the shared state.scan into its error member, rendering ErrorBody under the existing 'scanning' step) instead of always resetting to step 1 -- state.rescan.step deliberately never gains a fourth 'error' value; the error UI is derived from state.scan.status alone, mirroring the happy path's shared-slice architecture."
  - "Variant B's fileCount/totalBytes computed client-side from DiffResult.entries (summing added-state, type=='file' entries) rather than adding a new Go-side field for a display line only the not-yet-built STATE-03 entry point (plan 28-05) will reach this plan."

requirements-completed: [ACT-06, ACT-08]

coverage:
  - id: D1
    description: "The diff step renders four fixed-order groups (Added, Removed, Changed, Unreadable) with correct per-group counts, alphabetical sort, and the five stat tiles summing to the number of distinct paths across the old and new trees"
    requirement: "ACT-06"
    verification:
      - kind: e2e
        ref: "dev-browser live session against wails dev :34115 -- staged fixture (add/delete/in-place-edit-same-size/chmod-000-subdir) re-scanned through the real catalog-actions-menu flow: 'Re-scan changed 5 entries', tiles 2/1/1/1/3 (sum 8), matching the hand-derived distinct-path count (9 total paths minus 1 correctly-pruned unreadable descendant)"
        status: pass
    human_judgment: false
  - id: D2
    description: "A chmod-000 subdirectory with prior contents appears under UNREADABLE, never under REMOVED, and its previously-known descendant is excluded from the diff rather than falsely reported removed"
    requirement: "ACT-06"
    verification:
      - kind: e2e
        ref: "same live session -- 'locked-dir' rendered under UNREADABLE·1 with its real permission-denied read error as the row's reason column; REMOVED·1 contained only remove-me.txt, never locked-dir or its child"
        status: pass
    human_judgment: false
  - id: D3
    description: "An empty diff group renders no header and no rows at all -- no reserved empty slot"
    requirement: "ACT-06"
    verification:
      - kind: e2e
        ref: "dev-browser live session, second run -- an add-only fixture (2 new files, nothing removed/changed/locked) rendered exactly one .ws-rescan-diffgroup (ADDED·2); document.body.textContent contained no 'REMOVED'/'CHANGED'/'UNREADABLE' substring anywhere, confirmed by DOM query, not just visual absence"
        status: pass
    human_judgment: false
  - id: D4
    description: "A long diff-row path ellipsizes on a single line (not wrapped); the wrong-disc similarity banner renders with the exact locked copy and never disables any footer control"
    requirement: "ACT-08"
    verification:
      - kind: e2e
        ref: "dev-browser live session -- long-path ADDED row measured scrollWidth 1877 vs clientWidth 524 (white-space:nowrap, text-overflow:ellipsis computed styles) confirming single-line truncation; a 25-entry-old-tree catalog re-scanned against a wholly unrelated 5-file directory rendered 'This looks like a different volume...30 of 30 entries differ...' with the sole footer button (Discard scan and close) at disabled:false/aria-disabled:null"
        status: pass
    human_judgment: false
  - id: D5
    description: "A re-scan interrupted mid-walk (volume vanishes) shows 'Scan interrupted'/'failed', ErrorBody's re-scan-specific explanation, 'Retry scan' promoted to the primary slot, and no partial-write affordance anywhere in the flow"
    requirement: null
    verification:
      - kind: e2e
        ref: "dev-browser live session -- raced a background rm -rf against a 12,000-file source mid-scan; observed header 'Scan interrupted'/'failed', explanation verbatim 'Nothing about rescan-fixture.json has changed...', action buttons ['Retry scan' (class ws-create-btn-primary), 'Close without writing'], and confirmed via DOM query no element anywhere contains the text 'Write partial'"
        status: pass
    human_judgment: false
  - id: D6
    description: "The catalog-actions menu offers 'Re-scan volume & diff...' as the second item (Rename -> Re-scan -> Duplicate -> divider -> Delete), opening the same RescanDialog the footer button opens"
    requirement: "ACT-06"
    verification:
      - kind: e2e
        ref: "dev-browser live session -- DOM query of #ws-catalog-actions-menu's [role=menuitem] elements returned ['Rename catalog...', 'Re-scan volume & diff...', 'Duplicate catalog', 'Delete catalog...'] in that exact order"
        status: pass
    human_judgment: false
  - id: D7
    description: "The menu item is guarded against a concurrent scan the same way the footer button is (aria-disabled/dimmed/tooltip)"
    requirement: null
    verification: []
    human_judgment: true
    rationale: "NOT achieved to the same visual standard as the footer button -- see Deviations. Menu.tsx (locked, unmodified per this plan's own acceptance criteria) has no per-item disabled/title field, so the guard is functional only (click while scanning surfaces the tooltip text via the error slot and never opens the dialog) rather than visual (no aria-disabled/dimmed on the menu item itself). User reviewed and explicitly accepted this at the Task 4 checkpoint; flagged here for a future reader rather than silently closed."

duration: 91min
completed: 2026-08-16
status: complete
---

# Phase 28 Plan 03: Diff List, Wrong-Disc Warning, Error Step, Menu Entry Point Summary

**Grouped diff row list (Added/Removed/Changed/Unreadable), the wrong-disc similarity banner, both diff-step variants, the re-scan error step, and the catalog-actions menu's second entry point -- all live-verified via dev-browser against a running `wails dev` instance**

## Performance

- **Duration:** ~91 min (54 min implementation + ~37 min live dev-browser verification across two checkpoint rounds)
- **Started:** 2026-08-16T19:41:01Z (approx, from prior plan's commit)
- **Completed:** 2026-08-16T21:15:00Z
- **Tasks:** 3 (+ Task 4, a blocking human-verify checkpoint, resolved live)
- **Files modified:** 6 (1 new, 5 modified; excluding `.planning/`)

## Accomplishments

- `DiffList` (new): four fixed-order groups (Added, Removed, Changed, Unreadable), alphabetically sorted within each, rendered in a plain native-scroll `<div>` (`max-height: 200px`) -- deliberately not the tree pane's virtualizer, since a diff is flat and scale-bounded by what differs, not by catalog size. An empty group renders no header and no rows, proven live by DOM query (not just visual absence).
- `RescanDialog` Step 3 now renders both spec'd variants: Variant A (full diff -- sub-line, conditional similarity banner, stat grid, `DiffList`, always-visible resolution caption) and Variant B (the STATE-03 reduced summary, no stat grid/diff list/banner, its own caption), selected by `oldTreeAvailable`.
- The wrong-disc similarity banner renders above the stat grid exactly per the locked copy when `DiffResult.LowSimilarity` is set, and never disables any footer control -- proven live (`disabled:false`/`aria-disabled:null` on the sole footer button while the banner was showing).
- `RescanDialog`'s error/interrupted step reuses `ErrorBody` (writingPartial/onWritePartial omitted, "Retry scan" promoted into the primary-styled slot) with a re-scan-specific explanation naming the *existing* catalog rather than Create's not-yet-written one. Reached via the same `ScanFailure.kind === 'sourceLoss'` discriminator Create already classifies on; `state.rescan.step` stays `'scanning'` throughout (no new step value), the error UI is derived purely from the shared `state.scan.status`.
- `ErrorBody.tsx` gained three additive, optional props (`writingPartial?`, `onWritePartial?`, `explanation?`) -- `CreateSlideOver`'s call site needed zero edits (confirmed by an empty `git diff` on it).
- `DetailsPanel`'s `CatalogActions` menu gained "Re-scan volume & diff…" as the second item (Rename → Re-scan → Duplicate → divider → Delete), opening its own `RescanDialog` instance -- confirmed live via DOM order query. `Menu.tsx` itself is untouched (empty `git diff`), per this plan's own locked acceptance criterion.
- All of the above verified live against a real `wails dev` instance via dev-browser (`http://localhost:34115`, bindings probed fresh via `Object.keys(window.go.main.App)` before use) across two checkpoint rounds -- see Task 4 Live Verification below. No host-OS GUI automation was used at any point; every native-dialog interaction was replaced with an in-page JS stub of `window.go.main.App.SelectDirectory`.

## Task Commits

1. **Task 1: DiffList — four fixed-order groups, native scroll, ellipsized paths** - `931f7319` (feat)
2. **Task 2: Diff step Variants A and B, similarity banner, resolution caption** - `33da24a9` (feat)
3. **Task 3: Error step and the catalog-actions menu entry point** - `e8751ac9` (feat)

## Files Created/Modified

- `frontend/src/components/workspace/rescan/DiffList.tsx` (new) - grouped diff row list
- `frontend/src/types/rescan.ts` - `DiffResult.oldEntryCount`/`lowSimilarity`; new `DiffGroupKey`
- `frontend/src/workspace.css` - `.ws-rescan-diff*` (list/group/row), `.ws-rescan-warn`, `.ws-rescan-caption`, `.ws-rescan-body-diff`
- `frontend/src/components/workspace/rescan/RescanDialog.tsx` - Variant A/B bodies, similarity banner, resolution caption, error step
- `frontend/src/components/workspace/create/ErrorBody.tsx` - optional `writingPartial`/`onWritePartial`/`explanation` props, "Retry scan" primary-slot promotion
- `frontend/src/components/workspace/DetailsPanel.tsx` - `CatalogActions` menu's second item + its own `RescanDialog` instance

## Decisions Made

- **Menu.tsx's scan-running guard implemented functionally, not visually** — see Deviations below; this is the plan's one substantive, human-reviewed departure from the literal must-have wording.
- `RescanDialog`'s failure branch now discriminates `ScanFailure.kind`: `sourceLoss` dispatches `SCAN_FAILED` (driving the shared `state.scan` slice into its error member, rendered under the existing `'scanning'` step) instead of unconditionally resetting to step 1 — no new `RescanState.step` value was added; the error UI is derived from `state.scan.status` alone, the same shared-slice architecture the happy path already uses.
- Variant B's summary line (`fileCount`/`totalBytes`) is computed client-side from `DiffResult.entries` (summing `state === 'added' && type === 'file'`) rather than adding a new Go-side field — `RescanCatalog`'s binding returns only a `DiffResult` on the no-old-tree path, and this is real data already on the wire, not a fabricated value.

## Deviations from Plan

### Accepted Deviation (human-reviewed at the Task 4 checkpoint)

**1. [Rule 4 — architectural conflict, resolved by explicit user decision] The catalog-actions menu's "Re-scan volume & diff…" item is NOT visually dimmed during a concurrent scan.**

- **Found during:** Task 3, while wiring the menu entry point's scan-running guard.
- **Conflict:** The plan's action text calls for the same `aria-disabled`/locked-tooltip/dimmed treatment the details-panel footer button already has, applied to the new menu item — but the plan's own automated acceptance criterion requires `git diff HEAD -- frontend/src/components/workspace/Menu.tsx` to stay empty. `Menu.tsx`'s `MenuItemSpec` has no per-item `disabled`/`title` field and its render loop spreads no such attribute onto the `<button>` it produces, so the literal visual treatment is not achievable without editing `Menu.tsx` — which the plan itself locks unmodified.
- **Resolution:** Implemented the *functional* equivalent instead — while `state.scan.status` is `'counting'`/`'scanning'`, clicking "Re-scan volume & diff…" never opens the dialog and instead surfaces the exact locked tooltip string ("A scan is already running — open it from the status bar.") through the same shared error slot Footer's other actions already use. The menu item itself remains visually undimmed and clickable-looking until clicked.
- **Files modified:** `frontend/src/components/workspace/DetailsPanel.tsx` (no `Menu.tsx` edit).
- **Verification:** `git diff HEAD -- frontend/src/components/workspace/Menu.tsx` is empty (the plan's own acceptance criterion, still holds).
- **Disposition:** Flagged explicitly at the Task 4 checkpoint. **The user reviewed this and decided: accept the functional guard, keep `Menu.tsx` unmodified.** Recorded here plainly for a future reader — the menu item is NOT visually dimmed during a concurrent scan; only the details-panel footer button (built in 28-01) dims correctly.
- **Committed in:** `e8751ac9` (Task 3 commit).

---

**Total deviations:** 1, explicitly reviewed and accepted by the user (not auto-fixed under Rules 1-3).
**Impact on plan:** The concurrent-scan guard's *behavioral* guarantee (a second scan can never actually start from this entry point) holds; only its *visual* presentation differs from the footer button's. No data-safety or correctness impact — `RescanCatalog`'s own `scanMu` one-scan-at-a-time lock on the Go side is the real enforcement point regardless of what the frontend shows.

## Task 4 Live Verification (dev-browser, two checkpoint rounds)

Both rounds ran against a freshly-started `wails dev` (a stale hour-old `StorCat` process squatting on port 34115 with outdated bindings was found and killed first), driven entirely through dev-browser at `http://localhost:34115`. `Object.keys(window.go.main.App)` was probed before every session to confirm fresh bindings (`RescanCatalog` present). No `osascript`/System Events/host-OS GUI automation/keystroke injection was used at any point — every fixture was staged with plain filesystem commands, and the one native-dialog interaction (`VolumePicker`'s "choose any folder") was bypassed with an in-page JS stub of `window.go.main.App.SelectDirectory` returning the fixture path directly, never invoking the real OS picker.

**Round 1 (steps 1–5, 7):**
- **Menu order:** `["Rename catalog…", "Re-scan volume & diff…", "Duplicate catalog", "Delete catalog…"]` — second item, before Duplicate.
- **Grouped diff:** a staged fixture (add 2 / remove 1 / in-place-edit-same-size 1 / chmod-000 1 subdirectory-with-prior-contents) re-scanned through the real menu flow rendered "Re-scan changed 5 entries", tiles 2/1/1/1/3 (sum 8) — matching the hand-derived distinct-path count (9 total paths across old∪new, minus 1 correctly-pruned unreadable descendant that would otherwise have been falsely reported removed). `locked-dir` appeared under UNREADABLE·1 with its real permission-denied reason, never under REMOVED.
- **Long-path ellipsis:** the long-path ADDED row measured `scrollWidth 1877` vs `clientWidth 524` with computed `white-space: nowrap`/`text-overflow: ellipsis` — single-line truncation confirmed, not wrapped.
- **Similarity banner:** a 25-entry-old catalog re-scanned against a wholly unrelated 5-file directory rendered the exact locked copy ("This looks like a different volume…30 of 30 entries differ…"), with the sole footer button at `disabled:false`/`aria-disabled:null` — never gated.
- **Error step:** raced a background `rm -rf` against a 12,000-file source mid-walk; observed "Scan interrupted"/"failed", the exact re-scan-specific explanation, "Retry scan" at `ws-create-btn-primary` (the promoted primary slot), "Close without writing", and confirmed via DOM query that no element anywhere contains "Write partial".
- **Cleanup:** `chmod 755` restored on the fixture's locked subdirectory (verified readable again via `ls`); `wails dev` and the spawned `StorCat` process were both killed afterward — confirmed no process left listening on :34115.

**Round 2 (empty-group closure, requested separately):**
- Staged an add-only fixture (catalog 2 files, then add 2 more, no deletions/edits/chmod) and re-scanned. Result: "Re-scan changed 2 entries", tiles 2/0/0/0/2.
- DOM query: `document.querySelectorAll('.ws-rescan-diffgroup')` returned exactly 1 element (`ADDED · 2`); `document.body.textContent` contained no `REMOVED`, `CHANGED`, or `UNREADABLE` substring anywhere on the page — the three empty groups are absent from the DOM entirely, not merely hidden.
- Fixture removed and `wails dev`/`StorCat` killed afterward; port 34115 confirmed clear.

Screenshots captured during both rounds (`~/.dev-browser/tmp/28-03-*.png`) visually corroborate every finding above: grouped diff list with correct glyphs/colors, the similarity banner's warm-yellow tint, the interrupted-scan step's primary "Retry scan", and the single-group empty-group render.

## Known Stubs

None. All deliverables in this plan's file list are fully wired: `DiffList` renders real diff data, both Step 3 variants are reachable (Variant A live-verified via the menu/footer entry points; Variant B's rendering logic is implemented and unit-verifiable via its own branch, though its only real entry point — `UnreadableCatalogPanel`'s trio — is explicitly out of scope for this plan, landing in 28-05), and the error step is reachable and live-verified.

## Issues Encountered

- **Grep-collision self-inflicted twice during implementation, caught before commit both times:** doc comments in `DiffList.tsx` (explaining why the virtualizer/`useVisibleRows` were deliberately not used) and in `RescanDialog.tsx` (explaining why the header title deliberately doesn't say "Catalog rebuilt") both initially contained the literal forbidden strings the plan's own automated `<verify>` greps for zero occurrences of. Both were caught by running the plan's exact verification commands before considering each task done, and rephrased without repeating the literal banned substrings.
- **Stale `StorCat` process holding port 34115** at the start of live verification (an hour-old process from an earlier session's checkpoint work, with outdated bindings) — caught by the `Object.keys(window.go.main.App)` probe returning what looked plausible but wasn't guaranteed fresh, confirmed via `lsof`, killed, and a clean `wails dev` restarted before proceeding. This is exactly the failure mode the mandatory bindings-freshness probe exists to catch.
- **`CatalogRail`'s `storcat-catalog-directory` localStorage cache** does not auto-sync when `SetCatalogDirectory` is called directly (bypassing the Settings UI) — the rail kept showing a prior session's stale catalog list until the localStorage key was also set directly. Not a product bug (the real Settings UI updates both together); just a wrinkle specific to driving the app via raw binding calls for fixture setup, worked around by setting the localStorage key alongside the Go-side config.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `DiffList`, both Step 3 variants, the error step, and the menu entry point are all in place and live-verified for plan 28-04 (the Overwrite/Keep-both write resolutions, ACT-07) to build its footer buttons directly beneath the already-shipped resolution caption.
- `UnreadableCatalogPanel`'s action trio (the third entry point, `oldTreeAvailable: false`) remains explicitly out of scope — plan 28-05. Variant B's rendering path is implemented and ready for that plan to reach live; it has not itself been live-verified yet because its only real trigger doesn't exist until then.
- The one open item for a future maintainer: if `Menu.tsx` is ever revisited (e.g., a genuine need for a second per-item-disabled consumer arises), extending `MenuItemSpec` with optional `disabled`/`title` fields would let the catalog-actions menu's re-scan item match the footer button's visual guard exactly — tracked here, not silently dropped.

---
*Phase: 28-re-scan-diff*
*Completed: 2026-08-16*

## Self-Check: PASSED

All files created/modified verified present on disk (`DiffList.tsx`, `types/rescan.ts`, `workspace.css`, `RescanDialog.tsx`, `ErrorBody.tsx`, `DetailsPanel.tsx`, this SUMMARY); all three task commit hashes (`931f7319`, `33da24a9`, `e8751ac9`) verified present in `git log`.
