---
phase: 23-rail-virtualized-tree
plan: 06
subsystem: workspace-details-panel
tags: [go, os-exec, security, react, wails]

requires:
  - phase: 23-01
    provides: "AppContext's tree/expanded/selected/currentCatalogId reducer state, the TreeState union"
  - phase: 23-02
    provides: "CatalogMetadata.fileCount/.totalBytes/.parseError -- the cache-backed fallback fields for a not-yet-loaded catalog"
  - phase: 23-03
    provides: "frontend/src/lib/format.ts (formatBytes/formatCount/formatDate), the phase's fixed formatter module"
  - phase: 23-05
    provides: "wailsAPI.ts's extractErrorMessage() -- every catch block reads the real Go error instead of 'Unknown error'"
provides:
  - "internal/osutil.RevealInFileManager -- argv-only, per-platform (darwin/windows/linux), validated before spawn"
  - "App.RevealInFileManager binding + wailsAPI.revealInFileManager wrapper"
  - "DetailsPanel's catalog-level view (Files/Catalogued/JSON/Modified) and node-level view (Type/Size/Catalog/Depth/Indexed)"
  - "DetailsPanel's two-action footer: Open HTML catalog, platform-aware Reveal-in-file-manager"
affects: []

actuals:
  tokens: 7000
  tasks: 4
  commits: 4

tech-stack:
  added: []
  patterns:
    - "Per-platform argv builders taken as pure functions of (platform, path) rather than build-tagged files -- all three shapes are unit-tested by the same test binary on one development machine, a deliberate deviation from 23-RESEARCH.md's build-tagged sketch"
    - "Deferred-call-with-catch Environment() read (Toolbar.tsx's pattern) duplicated locally in DetailsPanel's Footer rather than extracted into a shared hook -- two call sites don't clear the 'need three real examples' bar for abstraction"

key-files:
  created:
    - internal/osutil/reveal.go
    - internal/osutil/reveal_test.go
  modified:
    - app.go
    - frontend/wailsjs/go/main/App.d.ts
    - frontend/wailsjs/go/main/App.js
    - frontend/src/services/wailsAPI.ts
    - frontend/src/components/workspace/DetailsPanel.tsx
    - frontend/src/workspace.css

key-decisions:
  - "No directory-containment check added to RevealInFileManager beyond existence + regular-file + extension-allowlist validation -- the plan's own threat model (T-23-02) explicitly considered and rejected a containment check against the configured catalog directory, since the locked binding signature carries no directory parameter and the frontend never passes a free-form/user-typed path. The orchestrator's checkpoint-authority text also names containment as a condition; reconciled by following the plan's own explicit, reasoned disposition over the more generic checkpoint boilerplate, since adding a second parameter would be an unrequested architectural change (Rule 4) the plan itself already declined"
  - "Argument-vector builders (revealArgvDarwin/Windows/Linux) are pure, untagged Go functions selected by revealArgvFor(platform, path) rather than three go:build-tagged files -- lets the hostile-path test assert exact argv equality for all three platforms from this one macOS machine, per the plan's own explicit deviation from 23-RESEARCH.md Pattern 5's sketch"
  - "DetailsPanel's catalog-level Files/Catalogued meta rows prefer the loaded FlatCatalog's exact fileCount/totalBytes (state.tree) over the rail's cache-backed, possibly-null CatalogMetadata fields when the tree for that exact catalog has finished loading -- same precedent TreeHeader.tsx (23-05) already established; falls back to catalog.fileCount/totalBytes (nullable) rendered as an em dash when neither is available yet"
  - "Both Name fields (catalog title, node name) use word-break: break-all; both Path fields use single-line ellipsis -- reconciling the UI-SPEC's explicit node-level wording ('Name word-break: break-all') with its silence on the catalog-level Name/Path styling, and the plan's own must_haves line ('Names break on any character and paths...ellipsize on one line') read as applying uniformly to both views"
  - "macOS OS-level reveal verified by the coordinator via a direct AppleScript Finder-selection readback against the exact `open -R <path>` argv this binding builds (including the hostile-filename case: a space, an apostrophe and a semicolon), rather than clicking inside the wails-built .app -- the production build carries no CDP/devtools hook this agent could script. Both checks returned exactly one selected item, the file itself, not its containing folder"
  - "Windows argv shape stays deferred, not verified -- no Windows machine or VM exists in this environment. 23-RESEARCH.md Assumption A1 (the `/select,` + path single-argument join) remains unit-tested-for-structure only; recorded here as an open platform-gated item to sweep before v3.0.0 ships, matching the disposition already accepted for Phase 22's SHELL-07 (native Windows drag arbitration)"

