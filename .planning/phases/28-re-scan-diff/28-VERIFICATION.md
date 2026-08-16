---
phase: 28-re-scan-diff
verified: 2026-08-16T23:15:04Z
status: human_needed
score: 13/15 must-haves verified
behavior_unverified: 0
overrides_applied: 0
behavior_unverified_items: []
human_verification:
  - test: "Volume enumeration in RescanDialog step 1 completes fast enough that no dedicated loading state is needed"
    expected: "wailsAPI.listVolumes() resolves quickly enough in practice (on real removable media, not just this dev machine) that the absence of a loading spinner in step 1 never reads as a frozen dialog"
    why_human: "28-01-PLAN.md's own must_haves marks this `verification: backstop` — a timing/UX judgment about real removable-media latency that cannot be inferred from source, only observed live against real hardware over time"
  - test: "The 0.6 low-similarity ratio and 20-entry floor constants (internal/catalog/diff.go similarityThreshold/similarityMinEntries) correctly flag a wrong-disc pick without excessive false positives on real re-scan usage"
    expected: "Over real usage, the wrong-disc banner fires when the user actually inserted the wrong disc and stays quiet for legitimate large changes"
    why_human: "28-03-PLAN.md's own must_haves marks this `verification: backstop` — an untuned default explicitly flagged by the plan itself as needing real-world observation, not inferable from source or a unit test"
---

# Phase 28: Re-scan & Diff Verification Report

