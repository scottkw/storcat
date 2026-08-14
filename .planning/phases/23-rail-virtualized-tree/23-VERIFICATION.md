---
phase: 23-rail-virtualized-tree
verified: 2026-08-14T05:20:00Z
status: passed
score: 17/17 must-have requirement groups verified (TREE-06 gap closed after the report — see gaps_closed)
behavior_unverified: 0
overrides_applied: 0
gaps_closed:
  - truth: "Switching catalogs resets the tree scroll position to the top; no stale offset survives the switch (TREE-06)."
    status: fixed
    fixed_in: "cac4aa06 — fix(23): reset tree scroll on the ready transition, not the load dispatch (TREE-06)"
    fix: >
      The verifier's root-cause diagnosis was exactly right. The scroll reset ran
      alongside the load dispatch while tree.status was still 'loading'; that branch
      renders no ref={scrollRef}, so scrollRef.current was null and the virtualizer
      had no scroll element — both reset calls were silent no-ops, and the virtualizer
      reapplied its own stale offset when the real element mounted. Moved the reset
      into a useLayoutEffect keyed on BOTH the catalog id and the 'loading'->'ready'
      transition.
    reverified: >
      Same sequence re-run against wails dev at localhost:34115 with a freshly
      generated 42,550-node fixture plus a second catalog: scroll to 1043, switch to
      the second catalog (scrollTop 0), switch BACK to the first (scrollTop 0).
      The step that previously returned 1043 now returns 0. tsc --noEmit and
      npm run build both exit 0.
    original_status: failed
    reason: >
      Reproduced live against wails dev (localhost:34115) with dev-browser: scroll
      resets to 0 correctly the FIRST time a given catalog is switched away from
      (a never-before-scrolled catalog opens at scrollTop 0), but when the user
      RETURNS to a catalog that was previously scrolled, the tree pane re-opens at
      the catalog's last scroll offset instead of 0. Sequence that reproduces it
      100% of the time: open fixture-dcim (42,550-node catalog) -> scroll to
      offset 2000 (clamps to 1043, the max) -> switch to a different, never-opened
      catalog (scrollTop correctly reads 0) -> switch back to fixture-dcim ->
      scrollTop reads 1043, not 0. Expansion and selection DO reset correctly on
      every switch (including the return case) -- only scroll leaks. Root cause:
      TreePane's reset effect (`if (scrollRef.current) scrollRef.current.scrollTop
      = 0; virtualizer.scrollToOffset(0);`) fires while `state.currentCatalogId`
      changes, but at that instant `tree.status` is still `'loading'` and the
      loading-state JSX branch does not attach `ref={scrollRef}` -- so both reset
      calls are no-ops against a null scroll element. Once the load resolves and
      the "ready" branch (which does attach the ref) mounts a fresh scroll
      element, `@tanstack/react-virtual`'s internal last-known offset (never
      cleared by the earlier no-op `scrollToOffset(0)`) gets re-applied to the new
      element instead of 0.
    artifacts:
      - path: "frontend/src/components/workspace/TreePane.tsx"
        issue: "Scroll-reset effect (lines ~38-59) resets scrollRef.current and calls virtualizer.scrollToOffset(0) while the scroll element is still unmounted (tree.status === 'loading'), so the reset silently fails for catalogs that were previously scrolled and are later revisited."
    missing:
      - "Defer the scroll reset until after the 'ready' branch's scroll element is actually mounted (e.g. reset in a layout effect keyed on tree.status transitioning to 'ready', or call virtualizer.scrollToOffset(0) again once nodes arrive), so the reset survives a revisit rather than only working the first time a catalog is opened."
      - "Add an explicit revisit case to the manual verification matrix (23-VALIDATION.md's TREE-06 row only exercises a single switch-away; it does not exercise switching back to the same, previously-scrolled catalog, which is exactly the path that fails)."
human_verification:
  - test: "TREE-08 Windows reveal: run the built app on a real Windows machine and confirm 'Reveal JSON in file manager' opens Explorer with the file pre-selected."
    expected: "Explorer opens with the .json file highlighted, using the explorer /select,<path> argv shape."
    why_human: "No Windows machine/VM available in this environment; already logged as an open item in .planning/WINDOWS.md (id 1) per the phase's own disclosure. Unit-tested for argv shape only (internal/osutil/reveal_test.go), not runtime-verified."
  - test: "RAIL-05: click the directory chip and confirm the native OS folder-picker dialog opens and a chosen directory both persists and reloads the rail."
    expected: "Native folder dialog appears; after choosing a directory, the rail reloads against it and the choice survives a relaunch."
    why_human: "The native OS dialog itself cannot be driven from dev-browser. The persistence-and-reload half was verified programmatically in this session (setting storcat-catalog-directory via localStorage + reload correctly reloaded the rail against the new directory), leaving only the dialog's own appearance for a human/manual pass, as the plan's own flagged assumption (FPA-23-04-B) already anticipated."
---

# Phase 23: Rail + Virtualized Tree Verification Report

**Phase Goal:** Users browse any catalog — including 40,000+-node ones — from a fast, filterable rail into a virtualized tree with a details panel that follows their selection
**Verified:** 2026-08-14
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Selecting a catalog in the rail loads its real contents as tree rows via one `LoadCatalogFlat` call, no per-expand round trip (TREE-01, RAIL-03) | VERIFIED | `TreePane.tsx`'s effect calls `wailsAPI.loadCatalogFlat` once per `currentCatalogId` change; `useVisibleRows` derives visibility client-side with no further Go calls. Confirmed live: selecting `fixture-dcim` (42,550 nodes) issued exactly one load, expand/collapse produced no network/binding activity. |
| 2 | v1 (array-wrapped) and v2 (bare-object) catalogs flatten identically (COMPAT-01) | VERIFIED | `TestLoadCatalogFlat_DualFormat` passes (`go test ./internal/search/...`). Live: a hand-written v1 array-wrapped fixture (`v1-legacy-volume.json`) opened and rendered correctly (1 file, 512B, correct header/metadata) with no conversion step. `LoadCatalog`/`CatalogItem` are byte-for-byte unmodified since before Phase 23 (`git log` shows no commits touching `internal/catalog/service.go` or `CatalogItem` in this phase); `cli` test suite (24 tests) stays green. |
| 3 | The 40,000+-node catalog scrolls smoothly with a bounded, viewport-proportional row count (TREE-01) | VERIFIED | Live against the 42,550-node generated fixture: mounted `.ws-tree-row` count stayed 30–41 across top/middle/bottom/9 intermediate scroll offsets after full expand-all (never proportional to node count). `BenchmarkLoadCatalogFlat40k` (run live in this session): 46.3ms Go-side parse+flatten, 5.641 MB marshalled, 42,550 nodes — consistent with the plan's own recorded 107.7ms/932.9ms-to-first-row figures. |
| 4 | Directory click toggles expansion AND selects; file click selects only (TREE-02) | VERIFIED | Live: clicking `VOL01` flipped its caret ▸→▾, revealed children, and populated the details panel with `VOL01`'s metadata in one click. Clicking a file (`IMG_0001.JPG`) selected it (details panel switched to file view) without altering any expansion state. |
| 5 | Expand-all / collapse-to-root work in one atomic state update and stay responsive at 40k+ (TREE-03) | VERIFIED | Live: "Expand all" against the 42,550-node fixture completed in ~800ms wall-clock (test-harness-inclusive), produced the full DFS pre-order visible set (scrollHeight 1,446,700px ÷ 34px/row ≈ 42,550 rows), and mounted row count stayed bounded (40) throughout a 5-point scroll sweep. "Collapse" returned to the 30-row top-level view. |
| 6 | Catalog header shows title, `.json`/`.html` chips, and the metadata line (TREE-04) | VERIFIED | Live: `fixture-dcim` header rendered "40,000 files \| 3.5M \| 131.1G catalogued \| modified 8/13/2026"; a catalog with no HTML companion correctly showed no HTML chip (matches the plan's documented rule). |
| 7 | Breadcrumb ancestors are accent-colored per-segment spans, current segment is not (TREE-05) | VERIFIED | Live DOM inspection: selecting `VOL01/100CANON/IMG_0001.JPG` produced 4 separate `span.ws-crumb` elements for `fixture-dcim`/`VOL01`/`100CANON` at `rgb(13,143,156)` (accent) and one `span.ws-crumb-current` for `IMG_0001.JPG` at `rgb(33,37,41)` (ordinary text) — confirmed NOT one pre-joined string. |
| 8 | **Switching catalogs resets scroll position, expansion, and selection — no stale state leaks (TREE-06)** | **FAILED** | Expansion and selection reset correctly on every switch, including revisits (confirmed live). **Scroll position does not**: see Gaps below — a previously-scrolled catalog re-opens at its old offset instead of 0, reproduced deterministically 2/2 attempts. |
| 9 | Details panel follows the current selection: catalog-level view with nothing selected, node-level view once a node is picked (TREE-07) | VERIFIED | Live: catalog selected/no node → "CATALOG" header with title/path/Files/Catalogued/JSON/Modified; node selected → "SELECTED FILE"/"SELECTED FOLDER" header with Type/Size/Catalog/Depth/Indexed, tracked correctly across 3+ selection changes in this session. |
| 10 | "Open HTML catalog" and "Reveal JSON in file manager" work from the details panel footer (TREE-08) | VERIFIED (macOS) / needs human (Windows) | Live: clicking "Reveal JSON in Finder" for `alpha-volume` completed with no error surfaced. Go tests confirm argv-only `exec.Command` construction (no shell), hostile-path tests (spaces/quotes/`;`/backticks/`$()`/pipe/newline) pass, and `containsPath` rejects out-of-directory and symlink-escape paths. Windows argv shape is unit-tested for structure only — logged as an open item in `.planning/WINDOWS.md` (id 1), consistent with this environment having no Windows machine. |
| 11 | Every rail row carries title, JSON size, filename, file count from one `BrowseCatalogs` call (RAIL-01) | VERIFIED | Live: 5-catalog fixture directory rendered correctly with per-row title/size/filename/count; counts absent (not zero) for catalogs whose sidecar cache is cold, matching the documented nullable-count contract. |
| 12 | Filter narrows the rail case-insensitively on title+filename without re-rendering the tree (RAIL-02) | VERIFIED | Live: typing "Beta" into the filter correctly narrowed the rail to `beta-volume` only, while the tree pane's rendered `.ws-tree-row` identities (`VOL01`/`VOL02`/`VOL03`) were byte-identical before and after — confirming the isolated-local-state mechanism holds. |
| 13 | Selecting a rail row loads its tree and clears the previous selection (RAIL-03) | VERIFIED | Live: selecting a new catalog always cleared the prior node selection and expansion (see truth 8 — this half of the atomic reset works correctly). |
| 14 | Red status dot on a catalog that failed to parse; row stays clickable (RAIL-04) | VERIFIED | Live: `corrupt-volume` (deliberately malformed with a trailing comma before `]`) showed a red dot, remained clickable, and opened the unreadable-catalog panel on selection. |
| 15 | Directory chip shows current directory, opens native picker, persists choice (RAIL-05) | VERIFIED (persistence/reload) / needs human (native dialog) | Live: setting the persisted directory key and reloading correctly re-populated the rail against the new directory. The native OS dialog itself is not driver-testable from a browser session — routed to human verification, as the plan itself anticipated (FPA-23-04-B). |
| 16 | Status bar reports live catalog/file/byte counts summed from the rail array, honest about cold-cache partiality (SHELL-06) | VERIFIED | Live: status bar read "5 catalogs · ≥40,005 files indexed · ≥131.1 GB" with the `≥` qualifier correctly present while `corrupt-volume`'s count was absent (parse failure) — matches the WR-03 fix's documented behavior exactly. |
| 17 | Empty-library and unreadable-catalog states show real diagnostic detail (STATE-01, STATE-02) | VERIFIED | Live: pointing at an empty directory produced the "Nothing catalogued yet" / "No catalogs here yet" pair in both rail and tree pane. Selecting `corrupt-volume` produced the STATE-02 panel with filename (`corrupt-volume.json`), byte offset (`byte 173`), reason (`invalid character ']' looking for beginning of value`), format attempted (`v2 object / v1 array`), and the raw parser error verbatim — all four required fields present. |

**Score:** 16/17 truths verified (1 failed: TREE-06's scroll-reset half)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/fixture/fixture.go` | 40k+ fixture generators | VERIFIED | `WriteDCIMCatalog` generated a live 42,550-node/3.6MB fixture in this session |
| `scripts/gen-fixture-catalog/main.go` | Committed CLI fixture writer | VERIFIED | Ran directly: `go run ./scripts/gen-fixture-catalog -out ... -shape dcim` produced the fixture used throughout this verification |
| `internal/search/flatten.go` | `LoadCatalogFlat` | VERIFIED | All flatten tests pass (`Structure`, `DualFormat`, `EmptyRoot`, `NotACatalog`, `DepthCap`, `MissingFile`, `PermissionDenied`, `OpportunisticCacheFill`) |
| `internal/search/flatten_bench_test.go` | `BenchmarkLoadCatalogFlat40k` | VERIFIED | Ran live: 46.3ms/op, 5.641MB, 42,550 nodes |
| `internal/config/counts_cache.go` | Concurrent-safe sidecar cache | VERIFIED | 8 tests pass including `TestCountsCache_ConcurrentPutGet_NoRace` under `-race` |
| `internal/osutil/reveal.go` | `RevealInFileManager` + argv builders | VERIFIED | Hostile-path, containment (`containsPath`), and per-platform argv tests all pass; live reveal call succeeded with no error |
| `pkg/models/catalog.go` | `FlatNode`/`FlatCatalog`, `CatalogMetadata` +3 fields | VERIFIED | `CatalogItem` untouched; `CatalogMetadata` gained exactly `FileCount`, `TotalBytes`, `ParseError` as specified |
| `frontend/src/hooks/useVisibleRows.ts` | O(n) visibility derivation | VERIFIED | Reviewed: single linear pass, `useMemo`-keyed on `[nodes, expanded]`, no rebuild/re-sort |
| `frontend/src/contexts/AppContext.tsx` | Reducer state + actions | VERIFIED | `SELECT_CATALOG`, `TREE_LOADED`/`TREE_FAILED` id-guard, `SET_CATALOG_DIR`, `TOGGLE_EXPAND`, `SET_EXPANDED` all present and correctly atomic — except the scroll-reset side effect, which lives in `TreePane.tsx`, not here (see Gaps) |
| `frontend/src/components/workspace/TreePane.tsx` | Virtualized rows + all pane states | PARTIAL | Rows/loading/empty/unreadable states all correct; scroll-reset effect has the TREE-06 bug documented in Gaps |
| `frontend/src/components/workspace/CatalogRail.tsx` | Rows, filter, chip, red dot | VERIFIED | All behaviors confirmed live |
| `frontend/src/components/workspace/StatusBar.tsx` | Live 3-segment summary | VERIFIED | `≥` qualifier confirmed live for cold-cache partial sums |
| `frontend/src/components/workspace/DetailsPanel.tsx` | Catalog/node views + footer | VERIFIED | Both views and both footer actions confirmed live |
| `frontend/src/components/workspace/TreeHeader.tsx`, `UnreadableCatalogPanel.tsx`, `BreadcrumbBar.tsx` | Header, error panel, breadcrumb | VERIFIED | All confirmed live with real data |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `LoadCatalog` (unmodified) | `LoadCatalogFlat` | Reuses dual-format parse | WIRED | `flatten.go` calls `catalog.LoadCatalog` directly; `cli` tests confirm no regression |
| `os.ReadDir` sort order | Rail row order | `BrowseCatalogs` return | WIRED | Confirmed no client-side re-sort; rows rendered in filesystem order |
| `json.Valid` fast path | `parseError` detection | `detectParseError` | WIRED | `TestCountsCache`/search tests confirm fast path; corrupt fixture correctly produced byte-offset diagnostics |
| `sync.Mutex` on counts cache | Concurrent background/opportunistic fill | `counts_cache.go` | WIRED | `-race` test passes |
| Filter string (local state) | Never enters `AppContext` | `CatalogRail.tsx` `useState`+`useDeferredValue` | WIRED | Confirmed live: tree rows unchanged while filter typed |
| `SELECT_CATALOG` | `TREE_LOADED` id guard | Superseded-load discard | WIRED | Reducer code inspected; guard present and correct |
| `currentCatalogId` change | Scroll-reset effect | `TreePane.tsx` `useEffect` | **PARTIALLY WIRED** | Effect fires correctly but targets a null `scrollRef.current` during the loading-state render, so the reset is a no-op for catalogs revisited after being scrolled — see Gaps |
| argv-slice `exec.Command` | No shell involvement | `reveal.go` | WIRED | Hostile-path tests confirm no metacharacter escape |
| `catalogDir` containment | `RevealInFileManager` | `containsPath` | WIRED | 4-scenario table test + 2 integration rejection tests pass |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|---------------------|--------|
| `TreePane` rows | `nodes`/`visibleIndices` | `LoadCatalogFlat` → `TREE_LOADED` → `useVisibleRows` | Yes — live 42,550-node fixture rendered real names/sizes/depths | FLOWING |
| `CatalogRail` rows | `state.catalogs` | `BrowseCatalogs` → `SET_CATALOGS` | Yes — 5 real catalogs listed with real sizes/counts | FLOWING |
| `StatusBar` totals | `filesIndexed`/`totalBytes` | `useMemo` over `state.catalogs` | Yes — live-summed, honestly qualified with `≥` when partial | FLOWING |
| `DetailsPanel` metadata | `catalog`/`selectedNode` | `AppContext` state, no separate binding | Yes — tracked live selection changes correctly | FLOWING |
| `UnreadableCatalogPanel` diagnostics | byte offset / reason / parser | `parseError` string from `BrowseCatalogs`/`LoadCatalogFlat` | Yes — real Go parser error surfaced verbatim | FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| 40k-node benchmark | `go test ./internal/search/... -bench=BenchmarkLoadCatalogFlat40k -run=^$ -benchtime=1x` | 46,261,125 ns/op, 5.641MB, 42,550 nodes | PASS |
| Reveal hostile-path argv construction | `go test ./internal/osutil/... -run "Hostile\|Contains\|RevealInFileManager" -v` | 12 tests, all PASS | PASS |
| Counts cache concurrency | `go test ./internal/config/... -run Cache -race -v` | 8 tests, all PASS | PASS |
| `LoadCatalogFlat` structure/format/edge cases | `go test ./internal/search/... -run Flat -v` | 8 tests, all PASS | PASS |
| CLI regression (`LoadCatalog` untouched) | `go test ./cli/... -v` | 24 tests, all PASS | PASS |
| Full workspace suite | `go build ./...`, `go test ./... -race`, `npx tsc --noEmit`, `npm run build` | All exit 0 | PASS |
| TREE-06 scroll reset (live, dev-browser at :34115) | Scroll fixture-dcim to offset, switch to unopened catalog, switch back | Scroll leaked back to 1043 instead of resetting to 0, reproduced 2/2 | **FAIL** |

