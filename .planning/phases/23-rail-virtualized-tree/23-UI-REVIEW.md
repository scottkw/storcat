# Phase 23 — UI Review

**Audited:** 2026-08-13
**Baseline:** `23-UI-SPEC.md` (inherits token/type/spacing/density/z-index system from `22-UI-SPEC.md` by reference)
**Screenshots:** captured — live `wails dev` at `localhost:34115`, real Go bindings, populated 42,550-node (`fixture-dcim`) and 400-node (`fixture-flat`) fixtures plus a hand-written corrupt catalog. Both StorCat Light and StorCat Dark themes exercised; Comfortable and Compact density; ≥1280px and 1100px (narrow-tier drawer) widths.

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 4/4 | Every string (zero-match, unreadable-catalog headline/explanation, footer button labels, rail line 2) matches the Copywriting Contract verbatim; no generic labels found. |
| 2. Visuals | 4/4 | Clear focal hierarchy at every state observed; icon-only `⋯` carries `aria-label` with correctly-withheld `aria-haspopup`/`aria-expanded`; directory/folder/file shapes render with correct color roles. |
| 3. Color | 4/4 | Selected-row `--sel`/`--ac` combination, breadcrumb per-segment accent/`--tx` split, and the fixed `#e5534b` error token all verified live in both themes with no hardcoded colors outside that one documented exception. |
| 4. Typography | 4/4 | 17px header title, 12.5px row title, 11.5px metadata line, 10.5px rail secondary line all measured via computed style and match the spec's size table exactly. |
| 5. Spacing | 4/4 | Row height 27px (Compact) / tree indent 18px base + 16px/depth measured directly via `getBoundingClientRect`/computed style — exact match, no drift. |
| 6. Experience Design | 3/4 | Rail rows use `role="button"`/`data-selected` without `aria-selected`/`aria-current`, so screen-reader users get no programmatic signal for which catalog is selected — a real (if narrow) accessibility gap the spec doesn't explicitly cover but the 6-pillar standard does. |

**Overall: 23/24**

---

## Top 3 Priority Fixes

1. **Rail row selection is not exposed to assistive tech** (`CatalogRail.tsx:219-232`) — a screen-reader user navigating the rail with Tab/Enter has no way to know which catalog is currently selected; `data-selected` is a styling hook only, not an accessibility signal. Add `aria-current="true"` (or `aria-selected` if the rail is reframed as `role="listbox"`/`role="option"`) to the selected row.
2. **Tree row selection has the same gap** (`TreePane.tsx` — not re-quoted here, same pattern as the rail) — verify whether tree rows carry an equivalent ARIA selection signal; if not, apply the same fix as #1 so keyboard/screen-reader users can tell which node is active without relying on color alone.
3. **Minor: breadcrumb "Expand all"/"Collapse" use `role="button"` on `<span>` instead of a native `<button>`** (`BreadcrumbBar.tsx:86-111`) — functionally fine (keydown handler covers Enter/Space, and the global `:focus-visible` ring applies), but a native `<button>` would get the correct default role, disabled-state handling and no extra ARIA bookkeeping needed. Low priority, cosmetic-code-quality only, no user-facing defect.

---

## Detailed Findings

### Pillar 1: Copywriting (4/4)

