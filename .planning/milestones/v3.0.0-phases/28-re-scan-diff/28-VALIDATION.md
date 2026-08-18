---
phase: 28
slug: re-scan-diff
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-16
---

# Phase 28 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by plan-phase from `28-RESEARCH.md`'s `## Validation Architecture`. The Per-Task
> Verification Map is filled in once PLAN.md task IDs exist.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing`, table-driven, beside source. `internal/catalog/` already has 6 test files (`service_test.go`, `duplicate_test.go`, `atomicwrite_test.go`, `atomicwrite_sigkill_test.go`, `measure_test.go`, `rename_test.go`). Frontend: **none by design** — TEST-01 deferred; do not add one. |
| **Config file** | none — plain `go test ./...` |
| **Quick run command** | `go test ./internal/catalog/... ./pkg/models/... -race -count=1` |
| **Full suite command** | `go build ./... && go vet ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build)` |
| **Estimated runtime** | ~90–120 seconds (the Phase 27 SIGKILL subprocess test dominates) |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/catalog/... ./pkg/models/... -race -count=1`
- **After every plan wave:** full suite command, plus a live dev-browser pass against `:34115`
- **Phase gate:** full suite green **plus a real observed green `build.yml` Actions run** for COMPAT-06 — no local command substitutes for that
- **Max feedback latency:** ~120 seconds

**Dev-server note:** browser verification runs against `wails dev` on **`:34115`**. Vite's `:5173` exposes no `window.go`. `curl` liveness proves the server is up, not that bindings are fresh — probe `Object.keys(window.go.main.App)` for new methods before recording evidence.