### Requirements Coverage

All 17 requirement IDs named for this phase (SHELL-06, RAIL-01–06, TREE-01–08, STATE-01–02, COMPAT-01) are claimed across the six plans' `requirements:` frontmatter, with no orphans against `REQUIREMENTS.md`'s Phase 23 mapping. RAIL-06 is correctly marked `Pending` in `REQUIREMENTS.md` (deferred to Phase 25 by locked decision; verified the pill renders inert with no handler, matching the documented deferral). STATE-03 is correctly out of scope (Phase 28).

| Requirement | Source Plan(s) | Status | Evidence |
|-------------|----------------|--------|----------|
| SHELL-06 | 23-02, 23-04 | SATISFIED | Status bar live-verified |
| RAIL-01 | 23-02, 23-04 | SATISFIED | Rail rows live-verified |
| RAIL-02 | 23-04 | SATISFIED | Filter isolation live-verified |
| RAIL-03 | 23-01, 23-04 | SATISFIED | Selection-clears-previous live-verified |
| RAIL-04 | 23-02, 23-04 | SATISFIED | Red dot live-verified |
| RAIL-05 | 23-04 | SATISFIED (persistence) / NEEDS HUMAN (native dialog) | See human_verification |
| RAIL-06 | 23-04 | CORRECTLY DEFERRED | Pill inert, matches Phase 25 boundary |
| TREE-01 | 23-01 | SATISFIED | 40k benchmark + live scroll sweep |
| TREE-02 | 23-03 | SATISFIED | Click semantics live-verified |
| TREE-03 | 23-03 | SATISFIED | Expand-all/collapse live-verified |
| TREE-04 | 23-05 | SATISFIED | Header live-verified |
| TREE-05 | 23-03 | SATISFIED | Breadcrumb per-segment color live-verified |
| **TREE-06** | 23-01, 23-03 | **BLOCKED** | Scroll-reset half fails on revisit — see Gaps |
| TREE-07 | 23-06 | SATISFIED | Details panel tracking live-verified |
| TREE-08 | 23-06 | SATISFIED (macOS) / NEEDS HUMAN (Windows) | See human_verification |
| STATE-01 | 23-03, 23-04 | SATISFIED | Empty-library states live-verified |
| STATE-02 | 23-02, 23-05 | SATISFIED | Unreadable panel live-verified with all 4 diagnostic fields |
| COMPAT-01 | 23-01, 23-02 | SATISFIED | v1 fixture live-verified; `LoadCatalog`/`CatalogItem`/`cli` untouched |

