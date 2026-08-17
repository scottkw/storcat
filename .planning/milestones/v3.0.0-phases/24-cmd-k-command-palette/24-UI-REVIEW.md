# Phase 24 — UI Review

**Audited:** 2026-08-14
**Baseline:** `24-UI-SPEC.md` (design contract, checker sign-off still shows "Approval: pending" — noted as a process gap, not re-litigated here)
**Screenshots:** Not captured this pass — code audit against the spec, cross-referenced with the live-verified `wails dev` evidence already on record in the task brief (focus trap, scroll lock, 50-cap/truncation, no-wrap keyboard clamp, XSS escape, centered-scroll delta). Live evidence is cited inline below where it substitutes for a fresh screenshot; it does not cover the two new findings below (scrim transition, aria-label collision), which were not part of that prior verification pass.

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 4/4 | Every locked string (PLT-03, PLT-06, placeholder formula, footer segments, hint/searching copy) matches verbatim, including the WR-04-fixed truncation line |
| 2. Visuals | 3/4 | The contract's mandatory "100ms scrim fade, no movement" transition is entirely absent — the panel hard-cuts in/out with no CSS transition anywhere in the palette rule set |
| 3. Color | 4/4 | Accent reserved exactly for magnifier/highlight/chip as specified; no hardcoded colors beyond the one explicitly-permitted scrim literal |
| 4. Typography | 4/4 | All sizes/weights match the locked scale exactly; implementation correctly follows the more specific Palette Contract section over a self-contradicting Typography-table line (see finding below) |
| 5. Spacing | 4/4 | All literal geometry (50px/32px/8px/74px/640px/15vh) matches the spec exactly, sourced from `--hp` density token where required |
| 6. Experience Design | 3/4 | Strong state machine (debounce, stale-response guard, no-wrap keyboard clamp, focus trap/restore) undercut by two real accessibility defects: a dangling `aria-controls` reference and a three-way `aria-label` collision |

**Overall: 22/24**

---

## Top 3 Priority Fixes

1. **No scrim/panel transition exists at all** — `CommandPalette.tsx:171` (`if (!isOpen) return null`) means the component isn't in the DOM until the instant it's open, and `.ws-palette-scrim`/`.ws-palette-panel` (`frontend/src/workspace.css:396-416`) declare no `transition`/`animation` property. The UI-SPEC's Palette Contract is explicit: "Transition: 100ms scrim fade, no movement — exactly the handoff's documented rule." Today the palette snaps in and out with a hard cut, not a fade. **Fix:** keep the component mounted through an `isClosing`/CSS-class transition state (or a `useState` that tracks a fade-in class applied on the frame after mount) and add `transition: opacity 100ms` to `.ws-palette-scrim`, opacity 0→1 on open. This is a locked contract line, not a nice-to-have.

2. **Three interactive elements share one accessible name: "Search every catalog"** — the toolbar trigger button (`Toolbar.tsx:62`), the palette's `role="dialog"` panel (`CommandPalette.tsx:205`), and the palette's `role="combobox"` input (`CommandPalette.tsx:233`) all carry the identical `aria-label`. A screen-reader user opens the palette by activating "Search every catalog" (button), lands in a dialog also announced as "Search every catalog," then tabs/starts typing into a combobox announced as "Search every catalog" a third time — three consecutive identical announcements with no differentiation between the *trigger*, the *dialog region*, and the *text field*. This was flagged in the audit brief as deferred and explicitly asked to be graded here: it is a real WCAG 4.1.2-adjacent labeling defect (redundant/non-distinguishing accessible names), not just a style nit. **Fix:** give the dialog a distinct label (e.g. `aria-label="Search palette"` or reference the input via `aria-labelledby` pointing at visually-hidden text), and drop the input's own `aria-label` (the combobox role + placeholder + dialog context already describe it) or make it read something like "Search query" instead of restating the trigger's label.

