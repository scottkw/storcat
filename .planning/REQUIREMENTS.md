# Requirements: StorCat v3.0.0 — Workspace Redesign

**Defined:** 2026-08-13
**Core Value:** Fast, lightweight directory catalog management — Go/Wails delivers 93% smaller binaries and 5x faster search than the original Electron version, with full feature parity.

**Design authority:** `design_handoff_storcat_ui/README.md` and `design_handoff_storcat_ui/designs/StorCat 1a Demo.dc.html` (direction **1a Workspace**). The handoff is high-fidelity and final — colors, type sizes, row heights, paddings, copy and geometry are matched exactly. Where this document and the handoff disagree, the handoff wins.

## v1 Requirements

### Shell — the workspace frame

- [x] **SHELL-01**: User sees a single workspace view — no tabs — with a 46px toolbar, catalog rail, tree pane, details panel, and 26px status bar
- [x] **SHELL-02**: User sees the three panes laid out as `268px 1fr 288px` at window widths ≥1280px
- [x] **SHELL-03**: User sees the details panel become a right drawer, toggled by a "Details" chip in the toolbar, at widths 1040–1279px (rail narrows to 236px)
- [x] **SHELL-04**: User sees the tree keep priority below 1040px (rail 200px, details stays a drawer)
- [x] **SHELL-05**: User can move the catalog rail to the right side, and the 1px divider moves with it
- [x] **SHELL-06**: User sees the status bar report catalog count, indexed file count, and total bytes
- [x] **SHELL-07**: User can drag the window by the toolbar without drag regions swallowing clicks on the search field, theme chip, or gear
- [x] **SHELL-08**: On macOS, user sees the real traffic lights sitting inside the 46px toolbar (TitleBarHiddenInset); on Windows and Linux the native title bar sits above it
- [x] **SHELL-09**: Overlays stack correctly at every window width — details panel below the create slide-over and search palette, which sit below dialogs and Settings

### Theme — token layer and appearance

- [x] **THEME-01**: User can switch between all 11 themes and the entire workspace repaints immediately
- [x] **THEME-02**: User sees legible text on accent-filled buttons and badges in every one of the 11 themes, including light accents (Gruvbox orange, Monokai green) and dark accents (GitHub blue)
- [x] **THEME-03**: User sees the theme colors defined by the handoff's `THEMES` array, with the extended token set (`--bg --p --p2 --ch --l --l2 --tx --dm --fn --ac --acs --onac --sel --hov`)
- [x] **THEME-04**: User can switch row density between Compact and Comfortable, changing tree row height, rail row padding, details row padding, palette row padding, and tree font size
- [x] **THEME-05**: User sees IBM Plex Sans for UI text and IBM Plex Mono for every path, filename, size, count and timestamp, with no network access required
- [x] **THEME-06**: User's theme, density, and rail position survive an app restart

### Rail — the catalog list

- [x] **RAIL-01**: User sees every catalog in the configured catalog directory listed in the rail, with title, JSON size, filename, and file count
- [x] **RAIL-02**: User can filter the rail by typing, matching case-insensitively against title and filename, without the tree re-rendering on each keystroke
- [x] **RAIL-03**: User can select a catalog from the rail, which loads its tree and clears the previous selection
- [x] **RAIL-04**: User sees a red status dot on the rail row of any catalog that failed to parse
- [x] **RAIL-05**: User sees the current catalog directory as a chip in the rail header, and can change it
- [ ] **RAIL-06**: User can open the create slide-over from the "＋ New" pill

### Tree — the catalog browser

- [x] **TREE-01**: User can browse a catalog of 40,000+ nodes with smooth scrolling and no freeze
- [x] **TREE-02**: User can expand and collapse directories; clicking a directory both toggles and selects it, clicking a file selects only
- [x] **TREE-03**: User can expand every directory in the current catalog, or collapse back to root, from the breadcrumb bar
- [x] **TREE-04**: User sees the catalog header with title, `.json`/`.html` chips, and the metadata line (file count, JSON size, bytes catalogued, modified date)
- [x] **TREE-05**: User sees a breadcrumb path for the current selection, with ancestor segments in the accent color
- [x] **TREE-06**: User's scroll position and expansion state do not leak between catalogs when switching
- [x] **TREE-07**: User sees the details panel follow the current selection, showing name, path, key/value metadata, and the actions footer
- [x] **TREE-08**: User can open the catalog's HTML from the details panel, and reveal its JSON in the OS file manager

### Palette — ⌘K search

- [x] **PLT-01**: User can open a search palette with ⌘K or by clicking the toolbar search field, with the input autofocused
- [x] **PLT-02**: User can search names and paths across every catalog in the directory
- [x] **PLT-03**: User sees at most 50 results, with a "Showing the first 50 of N hits" notice when more matched
- [x] **PLT-04**: User can navigate results by keyboard and dismiss the palette with Escape
- [x] **PLT-05**: User can click a hit to switch to its catalog, expand every ancestor, select it, scroll it into view, and close the palette
- [x] **PLT-06**: User sees "No file in any catalog matches that." when nothing matched
- [x] **PLT-07**: Focus is trapped inside the palette while open, and page scroll is locked behind it

