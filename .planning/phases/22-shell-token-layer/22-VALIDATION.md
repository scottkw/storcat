---
phase: 22
slug: shell-token-layer
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-13
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

Filled in by the planner against the actual task IDs. Every task must carry either an automated command from the table above or an explicit entry in the Manual-Only table below — no task may claim verification it cannot produce.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| *(planner fills)* | | | | — | N/A | | | | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

No test-framework installation. Wave 0 for this phase is the **dev-only state-exercise affordance** the research flagged, because several requirements are otherwise unverifiable by a human in this phase's shipped UI:

- [ ] A throwaway `import.meta.env.DEV`-gated affordance (keyboard shortcut or tiny dev panel) to cycle all 11 themes, toggle density Compact/Comfortable, and flip rail side Left/Right — none of these have real UI until Phase 26. Without it, THEME-01, THEME-02, THEME-04 and SHELL-05 cannot be exercised at all.
- [ ] The affordance must be `DEV`-gated so it cannot ship in a production build, and is removed or superseded when Phase 26's Settings lands.

*Everything else: existing infrastructure (tsc / vite build / go build / go test) covers what can be covered.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Single workspace view, no tabs; 46px toolbar / rail / tree / details / 26px status bar | SHELL-01 | Visual composition | Run `wails dev`; confirm no tab bar exists and all five regions render |
| `268px 1fr 288px` grid at ≥1280px | SHELL-02 | Computed-style check | Resize to ≥1280px; DevTools → inspect the middle row's computed `grid-template-columns` |
| Tier swap at 1279px and 1040px boundaries | SHELL-03, SHELL-04 | Resize behavior | Drag the window slowly through both boundaries; confirm rail 268→236→200px and details pane→drawer at 1279px |
| Rail-side swap moves the 1px divider with it | SHELL-05 | No Settings UI this phase | Flip rail side via the Wave 0 dev affordance; confirm divider follows and pane order swaps |
| Toolbar drag works; search field, theme chip, gear, Details chip all remain clickable | SHELL-07 | Native window behavior | Drag from empty toolbar space (window moves); click each control (each responds). **Must also be checked on a real Windows build** — click-swallowing is most visible there (PITFALLS.md #12) |
| macOS traffic lights inside the 46px toolbar; native title bar above it on Windows/Linux | SHELL-08 | Requires real per-OS builds | Build and launch on macOS: traffic lights sit in the toolbar band with correct left inset. On Windows/Linux: native title bar sits above the toolbar. Not verifiable from source |
| Details drawer stacks correctly at narrow tiers | SHELL-09 | Only partial this phase | Open the details drawer below 1280px; confirm it overlays the tree and the backdrop closes it. Full stacking order can only be verified from Phase 24 on, when a palette/dialog exists to stack against |
| All 11 themes repaint the entire workspace immediately | THEME-01 | No theme picker this phase | Cycle all 11 via the Wave 0 dev affordance; confirm every surface repaints with no stale colors |
| Legible text on every accent-filled button/badge in all 11 themes | THEME-02 | Human contrast judgement | Cycle all 11 and inspect the accent CTA and "+ New" pill in each. **Must check all 11** — the `--onac` bug only appears on light accents (Gruvbox orange, Monokai green) (PITFALLS.md #23) |
| 14 tokens present on `:root` per theme | THEME-03 | DevTools inspection | DevTools → Computed styles on `:root`; confirm all 14 custom properties resolve to concrete values (no unresolved `color-mix()`, no empty strings) |
| Density changes row height, rail/details/palette padding and tree font size | THEME-04 | No density toggle this phase | Toggle via the Wave 0 dev affordance; confirm `--rh` 27↔34px and the padding/font-size tokens change |
| IBM Plex Sans and Mono render with no network access | THEME-05 | Requires offline check | DevTools → Network → Offline (or physically disconnect), reload; confirm fonts still render as Plex, not a system fallback. Also confirm zero external font requests in the Network tab on a normal load |
| Theme, density and rail position survive an app restart | THEME-06 | Requires the real binary | Set non-default values, fully quit the built app (not a `vite dev` page reload), relaunch; confirm all three persist **and that the theme is applied before first paint — no flash of the default theme** |

---

## Validation Sign-Off

- [ ] All tasks have an automated verify command or an explicit Manual-Only row
- [ ] `tsc --noEmit` green after the tab-UI removal (proves no dead imports/state)
- [ ] `npm run build` and `go build ./...` green
- [ ] `go test ./...` green (unchanged — this phase touches one Go line)
- [ ] Wave 0 dev affordance exists and is `import.meta.env.DEV`-gated
- [ ] All 13 manual rows above completed, including the all-11-themes contrast pass
- [ ] SHELL-08 verified on a real macOS build; SHELL-07 flagged if no Windows build was available
- [ ] No watch-mode flags in any verify command
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