3. **`aria-controls` on the combobox references a nonexistent element outside the Results state** — `CommandPalette.tsx:227` sets `aria-controls={PALETTE_LISTBOX_ID}` unconditionally, but `id={PALETTE_LISTBOX_ID}` only exists on the `PaletteResultList` mounted inside the `isResults` branch (`CommandPalette.tsx:245`). In the Hint, Searching, and No-matches states — three of the four mutually-exclusive states — the input points assistive tech at a DOM id that isn't there. This is recorded in `24-REVIEW.md` as IN-01 and was left unfixed (Info severity was out of scope for the code-review fix pass, but it is squarely inside this UI review's accessibility remit). **Fix:** `aria-controls={isResults ? PALETTE_LISTBOX_ID : undefined}`, one line.

---

## Detailed Findings

### Pillar 1: Copywriting (4/4)

- `"Type to search…"` — header readout (`CommandPalette.tsx:189`) and body state (`CommandPalette.tsx:240`) both exact. ✓
- `"Searching…"` — header readout (`:191`) and body state (`:242`) both exact, and correctly gated by the `settled` flag so only the *first* in-flight query of a session shows it (`:32, :87`), matching the spec's "no flash on subsequent keystrokes" requirement. ✓
- `"{total} hits"` / `"50 of {total}"` — implemented as `total > results.length ? \`${results.length} of ${total}\` : \`${total} hits\`` (`:192-194`). Post-WR-04 this reads exactly `"50 of {total}"` whenever the cap is hit (`results.length` is always 50 at cap today) while being immune to future cap drift. Matches the spec's literal string today; verified correct by construction. ✓
- PLT-03 truncation — `` `Showing the first ${results.length} of ${total} hits` `` (`PaletteResultList.tsx:67`) — exact copy, no trailing clause, matches the locked "no full-table view" simplification from `24-CONTEXT.md`. ✓
- PLT-06 empty state — `"No file in any catalog matches that."` (`CommandPalette.tsx:255`) — character-for-character exact. ✓
- Footer — `"↵ reveal in catalog"` / `"esc close"` / `"searches names and paths"` (`:258-260`) — exact, correct left/left/right layout via `margin-left: auto` on the third span. ✓
- Placeholder — `` `Search ${formatCount(filesIndexed)} files across ${catalogCount} catalogs…` `` (`:183`), where `filesIndexed` is summed from `state.catalogs[].fileCount` (`:178-181`) — this is the **same derivation** `StatusBar.tsx:25-26` uses for its own "files indexed" figure, so the two numbers stay consistent app-wide rather than computing a second, possibly-divergent total. "catalogs" is unconditionally plural per the spec's convention. ✓
- No generic `Submit`/`OK`/`Click Here` labels anywhere in the reviewed files (`grep` clean).

### Pillar 2: Visuals (3/4)

- Focal point / hierarchy: magnifier → input → readout → flat result rows → footer reads as a single clear vertical flow; basename (12.5px, weight 400/600 on match) is the clear primary text, path (11px, dimmed) subordinate, chip and size flanking at equal 11px weight — correct hierarchy via size/weight/color, matches `23-UI-SPEC.md`'s established row-shape convention reused verbatim (`PaletteResultRow.tsx:48`, directory=square/`--ac`, file=circle/`--fn`).
- Icon-only elements: the magnifier SVG carries `aria-hidden="true" focusable="false"` (`CommandPalette.tsx:216-217`) and sits beside a labeled input rather than standing alone as an icon-only control — correctly not a bare icon button needing its own label.
- **Defect:** the contract's "100ms scrim fade, no movement" transition does not exist in any form. `CommandPalette.tsx` unmounts the whole tree on close (`if (!isOpen) return null`, line 171) and workspace.css's `.ws-palette-scrim`/`.ws-palette-panel` rules (lines 396-416) declare no `transition` or `animation` property at all — confirmed by `grep -n "transition\|animation"` against the palette CSS block returning zero hits. The open/close is a hard, un-animated cut. This is called out as a WARNING rather than a BLOCKER because it doesn't break the task (the palette still opens/closes/searches/reveals correctly) but it is a named, specific line in the locked contract that was simply not built.
- No overuse of accent color observed within the panel — accent is confined to the magnifier, the match highlight, and the chip text, per spec.

