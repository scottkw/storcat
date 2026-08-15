---
phase: 27-catalog-actions-watch
plan: 01
subsystem: catalog-actions
tags: [go, wails, json, html, rename, wailsapi, browsecatalogs]

# Dependency graph
requires:
  - phase: 26-settings
    provides: config.WatchDirectory scaffolding and the SettingsDialog shell pattern reused by later plans in this phase
provides:
  - "models.CatalogItem.Title (omitempty) as the authoritative on-disk catalog title field"
  - "internal/catalog.RenameCatalog -- order-preserving JSON root rewrite + dual HTML title rewrite"
  - "App.RenameCatalog -- containment-gated Wails binding (osutil.ContainsPath, matching GetCatalogHtmlPath/OpenExternal/RevealInFileManager)"
  - "wailsAPI.renameCatalog wrapper"
  - "BrowseCatalogs three-tier title resolution: JSON title -> unescaped HTML <title> -> filename"
affects: [27-02, 27-03, 27-04, 27-05, 27-06, 27-07]

# Actuals (#2632)
actuals:
  tokens: 8678
  tasks: 3
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Order-preserving JSON root rewrite via json.Decoder token-walk + json.RawMessage, never map[string]RawMessage or a struct round-trip -- preserves v1 array-envelope elements and any unrecognized key byte-for-byte"
    - "Surgical HTML substring replacement (strings.Index-delimited) for a single tag's content, never a full HTML regeneration"

key-files:
  created:
    - internal/catalog/rename.go
    - internal/catalog/rename_test.go
  modified:
    - pkg/models/catalog.go
    - internal/search/service.go
    - internal/search/service_test.go
    - app.go
    - frontend/src/services/wailsAPI.ts
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/wailsjs/go/models.ts

key-decisions:
  - "CatalogItem.Title tagged json:\"title,omitempty\", placed immediately after Name in the struct -- but RenameCatalog's on-disk rewrite always APPENDS a not-yet-present title key at the end of the root object's key list (per the plan's literal 'otherwise append the pair at the end' instruction), not at the struct's declared position -- struct field order only governs a fresh json.Marshal, not this byte-level edit"
  - "extractJSONTitle is a best-effort probe: unmarshal-into-struct-with-single-Title-field, tried bare-object then array-wrapped; any decode failure returns empty string, cascading to the HTML/filename fallback"

requirements-completed: [ACT-02]

coverage:
  - id: D1
    description: "A catalog's title can be changed via RenameCatalog without changing either filename, surviving a re-read through BrowseCatalogs"
    requirement: "ACT-02"
    verification:
      - kind: unit
        ref: "internal/catalog/rename_test.go#TestRenameCatalog_WritesJSONTitle"
        status: pass
      - kind: e2e
        ref: "live dev-browser against :34115 -- RenameCatalog then BrowseCatalogs round trip, see Task 3 evidence below"
        status: pass
    human_judgment: false
  - id: D2
    description: "Rename rewrites both the HTML <title> and <h1> sites with the same escaping the generator applies, and the read side unescapes them back"
    requirement: "ACT-02"
    verification:
      - kind: unit
        ref: "internal/catalog/rename_test.go#TestRenameCatalog_RewritesBothHTMLOccurrences"
        status: pass
      - kind: unit
        ref: "internal/catalog/rename_test.go#TestRenameCatalog_EscapesHTMLSpecials"
        status: pass
      - kind: unit
        ref: "internal/search/service_test.go#TestBrowseCatalogs_UnescapesHTMLTitle"
        status: pass
      - kind: e2e
        ref: "live dev-browser: renamed to 'Tom & Jerry <2024>', grep -c on .html returned 2, BrowseCatalogs returned the raw unescaped string"
        status: pass
    human_judgment: false
  - id: D3
    description: "A catalog with no .html renames cleanly; a v1 array-wrapped catalog preserves every other array element and every nested contents byte"
    requirement: "ACT-02"
    verification:
      - kind: unit
        ref: "internal/catalog/rename_test.go#TestRenameCatalog_NoHTMLIsNotAnError"
        status: pass
      - kind: unit
        ref: "internal/catalog/rename_test.go#TestRenameCatalog_PreservesArrayEnvelope"
        status: pass
      - kind: unit
        ref: "internal/catalog/rename_test.go#TestRenameCatalog_PreservesNestedContentsBytes"
        status: pass
    human_judgment: false
  - id: D4
    description: "App.RenameCatalog rejects a path outside the configured catalog directory before reading or writing anything, mirroring GetCatalogHtmlPath/OpenExternal/RevealInFileManager's containment gate"
    requirement: "ACT-02"
    verification:
      - kind: e2e
        ref: "live dev-browser: RenameCatalog on a real file outside the source dir rejected with 'outside configured catalog directory'; file unmodified after"
        status: pass
    human_judgment: false

duration: 40min
completed: 2026-08-15
status: complete
---

# Phase 27 Plan 01: Rename tracer slice Summary

