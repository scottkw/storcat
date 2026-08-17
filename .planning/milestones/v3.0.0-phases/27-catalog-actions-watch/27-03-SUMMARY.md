---
phase: 27-catalog-actions-watch
plan: 03
subsystem: catalog-actions
tags: [go, wails, wastebasket, trash, duplicate, filesystem, containment]

# Dependency graph
requires:
  - phase: 27-catalog-actions-watch
    provides: "27-01's containment-gated App.RenameCatalog binding shape and WriteFileAtomic reuse pattern; 27-02's fsync-hardened WriteFileAtomic that both duplicate copies route through"
provides:
  - "internal/catalog.DuplicateCatalog(jsonPath) (string, error) -- -copy/-copy-N collision-suffix loop, both-extensions freeness check, byte-identical copy via WriteFileAtomic"
  - "internal/osutil.TrashPaths(catalogDir, paths...) error -- containment-gated OS Trash wrapper around a swappable trashSeam, never falling back to local removal"
  - "App.DuplicateCatalog / App.DeleteCatalog Wails bindings, both containment-gated identically to RenameCatalog"
  - "wailsAPI.duplicateCatalog / wailsAPI.deleteCatalog service wrappers"
  - "github.com/Bios-Marcel/wastebasket/v2 v2.0.3 -- first of this phase's two new Go dependencies"
affects: [27-04, 27-05, 27-06, 27-07]

# Actuals (#2632)
actuals:
  tokens: 8652
  tasks: 3
  commits: 6

# Tech tracking
tech-stack:
  added:
    - "github.com/Bios-Marcel/wastebasket/v2 v2.0.3 -- cross-platform OS Trash mover; wraps macOS osascript, Windows SHFileOperationW, Linux FreeDesktop trash spec"
  patterns:
    - "Package-level swappable seam (trashSeam, unexported, same-package tests) for a third-party call a test must never actually invoke -- mirrors the test-seam shape this repo hadn't needed before now"
    - "Both-extensions freeness check (.json AND .html) before taking a collision-suffix candidate, so an orphaned HTML from a prior partial delete can never be silently clobbered"

key-files:
  created:
    - internal/catalog/duplicate.go
    - internal/catalog/duplicate_test.go
    - internal/osutil/trash.go
    - internal/osutil/trash_test.go
  modified:
    - go.mod
    - go.sum
    - app.go
    - frontend/src/services/wailsAPI.ts
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js

key-decisions:
  - "go.mod's `go` directive bumped from 1.23 to 1.23.4 -- not a discretionary choice: wastebasket's own go.mod requires go 1.23.4, and Go's module graph pruning rejects a lower directive once that dependency is added. Manually reverting to 1.23 was tried and confirmed to fail `go build` with 'updates to go.mod needed'. No other module was added or bumped beyond wastebasket's own transitive dependency (gobwas/glob, indirect)."
  - "Reworded a trash_test.go doc comment that happened to contain the bare word 'wastebasket' (in prose, not an import) after the plan's own <verification> block's literal grep (`grep -rn 'wastebasket' internal/` should show the import only in trash.go) would otherwise have failed on it -- same class of literal-grep-vs-intent mismatch 27-01 and 27-02 already documented precedent for fixing rather than silently claiming compliance."
  - "DeleteCatalog does not os.Stat the derived .html before appending it to the trash path list -- TrashPaths already skips a missing path via its own os.Lstat/os.IsNotExist branch, so a second check in app.go would just be a second place for the two to disagree, per the plan's explicit instruction."

requirements-completed: [ACT-03, ACT-04, ACT-05]