### Pillar 3: Color (4/4)

- Magnifier: `stroke="var(--ac)"`, 15×15, `strokeWidth 1.6` (`CommandPalette.tsx:209-221`) — matches spec exactly, correctly distinct from the toolbar's 13px `--dm` magnifier.
- Match highlight: `color: var(--ac); fontWeight: 600` on the matched substring only, basename-only, first-occurrence-only via `indexOf` (`PaletteResultRow.tsx:22-36`) — path is never touched by the highlight span, matches the "path stays plain-dimmed" lock from `24-CONTEXT.md`.
- Catalog chip: `.ws-palette-chip { color: var(--ac); background: var(--acs); border-radius: 5px; padding: 2px 7px }` (`workspace.css:497-504`) with an explicit code comment recording *why* it's a sibling class rather than a repurposed `.ws-chip` — the divergence the UI-SPEC pre-declared as deliberate was honored exactly as written, including "deliberately no `:hover` rule."
- `--hov` shared identically between mouse-hover (`.ws-palette-row:hover`) and keyboard-active (`.ws-palette-row[data-active]`) (`workspace.css:454-460`) — single shared visual, matches the spec's explicit "the two states share one visual" instruction.
- Scrim literal `rgba(4, 6, 9, .62)` (`workspace.css:399`) — matches the spec's declared literal exactly, correctly distinct from the `.ws-backdrop` drawer's `rgba(0,0,0,.35)` (line 392, four lines above).
- No `#e5534b` (destructive) usage anywhere in the reviewed files — confirmed by grep.

### Pillar 4: Typography (4/4)

- 14px mono input (`workspace.css:435-436`) ✓, basename 12.5px mono (`:479`) ✓, path/chip/size all 11px (`:487, :498, :507`) ✓, footer 10.5px (`:539`) ✓, truncation 11.5px (`:518`) ✓, hint/no-match state 12.5px (`:526`) ✓ — every one of these matches the Palette Contract section's per-element sizes exactly.
- **Note, not a defect:** the UI-SPEC's own Typography table (line 61) states "11.5px | Truncation notice line, 'Type to search…' / 'No file in any catalog matches that.' states" — bundling the truncation line *and* the two body-copy states into one 11.5px row — while the more detailed Palette Contract section (lines 127, 130) separately specifies 12.5px for the Hint/No-matches body copy and only the truncation line at 11.5px (line 129). These two sections of the spec disagree with each other. The implementation follows the more specific, later section (12.5px body copy, 11.5px truncation only) — a reasonable resolution of an internal spec inconsistency, and the one a checker would most likely have caught had the sign-off section not still read "pending." Flagging for the spec's own hygiene, not as an implementation defect.
- No new weights: match-highlight is the existing 600 (Semibold), everything else 400 — confirmed by grep, no `font-weight: 500` or other value introduced in palette rules.

### Pillar 5: Spacing (4/4)

- Panel: `width: min(640px, calc(100vw - 32px)); max-height: min(520px, 70vh); border-radius: 12px` (`workspace.css:407-409`) — exact match to the CONTEXT-locked 640px/15vh divergence from the 720px/96px handoff value, correctly implemented as a proportional `padding-top: 15vh` on the scrim (`:402`), not a fixed pixel value.
- Input row 50px (`:419`), footer 32px (`:531`), result-row shape 8px with 8px gap (`PaletteResultRow.tsx` + `workspace.css:449, 463-464`), size column fixed 74px right-aligned (`:511-512`) — every literal geometry value in the spec's table is present and correct.
- Result-row padding uses `var(--hp)` (`workspace.css:450`), the density token, not a hardcoded value — correctly inherits Compact/Comfortable density from Phase 22's token layer rather than reinventing row padding for this surface.
- No arbitrary/unexplained spacing values found in the palette CSS block beyond the ones the spec explicitly declares as literal-geometry exceptions.

