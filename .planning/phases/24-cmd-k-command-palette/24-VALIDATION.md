---
phase: 24
slug: cmd-k-command-palette
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-14
---

# Phase 24 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by plan-phase from `24-RESEARCH.md`'s `## Validation Architecture`. The Per-Task
> Verification Map is filled in once PLAN.md task IDs exist.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go: `go test` (table-driven, `*_test.go` beside source). Frontend: **none by design** — TEST-01 (Vitest + Testing Library) is an explicitly deferred milestone item; do not add a frontend test framework this phase. Frontend proof is `tsc --noEmit` + `vite build` + live `dev-browser` verification, matching Phase 22/23 precedent. |
| **Config file** | `go.mod`, `frontend/tsconfig.json`, `frontend/vite.config.ts` (all unchanged this phase) |
| **Quick run command** | `go test ./internal/search/... && (cd frontend && npx tsc --noEmit)` |
| **Full suite command** | `cd frontend && npx tsc --noEmit && npm run build && cd .. && go build ./... && go test ./... -race -count=1` |
| **Estimated runtime** | ~60–90 seconds (full suite, dominated by `vite build` and `-race`) |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/search/...` for Go tasks; `npx tsc --noEmit` for every frontend task
- **After every plan wave:** Full suite command above
- **Before `/gsd-verify-work`:** Full suite green **plus** a live `dev-browser` pass against `wails dev` covering PLT-01/04/05/06/07 and the multi-branch-open reveal regression
- **Max feedback latency:** ~90 seconds

**Dev-server note (carried from `23-VALIDATION.md:114`):** browser verification must run against `wails dev` on **`:34115`**, not Vite's `:5173` — `:5173` exposes no `window.go` bindings, so every binding-dependent behavior silently no-ops there.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| *pending — filled after PLAN.md task IDs are assigned* | | | | | | | | | ⬜ pending |

---

## Wave 0 Requirements

- [ ] A Go test file for the new capped-search surface (e.g. `internal/search/*_test.go`) — covers PLT-02/PLT-03 cap-at-50 and total-count behavior, **including the boundary cases: 0 matches, exactly 50, and 51 matches**
- [ ] Extend (do **not** replace) `cli/search_test.go` with a regression assertion that `cli/search.go`'s output for a fixed fixture is unchanged by this phase's Go edits

*No new frontend test infrastructure. Adding Vitest/Testing Library here would be scope creep against a locked deferral (TEST-01).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Palette opens via ⌘K and via the toolbar `.ws-search` click, input autofocused | PLT-01 | No frontend test framework (TEST-01 deferred); ⌘K delivery through WKWebView cannot be proven outside a real webview — this is RESEARCH Open Question #1 | In `wails dev` (`:34115`): press ⌘K from a neutral focus, then again with the rail filter focused; click the toolbar search field. Each opens the palette with the caret in the input. A second ⌘K while open is a no-op, not a re-open. |
| Keyboard navigation and Escape dismiss | PLT-04 | Same | ↓/↑ move the active row and clamp at both ends (no wrap); Home/End jump to first/last; Enter activates the active row; Escape closes from any focus position inside the palette. |
| Click/Enter a hit reveals it in the tree | PLT-05 | Same | Pick a hit in a *different* catalog: the rail switches catalog, every ancestor expands, the node is selected, it is scrolled to vertical centre, and the palette closes. |
| **Reveal does not collapse unrelated open branches** | PLT-05 | Same — this is RESEARCH Pitfall 2, caused by `SET_EXPANDED` replacing rather than merging the expanded map (`AppContext.tsx:123-126`) | Pre-expand 2–3 unrelated branches. Reveal a node in a *different* branch. Assert the pre-expanded branches are **still expanded** afterwards. This is the single highest-value manual check in the phase. |
| Empty and truncation copy render verbatim | PLT-03, PLT-06 | Same | Search a term matching nothing → exactly `No file in any catalog matches that.` Search a term matching >50 → exactly `Showing the first 50 of N hits` with the real N. A 1-character query shows `Type to search…`, never the no-match string. |
| Focus trap and scroll lock | PLT-07 | Same | With the palette open, Tab/Shift+Tab cannot land on the rail, toolbar, or tree behind the scrim; the page behind does not scroll; on close, focus returns to the element that was focused before opening. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
