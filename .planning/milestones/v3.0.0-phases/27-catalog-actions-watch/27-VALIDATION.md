---
phase: 27
slug: catalog-actions-watch
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-15
---

# Phase 27 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by plan-phase from `27-RESEARCH.md`'s `## Validation Architecture`. The Per-Task
> Verification Map is filled in once PLAN.md task IDs exist.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing`, table-driven, `*_test.go` beside source. Frontend: **none by design** — TEST-01 (Vitest + Testing Library) is an explicitly deferred milestone item; do not add one. |
| **Config file** | none — plain `go test ./...` |
| **Quick run command** | `go test ./internal/catalog/... ./internal/osutil/... ./internal/watch/... ./internal/search/... ./pkg/models/... -race -count=1` |
| **Full suite command** | `go build ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build)` |
| **Estimated runtime** | ~90–120 seconds (the SIGKILL subprocess test adds wall-clock over Phase 26's baseline) |

---

## Sampling Rate

- **After every task commit:** the quick run command above, scoped to touched packages
- **After every plan wave:** full suite command, plus a live dev-browser pass against `:34115`
- **Before `/gsd-verify-work`:** full suite green plus the manual-only checks below
- **Max feedback latency:** ~120 seconds

**Dev-server note:** browser verification runs against `wails dev` on **`:34115`**. Vite's `:5173` exposes no `window.go`, so every binding-dependent assertion passes vacuously there. `curl` liveness proves the server is up, not that bindings are fresh — verify `Object.keys(window.go.main.App)` includes the new/changed methods before recording any evidence.

**No host-OS GUI automation.** Do not drive native dialogs or the desktop via osascript / System Events / keystroke injection (STATE.md records a prior-phase incident). Where a check genuinely requires it, record it as a manual item rather than faking the evidence.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 27-01 T1 | 27-01 | 1 | ACT-02 | T-27-02, T-27-03, T-27-10 | Containment gate before any read/write; `html.EscapeString` at both HTML title sites; root-object rebuild preserves the v1 array envelope | unit (tdd) | `go test ./internal/catalog/... -run TestRenameCatalog -race -count=1` | ❌ Wave 0 — created by this task | ✅ pass — `TestRenameCatalog*` 12 passing |
| 27-01 T2 | 27-01 | 1 | ACT-02 | T-27-03 | `html.UnescapeString` on the HTML read path only, never on the JSON-sourced title | unit (tdd) | `go test ./internal/search/... -run TestBrowseCatalogs -race -count=1` | ✅ extends `internal/search/service_test.go` | ✅ pass — `TestBrowseCatalogs*` 15 passing |
| 27-01 T3 | 27-01 | 1 | ACT-02 | T-27-02 | Live rejection of an out-of-directory path | manual (dev-browser, `:34115`) | live binding round trip | n/a — TEST-01 deferred | ✅ pass — live, 27-UAT |
| 27-02 T1 | 27-02 | 1 | ACT-09 | T-27-06, T-27-11 | `tmp.Sync()` before close; best-effort parent-directory `Sync()` that never fails the write | unit (tdd) | `go test ./internal/catalog/... -run TestWriteFileAtomic -race -count=1` | ✅ extends `internal/catalog/atomicwrite_test.go` | ✅ pass — `TestWriteFileAtomic*` 10 passing |
| 27-02 T2 | 27-02 | 1 | ACT-09 | T-27-06, T-27-12, T-27-13 | Real external SIGKILL mid-write leaves the pre-existing catalog byte-identical | integration (subprocess) | `go test ./internal/catalog/... -run TestWriteFileAtomic_SurvivesKill -count=1 -timeout 120s` | ❌ Wave 0 — created by this task | ✅ pass — `TestWriteFileAtomic_SurvivesKill` 2 passing (real subprocess SIGKILL) |
| 27-03 T1 | 27-03 | 2 | ACT-03 | T-27-14, T-27-06 | Suffix candidate free only when both extensions absent; every write via `WriteFileAtomic` | unit (tdd) | `go test ./internal/catalog/... -run TestDuplicateCatalog -race -count=1` | ❌ Wave 0 — created by this task | ✅ pass — `TestDuplicateCatalog*` 11 passing |
| 27-03 T2 | 27-03 | 2 | ACT-04, ACT-05 | T-27-01, T-27-09, T-27-SC | Containment + extension + regular-file gate before the trash seam; no local removal call | unit (tdd, seam-mocked — never a real OS Trash) | `go test ./internal/osutil/... -run TestTrashPaths -race -count=1` | ❌ Wave 0 — created by this task | ✅ pass — `TestTrashPaths*` 11 passing (seam-mocked, no real Trash) |
| 27-03 T3 | 27-03 | 2 | ACT-03, ACT-04, ACT-05 | T-27-02, T-27-18 | Bindings gate on `ContainsPath`; the `.html` companion is derived in Go, never accepted from the renderer | build + typecheck | `go build ./... && go vet ./... && cd frontend && npx tsc --noEmit` | n/a | ✅ pass — `go build`/`go vet`/`tsc --noEmit` green |
| 27-04 T1 | 27-04 | 2 | ACT-01, ACT-02 | — | Danger tokens declared once; delete path boxes structurally cannot ellipsize | selector presence check | the `node -e` selector assertion in the plan's `<verify>` | n/a | ✅ pass — selector assertion green |
| 27-04 T2 | 27-04 | 2 | ACT-01 | T-27-15, T-27-16 | One focus-trap implementation only; no leaked pointer listener; ARIA emitted only when the menu is real | typecheck + build + grep gates | `cd frontend && npx tsc --noEmit && npm run build` | n/a — TEST-01 deferred | ✅ pass — typecheck + build + grep gates green |
| 27-04 T3 | 27-04 | 2 | ACT-01, ACT-02 | T-27-03, T-27-17 | Fail-closed on a null `catalogDir`; titles rendered as JSX text children only | typecheck + build + live | `cd frontend && npx tsc --noEmit && npm run build` + dev-browser | n/a — TEST-01 deferred | ✅ pass — typecheck + build green; live via 27-UAT |
| 27-05 T1 | 27-05 | 3 | ACT-04, ACT-05 | T-27-19, T-27-09 | No permanence vocabulary anywhere in the component; paths never truncated | typecheck + build + grep gates | `cd frontend && npx tsc --noEmit && npm run build` | n/a — TEST-01 deferred | ✅ pass — typecheck + build + grep gates green |
| 27-05 T2 | 27-05 | 3 | ACT-03, ACT-04 | T-27-01, T-27-18, T-27-20 | Paths always originate from the `BrowseCatalogs` listing; double-submit guarded | typecheck + build + live | `cd frontend && npx tsc --noEmit && npm run build` + dev-browser | n/a — TEST-01 deferred | ✅ pass — typecheck + build green; live via 27-UAT |
| 27-06 T1 | 27-06 | 3 | WATCH-02, WATCH-03 | T-27-04, T-27-05, T-27-21, T-27-22, T-27-SC | Errors channel drained in the same select; `Close()` once-guarded; no Wails import | unit (tdd, fake source + real temp dir) | `go test ./internal/watch/... -race -count=1 -timeout 60s` | ❌ Wave 0 — created by this task | ✅ pass — `internal/watch` 13 passing under `-race` |
| 27-06 T2 | 27-06 | 3 | WATCH-02, WATCH-03 | T-27-05, T-27-22 | `app.go` is the sole `runtime.EventsEmit` caller; emit guarded by `a.ctx == nil`; release on every quit path | build + repo-wide grep + live | `go build ./... && go vet ./... && go test ./... -race -count=1` | n/a | ✅ pass — full suite green; single-emitter grep holds |
| 27-07 T1 | 27-07 | 4 | WATCH-01, WATCH-02 | T-27-23, T-27-24, T-27-25 | Unsubscribe returned from the effect; no second rail read path | typecheck + build + grep gates | `cd frontend && npx tsc --noEmit && npm run build` | n/a — TEST-01 deferred | ✅ pass — typecheck + build + grep gates green |
| 27-07 T2 | 27-07 | 4 | WATCH-01, WATCH-02, WATCH-03 | T-27-26 | No ledger entry claims coverage that was not exercised | ledger consistency script | the `node -e` ledger assertion in the plan's `<verify>` | ✅ `.planning/WINDOWS.md` | ✅ pass — ledger assertion green against `.planning/WINDOWS.md` |
| 27-07 T3 | 27-07 | 4 | all 9 | all | The full 28-row phase matrix, live | manual (dev-browser, `:34115`) | live verification | n/a — TEST-01 deferred | ✅ pass — 27-UAT: 6 pass / 0 issues / 1 blocked (Windows/Linux, no host) |

### Requirement → verification approach (from RESEARCH.md)

| Req ID | Behavior | Test Type | Automated Command |
|--------|----------|-----------|-------------------|
| ACT-01 | Menu opens/closes, arrow-key nav, Escape, click-outside, focus restore to `⋯` | manual (live dev-browser) | dev-browser against `:34115` |
| ACT-02 | Rename writes the JSON title, rewrites **both** HTML title occurrences (`<title>` and `<h1>`), handles the no-`.html` case, round-trips `&`/special characters | unit | `go test ./internal/catalog/... -run TestRenameCatalog -v` |
| ACT-02 | `BrowseCatalogs` title-read fix (`html.UnescapeString`) | unit | `go test ./internal/search/... -run TestBrowseCatalogs_UnescapesTitle -v` |
| ACT-03 | Duplicate suffixes `-copy` / `-copy-2` / `-copy-3` across collisions; title inherited verbatim | unit | `go test ./internal/catalog/... -run TestDuplicateCatalog -v` |
| ACT-04, ACT-05 | Trash never falls back to permanent delete on error; every path containment-gated | unit (trash call behind a small interface — CI must not touch a real OS Trash) | `go test ./internal/osutil/... -run TestTrash -v` |
| ACT-09 | `WriteFileAtomic` calls `File.Sync()` before close+rename | unit | `go test ./internal/catalog/... -run TestWriteFileAtomic_Syncs -v` |
| ACT-09 | A real SIGKILL mid-write leaves the destination uncorrupted | integration (subprocess) | `go test ./internal/catalog/... -run TestWriteFileAtomic_SurvivesKill -v -timeout 60s` |
| WATCH-01 | Status bar shows the watching segment when the setting and directory are both set | manual (live dev-browser) | live verification |
| WATCH-02 | External add/remove/modify triggers a debounced `catalogs:changed` → rail refresh | unit (fake event source for the debounce) + manual live | `go test ./internal/watch/... -v` |
| WATCH-03 | `SetWatchDirectory(false)` and app quit both genuinely `Close()` the watcher | unit | `go test ./internal/watch/... -run TestWatcher_Close -v` |

---

## Wave 0 Requirements

- [x] `internal/catalog/rename_test.go` — ACT-02, including the two-HTML-occurrence case and special-character round-trip
- [x] `internal/catalog/duplicate_test.go` — ACT-03 collision sequence
- [x] `internal/osutil/trash_test.go` — ACT-04/ACT-05, with the trash call behind a small interface so CI never touches a real OS Trash
- [x] `internal/watch/watcher_test.go` — WATCH-02/WATCH-03. Test the debounce logic in isolation against a fake event source; reserve real `fsnotify` behavior for live verification (fsnotify against a real filesystem is unreliable in CI sandboxes)
- [x] `internal/catalog/atomicwrite_sigkill_test.go` + a standalone helper binary — ACT-09's crash-safety claim with an actual kill. **This closes `WINDOWS.md` #6**, which currently records the guarantee as unit-tested only
- [x] `internal/search/service_test.go` gains a title-unescape case — **confirmed at plan time: the file already exists**, so plan `27-01` Task 2 extends it rather than creating it
- [x] No new frontend test file — none needed, consistent with TEST-01's deferral and every prior phase's precedent

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Actions menu: open from `⋯`, arrow-key nav, Escape, click-outside, focus returns to the trigger | ACT-01 | No frontend test framework (TEST-01 deferred) | `wails dev` on `:34115`; dev-browser: open the menu from a selected catalog, walk items with arrows, close each way, assert focus lands back on `⋯` |
| Rename round-trip through the UI with a special-character title | ACT-02 | End-to-end across dialog → binding → disk → rail re-read | Rename a catalog to `Tom & Jerry <2024>`; confirm the rail, details panel, and the written `.html` all show it correctly and none shows `&amp;` |
| Delete confirmation names both real paths; HTML checkbox present-and-checked when an `.html` exists, absent when it does not | ACT-04 | Visual/copy correctness | Open delete on a catalog with an `.html`, then on one without; compare both dialogs |
| A failed Trash surfaces the real error with no permanent-delete affordance | ACT-05 | Requires inducing a genuine trash failure | Make the target undeletable (e.g. permissions), attempt delete, confirm the error text and that no "delete permanently" path exists anywhere in the state |
| "● watching `<dir>`" appears when enabled and disappears when disabled | WATCH-01 | Visual | Toggle the Settings watch switch; observe the status bar |
| Rail updates when catalogs change outside the app | WATCH-02 | Requires real external filesystem activity | With watching on, add / remove / modify a `.json` in the catalog directory from a terminal; confirm the rail reflects each within ~1s |
| Watcher is genuinely released on toggle-off and on quit | WATCH-03 | Release is not observable from the UI alone | Toggle off, then confirm via the unit test plus (optionally) `lsof` that no watch handle remains |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 120s
- [x] `nyquist_compliant` set honestly in frontmatter (`false` — frontend uncovered, TEST-01 deferred)

**Approval:** reconciled 2026-08-17 by `/gsd-validate-phase 27` against the shipped code — every automated row re-run green at audit time, manual rows closed by 27-UAT.

---

## Validation Audit 2026-08-17

| Metric | Count |
|--------|-------|
| Gaps found | 0 fillable |
| Resolved | 0 |
| Escalated | 0 |

**Auditor not spawned** — every Wave 0 item had already landed during execution. This file had simply
never been reconciled after the phase shipped: all 18 task rows still read `⬜ pending` and the Wave 0
boxes were unchecked, three days after the tests they describe were written and merged. That is why
`/gsd-audit-milestone` reported this phase as NOT-VALIDATED — coverage unknown, not coverage missing.

**Verified at audit time, not assumed.** Every Wave 0 file exists and every named test was re-run:

| Wave 0 file | Evidence |
|-------------|----------|
| `internal/catalog/rename_test.go` | `TestRenameCatalog*` — 12 passing |
| `internal/catalog/duplicate_test.go` | `TestDuplicateCatalog*` — 11 passing |
| `internal/osutil/trash_test.go` | `TestTrashPaths*` — 11 passing, seam-mocked |
| `internal/watch/watcher_test.go` | 13 passing under `-race -timeout 60s` |
| `internal/catalog/atomicwrite_sigkill_test.go` | `TestWriteFileAtomic_SurvivesKill` — 2 passing, real subprocess kill |
| `internal/search/service_test.go` | `TestBrowseCatalogs*` — 15 passing |

**Manual-only rows** are closed by `27-UAT.md` (`status: passed` — 6 pass, 0 issues, 1 blocked).
The blocked row is Windows/Linux platform behavior, which no macOS host can exercise; `WINDOWS.md`
8–11 stay open for exactly that reason.

**One defect this coverage did not catch,** recorded so the gap in the *approach* is visible: the
`CatalogRail` stale-response race was found by live UAT at milestone close, not by any automated
test — because the affected code is frontend, where TEST-01's deferral means there is no automated
layer to catch it. Fixed in `6d0597f5`; re-verified live. This is the concrete cost of the deferral,
and the reason `nyquist_compliant` stays `false` rather than being read as "good enough."
