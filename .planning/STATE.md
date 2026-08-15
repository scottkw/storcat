---
gsd_state_version: 1.0
milestone: v3.0.0
milestone_name: Workspace Redesign
current_phase: 26
current_phase_name: Settings
status: executing
stopped_at: Completed 26-02-PLAN.md
last_updated: "2026-08-15T13:55:44.650Z"
last_activity: 2026-08-15
last_activity_desc: Phase 26 execution started
progress:
  total_phases: 7
  completed_phases: 4
  total_plans: 30
  completed_plans: 27
  percent: 57
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-13)

**Core value:** Fast, lightweight directory catalog management — Go/Wails delivers 93% smaller binaries and 5x faster search, with full feature parity and CLI scriptability.
**Current focus:** Phase 26 — Settings

## Current Position

Phase: 26 (Settings) — EXECUTING
Plan: 3 of 5
Status: Ready to execute
Last activity: 2026-08-15 — Phase 26 execution started

Phases 22 and 23 are COMPLETE and verified. Phases 24-28 remain.

Progress: [█████████░] 90%

## Performance Metrics

**Velocity:**

- Total plans completed: 12
- Average duration: —
- Total execution time: —

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 24 | 5 | - | - |
| 25 | 7 | - | - |

**Recent Trend:**

- Last 5 plans: —
- Trend: —

*Updated after each plan completion*
**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 22 P01 | 11min | 2 tasks | 20 files |
| Phase 22 P03 | 5min | 2 tasks | 1 files |
| Phase 22 P04 | 4min | 2 tasks | 2 files |
| Phase 22 P05 | 5min | 2 tasks | 2 files |
| Phase 22 P02 | 15min | 2 tasks | 7 files |
| Phase 22 P06 | 3min | 3 tasks | 5 files |
| Phase 22 P07 | 24min | 3 tasks | 5 files |
| Phase 23 P01 | 11min | 3 tasks | 18 files |
| Phase 23 P02 | 24min | 3 tasks | 9 files |
| Phase 23 P03 | 17min | 3 tasks | 5 files |
| Phase 23 P04 | 14min | 3 tasks | 5 files |
| Phase 23 P05 | 45min | 2 tasks | 4 files |
| Phase 23 P06 | 50min | 4 tasks | 8 files |
| Phase 24 P01 | 15min | 3 tasks | 13 files |
| Phase 24 P02 | 35min | 2 tasks | 0 files |
| Phase 24 P03 | ~12min | 2 tasks | 2 files |
| Phase 24 P04 | ~50min | 3 tasks | 4 files |
| Phase 24 P05 | 35min | 3 tasks | 4 files |
| Phase 25 P01 | 26min | 3 tasks | 21 files |
| Phase 25 P02 | 16min | 2 tasks | 9 files |
| Phase 25 P03 | 3min | 3 tasks | 6 files |
| Phase 25 P04 | 4min | 3 tasks | 15 files |
| Phase 25 P05 | 18min | 3 tasks | 8 files |
| Phase 25 P06 | 35min | 3 tasks | 6 files |
| Phase 25 P07 | 47min | 3 tasks | 14 files |
| Phase 26 P01 | 25min | 2 tasks | 14 files |
| Phase 26 P02 | 8min | 2 tasks | 11 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table. v3.0.0 milestone decisions (Ant Design removal, macOS TitleBarHiddenInset, GUI-only new capabilities, sidecar count cache, vendored IBM Plex fonts, THEMES array authoritative) are recorded in PROJECT.md's "Milestone decisions" section.

