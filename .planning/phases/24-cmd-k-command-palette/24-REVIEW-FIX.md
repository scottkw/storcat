---
phase: 24-cmd-k-command-palette
fixed_at: 2026-08-14T17:45:00Z
review_path: .planning/phases/24-cmd-k-command-palette/24-REVIEW.md
iteration: 1
findings_in_scope: 4
fixed: 4
skipped: 0
status: all_fixed
---

# Phase 24: Cmd-K Command Palette - Code Review Fix Report

**Fixed at:** 2026-08-14T17:45:00Z
**Source review:** `.planning/phases/24-cmd-k-command-palette/24-REVIEW.md`
**Iteration:** 1

**Scope:** Critical + Warning findings only (default scope). IN-01 and IN-02 (Info severity) were explicitly out of scope per the fix-launch instructions and were not touched — IN-02 in particular is a deliberate, spec-locked decision (`24-CONTEXT.md`/`24-UI-SPEC.md` E1) and must not be "fixed".

**Environment note:** `workflow.use_worktrees` is `false` in `.planning/config.json` for this project, so all edits, verification, and commits below ran directly in the main checkout (`/Users/ken/dev/storcat`) on branch `main` — no isolated worktree was created. The gate results below are reproducible from the current working tree.

**Summary:**
- Findings in scope: 4
- Fixed: 4
- Skipped: 0

## Fixed Issues

### WR-01: `SET_EXPANDED`'s merge-vs-replace safety is call-site discipline only, not enforced

**Files modified:** `frontend/src/contexts/AppContext.tsx`, `frontend/src/components/workspace/TreePane.tsx`, `frontend/src/lib/reveal.ts`
**Commit:** `bb97ce90`
**Applied fix:** Added a new `MERGE_EXPANDED: string[]` reducer case in `AppContext.tsx` that spreads `state.expanded` internally and can only ever add entries — it has no way to express "replace" semantics, so a future reveal-path caller cannot accidentally collapse every other open branch the way a raw `SET_EXPANDED` dispatch could. The case also returns the identical `state` object (not just an equal `expanded` map) when no ancestor needed adding, so React's reducer bail-out makes a repeat reveal of an already-visible node a true no-op — the same idempotence the old call-site `mergeExpanded()` reference check provided, now enforced inside the reducer instead of at the call site. `TreePane.tsx`'s reveal effect A now dispatches `{ type: 'MERGE_EXPANDED', payload: ancestors }` directly. `lib/reveal.ts#mergeExpanded` (now dead — its only caller was TreePane) was deleted rather than kept as a vestigial pure helper, along with the module doc comment's reference to it; `findNodeIndexByPath` and `ancestorPathsOf` are unchanged. `SET_EXPANDED` itself is untouched and still full-replace, used only by `BreadcrumbBar.tsx`'s `handleExpandAll`/`handleCollapse` — verified no other call sites exist via `grep -rn` across `frontend/src`.
**Verification:** `npx tsc --noEmit` clean at the WR-01-only checkpoint (before WR-02 was layered on) and again on the final tree. Live-reverified per constraint 7 (see below): pre-expanded unrelated branches (VOL02, VOL03) survived a reveal into a third, previously-collapsed branch (VOL01/100CANON).

### WR-02: `handleActivate`'s dispatch order is load-bearing but enforced only by a comment

**Files modified:** `frontend/src/contexts/AppContext.tsx`, `frontend/src/components/workspace/CommandPalette.tsx`
**Commit:** `12d00bf8`
**Applied fix:** Added a new `REVEAL_HIT: { catalogId: string; path: string }` reducer case that applies the catalog switch and sets `pendingReveal` in one atomic state update — matching the review's suggested shape, with one refinement over the literal snippet in the review body: the switch-only fields (`tree`, `expanded`, `selected`) are reset to fresh/cleared values **only when `catalogId` actually differs from `currentCatalogId`**; when the hit is already in the current catalog they pass through unchanged, so a same-catalog reveal doesn't blow away in-progress expansion/selection state the way an unconditional reset would have. `CommandPalette.tsx`'s `handleActivate` now dispatches a single `REVEAL_HIT` action instead of the previous `SELECT_CATALOG` then `SET_PENDING_REVEAL` pair — there is no longer a two-statement ordering for a future edit, early return, or merge conflict to get wrong. `SELECT_CATALOG` itself is untouched and still used by `CatalogRail.tsx` for rail-driven catalog switches (verified via grep — it is a genuinely separate caller with different semantics, not a duplicate of the new action).
**Verification:** `npx tsc --noEmit` clean. Live-reverified per constraint 7: a cross-catalog reveal (search hit in `fixture-flat` while `fixture-dcim` was the open catalog) correctly switched catalogs, scrolled to, and selected the target row — screenshot confirms rail highlight, breadcrumb, and details panel all read `fixture-flat` with `FILE_000042.BIN` selected.

### WR-03: `useModalBehavior`'s no-`initialFocusRef` fallback focuses a non-focusable container