### Anti-Patterns Found

No debt markers (`TBD`/`FIXME`/`XXX`), no unreferenced `TODO`/`HACK`, no placeholder returns, and no fabricated data found in any file modified by this phase. `23-REVIEW.md` (standard-depth, 28 files) surfaced 3 Warning + 2 Info findings, all 5 subsequently fixed and verified in `23-REVIEW-FIX.md` with real commits (`786b8ddf`, `d5f41f1c`, `23692e32`, `ada22533`, all confirmed to exist in this session). No Critical findings from that review, and no over-engineering flagged.

### Human Verification Required

1. **TREE-08 — Windows OS-level reveal.** Test: run the built app on Windows, select a catalog, click "Reveal JSON in file manager." Expected: Explorer opens with the file pre-selected via `explorer /select,<path>`. Why human: no Windows machine/VM in this environment; already logged in `.planning/WINDOWS.md` (id 1, open) exactly as the phase itself discloses. Unit-tested for argv shape only.
2. **RAIL-05 — native folder-picker dialog.** Test: click the directory chip, confirm the OS-native folder dialog opens and a chosen directory persists across relaunch. Expected: dialog appears; choice sticks. Why human: dev-browser cannot drive a native OS file dialog. The persistence/reload half was verified programmatically in this session.

### Gaps Summary

