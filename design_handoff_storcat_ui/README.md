# Handoff: StorCat 3.0 UI/UX redesign

## Overview

StorCat is a Go + Wails desktop app (`scottkw/storcat`, branch `main`) that catalogs SD cards and other
external media into GNU-tree-compatible `.json` + `.html` files. The v2.3.0 UI is an Ant Design layout
with a top tab bar (Create / Search / Browse) and a wide left sidebar that changes meaning per tab.

This package contains a redesign in three directions plus a fully interactive demo of the chosen
direction (**1a — Workspace**). Nothing about the Go backend or the on-disk file formats changes;
the redesign is a frontend replacement that reads the same models
(`CatalogMetadata`, `CatalogItem`, `SearchResult`, `CreateCatalogResult`).

## About the design files

The files in `designs/` are **design references written in HTML** — running prototypes that show the
intended look, structure and behavior. They are **not production code to copy**. They use a small
streaming-component runtime (`support.js`) that has nothing to do with StorCat's stack.

The task is to **recreate these designs inside the existing frontend** —
`frontend/src` (React 18 + TypeScript + Vite + Ant Design v5) — using its established patterns:
`AppContext`, the `wailsjs/go` bindings, and the theme system in `frontend/src/themes.ts`.
Everything in the prototypes is expressible with plain React + inline styles or CSS modules; Ant
Design components can be kept where they earn their place (Modal, Select, Input, Tooltip), but the
tables and the tree are custom in this design and should be custom in code too (see *Scale*).

### How to open the prototypes

Open `designs/StorCat 1a Demo.dc.html` in any browser (Chrome/Safari/Edge). `support.js` must sit
next to it. Both files are self-contained apart from Google Fonts (IBM Plex Sans / IBM Plex Mono).

## Fidelity

**High fidelity.** Colors, type sizes, row heights, paddings and copy are final and intended to be
matched. The layout geometry (rail 268px, details 288px, 46px toolbar, 26px status bar, 27px tree
rows) is deliberate — it is what makes the workspace fit a 1280×820 window at "pro tool" density.

## Files

| File | What it is |
| --- | --- |
| `designs/StorCat 1a Demo.dc.html` | **The build target.** Interactive demo of direction 1a: rail, tree, details, ⌘K search, create flow, settings, error/empty states, responsive widths. |
| `designs/StorCat Redesign.dc.html` | The three static direction mockups (1a Workspace, 1b Console, 1c Refit) for context on what was considered and rejected. |
| `designs/support.js` | Prototype runtime. **Do not port.** |

---

## The three directions (context)

- **1a Workspace** *(chosen)* — no tabs. Catalog rail + tree + details in one view; create is a
  slide-over; search is a ⌘K palette over all catalogs.
- **1b Console** — tabs become a persistent left icon+label rail; Search and Create get full-canvas
  screens. Kept as the fallback if the single-view rewrite is too large for one release.
- **1c Refit** — keeps the three tabs, replaces the per-tab sidebar with toolbar controls and docks
  the tree preview beside the browse table. Lowest-risk port.

Everything below documents **1a**.

---

## Screens / views

### 1. Workspace (the whole app)

**Purpose** — answer "which card is that file on?" and "what's on this card?" without changing screens.

**Layout** — vertical flex, fixed chrome, three-column grid in the middle:

```
┌──────────────────────────────────────────────── 46px toolbar ─┐
│ traffic lights · StorCat mark · ⌘K search field · theme · gear │
├─────────────┬───────────────────────────────┬─────────────────┤
│ rail 268px  │ tree 1fr                      │ details 288px   │
├─────────────┴───────────────────────────────┴─────────────────┤
│ 26px status bar: 47 catalogs · 312,004 files · 4.1 TB · watch  │
└───────────────────────────────────────────────────────────────┘
```

- Grid: `display: grid; grid-template-columns: 268px 1fr 288px; min-height: 0`.
  Explicit `order: 1 / 2 / 3` on the three panes so the rail/details sides can swap (Settings →
  Catalog rail position). When swapped, the 1px divider moves with the pane (`border-right` on the
  rail becomes `border-left`).