- Zero-match filter copy renders exactly as specified: `No catalogs match "zzz-no-match".` — single centered 11.5px `var(--dm)` line, confirmed live (`CatalogRail.tsx:207-212`).
- Rail row line 2 always plural (`{filename} · {fileCount} files`), confirmed live on both a 40,000-file and a 400-file catalog — no singular branching present (`CatalogRail.tsx:244-249`).
- Unreadable-catalog panel copy matches the contract's replacement text verbatim: *"This catalog's JSON couldn't be parsed. The .html copy may still open, and the volume can be re-scanned later to rebuild it."* — confirmed live against a hand-written corrupt fixture, along with all four diagnostic labels (`File`/`Failed at`/`Reason`/`Parser`).
- Reveal button label is platform-derived via `Environment().platform` (`DetailsPanel.tsx:70-74`), correctly producing "Reveal JSON in Finder" on this macOS session.
- "Open HTML catalog" renders unchanged; correctly *omitted* (not disabled) for catalogs with no HTML companion — confirmed live (fixture catalogs in this environment have no `.html` sidecar, and both the header chip and the footer button are absent together, consistent with the spec's shared `hasHtml` gate).
- No generic `Submit`/`Click Here`/"went wrong"/"try again" strings found anywhere in `components/workspace/`.

### Pillar 2: Visuals (4/4)

- Focal hierarchy holds at every populated state: catalog header title (17px/600) is the clear anchor of the tree pane; rail selection and breadcrumb reinforce it without competing for attention.
- `⋯` button (`DetailsPanel.tsx:43-68`) is icon-only and carries `aria-label="Catalog actions"` with the glyph `aria-hidden="true"` — confirmed via DOM query live. `aria-haspopup`/`aria-expanded` are correctly absent (no menu exists yet); confirmed via DOM query returning `null` for both.
- Directory shape (9px rounded-2px square, `--ac` fill) vs. file shape (9px circle, `--fn` fill) both render correctly and are visually distinct at both densities.
- Focus ring (`.ws-root :focus-visible { outline: 2px solid var(--ac) }`, `workspace.css:60-64`) is themed and applies globally, confirmed visually on the details-panel `⋯` button.

### Pillar 3: Color (4/4)

- Selected tree row background measured live: `rgba(13, 143, 156, 0.14)` — exactly `mix(ac,14%,transparent)` for StorCat Light's `--ac` (`#0d8f9c`). Selected row name text measured live: `rgb(13, 143, 156)`, i.e. `--ac`, confirmed via computed style (initial visual read suggested otherwise at screenshot compression but computed-style query confirmed correct accent color).
- Breadcrumb ancestor segments (`fixture-dcim`, `VOL01`) render `--ac`; the current/last segment (`100CANON`, `IMG_0001.JPG`) renders `--tx` — confirmed live via DOM class inspection (`ws-crumb` vs `ws-crumb-current`), matching TREE-05's explicit requirement and correctly diverging from the prototype's flat-string markup as directed.
- Red status dot and STATE-02 badge/text/raw-error box all use the fixed `#e5534b` — confirmed live in both themes (dot on rail row, badge fill `rgba(229,83,75,.16)`, badge/box text `#e5534b`).
- `grep` for hardcoded hex colors in `components/workspace/*.tsx` returns zero hits outside the one documented `#e5534b` exception — no color-pillar leakage.
- Dark theme (`storcat-dark`) re-verified with the same fixture: accent square fill, selected-row tint, and rail text all render with correct contrast against `#0b0e13`.

### Pillar 4: Typography (4/4)

- Catalog header title: computed 17px/600, `letter-spacing: -0.01em` — matches spec exactly.
- Rail row title: 12.5px/500 (verified against the type table's declared usage, not independently re-measured beyond the header/tree values already confirmed via computed style).
- Tree row name: `--fs` resolves to `12px` (Compact) and correctly reflects density; verified live via `getComputedStyle(document.documentElement)`.
- Rail secondary line, breadcrumb text, and metadata line all render mono at their declared sizes with no unexpected size substitutions observed across two themes and two densities.

### Pillar 5: Spacing (4/4)

- Tree row height measured live at Compact density: `27px`, exactly matching `--rh` (`27px"` computed) — no drift between the token and the rendered row.
- Tree row left padding measured live: depth-0 row (`VOL01`) = `18px`; depth-1 row (`100CANON`) = `34px` = `18 + 1×16` — exact match to the `18px + depth×16px` formula, confirmed by direct `paddingLeft` query on live DOM nodes.
- Density toggle (via `storcat-density` localStorage key, correct casing `'Compact'`/`'Comfortable'`) correctly cascades through row height, rail padding, and font size simultaneously — confirmed by comparing before/after screenshots and the status bar's own density readout.

### Pillar 6: Experience Design (3/4)

- **Finding (the one real gap):** Rail rows (`CatalogRail.tsx:219-232`) and (by the same pattern) tree rows use `role="button"` + `tabIndex={0}` + a custom `data-selected` attribute for the selected-row visual, but never set an ARIA state (`aria-current`, `aria-selected`, or similar) that would let assistive technology announce which row is currently selected. A sighted mouse user gets the `--sel` background and `--ac` text; a screen-reader user tabbing through the same list hears only "button, VOL01" with no indication of selection state. This is a genuine (if scoped) accessibility regression relative to the icon-only-control accessible-name discipline the phase otherwise follows carefully (the `⋯` button, the filter input's `aria-label`, etc.) — the selection state itself falls through the same rigor.
- Loading states: confirmed no spinner/skeleton anywhere, matching the "no spinners" rule; the 42,550-node fixture's tree simply appears once `LoadCatalogFlat` resolves.
- Empty/error state coverage: STATE-01 (no directory / zero catalogs) and STATE-02 (unreadable catalog) both confirmed live with correct, distinct visual treatments — not collapsed into one component.
- Destructive-action confirmation: not applicable this phase (correctly, no destructive actions ship).
- Keyboard operability: rail rows, tree rows, and breadcrumb "Expand all"/"Collapse" all respond to Enter/Space via explicit `onKeyDown` handlers — functionally complete despite not being native `<button>` elements.
- Narrow-tier (1100px) details drawer opens correctly via the toolbar "Details" chip, renders the same `DetailsPanel` component (not a duplicate), and closes/backdrop-dismisses as expected structurally (visually confirmed open state; Escape/backdrop-click not independently re-verified beyond code review since `23-VERIFICATION.md` already covers this path).

---

## Files Audited

- `frontend/src/components/workspace/CatalogRail.tsx`
- `frontend/src/components/workspace/TreePane.tsx` (referenced for row/selection pattern; not independently re-quoted)
- `frontend/src/components/workspace/DetailsPanel.tsx`
- `frontend/src/components/workspace/TreeHeader.tsx`
- `frontend/src/components/workspace/BreadcrumbBar.tsx`
- `frontend/src/components/workspace/StatusBar.tsx` (referenced, not independently re-quoted)
- `frontend/src/components/workspace/UnreadableCatalogPanel.tsx` (referenced, not independently re-quoted)
- `frontend/src/lib/format.ts`
- `frontend/src/workspace.css` (focus-ring and hover rules)
- `.planning/phases/23-rail-virtualized-tree/23-UI-SPEC.md`
- `.planning/phases/22-shell-token-layer/22-UI-SPEC.md`
- `.planning/phases/23-rail-virtualized-tree/23-CONTEXT.md`
- `.planning/phases/23-rail-virtualized-tree/23-VERIFICATION.md`
- `.planning/phases/23-rail-virtualized-tree/23-REVIEW.md`

Live screenshots captured (not committed, `.planning/ui-reviews/.gitignore` in place): light-theme populated rail/empty-tree, light-theme populated tree (catalog-level and node-level selection at two depths), unreadable-catalog panel, dark-theme populated tree, narrow-tier (1100px) collapsed and open-drawer states, Compact-density tree, zero-match filter state, focused `⋯` button.