patterns-established:
  - "A Go OS-integration surface that spawns a process validates (absolute, symlink-resolved, existing regular file, allowlisted extension) BEFORE building any argument vector, and builds that vector as pure per-target functions of (path) with no shell ever in the call chain -- exec.Command(name, args...), never exec.Command(\"sh\", \"-c\", ...)"

requirements-completed: [TREE-07, TREE-08]

coverage:
  - id: D1
    description: "RevealInFileManager: argv-only per-platform builders (darwin/windows/linux) selected at runtime (not build tags), validated before spawn (existing regular file, catalog/HTML extension only, symlinks resolved), with a hostile-path test proving a space/quotes/semicolon/&&/backtick/$()/pipe/newline arrives as exactly one unmodified argument vector element on all three platforms"
    requirement: "TREE-08"
    verification:
      - kind: unit
        ref: "internal/osutil/reveal_test.go#TestRevealArgv_Darwin,_Windows,_Linux, TestRevealArgv_HostilePath_Darwin,_Windows,_Linux, TestRevealInFileManager_RejectsMissingPath,_RejectsDirectory,_RejectsDisallowedExtension,_ResolvesRelativePath -- go test ./internal/osutil/... -race -count=1, all pass"
        status: pass
      - kind: manual_procedural
        ref: "Coordinator's direct AppleScript Finder-selection readback against `open -R` (the exact command this binding builds) for /tmp/storcat-2306-built-fixtures/normal-catalog.json and the hostile-named .../weird catalog's name;here.json against the wails-built .app -- both selected exactly one item, the file itself, not the containing folder"
        status: pass
    human_judgment: true
    rationale: "Real OS process-spawn behavior against Finder cannot be exercised by this repo's automated Go/frontend test suites (no CDP/devtools hook exists in the production Wails build) -- verified out-of-band by the coordinator rather than assumed from unit tests alone."
  - id: D2
    description: "Windows explorer /select,<path> single-argument-vector shape (23-RESEARCH.md Assumption A1)"
    requirement: "TREE-08"
    verification:
      - kind: unit
        ref: "internal/osutil/reveal_test.go#TestRevealArgv_Windows, TestRevealArgv_HostilePath_Windows -- structural shape only"
        status: pass
    human_judgment: true
    rationale: "No Windows machine or VM exists in this environment. The argv shape is unit-tested for structure only and remains runtime-unverified; deferred, must be swept before the v3.0.0 milestone ships (matches the disposition already accepted for Phase 22's SHELL-07)."
  - id: D3
    description: "Details panel tracks the selection: catalog-level view (title/path, Files/Catalogued/JSON/Modified, no HTML row) reachable even for a catalog whose JSON failed to parse, and node-level view (heading names file/folder, name/path, Type/Size/Catalog/Depth/Indexed) for a selected tree row"
    requirement: "TREE-07"
    verification:
      - kind: automated_ui
        ref: "dev-browser at :34115 against 3 fixture catalogs (with HTML companion, without, and one deliberately truncated): no-selection placeholder unchanged; catalog-level view rendered exactly 4 meta rows and no HTML row; node-level view rendered exactly 5 meta rows and a heading naming 'Selected folder'/'Selected file' correctly for both a directory and a file click; selecting the truncated catalog rendered the catalog-level view with Files/Catalogued as em dashes (unknown) while JSON size and Modified stayed populated from the rail listing, alongside (not instead of) the tree pane's own unreadable-catalog diagnostic -- screenshots t2306-initial.png, t2306-catalog-level.png, t2306-node-level.png, t2306-node-file.png, t2306-broken-details.png"
        status: pass
      - kind: other
        ref: "cd frontend && npx tsc --noEmit && npm run build; grep -c ws-meta-row/aria-label >=1, grep -c aria-haspopup|aria-expanded|useMediaQuery|matchMedia ==0"
        status: pass
    human_judgment: false
  - id: D4
    description: "Footer: exactly two buttons (primary Open HTML catalog, outlined platform-aware Reveal-in-file-manager) in both views, omitting the open button entirely for a catalog with no HTML companion, self-disabling each button while its own action is in flight, and surfacing a failure as a short inline message"
    requirement: "TREE-08"
    verification:
      - kind: automated_ui
        ref: "dev-browser at :34115: 2 buttons rendered for catalog-with-html, exactly 1 (reveal only) for catalog-no-html; reveal button read 'Reveal JSON in Finder' (darwin Environment() check); rapid double-click left the reveal button disabled immediately after the first click and re-enabled after settling (one call, not two); deleting the HTML companion out from under a queued Open click surfaced 'HTML file not found: ...' as an inline footer message -- screenshots t2306-footer-withhtml.png, t2306-footer-nohtml.png, t2306-footer-error.png"
        status: pass
      - kind: other
        ref: "grep -c revealInFileManager >=1, grep -c Re-scan ==0 in DetailsPanel.tsx; go build ./... && go test ./... -count=1 all green"
        status: pass
    human_judgment: false

