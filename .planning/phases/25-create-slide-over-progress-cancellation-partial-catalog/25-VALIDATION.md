---
phase: 25
slug: create-slide-over-progress-cancellation-partial-catalog
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-14
---

# Phase 25 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by plan-phase from `25-RESEARCH.md`'s `## Validation Architecture`. The Per-Task
> Verification Map is filled in once PLAN.md task IDs exist.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing`, table-driven, `*_test.go` beside source. Frontend: **none by design** — TEST-01 (Vitest + Testing Library) is an explicitly deferred milestone item; do not add one. |
| **Config file** | none — plain `go test ./...`; `frontend/tsconfig.json`, `frontend/vite.config.ts` unchanged |
| **Quick run command** | `go test ./internal/catalog/... ./internal/search/... ./cli/...` |
| **Full suite command** | `go build ./... && go test ./... -race -count=1 && cd frontend && npx tsc --noEmit && npm run build` |
| **Estimated runtime** | ~60–90 seconds |

---

## Sampling Rate

- **After every task commit:** `go test ./internal/catalog/... ./internal/search/...` for Go tasks; `npx tsc --noEmit` for frontend tasks
- **After every plan wave:** full suite command above
- **Before `/gsd-verify-work`:** full suite green **plus** the two manual checks that cannot be automated without real removable media (volume-disappearance mid-scan; force-quit mid-scan)
- **Max feedback latency:** ~90 seconds

**Dev-server note:** browser verification runs against `wails dev` on **`:34115`**. Vite's `:5173` exposes no `window.go`, so every binding-dependent assertion passes vacuously there.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| *pending — filled after PLAN.md task IDs are assigned* | | | | | | | | | ⬜ pending |

---

## Requirement → Test Map (from RESEARCH)

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CRT-09 | Cancel actually stops the walk and writes nothing | unit | `go test ./internal/catalog/... -run TestCreateCatalogWithContext_Cancel -v` | ❌ Wave 0 |
| CRT-10 | Terminal (volume vanished) vs single-entry error classification | unit | `go test ./internal/catalog/... -run TestTraverseDirectory_TerminalError -v` | ❌ Wave 0 |
| CRT-11 | Partial catalog write produces the correct marker shape | unit | `go test ./internal/catalog/... -run TestWritePartialCatalog_Marker -v` | ❌ Wave 0 |
| COMPAT-02 | Clean-scan JSON byte-identical to the pre-milestone shape | unit | `go test ./internal/catalog/... -run TestCreateCatalog_JSONShapeUnchanged -v` | ❌ Wave 0 |
| COMPAT-03 | CLI `create` behaves identically, **including the `WriteHTML` default** | unit + smoke | `go test ./cli/... -v`, then `go run . create <tmpdir> --json` | ✅ extend `cli/*_test.go` |
| COMPAT-04 | `internal/catalog` imports no Wails package | static check | `go list -deps ./internal/catalog/... \| grep wailsapp` → expect empty | ❌ Wave 0 (assertion, not a Go test) |

---

## Wave 0 Requirements

- [ ] `internal/catalog/service_test.go` extensions — cancellation, terminal-error classification, partial-marker `omitempty` shape, and a JSON byte-diff against a pre-milestone fixture
- [ ] `internal/volumes/volumes_test.go` — new package; darwin-buildable assertions at minimum, windows/linux behind build tags and structurally testable only where the toolchain can build them
- [ ] A `COMPAT-04` guard asserting `internal/catalog` pulls in no `wailsapp` dependency

*No new frontend test infrastructure. Adding Vitest/Testing Library would be scope creep against a locked deferral (TEST-01).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| **Volume disappears mid-scan** | CRT-10, CRT-11 | Requires physically removing real removable media mid-walk; no way to simulate a vanished mount point faithfully in a unit test | Start a scan of a USB/external volume, physically eject it mid-scan. Assert: the error state appears naming where it stopped and the read errors seen; "Write partial catalog" produces a catalog carrying the unreadable-subtree marker; "Retry scan" and "Close without writing" both behave as specified. |
| **Force-quit mid-scan writes nothing** | CRT-13 | Requires killing the real process mid-write; the guarantee is about on-disk state after an abrupt exit | Start a large scan, quit the app (⌘Q / window close) mid-walk. Assert: no `.json` or `.html` appears in the output directory, and no `.tmp` residue is left behind. |
| **Atomic write survives a crash** | CRT-11 (and Phase 27's ACT-09, which reuses this primitive) | Same — needs a real kill signal during the write window | Kill the process during the write of a large catalog. Assert no truncated JSON exists at the destination path. |
| **`/Volumes` filtering** | CRT-02 | Depends on the live machine's actual mount table, including the boot-volume symlink and Apple-internal snapshot dirs RESEARCH found on this machine | `wails dev`, open the Create slide-over, compare the rendered volume cards against live `ls -la /Volumes`. Boot symlink and Apple-internal mounts must be absent; the two `d--x--x--x` mounts must render with the `read errors` tag. |
| Slide-over animation, progress rendering, states | CRT-01, CRT-04–08, CRT-12 | No frontend test framework (TEST-01 deferred) | Live `dev-browser` session against `:34115`, per Phases 22–24 precedent. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