- Every scroll region is `flex: 1; overflow-y: auto; min-height: 0`.

**Toolbar (46px, `--p2` background, 1px bottom border `--l`)**
- Traffic lights: three 11px circles, `#ff5f57 / #febc2e / #28c840`, gap 7px. (Wails owns real chrome —
  in the app this row is the draggable title bar area, keep the 46px height.)
- App mark: 16px rounded square in `--ac` + "StorCat" 13px/600, letter-spacing −0.01em.
- Search field: 460px × 30px, radius 8px, `--bg` fill, 1px `--l` border, hover border `--ac`.
  Magnifier 13px stroke 1.6 `--dm`; placeholder "Search every catalog…" 12.5px `--dm`;
  right-aligned `⌘K` badge 11px mono in a 1px `--l` box, radius 4px, padding 1px 5px.
- Theme chip: current theme name, 11.5px, 1px `--l` border, radius 6px, padding 3px 8px → opens Settings.
- Gear: 15px, two concentric circles, stroke `--dm` → opens Settings.

**Catalog rail (268px, `--p` background)**
- Header block, padding 12px 12px 10px, 1px bottom border:
  - "CATALOGS 47" — 12px/600, uppercase, letter-spacing 0.04em, `--dm`; count in `--fn`.
  - "＋ New" pill — 12px/600 `--ac` on `--acs`, radius 6px, padding 3px 8px; hover inverts to
    `--ac` background with `--onac` text. Opens the create slide-over.
  - Directory chip — 11px mono `--dm` on `--ch`, 1px `--l`, radius 6px, padding 5px 8px, ellipsized.
  - Filter input — 26px tall, radius 6px, `--bg` fill, 1px `--l`; filters title + filename, live.
- Rows (padding `6px 8px` compact / `10px` comfortable, radius 6px, gap 1px):
  - line 1: 6px status dot (red `#e5534b` when the catalog failed to parse, else transparent),
    title 12.5px/500 ellipsized, JSON size 10.5px mono `--fn`.
  - line 2: `sd12.json · 2,481 files` 10.5px mono `--fn`.
  - selected: `--sel` background + 2px left border in `--ac` + title in `--ac`; hover `--hov`.

**Tree pane (1fr, `--bg`)**
- Catalog header, padding 14px 18px 12px, 1px bottom border: title 17px/600 (−0.01em); two mono
  chips (`sd12.json`, `sd12.html`) 11px on `--ch`, radius 5px, padding 2px 7px; then a metadata line
  of 11.5px mono `--dm` values separated by `|` dividers in `--l`:
  `2,481 files | 143.6 KB | 29.4 GB catalogued | modified 2024-09-17`.
- Breadcrumb bar, 34px, `--p2`, 1px bottom border: mono path where ancestor segments are `--ac` and
  the current one is `--tx`; right side "Expand all" (`--ac`) and "Collapse" (`--dm`, hover `--ac`).
- Rows: height `--rh` (27px compact / 34px comfortable), `padding-left: 18px + depth × 16px`,
  `padding-right: 18px`, gap 8px:
  - 10px caret cell — `▾` open, `▸` closed, empty for leaves, 10px mono `--fn`;
  - 9px shape — 2px-radius square in `--ac` for directories, 50% circle in `--fn` for files;
  - name 12px mono (`--fs`), ellipsized, `--ac` when selected;
  - size right-aligned 11px mono `--fn`, `flex: none; white-space: nowrap`.
  - row background `--sel` when selected, `--hov` on hover. Click: directories toggle expand *and*
    select; files select only.

**Details panel (288px, `--p`)**
- Header row: heading ("Selected file" / "Selected folder" / "Catalog") 12px/600 uppercase `--dm`,
  and a 22px `⋯` button (1px `--l`, radius 6px, hover `--ac`) that opens the actions menu.
- Name 12.5px mono, `word-break: break-all`; path 11px mono `--dm` line-height 1.5.
- Key/value rows: label 11.5px `--dm` left, value 11.5px mono right, `padding: --mp 0`,
  1px bottom border `--l2`.