duration: 50min
completed: 2026-08-13
status: complete
---

# Phase 23 Plan 06: Reveal-in-File-Manager and the Details Panel's Two Actions Summary

**Argv-only RevealInFileManager (three platform shapes unit-tested for exact hostile-path equality on one machine, macOS confirmed for real via AppleScript against a wails-built app), and a details panel that tracks the tree selection with a two-button footer that self-disables and surfaces its own errors**

## Performance

- **Duration:** 50 min
- **Completed:** 2026-08-13
- **Tasks:** 4 (3 auto + 1 checkpoint)
- **Files modified:** 8 (2 created, 6 modified)

## Accomplishments

- `internal/osutil/reveal.go`: `RevealInFileManager(path string) error` -- resolves the path to an absolute, symlink-resolved form, rejects anything that isn't an existing regular file with a `.json`/`.html` extension, and only then builds a command via `exec.Command(name, args...)`, never a shell. Three pure per-platform argv builders (`revealArgvDarwin`/`Windows`/`Linux`), dispatched at runtime by `revealArgvFor(platform, path)` rather than `go:build` tags -- a deliberate deviation from `23-RESEARCH.md`'s sketch that lets all three platform shapes be unit-tested by one test binary on this macOS machine
- `internal/osutil/reveal_test.go`: table tests per platform, a hostile-path test (space, single quote, double quote, semicolon, `&&`, backtick, `$()`, pipe, newline) asserting exact argument-vector length and content -- not a substring match -- for all three platforms, plus rejection tests (missing path, directory, disallowed extension) and a relative-path-resolution test that proves resolution happens without ever spawning a real process
- `app.go`: `App.RevealInFileManager` binding, delegating entirely to `internal/osutil`; bridge regenerated via `wails generate module`
- `frontend/src/services/wailsAPI.ts`: `revealInFileManager` wrapper following the existing `{success, ...}` envelope
- `DetailsPanel.tsx`: catalog-level view (title/path + Files/Catalogued/JSON/Modified, no HTML row, reachable even for a catalog whose JSON failed to parse) and node-level view (heading names file/folder + name/path + Type/Size/Catalog/Depth/Indexed), plus an inert `⋯` overflow button (accessible name now, no menu semantics until Phase 27) and a two-action footer (Open HTML catalog / platform-aware Reveal-in-file-manager) that self-disables per-button while its own action is in flight and surfaces a failure inline
- macOS OS-level reveal verified for real: the coordinator ran the exact `open -R <path>` command this binding builds against a wails-built `.app`'s fixture directory, including a file named `weird catalog's name;here.json`, and read Finder's actual selection back via AppleScript -- both cases selected exactly one item, the file itself. Windows remains deferred (no Windows machine available)

## Task Commits

Each task was committed atomically:

1. **Task 1: RevealInFileManager -- three platform shapes, one argument vector, zero interpreters** - `20e1a5aa` (feat)
2. **Task 2: The details panel follows the selection** - `6bc74a0e` (feat)
3. **Task 3: The two footer actions** - `e9f2ef62` (feat)
4. **Task 4: Confirm the real OS reveal on a built app, and state the Windows result** - checkpoint, approved by the coordinator (see Decisions) -- no code commit, evidence recorded in this SUMMARY

## Files Created/Modified

- `internal/osutil/reveal.go` - `RevealInFileManager`, three pure argv builders, `revealArgvFor` dispatcher
- `internal/osutil/reveal_test.go` - platform/hostile-path/rejection/relative-path-resolution tests
- `app.go` - `RevealInFileManager` binding
- `frontend/wailsjs/go/main/App.d.ts`, `App.js` - regenerated via `wails generate module`, not hand-edited
- `frontend/src/services/wailsAPI.ts` - `revealInFileManager` wrapper
- `frontend/src/components/workspace/DetailsPanel.tsx` - catalog-level/node-level views, inert overflow button, two-action footer
- `frontend/src/workspace.css` - `.ws-details-overflow`/`.ws-details-reveal` hover rules

## Decisions Made

- No directory-containment check added beyond existence/regular-file/extension-allowlist validation -- the plan's own threat model (T-23-02) explicitly rejected a containment check against the configured catalog directory (the locked binding signature carries no directory parameter, and the frontend never passes a free-form path); reconciled the orchestrator's more generic checkpoint-authority condition in favor of the plan's own explicit, reasoned disposition rather than inventing a new parameter (Rule 4 territory the plan itself already declined)
- Argument-vector builders are pure, untagged functions selected at runtime, not three `go:build`-tagged files, so the hostile-path test proves exact argv equality for all three platforms from one development machine
- Catalog-level `Files`/`Catalogued` meta rows prefer the loaded `FlatCatalog`'s exact counts over the rail's cache-backed, possibly-null fields once that catalog's tree has actually loaded -- same precedent `TreeHeader.tsx` (23-05) established; falls back to an em dash when neither source has a value yet
- Both Name fields (catalog title, node name) use `word-break: break-all`; both Path fields use single-line ellipsis -- reconciles the UI-SPEC's explicit node-level wording with its silence on catalog-level styling, reading the plan's must_haves line ("Names break on any character and paths...ellipsize on one line") as applying to both views uniformly
- The Toolbar's deferred-call-with-catch `Environment()` pattern is duplicated locally in the footer rather than extracted into a shared hook -- two call sites don't clear the "need three real examples" bar for abstraction
- macOS reveal verified via a direct AppleScript Finder-selection readback against the same `open -R` argv this binding builds (not by scripting inside the built `.app`, which has no CDP/devtools hook available to this agent); Windows explicitly deferred, not verified, no machine available

## Deviations from Plan

None beyond what the plan itself already flagged and pre-authorized (FPA-23-06-B, the untagged-file structure; the checkpoint-authority's package/architecture conditions, addressed in Decisions above). Plan executed as written; the checkpoint's verification method (AppleScript readback rather than clicking inside the built app) was the coordinator's own choice, consistent with the plan's acknowledgment that the built-app click itself "cannot be confirmed from a browser."

## Issues Encountered

None. `wails dev` was running throughout for Task 2/3 verification; `wails build -platform darwin/arm64` succeeded cleanly for Task 4 (35.3s). A synthetic fixture directory (`/tmp/storcat-2306-fixtures`, deleted after Tasks 2/3) and a second one for the built-app checkpoint (`/tmp/storcat-2306-built-fixtures`, left in place per the coordinator's request for the verifier) were used for live verification, neither committed to the repo.

## Known Stubs

None. Both backend and frontend surfaces for TREE-07/TREE-08 are fully wired end-to-end and verified live, including the macOS OS-level reveal. The one open item is the Windows argv shape (`explorer /select,<path>`), which is unit-tested for structure but not runtime-verified -- recorded above as D2 (human_judgment: true) rather than silently marked done, and logged to `.planning/WINDOWS.md` below.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- Phase 23 (Rail + Virtualized Tree) is now fully implemented across all 6 plans: fixture/tracer/virtualization (23-01), sidecar counts cache + parse status (23-02), interactive tree/breadcrumb/formatters (23-03), rail population/filter/status bar (23-04), catalog header/unreadable-catalog panel (23-05), and this plan's reveal binding + details panel actions (23-06).
- The Windows `explorer /select,` argv shape should be swept (built and manually verified on a real Windows machine or VM) before the v3.0.0 milestone ships -- it is currently unit-tested-for-structure only.
- No blockers for Phase 24 (⌘K palette).

## Self-Check: PASSED

All 8 files claimed as created/modified confirmed present on disk; all 3 task commit hashes (`20e1a5aa`, `6bc74a0e`, `e9f2ef62`) confirmed in `git log`.

---
*Phase: 23-rail-virtualized-tree*
*Completed: 2026-08-13*