**No host-OS GUI automation.** No osascript / System Events / keystroke injection (STATE.md records a prior-phase incident). Filesystem operations from Bash are the correct way to stage diff fixtures.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 28-01 T1 | 01 | 1 | ACT-06, ACT-08 | T-28-01 | `Walk` extraction is behavior-preserving; re-scan never routes through `startScan` | unit + regression | `go test ./internal/catalog/... -run 'TestCreateCatalog\|TestDiff' -count=1` | ✅ `internal/catalog/diff_test.go`, `walk.go` | ✅ pass — `service_test.go` provably unmodified; 13 diff tests green |
| 28-01 T2 | 01 | 1 | STATE-03 | T-28-04 | Re-scan failure never populates `a.lastPartial`/`a.lastScanReq` | unit (app pkg) | `go test . -run TestRescan -count=1` | ✅ `app_test.go` | ✅ pass — `TestRescan_DoesNotRetainPartialForWritePartialCatalog` |
| 28-02 T1 | 02 | 2 | ACT-06 | T-28-05 | Skipped subtree leaves a marker instead of being silently dropped | unit (tdd) | `go test ./internal/catalog/... -run MarksSkipped -count=1` | ✅ `service_test.go` | ✅ pass — `TestTraverseDirectory_MarksSkippedNodeWhenFlagSet`; the pre-existing skip-and-continue regression still passes unmodified |
| 28-02 T2 | 02 | 2 | ACT-06 | T-28-05 | `unreadable` outranks `added`/size; descendants pruned, never reported `removed`; counts sum to distinct paths | unit (tdd) | `go test ./internal/catalog/... -run TestDiff -count=1` | ✅ `diff_test.go` | ✅ pass — 13 tests incl. `_UnreadableIsNotRemoved`, `_CountsSumToDistinctPaths`, `_TypeChangeYieldsRemovedAndAdded`, and an end-to-end real `chmod 000` fixture |
| 28-02 T3 | 02 | 2 | ACT-06 | — | Exactly one opt-in caller turns the marker on | static assertion | `grep -c 'MarkUnreadableOnSkip:\s*true'` → 1 | ✅ `app.go` | ✅ pass — single call site |
| 28-03 T1 | 03 | 3 | ACT-06 | T-28-08 | Paths/errors rendered as React text children only — no `innerHTML` path | typecheck + build + grep gate | `cd frontend && npx tsc --noEmit && npm run build` | n/a — TEST-01 deferred | ✅ pass — gates green; no `dangerouslySetInnerHTML`/`innerHTML` in `rescan/` |
| 28-03 T2 | 03 | 3 | ACT-06, ACT-08 | T-28-09, T-28-10 | Similarity banner is advisory and disables nothing; caption names what is overwritten | typecheck + build + live | `npx tsc --noEmit && npm run build` + dev-browser | n/a — TEST-01 deferred | ✅ pass — live: banner shown, footer `disabled:false` |
| 28-03 T3 | 03 | 3 | ACT-06 | — | Error step offers no partial-write affordance | typecheck + build + live | `npx tsc --noEmit && npm run build` + dev-browser | n/a — TEST-01 deferred | ✅ pass — live: "Scan interrupted", no "Write partial" element anywhere |
| 28-03 T4 | 03 | 3 | ACT-06, ACT-08 | — | The diff step, end to end, against real staged fixtures | manual (dev-browser, `:34115`) | live verification | n/a — TEST-01 deferred | ✅ pass — tiles 2/1/1/1/3 summed to 8 = distinct paths; `chmod 000` dir under UNREADABLE, absent from REMOVED |
| 28-04 T1 | 04 | 4 | ACT-07 | T-28-06, T-28-07 | Single write path: reuses `WriteCatalogFrom`/`WriteFileAtomic` and the shared `nextCopyRoot` | unit (tdd) | `go test ./internal/catalog/... -run TestWriteRescanResult -count=1` | ✅ `resolve_test.go` | ✅ pass — 4 tests; `grep -c WriteFileAtomic resolve.go` = 0, `nextCopyRoot` reused |
| 28-04 T2 | 04 | 4 | ACT-07 | T-28-07 | Write failure surfaces the real error and leaves every action reachable | typecheck + build + live | `npx tsc --noEmit && npm run build` + dev-browser | n/a — TEST-01 deferred | ✅ pass — live: `.ws-rescan-resolve-error` banner, all three buttons `disabled:false` |
| 28-04 T3 | 04 | 4 | ACT-07 | — | Overwrite / keep-both / discard each do exactly what they claim, on disk | manual (dev-browser + hashes) | live verification | n/a — TEST-01 deferred | ✅ pass — discard byte-identical (`e3bb4dbd…` unchanged); keep-both preserved original, collided to `-copy-2`; overwrite → `b12a275c…` |
| 28-05 T1 | 05 | 5 | STATE-03 | — | Three exits all reuse existing surfaces; no invented membership state | typecheck + build + grep gate | `npx tsc --noEmit && npm run build` | n/a — TEST-01 deferred | ✅ pass — `app.go`/`DeleteConfirmDialog.tsx` diffs empty; no hidden/excluded/archived state |
| 28-05 T2 | 05 | 5 | STATE-03 | — | The trio against genuinely unparseable catalogs | manual (dev-browser, `:34115`) | live verification | n/a — TEST-01 deferred | ✅ pass — 5 corrupted fixtures; real Trash move, recoverable; keep-both left the broken original byte-identical |
| 28-06 T1 | 06 | 6 | COMPAT-06 | — | Publication route is a user decision, not a silent step | decision checkpoint | n/a — blocking human decision | n/a | ✅ resolved — user selected `push-main` |
| 28-06 T2 | 06 | 6 | COMPAT-06 | — | Each of the three legs individually green, from one cited run | **manual, CI-observed** | a real push + an observed Actions run | n/a — no local substitute | ✅ pass — run 31976677486 @ `2715a42d`; `build-macos`/`build-linux`/`build-windows` all `success`; independently re-confirmed at audit (one run for that SHA, no `v3.0.0` tag, `release.yml` not fired) |
| 28-06 T3 | 06 | 6 | COMPAT-06 | — | No ledger entry closed on evidence that does not establish it | ledger consistency | manual review of `.planning/WINDOWS.md` | ✅ `.planning/WINDOWS.md` | ✅ pass — exactly one entry closed (#7, live-verified); #4/#5 reworded but open; 10 open |

### Requirement → verification approach (from RESEARCH.md)

| Req ID | Behavior | Test Type | Automated Command |
|--------|----------|-----------|-------------------|
| ACT-06 | Diff categorizes added/removed/changed/unchanged over a synthetic old/new tree pair | unit | `go test ./internal/catalog/... -run TestDiff -v` |
| ACT-06 | A skip-and-continue subtree failure under `MarkUnreadableOnSkip` yields an `unreadable` entry, NOT a `removed` one | unit | `go test ./internal/catalog/... -run TestTraverseDirectory_MarksSkippedNodeWhenFlagSet -v` |
| ACT-06 | The existing `TestTraverseDirectory_SingleEntryErrorSkipsAndContinues` (`service_test.go:462`) still passes **unmodified** after the flag lands | regression | `go test ./internal/catalog/... -run TestTraverseDirectory_SingleEntryErrorSkipsAndContinues -v` |
| ACT-06 | mtime-based `changed` detection for a same-size edit | unit | `go test ./internal/catalog/... -run TestDiff_SameSizeMtimeChange -v` |
| ACT-06 | An old entry with `ModTime == 0` falls back to size-only, never spurious `changed` | unit | `go test ./internal/catalog/... -run TestDiff_MissingOldModTimeFallsBackToSizeOnly -v` |
| ACT-07 | Overwrite reuses `WriteCatalogFrom` (crash-safe write already covered) | regression | `go test ./internal/catalog/... -run TestWriteCatalogFrom -v` |
| ACT-07 | Keep-both invokes the SAME `nextCopyRoot` collision loop, not a copy of it | unit (call-site) | `go test ./internal/catalog/... -run TestDuplicateCatalog -v` plus a new call-site assertion |
| ACT-08 | Volume selection is always presented; nothing is persisted or pre-selected | manual (live dev-browser) + static | dev-browser against `:34115`; grep for any volume persistence |
| STATE-03 | Re-scan's failure path never populates `a.lastPartial` / `a.lastScanReq` | unit (app package) | `go test . -run TestRescan_DoesNotRetainPartial -v` |
| STATE-03 | The unreadable-catalog action trio (re-scan / open html / remove) | manual (live dev-browser) | dev-browser against `:34115` |
| COMPAT-06 | `build.yml` compiles on real macOS / Windows / Linux runners against the full milestone diff | **manual, CI-observed** | a real `git push` + an observed green Actions run — no local command substitutes |

---

## Wave 0 Requirements

- [x] `internal/catalog/diff_test.go` — ACT-06's four-state categorization, the mtime-fallback edge case, and the type-change (file↔directory) edge case
- [x] `MarkUnreadableOnSkip` coverage in `internal/catalog/service_test.go` (or a new file), landing **alongside** the existing skip-and-continue regression test, which must keep passing unmodified
- [x] An `app`-package test proving re-scan's failure path does NOT populate `a.lastPartial`/`a.lastScanReq` — follow Phase 25's `TestStartScan_RetainsPartialOnSourceLoss` pattern in the existing `app_test.go`
- [x] A call-site test asserting keep-both invokes the shared `nextCopyRoot`, not a reimplementation
- [x] No framework install needed — Go stdlib `testing` is already wired
- [x] No new frontend test file — none needed, consistent with TEST-01's deferral

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| The three re-scan entry points, the 3-step dialog, and the diff view | ACT-06, ACT-08 | No frontend test framework (TEST-01 deferred) | `wails dev` on `:34115`; drive via dev-browser from each entry point |
| Volume picker always asks, with no pre-selection | ACT-08 | Absence-of-behavior, visual | Re-scan the same catalog twice; confirm no volume is pre-selected the second time |
| Diff shows all four states with correct counts | ACT-06 | Needs staged filesystem fixtures | Stage a directory with an added file, a removed file, a same-size edited file, and an unreadable subdirectory (`chmod 000`); re-scan and compare counts |
| Low-similarity warning appears and does NOT block | ACT-08 | Visual + interaction | Re-scan a catalog against a completely different directory; confirm the warning appears and resolution is still reachable |
| Overwrite / keep-both / discard each do what they say | ACT-07 | End-to-end across UI → binding → disk | Run all three; confirm overwrite replaces, keep-both writes a `-copy`, discard leaves the original byte-identical |
| The unreadable-catalog action trio | STATE-03 | Needs a genuinely unparseable catalog | Corrupt a catalog's JSON, open it, exercise all three actions |
| **`build.yml` green on all three runners against the full milestone diff** | COMPAT-06 | **Requires a real push — reading the YAML proves nothing** | Push the branch; observe the Actions run to completion on macOS, Windows and Linux |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 120s
- [x] COMPAT-06 backed by an observed CI run, not a local build
- [x] `nyquist_compliant` set honestly in frontmatter (`false` — frontend uncovered, TEST-01 deferred)

**Approval:** reconciled 2026-08-17 by `/gsd-validate-phase 28` — per-task map built from the shipped plans, every automated row re-run green at audit time.

---

## Validation Audit 2026-08-17

| Metric | Count |
|--------|-------|
| Gaps found | 0 fillable |
| Resolved | 0 |
| Escalated | 0 |

**Auditor not spawned** — no fillable gap exists. This pass built the per-task map, which had never
been filled: it still held its plan-time seed row, `*(filled in once PLAN.md task IDs exist)*`, so
this phase reported as NOT-VALIDATED at milestone audit despite having the strongest automated
coverage in the milestone.

**Verified at audit time, not assumed** — every Wave 0 item exists and was re-run:

| Wave 0 item | Evidence |
|-------------|----------|
| `internal/catalog/diff_test.go` | 13 tests green, incl. the sum invariant and a real `chmod 000` end-to-end fixture |
| `MarkUnreadableOnSkip` coverage | `TestTraverseDirectory_MarksSkippedNodeWhenFlagSet`; the pre-existing skip-and-continue regression still passes **unmodified** |
| app-package retained-partial test | `TestRescan_DoesNotRetainPartialForWritePartialCatalog` |
| keep-both call-site assertion | `grep -c WriteFileAtomic resolve.go` = 0; `nextCopyRoot` reused from `duplicate.go` |

**Post-SUMMARY fixes are covered too.** Code review found 1 Critical + 3 Warnings *after* the
SUMMARYs were written; all four are fixed and pinned by tests re-run green here — notably
`TestDiff_NewUnreadableEntryIsUnreadableNotAdded` (CR-01), which was confirmed to FAIL against the
buggy ordering before being accepted, so it is a real regression test rather than a tautology.

**COMPAT-06 stays honest.** The observed run proves native compilation on three real runner OSes
plus TypeScript type-checking. It does not prove signing, notarization, or release, and this file
does not claim it does — those are tag-gated in `release.yml`, and Windows signing skips with a
warning while CRED-04/CRED-05 are unprovisioned.

**Two backstop observations remain open** in `28-UAT.md` (`status: accepted`): volume-enumeration
latency on real removable media, and the untuned `0.6` / 20-entry wrong-disc constants. Both were
flagged `verification: backstop` by the plans themselves — real-world observations, not defects, and
not automatable.