- Footer buttons stacked with `margin-top: auto`, 30px tall, radius 7px, gap 8px:
  primary "Open HTML catalog" (`--ac` fill, `--onac` text, 12.5px/600), then outlined
  "Reveal JSON in Finder", then muted "Re-scan volume…".

**Status bar (26px, `--p2`, 1px top border)** — 11px mono `--fn`: catalog count, indexed file count,
total bytes; right side `● watching ~/dev/sd-catalogs` in `--ac`, replaced by
`● scanning sd48 · 68%` while a background scan runs.

### 2. Search palette (⌘K)

Overlay `rgba(4,6,9,.62)` over the window, panel centered horizontally, `padding-top: 96px`,
720px wide, max-height 520px, radius 12px, `--p` fill, 1px `--l`, shadow `0 30px 70px rgba(0,0,0,.6)`.

- Input row 50px: 15px magnifier in `--ac`, 14px mono input, right-side count ("28 hits" /
  "50 of 3,482" / "type to search") 11px mono `--fn`. Autofocus on open.
- Result rows (`padding: --hp`, i.e. 7px 14px compact): 8px shape (square = dir, circle = file),
  name 12.5px mono over path 11px mono `--dm` (both ellipsized), catalog chip 11px `--ac` on `--acs`,
  size 11px mono `--dm` right-aligned in a 74px column. Hover `--hov`.
- Cap notice when >50 matches: 11.5px `--dm` line above the footer —
  "Showing the first 50 of 3,482 hits — refine the term, or ↵ to open the full result table."
- Empty: "No file in any catalog matches that." centered, 12.5px `--dm`, padding 30px.
- Footer 32px `--p2`: `↵ reveal in catalog`, `esc close`, right "searches names and paths".
- Clicking a hit: switch catalog, expand every ancestor of the hit, select it, close the palette.

### 3. Create catalog (right slide-over, 560px)

Same overlay scrim; panel is full height, `--p`, 1px left border, shadow `-30px 0 70px rgba(0,0,0,.5)`.
Header 52px: title + step label (`step 1 of 2`, `step 2 of 2`, `failed`, `done`) + `×`.

**Step 1 — form** (scrollable, padding 18px, gap 20px)
- "SOURCE VOLUME" — one card per detected volume: 26×20px card-shaped chip, name 13px/500,
  mount path 11px mono `--dm`, size, and a status tag (`mounted` on `--ch`; `read errors` in
  `#f0b429` on `rgba(240,180,41,.14)`). Selected card: `--sel` fill, 1px `--ac`. Below:
  "…or **choose any folder**" (`--ac`).
- "Catalog title" (free text) and "Filename root" (mono input + fixed `.json / .html` suffix in `--fn`),
  two columns, 34px fields, radius 7px, `--p2` fill, 1px `--l`.
- Options — three toggles (30×17px track, 13px knob; on = `--ac` track + `--onac` knob):
  write HTML alongside JSON (`sd48.html`), copy to secondary location (`~/Dropbox/catalogs`),
  include hidden files. Each with a mono note in `--fn`.
- "WILL WRITE" preview box: `--ch` fill, 1px `--l2`, radius 9px, 11.5px mono; recomputed live from
  the filename root and the two output toggles.
- Footer 62px `--p2`: "Create catalog" (`--ac`, 34px, radius 8px) + `⌘↵` hint + "Cancel".

**Step 2 — scanning**
- Title + big percentage in `--ac`; 6px progress track (`--ch` under `--ac` fill).
- Counters line: files seen, bytes, "about 4s left" (11.5px mono `--dm`).
- "WALKING" — current absolute path, 11.5px mono `--ac`, `animation: scpulse 1.4s ease-in-out infinite`
  (opacity 1 → .35 → 1).
- Log box: `--ch`, radius 9px, newest-first `+ /path` lines, 11px mono `--dm`, max 9 lines.
- "Run in background" (outlined) hands progress to the status bar.

