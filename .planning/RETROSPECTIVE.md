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

## Milestone: v2.1.0 — CLI Commands

**Shipped:** 2026-03-26
**Phases:** 4 | **Plans:** 7

### What Was Built
- Full CLI subcommand framework in `cli/` package with stdlib flag.FlagSet dispatch
- 6 subcommands: create, search, list, show, open, version — all with --help and proper exit codes
- Table output (tablewriter) and --json flag for machine-readable output on data commands
- Colorized tree rendering with --depth N, --no-color, and NO_COLOR env var support
- Cross-platform browser launch for `storcat open`
- Platform compatibility: Windows console attachment, macOS -psn_* filtering, install script

### What Worked
- Phase 8 stub pattern: create dispatch skeleton with stubs, then replace each stub in subsequent phases — clean incremental approach
- Milestone audit between Phase 10 and Phase 11 caught 6 tech debt items before shipping
- Phase 11 as explicit tech debt cleanup phase resolved 5/6 items with surgical commits
- All 23 requirements verified via 3-source cross-reference (VERIFICATION.md, REQUIREMENTS.md traceability, code inspection)
- Same-day execution for all 4 phases — minimal overhead, focused scope

### What Was Inefficient
- SUMMARY.md one_liner field still not populated in Phase 10/11 summaries — same gap as v2.0.0
- Phase 11 Nyquist VALIDATION.md missing — same gap pattern as v2.0.0 Phase 7 (cleanup/verification phases get skipped)
- Some decisions in STATE.md accumulated as "pending todos" that were already resolved — stale context

### Patterns Established
- Interspersed flag parsing: pre-separate positional args from flags before flag.Parse
- `cli/output.go` shared helpers (printJSON, formatBytes) across all data commands
- TDD red-green for each command: failing tests first, then implementation
- fs.Usage override pattern for custom --help output on flag.FlagSet

### Key Lessons
1. Stub pattern for CLI commands is excellent — provides compile-time safety while enabling incremental delivery
2. Audit-then-cleanup as a two-step flow works well — audit identifies, dedicated phase resolves
3. SUMMARY.md template compliance remains the biggest recurring gap — needs enforcement in execution workflow

### Cost Observations
- Model mix: Primarily Opus for planning/execution, Sonnet for subagents
- Sessions: ~3 sessions in 1 day
- Notable: 4 phases in a single day — CLI commands were well-scoped with clear boundaries

---

## Cross-Milestone Trends

### Process Evolution

| Metric | v2.0.0 | v2.1.0 |
|--------|--------|--------|
| Phases | 7 | 4 |
| Plans | 11 | 7 |
| Tasks | 15 | 9 |
| Timeline | 3 days | 1 day |
| Requirements | 20/20 | 23/23 |
| Tech Debt Items | 11 | 6 (5 resolved) |

### Recurring Themes
- Bottom-up dependency ordering is effective for migration work
- Nyquist validation needs enforcement for final phases, not just middle ones
- SUMMARY.md one_liner/requirements_completed fields consistently empty — template gap persists across milestones
- Audit-then-cleanup two-step is a reliable pattern for shipping quality