### Create — the catalog creation flow

- [x] **CRT-01**: User can open a 560px right slide-over to create a catalog, which animates in over 340ms and out over 260ms without unmounting early
- [ ] **CRT-02**: User sees detected mounted volumes as selectable cards with name, mount path, size, and a `mounted` or `read errors` status
- [x] **CRT-03**: User can choose any folder instead of a detected volume
- [ ] **CRT-04**: User can set a catalog title and a filename root, and sees a live "WILL WRITE" preview of the files that will be produced
- [ ] **CRT-05**: User can toggle write-HTML-alongside-JSON, copy-to-secondary-location, and include-hidden-files
- [x] **CRT-06**: User can start the scan with the Create button or ⌘↵
- [x] **CRT-07**: User sees live scan progress — percentage, files seen, bytes, estimated time remaining, the current walking path, and a newest-first log
- [ ] **CRT-08**: User can hand a running scan to the status bar with "Run in background" and see `● scanning <name> · N%` there
- [x] **CRT-09**: User can cancel a scan and the underlying walk actually stops
- [x] **CRT-10**: User sees a distinct error state when the volume goes away mid-scan, showing where it stopped and the read errors encountered
- [x] **CRT-11**: User can write a partial catalog from the error state, retry the scan, or cancel
- [x] **CRT-12**: User sees a done state listing every file written with its size, and can open the new catalog in the workspace or catalog another volume
- [ ] **CRT-13**: Closing the window mid-scan cancels the walk and writes nothing

### Settings

- [ ] **SET-01**: User can open Settings with ⌘, , the gear, or the theme chip
- [ ] **SET-02**: User can pick a theme from 11 cards, each showing a 4-swatch strip and a light/dark tag
- [ ] **SET-03**: User can set row density and catalog rail position from segmented controls
- [ ] **SET-04**: User can set the catalog directory, a default filename root, and the four catalog toggles (write HTML, copy to secondary, watch directory, remember window size and position)
- [ ] **SET-05**: User's settings save as they are changed, with no explicit save step

### Actions — catalog management

- [ ] **ACT-01**: User can open a catalog actions menu from the `⋯` button in the details panel
- [ ] **ACT-02**: User can rename a catalog's title, which rewrites the `.html` `<title>` and leaves filenames unchanged
- [ ] **ACT-03**: User can duplicate a catalog, copying the `.json` and any `.html` with a suffixed filename root
- [ ] **ACT-04**: User can delete a catalog to the OS Trash after a confirmation that names both file paths, with an option to also delete the matching `.html`
- [ ] **ACT-05**: A failed trash operation surfaces as an error and never silently falls back to permanent deletion
- [ ] **ACT-06**: User can re-scan a catalog's source volume and see a diff of added, removed, changed, and unchanged entries with counts
- [ ] **ACT-07**: User can resolve a diff by overwriting the catalog, keeping both, or discarding
- [ ] **ACT-08**: Re-scan always asks the user to select the source volume rather than guessing which media the catalog came from
- [ ] **ACT-09**: No catalog write can corrupt an existing catalog file if the app crashes mid-write

### Watch — directory monitoring

- [ ] **WATCH-01**: User sees `● watching <catalog directory>` in the status bar when watching is enabled
- [ ] **WATCH-02**: User sees the rail update when catalogs are added, removed, or modified outside the app
- [ ] **WATCH-03**: User can turn watching off in Settings, and the watcher is released

### States — empty and error surfaces

- [x] **STATE-01**: User with no catalogs sees an empty-library state in both the rail and the tree pane, with "Catalog a volume" and "Choose catalog folder…" actions
- [x] **STATE-02**: User selecting an unreadable catalog sees why it failed — the file, the byte offset, the reason, and the raw parse error
- [ ] **STATE-03**: User can re-scan, open the `.html` instead, or remove an unreadable catalog from the library

### Compatibility — no regressions

- [x] **COMPAT-01**: User can open catalogs created by StorCat v1.x and v2.x without conversion
- [x] **COMPAT-02**: Catalogs written by v3.0.0 are byte-for-byte the same JSON shape as v2.3.0 wrote, so external tools keep working
- [x] **COMPAT-03**: All six CLI subcommands (`create`, `search`, `list`, `show`, `open`, `version`) behave exactly as they did in v2.3.0
- [x] **COMPAT-04**: `internal/catalog` remains usable from the CLI without a Wails runtime context
- [ ] **COMPAT-05**: Window size and position persistence continues to work, controlled by the Settings toggle
- [ ] **COMPAT-06**: The app builds, signs, notarizes, and releases on all existing CI platform targets

## v2 Requirements

Deferred — acknowledged but not in this roadmap.

### Interface

