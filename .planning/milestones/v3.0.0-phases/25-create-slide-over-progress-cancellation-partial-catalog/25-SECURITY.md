---
phase: 25
slug: create-slide-over-progress-cancellation-partial-catalog
threats_total: 24
threats_closed: 24
threats_open: 0
asvs_level: 1
block_on: high
verdict: SECURED
register_authored_at_plan_time: true
audited: 2026-08-15
---

# Phase 25 — Security Audit

**Verdict: SECURED** — 24/24 threats closed, `threats_open: 0`.

## Verification method

All 7 `PLAN.md` threat models, all 7 `SUMMARY.md` files, `25-REVIEW.md`, `25-REVIEW-FIX.md` and `STATE.md` were read, and every declared `mitigate`/`accept` disposition was cross-referenced against **current HEAD source** (post-fix) rather than against SUMMARY claims. `go build ./...`, `go test ./... -race -count=1`, `npx tsc --noEmit`, and the `go list -deps | grep wailsapp` / `git diff --stat` guards were re-run independently.

Mode was `register_authored_at_plan_time: true` — verify declared mitigations exist, not scan for new threats — plus an adversarial sweep for unmapped surface (`os/exec` in new Go files, `dangerouslySetInnerHTML`/`innerHTML` across the create UI, `slugifyRoot`'s output sanitization), which found nothing outside the mapped register.

## Threat register

| Threat ID | Category | Severity | Disposition | Evidence |
|---|---|---|---|---|
| T-25-01 | Tampering | high | mitigate | `app.go:271-282` — `outputRoot`'s `.json`/`.html` destinations both run through `osutil.ContainsPath` before any walk starts; `TestStartScan_RejectsEscapingOutputRoot` (`app_test.go:199`) passing |
| T-25-02 | Tampering | medium | mitigate | `app.go:285-302` — `opts.CopyToDirectory` resolved via `EvalSymlinks` then `ContainsPath`-checked identically to the primary output |
| T-25-03 | DoS | medium | mitigate | `app.go:304-308` — `scanMu`-guarded `activeScanCancel != nil` rejects a second scan; `TestStartScan_RejectsSecondConcurrentScan` passing |
| T-25-04 | Tampering | medium | mitigate | `atomicwrite.go:23` — `os.CreateTemp` (exclusive, random suffix) in the **destination** directory, never a formatted name; `TestWriteFileAtomic_TempIsCreatedInDestinationDirectory` passing |
| T-25-05 | Info Disclosure | low | mitigate | `atomicwrite.go:31,36,44,49` — every post-create error path calls `os.Remove`; `TestWriteFileAtomic_RemovesTempOnFailure` passing |
| T-25-06 | Tampering | medium | mitigate | `service.go:491,513` — `html.EscapeString` on title and item name, unchanged from pre-phase |
| T-25-07 | Info Disclosure | low | accept | Single-user local desktop app; the webview is the app's own UI. Rationale documented in `25-01-PLAN.md` |
| T-25-10 | DoS | low | mitigate | `service.go:45,83-85` — `maxReadErrorEntries = 50` caps the recorded slice |
| T-25-11 | Tampering | medium | mitigate | `pkg/models/catalog.go:21,24` — both marker fields carry `omitempty`; `TestCreateCatalog_JSONShapeUnchanged` passing (COMPAT-02) |
| T-25-12 | Tampering | medium | mitigate | `app.go:312-321` — `StartScan` clears `lastPartial`/`lastPartialResult`/`lastScanReq` and bumps `retainedGen` before every scan; `TestStartScan_ClearsRetainedPartialOnNewScan` passing |
| T-25-13 | Tampering | medium | mitigate | `app.go:448-488` (post CR-02 fix `b411d2d9`) — `writeMu` serializes the whole check-decide-write-record sequence; `TestWritePartialCatalog_WritesOnce` and `TestWritePartialCatalog_ConcurrentCallsWriteOnce` (8-goroutine, `-race`) both passing |
| T-25-14 | Tampering | low | mitigate | `app.go:401-409` — cancel handle read/written only under `scanMu`, cleared only by the owning goroutine's deferred cleanup |
| T-25-15 | DoS | low | accept | `app.go:628-634` — 3-second bounded wait on a spawned goroutine; the `beforeClose` hook returns immediately |
| T-25-16 | Info Disclosure | low | accept | `app.go:661-663` — `ListVolumes` is a thin delegate exposing mount metadata only |
| T-25-17 | DoS | medium | mitigate | `volumes.go:114-131` — `probeReadable` reads at most one directory entry; `diskUsage` failures return 0/0 and never abort enumeration |
| T-25-18 | Tampering | medium | mitigate | `volumes_darwin.go:36-47` — boot-volume symlink and `com.apple.`-prefixed entries excluded; `TestList_ExcludesBootVolumeSymlink` live-verified against this machine's real `/Volumes` |
| T-25-19 | DoS | low | accept | `measure.go` — `ctx.Err()` checked at every directory boundary, on the same cancellable context `CancelScan` reaches |
| T-25-20 | Tampering | medium | mitigate | Same containment gate as T-25-02 — the persisted secondary-directory value only reaches the binding as `CopyToDirectory`, which is containment-checked |
| T-25-21 | Info Disclosure | low | accept | The preview's "already exists" check reads only the already-loaded rail listing for the user's own configured directory |
| T-25-22 | Tampering | medium | mitigate | `grep -rn dangerouslySetInnerHTML frontend/src/` returns **zero** matches; `ScanningBody.tsx`, `ErrorBody.tsx`, `DoneBody.tsx` render paths as plain JSX text children (the Phase 24 T-24-11 discipline, held) |
| T-25-23 | DoS | low | mitigate | `AppContext.tsx:123,295` — `SCAN_LOG_CAP = 9` enforced in the reducer itself so state never grows unbounded, plus a defensive re-slice in `ScanningBody.tsx:16,29` |
| T-25-24 | Tampering | medium | mitigate | `app.go` `StartScan`'s `scanMu` rejection (T-25-03) + `CreateSlideOver.tsx:186` `submittingRef` guard; `ErrorBody.tsx:104` "Retry scan" additionally `disabled={writingPartial}` (WR-02 fix) |
| T-25-25 | Spoofing | medium | mitigate | `DoneBody.tsx:41-47,66` — three independent signals distinguish a partial from a complete catalog: the on-disk `Unreadable`/`ReadError` marker, the visible `partial` tag, and the stop-percent clause replacing duration |
| T-25-SC | Tampering | high | mitigate | `git diff --stat 0268f0cd..HEAD -- go.mod go.sum frontend/package.json frontend/package-lock.json` — empty, verified directly rather than from a SUMMARY claim. No package install occurred; the supply-chain row is genuinely n/a, not waived |

## Post-review fixes independently re-verified in current source

Not trusted from `25-REVIEW-FIX.md` — read at HEAD:

- **CR-01** (`copyFile` bypassed the atomic write, violating CONTEXT's locked crash-safe-writes decision) — `service.go:612-621` now reads `src` fully and calls `WriteFileAtomic`; `TestCopyFile_PreservesExistingDestinationOnFailure` present and passing.
- **CR-02** (`WritePartialCatalog` check-then-act race) — `writeMu` + `retainedGen` generation guard present; folded into T-25-13.
- **WR-01** (stale ⌘↵ closure) — `handleCreateRef` pattern at `CreateSlideOver.tsx:335-336,353`.
- **WR-02** (unguarded Retry / Close-without-writing during an in-flight partial write) — `ErrorBody.tsx:104,116` both carry `disabled={writingPartial}`.

## Carried obligations — recorded, not fixed, not re-accepted

- **T-22-05** — `CatalogModal`'s unsanitized `srcDoc` confirmed still present at `CatalogModal.tsx:133` (`srcDoc={htmlContent}`, no sanitizer). Phase 25 added **no** new dispatcher for `openCatalogModal`, so the risk is unchanged and still dormant. Correctly deferred to Phase 26 — not touched, not silently fixed, not re-accepted a fourth time.
- **FU-23-A** — `GetCatalogHtmlPath` / `OpenExternal` (`app.go:674-694`) confirmed to still lack the containment gate `RevealInFileManager` received. Real, open, correctly assigned to Phase 26.

  **Phase 25's four new bindings were checked for wrongful omission from that sweep, and the omission is justified rather than an oversight:** `StartScan` receives its own containment gate at introduction (T-25-01/T-25-02), `WritePartialCatalog` takes no renderer-supplied path parameter, and `CancelScan` / `ListVolumes` take no path parameter at all. None of the four needs the FU-23-A treatment.

## Unregistered flags

None. No `## Threat Flags` section exists in any of the 7 SUMMARY files.

## Minor observation (non-blocking)

T-25-18's mitigation text says the boot-volume/vendor filter is "unit-tested on every platform," but `TestList_ExcludesBootVolumeSymlink` is **darwin-only**. Windows (`GetLogicalDrives`) and Linux enumeration remain compile-verified only. This is not a new gap — the phase already disclosed it transparently as `.planning/WINDOWS.md` entries **#4** and **#5** — but the mitigation wording overstates the coverage slightly and should be read against those ledger entries.
