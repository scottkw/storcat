# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v2.0.0 — Go/Wails Migration

**Shipped:** 2026-03-26
**Phases:** 7 | **Plans:** 11 | **Tasks:** 15

### What Was Built
- Complete Go/Wails backend replacing Electron/Node.js — 6 Go source files, Wails bindings, config manager
- Full IPC surface with 17 wailsAPI wrappers maintaining `window.electronAPI` contract
- Window state persistence (size, position, toggle) via Go config lifecycle hooks
- macOS platform integration (drag region, build-time version injection)
- Cross-platform build verification and clean merge to main

### What Worked
- Bottom-up phase ordering (data models → services → config → app layer → shim → platform → verify) prevented rework
- TDD approach in Phase 2 (LoadCatalog) and Phase 3 (Config) caught issues early
- Strict `{success,...}` envelope pattern made IPC debugging straightforward
- 3-day execution timeline for 7 phases — fast cadence maintained throughout
- Milestone audit before completion caught 11 tech debt items cleanly

### What Was Inefficient
- SUMMARY.md `requirements_completed` frontmatter was empty across all 11 summaries — template gap never caught during execution
- Phase 7 Nyquist VALIDATION.md left in draft — validation workflow wasn't enforced for the final verification phase
- Some one-liner fields in SUMMARY.md were empty, degrading automated accomplishment extraction

### Patterns Established
- `{success,...}` envelope pattern for all IPC/API boundaries
- Dual-format parsing (v1 array + v2 bare object) for backward compatibility
- OnDomReady (not OnStartup) for window state restoration in Wails
- `//go:embed wails.json` as simpler alternative to ldflags for version injection
- `--wails-draggable: drag` CSS property for macOS titlebar

### Key Lessons
1. Plan every phase with observable success criteria — the "what must be TRUE" format caught ambiguity early
2. Template gaps (empty frontmatter fields) compound silently — validate template compliance at phase completion
3. Advisory tech debt items are fine to ship with — the audit surfaced them, and none blocked the release

### Cost Observations
- Model mix: Primarily Opus for planning/execution, Sonnet for subagents
- Sessions: ~5 sessions across 3 days
- Notable: Phase 5 (Frontend Shim) was single-plan, single-file — minimal overhead

---

## Cross-Milestone Trends

### Process Evolution

| Metric | v2.0.0 |
|--------|--------|
| Phases | 7 |
| Plans | 11 |
| Tasks | 15 |
| Timeline | 3 days |
| Requirements | 20/20 |
| Tech Debt Items | 11 |

### Recurring Themes
- Bottom-up dependency ordering is effective for migration work
- Nyquist validation needs enforcement for final phases, not just middle ones
