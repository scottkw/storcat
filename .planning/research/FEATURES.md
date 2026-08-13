# Feature Research

**Domain:** Desktop media-cataloging tool — pro-tool workspace UI (single-pane file-manager/IDE pattern) over a Go/Wails catalog backend
**Researched:** 2026-08-13
**Confidence:** MEDIUM (cross-verified against multiple independent, well-documented comparable tools — VS Code, rclone/rsync, FreeFileSync, Dropbox, Time Machine/Backblaze, Finder/Disk Utility, macOS/Nemo trash UX; no official StorCat-specific precedent exists since this is greenfield for the app)

This file researches **how the ten designed behaviors in `design_handoff_storcat_ui/README.md` work in comparable tools**, not alternative features — the handoff is final and authoritative. Every row below traces to a specific handoff section (cited inline) and is grounded in what comparable desktop tools do.

---

## Feature Deep Dives

### 1. Single-view workspace (rail + tree + details, no tabs)
**Handoff:** §"1. Workspace (the whole app)", §"Suggested build order" step 1–2

This is the classic **IDE project-pane / file-manager three-pane layout** — VS Code's Explorer (left) + editor (center) + a details/outline surface, or Finder's sidebar + list + Get-Info/Preview pane, or a DAM tool's folder tree + content grid + metadata panel (Adobe Bridge's Folders/Content/Metadata panel arrangement is the closest analog: a navigation tree, a content area, and a details panel that follows selection). What makes the pattern work in all of these:
- **One selection state drives everything** — clicking a tree node updates the details panel and (in StorCat's case) the status bar; there is no separate "which tab am I in" state to desync from. The handoff's `curId`/`selected`/`expanded` state block is exactly this — a single source of truth, not three tab-scoped state trees like the current v2.3.0 UI.
- **Independent scroll regions** — rail, tree, and details each scroll on their own (`flex: 1; overflow-y: auto; min-height: 0` per the handoff). This is standard in VS Code and Finder; the common bug is a parent container without `min-height: 0` that prevents inner flex children from scrolling at all.
- **Virtualization at scale** — VS Code's own Explorer virtualizes deep trees; without it, a 40k-node tree (explicitly called out in the handoff's *Scale* section) will jank on expand/collapse and select. This is the single highest-complexity item in the whole redesign.
- **Responsive collapse to drawers** — IDEs and Finder-like tools drop side panels into overlay drawers below a width threshold rather than reflowing to unusable narrow columns; the handoff's three-tier responsive table (≥1280 / 1040–1279 / <1040) mirrors this.

**What breaks it (comparable-tool failure modes to avoid):**
- Not virtualizing the tree — the #1 way these layouts die at scale.
- Z-index/stacking bugs where a drawer panel outranks a modal or slide-over — the handoff explicitly flags this risk ("The details panel outranking the slide-over is an easy bug to reintroduce").
- Selection state that isn't *derived* — hardcoding "select node 0" instead of deriving initial selection from data (handoff: "Initial selection is derived, never hardcoded: find the node, expand its ancestors").