coverage:
  - id: D1
    description: "Duplicating a catalog produces the next free -copy/-copy-N filename root, checking both .json and .html so an orphaned HTML is never clobbered, with the .html companion copied when present and the title inherited byte-identically"
    requirement: "ACT-03"
    verification:
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_FirstCopy"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_SecondCopy"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_ThirdCopy"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_SkipsRootWithOrphanHTML"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_CopiesHTMLWhenPresent"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_NoHTMLCopiesOnlyJSON"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_IsByteIdentical"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_LeavesSourceUntouched"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_RejectsNonJSON"
        status: pass
      - kind: unit
        ref: "internal/catalog/duplicate_test.go#TestDuplicateCatalog_RejectsMissingSource"
        status: pass
    human_judgment: false
  - id: D2
    description: "TrashPaths validates, symlink-resolves, and containment-checks every path before reaching the OS Trash seam; a missing path is skipped (free retry), an empty catalogDir or an out-of-bounds path is rejected before the seam is ever called, and a seam error propagates verbatim via errors.Is"
    requirement: "ACT-04"
    verification:
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_PassesResolvedPathsToSeam"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_RejectsOutsideCatalogDir"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_RejectsSiblingPrefixDirectory"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_RejectsDisallowedExtension"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_RejectsDirectory"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_SkipsMissingPath"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_AllMissingReturnsNil"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_EmptyCatalogDirIsAnError"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_PropagatesSeamErrorVerbatim"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_NonASCIIPathReachesSeamByteIdentical"
        status: pass
      - kind: unit
        ref: "internal/osutil/trash_test.go#TestTrashPaths_NoPathsIsNil"
        status: pass
    human_judgment: false
  - id: D3
    description: "No permanent-deletion fallback exists anywhere in the delete path -- the trash seam is the only removal mechanism reachable from TrashPaths, and DeleteCatalog carries no force parameter"
    requirement: "ACT-05"
    verification:
      - kind: unit
        ref: "acceptance grep: ! grep -Eq 'os\\.Remove|os\\.RemoveAll|os\\.Truncate' internal/osutil/trash.go"
        status: pass
      - kind: unit
        ref: "acceptance grep: awk '/func \\(a \\*App\\) DeleteCatalog/,/^}/' app.go | grep -Ec 'os\\.Remove|os\\.RemoveAll' returns 0"
        status: pass
    human_judgment: false
  - id: D4
    description: "App.DuplicateCatalog and App.DeleteCatalog reject a path outside the configured catalog directory before any read, write, or trash call, matching RenameCatalog's containment gate exactly"
    requirement: "ACT-03, ACT-04"
    verification:
      - kind: unit
        ref: "acceptance grep: awk '/func \\(a \\*App\\) DuplicateCatalog/,/^}/' app.go | grep -c osutil.ContainsPath returns 1; same for DeleteCatalog"
        status: pass
    human_judgment: true
    rationale: "The containment-gate binding logic is unit-grep-verified as present and correctly shaped, but no live dev-browser round trip was performed against a real wails dev session for this plan (unlike 27-01's tracer). A human/UAT pass exercising duplicateCatalog and deleteCatalog through the running app is the remaining proof this deliverable needs before being fully trusted."
duration: 6min
completed: 2026-08-15
status: complete
---

# Phase 27 Plan 03: Duplicate and Trash-delete backend Summary

**`catalog.DuplicateCatalog`'s -copy/-copy-N collision loop and `osutil.TrashPaths`'s containment-gated OS Trash wrapper (new dependency `wastebasket/v2 v2.0.3`) land as `App.DuplicateCatalog`/`App.DeleteCatalog` bindings, both gated identically to 27-01's `RenameCatalog`, with zero permanent-deletion fallback anywhere in the delete path.**

## Performance

- **Duration:** ~6 min
- **Started:** 2026-08-15T17:48:30Z
- **Completed:** 2026-08-15T17:53:56Z
- **Tasks:** 3
- **Files modified:** 10 (4 created, 6 modified)

## Accomplishments
- `internal/catalog.DuplicateCatalog` finds the next free `-copy`/`-copy-N` filename root -- checking **both** `.json` and `.html` at each candidate so an orphaned HTML from a prior partial delete can never be silently clobbered -- and copies the JSON (and HTML, if present) byte-identically through `WriteFileAtomic`, inheriting the title verbatim with no JSON parse at all.
- `internal/osutil.TrashPaths` validates every path (Lstat existence, `filepath.Abs`+`EvalSymlinks`, regular-file, `.json`/`.html` extension allowlist, `osutil.ContainsPath`) before handing a resolved list to a swappable `trashSeam` initialized to `wastebasket.Trash`. A missing path is skipped, not an error -- the retry-after-partial-failure UX this phase's `must_haves` require is free because `wastebasket` itself already tolerates a missing path per-backend, verified by `27-RESEARCH.md`'s full source read.
- `github.com/Bios-Marcel/wastebasket/v2 v2.0.3` added as this phase's first new Go dependency, exact-pinned in `go.mod`'s direct require block, per the ROADMAP's pre-approval and `27-RESEARCH.md`'s Package Legitimacy Audit (OK/Approved).
- `App.DuplicateCatalog` and `App.DeleteCatalog` gate every caller-supplied path through the identical `filepath.Abs` + `filepath.EvalSymlinks` + `osutil.ContainsPath` sequence `App.RenameCatalog` (27-01) already carries. `DeleteCatalog` derives its `.html` companion in Go via the repo's one `.json`/`.html` convention -- the renderer can never supply or name a second file for deletion.
- Wails bridge regenerated (`App.d.ts`/`App.js`) and `wailsAPI.duplicateCatalog`/`wailsAPI.deleteCatalog` added, both routed through the shared `wailsError` helper.
- No local filesystem-removal call exists anywhere in the new delete surface: the acceptance greps forbidding `os.Remove`/`os.RemoveAll`/`os.Truncate` in `trash.go` and in `App.DeleteCatalog`'s body both pass, and `wastebasket.Trash` itself was verified (27-RESEARCH.md, full source read of all four platform backends) to never fall back to permanent deletion.