**Step 2 error state** (triggered by a volume with read errors)
- 30px round `!` badge in `#e5534b` on `rgba(229,83,75,.16)`; headline
  "Stopped at 57% — the volume went away"; mono sub-line with mount point and files walked.
- Log box with red `read error: … — input/output error` lines and a `--dm` summary line.
- Explanatory 12.5px `--dm` paragraph, then actions: "Write partial catalog" (primary),
  "Retry scan" (outlined), "Cancel" (text).

**Done state**
- 30px round `✓` in `--ac` on `--acs`; title "… catalogued"; mono line "18,204 files · 119.2 GB · 41 s".
- Written-files list: 8px `--ac` square + mono path + size, 1px `--l2` separators (rows follow the
  toggles — HTML row only if HTML was requested, secondary-copy row only if enabled).
- "Open in workspace" (primary → inserts the catalog at the top of the rail and selects it) and
  "Catalog another volume" (outlined → back to step 1).

### 4. Settings (modal, 660px, max-height 700px)

Header 50px: "Settings" + `⌘,` hint + `×`. Body scrolls, padding 18px, gap 22px. Footer 56px `--p2`:
"StorCat 3.0.0 · settings save as you change them" + "Done".

- **THEME** — 2-column grid of 11 cards (8px 10px padding, radius 8px, `--p2` fill, 1px `--l`;
  selected `--sel` + 1px `--ac` + name in `--ac`). Each card: a 4-swatch strip (4 × 9×18px:
  background, panel, accent, text) + theme name 12.5px/500 + `light`/`dark` tag 10px mono `--fn`.
  The 11 themes are exactly the ones in `frontend/src/themes.ts`.
- **LAYOUT** — two labelled rows with segmented controls (3px padding, `--ch` track, 1px `--l`,
  radius 8px; active segment `--ac` fill + `--onac` text): Row density (Compact / Comfortable) and
  Catalog rail position (Left / Right).
- **CATALOGS** — catalog directory chip + "Change…"; "Default filename root" mono input (150px);
  then four toggles with sub-labels: write HTML alongside JSON, copy catalogs to a secondary
  location, watch catalog directory for changes, remember window size & position.

### 5. Catalog actions (from `⋯`)

Menu: 216px, `--p2`, 1px `--l`, radius 9px, shadow `0 18px 40px rgba(0,0,0,.5)`, items 12.5px with
`--hov` hover — Rename catalog…, Re-scan volume & diff…, Duplicate catalog, Delete catalog… (`#e5534b`).

- **Rename** (460px modal) — explains the title lives in the `.html` `<title>` and filenames don't
  change; focused input with 1px `--ac` border; Save / Cancel.
- **Delete** (480px) — "Delete “SD Card #12”?"; copy makes clear it deletes catalog files, not the
  card, and that files go to the Trash; a `--ch` box lists both paths; toggle "Also delete the
  matching .html"; destructive button "Move to Trash" (`#e5534b`, white text).
- **Re-scan & diff** (620px) — "Re-scan changed 334 entries"; four stat tiles on `--ch`
  (added `--ac`, removed `#e5534b`, changed `#f0b429`, unchanged `--dm`); a scrollable list of
  `+ / − / ~` rows with path and size (or `old → new`); actions "Overwrite catalog",
  "Keep both (write sd12-2026.json)", "Discard".

### 6. Empty and unreadable states

