# Phase 27: Catalog Actions + Watch - Context

**Gathered:** 2026-08-15
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous) — 4 grey areas, 16 questions, all accepted as recommended

<domain>
## Phase Boundary

Users manage existing catalogs — rename, duplicate, delete to Trash — from an actions menu on the details
panel's `⋯` button, and the rail stays current when catalogs change on disk outside the app.

In scope: ACT-01 (actions menu), ACT-02 (rename title), ACT-03 (duplicate), ACT-04/ACT-05 (delete to Trash
with confirmation, never silently falling back to permanent deletion), ACT-09 (crash-safe writes for every
catalog write path), WATCH-01/02/03 (status-bar watching indicator, rail updates on external change, watcher
released when turned off).

Out of scope: re-scan and diff (Phase 28), any change to catalog *filenames* on rename (ACT-02 explicitly
leaves filenames unchanged).

</domain>

<decisions>
## Implementation Decisions

### Catalog identity & rename semantics
- **A `title` field is added to the catalog JSON root and becomes authoritative.** The scout established that
  a catalog's title currently lives ONLY in the `.html` `<title>` (`pkg/models/catalog.go`'s `CatalogItem` has
  no title field; `internal/search/service.go:172-262`'s `BrowseCatalogs` extracts `<title>` and falls back to
  the JSON filename). That leaves a JSON-only catalog with nowhere to persist a rename. Backward compatibility
  is preserved: when the field is absent, the existing HTML-then-filename fallback still applies, so every v1.0
  and v2.x catalog keeps reading exactly as it does today.
- **Rename is allowed on a catalog with no `.html`** — it writes the JSON title field; there is simply nothing
  to rewrite in HTML. No blocked/greyed state.
- **The read-side escaping bug is fixed in this phase.** `BrowseCatalogs` extracts `<title>` via a raw
  `strings.Index` substring with no `html.UnescapeString` anywhere in the repo, so a title containing `&`
  currently displays as the literal `&amp;`. The write side already escapes correctly (`html.EscapeString`,
  twice, in `internal/catalog/service.go:491`). Rename is precisely the feature that makes users type `&`,
  so shipping rename on a broken reader would manufacture the bug report. Fix the reader here.
- **Duplicate suffixes the filename root `-copy`, then `-copy-2`, `-copy-3` on collision.** Matches ACT-03's
  "suffixed filename root" wording and stays human-readable.

### Destructive actions & crash safety
- **Trash uses `github.com/Bios-Marcel/wastebasket/v2`** — named in the ROADMAP as the milestone's one new Go
  dependency, therefore pre-approved. Do not hand-roll per-platform trash.
- **A failed trash operation surfaces the real error and stops.** No permanent-deletion fallback is offered or
  performed, and no "delete permanently instead?" escape hatch exists in this phase. This is ACT-05's explicit
  requirement, not a default that may be softened.
- **The "also delete the matching `.html`" option is checked by default when an `.html` exists**, and hidden
  entirely when it does not. The `.json`/`.html` pair is what a user thinks of as "the catalog"; leaving an
  orphaned HTML is the surprising outcome.
- **`WriteFileAtomic` gains `File.Sync()` before close+rename, and gets a real SIGKILL-mid-write verification.**
  ACT-09 says no write may corrupt an existing catalog if the app crashes mid-write; `os.Rename` atomicity
  alone does not survive power loss, and `.planning/WINDOWS.md` #6 is the open ledger entry recording that the
  current guarantee is unit-tested only, never verified under a real crash. This phase closes that entry rather
  than building three more write paths on an unproven assumption.

### Actions menu & confirmation UI
- **A new minimal `Menu` component built on the existing `useModalBehavior` hook** — anchored popover,
  `role="menu"`/`role="menuitem"`, arrow-key navigation, Escape and click-outside to close, focus restored to
  the `⋯` button. No menu/dropdown/popover primitive exists anywhere in `frontend/src` (Ant Design's was
  removed in Phase 22 and never replaced), and `useModalBehavior`'s own doc comment names Phase 27 as an
  intended consumer. Do not add a dropdown dependency; do not write a second overlay implementation.
- **The menu skips scroll-lock** (`useModalBehavior` exposes `scrollLockSelector`, so this is configurable, not
  a fork). A small anchored dropdown that locks page scroll feels broken. It still gets focus trap, Escape,
  and focus restore.