- **FUT-01**: Rail becomes a drawer below ~820px window width (not prototyped; implementer's call deferred)
- **FUT-02**: Frameless window with custom minimize/maximize/close controls on Windows and Linux
- **FUT-03**: CLI subcommands for the new capabilities (rename, duplicate, delete, rescan, volumes)

### Testing

- **TEST-01**: Frontend unit tests (Vitest + Testing Library) for the virtualizer, palette keyboard navigation, and modal behavior

## Out of Scope

| Feature | Reason |
|---------|--------|
| Full-text file-content search / indexing | StorCat catalogs names and paths; content indexing is a different product |
| Continuous two-way sync | StorCat is a cataloguer, not a sync tool — re-scan & diff covers the real need |
| Content-hash diffing for re-scan | Path + size (and mtime) matches rsync/FreeFileSync convention; hashing 40k files on removable media is prohibitively slow |
| Multi-scan queue | One scan at a time; "Run in background" already covers the real workflow |
| "Type to confirm" delete gate | Files go to Trash and are recoverable — a confirmation dialog is proportionate |
| Direction 1b Console / 1c Refit layouts | 1a Workspace was chosen; the others exist only as rejected context |
| Wails v3 migration | Still alpha, premature |
| Tailwind CSS migration | Different design direction, not prioritized |
| Changing the catalog JSON/HTML on-disk format | Frozen — v1 catalogs and external tools depend on it |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SHELL-01 | Phase 22 | Complete |
| SHELL-02 | Phase 22 | Complete |
| SHELL-03 | Phase 22 | Complete |
| SHELL-04 | Phase 22 | Complete |
| SHELL-05 | Phase 22 | Complete |
| SHELL-06 | Phase 23 | Complete |
| SHELL-07 | Phase 22 | Complete |
| SHELL-08 | Phase 22 | Complete |
| SHELL-09 | Phase 22 | Complete |
| THEME-01 | Phase 22 | Complete |
| THEME-02 | Phase 22 | Complete |
| THEME-03 | Phase 22 | Complete |
| THEME-04 | Phase 22 | Complete |
| THEME-05 | Phase 22 | Complete |
| THEME-06 | Phase 22 | Complete |
| RAIL-01 | Phase 23 | Complete |
| RAIL-02 | Phase 23 | Complete |
| RAIL-03 | Phase 23 | Complete |
| RAIL-04 | Phase 23 | Complete |
| RAIL-05 | Phase 23 | Complete |
| RAIL-06 | Phase 23 | Pending |
| TREE-01 | Phase 23 | Complete |
| TREE-02 | Phase 23 | Complete |
| TREE-03 | Phase 23 | Complete |
| TREE-04 | Phase 23 | Complete |
| TREE-05 | Phase 23 | Complete |
| TREE-06 | Phase 23 | Complete |
| TREE-07 | Phase 23 | Complete |
| TREE-08 | Phase 23 | Complete |
| PLT-01 | Phase 24 | Complete |
| PLT-02 | Phase 24 | Complete |
| PLT-03 | Phase 24 | Complete |
| PLT-04 | Phase 24 | Complete |
| PLT-05 | Phase 24 | Complete |
| PLT-06 | Phase 24 | Complete |
| PLT-07 | Phase 24 | Complete |
| CRT-01 | Phase 25 | Complete |
| CRT-02 | Phase 25 | Pending |
| CRT-03 | Phase 25 | Complete |
| CRT-04 | Phase 25 | Pending |
| CRT-05 | Phase 25 | Pending |
| CRT-06 | Phase 25 | Complete |
| CRT-07 | Phase 25 | Complete |
| CRT-08 | Phase 25 | Pending |
| CRT-09 | Phase 25 | Complete |
| CRT-10 | Phase 25 | Complete |
| CRT-11 | Phase 25 | Complete |
| CRT-12 | Phase 25 | Complete |
| CRT-13 | Phase 25 | Pending |
| SET-01 | Phase 26 | Pending |
| SET-02 | Phase 26 | Pending |
| SET-03 | Phase 26 | Pending |
| SET-04 | Phase 26 | Pending |
| SET-05 | Phase 26 | Pending |
| ACT-01 | Phase 27 | Pending |
| ACT-02 | Phase 27 | Pending |
| ACT-03 | Phase 27 | Pending |
| ACT-04 | Phase 27 | Pending |
| ACT-05 | Phase 27 | Pending |
| ACT-06 | Phase 28 | Pending |
| ACT-07 | Phase 28 | Pending |
| ACT-08 | Phase 28 | Pending |
| ACT-09 | Phase 27 | Pending |
| WATCH-01 | Phase 27 | Pending |
| WATCH-02 | Phase 27 | Pending |
| WATCH-03 | Phase 27 | Pending |
| STATE-01 | Phase 23 | Complete |
| STATE-02 | Phase 23 | Complete |
| STATE-03 | Phase 28 | Pending |
| COMPAT-01 | Phase 23 | Complete |
| COMPAT-02 | Phase 25 | Complete |
| COMPAT-03 | Phase 25 | Complete |
| COMPAT-04 | Phase 25 | Complete |
| COMPAT-05 | Phase 26 | Pending |
| COMPAT-06 | Phase 28 | Pending |

**Coverage:**

- v1 requirements: 75 total
- Mapped to phases: 75
- Unmapped: 0 ✓

---
*Requirements defined: 2026-08-13*
*Last updated: 2026-08-13 after roadmap creation — 7 phases (22-28), 100% coverage*
