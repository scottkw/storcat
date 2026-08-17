---
phase: 22
slug: shell-token-layer
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: 2026-08-13
reconciled: 2026-08-13
---

# Phase 22 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

**Honest constraint up front:** this project has **no frontend test framework** (TEST-01 — Vitest + Testing Library — is explicitly deferred to a future milestone). Do NOT introduce Vitest to satisfy this phase. The automated layer below is what genuinely exists today: TypeScript type-checking, the Vite production build, the Go build, and the existing Go test suite. Everything visual is human-verified.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | None for frontend behavior. Available automated checks: `tsc` (strict, `noUnusedLocals`/`noUnusedParameters`), `vite build`, `go build`, `go test` |
| **Config file** | `frontend/tsconfig.json`, `frontend/vite.config.ts`, `go.mod` |
| **Quick run command** | `cd frontend && npx tsc --noEmit` |
| **Full suite command** | `cd frontend && npx tsc --noEmit && npm run build && cd .. && go build ./... && go test ./...` |
| **Estimated runtime** | ~30–60 seconds |

`tsc --noEmit` is the load-bearing automated check this phase: with `noUnusedLocals` and `noUnusedParameters` on, it catches the entire dead-import/dead-state surface left behind when `Header.tsx`, `MainContent.tsx`, `WelcomeContent.tsx` and `components/tabs/*` are deleted. A green `tsc` is strong evidence the removal was complete and consistent.

---

## Sampling Rate

- **After every task commit:** `cd frontend && npx tsc --noEmit`
- **After every plan wave:** full suite command above, plus a visual pass in `wails dev`
- **Before `/gsd-verify-work`:** full suite green, plus the manual matrix below completed
- **Max feedback latency:** ~60 seconds

---

## Per-Task Verification Map

Reconciled against the seven `22-0*-SUMMARY.md` `coverage` blocks and `22-VERIFICATION.md`'s live-browser pass, which superseded most `human_judgment: true` summary rows with actual dev-browser measurements. "Status" reflects what is proven today, not what was planned.

