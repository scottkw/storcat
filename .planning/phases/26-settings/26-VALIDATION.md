---
phase: 26
slug: settings
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-15
---

# Phase 26 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by plan-phase from `26-RESEARCH.md`'s `## Validation Architecture`. The Per-Task
> Verification Map is filled in once PLAN.md task IDs exist.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing`, table-driven, `*_test.go` beside source (existing: `internal/config/config_test.go`, `internal/osutil/reveal_test.go`). Frontend: **none by design** — TEST-01 (Vitest + Testing Library) is an explicitly deferred milestone item; do not add one. |
| **Config file** | none — plain `go test ./...`; `frontend/tsconfig.json`, `frontend/vite.config.ts` unchanged |
| **Quick run command** | `go test ./internal/config/... ./internal/osutil/... -race -count=1` |
| **Full suite command** | `go build ./... && go test ./... -race -count=1 && (cd frontend && npx tsc --noEmit && npm run build)` |
| **Estimated runtime** | ~60–90 seconds |

---

## Sampling Rate

- **After every task commit:** `go build ./... && go test ./internal/config/... ./internal/osutil/... -race -count=1` for Go tasks; `npx tsc --noEmit` for frontend tasks
- **After every plan wave:** full suite command above, plus a live dev-browser pass against `:34115`
- **Before `/gsd-verify-work`:** full suite green plus the manual-only entry-point and visual checks below
- **Max feedback latency:** ~90 seconds

**Dev-server note:** browser verification runs against `wails dev` on **`:34115`**. Vite's `:5173` exposes no `window.go`, so every binding-dependent assertion passes vacuously there. Per STATE.md's standing note, `curl` liveness proves the server is up, not that bindings are fresh — verify `Object.keys(window.go.main.App)` includes the new/changed methods before recording any evidence.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| *(filled in once PLAN.md task IDs exist)* | — | — | — | — | — | — | — | — | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

### Requirement → verification approach (from RESEARCH.md)

| Req ID | Behavior | Test Type | Automated Command |
|--------|----------|-----------|-------------------|
| SET-01 | ⌘,/gear/theme chip open Settings | manual-only (live dev-browser) | dev-browser click + keyboard event against `:34115` |
| SET-02 | 11 theme cards apply immediately | manual-only + Go persistence test | `go test ./internal/config/... -run TestSetTheme` |
| SET-03 | Density/rail-position segmented controls | manual-only + Go persistence test | `go test ./internal/config/... -run 'TestSetDensity\|TestSetRailSide'` |
| SET-04 | Catalog directory, filename root, 4 toggles | manual-only + Go persistence tests | `go test ./internal/config/... -run 'TestSetCatalog\|TestSetWriteHTML\|TestSetCopyToSecondary\|TestSetWatchDirectory'` |
| SET-05 | Save-as-you-change, no debounce | unit (Go): each `Set*` calls `Save()` synchronously | `go test ./internal/config/...` |
| COMPAT-05 | Window persistence toggle continuity | unit (Go, existing) + manual restart check | `go test ./internal/config/... -run TestWindowPersistence` |
| T-22-05 | `CatalogModal.tsx` deleted, no dead listener | static assertion | `grep -rn openCatalogModal frontend/src` returns nothing |
| FU-23-A | `GetCatalogHtmlPath`/`OpenExternal` containment | unit (Go), table-driven, mirroring `reveal_test.go` | `go test ./... -run 'TestGetCatalogHtmlPath\|TestOpenExternal'` |

---

## Wave 0 Requirements

- [ ] New table-driven cases in `internal/config/config_test.go` for the new `Set*`/`Get*` pairs (`density`, `railSide`, `catalogDirectory`, `defaultFilenameRoot`, `writeHTML`, `copyToSecondary`, `secondaryDirectory`, `watchDirectory`), following `newTestManager(t)`'s existing temp-dir fixture pattern (`config_test.go:11-30`)
- [ ] Containment tests for `GetCatalogHtmlPath` / `OpenExternal` (new file or extending `reveal_test.go`), following the existing `ContainsPath` table structure — must cover `../` escape, symlink escape, and same-prefix sibling directory
- [ ] Migration test: localStorage-sourced values land in Go config on first run and are not re-migrated on subsequent runs
- [ ] No new frontend test file — none needed, consistent with TEST-01's deferral and every prior phase's precedent

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| ⌘, / gear / theme chip all open Settings | SET-01 | No frontend test framework (TEST-01 deferred) | `wails dev` on `:34115`; dev-browser: press ⌘,, click gear, click theme chip — dialog opens from each |
| Theme applies immediately on card click across all 11 themes | SET-02 | Visual token application | Click each of the 11 cards; confirm tokens repaint with no reload and no flash |
| Density and rail-position segmented controls take effect live | SET-03 | Visual/layout | Toggle each segment; confirm row height and rail side change immediately |
| Settings survive an app restart (all fields) | SET-05 | Requires real process restart | Change every control, quit, relaunch, confirm every value persisted |
| Window size/position persistence continues to work exactly as pre-milestone, gated by the toggle | COMPAT-05 | Requires real window manager + restart | Toggle on: resize/move, quit, relaunch → restored. Toggle off: resize/move, quit, relaunch → not restored |
| Flash-free boot after the localStorage→config migration | (architecture risk, RESEARCH Pitfall 1) | Only observable as a first-paint artifact | Cold-launch on a non-default theme; confirm no default-theme flash before first paint |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