- **The delete confirmation is a new centred dialog on `useModalBehavior`**, matching Phase 26's
  `SettingsDialog` shell. It names both file paths verbatim, carries the HTML checkbox, and uses a
  destructive-styled confirm button — introducing the project's first destructive color token.
- **Rename uses a text field in that same dialog shell**, pre-filled with the current title, Enter commits.
  Not an inline edit in the details-panel header.
- The `⋯` button already exists in `DetailsPanel.tsx:43-68` and is deliberately inert with its menu ARIA
  withheld. This phase wires it up; it does not add the button.

### File watching
- **Watching uses `fsnotify/fsnotify`.** Flagged explicitly: this is a **second** new Go dependency, beyond the
  one (`wastebasket`) the ROADMAP anticipated. Accepted as a USER decision — the alternative (polling
  `browseCatalogs` on a timer) is laggy and re-stats the whole catalog directory every tick.
- **Bursts are coalesced in Go with a ~300ms trailing debounce before emitting.** A large copy fires hundreds
  of fs events, and each rail refresh re-reads and re-stats every file in the directory
  (`CatalogRail.tsx:19-31` → `browseCatalogs`), with no existing debounce to absorb them.
- **The event is a bare `catalogs:changed` signal with no payload.** The rail already re-lists via
  `browseCatalogs`, which is idempotent and the single source of truth; per-file deltas would create a second
  code path that can disagree with it.
- **The watcher's lifecycle follows the `WatchDirectory` setting and the current catalog directory** — stop and
  restart on directory change, and genuinely release the underlying watcher on toggle-off and on app quit.
  WATCH-03 requires release, not merely ignoring events.
- **`runtime.EventsEmit` is called from `app.go` only.** `app.go:167-189`'s comment establishes this as a hard
  constraint: `internal/catalog` must stay usable from the CLI with no Wails runtime attached. A new watcher
  package must not emit directly.

### Post-research resolutions (2026-08-15, from 27-RESEARCH.md)
- **Rename rewrites BOTH the `<title>` tag and the `<h1>` heading.** Research found the generated HTML carries
  the title in two places (`internal/catalog/service.go:451,474,491`), not the one ACT-02's wording names.
  Patching only `<title>` would leave the page's visible heading disagreeing with its own tab title. Both get
  the same `html.EscapeString` treatment the generator already applies.