## Task Commits

1. **Task 1: DuplicateCatalog -- collision loop (RED)** - `0ad71616` (test) -- `internal/catalog/duplicate_test.go`, 10 cases
2. **Task 1: DuplicateCatalog -- collision loop (GREEN)** - `46cac9d5` (feat) -- `internal/catalog/duplicate.go`
3. **Task 2: TrashPaths -- containment-gated seam (RED)** - `91c99c54` (test) -- `internal/osutil/trash_test.go`, 11 cases
4. **Task 2: TrashPaths -- containment-gated seam (GREEN)** - `e62dedf9` (feat) -- `internal/osutil/trash.go`, `go.mod`, `go.sum`
5. **Task 2 follow-up: literal-grep comment fix** - `3d475fd3` (test) -- reworded a doc comment tripping the plan's own `wastebasket` scope-grep
6. **Task 3: DuplicateCatalog/DeleteCatalog bindings + bridge + wailsAPI** - `0c33425c` (feat) -- `app.go`, `frontend/src/services/wailsAPI.ts`, `frontend/wailsjs/go/main/App.d.ts`, `frontend/wailsjs/go/main/App.js`

**Plan metadata:** pending (this SUMMARY's own commit)

## Files Created/Modified
- `internal/catalog/duplicate.go` - `DuplicateCatalog`, `nextCopyRoot`, `isCandidateRootFree`
- `internal/catalog/duplicate_test.go` - 10 table-driven cases covering the full `<behavior>` list
- `internal/osutil/trash.go` - `TrashPaths`, unexported `trashSeam` package var wrapping `wastebasket.Trash`
- `internal/osutil/trash_test.go` - 11 cases, all exercising the seam only, never a real OS Trash
- `go.mod`, `go.sum` - `github.com/Bios-Marcel/wastebasket/v2 v2.0.3` (direct, exact-pinned); `go` directive bumped to `1.23.4` (required by wastebasket's own `go.mod`)
- `app.go` - `App.DuplicateCatalog(jsonPath, catalogDir) (string, error)`, `App.DeleteCatalog(jsonPath, catalogDir, deleteHtml) error`, `strings` import added
- `frontend/src/services/wailsAPI.ts` - `duplicateCatalog`, `deleteCatalog` wrappers
- `frontend/wailsjs/go/main/App.d.ts`, `App.js` - regenerated via `wails generate module`

## Decisions Made
- **`go` directive bump to 1.23.4 is required, not discretionary.** Verified by manually reverting it to `1.23` after `go mod tidy`: `go build ./...` immediately failed with "updates to go.mod needed", because `wastebasket`'s own `go.mod` declares `go 1.23.4` and Go's module graph pruning rejects a lower directive once that dependency is in the graph. `git diff go.mod` shows exactly `wastebasket` (direct) and `gobwas/glob` (its own transitive indirect dependency) added -- no other module bumped, satisfying the plan's verification intent even though the literal wording ("no other module... bumped except golang.org/x/sys") didn't anticipate a `go` directive bump.
- **Reworded a `trash_test.go` doc comment** that used the bare word "wastebasket" in prose (not an import) after confirming it would trip the plan's own `<verification>` block's `grep -rn 'wastebasket' internal/` scope check, which expects the import to appear only in `trash.go`. Same class of literal-grep-vs-intent mismatch 27-01-SUMMARY.md and 27-02-SUMMARY.md already established precedent for fixing rather than silently claiming compliance.
- **`DeleteCatalog` does not pre-stat the derived `.html`** before appending it to the trash path list, per the plan's explicit instruction: `TrashPaths` already skips a path that `os.Lstat` reports missing, so a second existence check in `app.go` would just create a second place for the two to disagree.
- **Reworded a comment in `duplicate.go`'s doc string** during Task 1 (`os.Create+io.Copy` → "a truncate-then-stream copy") after it tripped the acceptance grep `! grep -Eq 'os\.WriteFile|os\.Create'` by matching inside prose, not code -- caught and fixed before commit, not left as a deviation.

## Deviations from Plan

### Acceptance-criterion/verification-wording mismatches (not bugs, documented for the record)

**1. `go.mod`'s `go` directive changed, beyond the plan's literal "no other module... bumped" wording**
- **Found during:** Task 2 verification
- **Cause:** `wastebasket/v2 v2.0.3`'s own `go.mod` requires `go 1.23.4`; Go's module graph pruning enforces this transitively once the dependency is added, regardless of the importing project's own preferred directive.
- **Action taken:** Confirmed by direct experiment (manual revert to `go 1.23` → `go build` failure) that this is unavoidable, not an artifact of running `go mod tidy` carelessly. Kept the required `1.23.4` directive; no other module was added or bumped beyond `wastebasket` itself and its own transitive `gobwas/glob` dependency.
- **Verification:** `git diff go.mod` shows exactly two additions (the direct `wastebasket` requirement and its indirect `gobwas/glob` dependency) plus the `go` directive line; `go build ./... && go vet ./... && go test ./... -race -count=1` green throughout.

**2. A `trash_test.go` doc comment tripped the plan's own `wastebasket`-scope verification grep**
- **Found during:** final plan-level `<verification>` pass
- **Cause:** `grep -rn 'wastebasket' internal/` is meant to prove the import is scoped to `trash.go` alone; a doc comment in `trash_test.go` used the bare word "wastebasket" in prose (explaining the macOS backend's escaping behavior), which the substring grep also matched even though nothing was imported.
- **Action taken:** Reworded the comment ("the underlying trash library's macOS backend") to preserve the same explanation without the literal substring. Fixed before this SUMMARY was written; no functional test change.
- **Verification:** `grep -rn 'wastebasket' internal/` now shows matches only in `trash.go`; `go test ./internal/osutil/... -run TestTrashPaths -race -count=1` still green.

**3. A `duplicate.go` doc comment tripped its own no-truncating-write acceptance grep, caught pre-commit**
- **Found during:** Task 1 verification, before the GREEN commit
- **Cause:** The doc comment originally said "rather than `os.Create+io.Copy`", which matched the acceptance grep `! grep -Eq 'os\.WriteFile|os\.Create'` inside prose rather than code.
- **Action taken:** Reworded to "a truncate-then-stream copy" before committing -- caught during the same verification pass that wrote the code, not a later fix.
- **Verification:** `! grep -Eq 'os\.WriteFile|os\.Create' internal/catalog/duplicate.go` passes; `go test ./internal/catalog/... -run TestDuplicateCatalog -race -count=1` green.

---

**Total deviations:** 3, all literal-grep/verification-wording mismatches against required or accidental prose matches -- no functional bug, no scope creep, all fixed or confirmed-unavoidable before their respective commits.
**Impact on plan:** None on correctness. All three are the same class of literal-vs-intent mismatch 27-01-SUMMARY.md and 27-02-SUMMARY.md already established precedent for documenting/fixing rather than silently claiming compliance.

## Issues Encountered
- The first `go get`/`go mod tidy` cycle silently stripped `wastebasket` back out of `go.mod` because nothing in the source tree imported it yet at that point (Go's module tidy removes an unused requirement). Resolved by writing `trash.go`'s import first, then re-running `go get`/`go mod tidy` -- the dependency then landed correctly in the direct `require` block.

## Next Phase Readiness
- `internal/osutil.ContainsPath` and `allowedRevealExtensions` are now reused by a third consumer (`TrashPaths`, after `RevealInFileManager`/`ResolveContainedFileURL`/`GetCatalogHtmlPath`/`RenameCatalog`), reinforcing the phase's shared-containment-gate pattern for `27-04`'s menu wiring and `27-05`'s confirmation dialogs to build their UI on top of.
- `App.DuplicateCatalog`/`App.DeleteCatalog` and their `wailsAPI` wrappers are live and ready for `27-04` (menu/context-menu wiring) and `27-05` (confirmation dialogs) to call.
- No live `wails dev`/dev-browser round trip was performed for this plan's two new bindings (unlike 27-01's tracer gate) -- `D4`'s `coverage` entry above flags this as the remaining human-judgment item; `27-04`/`27-05`, which wire these bindings into actual UI, are the natural place for that live verification to happen.

## Self-Check: PASSED

All 4 created files verified present on disk (`internal/catalog/duplicate.go`, `internal/catalog/duplicate_test.go`, `internal/osutil/trash.go`, `internal/osutil/trash_test.go`); all 6 commits (`0ad71616`, `46cac9d5`, `91c99c54`, `e62dedf9`, `3d475fd3`, `0c33425c`) verified present in `git log`.

---
*Phase: 27-catalog-actions-watch*
*Completed: 2026-08-15*
