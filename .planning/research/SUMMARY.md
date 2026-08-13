# Project Research Summary

**Project:** StorCat v3.0.0 — Workspace Redesign
**Domain:** Go/Wails desktop app — full custom-UI frontend replacement (Ant Design to hand-built single-view workspace) plus new Go backend capabilities (progress events, volume enumeration, OS trash, filesystem watching)
**Researched:** 2026-08-13
**Confidence:** MEDIUM-HIGH

## Executive Summary

StorCat v3.0.0 replaces the entire three-tab Ant Design React frontend with a single-view "pro tool" workspace (rail + tree + details, Cmd-K palette, slide-over create flow) driven by a high-fidelity, pixel-exact design handoff, and adds the backend capabilities that design implies: live scan progress, volume detection, error-tolerant partial-catalog scanning, re-scan and diff, directory watching, and OS-trash delete. Comparable tools across every dimension of this redesign are well-documented (VS Code's Explorer/palette pattern, rclone/rsync-style progress UX, FreeFileSync/rsync-style diffing, Finder/Disk-Utility-style trash and volume conventions), so the "how do experts build this" question has strong, convergent answers.

The recommended approach across all four research tracks converges on extending the project's existing dependency-light discipline rather than reaching for new frameworks. Virtualization should be hand-rolled (~100 lines) because the design fixes row height to two constants, making windowing pure arithmetic. Theme tokens should be computed in TypeScript rather than via CSS color-mix(), because Wails does not control the WebKitGTK version on Linux and older distros would silently render broken colors. The one deliberate exception is OS trash: github.com/Bios-Marcel/wastebasket/v2 is worth adding because correct per-platform trash semantics are genuinely fiddly to hand-roll, and this library is verified zero-cgo everywhere, which matters since the project's cross-compiled/universal-binary CI matrix depends on staying cgo-free. Architecturally, most new capability slots into an existing seam: internal/catalog already has an unused ProgressCallback parameter with a stub no-op implementation and an explicit "future: use Wails events" comment - this milestone's progress/cancellation work fills that gap rather than inventing new plumbing. The CLI's 6 subcommands need zero changes throughout, because every new capability is additive.

The key risk is concentrated in one place: the progress-events + cancellation + partial-catalog rework touches the one piece of existing, tested code this milestone modifies (traverseDirectory's error-return contract and the ProgressCallback signature), making it the single highest-regression-risk work in the whole milestone. Live progress and the error-tolerant walk are two faces of the same Go code path and must be built together, not sequentially. Secondary risks cluster around the frontend rewrite: losing antd's implicit accessibility scaffolding silently, stacking-order regressions across five new overlay types, and a long tail of "looks done but isn't" traps specific to 40k-node virtualized trees. Mitigation for nearly all of these is establishing shared primitives early (a tokens/z-index module, a shared modal-behavior hook, a flatten-once contract) before the phases that reuse them.

## Key Findings

### Recommended Stack

The existing stack (Go 1.23, Wails v2.10.2, React 18, TypeScript, Vite) is unchanged and validated; this milestone adds almost nothing new. golang.org/x/sys is already resolved indirectly via Wails and just needs promotion to direct for volume/mount enumeration. fsnotify v1.10.1 is the standard for directory watching. Removing antd/@ant-design/icons is likely a net removal with no offsetting addition - all four remaining UI primitives are fully pixel-specced and narrow enough that a headless library would add more surface than it saves. IBM Plex fonts follow the exact self-hosted-woff2 pattern already used for Nunito.

**Core technologies:**
- Hand-rolled fixed-row virtualization (no npm library) - windows the 40k+-node tree; fixed 27px/34px rows make this ~100 lines of index arithmetic
- github.com/Bios-Marcel/wastebasket/v2 v2.0.3 - cross-platform OS trash, verified zero-cgo, MPL-2.0
- golang.org/x/sys (promote indirect to direct, v0.30.0) - volume enumeration/capacity, zero new dependency cost
- github.com/fsnotify/fsnotify v1.10.1 - watches catalog storage directory for the "watching" status indicator

### Expected Features

The design handoff is authoritative and final; feature research validated how each of its 10 designed behaviors works in comparable tools. All 10 map onto established patterns: VS Code/Linear/Raycast for the palette, rclone/rsync/Time-Machine for progress, macOS Disk Utility/Finder for volumes, rsync-partial-mode/TeraCopy for tolerant failure, FreeFileSync/rsync for diffing (path+size/mtime, not content hash).

**Must have (table stakes, P1):**
- Single-view workspace (rail + tree + details, no tabs)
- Cmd-K palette with 50-result cap + "showing N of M"
- Live scan progress (%, files, bytes, path, ETA, background hand-off)
- Volume detection & selection
- Create slide-over (form to progress to done/error)

**Should have (differentiators, P2/P3):**
- Partial catalog on scan failure - solves failing-SD-card pain better than general sync tools
- Re-scan & diff (added/removed/changed/unchanged) - handoff's own "biggest backend piece"
- Directory watching - cheap, also solves "deleted outside the app" staleness
- Rename/Duplicate/Delete-to-Trash, Settings, Empty/unreadable-catalog states

**Explicitly out of scope:**
- Full-text file-content search/indexing
- Continuous two-way sync
- Content-hash diffing for re-scan
- Multi-scan queue
- "Type to confirm" delete gate

### Architecture Approach

The frontend's single monolithic useReducer+Context needs to split by change-frequency, not feature area: hot per-keystroke state (rail filter, palette query) stays local, never entering shared context; ThemeContext handles rare/global settings; WorkspaceContext handles catalog/selection/expand state. Tree flattening must be a new Go method (LoadCatalogFlat), not a change to LoadCatalog, because cli/show.go depends on LoadCatalog's current nested-tree shape. Progress events reuse the existing-but-stubbed ProgressCallback seam; app.go is the only place runtime.EventsEmit is called, keeping internal/catalog and internal/search Wails-free and CLI-safe. Three new Go packages are needed (internal/volumes, internal/trash, internal/watch), each isolated as OS-integration concerns orthogonal to "what a catalog is."

**Major components:**
1. internal/catalog (extended) - walk/create/flatten/rename/duplicate/diff; ProgressCallback signature and traverseDirectory's error contract change here
2. internal/volumes / internal/trash / internal/watch (new) - OS-integration packages, Wails-free and CLI-agnostic
3. app.go - sole adapter layer for runtime.EventsEmit and new bound methods
4. Frontend workspace shell - toolbar/rail/tree/details/status-bar consuming split ThemeContext/WorkspaceContext, hand-rolled fixed-row virtualizer as the tree pane's core primitive

### Critical Pitfalls

1. **No real cancellation plumbing** - CreateCatalog has no context.Context today; without threading one, "Cancel" only hides the UI while the walk keeps running. Thread context.Context through traverseDirectory/CreateCatalog, checked at directory boundaries, wired into beforeClose too.
2. **Removing antd silently drops focus trap, Escape handling, and scroll-lock** across all four new overlay types - invisible on mouse-only testing. Build one shared useModalBehavior hook when the first custom overlay ships and reuse everywhere.
3. **Stacking-order regression** - the details panel outranking the slide-over/palette/dialogs is explicitly flagged by the design handoff as "an easy bug to reintroduce," only visible at the 1040-1279px responsive tier. Centralize z-index into one named-constants module early; re-verify at every phase adding an overlay.
4. **Non-atomic catalog writes risk corrupting existing data on crash** - re-scan & diff introduces the first "overwrite an existing file" path. Build a temp-file-same-dir + os.Rename helper once, reuse for create and overwrite.
5. **Silent fallback from OS-trash failure to os.Remove** converts a recoverable delete into a permanent one. Trash failures must surface through the existing error envelope, never fall through to permanent delete.

## Implications for Roadmap

The design handoff's own suggested build order (shell to rail+tree to palette to create slide-over to settings to catalog actions) is validated as sound, with one refinement: LoadCatalogFlat should land as its own small backend phase immediately before the virtualized rail+tree phase.

### Phase 1: Shell + Token Layer
**Rationale:** Everything else mounts inside or over this shell; establishes every shared primitive (z-index tokens, CSS reset, TS-computed theme tokens, no-drag convention) that later phases depend on.
**Delivers:** Toolbar, 3-pane grid, status bar shell, theme switching across all 11 themes with luminance-derived contrast, CSS reset.
**Addresses:** Table-stakes workspace layout; density/rail-position settings groundwork.
**Avoids:** Stacking-order, drag-region click-swallowing, contrast, color-mix() compat, missing CSS reset, scrollbar-width drift.

### Phase 2: LoadCatalogFlat (backend)
**Rationale:** Small, isolated, additive Go change that must precede the virtualized tree; keeps this milestone's lowest-risk backend work cleanly separated from its highest-risk backend work (Phase 5).
**Delivers:** App.LoadCatalogFlat, FlatNode/FlatCatalog types, DFS flattener reusing search.Service.LoadCatalog internally.
**Uses:** Existing internal/search parsing, new file in internal/catalog.
**Implements:** Flatten-in-Go boundary, keeps cli/show.go untouched.

### Phase 3: Rail + Virtualized Tree
**Rationale:** Independently shippable once Phase 2 lands; the single highest-complexity frontend item (40k+-node scale).
**Delivers:** Rail (BrowseCatalogs-driven), hand-rolled fixed-row virtualizer keyed on stable node id, expand/collapse, selection state.
**Avoids:** Key instability, "expand all" freeze, scroll-state leak across catalog switch, dynamic-measuring virtualizer, sidecar cache leaking into catalog JSON.

### Phase 4: Cmd-K Command Palette
**Rationale:** Depends on Phase 3's flat-array/ancestor-expansion primitives; reuses existing uncapped search for MVP.
**Delivers:** Palette overlay, keyboard-nav state machine, "reveal in tree" (expand ancestors + select + scroll into view).
**Uses:** Existing internal/search; establishes the shared useModalBehavior hook for reuse in Phases 5-7.
**Avoids:** Missing focus trap on first overlay, expand-then-scroll race across async state boundary.

### Phase 5: Create Slide-over + Progress/Cancellation/Partial-Catalog (backend + frontend, paired)
**Rationale:** The milestone's single highest-regression-risk work - the only phase modifying existing, tested code. Live progress and the error-tolerant walk share one Go code path and must ship together.
**Delivers:** Volume detection/selection (internal/volumes), threaded context.Context cancellation, throttled EventsEmit progress stream, error-tolerant walk distinguishing "volume vanished" from "one bad file," atomic temp+rename write helper, create-form to progress to done/error slide-over UI.
**Avoids:** Slide-over exit-animation lifecycle bugs, EventsOn StrictMode leak, unthrottled event flood, no real cancellation, goroutine leak on quit, naive walk conflating failure types, non-atomic write, partial-catalog marker schema drift, CLI signature breaks.
**Research flag:** Most needs deeper research during planning - largest genuine engineering risk, touches working tested code.

### Phase 6: Settings
**Rationale:** Mostly additive internal/config fields mirroring the existing SetTheme pattern exactly - low risk, can run in parallel with Phases 3-5.
**Delivers:** Density/rail-position/theme-card settings modal, catalog-directory and default-behavior toggles.
**Avoids:** Hand-rolled controls losing keyboard/ARIA - use real button/input under styling.

### Phase 7: Catalog Actions + Watch
**Rationale:** Correctly sequenced last among lower-risk items; groups the milestone's one new third-party dependency (wastebasket/v2) with directory watching.
**Delivers:** Rename (HTML-safe title rewrite), Duplicate (shared filename-suffix helper), Delete-to-Trash (internal/trash, no silent os.Remove fallback), internal/watch (non-recursive, debounced, Errors-channel-aware).
**Avoids:** Trash silent fallback, watcher leak/recursive-watch temptation, event-storm self-triggering, silent watch-reliability degradation, HTML title-rewrite corruption.

### Phase 8: Re-scan & Diff
**Rationale:** The handoff's own explicitly named "biggest backend piece" - correctly sequenced last, needs the error-tolerant walk from Phase 5 plus the flat-array infrastructure from Phase 2/3.
**Delivers:** Path-keyed diff (size/mtime comparison), four-way classification, overwrite-vs-keep-both write paths reusing Phase 5's atomic-write helper.
**Depends on:** Phases 2, 3, 5.
**Research flag:** Needs its own research pass - the open questions about re-locating the source volume and forced-close partial-write semantics must be resolved before implementation.

### Phase Ordering Rationale

- Backend prerequisites are pulled forward of the frontend steps that need them, since building frontend against a shape that will change is strictly more work than sequencing the backend first.
- The highest-regression-risk work (Phase 5) is isolated as its own phase rather than interleaved with lower-risk additive work.
- Re-scan & diff is last because it's a hard dependency on Phase 5's error-tolerant walk - building it earlier would mean rewriting the walk twice.
- Shared primitives (z-index tokens in Phase 1, modal-behavior hook in Phase 4) are established at their first point of need and reused thereafter, since "reimplemented per-surface" is the actual failure mode, not "never implemented."

### Research Flags

Needs deeper research during planning:
- **Phase 5** - highest engineering risk, touches existing tested code, needs the forced-close partial-write policy resolved.
- **Phase 8** - largest new backend surface, needs the volume-relocation policy resolved and the on-disk "unreadable subtree" marker format finalized.

Standard patterns (skip research-phase):
- **Phase 1** - established CSS/theming patterns with concrete recipes for every risk item.
- **Phase 2** - small, isolated, fully specified.
- **Phase 3** - virtualization approach and pitfalls fully enumerated with concrete prevention patterns.
- **Phase 4** - reuses existing search backend; UX conventions well-established.
- **Phase 6** - mirrors an existing, working internal/config pattern exactly.
- **Phase 7** - library choice and fsnotify patterns already researched and concrete.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Go stdlib/x/sys/Wails runtime source and npm registry verified directly at specific tags/versions; MEDIUM only on third-party Go library maintenance signals |
| Features | MEDIUM | No official StorCat-specific precedent exists (greenfield), but cross-verified against multiple well-documented comparable tools with every claim traced to a specific handoff section |
| Architecture | HIGH for structural findings (read directly from the repo); MEDIUM on two externally-verified claims (WebView2/WebKitGTK color-mix() support) |
| Pitfalls | MEDIUM-HIGH | Codebase-grounded claims are HIGH; Wails/browser/OS-specific claims are cross-checked but web-sourced, so MEDIUM |

**Overall confidence:** HIGH - all four research tracks independently converge on the same architectural boundaries and phase-sequencing logic, a strong cross-validation signal for a greenfield redesign.

### Gaps to Address

- **Re-scan & diff volume relocation:** should the app auto-relocate the originating volume for a re-scan, or always require re-selecting a source volume? Architecture research leans toward not over-engineering a "same card?" check, treating any mismatch as a normal diff. Resolve during Phase 8 requirements definition.
- **Partial catalog on forced window-close mid-scan:** architecture research recommends cancel, do not silently auto-write, since "Write partial catalog" is designed as an explicit, informed user action. Flagged as a product decision for requirements, must be explicitly decided before Phase 5 implementation.
- **Go trash library go/no-go:** Bios-Marcel/wastebasket/v2 is verified zero-cgo but carries a MEDIUM maintenance-confidence signal (single maintainer, 45 stars). Worth a final go/no-go check before Phase 7.
- **On-disk "unreadable subtree" marker format:** the partial-catalog feature needs a way to mark a failed subtree in the catalog JSON while keeping the format frozen and v1-compatible. The constraint is specified (must be omitempty, tested byte-for-byte against pre-milestone JSON shape) but the actual field name/shape needs design during Phase 5 planning.

## Sources

### Primary (HIGH confidence)
- Direct code reading: app.go, internal/catalog/service.go, internal/search/service.go, internal/config/config.go, pkg/models/catalog.go, frontend/src/contexts/AppContext.tsx, frontend/src/services/wailsAPI.ts, frontend/src/themes.ts, cli/create.go, cli/show.go, cli/search.go, go.mod, design_handoff_storcat_ui/README.md, .planning/PROJECT.md
- pkg.go.dev + raw GitHub source at Wails tag v2.10.2 for runtime.EventsEmit/EventsOn signatures
- npm view (live registry) for virtualization library versions/peer ranges; bundlephobia.com for verified bundle sizes
- GitHub Releases API for fsnotify (v1.10.1), gopsutil (v4.26.7), wastebasket (v2.0.3)
- Raw GitHub source inspection of Bios-Marcel/wastebasket confirming no-cgo and MPL-2.0

### Secondary (MEDIUM confidence)
- UX Patterns for Developers (command palette), VS Code UI docs, Adobe Bridge workspace docs - workspace/palette pattern validation
- FreeFileSync, rsync sync guidance - diff/comparison-key conventions
- rclone progress output docs, Backblaze/Time Machine failure-handling docs - progress UX and partial-failure posture comparables
- CSS color-mix() Chrome docs, WebKit color-mix commit - browser-compat basis for TS-computed-tokens recommendation
- Wails GitHub issues #2448/#2453, #3796, #3971/#1861/#5547 - integration gotchas
- fsnotify GitHub issue #18, watchexec inotify-limits doc - watch-scope and Linux inotify-limit context

### Tertiary (LOW confidence)
- PkgPulse virtualization library comparison - aggregator summary, not decision-critical since hand-rolling is the primary recommendation regardless

---
*Research completed: 2026-08-13*
*Ready for roadmap: yes*
