# Roadmap: StorCat

## Milestones

- ✅ **v2.0.0 Go/Wails Migration** — Phases 1-7 (shipped 2026-03-26) — [Archive](milestones/v2.0.0-ROADMAP.md)
- ✅ **v2.1.0 CLI Commands** — Phases 8-11 (shipped 2026-03-26) — [Archive](milestones/v2.1.0-ROADMAP.md)
- ✅ **v2.2.0 Repo Consolidation & CI/CD** — Phases 12-15 (shipped 2026-03-27) — [Archive](milestones/v2.2.0-ROADMAP.md)
- ✅ **v2.3.0 Code Signing & Package Manager CLI** — Phases 16-21 (shipped 2026-03-28) — [Archive](milestones/v2.3.0-ROADMAP.md)
- 🚧 **v3.0.0 Workspace Redesign** — Phases 22-28 (in progress)

## Phases

<details>
<summary>✅ v2.0.0 Go/Wails Migration (Phases 1-7) — SHIPPED 2026-03-26</summary>

- [x] Phase 1: Data Models + Catalog Service (2/2 plans) — completed 2026-03-24
- [x] Phase 2: Search Service + Browse Metadata (1/1 plans) — completed 2026-03-25
- [x] Phase 3: Config Manager (2/2 plans) — completed 2026-03-25
- [x] Phase 4: App Layer + Lifecycle (2/2 plans) — completed 2026-03-25
- [x] Phase 5: Frontend Shim (1/1 plans) — completed 2026-03-26
- [x] Phase 6: Platform Integration (1/1 plans) — completed 2026-03-26
- [x] Phase 7: Verification + Merge (2/2 plans) — completed 2026-03-26

</details>

<details>
<summary>✅ v2.1.0 CLI Commands (Phases 8-11) — SHIPPED 2026-03-26</summary>

- [x] Phase 8: CLI Foundation and Platform Compatibility (2/2 plans) — completed 2026-03-26
- [x] Phase 9: Core Subcommands — Create, List, Search (2/2 plans) — completed 2026-03-26
- [x] Phase 10: Show, Open, and Output Polish (2/2 plans) — completed 2026-03-26
- [x] Phase 11: Tech Debt Cleanup (1/1 plan) — completed 2026-03-26

</details>

<details>
<summary>✅ v2.2.0 Repo Consolidation & CI/CD (Phases 12-15) — SHIPPED 2026-03-27</summary>

- [x] Phase 12: Repo Consolidation (3/3 plans) — completed 2026-03-27
- [x] Phase 13: CI Scaffold and Multi-Platform Build (1/1 plans) — completed 2026-03-27
- [x] Phase 14: Platform Packaging (1/1 plans) — completed 2026-03-27
- [x] Phase 15: Distribution Channel Automation (2/2 plans) — completed 2026-03-27

</details>

<details>
<summary>✅ v2.3.0 Code Signing & Package Manager CLI (Phases 16-21) — SHIPPED 2026-03-28</summary>

- [x] Phase 16: Secrets & Certificate Procurement (3/3 plans) — completed 2026-03-28
- [x] Phase 17: macOS Signing & Notarization (1/1 plans) — completed 2026-03-28
- [x] Phase 18: Windows Authenticode Signing (1/1 plans) — completed 2026-03-28
- [x] Phase 19: Homebrew CLI PATH (1/1 plans) — completed 2026-03-28
- [x] Phase 20: Windows CLI PATH via NSIS (1/1 plans) — completed 2026-03-28
- [x] Phase 21: Auto version and auto distribution (1/1 plans) — completed 2026-03-28

</details>

### 🚧 v3.0.0 Workspace Redesign (In Progress)

**Milestone Goal:** Replace the three-tab Ant Design frontend with the 1a Workspace design from `design_handoff_storcat_ui/`, matching it exactly, and build every backend capability the design implies.

