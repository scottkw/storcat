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
| 25-01-T1 | 01 | 1 | CRT-06, COMPAT-02, COMPAT-03, COMPAT-04 | T-25-06 | Hostile filenames stay HTML-escaped through the unchanged writer | unit (tdd) | `go test ./internal/catalog/... ./cli/... -count=1` | ✅ extend `internal/catalog/service_test.go` | ⬜ pending |
| 25-01-T2 | 01 | 1 | CRT-07, COMPAT-04 | T-25-01, T-25-02, T-25-03 | Write-path containment gate; one-scan-at-a-time refusal | unit (tdd) | `go test . ./internal/osutil/... -count=1 -race` | ✅ extend `app_test.go` | ⬜ pending |
| 25-01-T3 | 01 | 1 | CRT-01, CRT-03, CRT-12 | — | — | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-02-CP | 02 | 2 | CRT-11 | T-25-11 | On-disk marker shape approved before it becomes a one-way door | decision checkpoint | n/a — blocking human decision | n/a | ⬜ pending |
| 25-02-T1 | 02 | 2 | CRT-11 | T-25-04, T-25-05 | Random-suffix exclusive temp file in the destination dir; removed on every error path | unit (tdd) | `go test ./internal/catalog/... ./cli/... -count=1` | ❌ create `internal/catalog/atomicwrite_test.go` | ⬜ pending |
| 25-02-T2 | 02 | 2 | CRT-09, CRT-10, COMPAT-02, COMPAT-03 | T-25-10, T-25-11 | Read-error record cap; omitempty marker leaves clean catalogs byte-identical | unit (tdd) | `go test ./internal/catalog/... ./internal/search/... ./cli/... -count=1 -race` | ✅ extend `internal/catalog/service_test.go`, `internal/search/service_test.go` | ⬜ pending |
| 25-03-T1 | 03 | 3 | CRT-09, CRT-11, COMPAT-04 | T-25-12, T-25-13, T-25-14 | Retained tree cleared with its parameters; partial write is idempotent | unit (tdd) | `go test . -count=1 -race` | ✅ extend `app_test.go` | ⬜ pending |
| 25-03-T2 | 03 | 3 | CRT-13 | T-25-15 | Bounded wait on a spawned goroutine; hook never blocks the UI thread | unit (tdd) + manual | `go test . -count=1 -race` | ✅ extend `app_test.go` | ⬜ pending |
| 25-03-T3 | 03 | 3 | CRT-09, CRT-11 | — | Every new bridge entry routes through the shared error wrapper | build | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-04-T1 | 04 | 4 | CRT-02 | T-25-17, T-25-18 | Boot-volume and vendor-internal mounts filtered; single-entry readability probe | unit (tdd) + cross-build | `go test ./internal/volumes/... -count=1 && GOOS=windows GOARCH=amd64 go build ./internal/volumes/ && GOOS=linux GOARCH=amd64 go build ./internal/volumes/` | ❌ create `internal/volumes/volumes_test.go`, `internal/volumes/volumes_darwin_test.go` | ⬜ pending |
| 25-04-T2 | 04 | 4 | CRT-02, CRT-03, CRT-07 | T-25-16, T-25-19 | Pre-pass is context-cancellable; volume metadata is read-only | unit (tdd) | `go test ./internal/catalog/... . -count=1 -race` | ❌ create `internal/catalog/measure_test.go`; ✅ extend `app_test.go` | ⬜ pending |
| 25-04-T3 | 04 | 4 | CRT-02 | — | Unverifiable platform paths logged, never claimed | ledger assertion | `gsd-tools windows status` reports 4 open entries | ✅ `.planning/WINDOWS.md` | ⬜ pending |
| 25-05-T1 | 05 | 5 | CRT-02, CRT-03 | T-25-21 | Card list built from read-only mount metadata; staleness-guarded enumeration | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-05-T2 | 05 | 5 | CRT-04 | T-25-01, T-25-21 | Paths built by joining, never concatenation; existing-file qualifier before a replace | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-05-T3 | 05 | 5 | CRT-05, CRT-06 | T-25-20 | Persisted secondary location validated by the binding's containment gate | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-06-T1 | 06 | 6 | CRT-07 | T-25-22, T-25-23 | Volume-sourced paths rendered as text children only; log capped | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-06-T2 | 06 | 6 | CRT-09 | T-25-14 | Cancel reaches the mutex-guarded handle; nothing written | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-06-T3 | 06 | 6 | CRT-08 | T-25-07 | One rounding helper shared by both surfaces | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-07-T1 | 07 | 7 | CRT-10, CRT-11 | T-25-13, T-25-22, T-25-24 | Partial write fires once; retry never overlaps a live walk | build + live + manual (eject) | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-07-T2 | 07 | 7 | CRT-12 | T-25-12, T-25-25 | Partial can never be mistaken for complete — marker, tag, and stop-point summary | build + live | `cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |
| 25-07-T3 | 07 | 7 | CRT-01 | all of the above | Full-phase regression and containment sweep | full suite | `go build ./... && go test ./... -count=1 -race && GOOS=windows GOARCH=amd64 go build ./internal/volumes/ && GOOS=linux GOARCH=amd64 go build ./internal/volumes/ && cd frontend && npx tsc --noEmit && npm run build` | ✅ toolchain | ⬜ pending |

