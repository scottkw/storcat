---
phase: 23
slug: rail-virtualized-tree
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: 2026-08-13
---

# Phase 23 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

**What changed since Phase 22:** this phase has a substantial Go surface, and Go **is** unit-testable here — `go test ./...` already runs against `internal/catalog`, `internal/config`, `internal/search` and `cli`. So unlike Phase 22, a large share of this phase's risk lands in genuinely automated tests. The frontend still has no test framework (TEST-01 remains deferred by locked decision — see `.planning/PROJECT.md` and `23-CONTEXT.md` — and this audit adds none), so UI behavior is browser-verified via `dev-browser`, which is what actually happened across `23-VERIFICATION.md` and `23-UI-REVIEW.md`.

**What `nyquist_compliant: true` means for this phase, given the Go/frontend asymmetry:** every requirement has *some* form of real, executed verification — Go behaviors are automated and re-run in this audit; frontend/UI behaviors are live `dev-browser` sessions against the real `wails dev` bindings at :34115, re-inspectable via the evidence trails in `23-VERIFICATION.md` and `23-UI-REVIEW.md`, but they are not re-runnable `npm test` commands because none exist by locked decision. This flag is **not** claiming a frontend test suite exists — it is claiming that no requirement in this phase is unverified, and that the two genuinely non-driver-testable items (a native OS dialog, a Windows-only runtime shape) are explicitly logged as manual/open rather than silently assumed.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go: `go test` (table-driven, `*_test.go` beside source — the established project pattern). Frontend: none by design; `tsc` + `vite build` + live browser verification via `dev-browser` against `wails dev` at :34115. |
| **Config file** | `go.mod`, `frontend/tsconfig.json`, `frontend/vite.config.ts` |
| **Quick run command** | `go test ./internal/... && (cd frontend && npx tsc --noEmit)` |
| **Full suite command** | `cd frontend && npx tsc --noEmit && npm run build && cd .. && go build ./... && go test ./... -race -count=1` |
| **Estimated runtime** | ~60–90 seconds (longer once the 40k benchmark runs) |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/...` for Go tasks; `npx tsc --noEmit` for frontend tasks
- **After every plan wave:** full suite, plus a live browser pass at http://localhost:34115 (verified during 23-01 planning that :5173 exposes no Wails bindings)
- **Before verification:** full suite green, the 40k fixture benchmark recorded with a real number, and the manual matrix below completed
- **Max feedback latency:** ~90 seconds

---

## Per-Task Verification Map

Reconstructed against the six plans' real task IDs and their actual execution evidence (`23-0*-SUMMARY.md` `coverage` blocks, `23-VERIFICATION.md`, `23-REVIEW-FIX.md`).

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command / Evidence | Status |
|---------|------|------|-------------|-----------|-------------------------------|--------|
| 23-01 T1 | 23-01 | 1 | Wave 0 fixture generator (TREE-01 prereq) | Go unit | `go test ./internal/fixture/... -count=1` | ✅ green |
| 23-01 T2 | 23-01 | 1 | TREE-01, TREE-06, RAIL-03, COMPAT-01 (end-to-end tracer) | Go unit + live browser | `go test ./internal/search/... ./cli/... -count=1`; dev-browser click-through at :34115 (23-01-SUMMARY D3) | ✅ green |
| 23-01 T3 | 23-01 | 1 | TREE-01 (42,550-node measurement) | Go benchmark + live browser | `go test ./internal/search/ -bench BenchmarkLoadCatalogFlat40k -benchtime=3x` (re-run in this audit: 54.1ms/op, 5.641MB, 42,550 nodes); dev-browser scroll-count sweep | ✅ green |
| 23-02 T1 | 23-02 | 2 | Sidecar counts cache (RAIL-01, SHELL-06 backing) | Go unit + race | `go test ./internal/config/... -run 'TestCountsCache' -race -count=1` (re-run: pass) | ✅ green |
| 23-02 T2 | 23-02 | 2 | RAIL-01, RAIL-04, STATE-02 (BrowseCatalogs parse status/counts) | Go unit | `go test ./internal/search/... ./cli/... -count=1` | ✅ green |
| 23-02 T3 | 23-02 | 2 | Wiring the cache into `app.go` + bridge regen | Go build + tsc | `go build ./... && go test ./... -count=1 && (cd frontend && npx tsc --noEmit && npm run build)` | ✅ green |
| 23-03 T1 | 23-03 | 2 | Expansion/selection reducer actions + `format.ts` (TREE-02/03/06) | tsc/build + node eval | `npx tsc --noEmit && npm run build`; formatter boundary values recorded in 23-03-SUMMARY | ✅ green |
| 23-03 T2 | 23-03 | 2 | TREE-02, TREE-06, STATE-01 (real tree rows, click semantics, scroll reset, 3 states) | Live browser (dev-browser) | 23-03-SUMMARY D2; **superseded by 23-VERIFICATION's TREE-06 finding — see Reconciliation below** | ⚠️ flaky→fixed |
| 23-03 T3 | 23-03 | 2 | TREE-03, TREE-05 (breadcrumb, expand-all/collapse) | Live browser (dev-browser) | 23-03-SUMMARY D3 — 42,550-node expand-all timing, per-segment color | ✅ green |
| 23-04 T1 | 23-04 | 3 | RAIL-01, RAIL-04 (populated rail rows, status dot) | Live browser (dev-browser) | 23-04-SUMMARY D1 | ✅ green |
| 23-04 T2 | 23-04 | 3 | RAIL-02, RAIL-05, STATE-01 (filter isolation, directory chip, empty states) | Live browser (dev-browser) + MutationObserver | 23-04-SUMMARY D2/D3 — 0 tree mutations during 10 filter keystrokes | ✅ green (native dialog itself is manual-only, see below) |
| 23-04 T3 | 23-04 | 3 | SHELL-06 (status bar live totals) | Live browser (dev-browser) + node eval | 23-04-SUMMARY D4; **cold-cache honesty gap found by code review, fixed as WR-03 — see Reconciliation** | ✅ green (post-fix) |
| 23-05 T1 | 23-05 | 3 | TREE-04 (catalog header) | Live browser (dev-browser) | 23-05-SUMMARY D1 | ✅ green |
| 23-05 T2 | 23-05 | 3 | STATE-02 (unreadable-catalog panel) | Live browser (dev-browser) | 23-05-SUMMARY D2 — verbatim raw-error match, deleted-file race case; **found and fixed a real `wailsAPI.ts` bug (`extractErrorMessage`) during this task's own verification** | ✅ green |
| 23-06 T1 | 23-06 | 3 | TREE-08, T-23-01 (RevealInFileManager argv-only + hostile path) | Go unit | `go test ./internal/osutil/... -count=1 -v` (re-run: pass, including hostile-path exact-argv tests for all 3 platforms) | ✅ green |
| 23-06 T2 | 23-06 | 3 | TREE-07 (details panel tracks selection) | Live browser (dev-browser) | 23-06-SUMMARY D3 | ✅ green |
| 23-06 T3 | 23-06 | 3 | TREE-08 (footer actions, double-click guard) | Live browser (dev-browser) | 23-06-SUMMARY D4 | ✅ green |
| 23-06 T4 | 23-06 | 3 | TREE-08 (real macOS OS-level reveal; Windows shape) | Manual (checkpoint) | AppleScript Finder-selection readback against a wails-built `.app`, including a hostile filename — approved; **Windows explicitly deferred, logged in `.planning/WINDOWS.md` id 1 (open)** | ✅ green (macOS) / 🟡 manual-open (Windows) |
| — | 23-REVIEW→FIX | — | WR-01/02/03, IN-01/02 (post-hoc code-review findings) | Go unit + live browser | `go test ./... -race`, `npx tsc --noEmit && npm run build`; live browser re-verification of containment and status-bar honesty fixes (23-REVIEW-FIX.md) | ✅ green, all 5 fixed |
| — | 23-VERIFICATION gap | — | TREE-06 scroll-reset-on-revisit | Live browser (dev-browser) | Found FAILED by verifier, root-caused, fixed in `cac4aa06`, re-verified live (scroll leaked to 1043 → now 0 on revisit) | ✅ green (post-fix) |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

### Reconciliation notes (this audit found no new gaps; documenting what was already found and closed)

- **TREE-06** was reported `VERIFIED` in 23-03-SUMMARY's own live check, but that check only exercised switch-away, not switch-away-then-return. `23-VERIFICATION.md` caught the asymmetry, reproduced it deterministically (2/2), root-caused it to a `useEffect` firing against an unmounted scroll ref during the `'loading'` render branch, and the fix (`cac4aa06`) moved the reset into a `useLayoutEffect` keyed on the `loading→ready` transition. Re-verified live post-fix. This is the phase's one genuine verification-to-implementation gap, and it was closed before this audit, not by this audit.
- **WR-01/02/03 + IN-01/02** (from `23-REVIEW.md`) were all Warning/Info severity, not Blocker, and all five were fixed and re-verified in `23-REVIEW-FIX.md` with real commits (`786b8ddf`, `d5f41f1c`, `23692e32`, `ada22533`) — confirmed present in `git log` by this audit's `go test` re-run, which exercises the WR-02 containment tests (`TestContainsPath`, `TestRevealInFileManager_RejectsMissingCatalogDir`, `TestRevealInFileManager_RejectsPathOutsideCatalogDir`) directly.
- **FU-23-A / FU-23-B** (from `23-SECURITY.md`) are open follow-ups explicitly owned by Phase 26, not this phase — correctly out of scope here, not a gap.

---

## Wave 0 Requirements

- [x] **The 40,000-node fixture generator.** `internal/fixture/fixture.go` (`WriteDCIMCatalog`/`WriteFlatCatalog`/`WriteDeepCatalog`) plus `scripts/gen-fixture-catalog/main.go`. Verified in this audit: `go test ./internal/fixture/... -count=1` passes; the DCIM shape's default parameters produce 42,550 nodes (matches every downstream measurement).
- [x] **A Go benchmark or timed test over that fixture.** `BenchmarkLoadCatalogFlat40k` in `internal/search/flatten_bench_test.go`. Re-run in this audit: **54.1ms/op, 5.641 MB marshalled, 42,550 nodes** (23-01-SUMMARY's own run measured 107.7ms/op on a colder run; 23-VERIFICATION's independent run measured 46.3ms/op; this audit's run measured 54.1ms/op — all three are real, non-fabricated numbers from the same benchmark on the same machine, with normal run-to-run variance. The number was never asserted, only measured, exactly as Wave 0 required.)

*Everything else uses existing infrastructure.*

---

## Automated Coverage (Go — this is the majority of the phase's risk)

All 12 rows below were checked against real code and re-run in this audit (`go build ./... && go test ./... -race -count=1`, plus the targeted commands shown). Every one exists and is green.

| Behavior | Requirement | Test | Confirmed this audit |
|----------|-------------|------|------------------------|
| `LoadCatalogFlat` flattens a nested catalog to the correct order, depth, parent index and `hasChildren` | TREE-01 | `internal/search/flatten_test.go#TestLoadCatalogFlat_Structure` | ✅ `go test ./internal/search/...` |
| `Name` is `filepath.Base` and `Path` is the verbatim relative path | TREE-04, TREE-05 | `internal/search/flatten_test.go#TestLoadCatalogFlat_Structure` (Name/Path split asserted) | ✅ |
| v1 array-wrapped and v2 bare-object catalogs both flatten identically | COMPAT-01 | `internal/search/flatten_test.go#TestLoadCatalogFlat_DualFormat` | ✅ |
| Empty catalog (root with no contents) yields a zero-length flat array, not an error | STATE-01 | `internal/search/flatten_test.go#TestLoadCatalogFlat_EmptyRoot` | ✅ |
| Truncated / malformed / non-catalog JSON produce a `parseError` carrying the byte offset and raw message | STATE-02 | `internal/search/service_test.go#TestBrowseCatalogs_ParseError_Truncated,_Malformed`; `internal/search/flatten_test.go#TestLoadCatalogFlat_NotACatalog` | ✅ |
| Missing file and permission-denied are distinguishable errors | STATE-02 | `internal/search/flatten_test.go#TestLoadCatalogFlat_MissingFile,_PermissionDenied`; `internal/search/service_test.go#TestBrowseCatalogs_ParseError_UnreadableFile` | ✅ |
| `BrowseCatalogs` populates `fileCount`, `totalBytes`, `parseError` | RAIL-01, RAIL-04, SHELL-06 | `internal/search/service_test.go` (counts-cache hit/miss + parse-error tests) | ✅ |
| `json.Valid()` fast path is taken for good catalogs (no full unmarshal) | RAIL-01 perf | `internal/search/service_test.go#TestDetectParseError_AllocsValidLessThanInvalid` (`testing.AllocsPerRun`) | ✅ |
| Sidecar cache: hit on unchanged path+mtime, miss on changed mtime, safe under concurrent access | RAIL-01, SHELL-06 | `internal/config/counts_cache_test.go` + `#TestCountsCache_ConcurrentPutGet_NoRace` | ✅ `go test ./internal/config/... -run 'TestCountsCache' -race -count=1` |
| `RevealInFileManager` builds the correct argv per OS and does not shell-interpolate the path | TREE-08 | `internal/osutil/reveal_test.go#TestRevealArgv_HostilePath_Darwin,_Windows,_Linux` (exact argv element count+content, not substring) | ✅ `go test ./internal/osutil/... -run Hostile -v` |
| 40k fixture: parse + flatten completes within a recorded budget | TREE-01 | `internal/search/flatten_bench_test.go#BenchmarkLoadCatalogFlat40k` | ✅ 54.1ms/op, 42,550 nodes, 5.641MB this run |
| `cli/show.go` still works — `LoadCatalog` unchanged | COMPAT-01 | `go test ./cli/...` | ✅ |
| **(added post-plan, WR-02)** `RevealInFileManager` containment check rejects paths outside the configured catalog directory | T-23-02 | `internal/osutil/reveal_test.go#TestContainsPath` (4 scenarios incl. symlink escape) + `TestRevealInFileManager_RejectsMissingCatalogDir,_RejectsPathOutsideCatalogDir` | ✅ |