- [Phase ?]: Implemented full OKLab conversion (Ottosson matrices) for theme token color mixing, not sRGB averaging, for pixel-exact fidelity with the design handoff
- [Phase ?]: Toolbar: kept className="no-drag ws-search"/"no-drag ws-chip" per the plan's action text rather than restructuring to satisfy the plan's own literal className="no-drag" exact-match grep -- moving .no-drag to a wrapper would break .ws-search:hover/.ws-chip:hover
- [Phase ?]: 22-05: DetailsPanel takes no props parameter (not even unused/underscore-prefixed) -- the exported DetailsPanelProps interface exists purely for plan 22-07 to extend additively
- [Phase ?]: Task 1's blocking package-legitimacy checkpoint resolved via orchestrator's standing checkpoint_authority approval, not re-raised interactively
- [Phase ?]: IBM Plex @font-face src paths use no leading ./ to exactly match the existing Nunito rule's form
- [Phase ?]: 22-06: DevStateSwitcher reads density/railSide from useAppContext() instead of a shadow local copy, theme stays local pending Phase 26's Settings-owned theme state
- [Phase ?]: 22-06: Rail-side CSS swap scoped entirely inside the 1280px media block so the rail snaps back to left below that tier regardless of the stored setting
- [Phase ?]: 22-07: Fixed New pill hover-inversion CSS that was dead code -- an inline color/background on the button always outranked the .ws-new-pill:hover class rule, so the hover state never actually fired in any theme; moved base fill/text into a CSS rule so hover can win
- [Phase ?]: 22-07: Verified the phase's full manual matrix directly via dev-browser against the running Vite dev server rather than deferring to a literal human checkpoint pass -- all rows recorded with actual results
- [Phase ?]: 23-01: Excluded catalog root from LoadCatalogFlat's flat array (root's direct children get Depth 0/ParentIdx -1) so an empty catalog yields a zero-length slice and a single top-level file renders as one row
- [Phase ?]: 23-01: useVisibleRows requires a node's parent to be BOTH visible AND expanded, correcting 23-RESEARCH.md Pattern 3's visible-only sketch
- [Phase ?]: 23-01: @tanstack/react-virtual@3.14.9 installed without re-raising the package-legitimacy checkpoint, per 23-CONTEXT.md's standing pre-approval
- [Phase ?]: 23-02: Cache JSON serializes map[string]CountEntry directly (no wrapping struct) -- one less layer than the research sketch
- [Phase ?]: 23-02: storcatConfigDir() extracted from config.NewManager as a pure, behavior-unchanged refactor so the counts cache resolves the same directory
- [Phase ?]: 23-02: config.NewCountsCacheAt(path) exported (not test-private) so internal/search's tests can build a cache pointed at a temp path from outside the config package
- [Phase ?]: 23-03: TOGGLE_EXPAND gates on node.type==='directory' (not hasChildren) per the plan's literal wording -- toggling a childless directory is a harmless no-op
- [Phase ?]: 23-03: TreePane's empty-library branch also covers 'nothing selected yet' (catalogs exist but none picked) since the plan names no fifth state and this reuses the Phase 22 landing block
- [Phase ?]: 23-03: BreadcrumbBar mounts only inside TreePane's rows branch, not during loading/empty-catalog/empty-library, since its catalog-header slot doesn't exist until 23-05
- [Phase ?]: 23-03: Fixed a CSS Grid min-content bug -- .ws-tree needed its own min-width:0 (not just the breadcrumb path span's) so a deep selection's unbounded nowrap text truncates instead of widening the whole app
- [Phase ?]: [Phase 23-04]: Directory chip's hover gets its own .ws-dir-chip:hover rule (border-color only), not a reuse of .ws-chip:hover which also recolors text -- 23-UI-SPEC calls for hover-border-only
- [Phase ?]: [Phase 23-04]: SET_CATALOG_DIR clears currentCatalogId/tree/expanded/selected but deliberately leaves state.catalogs untouched -- the immediately-following SET_CATALOGS dispatch replaces the list
- [Phase ?]: [Phase 23-04]: themeTokens.ts' safeGetItem/safeSetItem exported (were module-private) for reuse by CatalogRail.tsx's directory-persistence, instead of a second try/catch wrapper
- [Phase ?]: Title stays base sans (not mono) in the catalog header, matching 23-04's rail-row precedent and the UI-SPEC's per-section contract over its general Typography summary
- [Phase ?]: Header renders in both ready branches (empty-catalog and rows) so a zero-file catalog still shows exact 0-value metadata
- [Phase ?]: isUnreadable check runs before the loading branch so a catalog already known broken from the rail listing never flashes a loading state first
- [Phase ?]: Fixed wailsAPI.ts's silent 'Unknown error' bug (Wails rejects with a plain string, not an Error instance) via extractErrorMessage(), applied to all 12 affected call sites
- [Phase ?]: 23-06: No directory-containment check added to RevealInFileManager -- the plan's own threat model (T-23-02) explicitly rejected it since the locked signature carries no directory param
- [Phase ?]: 23-06: macOS reveal verified via coordinator's direct AppleScript Finder-selection readback against the exact open -R argv this binding builds, including a hostile filename; Windows argv shape deferred (no Windows machine available), logged to WINDOWS.md
- [Phase ?]: 24-01: Measured wailsError(error) baseline directly (18) rather than trusting the plan's claimed number -- matched exactly
- [Phase ?]: 24-01: Palette this plan renders only basename/path/catalog/size -- shape icon, chip styling, truncation footer, keyboard nav, and match highlighting are owned by plans 24-03/24-04/24-05 per the plan's own artifact table
- [Phase ?]: 24-02: RESEARCH Open Question #1 resolved POSITIVE -- macOS WKWebView does not reserve Cmd-K; no options.App.Menu accelerator fallback needed, 24-CONTEXT.md's locked architecture stands
- [Phase ?]: 24-02: Stale wails dev binary (built before App.SearchIndexed landed) blocked live evidence at task start; curl -sf proves liveness not freshness -- restarted with user approval, verified via page.evaluate(() => Object.keys(window.go.main.App))
- [Phase ?]: useModalBehavior hook (frontend/src/hooks/useModalBehavior.ts) is the single implementation of focus trap/Escape/scroll-lock/focus-restore for Phases 25-27 to import unchanged
- [Phase ?]: 24-04: Highlight rendered as JSX text children only (never HTML string) to satisfy T-24-11 -- verified live with a fixture filename containing a literal <b> rendering as escaped text, zero real <b> elements
- [Phase ?]: 24-04: PLT-03 truncation line reads the Go-computed total prop, never results.length -- prevents ever presenting a capped set as complete
- [Phase ?]: 24-04: Added optional id prop to PaletteResultList (beyond its originating task's file list) so CommandPalette's combobox input can wire aria-controls to the listbox's real DOM id
- [Phase ?]: 24-05: Removed the word dispatch from reveal.ts's own comments (used issued/call site instead) after the Task 1 purity grep matched the word inside prose, not just code
- [Phase ?]: 24-05: mergeExpanded's idempotence check is a .some() scan for any ancestor path not already true, returning the input object unchanged by reference when nothing needs adding
- [Phase ?]: 24-05: pendingReveal cleared inside SELECT_CATALOG's and SET_CATALOG_DIR's existing atomic reducer updates, giving stale-discard for a rail switch superseding a reveal for free, with no separate staleness comparison at the consuming effect
- [Phase ?]: 25-01: Traverse-error classification (terminal vs. single-entry) deferred to plan 25-02; Task 1 only adds a ReadErrors counter on today's existing skip-and-continue paths
- [Phase ?]: 25-01: StartScan resolves outputDir/copyToDirectory via filepath.EvalSymlinks before containment-checking -- macOS's /var -> /private/var symlink otherwise produces a false escape rejection for every legitimately-nested write
- [Phase ?]: 25-01: CreateSlideOver's animated-exit closing flag uses a single useLayoutEffect keyed on isOpen, not the render-time setState pattern -- the latter breaks under React 18 StrictMode's development double-invoke, closing the panel with no visible exit animation
- [Phase ?]: 25-01: ScanResultFile.size is optional and left unset -- CreateCatalogResult has no per-output-file byte count, and using totalSize (the scanned tree's sum) would misrepresent an individual file's own size
- [Phase ?]: 25-02: Marker shape resolved to option-a (two flat omitempty scalars: Unreadable bool, ReadError string) at the plan's blocking checkpoint -- a USER decision, not Claude's discretion; option-b (nested object) and option-c (bare presence-implies-unreadable string) rejected
- [Phase ?]: 25-02: Errors propagate as *SourceUnavailableError with Partial populated once by the outermost CreateCatalogWithContext, not rebuilt at every recursion level; marker fields set only on the origin node where classify() first returns true
- [Phase ?]: 25-03: startScan (unexported core, optional progress test-hook) split from the Wails-bound StartScan so app_test.go can deterministically reproduce a mid-walk source loss headlessly -- a.throttledProgress's real path needs a live Wails runtime context that would log.Fatal on a fake one
- [Phase ?]: 25-03: CancelScan delegates to the new cancelActiveScan() helper (added in Task 2) instead of keeping Task 1's standalone body -- avoids two copies of the same cancel-and-report logic
- [Phase ?]: 25-03: CRT-13 force-quit-mid-scan live verification not performed -- wails dev was not running at task start and the executor was instructed not to start a long-lived dev server; recorded as an open manual item, not asserted as proven
- [Phase ?]: 25-04: internal/volumes uses stdlib syscall (not golang.org/x/sys) per this plan's phase-specific stdlib-only instruction, overriding 25-RESEARCH.md's x/sys recommendation -- go.mod/go.sum untouched
- [Phase ?]: 25-04: MeasureTree count-only pre-pass has no terminal-vs-single-entry classification -- tolerates every read failure via readErrors, matching traverseDirectory's ordinary skip-and-continue only
- [Phase ?]: 25-04: TestStartScan_RetainsPartialOnSourceLoss (25-03) updated to supply TotalBytesHint so its mid-walk-removal timing stays keyed to the real walk, not the new pre-pass
- [Phase ?]: 25-04: WINDOWS.md ledger entries landed as #4 (Windows disk-space/drive-letter) and #5 (Linux enumeration heuristic, first non-Windows entry), not the plan's anticipated #3/#4 -- entry #3 was already taken by 25-03's CRT-13 gap
- [Phase ?]: 25-05: Placeholder-only default application (native HTML input placeholder) satisfies CRT-04's field-independence contract with no separate touched-flag state
- [Phase ?]: 25-05: wailsAPI.startScan's opts gained totalBytesHint (deviation, file not in task's listed set) -- required to thread the selected volume's known total into the StartScan binding
- [Phase ?]: CreateSlideOver's scan:progress subscription changed from isOpen-gated to always-on (it is already permanently mounted), the minimal fix needed for CRT-08's background handoff to keep the status bar live while the panel is closed
- [Phase ?]: handleCloseRequest keys cancellation on an explicit CloseReason argument ('cancel-the-scan' vs 'leave-it-running'), never inferred from trigger identity or state alone, so Run in background can never accidentally cancel
- [Phase ?]: Fixed a real classification gap: the scan root's own initial os.Stat failure now correctly classifies as source-loss under HaltOnSourceLoss (was silently falling through as a generic error, misclassified by the frontend as a cancellation)
- [Phase ?]: Added real per-output-file byte sizes (CreateCatalogResult.JsonSize/HtmlSize/CopyJsonSize/CopyHtmlSize) so the done state's written-file rows show honest sizes instead of staying undefined
- [Phase ?]: All four create entry points reset a stale done/error scan state before opening, landing on the form step per 25-UI-SPEC's Entry Points contract, rather than reopening into whatever state.scan happened to hold
- [Phase ?]: 26-01: Get() returns a copy, not the live pointer -- required for concurrent Set* race-freedom
- [Phase ?]: 26-01: Proceeded past the tracer's interactive human-verify checkpoint after full live dev-browser proof, given config.json's mode:yolo and this project's established precedent (22-07, 24-02) of self-verifying live rather than deferring to a checkpoint
- [Phase ?]: 26-01: Restarted wails dev mid-plan so the wails.json version bump (compile-time go:embed) was actually live before recording evidence
- [Phase ?]: 26-02: RailSide is a genuinely new Config field distinct from the orphaned SidebarPosition, pinned by TestSetRailSide_DoesNotTouchSidebarPosition
- [Phase ?]: 26-02: ThemeGrid second swatch reads theme.tokens.p2 (not p) matching 26-UI-SPEC's literal code-mapping correction
- [Phase ?]: 26-02: Skipped full OS quit-and-relaunch for RailSide persistence proof to avoid disrupting shared wails dev process plans 26-03..05 depend on; substituted TestSetRailSide_Persists + on-disk config.json readback, logged WINDOWS.md entry #7

### Key Research Findings

Research complete — see `.planning/research/SUMMARY.md`, `ARCHITECTURE.md`, `PITFALLS.md` (researched 2026-08-13, confidence HIGH). Highlights carried into the roadmap:

- Flatten the catalog tree in Go (`LoadCatalogFlat`, new method) rather than modifying `LoadCatalog`, which `cli/show.go` depends on — folded into Phase 23 as backend prep, not a standalone phase.
- Theme tokens are computed in TypeScript, not CSS `color-mix()`, because Wails doesn't control WebKitGTK's version on Linux — Phase 22.
- Phase 25 (Create) is the milestone's single highest-regression-risk phase — the only one modifying existing tested code (`traverseDirectory`, `ProgressCallback`). Flagged for a dedicated research pass during planning.
- Phase 28 (Re-scan & Diff) is the handoff's own "biggest backend piece" — also flagged for a dedicated research pass, to resolve volume-relocation policy and the on-disk partial-catalog marker shape.
- COMPAT-01–06 are distributed as success criteria across Phases 23/25/26/28 rather than a final hardening phase — see ROADMAP.md's "Compatibility strategy" note.

### Pending Todos

**Carried security obligations for Phase 26** (from `23-SECURITY.md` — must not be re-accepted a third time):

- **T-22-05 — `CatalogModal` unsanitized `srcDoc`.** Accepted in Phases 22 and 23 *only* because no dispatcher of `openCatalogModal` exists. The plumbing beneath it is now live (`window.electronAPI` shim routes to real Go), so only the missing dispatcher stands between attacker-influenceable catalog HTML and an unsanitized `srcDoc`. Phase 26 must sanitize `htmlContent` or delete `CatalogModal.tsx` — not re-affirm the acceptance.
- **FU-23-A — `GetCatalogHtmlPath` / `OpenExternal` / `SearchIndexed` containment.** All three are Wails-exposed and callable from any renderer JS but lack the containment/symlink/regular-file gate that `RevealInFileManager` got in review finding WR-02. Mechanical fix: thread `catalogDir` through each and reuse `internal/osutil`'s existing `containsPath` helper. **`App.SearchIndexed` was added to this sweep's target list by Phase 24** (threat T-24-01, recorded in all five 24-*-PLAN.md threat registers and confirmed by `24-SECURITY.md`) — it takes an arbitrary directory with the same parameter surface as its siblings, so it must not be missed when this sweep runs.

**Open platform-gated items** (tracked in `.planning/WINDOWS.md`, sweep before v3.0.0 ships):

- SHELL-07 — native Windows drag-vs-click arbitration (Phase 22, user-accepted as platform-gated).
- TREE-08 — Windows `explorer /select,` argv shape at runtime (Phase 23). Unit-tested only; macOS reveal was verified against Finder's real selection.

### Blockers/Concerns

active. Two phases (25, 28) are pre-flagged as needing their own research pass before planning — not blockers, but plan-phase should account for the extra step.

- 25-01: An unintended side effect was caused on the user's live desktop during live verification -- a stray osascript/System Events keystroke (attempting to target the StorCat app window, which did not reliably come to accessibility focus) landed in a Raycast Settings/Store window instead and appears to have installed a 'GIF Search' Raycast extension. Not requested. User should check Raycast's installed extensions and remove it if unwanted. All further OS-level UI automation was stopped immediately once noticed.

- **Decision-coverage gate override (Phase 26, USER-approved 2026-08-15).** `check.decision-coverage-plan` returns `could-not-parse` on this project's CONTEXT.md format — discuss-phase here writes prose bullets under topic subheadings, not the `- **D-NN:** text` bullets the gate's parser requires, so it can name zero decisions rather than any uncovered one. It is a parser/format mismatch, not evidence of a dropped decision: the gsd-plan-checker's independent source audit confirmed every 26-CONTEXT.md decision (persistence, both carried security obligations, the `SearchIndexed` exclusion, dialog/theme/segmented decisions, the version bump, all discretion items) is carried into a plan. User elected to proceed and record the override. **Same override applies to phases 27 and 28 in this autonomous run.** verify-phase may re-surface this.

## Deferred Items

Items acknowledged and carried forward from previous milestone close:

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Windows signing | CRED-04/CRED-05: Windows OV code signing credentials not yet provisioned (SSL.com eSigner, 4 secrets missing) | Open | v2.3.0 close |
| Testing | TEST-01: Frontend unit tests (Vitest + Testing Library) for virtualizer, palette keyboard nav, modal behavior | v2 requirement, deferred | v3.0.0 requirements definition |
| Interface | FUT-01–03: Rail-as-drawer below ~820px, frameless Windows/Linux window chrome, CLI subcommands for new capabilities | v2 requirements, deferred | v3.0.0 requirements definition |

## Session Continuity

Last session: 2026-08-15T13:55:44.633Z
Stopped at: Completed 26-02-PLAN.md
Resume file: None
Resume command: `/gsd-autonomous --from 26`

**Worktree isolation: intentionally OFF.** `workflow.use_worktrees` is `false`, which is the correct and working configuration — Phases 22 and 23 were both built this way. Executors run sequentially on the main working tree. Do not "fix" this by setting it to `true`: Claude Code's harness forks agent worktrees from `origin/HEAD`, not local HEAD, and local `main` runs ~100 commits ahead of `origin/main` during a milestone, so every executor would land in a checkout predating all of v3.0.0 and halt on a base mismatch. The only reason to turn it on is speed (parallel executors within a wave), and it requires pushing `main` first.

## Phase 26 handoff (written 2026-08-15)

`26-CONTEXT.md` and `26-UI-SPEC.md` are committed and approved (UI checker: 6/6 PASS, no FLAGs — the first phase to pass all six cleanly). Research, planning and execution have **not** started.

**Phase 26 discharges two security obligations that have now been deferred through Phases 22, 23, 24 and 25.** The Pending Todos section above says they must not be re-accepted again; `26-CONTEXT.md` locks *how*:

- **T-22-05 — DELETE `frontend/src/components/CatalogModal.tsx` and its `App.tsx` wiring**, rather than sanitizing. Verified during discuss: the only `dispatchEvent` in the entire frontend is `themeChange`, so nothing dispatches `openCatalogModal` — the listener at `App.tsx:37` is registered but unreachable. The component still calls the legacy `window.electronAPI` shim and antd's `message` (antd was removed this milestone). `DetailsPanel.tsx:121-123` already provides the real "Open HTML catalog" path. Deleting removes the threat instead of mitigating it; sanitizing would mean hand-rolling HTML sanitization, since no new dependency is permitted.
- **FU-23-A — thread `catalogDir` through `GetCatalogHtmlPath` and restrict `OpenExternal` to `file://` paths contained in the catalog directory**, reusing `internal/osutil`'s `containsPath` exactly as `RevealInFileManager` did in Phase 23. **`SearchIndexed` comes OFF the sweep list** — Phase 25's security audit verified it takes no renderer-supplied path reaching a filesystem write and got its own containment at introduction. Correct the target list rather than carrying a redundant item.

**Also queued for Phase 26:**

- Mark `.planning/WINDOWS.md` **entry #3 fixed** — Phase 25 wave 7 closed CRT-13's force-quit-mid-scan live via `window.runtime.Quit()`. Entries #1, #2, #4, #5, #6 remain genuinely open for the pre-ship sweep.
- Migrate six `localStorage` keys into the Go config as the single source of truth: `storcat-theme-id`, `storcat-density`, `storcat-rail-side`, `storcat-catalog-directory`, `storcat-secondary-directory` (plus `storcat-dev-switcher`, which is dev-only and can stay). Phase 22 explicitly deferred this ("theme stays local pending Phase 26's Settings-owned theme state").

**Standing operational constraints for executors (learned the hard way this session):**

- **No host-OS GUI automation** — no `osascript`, System Events, `cliclick`, or synthetic keystrokes. An executor drove a native macOS folder dialog this way and delivered a keystroke to an unverified focused window. When CDP cannot reach a native dialog, call the binding directly in the live webview (`window.go.main.App.<Binding>(...)`) — the same code path a real click takes — or record it manual-only.
- **`curl` liveness is not binding freshness.** A `wails dev` server can answer `curl -sf` 200 while running a binary that predates a just-landed binding. Probe `Object.keys(window.go.main.App)` before recording any binding-dependent evidence.
- **Verify against `:34115` only.** Vite's `:5173` serves the same frontend but exposes no `window.go`, so binding assertions pass vacuously there.
- **`wails dev` is currently NOT running** — it exited when Phase 25 tested force-quit. Restart it before live verification.
- **No frontend test framework** (TEST-01 explicitly deferred). Frontend proof is `tsc --noEmit` + `vite build` + live dev-browser.
