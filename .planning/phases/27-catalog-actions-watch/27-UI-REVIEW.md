# Phase 27 — UI Review

> ## Orchestrator corrections (2026-08-16, post-audit)
>
> **Priority fix 1 is STALE — the menu focus defect is fixed, not open.** The auditor read `WINDOWS.md`
> entry 13's *description* (which narrates the original defect) and missed its `fixed` status column.
> The defect was root-caused during this phase's code review (`pointerdown` fires before `mousedown`,
> whose default action stole focus to `<body>` *after* `useModalBehavior` had already restored it), fixed
> in commit `12c32dd3` with `event.preventDefault()`, and then **re-scoped** in commit `ea3a51e0` when
> that blanket `preventDefault()` was found to swallow focus for every focusable element on the page.
> The final rule — suppress the default action only when the outside-click target is not itself focusable
> — was live-verified with real CDP-trusted mouse events in both directions: a non-focusable click closes
> the menu and restores focus to `⋯`, and a click straight onto another input focuses that input on the
> first click. The auditor's suggested remedy (call `triggerRef.current?.focus()` directly) would
> reintroduce the exact race the `preventDefault()` exists to win. **No action taken; finding withdrawn.**
>
> **Priority fix 2 is REAL and has been applied.** `DetailsPanel.tsx:327` hardcoded `#e5534b` instead of
> the `var(--danger)` token this phase introduced. Fixed. Scope note the audit did not surface: the
> literal appears in **8** places repo-wide, 6 of which predate this phase (`workspace.css` ×3,
> `UnreadableCatalogPanel.tsx` ×2, `CatalogRail.tsx` ×1) — those were written before `--danger` existed
> and sit in files this phase did not touch, so they are left as a recorded follow-up rather than swept
> into a phase that cannot re-verify them.
>
> **Priority fix 3 is tracked, not fixed.** `RenameDialog`'s failure sub-state was verified by binding
> rejection plus code-review parity, not clicked through end-to-end. Added to `27-UAT.md` as a human item.
>
> Adjusted score: **22/24** (Experience Design 2/4 → 4/4; the pillar's deduction rested entirely on the
> withdrawn finding 1 and on finding 3, which is an evidence-method gap rather than a defect).



