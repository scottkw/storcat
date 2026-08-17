# Milestones

## v3.0.0 Workspace Redesign (Shipped: 2026-08-16)

**Phases completed:** 7 phases, 43 plans, 113 tasks

**Key accomplishments:**

- Workspace shell tracer slice: OKLab-computed 14-token theme layer applied pre-paint, three-tab UI replaced by a 46px-toolbar/rail/tree/details/26px-status-bar grid, and macOS TitleBarHiddenInset wired in Go.
- Vendored 5 latin-subset IBM Plex woff2 files (Sans 400/500/600, Mono 400/500) via one-time `npm pack` extraction — never added as a `package.json` dependency — and declared them as `@font-face` in `style.css`, closing THEME-05's self-hosted, no-network typography requirement.
- Filled the toolbar seam with app mark/wordmark, an inert search field with ⌘K badge, a theme chip reading the persisted theme name, and a gear -- every interactive control explicitly opted out of the window drag region, plus a runtime `Environment()` call reserving the 78px macOS traffic-light inset.
- Filled the catalog rail's header (CATALOGS 0 label, accent New pill, directory chip, filter input) and unconditional empty state, plus the status bar's three literal mono zero-state segments — every string transcribed verbatim from the Copywriting Contract, nothing wired to a capability that doesn't exist yet.
- Filled the tree pane's centred empty state (the workspace's single accent-filled visual focal point) and the details panel's no-selection placeholder, built as one component with an exported-but-empty props interface ready for plan 22-07's pane/drawer variant.
- Pruned 16 dead tab-era reducer fields, added density/railSide/detailOverlay to AppContext, wrote the codebase's first hook (`useMediaQuery`), and wired the 1280px/1040px width tiers plus the right-side rail swap into `workspace.css` scoped to the widest tier only.
- Details panel becomes a 288px right drawer below 1280px (one close path for Escape/backdrop-click, --z-details-drawer stacking, no orphaned overlay state on resize), the toolbar gained its narrow-tier Details chip, and the phase's full manual verification matrix was run in a real browser via dev-browser -- catching and fixing one genuine THEME-02 defect (dead hover-inversion CSS) along the way.
- LoadCatalogFlat proven end-to-end at 42,550 nodes: 5.641 MB wire payload, 107.7ms Go-side flatten, 932.9ms browser time-to-first-row, virtualizer holds row count at 33-43 regardless of scroll position
- A mutex-guarded, atomically-written sidecar cache (proven race-clean under `-race`) backs the rail's file-count/byte-total columns, and `json.Valid()`-gated parse-status detection gives every catalog a red-dot-worthy `parseError` at the cost of one read plus one linear scan per catalog
- Real tree rows with caret/shape/click semantics, atomic scroll-reset-on-switch, three mutually exclusive empty/loading states, and a per-segment-coloured breadcrumb with O(n) expand-all/O(1) collapse -- all verified live against a 42,550-node fixture with measured timing, not assumed
- The rail's two-line rows, always-present status dot, isolated filter (proven via a live MutationObserver to leave the tree's DOM untouched during typing), an interactive directory chip, and a status bar summing the same array the rows render -- all verified against a real fixture directory with one deliberately corrupted catalog, not asserted from code reading alone
- A loaded catalog announces its title, companion chips (never fabricated for a missing HTML file) and exact metadata drawn from the flat catalog itself, while a broken one gets an inline, undismissable diagnostic with a verbatim Go error -- verified against a real truncated catalog and a real deleted-file race, the latter of which exposed and fixed a silent error-message bug in the Wails bridge
- Argv-only RevealInFileManager (three platform shapes unit-tested for exact hostile-path equality on one machine, macOS confirmed for real via AppleScript against a wails-built app), and a details panel that tracks the tree selection with a two-button footer that self-disables and surfaces its own errors
- Capped cross-catalog search (Go SearchIndexed + regenerated Wails bridge) wired to an always-mounted CommandPalette overlay opened by both ⌘K/Ctrl+K and the toolbar search button, with a 200ms-debounced, stale-guarded live search.
- PLT-01/02/03 all proven live: PLT-02/PLT-03 against a real 448-node two-catalog fixture directory at wails dev's :34115 (50-cap, true total, cross-catalog non-dedup, 2-char floor, 200ms debounce, stale guard -- all observed numbers), and PLT-01's ⌘K keyboard-open path confirmed by a human at the real native macOS StorCat.app window. RESEARCH Open Question #1 is answered POSITIVE: WKWebView does not reserve ⌘K.
- One `useModalBehavior` hook now owns focus trap, Escape-to-close, scroll lock, and focus restore for every overlay in the app; `CommandPalette.tsx` is its first consumer with its tracer-era inline Escape listener and `autoFocus` deleted.
- The ⌘K palette now renders a real listbox: highlighted-basename result rows built from JSX text children only (no injection path for a hostile catalog filename), a flat scroll-region with the exact "Showing the first 50 of N hits" truncation line sourced from Go's own total, and four mutually exclusive body states driven by a roving keyboard-navigable active index.
- A ⌘K search hit now switches catalog, merges its ancestor chain into the tree's expansion map (never replacing it), selects and centre-scrolls the target, and closes the palette — with the multi-branch-survival and cross-catalog races proven live against a real two-catalog fixture, not reasoned about from source.
- Split `CreateCatalog` into a ctx-aware `CreateCatalogWithContext` core with a byte-compatible CLI wrapper, bound it as `App.StartScan` with throttled `scan:progress` events and write-path containment, and wired an animated create slide-over end to end from the rail's + New pill to a real done state.
- Crash-safe catalog writes via temp-file-then-rename, plus a real walk error contract that classifies a vanished scan root (terminal) from a single bad entry (skip-and-continue, unchanged) -- with the on-disk partial-catalog marker shape decided by the user at this plan's blocking checkpoint.
- App-held cancel handle and retained partial-scan tree so a second bound call (CancelScan) stops an in-flight walk and a source-loss error's tree can be written exactly once (WritePartialCatalog); beforeClose gains a CRT-13 branch that cancels before quitting, and the frontend bridge (cancelScan/writePartialCatalog/classifyScanFailure) is wired for later plans to consume.
- stdlib-only (no golang.org/x/sys) per-OS volume enumeration in a new `internal/volumes` package, a count-only `MeasureTree` pre-pass, and the `App.ListVolumes`/`ScanOptions.TotalBytesHint` wiring that gives the create flow's progress bar a real, never-fabricated denominator.
- Turned the tracer's bare folder-only form into the real one: selectable volume cards with live size/status, independent title and filename-root fields backed by a deterministically-ordered live WILL WRITE preview, and the three creation toggles including a working secondary-location bootstrap -- all with zero new npm/Go dependencies.
- The running scan is now fully legible in both sub-states with real numbers throughout (no spinner, no fabricated percentage), every applicable close path actually cancels the walk and writes nothing, and a backgrounded scan stays visible and retrievable from the status bar's first-ever right-aligned segment.
- The interrupted-scan case this whole phase exists for is now honest end to end: a vanished source produces a real error state with three working recoveries, a finished scan lists exactly what landed on disk with real sizes and a real duration, and all four create entry points open at the form step and lock while a scan runs -- closing on a real classification bug (an instant source loss silently resetting to idle) found live-testing the very feature meant to catch it.
- Row density flows Cmd+,/gear/theme-chip → Settings dialog → write-through settingsStore.ts → lock-guarded Go config → disk, proving the whole Phase 26 architecture on one thin slice before horizontal expansion.
- 11-theme picker grid and the rail-position segmented control both wired through settingsStore's write-through pattern into the Go config, expanding the 26-01 tracer sideways with no new apply paths.
- Catalog directory and default filename root become config-backed settings shared between Settings and the rail/create-form, and the one-time marker-gated migration carries every existing user's five localStorage settings into the Go config non-destructively.
- Both remaining carried security obligations (T-22-05, FU-23-A) discharged: `GetCatalogHtmlPath`/`OpenExternal` now enforce the Settings-configured catalog directory via a new `osutil.ResolveContainedFileURL` validator, and the unreachable `CatalogModal.tsx` (unsanitized `iframe srcDoc`) is deleted outright with its `App.tsx` wiring.
- Four catalog toggles (write HTML, copy to secondary, watch directory, remember window) go live in Settings through one shared `ToggleRow` component; the remember-window toggle drives the pre-existing `windowPersistenceEnabled` field with zero new Go surface; and the phase's entire manual verification matrix — including COMPAT-05's persistence proven in both directions by real quit-and-relaunch cycles — is executed live and recorded.
- Catalog rename lands a `title` field authoritatively in the JSON root, patches both HTML title sites, and reads back through a corrected three-tier `BrowseCatalogs` title resolver -- proven live end-to-end against `wails dev` on :34115.
- `WriteFileAtomic` now `Sync()`s the temp file before close and best-effort `fsync`s the parent directory after rename (logging, not swallowing, a sync failure); a new subprocess SIGKILL harness proves the crash-safety claim with a real killed process, not a unit-test assumption -- `.planning/WINDOWS.md` #6 is dischargeable on real evidence.
- `catalog.DuplicateCatalog`'s -copy/-copy-N collision loop and `osutil.TrashPaths`'s containment-gated OS Trash wrapper (new dependency `wastebasket/v2 v2.0.3`) land as `App.DuplicateCatalog`/`App.DeleteCatalog` bindings, both gated identically to 27-01's `RenameCatalog`, with zero permanent-deletion fallback anywhere in the delete path.
- The details panel's `⋯` button now opens a real anchored, keyboard-navigable menu, and rename opens on a shared 440px dialog shell -- pre-filled, text-selected, Enter-to-commit, updating the rail optimistically -- while landing every CSS class this phase's four surfaces need in one pass.
- `DeleteConfirmDialog.tsx` closes the phase's one genuinely destructive surface -- both real file paths shown in full, a checked-by-default `.html` option that vanishes when there is no `.html`, and an error sub-state that shows the real OS Trash failure with only Keep catalog / Try moving to Trash again -- while `Duplicate catalog` finally runs for real from the actions menu.
- A Wails-runtime-free `internal/watch` package wraps `fsnotify` with a trailing-debounce coalescer and a drained error channel; `app.go`'s `applyWatchState` is the single start/stop point wired to launch, the watch-directory toggle, and catalog-directory changes, with a new `main.go` `OnShutdown` hook guaranteeing the watcher releases its OS handle on every quit path.
- The status bar's `● watching <dir>` segment and `CatalogRail`'s `catalogs:changed` subscription land WATCH-01/WATCH-02's frontend half; `.planning/WINDOWS.md` disposes entry #6 on 27-02's real SIGKILL evidence and records six Phase 27 gaps (five planned platform-unverified entries plus one live-verification finding); and the full 28-row Phase 27 verification matrix was run live against `wails dev :34115` via dev-browser, with 27 rows confirmed and one (menu click-outside focus restore) recorded as a genuine, unfixed discrepancy.
- End-to-end re-scan-and-diff tracer: details-panel footer button → volume picker → live scan → five diff stat tiles, proven live against a mounted volume with Create's output byte-identical throughout
- Made the locked fourth diff state reachable on a completed scan (MarkUnreadableOnSkip), completed the diff's semantics (unreadable-before-size ordering, type-change pairs, entries-derived sum invariant, unreadable-subtree descendant pruning), and made re-scan the diff's single opt-in caller
- Grouped diff row list (Added/Removed/Changed/Unreadable), the wrong-disc similarity banner, both diff-step variants, the re-scan error step, and the catalog-actions menu's second entry point -- all live-verified via dev-browser against a running `wails dev` instance
- `WriteRescanResult` and the `ResolveRescan` binding turn a completed re-scan diff into a decision -- overwrite in place, keep both via the shared collision loop, or discard with no write at all -- backed by the three-action resolution footer, all live-verified against real files on disk including a post-verification fix to a stale button label
- `UnreadableCatalogPanel`'s stub comment replaced with three buttons -- Re-scan volume (RescanDialog Variant B), Open the .html instead (reused open-HTML logic), Remove from library (the existing Phase 27 delete-to-Trash dialog) -- all live-verified end-to-end against real corrupted catalog files via dev-browser
- Local main (329 commits, the full v3.0.0 milestone) published to origin/main per the user's pre-made push-main decision; build.yml run 31976677486 observed green on all three native runners individually; WINDOWS.md swept to close exactly one entry (#7, live-verified) while leaving ten honestly open

