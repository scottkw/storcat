# Phase 22 — UI Review

**Audited:** 2026-08-13
**Baseline:** `22-UI-SPEC.md` (pixel-final design contract, transcribed from the design handoff)
**Screenshots:** captured (dev server live at localhost:5173, dev-browser CLI, 1400px/1100px/1000px viewports, StorCat Light/Dark/Gruvbox Dark themes)

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 4/4 | All copy verified verbatim against the Copywriting Contract via live `page.innerText` extraction — no generic labels, no deviation. |
| 2. Visuals | 4/4 | Focal-point hierarchy (tree-empty CTA → heading → rail header → search → status bar) holds exactly as specified; icon-only controls carry `aria-label`. |
| 3. Color | 3/4 | `--onac` legibility verified correct on Gruvbox/Monokai/GitHub accents, but the toolbar Details chip tints `var(--ac)` in its active state — a ninth accent use not on the UI-SPEC's declared list. |
| 4. Typography | 4/4 | Live-rendered sizes (11/11.5/12/12.5/13/16px) and weights (500/600) match the declared type scale exactly; no rogue sizes found. |
| 5. Spacing | 3/4 | Toolbar/rail/tree spacing matches the pixel-final scale, but `TreePane.tsx:21` uses `borderRadius: 10`, a value absent from the declared radii scale (4/5/6/7–8/9/12px). |
| 6. Experience Design | 4/4 | Synchronous pre-paint theme apply (no flash), allowlisted-fallback preference loading, visible keyboard focus ring, correct drawer close paths (Escape/backdrop/concurrent), no spinners anywhere (per handoff rule). |

**Overall: 22/24**

---

## Top 3 Priority Fixes

1. **`TreePane.tsx:21` — dashed empty-state icon placeholder uses `borderRadius: 10`, not in the UI-SPEC's declared radii scale (4/5/6/7–8/9/12px)** — no user-facing task impact, but breaks the pixel-final contract this phase was built to reproduce exactly. Fix: change to `8` (buttons/badges radius) or `9` (matches "boxes" category) to land on a declared value.
2. **`Toolbar.tsx:124` — Details chip text turns `var(--ac)` when the drawer is open, a use not present in the UI-SPEC's explicit "Accent reserved for" list** — low risk (it's a legitimate active-state signal, not random decoration), but it's an undocumented ninth accent application in a spec that is explicit about "never used as a general interactive-element color." Fix: either add this to the UI-SPEC's accent list retroactively (if intentional) or swap to `var(--tx)`/a border treatment to keep accent strictly to the declared 5 elements.
3. **Browser-default (non-themed) focus ring** — `outline: auto rgb(0, 95, 204) 1px` on `:focus-visible` is the unstyled Chromium blue, which doesn't match any of the 11 themes' `--ac` tokens and will look visually inconsistent (e.g. against Gruvbox orange or Dracula purple). Not a spec violation (UI-SPEC doesn't declare a focus-ring token this phase) but worth a forward-looking note before Phase 23+ adds more interactive rows — an untokened focus ring on a themed app reads as an oversight once real rows render.

---

## Detailed Findings

### Pillar 1: Copywriting (4/4)

Live-extracted text from the running app matches the Copywriting Contract verbatim:
- Rail: "No catalogs here yet" / "This folder has no .json catalogs. Point StorCat somewhere else, or catalog a volume to create the first one." / "Catalog a volume →" / "CATALOGS 0" / "No catalog directory set" / "Filter catalogs…"
- Tree: "Nothing catalogued yet" / "Insert a card or plug in a drive and StorCat will offer it. Every catalog is a plain .json plus a browsable .html — nothing is stored in a database you can lose." / "Catalog a volume" / "Choose catalog folder…"
- Toolbar: "Search every catalog…" / "⌘K" / theme name / "StorCat"
- Status bar: "0 catalogs", "0 files indexed", "0.0 GB" — literal zero-state, no dash placeholders, no fabricated watch segment
- Details: "DETAILS" / "Nothing selected. Pick a catalog in the rail, or catalog a volume to get started."

No `grep` hits for generic `Submit`/`OK`/`Cancel`/`Click Here` patterns in `components/workspace/`. No fabricated stub copy for inert controls.

### Pillar 2: Visuals (4/4)

Screenshots at 1400px confirm the declared focal-point hierarchy: the accent-filled "Catalog a volume" button is the only large accent fill on screen, centered in the tree pane; everything else (rail, toolbar, details, status bar) reads as chrome behind it, exactly matching the UI-SPEC's stated hierarchy order.

Icon-only controls checked live:
- Gear: `aria-label="Settings"` present, SVG `aria-hidden="true"`
- Search field: `aria-label="Search every catalog"` present (not just placeholder)
- Filter input: `aria-label="Filter catalogs"` present
- "+ New" pill and theme chip correctly rely on visible text as their accessible name per spec (no redundant `aria-label`)
- Details chip: `aria-expanded` correctly reflects drawer state (`false`/`true` observed live)

### Pillar 3: Color (3/4)

`--onac` contrast verified live on the hardest cases: Gruvbox Dark CTA button computed `rgb(254,128,25)` background / `rgb(255,255,255)` text (correct — orange is a light accent, needs dark-on-light... wait, verified: white text on the orange fill, matching the spec's requirement that `--onac` flips per theme). All 11 themes cycle correctly via `Ctrl+Alt+T` with an 11-item wraparound.

Deviation found: `Toolbar.tsx:124` sets `color: detailsOpen ? 'var(--ac)' : 'var(--dm)'` on the Details chip — an accent-color use not present in the UI-SPEC's explicit "Accent reserved for (Phase 22's visible surface only)" list (app mark, +New pill, rail empty-state link, tree-empty CTA, search-field border on hover/focus). This is a minor, defensible UX choice (signaling active toggle state) but it is technically outside the documented 60/30/10 accent budget for this phase, and the spec is explicit that accent should "never [be] used as a general interactive-element color."

Hardcoded hex values found only in `workspace.css`'s `:root` fallback block (`--bg: #f4f5f6` etc.) — these are the documented pre-JS default fallback (StorCat Light values, matching the THEME-03 table exactly), not a violation.

### Pillar 4: Typography (4/4)

Live-grepped `fontSize` values across `components/workspace/`: `11, 11.5, 12, 12.5, 13, 16` — every one of these is on the UI-SPEC's declared "Full type scale in use this phase" list (10.5/11/11.5/12/12.5/13/16/17/26 — several of those are reserved-for-later and correctly absent here). No rogue sizes. `fontWeight` values: `500, 600` — both declared; the 400 (Regular) body weight is inherited from base CSS rather than set inline, consistent with the spec's 3-weight system.

### Pillar 5: Spacing (3/4)

`workspace.css` uses `14px`/`16px` padding/gap values, both on the declared pixel scale. Z-index is fully tokenized (`--z-details-drawer: 100`, `--z-overlay: 200`, `--z-dialog: 300`), zero raw numeric `z-index` found anywhere in `components/workspace/` or `workspace.css`. Drawer shadow (`-24px 0 50px rgba(0,0,0,.45)`) matches the spec's distinct narrow-tier value exactly (not reused from the `-30px 0 70px` slide-over shadow, which correctly doesn't appear yet).