### Pillar 6: Experience Design (3/4)

**Strengths, verified either by code or by the live-verified evidence cited in the audit brief:**
- Debounce (200ms, `PALETTE_DEBOUNCE_MS`, `CommandPalette.tsx:14, 75`) with a stale-response guard via a monotonic `requestIdRef` (`:40, 74, 77`) — matches the 8-char-burst → 1-call live verification already on record.
- Keyboard clamp with no wrap at both boundaries (`:138, 142, 146-158`) — matches the 60×ArrowDown-stays-at-49 live verification.
- `useModalBehavior` correctly implements all four contracted behaviors (focus trap, Escape, scroll lock keyed on `.ws-root`, focus restore) with the `isOpen`-transition-not-unmount design Phase 25's animated exit will depend on (`useModalBehavior.ts:54-142`) — the two post-review WR-01/WR-02 fixes (`MERGE_EXPANDED`, `REVEAL_HIT`) close the ancestor-collapse and dispatch-order landmines the code review found, and both were live-reverified per the fix report.
- XSS: highlight built from three JSX text children, never `dangerouslySetInnerHTML` (`PaletteResultRow.tsx:29-35`) — matches the `<b>`-literal live verification.
- Zero-catalog / empty-library case is honest, not disabled (`E4` in the spec) — the trigger stays live and typing yields the correct PLT-06 copy rather than hiding the affordance.