- **Every Trash-bound path goes through `osutil.ContainsPath` before reaching `wastebasket`.** Research read
  all four of the library's platform backends: `Trash()` never falls back to permanent deletion anywhere
  (so ACT-05 is satisfied by the library itself, not only by our error handling), and it silently no-ops on
  an already-missing path (so the UI-SPEC's retry button needs no "what already succeeded" bookkeeping). But
  its macOS backend shells out to `osascript` with a hand-built AppleScript string that escapes only literal
  `"` before interpolation — weaker than this codebase's own convention of passing every path as its own
  `exec.Command` argv element. The containment gate is the mitigation, applied for the same reason
  `RevealInFileManager` and `OpenExternal` have it.
- **The new JSON `title` field is `omitempty`** — preserves byte-parity for catalogs that have no title
  override, which COMPAT-02 depends on.
- **`DuplicateCatalog` inherits the source title verbatim.** Only the filename root gets the `-copy` suffix;
  ACT-03 speaks to the filename, not the title, and a user who wants a different title can rename after.

### Claude's Discretion
- Menu item ordering, labels, and icons within the actions menu.
- The exact JSON `title` field name and its position in the struct.
- Debounce implementation shape (timer reset vs. channel-based coalescing) and whether the 300ms is a named
  constant or configurable.
- Whether the destructive color token is a new theme token per-theme or a single derived value.
- Package layout for the watcher (`internal/watch` vs. a file beside an existing package).
- Whether `WriteFileAtomic` also fsyncs the parent directory after `os.Rename` (research rates this MEDIUM
  confidence and platform-divergent — Linux/macOS support it, Windows does not the same way). Decide it
  explicitly in the plan and record the choice; do not silently include or omit it.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/catalog/atomicwrite.go:21` — `WriteFileAtomic(path string, data []byte, perm os.FileMode) error`.
  Creates the temp file in the destination directory (never `os.TempDir()`), chmods, renames. Its doc comment
  at `:16-18` states it is exported specifically for Phase 27's rename/duplicate/delete. **No `fsync` today.**
- `internal/catalog/service.go:617` — `copyFile`, reads source fully then atomic-writes the destination. The
  exact pattern ACT-03's duplicate should reuse.
- `frontend/src/hooks/useModalBehavior.ts` — focus trap, Escape, scroll lock (configurable via
  `scrollLockSelector`), focus restore. Already used by `CommandPalette`, `CreateSlideOver`, `SettingsDialog`;
  its doc comment names Phase 27.
- `internal/osutil/reveal.go:93-109` — `ContainsPath` (`filepath.Abs` + `EvalSymlinks` + `filepath.Rel`), and
  `RevealInFileManager`'s pattern of always invoking via `exec.Command(name, args...)` with a distinct argv,
  never a shell string. The safety pattern any Trash code must mirror.
- `frontend/src/components/workspace/settings/` — Phase 26's `SettingsDialog` shell, the visual model for the
  new confirm dialog.

### Established Patterns
- **Backend → frontend push:** `runtime.EventsEmit(a.ctx, "scan:progress", …)` in `app.go:185`, guarded by
  `if a.ctx == nil { return }` so it is safe before Wails startup and from tests. Frontend subscribes with
  `EventsOn` from `wailsjs/runtime/runtime` and calls the returned unsubscribe in a `useEffect` cleanup
  (`CreateSlideOver.tsx:194`).
- **Catalog file pairing is filename convention only** — same directory, same basename
  (`strings.TrimSuffix(filePath, ".json") + ".html"`, at both `internal/search/service.go:189` and
  `app.go:759-765`). No JSON field records the HTML path.
- Go tests are table-driven `*_test.go` beside source. There is **no frontend test framework by design**
  (TEST-01 deferred) — frontend proof is `npx tsc --noEmit` + `npm run build` + live dev-browser on `:34115`.

### Integration Points
- `pkg/models/catalog.go` → new `title` field on the catalog root.
- `internal/catalog/` → rename/duplicate/delete operations; `atomicwrite.go` gains `fsync`.
- `internal/search/service.go:196-199` → the `<title>` reader gains `html.UnescapeString`.
- `internal/osutil/` → new trash helper, beside `openexternal.go` and `reveal.go`.
- New watcher package → consumed by `app.go`, which owns the `EventsEmit` call.
- `app.go` → new bindings for rename/duplicate/delete; watcher start/stop wired to `SetWatchDirectory` and to
  catalog-directory changes; `catalogs:changed` emission.
- `frontend/src/components/workspace/DetailsPanel.tsx:43-68` → wire the existing inert `⋯` button.
- New `Menu` + confirm-dialog components under `frontend/src/components/workspace/`.
- `frontend/src/components/workspace/StatusBar.tsx:60-81` → the "● watching …" segment, following the existing
  conditional `●`-prefixed scan segment's pattern.
- `frontend/src/components/workspace/CatalogRail.tsx:50-53` → the `state.catalogDir` effect is the reusable
  refresh path a `catalogs:changed` handler re-triggers.

</code_context>

<specifics>
## Specific Ideas

- Phase 26 shipped `config.WatchDirectory` + `SetWatchDirectory` (`internal/config/config.go:46,315-321`,
  `app.go:632-638`) with the Settings toggle note reading **"applies once file watching ships"**
  (`CatalogSettingsSection.tsx:136`). Nothing reads the value yet — this phase is what makes that note true,
  and the copy should be revisited once watching actually works.
- `.planning/WINDOWS.md` #6 (atomicwrite crash-safety, explicitly tied to "CRT-11/Phase 27 ACT-09") is the
  ledger entry this phase closes via the fsync + SIGKILL decision above.
- The project convention from WINDOWS.md is that platform-specific behavior which cannot be exercised on this
  dev machine gets logged as an open ledger item rather than silently claimed as done. Both new dependencies
  here are platform-specific under the hood (`wastebasket`'s trash backends, `fsnotify`'s watch backends), so
  expect new Windows/Linux entries rather than overclaimed coverage.
- `CLAUDE.md` at the repo root is stale v1.2.3 Electron-era content (references Ant Design and Electron IPC).
  Treat it as historical background; the live Go/Wails conventions are those established by Phases 22–26 in
  the code itself.

</specifics>

<deferred>
## Deferred Ideas

- Re-scan and diff (ACT-06/07/08, STATE-03) — Phase 28, as roadmapped.
- Renaming a catalog's *filenames* (as opposed to its title) — ACT-02 explicitly leaves filenames unchanged.
- Per-file incremental rail patching from watch events — rejected in favour of the idempotent full re-list;
  revisit only if a large catalog directory makes re-listing measurably slow.
- Undo for delete-to-Trash beyond what the OS Trash itself provides.

</deferred>