- **Empty library** — rail shows a short block ("No catalogs here yet" + explanation + "Catalog a
  volume →"); the tree pane centers a 46px dashed placeholder square, "Nothing catalogued yet",
  a 420px explanatory paragraph, and two buttons: "Catalog a volume", "Choose catalog folder…".
- **Unreadable catalog** — replaces the tree pane: 22px `!` badge + "This catalog can’t be read",
  explanation, key/value rows (file, failed at byte, reason, parser), a `--ch` box with the raw
  parse error in `#e5534b`, then "Re-scan volume" / "Open the .html instead" /
  "Remove from library" (`#e5534b` text). The catalog's rail row carries the red dot.

---

## Interactions & behavior

| Trigger | Behavior |
| --- | --- |
| Click rail row | Set current catalog, clear selection, expand its root |
| Type in rail filter | Case-insensitive match on title + filename |
| Click directory row | Toggle expand and select |
| Click file row | Select (details panel follows) |
| "Expand all" / "Collapse" | Expand every directory of the current catalog / collapse to root |
| Overflow row | Reveal the next 12 entries; row becomes "↑ collapse — showing 15 of 1,211 files" |
| `⌘K` / click search field | Open palette, autofocus input |
| Click hit | Switch catalog + expand ancestors + select + close |
| `Esc` | Close palette / create / settings |
| `⌘,` | Open Settings |
| `⌘↵` in create form | Start the scan |
| "＋ New" / "Re-scan volume…" | Open create slide-over at step 1 |
| Create → scan | ~220ms ticks, +2–9% each, ~4s total; volumes flagged `read errors` fail past 54% |
| "Run in background" | Close the slide-over; status bar shows `● scanning sd48 · 68%` |
| "Open in workspace" | Prepend the new catalog to the rail, select it, close |
| Theme card | Repaint immediately — every color derives from the theme |
| Density | Row height, rail row padding, details row padding, palette row padding, tree font size |
| Rail position | Swap grid order of rail and details, move the divider borders |

**Transitions** — deliberately few:

- Create slide-over: enters `translateX(100%) → 0` over **340ms** `cubic-bezier(.16,.84,.24,1)`,
  exits over **260ms** `cubic-bezier(.4,0,.7,.2)`. The scrim fades in 200ms ease-out / out 260ms
  ease-in alongside it. The panel must stay mounted for the exit — keep a `createClosing` flag and
  unmount on a 260ms timer (Escape, ×, Cancel, the scrim and "Open in workspace" all route through
  the same close path). `will-change: transform` on the panel.
- Scanning "walking" path: `scpulse` 1.4s ease-in-out infinite (opacity 1 → .35).
- Modals and the ⌘K palette: 100ms scrim fade, no movement.
- Rows: 80ms background transition on hover.

No spinners: progress is always a real number.

**Stacking order** — details panel `z-index: 3` (it becomes a drawer on narrow windows), create
slide-over and search palette `6`, dialogs and Settings `7`. The details panel outranking the
slide-over is an easy bug to reintroduce.

**Responsive** — the window is resizable; the design has three tiers:

| Window width | Behavior |
| --- | --- |
| ≥1280px | Three panes inline (`268px 1fr 288px`) |
| 1040–1279px | `236px 1fr`; details panel becomes a right drawer, toggled by a "Details" chip in the toolbar (absolute, 288px, top 46px, bottom 26px, shadow `-24px 0 50px rgba(0,0,0,.45)`) |
| <1040px | `200px 1fr`; details stays a drawer; tree keeps priority |

Below ~820px the rail should become a drawer too (not prototyped — implementer's call).

**Scale** — this is the one place the prototype is a sketch and the implementation must be real:
- Catalog trees reach 40k+ nodes. Flatten the `CatalogItem` tree once into an array of
  `{depth, name, kind, size, parentIdx}`, derive visibility from an expanded-id set, and
  **virtualize** the row list (fixed 27/34px rows make windowing trivial). Do not render the whole
  tree — the prototype's `__more__` overflow row exists to keep the mock honest; in code, either
  virtualize the full list or cap per-folder rendering at ~200 with the same "show more" affordance.
- Search should cap the result list at 50 with the "first 50 of N" line, and run off the main
  thread (Go side already walks the JSON — keep it in `internal/search`).

---

## State management

Recreate roughly this state (names from the prototype):

```
themeId: string           // one of themes.ts ids; persist to localStorage
density: 'Compact' | 'Comfortable'
side: 'Left' | 'Right'    // catalog rail position
catalogs: CatalogMetadata[]
curId: string             // current catalog id (filename root)
railFilter: string
expanded: Record<string, boolean>   // "<catalogId>:<nodeIndex>" → open
selected: string | null             // same id form
paletteOpen, createOpen, settingsOpen: boolean
query: string
step: 'form' | 'scanning' | 'error' | 'done'
form: { title, root, vol }
opts: [writeHtml, copySecondary, includeHidden]
pct, seen, log[], walkIdx            // scan progress
dialog: null | 'rename' | 'delete' | 'rescan'
detailOverlay: boolean               // narrow-window drawer
```

Initial selection is **derived, never hardcoded**: find the node, expand its ancestors.
Data needs: `BrowseCatalogs(dir)` for the rail, `LoadCatalog(path)` for the tree,
`SearchCatalogs(term, dir)` for the palette, `CreateCatalog(...)` for the scan, plus a
progress event stream (see *Mocked functionality*).

---

## Design tokens

Colors are CSS custom properties on a wrapper; every value is derived from the active theme so all
11 themes work without per-theme overrides.

| Token | Meaning | Dark default (StorCat Dark) |
| --- | --- | --- |
| `--bg` | app background / tree pane | `#0b0e13` |
| `--p` | panels (rail, details, modals) | `#0f1319` |
| `--p2` | toolbars, footers, cards | `#12161d` |
| `--ch` | chips, inputs, log boxes | `#151a22` |
| `--l` | borders, dividers | `#232a35` |
| `--l2` | subtle row separators | `mix(--l 55%, --p)` |
| `--tx` | primary text | `#e6ebf2` |
| `--dm` | secondary text | `mix(--tx 66%, --bg)` |
| `--fn` | tertiary / mono metadata | `mix(--tx 44%, --bg)` |
| `--ac` | accent | `#4fd6e0` |
| `--acs` | accent tint (chips, badges) | `mix(--ac 16%, transparent)` |
| `--onac` | text on accent fills | `#0b0e13` or `#ffffff` by luminance |
| `--sel` | selected row fill | `mix(--ac 14%, transparent)` |
| `--hov` | hover fill | `mix(--tx 8%, --bg)` |
| `--rh` | tree row height | 27px / 34px |
| `--rp` / `--mp` / `--hp` / `--fs` | rail row padding / meta row padding / palette row padding / tree font size | `6px 8px` / `6px` / `7px 14px` / `12px` |

`--onac` is computed from relative luminance (`> 0.45 → dark text`), which is what keeps light
accents (Gruvbox orange, Monokai green) and dark accents (GitHub blue) both legible. Port this
helper — it is 6 lines and it prevents the contrast bug that light themes otherwise hit.

Per-theme base values (bg / panel / panel2 / chip / line / text / accent) for all 11 themes are in
the demo's logic (`THEMES` array) and map 1:1 onto `frontend/src/themes.ts` entries — the array is
the authoritative version for this design; `themes.ts` should be extended with `p2`, `ch` and an
explicit accent rather than reusing `headerBg`/`antdPrimaryColor`.

**Type** — IBM Plex Sans (400/500/600) for UI, IBM Plex Mono (400/500) for every path, filename,
size, count and timestamp. Scale in use: 10.5, 11, 11.5, 12, 12.5, 13, 14, 15, 17, 26px.
Titles carry `letter-spacing: -0.01em`; uppercase section labels 11–12px/600 with `+0.04–0.05em`.

**Spacing** — 1, 2, 3, 5, 6, 7, 8, 10, 12, 14, 16, 18, 20, 22px. Radii: 4, 5, 6, 7, 8, 9, 12px
(rows 6, buttons 7–8, panels/modals 12). Shadows: rows/none, menu `0 18px 40px rgba(0,0,0,.5)`,
modals `0 30px 70px rgba(0,0,0,.6)`, drawers `-30px 0 70px rgba(0,0,0,.5)`.

**Status colors** (theme-independent): error `#e5534b`, warning `#f0b429`, success = `--ac`.

## Assets

No images or icon fonts. The five icons (magnifier, folder/card, plus, gear, caret) are inline SVGs
of circles, lines and rects, 10–15px, `stroke-width: 1.4–1.8`, `stroke: currentColor` or `--dm`.
The app mark is a 16px rounded square in `--ac` — a placeholder for the existing
`frontend/src/storcat-icon.svg`, which should be dropped in at 16–18px (and is the only place the
owl/disc logo appears; the old 200px hero logo is gone).

Traffic-light circles in the prototypes are mock window chrome — Wails provides the real ones.

---

## Mocked functionality (not in the codebase today)

Everything below is designed but has no backend behind it. Treat this as the feature list the
redesign implies; the UI degrades sensibly if any of it is deferred.

| Design element | Status in `main` | What implementing it needs |
| --- | --- | --- |
| Live scan progress (%, files seen, walking path, log) | **Missing** — `CreateCatalog` is a single blocking call | Emit Wails runtime events (`EventsEmit`) from `internal/catalog` during the walk; the UI already assumes an event stream. Without it: show an indeterminate state and keep the slide-over open |
| Detected volumes list (`/Volumes/*`, size, `mounted` / `read errors`) | **Missing** — only a folder picker exists (`SelectDirectory`) | Enumerate mount points per OS (`/Volumes` on macOS, `GetLogicalDrives` on Windows, `/media/$USER` + `/run/media` on Linux) with capacity; "read errors" comes from a cheap stat pass or the previous failed scan |
| Partial catalog on scan failure | **Missing** | Write what was walked plus a marker for the unreadable subtree; needs an error-tolerant walk (currently an error aborts) |
| Rename catalog | **Missing** | Rewrite the `<title>` in the `.html` (the title is read from there by `BrowseCatalogs`); JSON/filenames untouched |
| Delete catalog (to Trash) + optional `.html` | **Missing** | OS trash rather than `os.Remove` |
| Duplicate catalog | **Missing** | Copy `.json` (+ `.html`) with a suffixed filename root |
| Re-scan & diff (added / removed / changed, overwrite vs keep-both) | **Missing** | Walk the volume, compare against the loaded `CatalogItem` tree, return a diff; needs stable path keys and a size comparison |
| Watch catalog directory | **Missing** | `fsnotify` on the catalog dir → re-run `BrowseCatalogs`; drives the status-bar "watching" indicator |
| Malformed-catalog error surface (byte offset, reason) | **Partly** — parse failures are currently swallowed (`continue`) in `BrowseCatalogs` / `searchInCatalogFile` | Return per-file errors instead of skipping, and surface `json.SyntaxError.Offset` |
| Per-catalog file count and total bytes in the rail | **Missing** — `CatalogMetadata` has only the JSON file size | Add `fileCount` and `totalBytes` (computed at create time, or lazily on load and cached) |
| Search hit total when capped ("50 of 3,482") | **Partly** — `SearchCatalogs` returns everything | Return a total alongside a capped slice, or cap in the frontend and count |
| Remember window size & position | **Missing** | Persist bounds in `internal/config` and apply on startup |
| Row density, rail position, default filename root settings | **Missing** | New fields in `internal/config` (theme persistence already exists via localStorage) |
| "Open HTML catalog" / "Reveal in Finder" | **Exists** (`OpenCatalogFile` / equivalent) | Wire up |
| Tree browsing of a catalog | **Exists** (`LoadCatalog`, `CatalogModal`) | Reuse; render into the docked pane instead of a modal, virtualized |
| Search across catalogs | **Exists** (`internal/search`) | Reuse behind the palette |
| Themes | **Exists** (`themes.ts`, 11 themes) | Extend token set as noted above |

## Suggested build order

1. Shell: toolbar + 3-pane grid + status bar, token layer, theme switching from `themes.ts`.
2. Rail from `BrowseCatalogs` + tree from `LoadCatalog`, virtualized. (Ship-able already.)
3. ⌘K palette over `SearchCatalogs`, with cap + total.
4. Create slide-over: form + `CreateCatalog`, then progress events, then the error path.
5. Settings modal; move theme out of the old modal, add density / rail position / defaults.
6. `⋯` actions: rename, duplicate, delete; re-scan & diff last (biggest backend piece).