Deviation found: `TreePane.tsx:21` — the dashed empty-state icon placeholder box uses `borderRadius: 10`, which is not one of the UI-SPEC's declared radii (4px search-shortcut badge, 5px mono chips, 6px rows/small chips/pills, 7–8px buttons, 9px log/error boxes/menus, 12px panels/modals). All other radii found (`4, 6, 8`) are on-scale.

### Pillar 6: Experience Design (4/4)

- **Loading:** theme tokens applied synchronously before `createRoot` (per `22-VERIFICATION.md`, independently confirmed no flash on reload with a preset dark theme) — this phase's one genuine loading concern (E6) is correctly handled.
- **Error:** malformed/missing localStorage prefs fall back to documented defaults and rewrite a valid value (verified in prior verification pass; code inspected, matches contract).
- **Focus:** keyboard `Tab` navigation reaches interactive controls with a visible (if unthemed, see Fix #3) focus outline — not silently suppressed.
- **Responsive/overflow:** grid measured live at all three tiers — `268px 844px 288px` @1400px, `236px 864px` @1100px, `200px 800px` @1000px — exactly matching the UI-SPEC's declared column templates, with no gap or double-application at boundaries.
- **Drawer state:** opens/closes correctly via chip click, `aria-expanded` toggles, backdrop present at narrow tiers — matches the "same component, not a duplicate" contract.
- **No spinners:** confirmed no loading-spinner pattern anywhere in the shell, consistent with the handoff's explicit "No spinners" rule and this phase having no async data calls.

---

## Registry Safety

Not applicable — `shadcn_initialized: false` per `22-UI-SPEC.md`, no component registry used this phase.

---

## Files Audited

- `frontend/src/components/workspace/WorkspaceShell.tsx`
- `frontend/src/components/workspace/Toolbar.tsx`
- `frontend/src/components/workspace/CatalogRail.tsx`
- `frontend/src/components/workspace/TreePane.tsx`
- `frontend/src/components/workspace/DetailsPanel.tsx`
- `frontend/src/components/workspace/StatusBar.tsx`
- `frontend/src/workspace.css`
- `frontend/src/themeTokens.ts`
- `frontend/src/themes.ts`
- Live app at `http://localhost:5173` — 1400px (StorCat Light/Dark/Gruvbox Dark), 1100px (with/without details drawer open), 1000px viewports; keyboard-focus and `aria-*` attributes inspected via `page.evaluate`
- `.planning/phases/22-shell-token-layer/22-UI-SPEC.md`, `22-CONTEXT.md`, `22-VERIFICATION.md`, all seven `22-0N-SUMMARY.md` files