**Phase Goal:** Users re-scan a catalog's source volume and safely reconcile what changed, without risking the existing catalog
**Verified:** 2026-08-16T23:15:04Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | ACT-06: A user re-scans a catalog and sees a diff of added/removed/changed/unreadable/unchanged with counts | ✓ VERIFIED | `internal/catalog/diff.go` `ComputeDiff` implements all five states; `DiffList`/`RescanDialog` render them; `app.go RescanCatalog` binds it; `go test ./internal/catalog/... -run TestDiff -race` all pass; live-verified per 28-01/28-02/28-03 SUMMARYs with DOM queries and exact tile counts |
| 2 | An unreadable subtree is a FOURTH diff state, distinct from `removed`, and is never collapsed into it | ✓ VERIFIED | CR-01 fix confirmed live in `diff.go:77-80` — `case newItem.Unreadable:` now precedes `case !existed:`; regression test `TestDiff_NewUnreadableEntryIsUnreadableNotAdded` passes; descendant-pruning (`hasUnreadableAncestor`) confirmed in code, preventing a locked directory's prior contents from being falsely reported `removed` |
| 3 | Directories are diffed for existence only, never reported `changed` | ✓ VERIFIED | `diff.go` categorize switch has no directory-mtime branch; doc comment states the A2 resolution explicitly; `TestDiff_DirectoryNeverReportsChanged` referenced in 28-01-SUMMARY |
| 4 | ACT-08: Re-scan always asks for the source volume; nothing is persisted or pre-selected | ✓ VERIFIED | `RescanDialog`'s `selectedSource` is local `useState`, reset on every mount; no `localStorage`/config write anywhere in the re-scan path (grep confirms); volume choice is never round-tripped through `ResolveRescan`'s inputs |
| 5 | A wrong-disc pick shows an honest, non-blocking warning | ✓ VERIFIED | `LowSimilarity` computed in `diff.go`; banner in `RescanDialog.tsx` renders only when set, never disables footer buttons (grep + live DOM check per 28-03-SUMMARY: `disabled:false`/`aria-disabled:null` while banner showing) |
| 6 | ACT-07: A user resolves a diff via overwrite, keep-both, or discard — no per-entry picker, no merge | ✓ VERIFIED | `RescanDialog` footer has exactly the three locked actions; `ResolveRescan` binding accepts only `overwrite`/`keep-both` and explicitly rejects any other mode (`TestResolveRescan_DiscardIsNotAWritePath` passes) |
| 7 | The write path is irreversible and reuses the hardened atomic-write primitive and `nextCopyRoot` — no second write/naming path | ✓ VERIFIED | `internal/catalog/resolve.go` calls `nextCopyRoot` (from `duplicate.go`, unmodified) for keep-both and `WriteCatalogFrom` (→ `WriteFileAtomic`) as the only write call; `grep -c WriteFileAtomic resolve.go` = 0; `git diff` on `atomicwrite.go`/`duplicate.go` confirmed empty by the plan's own gate |
| 8 | The `.html` sidecar is rewritten when present, never created where absent | ✓ VERIFIED | `resolve.go:58-62` stats the original `.json`'s sibling `.html` and sets `WriteHTML` from disk state, never from renderer input; `TestWriteRescanResult_RewritesHtmlWhenPresent`/`DoesNotCreateHtmlWhenAbsent` pass |
| 9 | Discard writes nothing; the original catalog is byte-identical after a discard | ✓ VERIFIED | Discard has no Go binding at all (`ResolveMode` only defines `overwrite`/`keep-both`); dialog close-only, confirmed by code inspection and 28-04-SUMMARY's live sha256 checks |
| 10 | The overwrite target is re-derived and re-validated at write time, never trusted from the renderer | ✓ VERIFIED | `ResolveRescan` (app.go:571-604) re-runs `filepath.Abs → filepath.EvalSymlinks → osutil.ContainsPath` against a config-manager-derived `catalogDir`; `TestResolveRescan_RejectsPathOutsideCatalogDir` passes |
| 11 | A re-scan failure never populates Create's retained-partial state (`lastPartial`/`lastPartialResult`/`lastScanReq`) | ✓ VERIFIED | `RescanCatalog` (app.go) contains zero references to those three fields (grep-verified per the plan's own acceptance criterion); `TestRescan_DoesNotRetainPartialForWritePartialCatalog` passes |
| 12 | A held re-scan tree survives a failed write and can be retried without a full re-scan (WR-01) | ✓ VERIFIED | `app.go:606-645` — tree is cleared only after `WriteRescanResult` succeeds, guarded by a `lastRescanJSONPath`-match check against a concurrent new re-scan; `TestResolveRescan_RetainsTreeAcrossFailedWriteAndSucceedsOnRetry` passes |
| 13 | STATE-03: An unreadable catalog can be re-scanned (fresh scan, no diff), have its `.html` opened, or be removed from the library — three actions from the panel itself | ✓ VERIFIED | `UnreadableCatalogPanel.tsx` action trio wired to `RescanDialog oldTreeAvailable={false}`, a verbatim second copy of the footer's `handleOpenHtml`, and the unmodified Phase 27 `DeleteConfirmDialog`; `git diff -- app.go` and `-- DeleteConfirmDialog.tsx` both empty per the plan's own gate |
| 14 | COMPAT-06 is honestly recorded as build-proven only — never over-claiming signing/notarization/release | ✓ VERIFIED | Independently re-queried GitHub: exactly one `build.yml` run for commit `2715a42d` (id `31976677486`), all three jobs (`build-macos`/`build-linux`/`build-windows`) individually `success`; no `v3.0.0` tag on `origin`; `release.yml` last ran 2026-03-28 (pre-dates this push), confirming it was NOT fired. `REQUIREMENTS.md` COMPAT-06 checkbox is `[ ]` (unchecked) with "Partial — build proven, sign/notarize/release open" — matches evidence exactly |
| 15 | WINDOWS.md sweep closes exactly one entry (#7) on real evidence; #4/#5 reworded but stay open; #10/#11 annotated but stay open | ✓ VERIFIED | `.planning/WINDOWS.md`: `open_count: 10`, `fixed_count: 4` (up from 3 — only #7 moved, itself carrying a detailed live quit-and-relaunch account); #4/#5 both still `open`, both cite the `31976677486` run and say "still COMPILE evidence only"; #10/#11 both still `open`, both note the new Phase 28 call sites without changing status |

**Score:** 13/15 truths verified (2 backstop/non-inferable, routed to human verification per the plan's own explicit `verification: backstop` markers)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `pkg/models/catalog.go` | `ModTime`, `DiffState`/`DiffEntry`/`DiffResult` wire types | ✓ VERIFIED | Present, `omitempty` on `ModTime`, builds and used throughout |
| `internal/catalog/walk.go` | Exported `Walk`, extracted verbatim | ✓ VERIFIED | Single `func (s *Service) Walk` match; `CreateCatalogWithContext` reduced to a 2-line composition |
| `internal/catalog/diff.go` | `ComputeDiff`, pure, `pkg/models`-only | ✓ VERIFIED | All five states, descendant pruning, sum-invariant-by-construction, no non-stdlib/non-`pkg/models` imports |
| `internal/catalog/resolve.go` | `WriteRescanResult`, `ResolveMode` | ✓ VERIFIED | Reuses `nextCopyRoot`/`WriteCatalogFrom` exclusively; no `WriteFileAtomic` call |
| `internal/catalog/diff_test.go`, `resolve_test.go`, `service_test.go` | Table-driven regression coverage | ✓ VERIFIED | All named tests pass under `-race -count=1`; `service_test.go`'s pre-existing lines are diff-empty (git history confirms only additive `feat(28-02)` commit touched it) |
| `frontend/src/components/workspace/rescan/RescanDialog.tsx` | 3-step dialog, both variants, resolution footer | ✓ VERIFIED | Present, compiles, three-action footer with locked copy confirmed by grep |
| `frontend/src/components/workspace/rescan/DiffList.tsx` | Grouped diff list, native scroll | ✓ VERIFIED | No virtualizer/`useVisibleRows`; `formatBytes` reused; 4+ `.ws-rescan-diff*` CSS rules present |
| `frontend/src/components/workspace/UnreadableCatalogPanel.tsx` | Action trio | ✓ VERIFIED | Three buttons wired to real surfaces; zero Go diff; `DeleteConfirmDialog` diff empty |
| `.planning/WINDOWS.md` | Honest ledger sweep | ✓ VERIFIED | See truth #15 above |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `RescanCatalog` (app.go) | `catalogService.Walk` | direct call | ✓ WIRED | `grep -A40 RescanCatalog app.go` contains zero `startScan`/`CreateCatalogWithContext` references |
| `RescanCatalog` (app.go) | `catalog.ComputeDiff` | via `searchService.LoadCatalog` | ✓ WIRED | Confirmed in source; `oldTreeAvailable` gates old-tree load |
| `Options.MarkUnreadableOnSkip` | both `traverseDirectory` skip branches → node markers → `ComputeDiff` `unreadable` | ✓ WIRED | `RescanCatalog`'s `catalog.Options{..., MarkUnreadableOnSkip: true}` is the repo's single `true` opt-in (grep-confirmed); Create's call sites untouched |
| `ResolveRescan` (app.go) | `WriteRescanResult` → `nextCopyRoot`/`WriteCatalogFrom` → `WriteFileAtomic` | ✓ WIRED | Traced directly in `resolve.go` and `app.go`; no bypass |
| `UnreadableCatalogPanel` | `RescanDialog` (oldTreeAvailable=false) / `DeleteConfirmDialog` / footer's `handleOpenHtml` logic (duplicated) | ✓ WIRED | Confirmed in source, lines 161-227 |
| Catalog-actions menu | `RescanDialog` (2nd item, "Re-scan volume & diff…") | ✓ WIRED | `DetailsPanel.tsx` menu items array confirmed in order: Rename → Re-scan → Duplicate → Delete; `Menu.tsx` itself untouched (git log confirms no phase-28 commits touch it) |

### Behavioral Spot-Checks / Test Execution

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full Go build | `go build ./...` | exit 0 | ✓ PASS |
| Full Go vet | `go vet ./...` | exit 0 | ✓ PASS |
| Full Go test suite, race-enabled | `go test ./... -race -count=1` | all packages `ok` | ✓ PASS |
| Diff regression tests (incl. CR-01 fix) | `go test ./internal/catalog/... -run 'TestDiff_NewUnreadableEntryIsUnreadableNotAdded\|TestDiff_UnreadableIsNotRemoved\|TestDiff_CountsSumToDistinctPaths\|TestComputeDiff_EndToEndWithRealUnreadableSubdirectory' -race -v` | all PASS | ✓ PASS |
| Walk-marking regression tests | `-run 'TestTraverseDirectory_MarksSkippedNodeWhenFlagSet\|TestTraverseDirectory_SingleEntryErrorSkipsAndContinues'` | both PASS | ✓ PASS |
| ResolveRescan regression tests (incl. WR-01/WR-02 fixes) | `-run TestResolveRescan` | all 3 PASS, incl. `TestResolveRescan_RetainsTreeAcrossFailedWriteAndSucceedsOnRetry` | ✓ PASS |
| Frontend typecheck | `cd frontend && npx tsc --noEmit` | exit 0 | ✓ PASS |
| Frontend build | `npm run build` | exit 0, `dist/` produced | ✓ PASS |
| CI evidence re-query | `gh run view 31976677486` | 3/3 jobs `success`, `headSha` matches | ✓ PASS |
| No `v3.0.0` tag published | `git ls-remote --tags origin \| grep 3.0.0` | no match | ✓ PASS (confirms honest non-claim) |
| `release.yml` not fired by this push | `gh run list --workflow=release.yml` | latest run predates the push (2026-03-28) | ✓ PASS (confirms honest non-claim) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|--------------|--------|----------|
| ACT-06 | 28-01, 28-02, 28-03 | Re-scan + diff of added/removed/changed/unreadable/unchanged with counts | ✓ SATISFIED | Truths 1-3 above; `REQUIREMENTS.md` marks `[x]` Complete |
| ACT-07 | 28-04 | Resolve diff via overwrite/keep-both/discard | ✓ SATISFIED | Truths 6-10 above; `REQUIREMENTS.md` marks `[x]` Complete |
| ACT-08 | 28-01, 28-03 | Always ask for source volume, never guess | ✓ SATISFIED | Truth 4-5 above; `REQUIREMENTS.md` marks `[x]` Complete |
| STATE-03 | 28-01, 28-05 | Unreadable catalog: re-scan/open-html/remove | ✓ SATISFIED | Truth 13; `REQUIREMENTS.md` marks `[x]` Complete |
| COMPAT-06 | 28-06 | Build/sign/notarize/release on all CI targets | ✓ SATISFIED, honestly partial | Truth 14-15; `REQUIREMENTS.md` correctly marks `[ ]` unchecked with "Partial" note — this is the CORRECT outcome, not a gap, per this phase's own honesty mandate |

No orphaned requirements found — all five phase requirement IDs (ACT-06, ACT-07, ACT-08, STATE-03, COMPAT-06) appear in plan frontmatter and in `.planning/REQUIREMENTS.md`'s Phase 28 traceability rows with matching status.

### Anti-Patterns Found

None. Grep for `TBD|FIXME|XXX|TODO|HACK|PLACEHOLDER` across all phase-modified Go and TSX files under review returned zero hits. No stub returns (`return null`/`return {}`/`return []`), no hardcoded-empty props flowing to render, no console.log-only implementations found in the reviewed files.

### Accepted Deviations (per user instruction — not reported as gaps)

1. **`Menu.tsx` unmodified** — the catalog-actions menu's re-scan item is NOT visually dimmed during a concurrent scan (functional-only guard: click surfaces the locked tooltip via the shared error slot, never opens the dialog). The details-panel footer button dims correctly. Confirmed: `git log -- frontend/src/components/workspace/Menu.tsx` shows zero phase-28 commits.
2. **`ResolveRescan`'s `mode` parameter is a plain `string`**, not `catalog.ResolveMode` — Wails codegen limitation. `WriteRescanResult` itself keeps the typed constant. Confirmed in `app.go:571` and `resolve.go:40`.
3. **Keep-both button label is plain "Keep both"** with no filename — confirmed in `RescanDialog.tsx:373`; the filename claim was deliberately removed post-verification (commit `33dcdc9d`) because it was not collision-checked and could go stale.
4. **`ComputeDiff` prunes old-tree descendants under any unreadable-marked new-tree path** rather than reporting them `removed` — confirmed in `diff.go:52-69`; this is the phase's primary data-integrity control (T-28-05) and goes beyond the plan's literal text, added correctly per 28-02-SUMMARY's own disclosed deviation.

### Code Review Findings — Fix Verification

All four 28-REVIEW.md findings (1 critical, 3 warning) were found and fixed AFTER the six plan SUMMARYs were written. Independently confirmed each fix is actually in the code and covered by a passing regression test:

| Finding | Fix Location | Verified |
|---------|--------------|----------|
| CR-01 (unreadable-and-new miscategorized as `added`) | `diff.go:77-80`, commit `1eff66ea` | ✓ Code inspected; `TestDiff_NewUnreadableEntryIsUnreadableNotAdded` passes |
| WR-01 (tree cleared before write attempted) | `app.go:606-645`, commit `121b9547` | ✓ Code inspected; `TestResolveRescan_RetainsTreeAcrossFailedWriteAndSucceedsOnRetry` passes |
| WR-02 (`catalogDir` trusted from renderer) | `app.go:577-582`, commit `3b2de852` | ✓ Code inspected; `catalogDir` now derived from `a.configManager` only |
| WR-03 (dead CSS `.ws-rescan-error`) | `frontend/src/workspace.css`, commit `7d72cd91` | ✓ `grep -c ws-rescan-error\b` returns 0; only `ws-rescan-resolve-error` remains, and it is used |

### Human Verification Required

1 and 2 items above (frontmatter `human_verification`) — both are **explicit `verification: backstop` items self-flagged by the plans themselves** (28-01-PLAN.md and 28-03-PLAN.md must_haves), not defects found during this verification pass. Neither is inferable from source; both require observation against real removable-media hardware over real usage.

### 1. Volume-enumeration loading state omission

**Test:** Insert a real slow/large removable volume (optical media, a large USB drive with many partitions) and open the re-scan volume picker (step 1).
**Expected:** `wailsAPI.listVolumes()` returns fast enough in practice that the absence of a dedicated loading spinner never reads as a frozen or broken dialog.
**Why human:** Self-flagged `verification: backstop` in 28-01-PLAN.md — a real-hardware timing/UX judgment, not inferable from source or a fast dev-machine test.

### 2. Low-similarity threshold tuning (0.6 ratio / 20-entry floor)

**Test:** Over a period of real re-scan usage across a range of catalog sizes and genuine wrong-disc mistakes, observe whether the wrong-disc banner fires appropriately (present when the user genuinely picked the wrong disc, absent for legitimate large changes).
**Expected:** The named constants in `internal/catalog/diff.go` (`similarityMinEntries = 20`, `similarityThreshold = 0.6`) produce a banner that matches user intuition often enough to be useful, without excessive false positives/negatives.
**Why human:** Self-flagged `verification: backstop` in 28-03-PLAN.md — an untuned default the plan itself says is "not tuned against real removable-media re-scan data."

### Gaps Summary

No gaps found. This phase's must-haves — across all six plans, all four review-finding fixes, the honest COMPAT-06 record, and the WINDOWS.md sweep — are all substantively implemented, wired, and tested. The only items not marked fully VERIFIED are two constants/timing assumptions the plans themselves explicitly flagged as needing real-world observation (`verification: backstop`), which is why overall status is `human_needed` rather than `passed` — not because anything is broken, missing, or misrepresented.

Independent re-verification specifically targeted the phase's stated highest-risk claims (per the verification_emphasis instructions) and found every one to hold:
- COMPAT-06's CI evidence is accurate and not over-claimed (independently re-queried against GitHub, matching the phase record exactly).
- WINDOWS.md's sweep closed exactly the one entry the plan intended (#7), on real live evidence, and left the other ten honestly open.
- The write path (`resolve.go`) genuinely reuses the hardened atomic-write primitive and `nextCopyRoot` — no second write or naming path exists.
- All four post-summary code-review fixes (1 critical, 3 warning) are actually present in the code and covered by passing regression tests, not just claimed.

---

_Verified: 2026-08-16T23:15:04Z_
_Verifier: Claude (gsd-verifier)_
