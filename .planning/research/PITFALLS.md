# Pitfalls Research

**Domain:** Full custom-UI replacement (antd → hand-built workspace) plus new OS-level backend capability, in an existing Go 1.23 / Wails v2.10.2 desktop app
**Researched:** 2026-08-13
**Confidence:** MEDIUM-HIGH — codebase-grounded claims (read directly from `app.go`, `internal/catalog/service.go`, `go.mod`, `design_handoff_storcat_ui/README.md`) are HIGH confidence; Wails/browser/OS-specific claims are cross-checked against official docs and GitHub issues (MEDIUM confidence per the classify-confidence seam — web search is not a primary-source tier)

Phase names below follow the design handoff's **Suggested build order**: Phase 1 Shell, Phase 2 Rail+Tree, Phase 3 ⌘K Palette, Phase 4 Create Slide-over, Phase 5 Settings, Phase 6 Catalog Actions. Map these onto whatever phase numbers the roadmap assigns — the ordering and dependencies are what matter.

---

## Critical Pitfalls

### Pitfall 1: Stacking-order regression (details panel outranks the slide-over/palette)

**What goes wrong:**
The details panel (`z-index: 3`) ends up rendering above the create slide-over or ⌘K palette (`z-index: 6`) or above Settings/dialogs (`z-index: 7`) — e.g. the slide-over appears to open "behind" the details panel, or a dialog opened from `⋯` while a narrow-window details drawer is showing gets clipped underneath it. The design handoff calls this out explicitly as "an easy bug to reintroduce."

**Why it happens:**
Z-index values get set ad hoc per component (`style={{zIndex: 5}}` scattered across files) as each overlay is added in a later phase (Settings in Phase 5, catalog-action dialogs in Phase 6) by someone who doesn't recheck the details panel's value from Phase 1. Because the details panel only becomes an overlay/drawer at the 1040px responsive tier, the bug is invisible at the default 1280px+ window size used during normal development and CI screenshots.

**How to avoid:**
Define a single `z.ts`/`tokens.ts` module with named constants (`Z_DETAILS = 3`, `Z_OVERLAY = 6`, `Z_DIALOG = 7`) imported everywhere an overlay is styled; never hardcode a numeric z-index elsewhere. Add a narrow-window (1040–1279px) visual check to the Phase 4, 5, and 6 verification steps specifically, not just Phase 1, since new overlay types are introduced in each of those phases.

**Warning signs:**
Any component file containing a raw `zIndex:` or `z-index:` number outside the tokens module; a dialog/palette that "flashes behind" content when the window is resized below 1280px.

**Phase to address:**
Phase 1 (define the tokens); re-verify at Phase 4, 5, and 6 (each adds a new overlay type).

---

### Pitfall 2: Removing antd silently drops focus trap, Escape handling, and scroll locking

**What goes wrong:**
antd's `Modal`/`Drawer` provided focus trapping (Tab cycles within the dialog, not the page behind it), Escape-to-close, `aria-modal`/`role="dialog"` semantics, and scroll-lock with scrollbar-width compensation (so the background doesn't jump when its scrollbar disappears). Deleting antd wholesale and hand-rolling four different overlay surfaces (⌘K palette, create slide-over, Settings modal, `⋯` action dialogs) without deliberately reimplementing these means: Tab from the last field in Settings moves focus into the tree behind it; Escape works on some overlays and not others because each was wired independently; opening the palette or a dialog causes a 1-2px layout shift as the rail/tree's scrollbar appears/disappears.

**Why it happens:**
These behaviors are invisible when clicking through the happy path with a mouse — the kind of thing "a human would catch by eye" only via keyboard navigation, which an autonomous agent executing this milestone is unlikely to test by default.

**How to avoid:**
Build one shared `useModalBehavior` (or equivalent) hook — focus trap, Escape binding, scroll lock with scrollbar-width compensation — and apply it to all four overlay types from the moment antd's Modal is removed, rather than reimplementing per-surface. Write it once in Phase 1 or Phase 3 (whichever ships the first non-antd overlay) and reuse in Phases 4, 5, 6.

**Warning signs:**
Any overlay component that doesn't call the shared hook; Tab-key testing (not just click testing) not included in phase verification; background content visibly shifting when an overlay opens.

**Phase to address:**
Phase 3 (first custom overlay — the ⌘K palette) establishes the shared hook; Phases 4, 5, 6 reuse it, not reimplement it.

---

### Pitfall 3: Custom form controls (Select/Switch/Input) lose antd's keyboard and ARIA behavior

**What goes wrong:**
antd's `Select` provided arrow-key navigation and typeahead; `Switch` exposed `role="switch"` / `aria-checked`; `Input`/`Modal` carried proper labelling. The handoff explicitly permits keeping antd for `Modal/Select/Input/Tooltip` "where they earn their place," but if the theme cards, density/rail-position segmented controls, and the three toggle switches in Settings and the Create form are hand-rolled instead (since they're custom-spec'd visuals, not stock antd components), their keyboard/ARIA behavior needs to be built, not assumed.

**Why it happens:**
The visual spec (30×17px track, 13px knob) makes these controls look trivial to build as a `<div onClick>`, which works with a mouse but has no keyboard affordance or screen-reader semantics at all.

**How to avoid:**
For every hand-rolled interactive control (toggle, segmented control, theme card grid), use a real `<button>`/`<input type=checkbox role=switch>` under the hood so keyboard/AT behavior comes free, then style it to spec — don't build custom hit-testing on a styled `<div>`.

**Warning signs:**
Any interactive control implemented as a `<div>` with only an `onClick` handler and no `onKeyDown`/native element underneath.

**Phase to address:**
Phase 5 (Settings — most of these controls live here) and Phase 4 (Create form toggles).

---

### Pitfall 4: React key instability across expand/collapse breaks virtualization identity

**What goes wrong:**
The flattened `{depth, name, kind, size, parentIdx}` array's *visible* subset changes shape every time a directory is expanded or collapsed. If the virtualized list's `key`/`itemKey` is derived from the row's position in the currently-visible list (an array index) rather than a stable per-node identity, the virtualizer misattributes DOM nodes across renders — visible as a flicker, wrong-row content for one frame, or lost scroll anchoring, especially noticeable at 40k+ rows where the effect is amplified by React's reconciliation cost.

**Why it happens:**
It's the path of least resistance to key rows by their position in `.map()` over the currently-visible array, since that's what most virtualization tutorials show for simple flat lists — trees specifically need a stable id independent of visibility.