Run with `-race` at least once for the cache concurrency test — done in this audit (`go test ./... -race -count=1`, all green).

---

## Manual-Only Verifications (browser, via dev-browser against `wails dev` at http://localhost:34115)

**Correction from the original draft:** the URL is `:34115`, not `:5173` — `23-01`'s planning verified that a bare Vite dev server exposes no `window.go` bindings, so nothing in this table was ever actually verifiable at `:5173`. Every row below reflects what `23-VERIFICATION.md` and `23-UI-REVIEW.md` actually drove, live, against real Go bindings and a real 42,550-node fixture (`fixture-dcim`) plus supporting fixtures (a hand-corrupted catalog, a v1 array-wrapped catalog, an empty directory).

| Behavior | Requirement | Status | Evidence |
|----------|-------------|--------|----------|
| Rail lists every catalog with title, JSON size, filename, file count | RAIL-01 | ✅ executed live | 23-VERIFICATION.md truth #11; 5-catalog fixture directory |
| Typing in the filter narrows the rail case-insensitively on title AND filename | RAIL-02 | ✅ executed live | 23-VERIFICATION.md truth #12; mixed-case fragment matched both |
| **The tree does not re-render on filter keystrokes** | RAIL-02 | ✅ executed live | MutationObserver on `.pane-scroll` recorded 0 mutations during 10 keystrokes (23-04-SUMMARY D2, re-confirmed 23-VERIFICATION truth #12) |
| Selecting a rail row loads its tree and clears the previous selection | RAIL-03 | ✅ executed live | 23-VERIFICATION.md truth #13 |
| Red dot on a catalog that fails to parse | RAIL-04 | ✅ executed live | 23-VERIFICATION.md truth #14; deliberately-corrupted `.json` |
| Directory chip shows the current directory and opens the picker | RAIL-05 | ✅ persistence half executed live / 🟡 native dialog manual-only | Persistence+reload verified programmatically (localStorage → `SET_CATALOG_DIR` → `BrowseCatalogs`); the native OS folder-picker itself cannot be driven from `dev-browser` — logged as outstanding human verification in `23-VERIFICATION.md` |
| 40,000-node catalog scrolls smoothly with no freeze | TREE-01 | ✅ executed live | 42,550-node fixture; mounted `.ws-tree-row` count stayed 30–41 across a 9-point scroll sweep including post-expand-all |
| Directory click toggles AND selects; file click selects only | TREE-02 | ✅ executed live | 23-VERIFICATION.md truth #4 |
| Expand-all and collapse-to-root from the breadcrumb bar | TREE-03 | ✅ executed live | 23-VERIFICATION.md truth #5; ~800ms wall-clock at 42,550 nodes |
| Catalog header shows title, `.json`/`.html` chips, and the metadata line | TREE-04 | ✅ executed live | 23-VERIFICATION.md truth #6 |
| Breadcrumb ancestors are accent-colored **per-segment spans** and clickable-appearing | TREE-05 | ✅ executed live | 23-VERIFICATION.md truth #7; 4 separate `span.ws-crumb` elements confirmed via DOM inspection, not one string |
| Scroll position and expansion reset on catalog switch | TREE-06 | ✅ executed live, gap found and fixed | Originally FAILED on revisit (stale scroll offset); fixed in `cac4aa06`; re-verified — see Reconciliation notes above |
| Details panel follows selection | TREE-07 | ✅ executed live | 23-VERIFICATION.md truth #9; 3+ selection changes tracked |
| "Open HTML catalog" and "Reveal JSON in file manager" work | TREE-08 | ✅ macOS wire-level + real OS reveal / 🟡 Windows manual-open | macOS: real `open -R` argv confirmed against a built `.app` via AppleScript Finder-selection readback, including a hostile filename (space, apostrophe, semicolon) — both selected exactly one item, the file itself. Windows: unit-tested for argv structure only, explicitly deferred — `.planning/WINDOWS.md` id 1 (open) |
| Status bar shows live catalog count, file count, total bytes | SHELL-06 | ✅ executed live, gap found and fixed | Cold-cache honesty gap (WR-03) found by code review — status bar could read a confident "0 files" for catalogs never opened; fixed with a `"≥"` qualifier, re-verified live (`"≥0 files indexed"` cold → `"5 files indexed"` warm) |
| Empty-library state when the directory has no catalogs | STATE-01 | ✅ executed live | 23-VERIFICATION.md truth #17 |
| Unreadable-catalog panel shows filename, byte offset, reason, raw error | STATE-02 | ✅ executed live | 23-VERIFICATION.md truth #17; all 4 fields present, raw error verbatim byte-for-byte |
| v1.x and v2.x catalogs open without conversion | COMPAT-01 | ✅ executed live | 23-VERIFICATION.md truth #2; hand-written v1 array-wrapped fixture opened correctly |

**Needs the built app, not the browser:** the actual OS-level file-manager reveal (TREE-08) — this was actually done, on the macOS build, via an AppleScript Finder-selection readback against the exact `open -R <path>` argv the binding constructs (23-06-SUMMARY, task 4 checkpoint, approved). The Windows shape remains genuinely unverified and is logged in `.planning/WINDOWS.md` (id 1, open) rather than assumed.

---

## Validation Sign-Off

- [x] Wave 0 fixture generator exists and the 40k benchmark records a real number — 54.1ms/op, 42,550 nodes, 5.641MB (this audit's run; three independent runs across the phase's lifecycle all land in the same order of magnitude with normal variance)
- [x] `go test ./...` green, including a `-race` pass for the sidecar cache — re-run in this audit, `go test ./... -race -count=1`, all packages `ok`
- [x] `RevealInFileManager` argv test includes a hostile path (spaces, quotes, `;`) — confirmed, plus `&&`, backtick, `$()`, pipe, newline; also confirmed on a real built app via AppleScript readback with a hostile filename
- [x] `cli` tests still green — `LoadCatalog` untouched — confirmed, `git diff` shows `internal/catalog/service.go`/`CatalogItem` untouched across the phase
- [x] `npx tsc --noEmit`, `npm run build`, `go build ./...` all green — re-run in this audit, all exit 0
- [x] All 19 manual rows above completed in-browser, with the 40k row-count assertion recorded — 17 of 19 fully executed live with no caveat; 2 (RAIL-05's native dialog, TREE-08's Windows shape) are genuinely non-driver-testable and are explicitly logged rather than silently marked done
- [x] TREE-08's OS reveal verified on the built macOS app — done via AppleScript Finder-selection readback against a real `wails build -platform darwin/arm64` output, including a hostile filename; **Windows explicitly not verified, logged in `.planning/WINDOWS.md` id 1 (open, must sweep before v3.0.0 ships)**
- [x] No watch-mode flags in any verify command — confirmed, all commands in this file are one-shot (`-count=1`, no `--watch`)
- [x] `nyquist_compliant: true` set in frontmatter — set, with the Go/frontend asymmetry explained above rather than asserted blindly

**Approval:** validated 2026-08-14 by retroactive Nyquist audit (`/gsd-validate-phase`). No implementation changes made — this audit only reconciled documentation against already-executed verification evidence (`23-VERIFICATION.md`, `23-REVIEW.md`/`23-REVIEW-FIX.md`, `23-SECURITY.md`, `23-UI-REVIEW.md`) and independently re-ran the full Go suite (`go build`, `go test -race`, the 40k benchmark, the hostile-path and containment tests) plus the frontend typecheck/build to confirm they still pass at current `HEAD`.