**Sampling continuity:** no three consecutive tasks lack an automated verify — every task above carries one. The two hardware-dependent behaviours (volume disappearance, force quit) ride on tasks that also carry an automated command, so an automated signal never goes more than one task without firing.

**Standing regression gates**, asserted at every Go task and again at 25-07-T3:
- `test -z "$(git diff --stat -- cli/create.go)"` — the CLI compatibility anchor is unedited.
- `test -z "$(go list -deps ./internal/catalog/... | grep -i wailsapp)"` — COMPAT-04's import boundary holds.

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

Every gap the research pass identified is created by the task that needs it, inside its own plan — there is no separate scaffolding wave, because each missing test file belongs to exactly one task and creating it there keeps the red-green cycle inside the task that owns the behaviour.

- [ ] `internal/catalog/service_test.go` extensions — cancellation and wrapper defaults (25-01-T1), terminal-error classification and the partial-marker `omitempty` shape and the JSON byte-shape assertion (25-02-T2)
- [ ] `internal/catalog/atomicwrite_test.go` — new file, created by 25-02-T1
- [ ] `internal/catalog/measure_test.go` — new file, created by 25-04-T2
- [ ] `internal/volumes/volumes_test.go` and `internal/volumes/volumes_darwin_test.go` — new package's tests, created by 25-04-T1; the Windows and Linux implementations are proven by cross-build in the same task rather than by a test the toolchain cannot run here
- [ ] `internal/search/service_test.go` extensions — reader degradation against the marker fields, created by 25-02-T2
- [ ] `app_test.go` extensions — binding, containment, cancellation, partial-write idempotency and close-hook helpers (25-01-T2, 25-03-T1, 25-03-T2, 25-04-T2)
- [ ] The COMPAT-04 guard — a dependency-graph assertion rather than a Go test, wired into 25-01-T2's acceptance criteria and re-asserted at every later Go task

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
| **Windows disk-space and drive-letter enumeration** | CRT-02 | No Windows machine or VM available (`25-RESEARCH.md` A3) | Not runnable this phase. Proven only by `GOOS=windows GOARCH=amd64 go build ./internal/volumes/` in 25-04-T1 and logged as an open platform-ledger entry by 25-04-T3. Must not be claimed as working. |
| **Linux mount enumeration heuristic** | CRT-02 | No Linux machine or VM available (`25-RESEARCH.md` A4) | Not runnable this phase. Proven only by `GOOS=linux GOARCH=amd64 go build ./internal/volumes/` in 25-04-T1 and logged as an open platform-ledger entry by 25-04-T3. Must not be claimed as working. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