**Compatibility strategy:** COMPAT-01 through COMPAT-06 are cross-cutting regression guarantees, not new capability. They are distributed as success criteria across the phases whose changes could break them — Phase 23 (LoadCatalog/LoadCatalogFlat exercised against real v1.x/v2.x catalogs), Phase 25 (the only phase modifying the shared write path, the `ProgressCallback` signature, and the internal/catalog↔Wails boundary that the CLI depends on staying clean), Phase 26 (window-state persistence, rebuilt inside the new Settings modal), and Phase 28 (the closing full-CI-build/sign/release gate, run after the milestone's one new Go dependency and the frontend rewrite are both in place) — rather than collected into a standalone hardening phase at the end. A final-phase-only approach would let a regression introduced in Phase 23 or 25 go undetected until the last phase; distributing them means each guarantee is verified at the point where its supporting change actually lands, consistent with the milestone's autonomous-execution verification standard (Go tests + CLI re-run + dev-browser visual check per phase).

- [x] **Phase 22: Shell + Token Layer** - Single-view workspace shell (toolbar, 3-pane grid, status bar) with the full 11-theme token layer and responsive tiers (completed 2026-08-13)
- [x] **Phase 23: Rail + Virtualized Tree** - Catalog rail and a 40k-node virtualized tree pane with a details panel, backed by a new LoadCatalogFlat method and a sidecar count cache (completed 2026-08-13)
- [x] **Phase 24: Cmd-K Command Palette** - Cross-catalog ⌘K search with capped results and the shared modal-behavior hook (completed 2026-08-14)
- [ ] **Phase 25: Create Slide-over + Progress/Cancellation/Partial-Catalog** - Volume-aware create flow with live scan progress, real cancellation, and an error-tolerant partial-catalog path
- [ ] **Phase 26: Settings** - Theme, density, rail position, and catalog defaults in one settings modal
- [ ] **Phase 27: Catalog Actions + Watch** - Rename, duplicate, delete-to-Trash, and live directory watching
- [ ] **Phase 28: Re-scan & Diff** - Re-scan a catalog's source volume and resolve an added/removed/changed diff

## Phase Details

### Phase 22: Shell + Token Layer

**Goal**: Users interact with a single-view workspace (toolbar, rail, tree, details, status bar) instead of the old three-tab interface, fully responsive and themed across all 11 themes
**Depends on**: Nothing (first phase of v3.0.0)
**Requirements**: SHELL-01, SHELL-02, SHELL-03, SHELL-04, SHELL-05, SHELL-07, SHELL-08, SHELL-09, THEME-01, THEME-02, THEME-03, THEME-04, THEME-05, THEME-06
**Success Criteria** (what must be TRUE):

  1. User sees a single workspace view — no tabs — with a 46px toolbar, catalog rail, tree pane, details panel, and status bar, laid out as `268px 1fr 288px` at window widths ≥1280px (SHELL-01, SHELL-02)
  2. User sees the layout respond correctly at 1040–1279px (rail narrows to 236px, details becomes a toggleable drawer behind a "Details" chip) and below 1040px (rail 200px, details stays a drawer, tree keeps priority), with every overlay stacking correctly at each tier (SHELL-03, SHELL-04, SHELL-09)
  3. User can move the catalog rail to the right side with the 1px divider following it, drag the window from the toolbar without losing clicks on the search field, theme chip, or gear, and — on macOS — see the real traffic lights sitting inside the 46px toolbar, with the native title bar above it on Windows and Linux (SHELL-05, SHELL-07, SHELL-08)
  4. User can switch between all 11 themes and see the entire workspace repaint immediately, with legible text on every accent-filled button and badge across both light accents (Gruvbox orange, Monokai green) and dark accents (GitHub blue), using the handoff's extended token set (THEME-01, THEME-02, THEME-03)
  5. User can toggle row density between Compact and Comfortable and see it change tree row height, rail/details/palette row padding, and tree font size; sees IBM Plex Sans and Mono render with no network access; and finds their theme, density, and rail position preserved across an app restart (THEME-04, THEME-05, THEME-06)

**Plans**: 7/7 plans complete (4 waves)
Plans:
**Wave 1**

- [x] 22-01-PLAN.md — Tracer: theme token pipeline paints the workspace frame end to end, tab UI removed, macOS title-bar option wired (wave 1)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 22-02-PLAN.md — Typography: vendored latin-subset IBM Plex Sans/Mono with IBM's OFL text (wave 2)
- [x] 22-03-PLAN.md — Toolbar: controls, drag-region opt-outs, macOS traffic-light inset (wave 2)
- [x] 22-04-PLAN.md — Catalog rail and 26px status bar skeletons (wave 2)
- [x] 22-05-PLAN.md — Tree pane and details panel skeletons (wave 2)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 22-06-PLAN.md — Workspace state, responsive tiers, rail-side swap (wave 3)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 22-07-PLAN.md — Details drawer, overlay stacking scale, phase verification matrix (wave 4)

**UI hint**: yes

### Phase 23: Rail + Virtualized Tree

**Goal**: Users browse any catalog — including 40,000+-node ones — from a fast, filterable rail into a virtualized tree with a details panel that follows their selection
**Depends on**: Phase 22
**Requirements**: SHELL-06, RAIL-01, RAIL-02, RAIL-03, RAIL-04, RAIL-05, RAIL-06, TREE-01, TREE-02, TREE-03, TREE-04, TREE-05, TREE-06, TREE-07, TREE-08, STATE-01, STATE-02, COMPAT-01
**Success Criteria** (what must be TRUE):

  1. User sees every catalog in the configured catalog directory listed in the rail (title, JSON size, filename, file count), can filter it by typing (case-insensitive match on title and filename) without the tree re-rendering on each keystroke, and selecting a row loads its tree while clearing the previous selection (RAIL-01, RAIL-02, RAIL-03)
  2. User sees a red status dot on any catalog that failed to parse, sees the current catalog directory as a chip they can change, and can open the create slide-over from the "＋ New" pill (RAIL-04, RAIL-05, RAIL-06)
  3. User can browse a catalog of 40,000+ nodes with smooth scrolling and no freeze — verified against a synthetic 40,000+-node fixture catalog built for this phase — expand and collapse directories (directory click toggles and selects, file click selects only), and expand every directory or collapse to root from the breadcrumb bar (TREE-01, TREE-02, TREE-03)
  4. User sees the catalog header (title, `.json`/`.html` chips, file count/JSON size/bytes-catalogued/modified-date line), a breadcrumb path with accent-colored ancestor segments, and a details panel that follows the current selection with working "Open HTML catalog" and "Reveal JSON in file manager" actions (TREE-04, TREE-05, TREE-07, TREE-08)
  5. User's scroll position and expansion state reset cleanly on catalog switch (TREE-06); the status bar shows live catalog count, indexed file count, and total bytes (SHELL-06); catalogs created by StorCat v1.x and v2.x open without conversion (COMPAT-01); and empty-library / unreadable-catalog states appear with file/byte-offset/reason/raw-parse-error detail (STATE-01, STATE-02)

**Plans**: 6/6 plans complete

**Wave 1**

- [x] 23-01-PLAN.md — Tracer: fixture generator, LoadCatalogFlat, bridge, virtualized rows, 42k measurement (wave 1)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 23-02-PLAN.md — Sidecar counts cache, parse-error detection, three CatalogMetadata fields (wave 2)
- [x] 23-03-PLAN.md — Expand/collapse, expand-all and collapse-to-root, breadcrumb, formatters (wave 2)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 23-04-PLAN.md — Populated rail rows, filter, directory chip, live status bar (wave 3)
- [x] 23-05-PLAN.md — Catalog header and the unreadable-catalog diagnostic panel (wave 3)
- [x] 23-06-PLAN.md — RevealInFileManager and the details panel with its two actions (wave 3)

**UI hint**: yes

### Phase 24: Cmd-K Command Palette

**Goal**: Users find any file across every catalog instantly from a ⌘K palette without leaving the workspace
**Depends on**: Phase 23
**Requirements**: PLT-01, PLT-02, PLT-03, PLT-04, PLT-05, PLT-06, PLT-07
**Success Criteria** (what must be TRUE):

  1. User can open the search palette with ⌘K or by clicking the toolbar search field, with the input autofocused (PLT-01)
  2. User can search names and paths across every catalog in the directory, seeing at most 50 results with a "Showing the first 50 of N hits" notice when more matched, or "No file in any catalog matches that." when nothing matched (PLT-02, PLT-03, PLT-06)
  3. User can navigate results by keyboard, click a hit to switch to its catalog, expand every ancestor, select it, scroll it into view, and close the palette, and can dismiss the palette with Escape (PLT-04, PLT-05)
  4. Focus is trapped inside the palette while it's open and page scroll is locked behind it, via a shared modal-behavior hook (focus trap, Escape, scroll lock) that Phases 25-27's overlays reuse rather than reimplement (PLT-07)

**Plans**: 5/5 plans executed (5 waves)

Plans:
**Wave 1**

- [x] 24-01-PLAN.md — Tracer: capped cross-catalog search in Go, the regenerated Wails bridge, and both ⌘K open paths end to end (wave 1)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 24-02-PLAN.md — Tracer verification gate: ⌘K delivery inside the real WKWebView window, plus live proof of the cap and the stale guard at :34115 (wave 2)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 24-03-PLAN.md — The shared useModalBehavior hook (focus trap, Escape, scroll lock, focus restore) Phases 25-27 reuse (wave 3)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 24-04-PLAN.md — Result row and listbox, the four body states with verbatim copy, and keyboard navigation (wave 4)

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 24-05-PLAN.md — Reveal-to-tree: merge-not-replace ancestor expansion and the two-effect centred scroll (wave 5)

**UI hint**: yes

### Phase 25: Create Slide-over + Progress/Cancellation/Partial-Catalog

**Goal**: Users create a new catalog from a detected volume or folder, watch it scan live, and can cancel it or recover from a volume that disappears mid-scan — all without risking data loss or breaking the CLI
**Depends on**: Phase 23, Phase 24
**Requirements**: CRT-01, CRT-02, CRT-03, CRT-04, CRT-05, CRT-06, CRT-07, CRT-08, CRT-09, CRT-10, CRT-11, CRT-12, CRT-13, COMPAT-02, COMPAT-03, COMPAT-04
**Success Criteria** (what must be TRUE):

  1. User can open a 560px right slide-over that animates in over 340ms and out over 260ms without unmounting early, regardless of which of the five close paths (Escape, ×, Cancel, scrim, "Open in workspace") triggers the close (CRT-01)
  2. User sees detected mounted volumes as selectable cards (name, mount path, size, `mounted`/`read errors` status), can choose any folder instead, set a catalog title and filename root with a live "WILL WRITE" preview, toggle write-HTML/copy-secondary/include-hidden, and start the scan with the Create button or ⌘↵ (CRT-02, CRT-03, CRT-04, CRT-05, CRT-06)
  3. User sees live scan progress — percentage, files seen, bytes, estimated time remaining, current walking path, and a newest-first log — can hand a running scan to the status bar with "Run in background", and can cancel a scan such that the underlying walk actually stops (CRT-07, CRT-08, CRT-09)
  4. User sees a distinct error state when the volume goes away mid-scan (showing where it stopped and the read errors encountered), can write a partial catalog, retry the scan, or cancel from that state, and closing the window mid-scan cancels the walk and writes nothing (CRT-10, CRT-11, CRT-13)
  5. User sees a done state listing every file written with its size and can open the new catalog in the workspace or catalog another volume; catalogs this flow writes are byte-for-byte the same JSON shape v2.3.0 wrote; and all six CLI subcommands keep behaving exactly as before, with `internal/catalog` still usable from the CLI without a Wails runtime context (CRT-12, COMPAT-02, COMPAT-03, COMPAT-04)

**Plans**: 3/7 plans executed (7 waves)
Plans:

**Wave 1**

- [x] 25-01-PLAN.md — Tracer: ctx-threaded catalog service, byte-compatible CLI wrapper, StartScan binding, animated slide-over shell, end-to-end folder→catalog path (wave 1)

**Wave 2** *(blocked on Wave 1)*

- [x] 25-02-PLAN.md — Atomic writes, read-error classification, and the on-disk unreadable-subtree marker (wave 2)

**Wave 3** *(blocked on Wave 2)*

- [x] 25-03-PLAN.md — Cancel handle, retained partial scan, partial-write binding, cancel-on-window-close (wave 3)

**Wave 4** *(blocked on Wave 3)*

- [ ] 25-04-PLAN.md — internal/volumes per-OS enumeration, ListVolumes binding, count-only pre-pass, platform-ledger entries (wave 4)

**Wave 5** *(blocked on Wave 4)*

- [ ] 25-05-PLAN.md — Volume picker, folder alternative, title/filename-root fields, live WILL WRITE preview, the three toggles (wave 5)

**Wave 6** *(blocked on Wave 5)*

- [ ] 25-06-PLAN.md — Live scanning body in both sub-states, cancellation from every close path, background hand-off and status-bar segment (wave 6)

**Wave 7** *(blocked on Wave 6)*

- [ ] 25-07-PLAN.md — Error state with partial-write/retry/close, done state in both flavours, four entry points, phase verification matrix (wave 7)

**UI hint**: yes
**Research flag**: RESOLVED — the dedicated research pass ran (`25-RESEARCH.md`, confidence HIGH). Both flagged unknowns are answered: forced close and user cancel write nothing because the tree is complete in memory before any write is attempted, and the unreadable-subtree marker is two `omitempty` fields on `CatalogItem` (exact shape gated by a `checkpoint:decision` in 25-02, since it is a one-way on-disk format door). Research also surfaced a third, unflagged finding: `CreateCatalog` conflates the scanned source with the output directory, so the new context-aware method needs a genuinely separate `outputDir` parameter with the CLI wrapper defaulting it to today's behavior.

### Phase 26: Settings

**Goal**: Users configure theme, density, rail position, and catalog defaults from one settings surface that saves as they go
**Depends on**: Phase 22
**Requirements**: SET-01, SET-02, SET-03, SET-04, SET-05, COMPAT-05
**Success Criteria** (what must be TRUE):

  1. User can open Settings with ⌘,, the gear, or the theme chip (SET-01)
  2. User can pick a theme from 11 cards (each showing a 4-swatch strip and a light/dark tag), and set row density and catalog rail position from segmented controls (SET-02, SET-03)
  3. User can set the catalog directory, a default filename root, and the four catalog toggles (write HTML, copy to secondary location, watch directory, remember window size and position) (SET-04)
  4. User's settings save as they're changed with no explicit save step, and window size/position persistence continues to work exactly as it did pre-milestone, controlled by this same toggle (SET-05, COMPAT-05)

**Plans**: TBD
**UI hint**: yes

### Phase 27: Catalog Actions + Watch

**Goal**: Users manage existing catalogs — rename, duplicate, delete to Trash — and see the rail stay current when catalogs change outside the app
**Depends on**: Phase 23, Phase 26
**Requirements**: ACT-01, ACT-02, ACT-03, ACT-04, ACT-05, ACT-09, WATCH-01, WATCH-02, WATCH-03
**Success Criteria** (what must be TRUE):

  1. User can open a catalog actions menu from the `⋯` button in the details panel, rename a catalog's title (rewriting the `.html` `<title>` safely, including for titles with special characters, and leaving filenames unchanged), and duplicate a catalog with a suffixed filename root (ACT-01, ACT-02, ACT-03)
  2. User can delete a catalog to the OS Trash after a confirmation that names both file paths (with an option to also delete the matching `.html`), and a failed trash operation surfaces as an error and never silently falls back to permanent deletion (ACT-04, ACT-05)
  3. No catalog write — create, rename, duplicate, or delete — can corrupt an existing catalog file if the app crashes mid-write (ACT-09)
  4. User sees "● watching `<catalog directory>`" in the status bar when watching is enabled, sees the rail update when catalogs are added, removed, or modified outside the app, and can turn watching off in Settings with the underlying watcher released (WATCH-01, WATCH-02, WATCH-03)

**Plans**: TBD
**UI hint**: yes

### Phase 28: Re-scan & Diff

**Goal**: Users re-scan a catalog's source volume and safely reconcile what changed, without risking the existing catalog
**Depends on**: Phase 23, Phase 25, Phase 27
**Requirements**: ACT-06, ACT-07, ACT-08, STATE-03, COMPAT-06
**Success Criteria** (what must be TRUE):

  1. User can re-scan a catalog's source volume and see a diff of added, removed, changed, and unchanged entries with counts, always being asked to select the source volume rather than the app guessing which media it came from (ACT-06, ACT-08)
  2. User can resolve a diff by overwriting the catalog, keeping both, or discarding, with the overwrite path reusing the same crash-safe atomic write guarantee established in Phase 25 (ACT-07)
  3. User viewing an unreadable catalog can re-scan it, open its `.html` instead, or remove it from the library (STATE-03)
  4. The app builds, signs, notarizes, and releases on every existing CI platform target with the full milestone's changes included — confirming the frontend rewrite and the milestone's one new Go dependency (Bios-Marcel/wastebasket/v2) don't break any platform build (COMPAT-06)

**Plans**: TBD
**UI hint**: yes
**Research flag**: This is the handoff's own "biggest backend piece" — plan-phase should run its own research pass before planning, per SUMMARY.md's flag, to resolve the volume-relocation policy and finalize the on-disk "unreadable subtree" marker format from Phase 25.

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Data Models + Catalog Service | v2.0.0 | 2/2 | Complete | 2026-03-24 |
| 2. Search Service + Browse Metadata | v2.0.0 | 1/1 | Complete | 2026-03-25 |
| 3. Config Manager | v2.0.0 | 2/2 | Complete | 2026-03-25 |
| 4. App Layer + Lifecycle | v2.0.0 | 2/2 | Complete | 2026-03-25 |
| 5. Frontend Shim | v2.0.0 | 1/1 | Complete | 2026-03-26 |
| 6. Platform Integration | v2.0.0 | 1/1 | Complete | 2026-03-26 |
| 7. Verification + Merge | v2.0.0 | 2/2 | Complete | 2026-03-26 |
| 8. CLI Foundation and Platform Compatibility | v2.1.0 | 2/2 | Complete | 2026-03-26 |
| 9. Core Subcommands — Create, List, Search | v2.1.0 | 2/2 | Complete | 2026-03-26 |
| 10. Show, Open, and Output Polish | v2.1.0 | 2/2 | Complete | 2026-03-26 |
| 11. Tech Debt Cleanup | v2.1.0 | 1/1 | Complete | 2026-03-26 |
| 12. Repo Consolidation | v2.2.0 | 3/3 | Complete | 2026-03-27 |
| 13. CI Scaffold and Multi-Platform Build | v2.2.0 | 1/1 | Complete | 2026-03-27 |
| 14. Platform Packaging | v2.2.0 | 1/1 | Complete | 2026-03-27 |
| 15. Distribution Channel Automation | v2.2.0 | 2/2 | Complete | 2026-03-27 |
| 16. Secrets & Certificate Procurement | v2.3.0 | 3/3 | Complete | 2026-03-28 |
| 17. macOS Signing & Notarization | v2.3.0 | 1/1 | Complete | 2026-03-28 |
| 18. Windows Authenticode Signing | v2.3.0 | 1/1 | Complete | 2026-03-28 |
| 19. Homebrew CLI PATH | v2.3.0 | 1/1 | Complete | 2026-03-28 |
| 20. Windows CLI PATH via NSIS | v2.3.0 | 1/1 | Complete | 2026-03-28 |
| 21. Auto version and auto distribution | v2.3.0 | 1/1 | Complete | 2026-03-28 |
| 22. Shell + Token Layer | v3.0.0 | 7/7 | Complete   | 2026-08-13 |
| 23. Rail + Virtualized Tree | v3.0.0 | 6/6 | Complete   | 2026-08-13 |
| 24. Cmd-K Command Palette | v3.0.0 | 5/5 | Complete    | 2026-08-14 |
| 25. Create Slide-over + Progress/Cancellation/Partial-Catalog | v3.0.0 | 3/7 | In Progress|  |
| 26. Settings | v3.0.0 | 0/TBD | Not started | - |
| 27. Catalog Actions + Watch | v3.0.0 | 0/TBD | Not started | - |
| 28. Re-scan & Diff | v3.0.0 | 0/TBD | Not started | - |
