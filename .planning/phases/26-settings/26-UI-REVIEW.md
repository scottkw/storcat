---
phase: 26
slug: settings
status: reviewed
score: 22/24
baseline: 26-UI-SPEC.md
reviewed: 2026-08-15
method: code-level audit (Playwright-MCP unavailable; leaned on SUMMARYs' recorded live dev-browser evidence for visual-only claims)
---

# Phase 26: Settings — UI Review

**Baseline:** `26-UI-SPEC.md` (approved design contract)
**Overall: 22/24**

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 3/4 | All contract strings verbatim except the watch-directory toggle note, which intentionally diverges from the UI-SPEC's own locked Copywriting Contract table with no spec update to match |
| 2. Visuals | 4/4 | Consistent hierarchy, every icon-only control has an `aria-label`, exactly-one-selected invariant enforced on both theme grid and segmented controls |
| 3. Color | 4/4 | Accent reserved exactly for the declared list (selected theme card, active segment, "Change…" link, toggle-on track, footer button); zero hardcoded hex/rgb in the audited files |
| 4. Typography | 4/4 | No new sizes/weights outside the locked scale; the 10px→11px theme-tag round-up is applied and documented |
| 5. Spacing | 4/4 | Every literal geometry value (660/700px, 50/56px, 8px grid gap, 9×18 swatches, 30px input height, 3px/4px 12px vs 14px segment padding) matches the spec byte-for-byte |
| 6. Experience Design | 3/4 | Save-as-you-change contract fully honored (no Save/Cancel/dirty-state anywhere, grep-confirmed) but two real interaction branches are verified only by code-trace, not by a live click |

---

## Top 3 Priority Fixes

1. **Watch-directory toggle note diverges from the locked spec copy.** `CatalogSettingsSection.tsx:136`
   renders `"applies once file watching ships"`, but `26-UI-SPEC.md`'s Copywriting Contract table locked
   `"refresh the rail automatically"`. The *implemented* copy is the correct choice — it does not imply an
   active watcher, which is Phase 27's job (and the original copy was flagged as WR-02 in this phase's code
   review) — but the design contract was never amended, so the spec described UI that does not exist.
   **→ RESOLVED in this pass:** `26-UI-SPEC.md`'s Copywriting Contract table updated to the shipped copy,
   with the reason recorded inline.

2. **Two behaviors proven only by code-trace, not a live interaction.**
   (a) `WorkspaceShell.tsx`'s `openSettings()` early-return during an active foreground scan (⌘, no-op on
   `state.scan.status === 'counting'/'scanning'`) was never exercised against a real in-progress scan (26-01 D4).
   (b) `CatalogSettingsSection.handleToggleCopyToSecondary`'s empty-state branch — which opens the native
   folder picker and handles a cancel — was never driven through an actual click (26-05 D3), because this
   project's standing no-host-OS-GUI-automation rule forbids driving a native dialog.
   Both are exactly the class of edge case that regresses silently.
   **→ Tracked**, not resolved here: both are recorded in `26-UAT.md` (items 1 and 3) for a human pass.
   (a) needs a slow-scan fixture; (b) needs a human click or a stub harness.

3. **`26-UI-SPEC.md`'s UI Considerations header arithmetic does not close.** The header claimed
   "Applicable: 27 — resolved 27 (24 explicit, 3 backstop)" while its five tables yield 24 explicit + 2
   backstop = 26. The executor surfaced this rather than fabricating a third backstop row to make the
   header true — the correct call — but the header misreported its own coverage to the next reader.
   **→ RESOLVED in this pass:** header corrected to the tables' actual count, with a note explaining the
   original discrepancy.

---

## Detailed Findings

**Pillar 1 — Copywriting (3/4).** Header title, hint, footer status line (dynamic `GetVersion()`-sourced,
not hardcoded — `SettingsDialog.tsx:39-45`), footer close button ("Close settings", not the generic "Done"
the handoff literally specified), section labels, segment labels, and three of four toggle notes all match
the contract verbatim. The one deviation (fix #1) is a substantively correct fix to the spec's own copy
that was applied without updating the spec.

**Pillar 2 — Visuals (4/4).** Theme grid enforces exactly-one-active via `theme.id === activeThemeId` with
no code path that can leave zero or multiple selected. Segmented controls use `aria-checked` and roving
`tabIndex` correctly. Icon-only controls (`×`, gear) both carry `aria-label`. Polish note (not scored down):
`Toolbar.tsx` uses two different labels for the two Settings entry points — "Settings" on the gear (line 166)
and "Open settings" on the theme chip (line 147). Functionally harmless; worth aligning for screen-reader
consistency.

**Pillar 3 — Color (4/4).** `--ac`/`--sel`/`--onac`/`--ch`/`--fn` usage in `workspace.css:1486-1656` matches
the spec's reservation table exactly. No new destructive color introduced — correctly, since this phase has
no destructive actions.

**Pillar 4 — Typography (4/4).** Sizes (15px/600, 11px mono, 12.5px/500, 12px, 11.5px) all trace to
already-declared scale entries; no untracked new size found via grep.

**Pillar 5 — Spacing (4/4).** Verified `.ws-settings-panel` (660×700), `.ws-settings-header`/`.ws-settings-footer`
(50/56px), `.ws-theme-swatch` (9×18px), `.ws-seg` (3px padding/gap), `.ws-rail-seg .ws-seg-option`
(4px 14px vs the density row's inherited 4px 12px) — all literal values match line-for-line.

**Pillar 6 — Experience Design (3/4).** Grep confirms zero Save/Cancel button markup and zero dirty-state
indicator anywhere in the dialog; every setter writes through `settingsStore.ts` synchronously with no
debounce (`SettingsDialog.tsx`'s `onChange` handlers call the Wails setter in the same handler as the local
dispatch). The `useModalBehavior` hook is reused unchanged — no bespoke focus-trap/Escape/scroll-lock code
in `SettingsDialog.tsx`. Held to 3/4 for the two live-untested branches in fix #2.

---

## Registry Safety

Not applicable — `shadcn_initialized: false`, no component registry used this phase (confirmed in
`26-UI-SPEC.md`'s own Registry Safety table).

---

## Files Audited

- `frontend/src/components/workspace/SettingsDialog.tsx`
- `frontend/src/components/workspace/settings/SegmentedControl.tsx`
- `frontend/src/components/workspace/settings/ThemeGrid.tsx`
- `frontend/src/components/workspace/settings/CatalogSettingsSection.tsx`
- `frontend/src/components/workspace/create/OptionsToggles.tsx`
- `frontend/src/components/workspace/Toolbar.tsx`
- `frontend/src/components/workspace/WorkspaceShell.tsx`
- `frontend/src/workspace.css` (lines 1360-1656)
- `frontend/src/themes.ts`
- `.planning/phases/26-settings/26-UI-SPEC.md`, `26-CONTEXT.md`, `26-0{1..5}-SUMMARY.md`