**Catalog rename lands a `title` field authoritatively in the JSON root, patches both HTML title sites, and reads back through a corrected three-tier `BrowseCatalogs` title resolver -- proven live end-to-end against `wails dev` on :34115.**

## Performance

- **Duration:** ~40 min
- **Started:** 2026-08-15T17:24:54Z
- **Completed:** 2026-08-15T17:35:15Z
- **Tasks:** 3 (2 code tasks + 1 checkpoint gate, self-resolved with live evidence per standing checkpoint authority)
- **Files modified:** 10

## Accomplishments
- `models.CatalogItem.Title` (`omitempty`) is the new authoritative title field; absent for any catalog never renamed, preserving COMPAT-02 byte parity
- `internal/catalog.RenameCatalog` rewrites the JSON root order-preservingly (token-walk + `json.RawMessage`, never a struct round-trip or a `map[string]RawMessage`) and patches both `<title>` and `<h1>` in the sibling `.html` via the same `html.EscapeString` treatment the generator uses
- `App.RenameCatalog` gates every call through `filepath.Abs` + `filepath.EvalSymlinks` + `osutil.ContainsPath`, exactly like `GetCatalogHtmlPath`/`OpenExternal`/`RevealInFileManager` (T-27-02)
- Wails bridge regenerated (`App.d.ts`/`App.js`/`models.ts`) and `wailsAPI.renameCatalog` added, routed through the shared `wailsError` helper
- `BrowseCatalogs`'s read-side escaping bug is fixed: HTML-sourced titles are now `html.UnescapeString`'d; JSON-sourced titles are never unescaped, since they hold the user's literal string
- Full round trip proven live: `wails dev` restarted (binary was stale relative to the new binding), `Object.keys(window.go.main.App)` confirmed `RenameCatalog` present, then a real catalog was created, renamed twice, containment-rejected once, and every on-disk/read-back assertion from the plan's Task 3 checklist passed

## Task Commits

1. **Task 1: Rename a catalog end-to-end** - `f6b7912e` (feat) -- title field, `internal/catalog/rename.go`, `App.RenameCatalog`, regenerated bridge, `wailsAPI.renameCatalog`
2. **Task 2: Read side -- JSON title precedence + HTML unescape** - `bcae9b0c` (fix) -- `extractJSONTitle`, three-tier `BrowseCatalogs` title chain
3. **Task 3: Tracer verification gate** - no code commit (live-verification-only checkpoint); resolved via `checkpoint_authority` standing authority with real dev-browser evidence, documented below