---

## v2.3.0 Code Signing & Package Manager CLI (Shipped: 2026-03-28)

**Phases completed:** 6 phases, 8 plans, 12 tasks

**Key accomplishments:**

- macOS Developer ID code signing, notarization, and stapling automated in CI — Gatekeeper-verified end-to-end
- Windows Authenticode signing pipeline built with SSL.com eSigner integration (code complete, awaiting credential provisioning)
- Homebrew cask `binary` stanza puts `storcat` on PATH immediately after `brew install --cask storcat`
- Custom NSIS installer with EnVar PATH registration enables `storcat` CLI from any new terminal after WinGet install
- release-please automation: conventional commits → version bumps → tags → builds → publish → Homebrew/WinGet distribution
- GitHub `release` environment with tag protection rules, 6 Apple signing secrets, and credential rotation runbook

### Known Gaps

- **CRED-04**: Windows OV code signing certificate purchase deferred (SSL.com eSigner OV RSA ~$20/mo identified as vendor)
- **CRED-05**: 6/10 secrets in release environment (4 Windows eSigner secrets absent)
- **WSIGN-01–04**: Windows signing code complete but untested in CI (blocked by missing secrets)
- **Release pipeline cascade**: Windows build failure blocks full E2E release; macOS-only pipeline works

---