Sixteen of seventeen must-have truths verified cleanly against a live 42,550-node fixture, a hand-crafted corrupt catalog, a v1 array-wrapped catalog, and an empty directory — all exercised through `wails dev` at `:34115` with real Go bindings, not the unbound Vite dev server. The Go-side surface (flatten, counts cache, reveal, fixture generator) is thoroughly and correctly tested, including the `-race` concurrency case and the hostile-path/containment security fixes from the code review.

**One genuine, reproducible gap: TREE-06's scroll-reset guarantee is only half true.** Expansion and selection correctly reset to empty/null on every catalog switch, including when returning to a previously-visited catalog. Scroll position, however, only resets correctly the first time a catalog is switched away from; if a user scrolls a catalog, switches away, and switches back, the tree reopens at the stale offset instead of the top. This was reproduced deterministically twice in this session with careful `waitForReady` gating to rule out test-timing artifacts, and the root cause is identifiable in `TreePane.tsx`: the reset effect fires while the scroll-owning DOM element is unmounted (during the `'loading'` render branch), so both reset calls (`scrollRef.current.scrollTop = 0` and `virtualizer.scrollToOffset(0)`) are silent no-ops, and `@tanstack/react-virtual` reapplies its stale internal offset once the real scroll element remounts on `'ready'`.

Notably, `23-03-SUMMARY.md`'s own live-verification note for this exact must-have checked that "re-selecting the original catalog showed the expansion map cleared" but did not re-check scroll on that same return step — the asymmetry that let this ship. This is flagged as a gap rather than an override candidate: it's a small, well-localized fix (defer the reset until the ready-state scroll element exists), and TREE-06 is both an explicit roadmap Success Criterion and a named must-have in two separate plans (23-01, 23-03), not a peripheral nicety.

RAIL-06 (create slide-over) is correctly deferred to Phase 25 per locked decision — the "＋ New" pill renders in its Phase 22 form, is genuinely inert (no handler), and `REQUIREMENTS.md` already marks it `Pending` rather than `Complete`. This is not a gap for Phase 23.

---

_Verified: 2026-08-14_
_Verifier: Claude (gsd-verifier)_