| Task ID | Plan | Requirement(s) | How Verified | Status |
|---------|------|-----------------|---------------|--------|
| 22-01 D1/D2 | 22-01 | SHELL-01, SHELL-02 | `tsc`/`build` + live browser: `.ant-tabs` absent, `.ws-toolbar` 46px, `.ws-status` 26px, grid `268px 844px 288px` @1400px (22-VERIFICATION.md) | ✅ green |
| 22-01 D3 | 22-01 | SHELL-08 | `go build`/`go test` green + real macOS build (`wails build -platform darwin/arm64`), traffic lights confirmed inside toolbar band (22-VERIFICATION.md `verified_post_report`) | ✅ green |
| 22-01 D4 | 22-01 | THEME-03 | Live browser: all 14 `:root` tokens resolve concretely per theme (22-VERIFICATION.md) | ✅ green |
| 22-01 D5 | 22-01 | THEME-01, THEME-02 | Live browser: 12x `Ctrl+Alt+T` cycle, 11-theme wraparound confirmed; `--onac` correctness measured on Gruvbox/Monokai/GitHub (22-VERIFICATION.md) | ✅ green |
| 22-01 D6 | 22-01 | THEME-04 | Live browser: `Ctrl+Alt+D` flips `--rh` 34px↔27px (22-VERIFICATION.md) | ✅ green |
| 22-01 D7 | 22-01 | THEME-06 | Live browser: preset dark theme + reload shows tokens applied pre-paint, no flash (22-VERIFICATION.md) | ✅ green |
| 22-01 D8 | 22-01 | — (dev affordance) | `npm run build` + `grep -l storcat-dev-switcher dist/assets/*.js` → 0 matches | ✅ green |
| 22-02 D1–D4 | 22-02 | THEME-05 | `tsc`/`build` green; byte-size/magic-byte checks on 5 woff2 files; zero `package.json`/`vite.config.ts` diff; build asset count | ✅ green |
| 22-02 D5 | 22-02 | THEME-05 | Superseded by 22-07 D8's live dev-browser network trace (4 same-origin font requests, 0 external) — see Manual-Only THEME-05 row | ✅ green (superseded) |
| 22-03 D1 | 22-03 | SHELL-01 | `tsc`/`build` + zero-antd/zero-onClick/zero-zIndex grep; live browser confirms toolbar band and controls (22-VERIFICATION.md) | ✅ green |
| 22-03 D2 | 22-03 | SHELL-07 | `.no-drag` on 4 controls confirmed by grep and by `getComputedStyle` (`--wails-draggable: no-drag`) in-browser; native drag arbitration itself remains unverified — see Manual-Only SHELL-07 row | ⚠️ partial (structural only; accepted, see below) |
| 22-03 D3 | 22-03 | SHELL-08 | `tsc`/`build` + grep confirms platform gate logic; real macOS build confirms visually (22-VERIFICATION.md) | ✅ green |
| 22-04 D1 | 22-04 | SHELL-01 | `tsc`/`build` + Copywriting Contract string grep; live browser `page.innerText` extraction confirms verbatim (22-VERIFICATION.md, 22-UI-REVIEW.md) | ✅ green |
| 22-04 D2 | 22-04 | THEME-02 | Superseded by 22-07 D4's live hover-inversion fix and measurement across all 11 themes | ✅ green (superseded) |
| 22-04 D3 | 22-04 | SHELL-01 | Grep confirms 3 literal zero-state segments, no watch segment | ✅ green |
| 22-04 D4 | 22-04 | THEME-04 | Grep confirms no hardcoded density pixel literals; `.pane-scroll` class applied | ✅ green |
| 22-05 D1 | 22-05 | SHELL-01 | `tsc`/`build` + Copywriting Contract grep; live browser `page.innerText` confirms verbatim (22-VERIFICATION.md, 22-UI-REVIEW.md) | ✅ green |
| 22-05 D2 | 22-05 | THEME-02 | Live browser confirms `--onac` legibility on the CTA (22-VERIFICATION.md) | ✅ green |
| 22-05 D3 | 22-05 | SHELL-01 | Grep confirms literal placeholder text, `textTransform`, no `onClick` | ✅ green |
| 22-05 D4 | 22-05 | THEME-04 | Grep confirms no density literals, no `matchMedia`/`useMediaQuery`/`innerWidth` in DetailsPanel | ✅ green |
| 22-06 D1 | 22-06 | SHELL-03 | `tsc`/`build` + grep confirms 16 tab-era fields removed, `storcat-last-*` untouched | ✅ green |
| 22-06 D2 | 22-06 | THEME-04 | Grep confirms reducer actions/state added, `readPersistedPrefs` seeding | ✅ green |
| 22-06 D3 | 22-06 | SHELL-03 | Grep confirms `useMediaQuery` hook shape and single call site | ✅ green |
| 22-06 D4 | 22-06 | SHELL-04 | Grep confirms min-width-only ladder, zero `max-width`/`@container`; live browser confirms tier measurements (22-VERIFICATION.md: `236px 804px` @1040px, `200px 839px` @1039px) | ✅ green |
| 22-06 D5 | 22-06 | SHELL-05 | Grep confirms rail-side CSS scoped correctly; live browser confirms `Ctrl+Alt+R` flip, grid/order/divider swap, persistence across reload (22-VERIFICATION.md) | ✅ green |
| 22-06 D6 | 22-06 | THEME-04 | Grep confirms reducer wiring; live browser confirms `Ctrl+Alt+D` live toggle (22-VERIFICATION.md) | ✅ green |
| 22-06 D7 | 22-06 | THEME-06 | Grep confirms persistence wiring; live browser confirms reload-level persistence for all 3 keys (22-VERIFICATION.md). Full quit/relaunch of the packaged binary remains untestable until Phase 26 — see Manual-Only THEME-06 row | ✅ green (reload-level) / ⏭ deferred (full-binary, Phase 26) |
| 22-07 D1 | 22-07 | SHELL-03 | Live dev-browser: drawer geometry/shadow/z-index/close-paths/no-orphan-on-widen all measured directly (22-VERIFICATION.md) | ✅ green |
| 22-07 D2 | 22-07 | SHELL-09 | Live dev-browser: idempotent toggle, concurrent close-path race confirmed non-throwing (22-VERIFICATION.md). Full cross-overlay stacking deferred to Phase 24 — see Manual-Only SHELL-09 row | ✅ green (this phase's scope) / ⏭ deferred (cross-overlay, Phase 24) |
| 22-07 D3 | 22-07 | SHELL-03 | Live dev-browser: Details chip visibility measured at exact tier boundaries (1279/1280px) | ✅ green |
| 22-07 D4 | 22-07 | THEME-02 | Live dev-browser: found the New-pill hover bug (dead CSS), fixed it, re-measured across all 11 themes | ✅ green |
| 22-07 D5 | 22-07 | SHELL-09 (partial) | Explicitly out of this phase's scope by design — see Manual-Only SHELL-09 row | ⏭ deferred (Phase 24) |
| 22-07 D6 | 22-07 | SHELL-07 | Click registration confirmed live; native drag-vs-click arbitration itself accepted platform-gated by the user — see Manual-Only SHELL-07 row | ⚠️ accepted (platform-gated) |
| 22-07 D7 | 22-07 | SHELL-08 | Superseded by 22-VERIFICATION.md's post-report real macOS build check | ✅ green (superseded) |
| 22-07 D8 | 22-07 | THEME-05 | Live dev-browser: 4 font requests, all same-origin, 0 external; `document.fonts` reports all 4 in-use weights loaded. Physical network-disconnect variant not reproducible against a Vite dev server (a built binary has no HTTP dependency at all, so this gap is dev-environment-specific, not a real product gap) | ✅ green (network-trace half) / ⚠️ environment-limited (physical-disconnect half) |
| 22-07 D9 | 22-07 | THEME-06 | Superseded by 22-06 D7 / 22-VERIFICATION.md's reload-level persistence proof; full-binary case deferred to Phase 26 | ✅ green (reload-level) / ⏭ deferred (full-binary, Phase 26) |

*Status: ✅ green (verified) · ⚠️ partial/accepted/environment-limited (see note) · ⏭ deferred (explicit, scheduled) · ❌ red (unresolved gap)*

**No ❌ red rows.** Every task's requirement is either fully green, explicitly and narrowly deferred to a named future phase where the missing precondition first exists, or accepted as platform-gated by the user with the structural half (code-level wiring) independently confirmed.

---

## Wave 0 Requirements

No test-framework installation. Wave 0 for this phase was the **dev-only state-exercise affordance** the research flagged, because several requirements would otherwise be unverifiable by a human in this phase's shipped UI:

- [x] A throwaway `import.meta.env.DEV`-gated affordance (`DevStateSwitcher.tsx`, `Ctrl+Alt+T/D/R`) to cycle all 11 themes, toggle density Compact/Comfortable, and flip rail side Left/Right — built in 22-01 (theme/density) and extended in 22-06 (rail side). This is what made THEME-01, THEME-02, THEME-04, and SHELL-05 exercisable at all, and 22-VERIFICATION.md's live dev-browser pass used it directly (12x `Ctrl+Alt+T` cycle, `Ctrl+Alt+D` density flip, `Ctrl+Alt+R` rail-side flip, all measured).
- [x] The affordance is `DEV`-gated so it cannot ship in a production build. Confirmed twice independently: 22-01's own build (`grep -l storcat-dev-switcher dist/assets/*.js` → 0 matches) and 22-SECURITY.md's independent fresh rebuild against current HEAD (`rm -rf dist && npm run build`, same grep, same 0 matches — not merely trusted from source).

**Wave 0: complete.**

*Everything else: existing infrastructure (tsc / vite build / go build / go test) covers what can be covered.*

---

## Manual-Only Verifications

Reconciled against `22-VERIFICATION.md` (live dev-browser session against the actual Vite dev server, measuring computed values at every tier boundary and across all 11 themes) and `22-UI-REVIEW.md` (independent 6-pillar visual audit with its own live measurements). Ten of these thirteen rows were in fact executed live, not left as unexecuted instructions. Three genuinely were not, for the reasons stated.

| Behavior | Requirement | Actually Executed? | Evidence |
|----------|-------------|---------------------|----------|
| Single workspace view, no tabs; 46px toolbar / rail / tree / details / 26px status bar | SHELL-01 | ✅ Yes — live | `.ant-tabs` absent, `.ws-toolbar` height 46px, `.ws-status` height 26px, all measured in-browser (22-VERIFICATION.md); screenshot `phase22-final-wide.png` matches UI-SPEC |
| `268px 1fr 288px` grid at ≥1280px | SHELL-02 | ✅ Yes — live | Computed `grid-template-columns: 268px 844px 288px` at 1400px width (22-VERIFICATION.md) |
| Tier swap at 1279px and 1040px boundaries | SHELL-03, SHELL-04 | ✅ Yes — live | `236px 1043px` @1279px, `236px 804px` @1040px, `200px 839px` @1039px — exact boundary crossings, no gap/overlap (22-VERIFICATION.md) |
| Rail-side swap moves the 1px divider with it | SHELL-05 | ✅ Yes — live | `Ctrl+Alt+R` flips grid to `288px 1fr 268px`, rail `order: 3`, divider moves to rail's left edge, persists across reload (22-VERIFICATION.md) |
| Toolbar drag works; search field, theme chip, gear, Details chip all remain clickable | SHELL-07 | ⚠️ Partial — structural only, native arbitration **not** executed | All 4 controls confirmed to carry `.no-drag` (grep + live `getComputedStyle` on the search button). The click-vs-drag arbitration itself is native Wails webview runtime behavior with no plain-Chromium equivalent, and no Windows build exists in this environment. **Accepted by the user, 2026-08-13, as platform-gated** (22-VERIFICATION.md `human_verification_accepted`) |
| macOS traffic lights inside the 46px toolbar; native title bar above it on Windows/Linux | SHELL-08 | ✅ Yes — real macOS build | `wails build -platform darwin/arm64` (exit 0), app launched, window frame captured via System Events; traffic lights sit on the same horizontal band as the toolbar content, no native title bar above, TitleBarHiddenInset behavior confirmed (22-VERIFICATION.md `verified_post_report`) |
| Details drawer stacks correctly at narrow tiers | SHELL-09 | ✅ Yes — live (this phase's scope only) | Drawer geometry/shadow/z-index, both close paths (Escape + backdrop click), concurrency-safety, no-orphan-on-widen all measured live (22-VERIFICATION.md). Full cross-overlay stacking (drawer vs. a second overlay) genuinely **not executed** — no second overlay exists yet. **Deferred to Phase 24**, which introduces the ⌘K palette (the first second overlay to stack against); 22-07-PLAN.md's own Task 3 instructions state this explicitly |
| All 11 themes repaint the entire workspace immediately | THEME-01 | ✅ Yes — live | 12x `Ctrl+Alt+T` cycle, confirmed 11-entry wraparound (22-VERIFICATION.md) |
| Legible text on every accent-filled button/badge in all 11 themes | THEME-02 | ✅ Yes — live | `--onac` measured directly per theme (Gruvbox `#fe8019`→white, Monokai `#a6e22e`→dark, GitHub `#58a6ff`→white); New-pill hover-inversion bug found and fixed, then re-measured on Gruvbox Dark (22-VERIFICATION.md, 22-UI-REVIEW.md Pillar 3) |
| 14 tokens present on `:root` per theme | THEME-03 | ✅ Yes — live | All 14 custom properties confirmed resolving to concrete values for every theme (22-VERIFICATION.md) |
| Density changes row height, rail/details/palette padding and tree font size | THEME-04 | ✅ Yes — live | `Ctrl+Alt+D` flips `--rh` 34px↔27px live (22-VERIFICATION.md) |
| IBM Plex Sans and Mono render with no network access | THEME-05 | ⚠️ Partial — network-trace half executed live; physical-disconnect half **not** executed | Live network trace: 4 font requests, all same-origin, 0 external (22-VERIFICATION.md, superseding 22-02's inconclusive attempt). The literal "disconnect and reload" check could not be reproduced against a Vite dev server — Playwright's offline mode also severs the connection to the dev server itself, unlike a built binary which embeds its assets. This is a dev-environment limitation, not an unverified product behavior: the build asset trace independently proves zero external hosts are referenced anywhere in the shipped CSS |
| Theme, density and rail position survive an app restart | THEME-06 | ⚠️ Partial — reload-level executed live; full quit/relaunch of the packaged binary **not** executed | Reload-level persistence (the strongest a dev-server session can prove) confirmed for all 3 keys, exercising the identical `readPersistedPrefs()` → `applyTokens()` path a relaunch uses (22-VERIFICATION.md). Full quit/relaunch genuinely can't be tested yet: `DevStateSwitcher` is `DEV`-gated, so a **production** build has no way to set a non-default value in the first place — a fresh launch and a restored launch are indistinguishable until Settings exists. **Deferred to Phase 26**, when Settings (SET-01..04) first makes a non-default value settable in a production build |

**Reconciliation summary: 10/13 rows executed live and verified. 3/13 have a documented, narrow, non-arbitrary reason they could not be, each resolved to an explicit disposition (accepted-by-user, or deferred-to-named-future-phase) rather than left open.**

---

## Validation Sign-Off

- [x] All tasks have an automated verify command or an explicit Manual-Only row (see Per-Task Verification Map above)
- [x] `tsc --noEmit` green after the tab-UI removal (proves no dead imports/state) — reconfirmed 2026-08-13
- [x] `npm run build` and `go build ./...` green — reconfirmed 2026-08-13
- [x] `go test ./...` green (unchanged — this phase touches one Go line) — reconfirmed 2026-08-13, all packages `ok`
- [x] Wave 0 dev affordance exists and is `import.meta.env.DEV`-gated — confirmed twice (22-01 build check, 22-SECURITY.md independent fresh rebuild)
- [x] All 13 manual rows above completed — 10 executed live, 3 resolved to explicit accepted/deferred dispositions (SHELL-07 Windows, SHELL-09 cross-overlay, THEME-06 full-binary), none left silently open
- [x] SHELL-08 verified on a real macOS build; SHELL-07 flagged and accepted by the user as platform-gated (no Windows build available)
- [x] No watch-mode flags in any verify command
- [x] `nyquist_compliant: true` set in frontmatter

**`nyquist_compliant: true` — rationale:** This project has no frontend test framework by locked, explicit decision (TEST-01 deferred to a future milestone; documented in `PROJECT.md`, `22-CONTEXT.md`, and this file's own opening constraint). Nyquist compliance for this phase is therefore not "does an automated test suite cover every requirement" — it is "does every requirement resolve to real evidence via the means this project actually has: `tsc`/`vite build`/`go build`/`go test`, plus live dev-browser measurement, plus real per-OS builds where required." All 14 requirement IDs resolve to evidence by that standard: 11 are fully green, and the remaining 3 sub-claims (Windows drag arbitration, full cross-overlay stacking, full-binary restart) are not silent gaps — they are explicitly accepted-by-the-user or deferred to the specific future phase where the missing precondition (a Windows machine, a second overlay, a Settings UI) will first exist. Setting this field to `true` reflects that the validation contract this phase actually promised (see the file's own "Honest constraint up front") has been discharged, not that a test suite exists.

**Approval:** reconciled 2026-08-13 (retroactive audit, gsd-validate-phase)