**Plan metadata:** pending (this SUMMARY's own commit)

## Files Created/Modified
- `pkg/models/catalog.go` - `CatalogItem.Title string` (`json:"title,omitempty"`), placed after `Name`
- `internal/catalog/rename.go` - `RenameCatalog`, `setTitleInDocument`, `setRootStringField`, `rewriteHTMLTitle`, `replaceTagContent`
- `internal/catalog/rename_test.go` - 10 table-driven cases covering every `<behavior>` line
- `internal/search/service.go` - `extractJSONTitle`; `BrowseCatalogs` restructured to a single read of `filePath` feeding both the parse-error check and the title probe
- `internal/search/service_test.go` - `TestBrowseCatalogs_TitlePrecedence`, `_UnescapesHTMLTitle`, `_JSONTitleIsNotUnescaped`, `_ArrayWrappedTitle`, `_UnparseableJSONSkipsTitleProbe`
- `app.go` - `App.RenameCatalog(jsonPath, catalogDir, newTitle) error`
- `frontend/src/services/wailsAPI.ts` - `renameCatalog` wrapper
- `frontend/wailsjs/go/main/App.d.ts`, `App.js`, `frontend/wailsjs/go/models.ts` - regenerated via `wails generate module`

## Decisions Made
- Followed the plan's literal instruction for `setRootStringField`: an already-present key is replaced in place; a not-yet-present key is appended at the end of the root object's key list. For a catalog's first-ever rename this means the new `"title"` key lands at the END of the root object on disk (after `contents`), not "near the front" as the struct's declared field position (`Name`, then `Title`) might suggest -- struct field order only governs a fresh `json.Marshal`, and this is a surgical edit of existing bytes, not a re-marshal. Confirmed live: `"title":"..."` landed at byte offset 286 in a 330-byte document. This is the correct behavior per `<action>`'s explicit wording, and the only alternative (reordering to match struct position) would require exactly the map-based re-marshal the plan forbids for its key-order-preservation guarantee.
- `extractJSONTitle` uses a two-attempt unmarshal (bare-object then array-wrapped) rather than a manual byte scan, matching `LoadCatalog`'s own attempt order and reusing `encoding/json`'s tolerance for unrecognized keys.

## Deviations from Plan

### Acceptance-criterion mismatch (not a bug, documented for the record)

**1. Task 2's literal `grep -c 'os.ReadFile(filePath)' internal/search/service.go` returns 3, not 1**
- **Found during:** Task 2 verification
- **Cause:** `internal/search/service.go` already contained two pre-existing, unrelated occurrences of the exact substring `os.ReadFile(filePath)` in `searchInCatalogFile` and `LoadCatalog` before this task touched the file. The task's actual behavioral intent -- "the parse-error read and the title probe share one read" -- is fully satisfied: `BrowseCatalogs` itself contains exactly one `os.ReadFile(filePath)` call, and the title probe reuses those same bytes rather than reading again.
- **Action taken:** Implemented correctly (single read within `BrowseCatalogs`); did not modify `searchInCatalogFile` or `LoadCatalog` to force the whole-file grep to 1, since those functions are out of scope for this task (scope-boundary rule) and rewriting them would be unrelated churn.
- **Verification:** `go test ./internal/search/... -race -count=1` and the full `go test ./... -race -count=1` both green; manual read of `BrowseCatalogs`'s body confirms one read.

**2. Task 3's checkpoint wording "shows `"title":"..."` near the front" did not hold for a first-time rename**
- **Found during:** Task 3 live verification
- **Cause:** Same root cause as #1 above -- `setRootStringField` appends a new key at the end per the plan's own explicit instruction. The checkpoint script's informal "near the front" expectation implicitly assumed the struct's declared field position governs the on-disk byte position, which is only true for a fresh `json.Marshal`, not this rewrite.
- **Action taken:** Recorded the real byte offset (286 in a 330-byte document) as evidence rather than silently claiming compliance with the informal wording; the load-bearing assertions (title correct, both HTML sites correct, filenames unchanged, containment enforced, idempotent on the key) all passed exactly as specified.

---

**Total deviations:** 2, both acceptance-criterion/checkpoint-wording mismatches against pre-existing code or the plan's own explicit design choice -- no functional bug, no scope creep.
**Impact on plan:** None on correctness. Both are documented for `/gsd-verify-work` and future planners so a literal grep or "near the front" re-check isn't mistaken for a regression.

## Issues Encountered
- `wails dev` was running from a prior session and predated this plan's `RenameCatalog` binding. Per the standing operational constraint ("curl liveness is not binding freshness"), it was restarted before any live evidence was recorded, and `Object.keys(window.go.main.App)` was checked to confirm `RenameCatalog` was present post-restart.
- The configured catalog directory in the running app's persisted config pointed at a prior session's scratchpad path (`f5f2d8cb-...`) with no `.html` files present. Created a fresh scratchpad source directory and used the live `CreateCatalog` binding to produce a real `.json`/`.html` pair via the actual generator, rather than hand-crafting a fixture, so the live verification exercised the real write path end to end.

## Tracer Verification Gate -- Live Evidence (Task 3)

Performed directly (per `checkpoint_authority`'s standing authority to resolve tracer-gate checkpoints with real evidence, not to stop and ask), against `wails dev` on `:34115` after a fresh restart:

1. `Object.keys(window.go.main.App)` included `RenameCatalog` -- bridge confirmed fresh.
2. Created a real catalog (`my-test-catalog.json`/`.html`) via `CreateCatalog` in a scratch source directory, with `catalogDir` set to that same directory.
3. `RenameCatalog(jsonPath, catalogDir, "Tom & Jerry <2024>")` resolved without throwing.
4. `BrowseCatalogs(catalogDir)` returned `title: "Tom & Jerry <2024>"` -- exactly the raw form, never `&amp;`/`&lt;`/`&gt;`.
5. On disk: JSON held `"title":"Tom & Jerry <2024>"` (Go's default `&` JSON-string escaping of `&`/`<`/`>`, distinct from HTML entity escaping -- unmarshals back to the literal string); `grep -c 'Tom &amp; Jerry &lt;2024&gt;'` on the `.html` returned exactly `2`.
6. Filenames (`my-test-catalog.json`/`.html`) were unchanged before and after.
7. `RenameCatalog` on a real file that exists outside the configured catalog directory rejected with `"...outside configured catalog directory"`, and the file was verified byte-unmodified afterward.
8. A second rename to `"Second Title"` left exactly one `"title":` occurrence in the JSON root, carrying the newer value, and both HTML sites updated to match.

All 8 steps passed with real, recorded evidence -- no step was skipped or reasoned-only.

## Next Phase Readiness
- The four seams this tracer proves (`osutil.ContainsPath` in front of a path-taking binding, `WriteFileAtomic` as the one write primitive, the regenerated Wails bridge, and the `wailsAPI` wrapper shape) are all live and verified for plans `27-02` through `27-06` (duplicate, trash, menu, dialogs, watcher) to build on.
- `WriteFileAtomic` still has no `fsync` -- that hardening and its SIGKILL-mid-write verification are explicitly owned by plan `27-02`, per the threat register (T-27-06) and `.planning/WINDOWS.md` #6.

---
*Phase: 27-catalog-actions-watch*
*Completed: 2026-08-15*