**Files modified:** `frontend/src/hooks/useModalBehavior.ts`
**Commit:** `12c4afd2`
**Applied fix:** Applied the review's suggested fix verbatim (matches the code block in the finding almost exactly): in the `initialFocusRef` fallback branch, if the container has no focusable descendants, the hook now gives the container a `tabIndex = -1` (only if it doesn't already have a `tabindex` attribute) before calling `.focus()` on it, instead of calling `.focus()` on a `<div>` with no `tabindex` (a silent no-op). This path stays dead code for `CommandPalette` today (it always supplies `initialFocusRef={inputRef}`), matching the review's own characterization — the fix closes the gap for Phase 25–27 consumers per `24-CONTEXT.md`'s "must not reimplement any of these four behaviors" contract. Cleanup remains keyed on the `[isOpen]` transition, unchanged — constraint 2 (Phase 25's 260ms animated exit depends on cleanup firing on `isOpen: false`, not unmount) was not touched by this fix.
**Verification:** `npx tsc --noEmit` clean. Not independently re-verified live (this path is dead code for the current palette — `CommandPalette` always supplies `initialFocusRef`, so there is no live user-observable behavior change in Phase 24 to re-test; constraint 7 only required re-verifying the WR-01/WR-02 landmine behaviors).

### WR-04: The 50-result cap is duplicated as an unsourced literal in three frontend locations

**Files modified:** `frontend/src/components/workspace/CommandPalette.tsx`, `frontend/src/components/workspace/palette/PaletteResultList.tsx`
**Commit:** `461f7490`
**Applied fix:** Both frontend literal-`50` sites now derive from `results.length` (which is always `min(SearchIndexedCap, total)` per the Go contract) instead of restating the number: `CommandPalette.tsx`'s readout (`total > 50 ? \`50 of ${total}\` : ...` → `total > results.length ? \`${results.length} of ${total}\` : ...`) and `PaletteResultList.tsx`'s truncation line (`` `Showing the first 50 of ${total} hits` `` → `` `Showing the first ${results.length} of ${total} hits` ``). No shared Go-side constant was exported to the frontend for this — the review's own fix description frames `results.length` as the derivation target, not a shared constant, since Go's `SearchIndexedCap` isn't currently marshaled across the Wails bridge and adding a new binding just to expose an `int` would be more than the finding asked for. The exact-copy contract is preserved: today `results.length === 50` whenever these lines render (PLT-03's locked copy `Showing the first 50 of N hits` is unchanged), and the drift risk if `SearchIndexedCap` is ever bumped is now closed.
**Verification:** `npx tsc --noEmit` clean, `npm run build` clean (dist output unchanged in kind). Not independently re-verified live via dev-browser beyond the cross-catalog-reveal screenshot already taken for WR-02, which incidentally shows a non-truncated result set; the truncation-line and readout-derivation change is a pure string-computation change with no branching logic altered (the `>` comparison and truthy conditions are unchanged, only the right-hand operand), so `tsc`/build cleanliness plus a re-read of both edited lines was judged sufficient.

## Skipped Issues

None — all four in-scope findings (WR-01 through WR-04) were fixed. IN-01 and IN-02 were out of scope per the fix-launch instructions and were not attempted.

## Post-Fix Verification (constraint 6)

Run in the main checkout (no worktree — `workflow.use_worktrees: false`):

- `cd frontend && npx tsc --noEmit` — clean, no errors.
- `cd frontend && npm run build` — clean (`✓ 1460 modules transformed`, dist assets emitted; the repeated `Module level directives cause errors when bundled, 'use client' was ignored` lines are pre-existing dependency-level warnings, unrelated to this phase's changes).
- `go build ./...` — clean.
- `go test ./... -race -count=1` — all packages pass (`storcat-wails`, `cli`, `internal/catalog`, `internal/config`, `internal/fixture`, `internal/osutil`, `internal/search`).
- `git diff --stat -- frontend/package.json frontend/package-lock.json go.mod go.sum` — empty (constraint 3: no new dependencies).

## Live Re-Verification (constraint 7)

Performed via the dev-browser skill against `wails dev` on `:34115` (PID 48458), fixtures at `.../scratchpad/t2402-fixtures/` loaded via `localStorage['storcat-catalog-directory']` + reload.

**(a) Pre-expanded unrelated branches survive a reveal into a different branch:**
1. Selected `fixture-dcim` catalog, manually expanded `VOL02` and `VOL03` (leaving `VOL01` collapsed).
2. Opened the palette (⌘K), searched `IMG_0001` — first hit was `VOL01/100CANON` (the collapsed branch).
3. Pressed Enter to activate. Result: `VOL01` and `VOL01/100CANON` expanded, `IMG_0001.JPG` selected — and `VOL02`/`VOL03` remained expanded exactly as before (tree rows confirmed via `.ws-tree-row` text dump: `▾VOL01…▾100CANON…IMG_0001.JPG(selected)…▾VOL02…▾VOL03…`).

**(b) Cross-catalog reveal still lands:**
1. From `fixture-dcim` (still the open catalog), opened the palette and searched `FILE_000042` — a term unique to the other catalog, `fixture-flat`.
2. Pressed Enter to activate the single hit. Result (screenshot `cross-catalog-reveal.png`): the rail's active-catalog highlight, the breadcrumb, and the details panel all switched to `fixture-flat`, and `FILE_000042.BIN` is scrolled to center and selected.

Both landmine behaviors held after the WR-01 (`MERGE_EXPANDED`) and WR-02 (`REVEAL_HIT`) fixes.

---

_Fixed: 2026-08-14T17:45:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
