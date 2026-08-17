# Phase 25 — UI Review

**Audited:** 2026-08-14
**Baseline:** `25-UI-SPEC.md` (approval: pending in the source file's own sign-off block — noted as a process gap, not a defect this audit introduces)
**Screenshots:** captured (base workspace state only, via `wails dev` at `localhost:34115`, 1440×900 — `.planning/ui-reviews/25-20260814-225028/desktop.png`). The four create-slide-over states (form/scanning/error/done) were not screenshotted interactively in this pass (no scripted browser automation available in this environment beyond the one-shot `npx playwright screenshot` CLI); those states were verified by direct source/CSS audit instead, cross-checked against `25-REVIEW.md`'s and `25-REVIEW-FIX.md`'s recorded live-`wails dev` verification evidence (which this audit treats as evidence, not as a substitute for its own code-level checks).

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 4/4 | All locked strings verified character-for-character in source, including the two review-mandated changes ("Discard and close", "Close without writing"); E6's unresolved status is honestly carried forward, not quietly resolved. |
| 2. Visuals | 3/4 | Hierarchy, icon-only `aria-label`s, and state badges all correct; one real defect — the 34px vs. 32px outlined-button height distinction the spec explicitly calls "not a rounding error" is collapsed to a single 32px class, silently shrinking "Retry scan" and "Catalog another volume" by 2px. |
| 3. Color | 4/4 | `--onac` correction on the toggle knob is implemented exactly as specified (not the hardcoded handoff hex); all status colors are the correct theme-independent/theme-dependent split; no stray hardcoded hex outside the declared exceptions. |
| 4. Typography | 3/4 | Sizes and weights match the declared scale/roles; 11 distinct sizes in `workspace.css` is Phase 22's already-declared systemic exception, not new debt, per this audit's instructions — scored 3 rather than 4 only because of the same button-height/type-pairing slip noted under Visuals (12.5px outline-button label spec-correct, but sharing one height value across two contractually-different heights). |
| 5. Spacing | 3/4 | Full component-geometry table (52/62/34/32/30×20/30/30×17/13/6/3/300/8px) spot-checked against CSS and matches, with the single exception of BLOCKER-adjacent finding CRT-B1 below (32px applied where 34px is contractually required in two of three outlined-button contexts). |
| 6. Experience Design | 4/4 | CRT-01's full animation contract (340ms in / 260ms out, no early unmount, all five close paths routed through one handler, `prefers-reduced-motion` disables both) verified present in both CSS and component logic; no-spinner rule honored in both scanning sub-states; `useModalBehavior` consumed unchanged, called with real `isOpen` per the load-bearing E6 note; all four post-review fixes (CR-01/CR-02/WR-01/WR-02) confirmed landed in current source. |

**Overall: 21/24**

---

## Top 3 Priority Fixes

1. **Outlined secondary buttons are 32px everywhere, but the contract requires 34px for two of the three contexts.** `.ws-create-btn-outline` (`frontend/src/workspace.css:962-972`) is a single shared class at `height: 32px`, used by `ScanningBody.tsx`'s "Run in background" (spec-correct: intentionally 2px shorter), `ErrorBody.tsx`'s "Retry scan" (`ErrorBody.tsx:104`), and `DoneBody.tsx`'s "Catalog another volume" (`DoneBody.tsx:105`) — the latter two are spec'd at 34px ("outlined, 34px... transcribed exactly from the handoff; not a rounding error" — 25-UI-SPEC.md lines 42, 252, 263). User impact: the primary/outline button pair in the error and done footers now sit 2px out of vertical alignment with the primary button beside them, contradicting a distinction the spec's authors specifically flagged as deliberate and load-bearing. Fix: split the class — keep `.ws-create-btn-outline` (or rename it `.ws-create-btn-outline-32`) at 32px for the scanning-footer "Run in background" button only, and add a `.ws-create-btn-outline-34` (or equivalent modifier) at `height: 34px` for `ErrorBody`'s "Retry scan" and `DoneBody`'s "Catalog another volume".

2. **`25-UI-SPEC.md`'s own sign-off block is still `pending`.** All six checker dimensions at the bottom of the spec (`25-UI-SPEC.md:456-463`) are unchecked and "Approval: pending." This is a process gap rather than an implementation defect — the built UI substantially matches the contract — but it means this phase's design contract was never formally closed out before implementation proceeded. Fix: run the checker pass and update the sign-off block, or explicitly document why it was skipped.

3. **No interactive screenshot evidence exists for the four create-slide-over states in this repo's UI-review history for this phase.** `25-REVIEW.md`/`25-REVIEW-FIX.md` (code review) and `25-VALIDATION.md`/manual `wails dev` sessions (per those documents) cover functional/live verification, but no dated screenshot artifact of the form/scanning/error/done states exists under `.planning/ui-reviews/`. This audit was code-only for those four states because no scripted browser-automation tool was available in this pass beyond a single static screenshot of the base workspace. Fix: capture the four states (ideally via `dev-browser` or an equivalent scripted flow) at least once and archive under `.planning/ui-reviews/` so future audits have a visual baseline, not just a source-code one.

---

## Detailed Findings

### Pillar 1: Copywriting (4/4)

- Verified every row in the Copywriting Contract table against source, including the two deliberately-changed strings the audit notes flagged as high-risk for reversion:
  - `"Discard and close"` — `CreateSlideOver.tsx:480` (exact, not "Cancel").
  - `"Close without writing"` — `ErrorBody.tsx:119` (exact, not "Cancel").
- Header title/step table (`"New catalog"`/`"step 1 of 2"`, `"Cataloguing volume"`/`"step 2 of 2"`, `"Scan interrupted"`/`"failed"`, `"Catalog written"`/`"done"`) matches `CreateSlideOver.tsx:369-370` exactly.
- Error headline `"Stopped at {pct}% — the volume went away"` (`ErrorBody.tsx:44`) and explanation paragraph (`ErrorBody.tsx:73-75`, verbatim) both match. The counting-sub-state headline extension (`"Stopped — the volume went away"`, no `{pct}%` slot) is a reasoned, disclosed extension of the locked template for a gap the spec itself acknowledges (E6 unresolved), not an unlogged deviation.
- Done-state strings (`"{catalogTitle} catalogued"`, `"partial"` tag, both doneLine flavors) match `DoneBody.tsx:41-47, 59, 66`.
- Disabled-entry-point tooltip `"A scan is already running — open it from the status bar."` matches verbatim in both `CatalogRail.tsx:59` and `TreePane.tsx:155`.
- Status-bar segment copy (`"● scanning {name} · {pct}%"` / `"● scanning {name} · counting…"`) matches `StatusBar.tsx:77-79`.
- E6's unresolved status is honestly preserved in code: `ErrorBody.tsx:13-33`'s doc comment explicitly restates "this component's state coverage is an explicit, re-derived assumption, not a machine-checked one" rather than silently treating it as resolved — this is exactly the audit-notes' instruction to verify it wasn't quietly closed out.
- No generic "Cancel"/"Submit"/"OK" strings found anywhere in `components/workspace/create/` or `CreateSlideOver.tsx` (`grep` confirmed zero matches).

### Pillar 2: Visuals (3/4)

- Clear focal point in each state: volume-card list + WILL WRITE preview (form), title/percentage/progress bar (scanning), round `!` badge + headline (error), round `✓` badge + title (done) — all present and correctly weighted per `CreateSlideOver.tsx`/component render trees.
- Icon-only controls carry accessible names: header `×` has `aria-label="Close"` (`CreateSlideOver.tsx:409`); the folder-picker SVG icon in `VolumePicker.tsx:99-113` is `aria-hidden="true"` inside a card whose text content (name + path) already supplies the accessible name, which is correct (the icon is decorative, not the sole label).
- Toggle rows correctly implement `role="switch"` + `aria-checked` + `aria-disabled` (`OptionsToggles.tsx:41-58`) — a real interactive-control pattern, not a div styled to look like one.
- **Defect (CRT-B1):** the 34px-vs-32px outlined-secondary-button distinction the spec calls out twice as deliberate ("not a rounding error," 25-UI-SPEC.md line 42; "transcribed exactly from the handoff," same context) is not implemented — see Top 3 Fix #1. This is a visual-hierarchy regression relative to the contract: two of the three outlined buttons in this phase render 2px shorter than their spec'd height, subtly misaligning them against the adjacent 34px primary buttons in the error/done footers.
- Volume-card selected/unselected states (`--sel`/`--ac` border vs. `--p2`/`--l`) are wired correctly in CSS (spot-checked `ws-create-vol-card-selected` alongside the badge/tag rules).

### Pillar 3: Color (4/4)

- The `--onac` correction (spec's explicit fix to the handoff's hardcoded `#06232a` knob color) is implemented exactly as required: `workspace.css:1313` uses `var(--onac)` on `.ws-create-toggle-track-on .ws-create-toggle-knob`, not a literal hex — this was the single highest-risk color item called out in the spec ("the same derived-token fix... every other accent-adjacent text/icon color") and it is correctly applied.
- Theme-independent fixed colors (`#e5534b` error badge/log, `#f0b429` warning tags) are used only where declared — `workspace.css:1021-1024` (badge-error), `ws-create-tag-warn`/`ws-create-tag-partial` (grep-confirmed reuse, no new hex introduced for the partial tag, matching the "reuses the exact warning-tag styling... no new color" contract).
- Success color (`✓` badge) correctly uses `var(--ac)` on `var(--acs)`, theme-dependent as specified (`workspace.css:1026-1029`).
- No hardcoded hex found inside the create-slide-over CSS block beyond the two theme-independent exceptions explicitly declared in the spec.

### Pillar 4: Typography (3/4)

- Font sizes activated by this phase (11/11.5/12/12.5/13/14/14-mono/15px) all present and correctly role-assigned: 11px uppercase section labels (`ws-create-label`), 11.5px mono counters/paths, 12.5px form-input values and secondary-button labels, 13px volume-card name/primary-button label, 14px scanning title/percentage/error-done headline, 15px header title/badge glyph.
- The 26px-Display-not-activated correction is honored — no 26px usage found anywhere in the create-slide-over CSS.
- No new weight introduced beyond 400/500/600 (spot-checked; consistent with the rest of `workspace.css`).
- `workspace.css` overall carries 11 distinct `font-size` values — well past the ">4 sizes" generic flag threshold, but per this audit's explicit instruction this is Phase 22's already-declared, carried-forward systemic exception, not new debt introduced by Phase 25, and is not counted as a fresh finding.
- Docked one point (not a full pass) because the same button-height slip noted under Visuals also touches a typography/geometry pairing: the 12.5px label size is correctly applied to all three outlined buttons, but two of those three buttons are paired with the wrong contractual height for that role, which is a coupled visual+spacing regression best flagged once rather than three times — see CRT-B1.

### Pillar 5: Spacing (3/4)

- Full component-geometry spot-check against `25-UI-SPEC.md`'s table, all confirmed present and correct in `workspace.css`: 560px panel width, 52px header (`:693`), 62px footer (`:764` area), 34px form inputs/primary buttons (`.ws-create-btn { height: 34px }`), 30×20px volume-card chip, 30px round badges, 30×17px toggle track / 13px knob, 6px/3px progress-bar track/radius, 300px log max-height, 8px written-file-row shape.
- Slide-over panel radius correctly unset (flush against the viewport edge) and shadow value (`-30px 0 70px rgba(0,0,0,.5)`) correctly reused from `22-UI-SPEC.md`'s declared "Right slide-over" row, not re-derived.
- **Defect (CRT-B1, same root cause as the Visuals/Typography findings above):** the 32px "Run in background" height is correctly implemented, but that same 32px value is incorrectly reused for the two buttons contractually specified at 34px ("Retry scan," "Catalog another volume") — `workspace.css:962-972` is a single un-split class. This is the pillar where the defect is most directly a spacing-scale violation (a literal, spec'd pixel value not honored), so it is weighted here as the primary score deduction.
- Z-index reuse confirmed correct: no new z-index value introduced; `.ws-create-scrim` uses `var(--z-overlay)` (`workspace.css:656`), matching the locked 200 value shared with the palette.

### Pillar 6: Experience Design (4/4)

- **CRT-01 animation contract fully verified, not just described:** both keyframe pairs (`ws-create-scrim-in`/`-out`, `ws-create-slide-in`/`-out`) exist in `workspace.css:630-648` with the exact spec'd timings (340ms in / 260ms out on the panel, `cubic-bezier(.16,.84,.24,1)` / `cubic-bezier(.4,0,.7,.2)`; 200ms/260ms scrim). `prefers-reduced-motion: reduce` correctly disables all four animations plus the walking-path pulse (`workspace.css:682-690`).
- **No early unmount, across all five close paths:** `CreateSlideOver.tsx:362`'s render condition (`if (!(isOpen || closing)) return null;`) combined with the `useLayoutEffect`-driven `closing` flag (lines 77-100) keeps the panel mounted for the full 260ms exit regardless of which of the five paths (Escape, `×`, "Discard and close," scrim click, "Open in workspace") triggered the close — all five route through the single `handleCloseRequest`/`onClose` function, confirmed by grep across all five call sites.
- `useModalBehavior` is called with the real `isOpen` (line 129), not `isOpen || closing` — the load-bearing distinction the spec calls out by name (E6's cross-phase note) is correctly honored, and the hook itself (`useModalBehavior.ts`) is consumed unmodified, with no bespoke reimplementation of focus trap/Escape/scroll-lock/focus-restore found anywhere in `CreateSlideOver.tsx` or its children.
- No-spinner rule verified in both scanning sub-states: `ScanningBody.tsx:40-44` conditionally omits the progress-bar div entirely when `isCounting` is true (not a hidden/zero-width bar), and the counting-sub-state counter line (`"{seen} files found so far"`) is a live, real number with no bytes/ETA fabricated — matches the contract precisely.
- All four post-review fixes (CR-01 `copyFile` atomic-write, CR-02 `WritePartialCatalog` TOCTOU + frontend guard, WR-01 stale-closure ref, WR-02 disabled Retry/Close during in-flight partial write) confirmed present in current source: `handleCreateRef` pattern at `CreateSlideOver.tsx:335-336, 353`; `disabled={writingPartial}` on both "Retry scan" and "Close without writing" at `ErrorBody.tsx:104, 116`; `writingPartialRef` defense-in-depth guard in `handleCreate` at `CreateSlideOver.tsx:191`.
- Entry-point disabling verified correct and consistent: both `CatalogRail.tsx` and `TreePane.tsx` gate their respective affordances on `isScanningNow` with matching `aria-disabled`/dimmed/`title` treatment and identical tooltip copy.

---

## Files Audited

- `frontend/src/components/workspace/CreateSlideOver.tsx`
- `frontend/src/components/workspace/create/VolumePicker.tsx`
- `frontend/src/components/workspace/create/CreateForm.tsx`
- `frontend/src/components/workspace/create/OptionsToggles.tsx`
- `frontend/src/components/workspace/create/ScanningBody.tsx`
- `frontend/src/components/workspace/create/ErrorBody.tsx`
- `frontend/src/components/workspace/create/DoneBody.tsx`
- `frontend/src/components/workspace/StatusBar.tsx`
- `frontend/src/components/workspace/CatalogRail.tsx`
- `frontend/src/components/workspace/TreePane.tsx` (entry-point wiring only)
- `frontend/src/hooks/useModalBehavior.ts`
- `frontend/src/lib/scanFormat.ts`
- `frontend/src/workspace.css` (create-slide-over rule blocks, lines ~626-1320; status-bar segment rules ~590-624)
- `.planning/phases/25-create-slide-over-progress-cancellation-partial-catalog/25-UI-SPEC.md`
- `.planning/phases/25-create-slide-over-progress-cancellation-partial-catalog/25-CONTEXT.md`
- `.planning/phases/25-create-slide-over-progress-cancellation-partial-catalog/25-REVIEW.md`
- `.planning/phases/25-create-slide-over-progress-cancellation-partial-catalog/25-REVIEW-FIX.md`
- `22-UI-SPEC.md`, `23-UI-SPEC.md`, `24-UI-SPEC.md` (cross-reference for locked tokens/z-index/density/modal-hook contracts)
- Live `wails dev` instance at `localhost:34115` (base workspace-shell screenshot only; see Screenshots note above)

---

## Remediation — 2026-08-15

| # | Finding | Class | Disposition | Evidence |
|---|---------|-------|-------------|----------|
| 1 | "Retry scan" and "Catalog another volume" rendered 32px against the spec's 34px | **Contract violation** | **Fixed** | Added `.ws-create-btn-outline-34` and applied it at `ErrorBody.tsx:104` and `DoneBody.tsx:105`. **`ScanningBody.tsx`'s "Run in background" deliberately left at 32px** — `25-UI-SPEC.md` line 42 declares that height intentional ("2px shorter than the 34px primary buttons, transcribed exactly from the handoff; not a rounding error"), so normalising all three to 34px would have replaced one contract violation with another. `tsc --noEmit` + `npm run build` green. |
| 2 | UI-SPEC's checker sign-off block still reads "pending" | Process | **Acknowledged, not edited** | The block is the UI *checker's* artifact, not the auditor's or the orchestrator's. The checker did return **APPROVED** for this spec after one revision (4 PASS / 2 FLAG, both FLAGs inherited Phase 22 systemic exceptions). Back-filling another role's sign-off box would misrepresent who signed. Recorded here instead. |
| 3 | No archived visual-state screenshots under `.planning/ui-reviews/` for the four slide-over states | Coverage | **Open, recommended** | The four states were audited from source and CSS rather than captured images. A `dev-browser` pass archiving form/scanning/error/done as a visual baseline would strengthen future regression review — worth doing before ship, not blocking this phase. |

**Post-remediation score: 22/24.** Visuals rises to 4/4 with the button-height defect closed. Typography (3/4) and Spacing (3/4) remain as recorded — both are conformance to Phase 22's declared systemic exceptions (>4 type sizes; the non-4px-grid component-dimension class), not new debt from this phase, and are already carried as a recommended milestone-level design-token audit.

**Notable positive finding, restated because it is the phase's strongest signal:** CRT-01's full 340ms/260ms animation contract — both keyframe pairs, `prefers-reduced-motion` handling, no-early-unmount across all five close paths, and the load-bearing `useModalBehavior(isOpen)` rather than `useModalBehavior(isOpen || closing)` distinction — was verified present in both `workspace.css` and `CreateSlideOver.tsx`. Phase 24's UI review caught exactly this class of defect as *entirely absent* (its mandated 100ms scrim fade was specified and never implemented); Phase 25 did not repeat it.