**Defects:**
- **`aria-controls` dangling reference (IN-01, unfixed)** — `CommandPalette.tsx:227` unconditionally points at `PALETTE_LISTBOX_ID`, which only exists in the DOM during the Results state (`:245`). In the other three states this is an invalid ARIA reference. Left out of scope during the code-review fix pass (Info severity); in scope for this UI review's accessibility grading. See Top-3 fix #3.
- **Aria-label collision across trigger/dialog/combobox** — three distinct controls (`Toolbar.tsx:62`, `CommandPalette.tsx:205`, `CommandPalette.tsx:233`) share the exact string "Search every catalog." This was explicitly called out in the audit brief as deferred to this review; graded here as a genuine WARNING-level accessibility defect (redundant, non-distinguishing accessible names for three different interaction targets), not a cosmetic nit. See Top-3 fix #2.
- **IN-02, zero-results vs. failed-search render identically (spec-locked, judged acceptable)** — `CommandPalette.tsx:76-89` folds a failed `SearchIndexed` call into the same UI as a genuine zero-match. This is explicitly locked in UI-SPEC E1/E3 with a reasoned justification (the search walks already-known local files; a hard failure is expected to be unreachable in normal operation). Judgment: the reasoning holds for this phase's actual failure surface (no network, no auth, purely local disk I/O against paths the rail already proved exist) — this is a defensible, deliberate simplification rather than a silent-fallback anti-pattern in the sense `CLAUDE.md` warns against, since the *reason* it's safe is documented and the failure mode is genuinely unreachable in the phase's scope, not merely unhandled. Not scored as a defect, but flagged for re-examination if the search binding's failure surface ever grows (e.g. network-backed catalogs in a later milestone).
- **Ctrl+K on Windows/Linux unverified** — `WorkspaceShell.tsx:60` checks `event.metaKey || event.ctrlKey`, which is the correct cross-platform pattern in code, but per the audit brief this has not been exercised on an actual Windows/Linux machine (WINDOWS.md ledger #2). Not scored as a defect since the code pattern is standard and correct; flagged as an open verification gap, not a UI defect.

---

## Files Audited

- `/Users/ken/dev/storcat/frontend/src/components/workspace/CommandPalette.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/palette/PaletteResultRow.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/palette/PaletteResultList.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/Toolbar.tsx`
- `/Users/ken/dev/storcat/frontend/src/components/workspace/WorkspaceShell.tsx`
- `/Users/ken/dev/storcat/frontend/src/hooks/useModalBehavior.ts`
- `/Users/ken/dev/storcat/frontend/src/workspace.css` (palette rule block, lines 396-541)
- `/Users/ken/dev/storcat/.planning/phases/24-cmd-k-command-palette/24-UI-SPEC.md`
- `/Users/ken/dev/storcat/.planning/phases/24-cmd-k-command-palette/24-CONTEXT.md`
- `/Users/ken/dev/storcat/.planning/phases/24-cmd-k-command-palette/24-REVIEW.md`
- `/Users/ken/dev/storcat/.planning/phases/24-cmd-k-command-palette/24-REVIEW-FIX.md`
- `/Users/ken/dev/storcat/.planning/phases/23-rail-virtualized-tree/23-UI-SPEC.md` (cross-reference only, Status Bar Contract)

---

## Remediation — 2026-08-14

All three priority fixes were applied and verified live at `wails dev` :34115 in commit `f7994ab5`.

| # | Finding | Class | Disposition | Evidence |
|---|---------|-------|-------------|----------|
| 1 | Mandatory "100ms scrim fade, no movement" transition absent (`workspace.css`, `CommandPalette.tsx`) | **Contract violation** | **Fixed** | `@keyframes ws-palette-scrim-in` + `animation: … 100ms ease-out` on `.ws-palette-scrim`, opacity only so the panel stays unanimated per contract. Observed live: `animationName="ws-palette-scrim-in"`, `animationDuration="0.1s"`. A `prefers-reduced-motion: reduce` opt-out was added — not required by the contract, but the project had no motion at all before this, so this is the first rule that needed one. |
| 2 | Three controls sharing `aria-label="Search every catalog"` | a11y defect | **Fixed** | Dialog is now `aria-label="Search palette"`; the combobox input keeps `"Search every catalog"` (it names what you type into it). Observed live: `dialogLabel="Search palette"`, `inputLabel="Search every catalog"`. The toolbar trigger retains its original name and never coexists with the input in the a11y tree, since the dialog is `aria-modal="true"`. |
| 3 | `aria-controls` referencing a listbox id absent outside the Results state (IN-01) | a11y defect | **Fixed** | Now `aria-controls={isResults ? PALETTE_LISTBOX_ID : undefined}`. Observed live: `null` in the hint state, `"ws-palette-listbox"` resolving to a real element in the Results state, `null` in the empty state. |

**Regression check after remediation:** `npx tsc --noEmit` and `npm run build` green. Verbatim copy re-confirmed unchanged — `Showing the first 50 of 400 hits`, `No file in any catalog matches that.`, `Type to search…` — with 50 rows still rendered under the cap.

### Not remediated (deliberate)

- **IN-02 — a failed search renders identically to a genuine zero-match result.** Left as-is: this is spec-locked in `24-UI-SPEC.md` E1 ("a failed search binding call degrades to the zero-result branch rather than a distinct error UI"). Changing it here would contradict the locked contract; revisiting it is a design decision for a future phase, not a review fix. Noted as residual risk.
- **Ctrl+K unverified on Windows/Linux** — no machine available. Tracked as WINDOWS.md ledger entry #2 for the pre-v3.0.0 sweep.
- **UI-SPEC internal Typography-table inconsistency** — a defect in the spec document, not in the implementation. Recorded for the milestone-level design-token audit the Phase 24 UI checker also recommended (alongside the >4-font-size and non-4px-grid conventions inherited from Phase 22).

**Post-remediation score: 24/24.** Both previously-docked pillars (Visuals 3/4 for the missing transition, Experience Design 3/4 for the accessible-name collision and dangling `aria-controls`) now have their findings closed.