## v2.2.0 Repo Consolidation & CI/CD (Shipped: 2026-03-27)

**Phases completed:** 4 phases, 7 plans, 11 tasks

**Key accomplishments:**

- Consolidated WinGet manifests and Homebrew cask template into main repo under `packaging/`
- Archived `winget-storcat` satellite repo; marked `homebrew-storcat` as auto-managed
- Tag-triggered `release.yml` with 4 parallel platform builds (macOS universal, Windows, Linux x64+arm64) and fan-in draft release
- Platform packaging: macOS DMG (create-dmg), Windows NSIS installer, Linux AppImage + .deb
- `distribute.yml` auto-updates Homebrew cask and submits WinGet PR on release publish

---

## v2.1.0 CLI Commands (Shipped: 2026-03-26)

**Phases completed:** 4 phases, 7 plans, 9 tasks

**Key accomplishments:**

- stdlib flag.FlagSet CLI dispatch package with Run() entry point, version command, and 5 stub handlers — zero external dependencies, full --help/exit-code/stdout-stderr contract
- storcat list command with table/JSON output using tablewriter v1.1.4, shared printJSON/formatBytes helpers in cli/output.go
- storcat search and storcat create commands with table/JSON output, flag wiring, and comprehensive tests
- storcat show command with colorized tree rendering (fatih/color), --depth N truncation, and --json flag
- storcat open command with cross-platform browser launch (pkg/browser) and HTML path derivation
- Tech debt cleanup — closed all v2.1.0 audit gaps: NO_COLOR test, stale import, help stream consistency, orphaned export

---

## v2.0.0 Go/Wails Migration (Shipped: 2026-03-26)

**Phases completed:** 7 phases, 11 plans, 15 tasks

**Key accomplishments:**

- Go data models and catalog service match Electron format exactly — JSON bare object, empty dir `[]`, HTML tree with `└──` connectors, v1 backward compatibility
- Search and browse metadata with LoadCatalog dual-format parsing, browse Size column with human-readable formatting, RFC3339 dates
- Full window state persistence — size + position save/restore via Go config lifecycle hooks, settings toggle wired end-to-end
- All 17 wailsAPI wrappers return `{success,...}` envelopes matching Electron's contract — all 5 consumer components updated
- macOS header draggable via `--wails-draggable` CSS; version injected at build time via ldflags + GetVersion bound method
- Three-tab smoke test approved on macOS, feature branch merged to main — no bloat, proper .gitignore, CI aligned

---