**Category:** Table stakes (for the "pro tool" positioning the redesign is aiming at — this is the baseline layout convention of every comparable tool: VS Code, Sublime Merge, Adobe Bridge, Finder).
**Complexity:** HIGH — this is a full frontend rewrite (Ant Design removed entirely per PROJECT.md), the tree must be flattened + virtualized, and three independently-responsive panes with swappable order must be built from scratch.
**Depends on:** Existing `BrowseCatalogs` / `LoadCatalog` Go bindings (already exist, per handoff's Mocked-functionality table — "Exists"). Nothing else in the redesign is buildable before this — every other feature (palette, create slide-over, settings, actions menu, empty/error states) mounts inside or over this shell.

---

### 2. Command-palette search (⌘K), 50-result cap
**Handoff:** §"2. Search palette (⌘K)", §"Interactions & behavior" table, §"Scale"

⌘K/Ctrl+K command-and-search palettes are now a baseline expectation in 2026 professional tools — VS Code (`Cmd+P`/`Cmd+Shift+P`), Linear, Notion, Raycast, Spotlight, GitHub's own command palette. Common conventions, all of which the handoff already specifies:
- **Autofocus on open** — the input is focused the instant the overlay mounts; users never click before typing. Handoff: "Autofocus on open."
- **Keyboard-first, mouse-optional** — full up/down navigation and Enter-to-select without ever touching the mouse; Escape always closes. The handoff's footer hints (`↵ reveal in catalog`, `esc close`) match this convention exactly.
- **Live, debounced-feeling results, no loading spinner** — matches the handoff's global "no spinners: progress is always a real number" principle; palettes in this class filter as-you-type against an already-loaded/fast index rather than showing a spinner per keystroke.
- **A visible cap with an explicit "showing N of M" affordance** rather than silently truncating — this is what keeps a capped result list from feeling broken. VS Code's Quick Open and GitHub's palette both cap and communicate a "truncated results" state rather than paginate. The handoff's cap notice ("Showing the first 50 of 3,482 hits — refine the term, or ↵ to open the full result table") is exactly this pattern, and is good practice: it tells the user *why* they're not seeing everything and gives a next action (refine, not "load more").
- **Enter behavior on a capped list**: comparable tools bind Enter to the *highlighted* result, not "run the query" — in StorCat's design, Enter opens the highlighted hit (or, per copy, offers the full table when capped). This needs an explicit keyboard state machine: arrow keys move a highlighted index, Enter acts on it, typing resets highlight to index 0.
- **Hit navigation = switch context + expand ancestors + select + close** — this is the same "reveal in tree" behavior VS Code's "Reveal in Explorer" and Finder's "Reveal in enclosing folder" implement: don't just show the result, put the user *in* the surrounding structure. The handoff reuses the same expand-ancestors logic the rail-click interaction needs (§1), so this is a shared primitive, not a one-off.

**Category:** Table stakes in 2026 for any tool with a large searchable corpus and pro-tool ambitions — its absence would read as a regression from what users expect of modern desktop software, independent of what v2.3.0 had.
**Complexity:** MEDIUM. The search itself already exists (`internal/search`, per Mocked-functionality table — "Exists," reuse behind the palette). New work is: (a) capping + returning a total count (currently `SearchCatalogs` returns everything — handoff: "Partly"), (b) the keyboard-nav state machine, (c) the ancestor-expand-and-select action, which should be implemented once and shared with the rail.
**Depends on:** #1 (workspace shell + expand/select primitives), existing `internal/search`. Independent of scan/volume/diff features — can ship in parallel with or right after the shell (handoff build order step 3).

---

### 3. Live scan progress (%, files, bytes, walking path, log, ETA, background hand-off)
**Handoff:** §"Step 2 — scanning", §"Interactions & behavior" (`~220ms ticks, +2–9% each`), §"Mocked functionality" row 1

This is the UX of a **long-running file-walk operation**, the same class of problem as `rclone --progress` / `rsync -P`, Time Machine's backup progress, or Backblaze's upload status. Conventions observed across these tools:
- **A real percentage plus supporting counters, not just a spinner.** `rclone -P`-style output shows bytes transferred, percent, throughput, ETA, and a file counter on one line; StorCat's design mirrors this almost field-for-field (percentage, files seen, bytes, "about 4s left"). This also matches the handoff's explicit design principle: "No spinners: progress is always a real number."
- **A visible "what is it doing right now" line.** rclone/rsync tools show the current file being transferred; StorCat's "WALKING /current/absolute/path" with a pulse animation is the cataloging equivalent — it reassures the user the process is alive even when the percentage stalls (e.g., one huge directory).
- **A rolling log capped to a small window**, newest-first, not an ever-growing scrollback — comparable to `rsyncy`'s condensed status view rather than raw rsync's unbounded verbose log. The handoff's log box (max 9 lines, newest-first `+ /path`) matches this "just enough context, bounded" convention.
- **Throttled update rate.** Emitting a UI event per file at real-world scan speeds would flood the Wails IPC bridge; the handoff's own cadence (~220ms ticks) implies coalescing/batching progress events rather than emitting one per file — this must be implemented in the Go walk (batch N files or T milliseconds per `EventsEmit`), not left to the frontend to throttle.
- **"Run in background" hands off to a persistent, small status indicator**, not a second progress screen — this is exactly what Backblaze/Time Machine do with menu-bar/system-tray progress once you dismiss their main window, and what the handoff specifies with the status-bar `● scanning sd48 · 68%` replacing the "watching" indicator.

**Category:** Table stakes for this specific product (a tool whose core action is "scan an entire SD card," which can take real time) — comparable tools in this class (backup/sync software) never ship a blocking-call-with-no-feedback UX; it reads as broken/frozen otherwise. Differentiator relative to StorCat's *own* history, since v2.3.0 has none of this (`CreateCatalog` is currently a single blocking call).
**Complexity:** HIGH. Requires new Go-side event emission (`EventsEmit`) threaded through the directory walk, a throttling/batching strategy, ETA estimation (simple rate-based projection is sufficient — comparable tools don't do anything fancier for this class of operation), and frontend event-subscription plumbing for the "run in background" hand-off to keep receiving events after the slide-over closes.
**Depends on:** The error-tolerant walk needed by #5 (partial catalog) shares the same walk loop — these two should be built together, not sequentially, since the error path is a state the progress UI must also render (see "Step 2 error state"). Also depends on #4 (volume detection) supplying the "read errors" flag that triggers the error path at ~54–57%.

---

### 4. Volume detection and selection
**Handoff:** §"Step 1 — form" ("SOURCE VOLUME"), §"Mocked functionality" row 2

Enumerating mounted external media with capacity and a pre-flight health signal is standard in disk-utility-class tools — macOS Disk Utility and Finder's sidebar both show mounted volumes with capacity, and gray out/flag volumes that are detected but not cleanly mounted. StorCat's card-per-volume UI (name, mount path, size, `mounted`/`read errors` tag) is the same information architecture, simplified to a picker.

Key expected behaviors from comparable tools:
- **Enumeration must be OS-native**, not a single cross-platform API — Disk Utility-class tools always special-case per OS (macOS `/Volumes`, Windows logical drives, Linux `/media/$USER` + `/run/media`), which is exactly what the handoff prescribes.
- **A cheap pre-flight health check, not a full scan**, to populate the `read errors` tag — comparable tools (Disk Utility, `diskutil info`) use a fast stat/metadata pass, not a full read, to flag failing media before committing to a long operation. This is explicitly a "cheap stat pass or the previous failed scan," per the handoff.
- **Capacity display matches Finder/Disk Utility conventions** — total size shown per volume, decoupled from percentage-full (StorCat doesn't need free-space math since it's cataloging, not managing space).
- **"...or choose any folder" fallback** — mirrors Finder/Disk Utility's ability to also operate on an arbitrary path, not just recognized volumes; keeps the existing `SelectDirectory` picker relevant rather than replaced.

**Category:** Table stakes — a tool whose entire purpose is "catalog my SD card" needs to show what cards are present; forcing users to always manually browse to `/Volumes/SD12` would be a regression from what any comparable media tool (photo importers, disk utilities) offers.
**Complexity:** MEDIUM — three OS-specific enumeration paths plus a lightweight stat-based health check; no new UI complexity beyond the card list already specified.
**Depends on:** Nothing upstream; this is an independent Go-side addition (`internal/volumes` or similar) that can be built and tested standalone, then wired into the existing create-flow Step 1. It *feeds* #3/#5 (the `read errors` tag is what triggers the scan error path).

---

### 5. Partial catalog on scan failure
**Handoff:** §"Step 2 error state", §"Mocked functionality" row 3

This is the cataloging-domain equivalent of **resumable/tolerant transfer tools** (rsync's partial-transfer mode, TeraCopy's skip-and-continue) rather than all-or-nothing backup tools. It's worth noting Time Machine and Backblaze take the *opposite* stance — Time Machine typically discards or requires manual removal of an incomplete backup rather than presenting a usable partial artifact, and Backblaze simply stops backing up a drive that goes missing rather than writing a partial snapshot. StorCat's design deliberately chooses the more forgiving, rsync-like posture: **write what was walked, mark what wasn't**, because a catalog of "everything except the folder that failed" is still useful (unlike a partial backup, which is not a substitute for a full one).

What "write partial catalog" should mean concretely, grounded in the handoff:
- The walk must become **error-tolerant** — today an error aborts the whole operation (handoff: "currently an error aborts"). This means catching per-entry read errors during traversal and continuing rather than propagating.
- The **unreadable subtree needs an explicit marker** in the catalog data, not silent omission — silently dropping the failed folder would make the catalog look complete when it isn't (a classic sync-tool trust failure). A marker also gives the "Unreadable catalog"/error UI something concrete to point at later.
- The **UI must clearly state what happened and where**: the design's error state ("Stopped at 57% — the volume went away," mount point, files walked, a red `read error: … — input/output error` log, and a plain-language explanation) mirrors how comparable tools report I/O failures — surfacing the OS-level error text (`input/output error`) rather than a generic "something went wrong," which is what actually helps a user diagnose a failing SD card.
- **Three explicit recovery actions** — Write partial catalog (primary), Retry scan, Cancel — is a reasonable, minimal action set; comparable tools (TeraCopy: skip/retry/abort per error) converge on the same three verbs for a failed-mid-operation state.

**Edge case (media disappearing mid-scan):** the design assumes the walk can detect "the volume went away" distinctly from "a file is unreadable" — these need different handling in Go: a lost mount point should stop the walk cleanly and preserve everything gathered so far (feeding "write partial catalog"), while an unreadable *file* within an otherwise-present volume should be skipped and logged without stopping the walk at all. Comparable tools (rsync, TeraCopy) distinguish these two cases; StorCat's error-tolerant walk should too, even though the handoff's UI presents one error state — the underlying Go logic needs the distinction so a single bad file doesn't unnecessarily trigger the "volume went away" narrative.

**Category:** Differentiator — this is exactly the pain point of the domain (SD cards fail, cards get pulled early, USB readers flake) and comparable general-purpose sync/backup tools mostly punt on it (discard-and-retry) rather than solve it (partial-with-marker). Getting this right is a meaningful edge over "catalog just crashed" (today's behavior).
**Complexity:** HIGH — requires rewriting the walk to be error-tolerant, designing the on-disk marker for an unreadable subtree (and keeping it compatible with the existing v1/v2 dual-format `CatalogItem` schema), and distinguishing "volume vanished" from "one bad file" at the Go level.
**Depends on:** #3 (live scan progress plumbing — the error state is a variant of the progress screen, sharing the same event stream and log box) and #4 (volume health flag that pre-signals the likely failure).

---

### 6. Re-scan & diff (added / removed / changed / unchanged)
**Handoff:** §"Re-scan & diff (620px)", §"Mocked functionality" row 7, §"Suggested build order" step 6 ("biggest backend piece")

This is directory-sync-tool territory — FreeFileSync and rsync are the closest comparables, and their conventions map directly onto the handoff's spec:

- **Comparison key: stable path, not content hash.** FreeFileSync and rsync both default to comparing by **relative path + size + modification time**, not a full content hash, because hashing every file on a large volume is prohibitively slow for what's usually a "did anything change" check — and StorCat's catalogs are explicitly sized for 40k+ node volumes (whole SD cards), where hashing would turn a few-second re-scan into a multi-minute one. The handoff's own phrasing — "needs stable path keys and a size comparison" — confirms this is the intended approach: **path is the identity key, size (and reasonably mtime) is the change signal.**
- **What "changed" means:** research on file-change detection is consistent that metadata-only comparison (size and/or mtime) is fast but imperfect — mtime can be stale/forged and size-identical content changes are missed — while hash comparison is authoritative but expensive. Given the handoff explicitly calls for "a size comparison" (not a hash comparison), StorCat should treat **size difference at the same path as definitively "changed,"** and can optionally also flag an mtime-only difference (same size, newer mtime) as "changed" to catch same-size edits — this is exactly the compromise FreeFileSync's default "file time and size" comparison mode makes, and it's appropriate here since catalogs aren't meant to be forensic-grade, they're meant to be useful.
- **Four-way classification (added/removed/changed/unchanged)** is the standard vocabulary of every directory-diff tool (FreeFileSync literally uses this categorization, git's status output uses the analogous added/deleted/modified/unmodified). The handoff's four stat tiles map onto it directly.
- **Overwrite vs. Keep-both** is FreeFileSync's and most versioned-backup tools' standard conflict resolution: either replace the existing artifact or write a new one alongside it. The handoff's "Keep both (write sd12-2026.json)" is the same pattern as suffixed backup filenames elsewhere in the app (Duplicate catalog, #7) — these two features should share one "generate a non-colliding filename root" helper.
- **Discard** (the third action) has no sync-tool analog needed — it's simply "don't act on this diff," a safe no-op.

**Edge cases to flag for planning:**
- **The volume that produced the original catalog may no longer be the volume at that mount point** (a different SD card inserted into the same slot). Comparable sync tools don't generally guard against this (FreeFileSync just diffs whatever's at the path); StorCat likely shouldn't over-engineer a "is this the same card" check either — treat it as a normal (probably enormous) diff and let the numbers speak.
- **Media disappearing mid-re-scan** is the same failure mode as #5 and should reuse the same error-tolerant walk and progress UI, not a separate code path.
- **40k-node trees**: the diff list itself needs the same virtualization treatment as the tree pane (§1) — a diff of thousands of entries rendered naively will repeat the tree's performance problem.

**Category:** Differentiator (this is the single feature the handoff itself calls out as the largest undertaking, and it's the feature that turns StorCat from "one-shot cataloging" into a tool for tracking a card's contents over time — genuinely distinct value versus a one-off `tree` dump).
**Complexity:** HIGH — the largest backend lift in the milestone per the handoff's own ordering ("re-scan & diff last (biggest backend piece)"). Requires the error-tolerant walk (shared with #5), a diff algorithm keyed on stable relative paths, and two write paths (overwrite in place, or write a suffixed sibling).
**Depends on:** #3/#5 (walk infrastructure), #4 (re-locating the source volume for the catalog being re-scanned — "Re-scan volume…" in the details-panel footer implies the app must remember or re-resolve which volume a catalog came from), and the loaded `CatalogItem` tree (`LoadCatalog`, already exists) as the diff baseline. Should be sequenced **last** among the backend-heavy features, per the handoff's own build order.

---

### 7. Rename / Duplicate / Delete-to-Trash of a catalog
**Handoff:** §"5. Catalog actions (from `⋯`)", §"Mocked functionality" rows 4–6

These are file-manager-safety conventions, not sync-tool conventions:
- **Delete → OS Trash, never `os.Remove`.** This is the universal desktop-OS convention (macOS Finder, Windows Recycle Bin, Linux desktop environments' trash spec) precisely because destructive file operations need a recoverable undo path. Nemo/Caja (Linux file managers) even expose this as a *toggleable* confirmation because trashing is already considered safe-by-default in most of these tools — a single, clear confirmation dialog (not a "type to confirm" gate) is the right level of friction for a reversible action, which matches the handoff's single 480px confirm modal with one destructive button, not a multi-step confirmation.
- **Explicit disambiguation copy is the important UX detail here**, more than the mechanics: the handoff is explicit that the copy must state "it deletes catalog files, not the card" — this is because the action's *label* ("Delete catalog") is ambiguous about scope (does it touch the SD card? just the JSON/HTML?) in a way generic file-delete dialogs aren't. This is a domain-specific trust signal, not a generic pattern, and should not be cut for brevity.
- **A toggle for the paired `.html` file** (delete JSON only vs. JSON+HTML) matches how comparable tools handle "sidecar file" deletion (e.g., RAW+JPEG pairs in photo tools) — give the user the choice rather than silently deleting or silently preserving the second file.
- **Rename only touches the display title (`<title>` in the `.html`), not the filenames** — this is a deliberate and slightly unusual choice (most "rename" affordances rename the file), so the explanatory copy the handoff specifies ("explains the title lives in the `.html` `<title>` and filenames don't change") is again the load-bearing UX detail, mirroring how the disambiguation copy works for delete.
- **Duplicate** is a straightforward copy-with-suffixed-filename-root operation, the same pattern needed for #6's "Keep both."

**Category:** Table stakes — every comparable creative/asset tool (and every OS file manager) offers rename/duplicate/delete on generated artifacts; their complete absence today (all three rows marked "Missing" in the handoff) is the gap, not the ambition.
**Complexity:** LOW–MEDIUM. Trash requires a cross-platform "move to OS trash" mechanism (Go doesn't have this in stdlib; a small platform-specific helper or a maintained library is needed — this is the one piece of real engineering here). Rename is a targeted `<title>` string replace in the `.html` file, care needed not to corrupt `BrowseCatalogs`' title parsing. Duplicate is a file copy plus the same filename-suffix helper #6 needs.
**Depends on:** Rail refresh after any of the three (list must re-run `BrowseCatalogs`). Duplicate/Keep-both share a naming helper with #6. No dependency on the scan/diff/watch features.

---

### 8. Directory watching → live "watching ~/dev/sd-catalogs" status
**Handoff:** §"Status bar", §"Mocked functionality" row 8

This is the same pattern as cloud-sync-client status indicators (Dropbox's menu-bar icon and in-window status line cycling through states like "Indexing," "Syncing," "Up to date"). StorCat's status bar text (`● watching ~/dev/sd-catalogs`, replaced by `● scanning sd48 · 68%` during a background scan) is a simplified, single-line version of the same idea — one indicator, a small enumerated set of states, always visible rather than tucked in a menu.

Expected behavior grounded in the fsnotify-class implementation the handoff specifies:
- The watch should be **debounced**, not raw-event-driven — a directory receiving several rapid filesystem events (a scan finishing and writing two files) should trigger one `BrowseCatalogs` refresh, not several. Comparable sync clients coalesce rapid-fire fs events for exactly this reason.
- **This directly answers the "catalog deleted outside the app" edge case**: with `fsnotify` active, a catalog file removed via Finder/Explorer (not through StorCat's own Delete action) should disappear from the rail automatically on the next debounced refresh, the same way Dropbox's own folder reflects external changes. Without the watch, StorCat would need a manual refresh action to avoid showing stale/ghost rail entries — worth confirming the watch is treated as the *primary* mechanism for this, with no separate "Refresh" button implied anywhere in the handoff.

**Category:** Differentiator — nice-to-have polish that most comparable single-purpose cataloging tools skip (a simple "click to browse" tool wouldn't bother), but it's cheap relative to its payoff here since it also solves the "deleted outside the app" staleness problem for free.
**Complexity:** MEDIUM — `fsnotify` wiring plus debounce logic in Go, and one more status-bar text state in the frontend. Low frontend complexity, moderate backend plumbing (must not fight with the in-app rail refresh triggered by Rename/Duplicate/Delete).
**Depends on:** #1 (status bar exists), existing `BrowseCatalogs`. Independent of scan/diff work; can be built any time after the shell.

---

### 9. Density, rail-position preferences, per-theme card pickers
**Handoff:** §"4. Settings (modal, 660px…)" — THEME / LAYOUT / CATALOGS sections

Configurable density (compact/comfortable row heights) and panel-position swapping are standard "pro tool" settings surfaces — VS Code (panel position, zoom/density-adjacent settings), Linear (list density toggle), and most DAM/asset tools (thumbnail/row density) all expose exactly this pair of controls via a segmented control, which is what the handoff specifies (3px-padding segmented track, active segment filled). A visual theme-card grid (swatch strip + name + light/dark tag) rather than a plain dropdown is likewise standard in tools with more than a handful of themes — it lets users recognize a theme by its colors rather than reading names, which matters once you're choosing among 11 options.

**Category:** Table stakes for the "pro tool" positioning (a redesign explicitly targeting "pro tool density" per the handoff's Fidelity note should not ship without the settings that let users tune that density) — but low-stakes if deferred one phase, since defaults (Comfortable, rail-left) are reasonable out of the box.
**Complexity:** LOW. Mostly CSS variable-driven (`--rh`, `--rp`, etc., already defined as design tokens) plus a handful of new persisted fields in `internal/config` (density, rail side, default filename root, and the four catalog-behavior toggles). Theme card grid is a straightforward re-skin of the existing 11-theme data in `themes.ts`, extended with the `p2`/`ch`/explicit-accent fields the handoff calls for.
**Depends on:** #1 (settings modal mounts over the shell; density/rail-side values are consumed by the shell's CSS). No dependency on backend scan/diff/watch work.

---

### 10. Empty-library and unreadable-catalog states
**Handoff:** §"6. Empty and unreadable states", §"Mocked functionality" row 9

Good empty and error states in comparable pro tools share the same shape: a short explanation of *why* the view is empty/broken, and one or two concrete next actions — never a bare blank pane. VS Code's empty Explorer ("You have not yet opened a folder"), Finder's empty-trash/empty-folder states, and any DAM tool's "no assets imported yet" screen all follow this. StorCat's design matches this convention directly:
- **Empty library**: explanation + two clear CTAs ("Catalog a volume", "Choose catalog folder…") — mirrors the "empty project, here's how to start" pattern rather than showing nothing.
- **Unreadable catalog**: this is the more distinctive, domain-specific piece — surfacing the *actual parse error* (file, byte offset, reason, parser, raw error text) rather than a generic "couldn't load" message. This is good practice borrowed from developer-facing error surfaces (compiler/linter error output with a byte/line offset) applied to an end-user screen — it's more technical than most consumer empty-states, which is appropriate since the audience already sees mono paths and byte counts everywhere else in this design.
- Both screens give an explicit way out that doesn't dead-end the user: "Re-scan volume" / "Open the .html instead" / "Remove from library" for the broken case, "Catalog a volume" / "Choose folder" for empty.

**Category:** Table stakes — every well-regarded comparable tool treats empty/error states as first-class screens, not afterthoughts; shipping the workspace without them would leave a real, hittable dead-end (a fresh install, or a catalog corrupted by an interrupted write) with no explanation.
**Complexity:** LOW for the empty state (pure UI). LOW–MEDIUM for the unreadable state's UI, but it depends on a **backend prerequisite that's currently the wrong shape**: parse failures are silently swallowed today (`continue` in `BrowseCatalogs`/`searchInCatalogFile`, per the handoff's Mocked-functionality table, "Partly" done) — that must become per-file errors that surface `json.SyntaxError.Offset` before this screen has anything real to show.
**Depends on:** The backend error-surfacing change above (shared prerequisite with nothing else in this list — it's self-contained), and #1 for where these states render (replacing the tree pane / rail block).

---

## Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete relative to comparable pro tools.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Single-view workspace (rail+tree+details) | Baseline layout of every comparable IDE/file-manager/DAM tool | HIGH | Full frontend rewrite; virtualization is the hard part (40k+ nodes) |
| ⌘K command-palette search with cap + count | Standard in 2026 pro tools (VS Code, Linear, Raycast, GitHub) | MEDIUM | Search backend exists; needs cap+total, keyboard nav, ancestor-expand |
| Volume detection & selection | Standard in disk-utility-class tools (Finder, Disk Utility) | MEDIUM | OS-specific enumeration + cheap health stat pass |
| Rename / Duplicate / Delete-to-Trash | Universal desktop file-manager safety convention | LOW–MEDIUM | Trash needs a cross-platform helper; explanatory copy is load-bearing |
| Empty-library & unreadable-catalog states | Every comparable pro tool treats these as first-class screens | LOW–MEDIUM | Unreadable state blocked on backend error-surfacing prerequisite |
| Live scan progress (%, bytes, path, ETA) | Any tool with a long-running walk shows real progress (rclone/rsync/Time Machine convention) | HIGH | Needs throttled `EventsEmit` stream from Go |
| Density & rail-position settings | Standard "pro tool" configurability (VS Code, Linear, DAM tools) | LOW | Mostly CSS tokens + a few config fields |

## Differentiators (Competitive Advantage)

Features that set the product apart from a generic cataloging tool. Not required by the layout convention alone, but where StorCat gains real value over the status quo (v2.3.0) and over generic sync/backup tools.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Partial catalog on scan failure (error-tolerant walk + marker) | Solves the domain's real pain point (failing SD cards) better than general sync/backup tools, which mostly discard-and-retry | HIGH | Needs distinct "volume vanished" vs. "one bad file" handling in Go |
| Re-scan & diff (added/removed/changed/unchanged) | Turns StorCat from one-shot cataloging into tracking a card over time | HIGH | Largest backend lift in the milestone per handoff's own build order |
| Directory watching (live "watching …" status) | Also solves "catalog deleted outside the app" staleness for free | MEDIUM | Debounce fs events; coalesce with in-app refresh triggers |
| Themed card picker with derived contrast (`--onac` luminance helper) | Lets 11 themes (light and dark accents) stay legible without per-theme overrides | LOW | Explicitly a ported 6-line helper per the handoff, not novel design |

## Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but would create problems or scope creep beyond what the handoff specifies. None of these appear in the handoff — flagged so they don't get added during planning.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|------------------|-------------|
| Full-text file-content search/indexing | "Search everything" feels more powerful than name/path search | Catalogs are trees of arbitrary (often binary) files on SD cards — content indexing is enormous scope and speed cost for a use case that doesn't need it | Handoff explicitly scopes search to "searches names and paths" — keep it there |
| Continuous two-way sync between catalog and volume | Feels "smarter" than manual re-scan | Fights the point-in-time-catalog model; risks acting on a volume unexpectedly (deletion/overwrite) when a card is swapped in the same slot | On-demand re-scan & diff (#6), user-triggered only |
| Content-hash (SHA-256) diffing for re-scan | Feels more "correct" than size/mtime comparison | Hashing every file on a 40k+-node volume turns a fast re-scan into a slow one; the handoff explicitly calls for a size comparison, not a hash | Path + size (+ mtime) comparison, matching rsync/FreeFileSync's default mode |
| Multi-scan queue / background scan manager | Seems like a natural extension of "run in background" | The handoff's status bar has exactly one scan slot (`● scanning sd48 · 68%`) — no queue UI is designed | Single background scan at a time, as specified |
| "Type to confirm" delete gate | Feels extra-safe for a destructive action | The action is reversible (Trash) and the handoff already invests in explanatory copy instead of extra friction; a typed-confirmation gate would be nagging, not safety, for a reversible action | Single confirm dialog with clear copy (as designed) |

## Feature Dependencies

```
[Single-view workspace shell]  (existing: BrowseCatalogs, LoadCatalog)
    ├──feeds selection/expand primitives──> [⌘K command palette]        (existing: internal/search)
    ├──hosts──> [Settings modal: density / rail position / theme cards]
    ├──hosts──> [Empty-library state]
    └──hosts──> [Unreadable-catalog state]
                     └──requires (backend)──> [Malformed-catalog error surfacing
                                                (per-file errors, json.SyntaxError.Offset)]

[Volume detection & selection]
    └──feeds "read errors" flag──> [Live scan progress] ──shares walk loop──> [Partial catalog on scan failure]
                                                                                     └──shares progress/event plumbing with──> [Re-scan & diff]
                                                                                                                                    └──requires──> [LoadCatalog tree as diff baseline] (existing)
                                                                                                                                    └──shares "non-colliding filename" helper with──> [Duplicate catalog]

[Rename / Duplicate / Delete-to-Trash]
    └──requires──> [Rail refresh after mutation] (BrowseCatalogs re-run)
    └──Duplicate / Keep-both share──> [filename-suffix helper] (also used by Re-scan & diff)

[Directory watching]
    └──requires──> [Status bar] (workspace shell)
    └──enhances──> [Rail staleness handling] (catalogs deleted outside the app)
```

### Dependency Notes

- **Everything requires the workspace shell first.** The handoff's own build order puts "Shell: toolbar + 3-pane grid + status bar" as step 1 for exactly this reason — the palette, create slide-over, settings, and actions menu are all things that mount *inside or over* the shell, not standalone screens.
- **Live scan progress and partial-catalog-on-failure must be built together, not sequentially.** They share one Go walk loop and one event stream; the "error state" in Step 2 is a variant of the progress screen, not a separate feature with its own plumbing.
- **Re-scan & diff depends on the walk/progress infrastructure and is correctly sequenced last** among backend-heavy items — it needs the error-tolerant walk (from partial-catalog work) plus a new diff algorithm on top, so building it first would mean rebuilding the walk logic twice.
- **The unreadable-catalog empty state is blocked on a small but necessary backend fix** (stop swallowing parse errors in `BrowseCatalogs`/`searchInCatalogFile`) that has no other dependents — it can be done any time, but must land before that screen has real data to show.
- **Duplicate catalog, Keep-both (re-scan), and any future "avoid filename collision" need share one small helper** — build it once when the first of the three is implemented.
- **Directory watching both drives the status-bar indicator and quietly solves the "catalog deleted outside the app" edge case** — if watching is deferred, that edge case needs a fallback (e.g., a manual refresh action), since nothing else in the handoff currently covers it.

## MVP Definition

Structured around the handoff's own **Suggested build order** (§ near the end of the handoff), which is treated as authoritative sequencing, not just a research recommendation.

### Launch With (v1 — phases 1–3 of the handoff's build order)

- [ ] Workspace shell: toolbar, 3-pane grid, status bar, token layer, theme switching — the whole redesign is unshippable without this
- [ ] Rail (`BrowseCatalogs`) + virtualized tree (`LoadCatalog`) — handoff notes this slice is "Ship-able already" once virtualized
- [ ] ⌘K palette over `SearchCatalogs`, with cap + total — reuses existing search, moderate net-new work

### Add After Validation (v1.x — phases 4–5)

- [ ] Create slide-over: form + `CreateCatalog`, then live progress events, then the error/partial-catalog path — trigger: once the shell/rail/tree are stable, this is the next-most-visible gap (today's blocking `CreateCatalog` call)
- [ ] Volume detection & selection — trigger: needed to make the create-flow's Step 1 match the design rather than falling back to folder-picker-only
- [ ] Settings modal: theme cards, density, rail position, catalog defaults — trigger: once the shell's tokens exist, this is comparatively low-cost to add

### Future Consideration (v2+ — phase 6, explicitly called "biggest backend piece")

- [ ] `⋯` actions: Rename, Duplicate, Delete-to-Trash — lower complexity, could move earlier if sequencing allows, but the handoff groups them with re-scan & diff as the last build-order step
- [ ] Re-scan & diff (added/removed/changed/unchanged, overwrite vs. keep-both) — defer until the error-tolerant walk exists; this is correctly last both by handoff instruction and by genuine dependency chain
- [ ] Directory watching (live "watching …" indicator) — defer-able polish; no other feature depends on it, and its main payoff (catching external deletes) has a manual-refresh fallback if deferred

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|----------------------|----------|
| Workspace shell (rail+tree+details) | HIGH | HIGH | P1 |
| Virtualized tree at scale | HIGH | HIGH | P1 |
| ⌘K palette with cap | HIGH | MEDIUM | P1 |
| Live scan progress | HIGH | HIGH | P1 |
| Volume detection | MEDIUM | MEDIUM | P1 |
| Create slide-over + done state | HIGH | MEDIUM | P1 |
| Partial catalog on failure | MEDIUM | HIGH | P2 |
| Settings (density/rail/theme cards) | MEDIUM | LOW | P2 |
| Rename / Duplicate / Delete-to-Trash | MEDIUM | LOW-MEDIUM | P2 |
| Empty-library state | MEDIUM | LOW | P2 |
| Unreadable-catalog state + error surfacing | MEDIUM | LOW-MEDIUM | P2 |
| Re-scan & diff | HIGH | HIGH | P3 |
| Directory watching | LOW-MEDIUM | MEDIUM | P3 |

**Priority key:**
- P1: Must have for the shell to be usable/ship-able at all
- P2: Should have, meaningfully completes the design, moderate cost
- P3: Biggest remaining backend investment; correctly sequenced last per the handoff

## Competitor / Comparable-Tool Feature Analysis

| Feature area | Closest comparable(s) | How they do it | StorCat's approach |
|---------|--------------|--------------|--------------|
| Single-pane workspace | VS Code Explorer, Adobe Bridge (Folders/Content/Metadata panels), Finder | Tree/list + details panel driven by one selection state, virtualized at scale | Same pattern; rail = catalog list, tree = current catalog, details = selection |
| Command palette | VS Code Quick Open, Linear, Raycast, GitHub | Autofocus, keyboard nav, capped + counted results, no spinner | Matches closely; adds "reveal in tree" (expand ancestors) on hit |
| Scan progress | `rclone -P` / `rsync -P`, Time Machine, Backblaze | Percentage + bytes + ETA + current-item line; background hand-off to a small persistent indicator | Same fields; adds a rolling log and a pulse animation on the current path |
| Volume detection | macOS Disk Utility / Finder sidebar | OS-native enumeration, capacity shown, mount-health flagged | Same, simplified to a create-flow picker with a `read errors` tag |
| Partial/failed operation handling | rsync partial mode, TeraCopy skip/retry/abort | Continue past per-item errors, offer partial result + retry | More generous than Time Machine/Backblaze's discard-and-retry stance; matches rsync/TeraCopy's continue-and-report posture |
| Directory diff | FreeFileSync, rsync | Path-keyed, size/mtime comparison (not hash), added/removed/changed/unchanged categories, overwrite vs. keep-both | Same categorization and comparison key; "Discard" replaces a full third-way-sync mode FreeFileSync also offers (not needed here) |
| Destructive action safety | macOS Finder / Windows Recycle Bin / Linux desktop trash spec | Move to OS trash, single confirm, reversible | Same, plus domain-specific disambiguation copy ("not the card") |
| Live watch status | Dropbox menu-bar/status-line states | Small enumerated status text, debounced updates | Same idea, single status-bar line instead of a menu |
| Settings density/position | VS Code, Linear, DAM tools | Segmented controls, immediate visual effect | Same pattern, CSS-token driven |
| Empty/error states | VS Code empty Explorer, DAM "no assets" screens, compiler-style error surfaces | Explanation + concrete next action(s); technical detail for developer-facing errors | Same; unreadable-catalog state borrows compiler-error conventions (byte offset, raw parser text) |

## Sources

- [Command Palette UI Design: Best practices, Design variants & Examples — Mobbin](https://mobbin.com/glossary/command-palette)
- [Command Palette Pattern — UX Patterns for Developers](https://uxpatterns.dev/patterns/advanced/command-palette)
- [Designing a Command Palette — Destiner's notes](https://destiner.io/blog/post/designing-a-command-palette/)
- [VS Code: User interface / Custom Layout docs](https://code.visualstudio.com/docs/editing/userinterface)
- [Use and manage Adobe Bridge workspace — Adobe Help](https://helpx.adobe.com/bridge/using/adobe-bridge-workspace.html)
- [FreeFileSync: Open Source File Synchronization & Backup Software](https://freefilesync.org/)
- [Rsync: Synchronize New or Modified Files in Linux — Tecmint](https://www.tecmint.com/sync-new-changed-modified-files-rsync-linux/)
- [How to interpret rclone's progress output — rclone forum](https://forum.rclone.org/t/how-to-interpret-rclones-progress-output/21969)
- [Progress Bar & Stats — Rclone CLI docs](https://rcloneui.com/docs/cli/tips/progress-bar)
- [rsyncy (rsync progress wrapper) — PyPI](https://pypi.org/project/rsyncy/0.2.0)
- [Viewing Volumes in macOS Finder — Grokipedia](https://grokipedia.com/page/Viewing_Volumes_in_macOS_Finder)
- [The Finder confuses with wildly inaccurate figures for available space — The Eclectic Light Company](https://eclecticlight.co/2023/04/17/the-finder-confuses-with-wildly-inaccurate-figures-for-available-space/)
- [Double-check user actions: All about warning message UI — LogRocket Blog](https://blog.logrocket.com/ux-design/double-check-user-actions-confirmation-dialog/)
- [Delete Button UI: Best Practices for Designing Destructive Actions — Design Monks](https://www.designmonks.co/blog/delete-button-ui)
- [Nemo "Move to trash" confirmation dialog discussion — linuxmint/nemo#1135](https://github.com/linuxmint/nemo/issues/1135)
- [Dropbox sync icons in the desktop app for macOS — Dropbox Help](https://help.dropbox.com/sync/macos-sync-icons)
- [How to check if your files and folders are syncing — Dropbox Help](https://help.dropbox.com/sync/check-sync-status)
- [Backing up External Hard Drives — Backblaze Help](https://help.backblaze.com/hc/en-us/articles/217665398-Backing-up-External-Hard-Drives)
- [Time Machine backup failed? Top common reasons and how to fix them — Nektony](https://nektony.com/how-to/fix-time-machine-couldnt-complete-backup)
- [Why Diff Tools Lie: Detecting Hidden File Changes with PowerShell Hash Verification — DEV Community](https://dev.to/shadowstrike/why-diff-tools-lie-detecting-hidden-file-changes-with-powershell-hash-verification-10ak)
- [File Integrity Check Script (size+mtime) — IT Blog](https://ixnfo.com/en/file-integrity-check-script-size-mtime.html)
- `design_handoff_storcat_ui/README.md` (primary source — authoritative design spec, all feature grounding traces here first)
- `.planning/PROJECT.md` (project/milestone context)

---
*Feature research for: StorCat v3.0.0 Workspace Redesign*
*Researched: 2026-08-13*
