---
phase: 23
slug: rail-virtualized-tree
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-13
---

# Phase 23 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

**What changed since Phase 22:** this phase has a substantial Go surface, and Go **is** unit-testable here — `go test ./...` already runs against `internal/catalog`, `internal/config`, `internal/search` and `cli`. So unlike Phase 22, a large share of this phase's risk lands in genuinely automated tests. The frontend still has no test framework (TEST-01 remains deferred; do NOT add Vitest), so UI behavior is browser-verified via `dev-browser`.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go: `go test` (table-driven, `*_test.go` beside source — the established project pattern). Frontend: none by design; `tsc` + `vite build` + live browser verification. |
| **Config file** | `go.mod`, `frontend/tsconfig.json`, `frontend/vite.config.ts` |
| **Quick run command** | `go test ./internal/... && (cd frontend && npx tsc --noEmit)` |
| **Full suite command** | `cd frontend && npx tsc --noEmit && npm run build && cd .. && go build ./... && go test ./...` |
| **Estimated runtime** | ~60–90 seconds (longer once the 40k benchmark runs) |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/...` for Go tasks; `npx tsc --noEmit` for frontend tasks
- **After every plan wave:** full suite, plus a live browser pass at http://localhost:5173
- **Before verification:** full suite green, the 40k fixture benchmark recorded with a real number, and the manual matrix below completed
- **Max feedback latency:** ~90 seconds

---

## Per-Task Verification Map

Filled by the planner against real task IDs. Every task carries either an automated command or an explicit Manual-Only row — no task may claim verification it cannot produce.

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| *(planner fills)* | | | | | | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] **The 40,000-node fixture generator.** A Go test helper (plus a small committed script for manual/browser use) that writes a synthetic catalog of ≥40,000 nodes into a temp dir on demand. Nothing downstream can honestly claim TREE-01 without it, and the 40k blob is deliberately NOT committed to the repo.
- [ ] **A Go benchmark or timed test** over that fixture covering parse → flatten → marshal, so the 5.83 MB / time-to-first-row figures are measured rather than asserted.

*Everything else uses existing infrastructure.*

---

## Automated Coverage (Go — this is the majority of the phase's risk)

| Behavior | Requirement | Test |
|----------|-------------|------|
| `LoadCatalogFlat` flattens a nested catalog to the correct order, depth, parent index and `hasChildren` | TREE-01 | Go table test, `internal/catalog` |
| `Name` is `filepath.Base` and `Path` is the verbatim relative path | TREE-04, TREE-05 | Go table test — this is the corrected finding from research; a regression here silently breaks row labels and breadcrumbs |
| v1 array-wrapped and v2 bare-object catalogs both flatten identically | COMPAT-01 | Go table test with one fixture of each format |
| Empty catalog (root with no contents) yields a zero-length flat array, not an error | STATE-01 | Go table test |
| Truncated / malformed / non-catalog JSON produce a `parseError` carrying the byte offset and raw message | STATE-02 | Go table test, one case per failure mode |
| Missing file and permission-denied are distinguishable errors | STATE-02 | Go table test |
| `BrowseCatalogs` populates `fileCount`, `totalBytes`, `parseError` | RAIL-01, RAIL-04, SHELL-06 | Go test |
| `json.Valid()` fast path is taken for good catalogs (no full unmarshal) | RAIL-01 perf | Go test asserting the fast path, or a benchmark showing the cost |
| Sidecar cache: hit on unchanged path+mtime, miss on changed mtime, safe under concurrent access | RAIL-01, SHELL-06 | Go table test + a `-race` test for the concurrent case |
| `RevealInFileManager` builds the correct argv per OS and does not shell-interpolate the path | TREE-08 | Go table test asserting argv construction; **must include a path containing spaces, quotes and a `;`** — this is a new process-spawning surface |
| 40k fixture: parse + flatten completes within a recorded budget | TREE-01 | Go benchmark over the Wave 0 fixture |
| `cli/show.go` still works — `LoadCatalog` unchanged | COMPAT-01 | existing `cli` tests must stay green |

Run with `-race` at least once for the cache concurrency test.

---

## Manual-Only Verifications (browser, via dev-browser at http://localhost:5173)

| Behavior | Requirement | Test Instructions |
|----------|-------------|-------------------|
| Rail lists every catalog with title, JSON size, filename, file count | RAIL-01 | Point at a directory with several catalogs; read the rendered rows |
| Typing in the filter narrows the rail case-insensitively on title AND filename | RAIL-02 | Type a mixed-case fragment matching one by title and one by filename |
| **The tree does not re-render on filter keystrokes** | RAIL-02 | Load a catalog, then type in the filter; assert the tree's rendered row identities/count are unchanged during typing |
| Selecting a rail row loads its tree and clears the previous selection | RAIL-03 | Select A, expand and select a node, select B; assert selection is null and expansion empty |
| Red dot on a catalog that fails to parse | RAIL-04 | Put a deliberately corrupt `.json` in the directory |
| Directory chip shows the current directory and opens the picker | RAIL-05 | Observe the chip; the native dialog itself is not browser-testable — verify the handler is wired |
| 40,000-node catalog scrolls smoothly with no freeze | TREE-01 | Load the generated fixture; assert rendered DOM row count stays proportional to viewport (not 40k) while scrolling, and record scroll timing |
| Directory click toggles AND selects; file click selects only | TREE-02 | Click one of each; assert expansion and selection state after each |
| Expand-all and collapse-to-root from the breadcrumb bar | TREE-03 | Trigger both on the fixture; assert visible row count changes and stays responsive |
| Catalog header shows title, `.json`/`.html` chips, and the metadata line | TREE-04 | Read the rendered header |
| Breadcrumb ancestors are accent-colored **per-segment spans** and clickable | TREE-05 | Assert multiple span elements, not one string; check computed color |
| Scroll position and expansion reset on catalog switch | TREE-06 | Scroll deep, expand several, switch catalog; assert scrollTop 0 and empty expansion |
| Details panel follows selection | TREE-07 | Select several nodes in turn; assert the panel's values track |
| "Open HTML catalog" and "Reveal JSON in file manager" work | TREE-08 | Wire-level in browser; the actual OS reveal needs the built app |
| Status bar shows live catalog count, file count, total bytes | SHELL-06 | Compare against the rail's own numbers |
| Empty-library state when the directory has no catalogs | STATE-01 | Point at an empty directory |
| Unreadable-catalog panel shows filename, byte offset, reason, raw error | STATE-02 | Select the deliberately corrupt catalog; assert all four are present |
| v1.x and v2.x catalogs open without conversion | COMPAT-01 | Open one of each format in the UI |

**Needs the built app, not the browser:** the actual OS-level file-manager reveal (TREE-08) — verify on the macOS build as Phase 22 did.

---

## Validation Sign-Off

- [ ] Wave 0 fixture generator exists and the 40k benchmark records a real number
- [ ] `go test ./...` green, including a `-race` pass for the sidecar cache
- [ ] `RevealInFileManager` argv test includes a hostile path (spaces, quotes, `;`)
- [ ] `cli` tests still green — `LoadCatalog` untouched
- [ ] `npx tsc --noEmit`, `npm run build`, `go build ./...` all green
- [ ] All 19 manual rows above completed in-browser, with the 40k row-count assertion recorded
- [ ] TREE-08's OS reveal verified on the built macOS app
- [ ] No watch-mode flags in any verify command
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