**Audited:** 2026-08-16
**Baseline:** `27-UI-SPEC.md` (design contract, approval pending, checker sign-off unchecked at time of audit)
**Screenshots:** not captured (no `localhost:3000`/dev server reachable in this session — this is a Wails desktop app; `wails dev` runs on `:34115` and was not running during this audit). Audit performed at code level, cross-referenced against the 27-04/27-05/27-07 SUMMARYs' recorded live dev-browser evidence (28-row verification matrix) where a claim could only be settled visually.

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 4/4 | Every string in the four new surfaces matches the Copywriting Contract verbatim; no generic labels; ACT-05's "no permanence vocabulary" constraint holds by direct grep. |
| 2. Visuals | 3/4 | Menu/dialog geometry, hierarchy, and destructive-button treatment all match spec, but the actions-menu's click-outside dismissal has a live-verified focus-restore defect (focus lands on `<body>`, not the trigger) — a real, currently-open interaction bug, not a visual one, but it undercuts the "every overlay restores focus" contract the menu explicitly claims. |
| 3. Color | 3/4 | `--danger`/`--ondanger` declared once, theme-independent, correctly consumed by all three new spec-named consumers (menu item, dialog buttons, new dialog error banners). But the pre-existing `Footer` action-error span in the same file (`DetailsPanel.tsx:327`) — semantically identical to the new dialog error banners — was left as a raw `#e5534b` hex literal instead of migrated to `var(--danger)`, undermining the stated "one declaration instead of three copies of a hex literal" goal within the very file this phase touched. |
| 4. Typography | 4/4 | Rename field correctly renders in Sans (not mono) per the spec's explicit correction; all sizes/weights (15/600, 12.5/500, 11.5px mono, 12px error) match the locked scale with no new values introduced. |
| 5. Spacing | 4/4 | 440px dialog width, 160px menu min-width, 6px/10px item padding, 9px/12px radii, and both shadow values all match the spec's literals exactly; z-index tier (`--z-menu: 150`) correctly slotted between `--z-details-drawer` (100) and `--z-overlay` (200). |
| 6. Experience Design | 2/4 | Two real, live-verified defects are open and unresolved at time of audit: (1) Menu click-outside focus-restore loses focus to `<body>` (WINDOWS.md #13, reproduced twice with independent evidence) — breaks the app's "focus restores on every close path" convention for one specific dismissal method; (2) `RenameDialog`'s failure sub-state (banner render, dialog-stays-open, retry-enabled) was verified only by code-review parity with `Footer`'s pattern, not exercised end-to-end through the dialog UI (27-04 SUMMARY's own flagged gap, `human_judgment: true`). Both are documented honestly in the SUMMARYs rather than hidden, which is a genuine credit to the process, but the defects themselves are real and open. |

**Overall: 20/24**

---

## Top 3 Priority Fixes

1. **Menu click-outside dismissal does not restore focus to the `⋯` trigger — it lands on `<body>`.** User impact: a keyboard/screen-reader user who dismisses the menu by clicking elsewhere in the app loses their place entirely, breaking the one accessibility guarantee this component's own JSDoc and the UI-SPEC both promise ("Focus restores to the `⋯` trigger on close, matching every other overlay in this app"). Concrete fix: in `Menu.tsx`'s `pointerdown` handler, when the target is non-focusable, call `triggerRef.current?.focus()` directly after `onClose()` rather than relying on `useModalBehavior`'s own `restoreTarget.focus()` cleanup to win the race against the browser's native focus-follows-click default action on `<body>` — the current `preventDefault()` scoping (WR-02) stops the browser from stealing focus to *other focusable elements* but does not stop the click's default action from moving focus to `<body>` when the click target itself isn't focusable. This is `WINDOWS.md` entry #13, already tracked but not yet closed.

2. **Inconsistent adoption of the new `--danger` token within the very file this phase touched.** User impact: none directly visible (both render identically today, since `--danger` is a fixed `#e5534b` non-theme-derived value), but this is technical debt with a real regression risk — if `--danger` is ever retuned per-theme, `DetailsPanel.tsx:327`'s Footer error span (and `CatalogRail.tsx:317`, `UnreadableCatalogPanel.tsx:57/104`) will silently diverge from the dialogs' error banners despite being the same semantic "action failed" message class. Concrete fix: replace `color: '#e5534b'` at `DetailsPanel.tsx:327` with `color: 'var(--danger)'` — this is the Footer's own action-error slot that `CatalogActions`' menu-item error messages (duplicate/rename failures reported via `onError`) already flow through, making it a direct sibling of the newly-tokenized dialog banners.

3. **`RenameDialog`'s failure sub-state has no end-to-end live verification, only code-review parity with an existing pattern.** User impact: a failed rename (e.g. a permissions error, a disk-full condition) is the one path in this dialog that was never actually clicked through in a running app this phase — if the banner fails to render, the dialog silently closes, or the field value is lost, that would surface as data loss on a common failure mode. Concrete fix: run one manual UAT pass — trigger a real rename failure (e.g. temporarily clear `catalogDir` or point it at a read-only path) and confirm the banner text, dialog-stays-open, edited-value-preserved, and button-re-enabled behaviors all hold, closing the `human_judgment: true` flag `27-04-SUMMARY.md` (coverage D3) left open.

---

## Detailed Findings

### Pillar 1: Copywriting (4/4)
- All menu item labels, dialog titles, field labels, button labels, and error-banner templates match `27-UI-SPEC.md`'s Copywriting Contract table verbatim: `RenameDialog.tsx:68` (`` `Couldn't rename this catalog: ${result.error}.` ``), `DeleteConfirmDialog.tsx:128` (`` Couldn't move {catalog.title} to the Trash: {error}. ``), `DeleteConfirmDialog.tsx:42` (two-branch `Move to Trash` / `Move both to Trash` primary label).
- `grep -Eqi 'permanent|forever|erase|unrecoverab'` against `DeleteConfirmDialog.tsx` returns nothing (confirmed both via direct grep in this audit and via 27-05's own acceptance-criterion D6) — ACT-05's "no permanence vocabulary, no third button" constraint holds.
- No bare generic labels (`Submit`/`Cancel`/`OK`/`Done`) found in any of the four new surfaces (`Menu.tsx`, `DialogShell.tsx`, `RenameDialog.tsx`, `DeleteConfirmDialog.tsx`, `StatusBar.tsx`).
- The Settings watch-directory note was correctly amended from Phase 26's placeholder ("applies once file watching ships") to "Detects catalogs added, removed, or edited outside the app" (`CatalogSettingsSection.tsx:136`), confirmed both statically and via 27-07's live DOM string search.

### Pillar 2: Visuals (3/4)
- Menu positioning (`position: fixed`, computed once via `getBoundingClientRect()`, right-aligned to the trigger) correctly escapes the details panel's `.pane-scroll` overflow region — no `transform`/`will-change`/`filter` on any ancestor (`ws-details`, `.pane-scroll`, `.ws-root`) that would create a competing containing block and break the fixed-positioning trick.
- Destructive styling is visually isolated to the delete path as spec'd: the rename dialog's primary button uses `--ac` (accent), never touching `--danger`; the delete dialog's primary uses `--danger`/`--ondanger` exclusively, with no accent color anywhere in that component — correctly avoiding the "two colors competing for attention" failure mode the spec calls out.
- **Live-verified defect (not fixed at time of audit):** clicking outside the open menu on a non-focusable page area closes the menu but leaves `document.activeElement` on `<body>` instead of restoring to the `⋯` trigger (`WINDOWS.md` #13, reproduced twice with independent focus-event-log evidence in 27-07's Task 3). Escape-driven close restores focus correctly every time — only the click-outside path is affected. This is the one console-hierarchy visual/interaction contract this phase's own component doc comment claims to guarantee ("Focus restores to the ⋯ trigger on close ... matching every other overlay in this app") and currently does not, for one of its two close paths.

### Pillar 3: Color (3/4)
- `--danger: #e5534b` / `--ondanger: #ffffff` declared exactly once in `workspace.css:30-31`, theme-independent (confirmed absent from `themes.ts`'s 458-line theme registry — not per-theme-derived, consistent with the spec's Claude's-Discretion resolution).
- All three spec-named new consumers correctly use the token, not a literal: `.ws-menu-item-danger { color: var(--danger); }` (`workspace.css:1723`), `.ws-dialog-btn-danger { background: var(--danger); color: var(--ondanger); }` (`workspace.css:1845-1848`), `.ws-dialog-error { color: var(--danger); }` (`workspace.css:1861`, used by both new dialogs' inline error banners).
- **Gap:** `DetailsPanel.tsx:327`'s pre-existing `Footer` action-error span still hardcodes `color: '#e5534b'` rather than `var(--danger)`. This span is in the same file `27-04` modified to hoist the error state (`actionError`/`setActionError`), and per the spec's own framing ("this phase's three consumers ... share one declaration instead of three copies of a hex literal"), it is a fourth de facto consumer of the same semantic color that was not migrated. `27-04-SUMMARY.md` explicitly documents that "the three pre-existing `#e5534b` literals were left untouched" (referring to `workspace.css:831/1052/1099` — parse-error text, scan-error badge, error log lines) — but does not mention `DetailsPanel.tsx:327`, `CatalogRail.tsx:317`, or `UnreadableCatalogPanel.tsx:57/104`, which are additional undocumented hex-literal siblings outside that acknowledged set.
- No accent (`--ac`) misuse found in the destructive dialog; no new base or per-theme color tokens introduced beyond the two declared.

### Pillar 4: Typography (4/4)
- Rename field (`ws-rename-input`) correctly renders `font-weight: 400` with no `mono` class — matches the spec's explicit correction that a catalog title is display prose, not a mono-qualifying value.
- Dialog titles at 15px/600 (`ws-dialog-title`), item/field labels at 12.5px/500 (`ws-menu-item`, `ws-dialog-label`, `ws-delete-check`), delete lead sentence and path boxes at 11.5px (`ws-delete-lead`, `ws-delete-path-box`), error banners at 12px/1.5 line-height (`ws-dialog-error`) — every value traced to `workspace.css` matches the spec's Typography table with no new sizes or weights introduced.
- Status-bar watching segment inherits `mono` from `.ws-status`'s existing class per spec, confirmed at `StatusBar.tsx:85-90`.

### Pillar 5: Spacing (4/4)
- New literal geometry values all match: 440px dialog width (`workspace.css:1769`), 160px menu min-width (`:1694`) and directory-truncation max-width (`:1969`, correctly reused rather than a new number), 6px 10px menu-item padding (`:1707`), 8px 10px delete path-box padding (`:1913`).
- Reused-as-is values confirmed: 50px header / 56px footer (`:1781`, `:1815`), 20px body padding (`:1808`), 16px body gap (`:1811`), 12px row gap (`:1922`), 8px footer gap (`:1822`), 32px footer-button height (`:1827`), 30px rename-field height (`:1879`), 9px menu/path-box radius (`:1697`, `:1912`), 12px dialog-panel radius (`:1771`), 7px rename-field radius (`:1880`).
- Both shadow values (`0 18px 40px rgba(0,0,0,.5)` menu, `0 30px 70px rgba(0,0,0,.6)` dialogs) match verbatim (`:1700`, `:1774`).
- `--z-menu: 150` correctly slotted between `--z-details-drawer: 100` and `--z-overlay: 200` (`workspace.css:41,48-50`), consumed only by `.ws-menu` (`:1693`); dialogs correctly reuse `--z-dialog: 300` unchanged (`:1758`).

### Pillar 6: Experience Design (2/4)
- Menu correctly implements roving `tabIndex` with ArrowUp/ArrowDown wraparound (`Menu.tsx:111-126`), Escape-to-close via the shared hook, and its own click-outside `pointerdown` listener since `useModalBehavior` provides no scrim for a menu. Escape-driven focus restore verified live and passing (27-07 matrix row 4).
- Delete dialog's busy-state treatment (`opacity: 0.7` + both buttons `disabled`) correctly reuses the existing `Footer` pattern (`DeleteConfirmDialog.tsx:82-83,97-98`) — no spinner introduced, consistent with project convention.
- Watching status-bar segment correctly omits itself entirely (not a placeholder/dash) when either `watchDirectory` is off or `catalogDir` is unset (`StatusBar.tsx:58,84`), matching the scan segment's established "no reserved empty slot" rule; live-verified (27-07 matrix rows 20-22).
- **Open defect #1:** Menu click-outside focus-restore lands on `<body>`, not the trigger (see Visuals/Color findings above) — a genuine, currently-unfixed interaction gap in a component whose own documentation claims full focus-restore coverage. Tracked as `WINDOWS.md` entry 13, `kind: unmet-truth`, but not yet closed.
- **Open defect #2:** `RenameDialog`'s error sub-state (banner render on failure, dialog staying open, edited value preserved, button re-enabling) is unverified end-to-end through the live UI — only code-review parity with `Footer`'s already-proven identical pattern and a direct binding-rejection check (`27-04-SUMMARY.md` coverage D3, `human_judgment: true`). The underlying error-string fidelity is proven; the dialog's own render behavior on that error path is not.
- No new frontend test framework added or recommended, consistent with `TEST-01` deferred / project convention — proof relies on `tsc --noEmit`, `npm run build`, and live dev-browser verification, all present and cited in the SUMMARYs.

---

## Files Audited
- `/Users/ken/dev/storcat/frontend/src/components/workspace/Menu.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/DialogShell.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/RenameDialog.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/DeleteConfirmDialog.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/DetailsPanel.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/StatusBar.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/CatalogRail.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/settings/CatalogSettingsSection.tsx`
- `/Users/ken/dev/storcat/frontend/src/hooks/useModalBehavior.ts`
- `/Users/ken/dev/storcat/frontend/src/workspace.css`
- `/Users/ken/dev/storcat/frontend/src/themes.ts`
- `/Users/ken/dev/storcat/pkg/models/catalog.go` (spot-checked for `title`/`hasHtml` field shape referenced by the frontend)
- `.planning/phases/27-catalog-actions-watch/27-UI-SPEC.md`, `27-CONTEXT.md`, `27-04-SUMMARY.md`, `27-05-SUMMARY.md`, `27-07-SUMMARY.md` (cross-referenced for live-verification evidence)

Registry audit: not applicable — `shadcn_initialized: false`, no component registry used this phase (confirmed via `27-UI-SPEC.md`'s own Registry Safety section).
