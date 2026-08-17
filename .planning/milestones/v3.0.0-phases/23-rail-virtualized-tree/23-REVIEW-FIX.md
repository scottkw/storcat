---
phase: 23-rail-virtualized-tree
fixed_at: 2026-08-13T00:00:00Z
review_path: .planning/phases/23-rail-virtualized-tree/23-REVIEW.md
iteration: 1
findings_in_scope: 5
fixed: 5
skipped: 0
status: all_fixed
---

# Phase 23: Code Review Fix Report

**Fixed at:** 2026-08-13
**Source review:** .planning/phases/23-rail-virtualized-tree/23-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 5 (0 Critical / 3 Warning / 2 Info -- full scope)
- Fixed: 5
- Skipped: 0

## Fixed Issues

### WR-01: `extractErrorMessage` fix not applied to all `wailsAPI` call sites

**Files modified:** `frontend/src/services/wailsAPI.ts`
**Commit:** `786b8ddf`
**Applied fix:** Routed the five broken catch blocks (`setTheme`, `setSidebarPosition`, `setWindowSize`, `openExternal`, `setWindowPersistence`) through `extractErrorMessage` as the finding specified. Went one step further per the fix guidance's explicit ask ("make the omission structurally impossible to repeat... one place, not 17 copies"): extracted a single `wailsError(error)` helper (`{success: false, error: extractErrorMessage(error)}`) and routed every one of the file's 17 catch blocks through it, including the 12 that were already correct. There is now exactly one line in the file that reads a caught Wails rejection -- no new abstraction layer, just deduplicating an already-identical three-line block that had drifted five times.

### WR-02: `RevealInFileManager` has no containment check against the configured catalog directory