**How to avoid:**
Flatten the full tree once at catalog load into a fixed array with a stable `id` per node (e.g., the node's own index in that *original* full array, or its path), and key/derive visibility from that id — never from position in the currently-filtered visible list.

**Warning signs:**
`itemKey`/`key` prop is a function of the loop index in the rendering pass, not of a value stored on the node itself.

**Phase to address:**
Phase 2 (Rail + Tree, virtualized).

---

### Pitfall 5: "Expand all" and per-render re-flattening freeze the UI at 40k+ nodes

**What goes wrong:**
"Expand all" on a catalog with 40k+ nodes needs to add every directory's id to the `expanded` set and recompute which rows are visible — if this recomputation re-walks/re-flattens the *entire original tree* recursively (rather than operating on the already-flattened array), or if any hot-path handler (rail filter keystroke, hover) re-derives the flattened array from `CatalogItem.contents` each render instead of once at load, the app visibly freezes for large catalogs.

**Why it happens:**
The recursive `CatalogItem.contents` structure is the natural thing to walk again on any UI event, since that's the shape the Go backend returns; it's easy to forget the flatten-once contract established in Phase 2 when adding a feature in a later phase (search-hit expansion in Phase 3, re-scan diff highlighting in Phase 6).

**How to avoid:**
Flatten once per `LoadCatalog` call (memoized on catalog id), keep a separate cheap "visibility" derivation (a `Set`/`Map` lookup per row, O(n) but non-recursive) for the `expanded` state, and treat any new feature that needs "is this node visible/expanded" as consuming the existing flattened array, never re-walking `CatalogItem`.

**Warning signs:**
Any `function` that recursively walks `.contents` outside of the initial `LoadCatalog`-triggered flatten; noticeable input lag on "Expand all" or rail-filter typing tested only against small demo catalogs (dev testing typically uses small directories, not 40k-node ones).

**Phase to address:**
Phase 2 — and explicitly test with a synthetic 40k+ node catalog, not just the developer's own directory tree.

---

### Pitfall 6: Scroll position and virtualizer state leak across catalog switches

**What goes wrong:**
Clicking a different rail row is specified to "clear selection, expand its root" — but if the virtualized list component instance is *not* remounted (or explicitly scrolled to top) on catalog change, its internal scroll offset from the previous catalog persists. On a catalog with fewer total rows than the previous scroll offset, this renders a blank pane (scrolled past the end of content) that looks like nothing loaded.

**Why it happens:**
Virtualizers cache scroll state internally keyed by nothing more than "this component instance," so switching the data prop without changing the component's `key` silently reuses stale scroll state.

**How to avoid:**
Key the virtualized list component by `catalogId` (forces remount, resets scroll) or explicitly call the virtualizer's `scrollTo(0)` API in an effect keyed on catalog-id change. Decide (and note as a deliberate choice, not an accident) whether scroll position should be remembered per catalog — if so, store it explicitly per catalog id, don't rely on incidental component reuse.

**Warning signs:**
Switching catalogs (especially from a large one scrolled deep to a small one) shows an empty tree pane.

**Phase to address:**
Phase 2.

---

### Pitfall 7: Fixed-height rows implemented with a dynamic-measuring virtualizer

**What goes wrong:**
The spec is explicit that rows are fixed height (27px compact / 34px comfortable) specifically "to make windowing trivial." Reaching for a dynamic-measuring approach (e.g., `CellMeasurer`-style measure-then-cache) anyway — out of habit, or "for future-proofing" — reintroduces measurement-pass layout thrash and a class of bugs (measurement racing with density toggle changes) the fixed-height design was chosen to avoid.

**Why it happens:**
Many popular virtualization examples default to dynamic measuring because most real-world lists don't have uniform row heights; it's an easy default to reach for without checking whether it's needed here.

**How to avoid:**
Use a fixed-size virtualizer (or a small hand-rolled windowing calculation — `Math.floor(scrollTop / rowHeight)`) keyed on the current `--rh` density value; when density toggles, this is a simple constant swap, not a remeasure.

**Warning signs:**
Any dependency or custom code that measures row DOM height at runtime instead of using the known `--rh` constant.

**Phase to address:**
Phase 2.

---

### Pitfall 8: ⌘K "expand ancestors then scroll into view" lands on the wrong row

**What goes wrong:**
Clicking a search hit must switch catalogs, expand every ancestor of the hit, select it, and close the palette. If the scroll-to-index call uses the visible-list index computed *before* the ancestor-expansion state update is applied (i.e., scrolling and expanding happen in the same synchronous block using stale data), the list scrolls to where that index used to be pre-expansion, landing on an unrelated row.

**Why it happens:**
`expanded` state and `scrollToIndex` are two independent operations; the natural way to write this ("expand ancestors, then scroll") reads correctly but runs against React's async state update — the scroll call often executes against the pre-update render.

**How to avoid:**
Batch the ancestor-expansion state update, then compute and perform the scroll in a `useLayoutEffect` gated on a "pending scroll target" id that runs *after* the list has re-rendered with the new expanded set (not inline in the click handler).

**Warning signs:**
Search-hit selection lands near-but-not-exactly on the target row, or on a visibly wrong row, especially for deeply nested hits.

**Phase to address:**
Phase 3 (⌘K palette), depends on Phase 2's virtualization being in place first.

---

### Pitfall 9: Create slide-over unmounts before its 260ms exit animation plays

**What goes wrong:**
The most natural React implementation of "closing" a panel is `if (createOpen) return <Panel/>; return null;` — which removes the DOM node the instant `createOpen` flips to `false`, before the CSS transition has any frames to run. The panel simply vanishes instead of sliding out over 260ms.

**Why it happens:**
Conditional-render-on-boolean is the default mental model for "open/closed" UI in React; animating an *exit* requires keeping the node mounted past the state flip, which is a genuinely different pattern most component code doesn't need.

**How to avoid:**
Implement exactly what the spec names: a `createClosing` boolean, set true on any close trigger, with a `setTimeout(unmount, 260)` that flips the real unmount flag; render condition becomes `if (createOpen || createClosing)`. Apply the closing CSS class synchronously so the transform starts animating on the same frame `createClosing` becomes true.

**Warning signs:**
The panel disappears instantly instead of sliding; no `setTimeout`/`onTransitionEnd` tied to a close action anywhere in the component.

**Phase to address:**
Phase 4 (Create slide-over) — this is this phase's headline behavioral risk per the spec's own wording.

---

### Pitfall 10: Multiple independent close paths diverge and reintroduce the snap-close bug

**What goes wrong:**
Escape, the `×` button, "Cancel", the scrim click, and "Open in workspace" are five separate triggers. If each wires its own `setCreateOpen(false)` instead of routing through one function, a future edit to "fix" the animation on one path (say, Escape) won't propagate to the other four, and the bug in Pitfall 9 comes back piecemeal — some close paths animate, others don't.

**Why it happens:**
Each trigger is implemented where it's most convenient (Escape in a keydown handler, × in a button onClick, scrim in an overlay onClick) rather than being centralized, since they're written at different times/by different code paths even within one execution session.

**How to avoid:**
A single `closeCreatePanel()` function is the *only* thing that sets `createClosing`/schedules the unmount timer; all five triggers call only that function, never touch `createOpen`/`createClosing` directly.

**Warning signs:**
`grep` for `setCreateOpen\|createClosing` finding more than one call site setting the closing state.

**Phase to address:**
Phase 4.

---

### Pitfall 11: Re-opening the slide-over during its own exit animation double-fires the unmount timer

**What goes wrong:**
User clicks "+ New" again during the 260ms window after a previous close was triggered. If the pending `setTimeout` from the *previous* close isn't cancelled, it fires 260ms after the *original* close regardless of the new open state, incorrectly closing the freshly-reopened panel out from under the user (or, if the unmount flag is unconditional, causing a flash-close/reopen).

**Why it happens:**
Timers scheduled by a state transition are easy to forget need explicit cancellation when the same state is flipped again before the timer fires — a classic "stale closure" timer bug.

**How to avoid:**
Store the timeout id and `clearTimeout` it whenever `closeCreatePanel()` runs again or the panel is explicitly reopened; alternatively guard the timer callback by checking `createOpen === false` before performing the unmount.

**Warning signs:**
Rapidly clicking "+ New" → × → "+ New" within under 260ms closes the panel unexpectedly.

**Phase to address:**
Phase 4.

---

### Pitfall 12: `--wails-draggable` swallows clicks on toolbar controls added in later phases

**What goes wrong:**
The 46px toolbar is Wails' real draggable title-bar region (`--wails-draggable: drag`), and that property is inherited by all descendant elements unless explicitly overridden. Any interactive control placed inside it — search field, theme chip, gear icon in Phase 1, plus the "Details" toggle chip added later for the 1040-1279px responsive tier — needs an explicit `--wails-draggable: no-drag`, or clicks are consumed as (failed) window-drag gestures instead of firing the click handler, most visibly on Windows.

**Why it happens:**
The rule is easy to apply to the controls present when the toolbar is first built (Phase 1) and easy to forget for a control added in a later phase that also happens to live inside the same 46px band (the responsive-tier "Details" chip is added well after Phase 1, potentially by a different plan/session).

**How to avoid:**
Establish a rule (and ideally a shared `.no-drag` utility class) at Phase 1: *every* interactive element inside the toolbar gets `no-drag` by convention, checked in code review whenever a new control is added to that row — including the later-phase responsive "Details" chip.

**Warning signs:**
A toolbar button that requires an unnaturally precise/still click to register, or that occasionally starts a window-drag instead of activating, particularly on Windows/WebView2.

**Phase to address:**
Phase 1 (establish the convention); re-check whenever Phase 4-6 or the responsive-tier work adds a new toolbar-region control.

---

### Pitfall 13: EventsOn listener leaks / duplication under React 18 StrictMode

**What goes wrong:**
Wails' `EventsOn` returns an unsubscribe function (or requires an explicit `EventsOff` call); if a `useEffect` registers a listener for scan-progress events without returning a proper cleanup, React 18 StrictMode's development-mode double-invoke of effects (mount→unmount→mount) leaves two listeners active, so progress percentages/log lines appear to update twice or jump inconsistently — masked in a quick manual dev test but present in the shipped code path too (multiple mounts of the create panel over an app session compound it further).

**Why it happens:**
It's easy to call `EventsOn(name, cb)` in an effect and never write the cleanup, since the *first* mount looks correct; only StrictMode's dev-mode double-invoke (or genuinely remounting the panel) surfaces the duplication.

**How to avoid:**
Every `useEffect` that calls `EventsOn` must return a cleanup that unregisters that exact listener (`EventsOff(name)` or the handle Wails' JS API returns); verify by opening/closing the create slide-over several times in a StrictMode-enabled dev build and confirming progress log lines don't duplicate.

**Warning signs:**
Progress percentage or log lines appearing twice per actual event; a `useEffect` with `EventsOn` and no return statement.

**Phase to address:**
Phase 4 (progress events) — the noisiest and highest-frequency listener in the app.

---

### Pitfall 14: Emitting scan-progress events at raw filesystem speed floods the bridge

**What goes wrong:**
The prototype mocks progress at ~220ms ticks; a real walk over a fast SSD/SD card can produce many thousands of file-visit events per second. Emitting a Wails event per file (rather than batching/throttling) generates far more IPC traffic than the frontend needs, causing dropped frames on the "walking" path repaint, a stuttering percentage counter, and unnecessary marshal/dispatch overhead on both sides of the bridge.

**Why it happens:**
"Emit progress per file visited" is the natural place to put `EventsEmit` inside `traverseDirectory`'s existing per-entry loop; it's not obviously wrong until tested against a large, fast volume.

**How to avoid:**
Throttle emission from Go — e.g. a ticker emitting the latest state every ~150-250ms (matching the prototype's own cadence) regardless of how many files were visited in between, rather than emitting on every entry. Coalesce the "current path" and counters into one event payload per tick.

**Warning signs:**
Progress percentage or the "walking" path field updating in a stutter/burst pattern rather than smoothly; CPU spent on IPC marshaling visible in profiling during a scan of a large volume.

**Phase to address:**
Phase 4.

---

### Pitfall 15: No cancellation plumbing — "Cancel" only hides the UI, the walk keeps running

**What goes wrong:**
`CreateCatalog` today is a single blocking call with no `context.Context` parameter (confirmed in `internal/catalog/service.go`). Adding progress events without also adding real cancellation means a "Cancel" click in the UI can, at best, stop *listening* to progress events and close the slide-over — the Go-side walk goroutine keeps running, keeps consuming disk/CPU, and may still write output the user believed they cancelled.

**Why it happens:**
Progress events (the visible, spec'd feature) are the obvious thing to build; cancellation (implied but not explicitly walked through in the "Mocked functionality" table beyond a UI button) is easy to treat as "just close the panel."

**How to avoid:**
Thread a `context.Context` into `traverseDirectory`/`CreateCatalog`, check `ctx.Err()` at each directory boundary (cheap, since directories are the natural checkpoint granularity), and have the frontend's Cancel path call a Go method that cancels the context for that specific scan (track cancel funcs keyed by a scan id, since multiple scans could theoretically overlap via "Run in background").

**Warning signs:**
`CreateCatalog`/`traverseDirectory` signatures unchanged after adding progress events (no `context.Context` parameter); clicking Cancel and immediately re-triggering a scan on the same volume still shows disk activity from the "cancelled" scan.

**Phase to address:**
Phase 4.

---

### Pitfall 16: Goroutine leak / dangling write on app quit mid-scan

**What goes wrong:**
Running the scan in a background goroutine (required so `EventsEmit`-based progress doesn't block the bound Wails method / UI) without registering it against the app's shutdown lifecycle means quitting StorCat mid-scan doesn't cancel the walk — the goroutine may keep running past window close, or in the worst case be interrupted mid-write to the output JSON file.

**Why it happens:**
The existing `beforeClose(ctx) bool` hook in `app.go` currently only handles window-state persistence; it's easy to add scan logic elsewhere without wiring it into that same hook.

**How to avoid:**
Register each in-flight scan's cancel function with the app (a simple slice/map guarded by a mutex) and call cancel-all from `beforeClose` before allowing the app to quit, or block close briefly to let a fast cancel+cleanup complete. Combine with Pitfall 19's atomic-write approach so even an abrupt process kill can't corrupt existing output.

**Warning signs:**
`beforeClose` unchanged after adding scanning; force-quitting the app during a scan leaves a `.tmp` file behind or leaves the previous catalog JSON truncated.

**Phase to address:**
Phase 4.

---

### Pitfall 17: Naive walk treats "volume disappeared" the same as "one file unreadable" — losing the whole catalog instead of writing a partial one

**What goes wrong:**
The current convention (per `CLAUDE.md`/project conventions) is to silently skip inaccessible files/directories with a `console.warn`-equivalent and continue. That's correct for an ordinary permission error on one file, but the design's error state ("Stopped at 57% — the volume went away") requires distinguishing *that* case — where the root/mount itself becomes unreadable mid-walk — and stopping to offer "Write partial catalog" / "Retry scan" / "Cancel", rather than either (a) silently continuing as if nothing happened and producing a catalog that's missing a chunk of the volume with no error surfaced, or (b) aborting the entire walk and losing everything already collected.

**Why it happens:**
Go's `filepath.WalkDir` returning an error from a callback conflates "skip this one entry" and "stop entirely" behind the same return-value mechanism; getting the right one for I/O errors on a directory subtree (as opposed to a single stat failure) takes deliberate handling, not the default pattern already in use for per-file skips.

**How to avoid:**
Classify walk errors: single-entry permission/stat errors → skip + warn (existing behavior, keep it); I/O errors on a directory itself (`errors.Is(err, syscall.EIO)` or similar, indicating the underlying device went away) → treat as a terminal condition for that subtree, preserve everything walked so far, and surface it as the "volume went away" error path rather than either silently continuing or discarding partial results.

**Warning signs:**
Pulling the SD card mid-scan either aborts with no catalog written at all, or produces a "successful" catalog silently missing a chunk of the tree with no error indicator.

**Phase to address:**
Phase 4 (error path is explicitly named as the third sub-step of this phase, after form and progress events).

---

### Pitfall 18: Non-atomic catalog write risks corrupting an existing catalog on crash

**What goes wrong:**
Writing directly to `sd48.json` (the real output path) while generating it — whether on initial create or on "Overwrite catalog" from re-scan & diff — means a crash, kill, or power loss mid-write leaves a truncated/corrupt file where a valid previous catalog used to be. This is strictly worse for "Overwrite" than for "Create," since overwrite destroys previously-good data.

**Why it happens:**
`writeJSONFile`/`writeHTMLFile` (confirmed present in `internal/catalog/service.go`) write directly to the target path today, which is fine for "always create new, never overwrite," but re-scan & diff (Phase 6) introduces the first "overwrite existing file" code path in this codebase, and it's easy to reuse the existing write function verbatim without revisiting its safety properties for that new use case.

**How to avoid:**
Write to a temp file *in the same directory as the target* (same filesystem, so `os.Rename` is atomic — using `/tmp` or `os.TempDir()` would put it on a different filesystem/volume and break atomicity, or fail outright across an SD card boundary), then `os.Rename` over the final path only after the write succeeds and is flushed. Apply this to both the create path and the re-scan-overwrite path.

**Warning signs:**
`writeJSONFile`/`writeHTMLFile` open the destination path directly with `os.Create`/`os.OpenFile` rather than a temp-then-rename sequence; "Overwrite catalog" reuses these functions unchanged.

**Phase to address:**
Phase 4 (introduce the safe write helper); Phase 6 (re-scan & diff's "Overwrite catalog" must use it, not the raw create-time writer).

---

### Pitfall 19: "Move to Trash" silently degrades to permanent `os.Remove` on failure

**What goes wrong:**
Go's stdlib has no cross-platform trash API. The catalog-delete dialog explicitly promises "Move to Trash" with real recycle-bin semantics; a common shortcut is to reach for a trash library, and when it errors (permission denied, unsupported filesystem, no trash service available — genuinely common on minimal Linux setups) catch the error and silently fall back to `os.Remove` "so delete still works." That converts a user-recoverable action into a silent, permanent, unrecoverable one — precisely the "Silent Fallbacks" anti-pattern this project's own conventions warn against (`or {} converts hard failures into silent corruption`).

**Why it happens:**
Trash failures are rare enough in normal development/testing to go unnoticed, and "delete should always work" feels like reasonable resilience — but it silently breaks the specific guarantee ("goes to the Trash") the dialog copy promises the user.

**How to avoid:**
Use a maintained cross-platform trash library (wraps `SHFileOperationW` on Windows, the FreeDesktop.org Trash spec on Linux, a Cocoa-equivalent move on macOS) rather than hand-rolling per-OS trash paths — hand-rolled `~/.Trash` moves on macOS in particular lose Finder's "put back" metadata and can't handle a cross-device move (e.g. deleting a copy on a different mounted volume) the way a proper library does. On any trash failure, return the error through the existing `{success:false, error}` envelope and show it in the UI — never fall through to `os.Remove`.

**Warning signs:**
A `catch`/error branch around the trash call that calls `os.Remove` as a fallback; delete succeeding with no visible difference in behavior when the trash service is deliberately broken in a test environment (e.g. a container with no trash implementation).

**Phase to address:**
Phase 6 (Catalog actions — delete).

---

### Pitfall 20: fsnotify watcher not cleaned up on directory change or app quit; recursive-watch temptation on Linux

**What goes wrong:**
Two related mistakes: (1) not calling `watcher.Close()` on the old `fsnotify.Watcher` when the user changes the catalog directory in Settings, leaking one OS-level watch/file descriptor per change over the app's lifetime; (2) "helpfully" watching every subdirectory recursively (fsnotify is not recursive by default, and this feature only needs to watch one flat directory of catalog files) — which compounds leak (1) and runs headfirst into Linux's low default `fs.inotify.max_user_watches` (8192, shared per real UID across every app on the system, not per-process), risking `ENOSPC` failures that have nothing to do with StorCat's own watch count once combined with the user's editor/IDE/other file watchers.

**Why it happens:**
`watcher.Close()` on directory-change is easy to forget since the happy path (watch once at startup, never change directories) never exercises it; recursive watching feels "more correct" for future-proofing but isn't needed for a flat catalog directory.

**How to avoid:**
Keep the watch strictly non-recursive and scoped to the single catalog directory; on Settings "Change…", explicitly `Close()` the old watcher before creating a new one; on app quit, close the watcher in the same `beforeClose` hook used for scan cancellation (Pitfall 16).

**Warning signs:**
Any recursive directory walk that calls `watcher.Add()` per subdirectory; no `watcher.Close()` call anywhere in the codebase.

**Phase to address:**
Phase 5/6 (wherever "watch catalog directory for changes" and the directory-change setting are implemented — both live under Settings/Catalogs per the handoff).

---

### Pitfall 21: fsnotify event storms and self-triggered refresh loops

**What goes wrong:**
Naively calling `BrowseCatalogs` (a full directory re-scan) on every single fsnotify event — including events from unrelated file churn in the watched directory (sync-tool temp files, `.DS_Store`, etc.) and, notably, events generated by StorCat's *own* atomic rename-on-write (Pitfall 18) — causes redundant or excessive re-scans, visible as rail flicker or wasted CPU on a directory experiencing heavy unrelated activity.

**Why it happens:**
The obvious implementation is "on any fsnotify event, re-run BrowseCatalogs" — correct in spirit but naive about event volume and about the fact that the watched directory now includes writes StorCat itself just made moments earlier.

**How to avoid:**
Debounce fsnotify events (coalesce a burst within ~200-500ms into a single refresh) and filter to `.json`/`.html` extensions before reacting at all; treat a redundant refresh from a self-triggered event as an accepted, cheap no-op rather than trying to suppress it via write-tracking (simpler and less bug-prone than trying to distinguish "our write" from "someone else's write").

**Warning signs:**
The rail visibly re-renders/flickers immediately after StorCat's own create/rename/delete operations, beyond the expected optimistic UI update.

**Phase to address:**
Phase 5/6.

---

### Pitfall 22: Watch reliability silently degrades on removable/network media and on unmount

**What goes wrong:**
If the catalog directory being watched is ever on removable media or a network share (NFS/SMB notoriously don't deliver inotify/FSEvents reliably, and Linux inotify + macOS FSEvents both behave inconsistently across mount/unmount of the watched path itself), the watcher can silently stop delivering events with no error — yet the status bar's "● watching ~/dev/sd-catalogs" indicator has no reason to know that and keeps claiming to be active.

**Why it happens:**
fsnotify's `Errors` channel is easy to leave unhandled (just reading `Events` and ignoring `Errors`), and a watch that "goes quiet" rather than erroring outright has no natural signal to react to.

**How to avoid:**
Read and act on the watcher's `Errors` channel explicitly, treating repeated errors as "watch lost" and downgrading the status bar to a neutral/non-watching state rather than leaving a stale "watching" claim. For the specific case of the catalog *output* directory (normally on the internal disk, per the project's own conventions) this is lower-risk than watching removable media directly — scope the feature's documented guarantee to local/internal disks, and treat network or removable watch targets as best-effort.

**Warning signs:**
No code path reads `watcher.Errors`; the "watching" status never changes state once set, regardless of what happens to the underlying path.

**Phase to address:**
Phase 5/6.

---

### Pitfall 23: `--onac` accent-on-fill text isn't luminance-derived, breaking contrast on light-accent themes

**What goes wrong:**
Hardcoding a single "text on accent" color (always dark, or always white) for every primary button, selected-chip, and badge breaks legibility the moment a theme's accent is light (e.g. Gruvbox orange, Monokai green) — the design handoff calls this out by name and provides the fix (`--onac`, computed via relative luminance, `> 0.45 → dark text`).

**Why it happens:**
It's the single easiest token to skip deriving, since a hardcoded value looks fine for whichever theme happens to be active during development (likely StorCat Dark, which has a dark-appropriate accent) and the bug only appears when cycling through all 11 themes — something manual click-through testing tends to skip past quickly.

**How to avoid:**
Port the exact 6-line relative-luminance helper the spec names, apply it per-theme at theme-selection time (not per-button), and explicitly verify all 11 theme cards/primary buttons/badges as part of Phase 1's (or wherever theming lands) verification — cycling every theme, not just spot-checking the default.

**Warning signs:**
A single constant (not a per-theme computed value) used for text-on-accent color anywhere in the codebase; low-contrast/illegible text on primary buttons when switching to a light-accent theme.

**Phase to address:**
Phase 1 (token layer + theme switching).

---

### Pitfall 24: Runtime `color-mix()` CSS silently fails on older system webkit2gtk (Linux AppImage)

**What goes wrong:**
Several derived tokens (`--l2`, `--dm`, `--fn`, `--sel`, `--hov`, `--acs`) are spec'd as literal CSS `mix(...)` expressions, implying `color-mix()` at runtime. macOS (any Wails-supported version) and Windows (WebView2, evergreen Chromium) are safe, but this project's own packaging decision has the Linux AppImage depend on the **system-installed** webkit2gtk rather than a bundled runtime — an older LTS distro's webkit2gtk can predate WebKit's `color-mix()` support (shipped in Safari 15 / autumn 2021), silently resolving borders, hover states, and muted text to invalid values. The UI doesn't crash — it just looks subtly broken (missing borders, no hover feedback, wrong text contrast) in a way that's easy to miss unless tested on an old-enough distro.

**Why it happens:**
The spec's token table is written in `mix(...)` shorthand as documentation of the *math*, not necessarily as a literal instruction to ship raw CSS `color-mix()`; it's easy to read it as "use CSS color-mix()" when the safer reading — given this project's own already-known Linux WebKit constraint — is "compute these values in TypeScript."

**How to avoid:**
Precompute all derived theme tokens in TypeScript at theme-load time (the same relative-luminance/alpha-blend math either way, just run in JS rather than shipped as CSS) rather than relying on runtime `color-mix()` support. This sidesteps the webkit2gtk floor question entirely and is consistent effort with already having to port the `--onac` luminance helper (Pitfall 23).

**Warning signs:**
Raw `color-mix(in srgb, ...)` strings appearing in shipped CSS/CSS-in-JS rather than precomputed hex/rgba values; no verification of the app on an older Linux distro's webkit2gtk before release.

**Phase to address:**
Phase 1 (token layer) — decide TS-computed vs. CSS `color-mix()` here, since every later phase's styling depends on this choice.

---

### Pitfall 25: Default browser styles and missing CSS reset surviving antd's removal

**What goes wrong:**
antd ships/relies on a normalize baseline; deleting it wholesale without adding an explicit minimal reset leaves default UA styles (button padding/border, input appearance, list margins, focus rings) bleeding through in exactly the tight, exact-spec areas most likely to be visually checked against the prototype — the 30px footer buttons, 26px filter input, 27/34px tree rows.

**Why it happens:**
It's easy to focus effort on the custom components that are visually obvious (rail, tree, details) and miss that the *absence* of antd's baseline reset is itself a regression that needs an explicit replacement, not just "component code that doesn't use antd."

**How to avoid:**
Add a minimal explicit reset (`box-sizing: border-box` globally, zeroed margin/padding on the actual element types used, `-webkit-font-smoothing: antialiased`) as an early Phase 1 deliverable, and verify it against the prototype pixel-for-pixel (side-by-side with `StorCat 1a Demo.dc.html` open) rather than trusting visual memory of "close enough."

**Warning signs:**
Buttons/inputs with visible default OS chrome (native focus ring style, default border radius) anywhere in the built app; no explicit reset stylesheet in the diff that removes antd.

**Phase to address:**
Phase 1.

---

### Pitfall 26: Scrollbar width steals content space differently across platforms in the fixed 268px/288px panes

**What goes wrong:**
Windows (WebView2) reserves scrollbar width by default (classic layout-affecting scrollbar) where macOS overlay scrollbars don't; in the fixed-width rail (268px) and details (288px) panes, this means the same content that fits exactly on macOS truncates/ellipsizes a character earlier — or right-aligned values (sizes, counts) shift — on Windows, purely because ~15-17px of the pane's width is consumed differently per platform.

**Why it happens:**
Development and visual comparison against the prototype happen almost entirely on macOS (per this project's build history and CLAUDE.md's macOS-centric conventions); the Windows-specific scrollbar behavior is invisible until tested on an actual Windows build.

**How to avoid:**
Apply `scrollbar-gutter: stable` (or explicit thin/overlay-style scrollbar CSS) uniformly across the rail, tree, and details scroll regions so content width is consistent regardless of scrollbar visibility/platform, and cross-check visually on an actual Windows build (CI artifact or VM), not only macOS dev.

**Warning signs:**
Text/values that fit cleanly in local (macOS) dev testing but ellipsize or misalign in a Windows build; no `scrollbar-gutter` or scrollbar styling anywhere in the CSS.

**Phase to address:**
Phase 1 (establish the rule); explicitly verify on Windows at whichever phase first ships a scrollable fixed-width pane (Phase 2 — rail/tree).

---

### Pitfall 27: Sidecar file-count/byte-total cache leaks into the frozen catalog JSON schema

**What goes wrong:**
The project's own milestone decision is that per-catalog file count and total bytes come from "a sidecar cache keyed by path+mtime; catalog JSON on disk is unchanged." The natural shortcut when wiring up the rail's `2,481 files` line is to compute these once and just add `fileCount`/`totalBytes` fields directly onto the catalog JSON's root object — which is simpler to implement than a separate cache but silently changes the frozen on-disk schema that v1 compatibility and external tools depend on.

**Why it happens:**
Adding a field to an existing struct and re-serializing it is far less code than standing up a separate cache file/keying scheme, so it's the path of least resistance for whoever wires up this specific UI line — especially under autonomous execution optimizing for "make the number show up."

**How to avoid:**
Keep the sidecar cache in a genuinely separate location (e.g. under `internal/config`'s data directory, never inside the `.json`/`.html` output), and add an explicit regression check comparing a freshly-created catalog's JSON top-level key set against the pre-milestone schema (should be unchanged) as part of whichever phase adds this feature.

**Warning signs:**
A diff that adds new fields directly to the `CatalogItem` or catalog-root struct that gets marshaled to the `.json` output; `git diff` on a freshly generated catalog JSON showing new top-level keys.

**Phase to address:**
Phase 2 (rail needs file count/bytes) — verify before Phase 6 reuses/depends on the same cache for diffing.

---

### Pitfall 28: Error-tolerant walk / partial-catalog markers change the JSON shape even for successful scans

**What goes wrong:**
Building the "volume went away" partial-catalog feature (Pitfall 17) plausibly involves adding a new field/marker to `CatalogItem` for "this subtree failed." If that field is added unconditionally to the struct's JSON tags rather than being genuinely optional/omitted on the common (fully successful) path, every normal catalog created after this milestone carries a new field absent from v1/v2.0 catalogs — which is a schema drift even if each individual value is harmless, and risks confusing external tools or a subsequent `LoadCatalog` dual-format detection heuristic that inspects shape, not just field presence.

**Why it happens:**
It's simpler to always include a field (even as `null`/empty) than to carefully `omitempty` it and test both branches; the difference is invisible unless someone diffs a successful scan's output before and after the change.

**How to avoid:**
Any error/partial marker must be `omitempty`/entirely absent on the clean-scan path, added only to the specific node that failed. Add a test that generates a catalog via the new create path with no errors and asserts byte-for-byte (or at minimum key-for-key) equivalence with a pre-milestone catalog's JSON shape.

**Warning signs:**
A new struct field without `omitempty` (or Go-equivalent conditional serialization) on `CatalogItem`; no test comparing pre/post-milestone JSON shape for a clean scan.

**Phase to address:**
Phase 4 (error path).

---

### Pitfall 29: New Go-side capability breaks the CLI's shared use of `internal/catalog`/`internal/search`

**What goes wrong:**
The 6 CLI subcommands call directly into `internal/catalog`/`internal/search`. Any signature change made to support new GUI-only features — a `context.Context` parameter becoming required for cancellation, a progress callback becoming non-optional, error-tolerant-walk changing what counts as a fatal vs. skippable error — either fails to compile against the CLI's call sites, or compiles but silently changes CLI *behavior* (e.g. `storcat create` starting to swallow errors it used to report, or blocking on an event-emit call that has no listener outside a Wails runtime).

**Why it happens:**
The GUI-facing features are what the milestone is about, so it's natural to design new function signatures around what the GUI needs (a live Wails runtime context for `EventsEmit`, a cancellable context) without re-checking the 6 CLI call sites that share the same package.

**How to avoid:**
Keep every new capability additive and optional: a `context.Context` parameter defaults to `context.Background()` for CLI callers who don't need cancellation; progress reporting stays behind the same `ProgressCallback` pattern already used in `traverseDirectory(dirPath, basePath string, onProgress ProgressCallback)` — extend that existing interface rather than introducing a Wails-coupled `runtime.EventsEmit` call inside `internal/catalog` itself (which would panic or hang if invoked outside an active Wails app context, e.g. from the CLI). Re-run the existing `cli/*_test.go` suite after every `internal/catalog`/`internal/search` change, not just once at the end of the milestone.

**Warning signs:**
A `runtime.EventsEmit` call appearing inside `internal/catalog`/`internal/search` rather than behind a callback the CLI can pass as a no-op; `cli/*_test.go` not run as part of Phase 4/6 verification.

**Phase to address:**
Phase 4 (progress/cancellation) and Phase 6 (re-scan & diff's overwrite path) — re-run CLI tests at both.

---

### Pitfall 30: Rename catalog's `<title>` rewrite corrupts the HTML file on special characters

**What goes wrong:**
Rename catalog rewrites the `.html` file's `<title>` tag (the source of truth `BrowseCatalogs` reads back for the rail display, per the handoff). A naive implementation — regex or plain string replace on `<title>...</title>` — breaks on titles containing `<`, `&`, `"`, or non-ASCII characters (e.g. "SD Card — Kids' Photos 2024"), either corrupting the HTML structure outright or writing unescaped characters that break parsing, silently breaking the rail's display of that catalog on next launch.

**Why it happens:**
String-replace on an HTML tag looks like the simplest possible implementation of "change the title," and works for the obvious test case (a plain alphanumeric title) that's likely to be used during development/verification.

**How to avoid:**
Use proper HTML-safe escaping (Go's `html/template` escaping, or parse via `golang.org/x/net/html` and mutate the title node) rather than blind string replacement; explicitly test rename with titles containing `<`, `&`, `"`, an em-dash, and non-ASCII characters.

**Warning signs:**
A regex or `strings.Replace` used to rewrite the `<title>` tag; no test case for rename with special characters in the new title.

**Phase to address:**
Phase 6 (Catalog actions — rename).

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|-----------------|------------------|
| Capping per-folder rendering at ~200 with a "show more" affordance instead of full virtualization | Less virtualization plumbing to build | Doesn't scale to a single folder with 10k+ direct children (large flat directories exist on SD cards); the spec explicitly offers this as the *fallback* to full virtualization, not the preferred approach | Only if full virtualization genuinely can't ship in Phase 2's timebox — the spec names it as acceptable, but full virtualization should still be the default target |
| Hardcoding a single `--onac` value instead of the luminance helper "for now" | One less helper to write in Phase 1 | Breaks contrast on ~half the 11 themes; expensive to retrofit later since every primary-button/badge surface needs re-touching once discovered | Never — the helper is spec'd as "6 lines," there's no meaningful time saved |
| Skipping the atomic temp+rename write pattern for the *first* create path and adding it only for re-scan/overwrite | Simpler Phase 4 implementation | The create path is far more common than overwrite and is just as vulnerable to a mid-write crash corrupting a *previous* create if the user retries after a failed scan | Never — implement once in a shared write helper, use it everywhere |
| Falling back to `os.Remove` when a trash library call fails | Delete "always works" | Silently converts a recoverable action into a permanent one; violates project's own no-silent-fallback convention | Never |
| Watching the catalog directory recursively "to be safe" | Feels more robust against future subdirectory structures | Multiplies OS watch-handle usage and Linux inotify limit exposure for a feature that only needs one flat directory watched | Never for this feature's actual scope |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|-----------------|-------------------|
| Wails `EventsOn`/`EventsEmit` bridge | Registering listeners without cleanup; emitting per-file instead of throttled | `useEffect` cleanup always calls `EventsOff`; Go-side ticker throttles emission to ~150-250ms cadence |
| Wails frameless title bar (`--wails-draggable`) | Assuming the toolbar's drag region doesn't need per-control opt-out as new controls are added in later phases | Every interactive element inside the 46px toolbar band gets `no-drag`, checked whenever a new one is added |
| CLI ↔ `internal/catalog`/`internal/search` shared package | Designing new function signatures purely around GUI needs (Wails runtime context, required cancellation) | Additive/optional parameters only; extend the existing `ProgressCallback` pattern rather than coupling `internal/catalog` to `runtime.EventsEmit` directly |
| fsnotify ↔ catalog directory | Watching recursively or reacting to every raw event including the app's own writes | Non-recursive single-directory watch; debounce + extension-filter events; treat self-triggered refresh as an accepted no-op |
| OS Trash (no Go stdlib API) | Hand-rolling `~/.Trash`/Recycle Bin paths, or falling back to `os.Remove` on failure | Maintained cross-platform trash library; surface trash errors, never silently degrade to permanent delete |
| Linux AppImage's system webkit2gtk | Assuming CSS features available in the *newest* WebKit (used by the design prototype's dev browser) work identically on an end user's distro webkit2gtk | Precompute theme-derived colors in TypeScript rather than relying on runtime `color-mix()`, given the project's own already-accepted system-WebKit dependency |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|-----------------|
| Re-flattening `CatalogItem.contents` recursively on every keystroke/hover instead of once per load | Input lag on rail filter or tree interaction | Flatten once per `LoadCatalog`, derive visibility from a cheap non-recursive lookup | Catalogs above a few thousand nodes; severe at 40k+ |
| Emitting a Wails event per file visited during a scan | Stuttering progress percentage/log, high IPC overhead | Throttle emission to a fixed cadence (~150-250ms) regardless of walk speed | Any volume fast/large enough to visit hundreds of files per second |
| Dynamic-measuring virtualizer for provably-fixed-height rows | Layout thrash, measurement races with density toggle | Use fixed-size windowing keyed on the `--rh` constant | Any catalog large enough that virtualization matters at all (40k+ target) |
| Un-keyed / index-keyed virtualized rows across expand/collapse | Flicker, wrong-row content for a frame, lost scroll anchor | Stable per-node id independent of current visibility, used as the row key | Large trees with frequent expand/collapse, worst at 40k+ |
| Un-debounced fsnotify → full `BrowseCatalogs` re-scan per event | Rail flicker, wasted CPU on directories with unrelated churn | Debounce bursts (~200-500ms), filter to relevant extensions | Directories with sync-tool activity or StorCat's own writes triggering self-events |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Silent fallback from OS Trash to `os.Remove` on failure | User believes a destructive action is recoverable when it is not — data loss with no warning | Never fall back silently; surface the trash error through the `{success:false, error}` envelope |
| Non-atomic overwrite of existing catalog JSON on re-scan | A crash mid-write corrupts a previously good catalog with no recovery path | Temp-file-in-same-dir + `os.Rename` for every write that can overwrite existing data |
| Unescaped user-provided title written into HTML `<title>` on rename | HTML injection into a locally-generated file is low severity here, but still corrupts the file/breaks parsing on special characters | Use `html/template` escaping or a real HTML parser for the rewrite, not string replace |
| Exposing raw parse errors (`json.SyntaxError.Offset`, byte content) in the "unreadable catalog" state without bounding message size | A very large/malformed JSON file could produce an unwieldy error dump in the UI (denial-of-usability, not security per se, but worth bounding) | Truncate/summarize the raw parse error surfaced to the UI rather than dumping arbitrarily large content |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-------------------|
| Focus escaping a modal/palette/slide-over into the tree behind it | Keyboard users lose track of where input goes; feels broken even if visually fine | Shared focus-trap hook applied to all four overlay types from the first one built |
| Background scroll/layout shift when an overlay opens (missing scrollbar-width compensation) | Visible 1-2px jump every time Settings/palette/create opens — reads as "janky" even though functionally correct | Reimplement antd's scroll-lock-with-compensation behavior explicitly, don't assume `overflow:hidden` on `body` alone is equivalent |
| "Expand all" freezing the UI momentarily on very large catalogs | Feels like the app hung; worst first impression at exactly the scale (40k+ nodes) the feature is meant to handle | Operate on the pre-flattened array (Pitfall 5), verify against a synthetic large catalog, not just dev's own small one |
| Search-hit selection landing on the wrong row after ancestor expansion | User loses trust in ⌘K as a reliable "find the file" tool | Sequence expand-then-scroll correctly across the async state boundary (Pitfall 8) |
| Slide-over closing instantly instead of animating out | Small thing, but visibly cheapens the "pro tool" feel the whole redesign is going for, and it's the one interaction explicitly called out in the spec | `createClosing` + timed unmount, single close path (Pitfalls 9-11) |

## "Looks Done But Isn't" Checklist

- [ ] **Virtualized tree:** Renders fine with the developer's own small test directory — verify separately against a synthetic 40k+ node catalog before considering Phase 2 done.
- [ ] **Theme switching:** Looks correct on the default theme (StorCat Dark) — cycle through all 11 themes and check every accent-filled surface (primary buttons, selected rows, chips, badges) for contrast, not just the panel backgrounds.
- [ ] **Create slide-over close:** Clicking × once looks fine — test all five close triggers (Escape, ×, Cancel, scrim, "Open in workspace") individually, and test rapid re-open during the exit animation.
- [ ] **Scan cancellation:** "Cancel" button closes the UI — verify the Go-side goroutine actually stops (no continued disk activity, no orphaned output file) after cancelling.
- [ ] **Delete to Trash:** File visibly disappears from the rail — verify it's actually recoverable from the OS Trash/Recycle Bin, and verify what happens when the trash operation is deliberately made to fail.
- [ ] **fsnotify watch:** Status bar shows "● watching" once at startup — verify it correctly reflects reality after changing the catalog directory in Settings, and after simulating a watch failure (unplug removable media if that's ever a valid target).
- [ ] **CLI parity:** GUI features ship and work — re-run `cli/*_test.go` and manually exercise all 6 subcommands after any change to `internal/catalog`/`internal/search`, not just at milestone end.
- [ ] **Cross-platform visual check:** Looks correct on the macOS dev machine — verify the fixed-width rail/details panes, scrollbar behavior, and font rendering on an actual Windows build before considering any visual-parity phase done.
- [ ] **Backward-compat regression:** New create path produces a catalog that "looks right" in the app — diff its JSON schema (top-level keys) against a pre-milestone catalog to confirm no accidental new required fields.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|----------------|------------------|
| Stacking-order bug (Pitfall 1) shipped | LOW | Centralize into the tokens module retroactively; audit every `z-index` usage in one pass |
| Non-atomic write already corrupted a user's catalog (Pitfall 18) | MEDIUM | Cannot recover the corrupted file itself, but the fix (temp+rename) prevents recurrence; consider a "repair/re-scan" prompt if `LoadCatalog` detects a truncated/invalid JSON on open |
| Trash silently fell back to permanent delete (Pitfall 19) already shipped | HIGH | Data is genuinely unrecoverable for affected users; fix the fallback immediately and treat as a release-blocking regression if discovered pre-ship |
| Schema drift from an unconditional new field (Pitfall 27/28) already shipped catalogs with the extra field | MEDIUM | Field is likely harmless if `omitempty`-able going forward, but requires an explicit compatibility check in `LoadCatalog` to ignore unknown fields gracefully (already true for JSON unmarshal into a defined struct, but verify no downstream code chokes on the extra key) |
| CLI signature break (Pitfall 29) already merged | LOW-MEDIUM | Compile failure would be caught immediately by CI/tests if `cli/*_test.go` is run; behavioral-only breaks (no compile error) are the higher-cost case — add the missing test coverage retroactively |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|-------------------|----------------|
| 1. Stacking-order regression | Phase 1 | Centralized z-index tokens; narrow-window visual check repeated at Phases 4, 5, 6 |
| 2. Missing focus trap / Escape / scroll lock | Phase 3 (first custom overlay) | Tab-key and Escape-key testing on every overlay type, not just click testing |
| 3. Hand-rolled controls losing keyboard/ARIA | Phase 4, 5 | Every toggle/segmented control built on a real `<button>`/`<input>` under styling |
| 4. Key instability across expand/collapse | Phase 2 | Row key derived from stable node id, not visible-list position |
| 5. "Expand all" freeze at scale | Phase 2 | Test against a synthetic 40k+ node catalog |
| 6. Scroll position leaking across catalog switch | Phase 2 | Switch catalogs of very different sizes, confirm no blank pane |
| 7. Fixed-height rows implemented as dynamic-measuring | Phase 2 | Windowing keyed on `--rh` constant, no runtime row measurement |
| 8. ⌘K expand-then-scroll race | Phase 3 | Search hits at varying tree depths land exactly on target row |
| 9. Slide-over unmounts before exit animation | Phase 4 | `createClosing` + timed unmount visually confirmed (260ms slide, not instant) |
| 10. Divergent close paths | Phase 4 | Single `closeCreatePanel()`; grep confirms no other state-setter |
| 11. Re-open during exit animation double-fires timer | Phase 4 | Rapid open/close/open test within 260ms window |
| 12. `--wails-draggable` swallowing toolbar clicks | Phase 1, re-check 4-6 | Every new toolbar-region control gets `no-drag`, tested on Windows |
| 13. EventsOn listener leak under StrictMode | Phase 4 | Open/close create panel repeatedly in StrictMode dev build, confirm no duplicate progress lines |
| 14. Progress events flooding the bridge | Phase 4 | Throttled ticker cadence confirmed against a large/fast volume scan |
| 15. No real cancellation | Phase 4 | `context.Context` threaded through walk; Cancel verified to stop disk activity |
| 16. Goroutine leak on app quit | Phase 4 | `beforeClose` cancels in-flight scans; force-quit-mid-scan test |
| 17. Partial-catalog error path missing | Phase 4 | Simulated volume-disappearance test produces partial catalog + correct error UI |
| 18. Non-atomic catalog write | Phase 4 (create), Phase 6 (overwrite) | Kill process mid-write test; confirm no corruption of prior catalog |
| 19. Trash silent fallback to permanent delete | Phase 6 | Simulated trash-failure test confirms error surfaced, no `os.Remove` fallback |
| 20. Watcher not cleaned up / recursive watch | Phase 5/6 | Directory-change and app-quit both close prior watcher; watch confirmed non-recursive |
| 21. fsnotify event storms / self-trigger loops | Phase 5/6 | Debounce verified under rapid unrelated file churn |
| 22. Watch reliability silently degrading | Phase 5/6 | `Errors` channel handled; status bar reflects actual watch health |
| 23. `--onac` not luminance-derived | Phase 1 | All 11 themes checked for accent-fill text contrast |
| 24. `color-mix()` unsupported on old webkit2gtk | Phase 1 | Tokens computed in TypeScript, not shipped as raw CSS `color-mix()` |
| 25. Missing CSS reset after antd removal | Phase 1 | Pixel-comparison against the `.dc.html` prototype for buttons/inputs/rows |
| 26. Scrollbar width shifting fixed-width panes | Phase 1 (rule), Phase 2 (verify) | `scrollbar-gutter: stable` applied; visual check on an actual Windows build |
| 27. Sidecar cache leaking into catalog JSON | Phase 2 | JSON schema diff of a fresh catalog against pre-milestone output |
| 28. Partial-catalog markers changing clean-scan JSON shape | Phase 4 | `omitempty` verified; clean-scan JSON shape test |
| 29. CLI ↔ shared package signature breaks | Phase 4, Phase 6 | `cli/*_test.go` re-run after every `internal/catalog`/`internal/search` change |
| 30. Rename `<title>` rewrite corrupting HTML | Phase 6 | Rename tested with `<`, `&`, `"`, em-dash, non-ASCII titles |

## Sources

- `design_handoff_storcat_ui/README.md` — authoritative design spec (read directly; Interactions & behavior, Scale, Stacking order, Suggested build order sections)
- `.planning/PROJECT.md` — milestone decisions (sidecar cache, Ant Design removal, GUI-only new capability, CLI parity constraint)
- `.planning/RETROSPECTIVE.md` — reviewed to avoid repeating already-internalized lessons (SUMMARY.md template gaps, Nyquist validation discipline, credential/ops separation — none of which are repeated in this document)
- `app.go`, `internal/catalog/service.go`, `go.mod` — read directly to confirm current function signatures (`CreateCatalog` has no `context.Context` today; `traverseDirectory` already has a `ProgressCallback` parameter to extend; `beforeClose` hook exists in `app.go`)
- Wails `EventsOn`/StrictMode cleanup — GitHub issue wailsapp/wails#3796 (MEDIUM confidence, web-sourced, cross-checked)
- Wails frameless drag region behavior — GitHub issues wailsapp/wails#3971, #1861, #5547; wails.io Frameless Applications guide (MEDIUM confidence, web-sourced)
- Go cross-platform trash libraries (go-trash, trash-go: SHFileOperationW on Windows, FreeDesktop.org Trash spec on Linux) (MEDIUM confidence, web-sourced)
- fsnotify editor-rename behavior and Linux inotify `max_user_watches` default of 8192 per UID — fsnotify GitHub issues/docs, watchexec inotify-limits doc (MEDIUM confidence, web-sourced)
- CSS `color-mix()` browser support (Safari 15 WebKit baseline, full support Safari 18/Chromium 125; WebView2 inherits Chromium) — caniuse/MDN browser-compat-data discussion (MEDIUM confidence, web-sourced)
- Project's own conventions (`CLAUDE.md`): "Silent Fallbacks" principle (`or {}` converting hard failures into silent corruption) applied directly to the Trash-fallback and partial-catalog pitfalls

---
*Pitfalls research for: StorCat v3.0.0 Workspace Redesign — full custom-UI replacement + new backend capability in an existing Go/Wails app*
*Researched: 2026-08-13*
