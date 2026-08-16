---
phase: 28
slug: re-scan-diff
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
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
| *(filled in once PLAN.md task IDs exist)* | — | — | — | — | — | — | — | — | ⬜ pending |

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

- [ ] `internal/catalog/diff_test.go` — ACT-06's four-state categorization, the mtime-fallback edge case, and the type-change (file↔directory) edge case
- [ ] `MarkUnreadableOnSkip` coverage in `internal/catalog/service_test.go` (or a new file), landing **alongside** the existing skip-and-continue regression test, which must keep passing unmodified
- [ ] An `app`-package test proving re-scan's failure path does NOT populate `a.lastPartial`/`a.lastScanReq` — follow Phase 25's `TestStartScan_RetainsPartialOnSourceLoss` pattern in the existing `app_test.go`
- [ ] A call-site test asserting keep-both invokes the shared `nextCopyRoot`, not a reimplementation
- [ ] No framework install needed — Go stdlib `testing` is already wired
- [ ] No new frontend test file — none needed, consistent with TEST-01's deferral

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

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] COMPAT-06 backed by an observed CI run, not a local build
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