**Files modified:** `app.go`, `internal/osutil/reveal.go`, `internal/osutil/reveal_test.go`, `frontend/wailsjs/go/main/App.d.ts`, `frontend/wailsjs/go/main/App.js`, `frontend/src/services/wailsAPI.ts`, `frontend/src/components/workspace/DetailsPanel.tsx`
**Commit:** `d5f41f1c`
**Applied fix:** Per explicit direction to close this gap properly rather than re-record it as accepted risk, `RevealInFileManager` now takes a `catalogDir` parameter end to end: Go signature (`RevealInFileManager(path, catalogDir string) error`), the Wails binding in `app.go`, the regenerated TS bindings (`wails generate module`), `wailsAPI.ts`'s wrapper, and `DetailsPanel.tsx`'s one real call site (reads `state.catalogDir` from `AppContext`, fails closed with a "No catalog directory configured." message if it's null rather than calling the backend with an empty string).

A new pure `containsPath(catalogDir, resolved)` helper -- extracted the same way `revealArgvFor` already was, so it is testable without ever reaching `exec.Command` -- compares `filepath.Clean`ed, symlink-resolved absolute paths via `filepath.Rel`, not `strings.HasPrefix`: `/catalogs-evil` does not pass a `/catalogs` prefix test with this approach (`Rel` returns a path starting with `..` for anything outside, which a naive string prefix check cannot detect). `RevealInFileManager` rejects an empty `catalogDir` outright (fail closed) and rejects any resolved path `containsPath` reports as outside it.

Table-driven Go tests (`TestContainsPath`) cover all four scenarios the fix guidance named: a legitimate in-directory path, a sibling directory sharing a name prefix, a `../` escape, and a symlink whose resolved target lands outside `catalogDir`. Two additional integration-level tests on `RevealInFileManager` itself (`TestRevealInFileManager_RejectsMissingCatalogDir`, `TestRevealInFileManager_RejectsPathOutsideCatalogDir`) prove containment is actually wired into the real function, not just the helper in isolation -- both stop before `exec.Command` is ever reached, so neither pops a real Finder/Explorer window during `go test`.

**Verified live in a browser** (`wails dev` at `http://localhost:34115`, `dev-browser`) against a real two-catalog fixture directory: selected a catalog and clicked "Reveal JSON in Finder" -- the call completed with no error surfaced in the details panel footer, confirming the legitimate call site (real `catalogDir` from `AppContext`) still works end to end with the new containment check in place.

### WR-03: Sidecar counts cache has no eager/background fill -- rail and status bar can silently show 0 for catalogs that were never opened

**Files modified:** `frontend/src/components/workspace/StatusBar.tsx`
**Commit:** `23692e32`
**Applied fix:** Option (b) from the finding's own fix text -- deliberately not a background-fill subsystem (explicitly ruled out as more machinery than the problem deserves). `StatusBar.tsx`'s existing memo now also tracks whether any catalog's `fileCount`/`totalBytes` was absent (cold cache) and renders a `"≥"` qualifier on the files and bytes segments whenever it was, distinguishing "not computed yet" from "genuinely counted as zero" per the UI-SPEC's status bar contract.

**Verified live in a browser** against a real two-catalog fixture directory (`/private/tmp/.../scratchpad/catalogs`, `alpha-volume.json` + `beta-volume.json`, neither opened yet):
- **Counts not yet computed:** status bar read `"2 catalogs · ≥0 files indexed · ≥0.0 GB"`; rail rows correctly omitted their `· N files` fragment (RAIL-04 precedent, unaffected by this fix).
- **Counts known:** after selecting both catalogs (each `LoadCatalogFlat` call opportunistically warms its own cache entry, per `flatten.go`) and reloading, status bar read `"2 catalogs · 5 files indexed · 0.0 GB"` (no qualifier), with rail rows showing `alpha-volume.json · 2 files` and `beta-volume.json · 3 files` -- 2 + 3 = 5 confirms the sum is correct once every entry is warm.

### IN-01: TOCTOU window between path validation and `exec.Command` in `reveal.go`

**Files modified:** `internal/osutil/reveal.go` (folded into commit `d5f41f1c`, same function WR-02 modifies)
**Commit:** `d5f41f1c`
**Applied fix:** Added the one-line `ponytail:`-style comment the finding asked for, directly above the `EvalSymlinks` call: names the accepted ceiling (exploiting it requires local write access to the exact resolved path, which already grants stronger primitives than "reveal the wrong file") and the upgrade path (`O_NOFOLLOW` + fd-based open) if this file is ever revisited. No behavioral change, as the finding itself specified this is not a blocking fix.

### IN-02: `maxFlattenDepth` cap allows one more level than its name suggests

**Files modified:** `internal/search/flatten.go`
**Commit:** `ada22533`
**Applied fix:** Changed the depth guard from `if depth > maxFlattenDepth` to `if depth >= maxFlattenDepth`, so the walk now tolerates exactly 512 levels (depths 0 through 511) before rejecting at depth 512, matching the constant's documented cap exactly. `TestLoadCatalogFlat_DepthCap`'s 600-level fixture exceeds the cap under either guard and continues to pass unchanged -- confirmed by `go test ./internal/search/... -race`.

## Skipped Issues

None -- all five findings were fixed.

## Verification

All four project gates run clean after all five fixes, in the main checkout (worktrees disabled via `workflow.use_worktrees: false` in `.planning/config.json` -- fixes were applied and committed directly on `main`, not in an isolated worktree, so these results are reproducible from this tree as-is):
- `go build ./...` — exit 0
- `go test ./... -race` — exit 0, all packages `ok` (including the new `containsPath` table-driven tests and the two new `RevealInFileManager` integration tests, all passing under `-race`)
- `cd frontend && npx tsc --noEmit` — exit 0, no errors
- `cd frontend && npm run build` — exit 0, 1455 modules transformed, `dist/` produced

Live browser verification (`wails dev` at `http://localhost:34115`, `dev-browser`, against a generated two-catalog fixture directory) confirmed both the WR-02 reveal-containment fix (legitimate call succeeds, no error surfaced) and the WR-03 status-bar honesty fix (cold state shows `"≥0 files indexed / ≥0.0 GB"`, warm state shows the correct unqualified `"5 files indexed / 0.0 GB"`) -- see WR-02 and WR-03 above for the captured evidence. `wails dev` was stopped after verification; no lingering processes remain.

`23-REVIEW.md` was updated in place with a **Disposition:** line under each of the 5 findings recording fixed status and commit hashes, and committed separately (docs commit, left for the orchestrator per instructions -- this file itself is also left uncommitted).

---

_Fixed: 2026-08-13_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
