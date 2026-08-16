---
schema_version: 1
open_count: 11
waived_count: 0
fixed_count: 2
total_count: 13
last_updated: 2026-08-16T00:23:51.065Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 23 | deviation | internal/osutil/reveal.go |  | Windows explorer /select,<path> argv shape (23-RESEARCH.md Assumption A1) unit-tested for structure only, not runtime-verified -- no Windows machine/VM available; sweep before v3.0.0 ships | open |  | 2026-08-14T02:17:12.499Z |  |
| 2 | 24 | deviation | frontend/src/components/workspace/WorkspaceShell.tsx |  | Ctrl+K open path (PLT-01, non-macOS) not runtime-verified -- the global keydown listener handles metaKey\|\|ctrlKey and macOS Cmd-K was confirmed in the real WKWebView window, but no Windows or Linux machine was available to press Ctrl+K in a native webview; sweep before v3.0.0 ships | open |  | 2026-08-14T15:23:41.035Z |  |
| 3 | 25 | unrun-verify | app.go |  | CRT-13 force-quit-mid-scan live verification not performed this session (wails dev not running at task start; executor instructed not to start a long-lived dev server) -- beforeClose's cancel-then-wait-then-requery sequence is unit-tested in isolation only, not proven end-to-end | fixed |  | 2026-08-15T00:33:49.624Z | 2026-08-15T02:39:39.103Z |
| 4 | 25 | deviation | internal/volumes/volumes_windows.go |  | Windows disk-space and drive-letter path: GetDiskFreeSpaceExW (via stdlib syscall.NewLazyDLL/LazyProc, not golang.org/x/sys) and GetLogicalDrives-based drive-letter enumeration compile under GOOS=windows GOARCH=amd64 (cross-build verified this session) but no Windows machine or VM was available to run them -- runtime behavior (correct byte totals, correct drive-letter set, permission/removed-media edge cases) is unverified and must be swept before v3.0.0 ships | open |  | 2026-08-15T00:48:22.173Z |  |
| 5 | 25 | deviation | internal/volumes/volumes_linux.go |  | Linux volume-enumeration heuristic: the /mnt and /media (including its per-user second level) removable-media roots, cross-checked against /proc/mounts, compile and are reasoned about statically only -- no Linux machine or VM was available this session to run them. This is a genuine, symmetrical unverified-platform risk 25-RESEARCH.md's Assumption A4 flagged as not previously tracked in this ledger (which began Windows-only); recorded here rather than left untracked. Must be swept before v3.0.0 ships, alongside entry 4's Windows gap | open |  | 2026-08-15T00:48:29.765Z |  |
| 6 | 25 | unrun-verify | internal/catalog/atomicwrite.go |  | Atomic write survives a real process kill mid-write (CRT-11/Phase 27 ACT-09) not empirically verified -- timing a SIGKILL inside WriteFileAtomic's few-ms temp-then-rename window is not reliably schedulable in this environment; the guarantee rests on unit tests (TestWriteFileAtomic_RemovesTempOnFailure, TestWriteFileAtomic_LeavesNoTempResidue) and os.Rename's OS-level atomicity, not a live kill test | fixed |  | 2026-08-15T02:39:08.862Z | 2026-08-16T00:03:50.836Z |
| 7 | 26 | unrun-verify | internal/config/config.go |  | 26-02 Task 2: full OS quit-and-relaunch proof of RailSide persistence (no visible flash) not performed live -- verified instead via TestSetRailSide_Persists (same Load() path) and on-disk config.json readback, to avoid disrupting the shared wails dev process plans 26-03..05 depend on | open |  | 2026-08-15T13:54:20.843Z |  |
| 8 | 27 | deviation | internal/watch/watcher.go |  | fsnotify Windows rename-release divergence (WATCH-03): fsnotify auto-removes a watch when the watched path itself is deleted or renamed on macOS/Linux, but its Windows backend does NOT remove the watcher on a rename of the watched directory (documented in fsnotify's own source, upstream behavior, not something internal/watch closes). If a user renames the catalog directory itself while StorCat is watching it outside of Settings, macOS/Linux drop the watch cleanly while Windows would keep a dangling watch on the old handle. applyWatchState's existing stop-and-restart-on-directory-change rule (the user-triggered path via Settings) is unaffected; the gap is specifically the watched directory vanishing out from under an active watch with no explicit directory-change action. No Windows machine was available to verify the resulting behavior. Sweep before v3.0.0 ships. | open |  | 2026-08-16T00:04:19.774Z |  |
| 9 | 27 | deviation | internal/watch/watcher.go |  | fsnotify Windows and Linux backends runtime-unverified (WATCH-02/WATCH-03): the ReadDirectoryChangesW (Windows) and inotify (Linux) backends compile for their targets but no Windows or Linux machine was available this session to run them; only the macOS kqueue backend was exercised live (dev-browser against wails dev, real touch/rm/cp-burst producing single debounced catalogs:changed events). Same class as existing entries 1, 2, 4, 5. Sweep before v3.0.0 ships. | open |  | 2026-08-16T00:04:19.902Z |  |
| 10 | 27 | deviation | internal/osutil/trash.go |  | wastebasket Windows and Linux backends runtime-unverified (ACT-04/ACT-05): the SHFileOperationW (Windows) and FreeDesktop-trash-spec (Linux) backends were read in full at research time and cross-build cleanly, but only the macOS osascript/Finder backend was actually exercised live this phase (real Trash move, real induced failure via chflags uchg, real retry). Runtime behavior on the other two platforms is unverified. Sweep before v3.0.0 ships. | open |  | 2026-08-16T00:04:20.018Z |  |
| 11 | 27 | deviation | internal/catalog/atomicwrite.go |  | WriteFileAtomic parent-directory fsync unsupported on Windows (ACT-09): the best-effort syncDir(filepath.Dir(path)) added in 27-02 works on Linux and macOS (confirmed live via the SIGKILL harness on this macOS host), but Windows does not expose a directory handle that can be fsync'd the same way, so the call is expected to fail there and its error is deliberately discarded (the temp file's own tmp.Sync() still provides the primary durability guarantee on Windows). The discard path itself is unverified on Windows -- no Windows machine was available this session. Sweep before v3.0.0 ships. | open |  | 2026-08-16T00:04:20.134Z |  |
| 12 | 27 | deviation | internal/osutil/trash.go |  | wastebasket's macOS AppleScript interpolation is an accepted, upstream-owned residual risk, not a defect to fix (ACT-04/ACT-05, threat T-27-01): the library's macOS backend builds an AppleScript string escaping only literal double-quote characters before osascript -e. This phase's mitigation is the osutil.ContainsPath containment gate applied to every path before the library is ever reached, which bounds the worst case but does not close the interpolation itself -- it is third-party code this project does not own. A future feature letting a user type an arbitrary filename that reaches the trash binding would need this revisited. | open |  | 2026-08-16T00:04:20.253Z |  |
| 13 | 27 | unmet-truth | frontend/src/components/workspace/Menu.tsx |  | Menu click-outside focus-restore does not reliably survive in this session's live re-test (matrix row 5, 27-07): 27-04-SUMMARY.md's coverage D1 claimed 'focus restore to the trigger button on every close path' verified live and passing, but re-testing this phase in the same Chromium-based dev-browser environment against wails dev :34115 shows the trigger DOES receive focus() momentarily during the close (confirmed via a focus()-call instrumentation log showing ws-details-overflow.focus() firing) but then loses it again to document.body by the time the click gesture's own native focus-follows-click default action completes, landing on <body> instead. Escape-driven close (a keyboard-only dismissal with no competing native click-focus event) restores focus reliably every time. Not fixed here -- Menu.tsx/useModalBehavior.ts are outside this plan's files_modified scope. Not confirmed whether this reproduces in the actual shipped app's WKWebView engine on macOS (Chromium and WebKit can differ on native focus-follows-click timing) -- host-OS GUI automation is prohibited by this project's standing constraint, so the real WKWebView window could not be driven to check. Needs a manual click-through in the built app before treating as a confirmed regression. | open |  | 2026-08-16T00:23:51.065Z |  |

````json
[
  {
    "id": 1,
    "kind": "deviation",
    "phase": "23",
    "file": "internal/osutil/reveal.go",
    "line": null,
    "description": "Windows explorer /select,<path> argv shape (23-RESEARCH.md Assumption A1) unit-tested for structure only, not runtime-verified -- no Windows machine/VM available; sweep before v3.0.0 ships",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-14T02:17:12.499Z",
    "resolved_at": null
  },
  {
    "id": 2,
    "kind": "deviation",
    "phase": "24",
    "file": "frontend/src/components/workspace/WorkspaceShell.tsx",
    "line": null,
    "description": "Ctrl+K open path (PLT-01, non-macOS) not runtime-verified -- the global keydown listener handles metaKey||ctrlKey and macOS Cmd-K was confirmed in the real WKWebView window, but no Windows or Linux machine was available to press Ctrl+K in a native webview; sweep before v3.0.0 ships",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-14T15:23:41.035Z",
    "resolved_at": null
  },
  {
    "id": 3,
    "kind": "unrun-verify",
    "phase": "25",
    "file": "app.go",
    "line": null,
    "description": "CRT-13 force-quit-mid-scan live verification not performed this session (wails dev not running at task start; executor instructed not to start a long-lived dev server) -- beforeClose's cancel-then-wait-then-requery sequence is unit-tested in isolation only, not proven end-to-end",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-08-15T00:33:49.624Z",
    "resolved_at": "2026-08-15T02:39:39.103Z"
  },
  {
    "id": 4,
    "kind": "deviation",
    "phase": "25",
    "file": "internal/volumes/volumes_windows.go",
    "line": null,
    "description": "Windows disk-space and drive-letter path: GetDiskFreeSpaceExW (via stdlib syscall.NewLazyDLL/LazyProc, not golang.org/x/sys) and GetLogicalDrives-based drive-letter enumeration compile under GOOS=windows GOARCH=amd64 (cross-build verified this session) but no Windows machine or VM was available to run them -- runtime behavior (correct byte totals, correct drive-letter set, permission/removed-media edge cases) is unverified and must be swept before v3.0.0 ships",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-15T00:48:22.173Z",
    "resolved_at": null
  },
  {
    "id": 5,
    "kind": "deviation",
    "phase": "25",
    "file": "internal/volumes/volumes_linux.go",
    "line": null,
    "description": "Linux volume-enumeration heuristic: the /mnt and /media (including its per-user second level) removable-media roots, cross-checked against /proc/mounts, compile and are reasoned about statically only -- no Linux machine or VM was available this session to run them. This is a genuine, symmetrical unverified-platform risk 25-RESEARCH.md's Assumption A4 flagged as not previously tracked in this ledger (which began Windows-only); recorded here rather than left untracked. Must be swept before v3.0.0 ships, alongside entry 4's Windows gap",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-15T00:48:29.765Z",
    "resolved_at": null
  },
  {
    "id": 6,
    "kind": "unrun-verify",
    "phase": "25",
    "file": "internal/catalog/atomicwrite.go",
    "line": null,
    "description": "Atomic write survives a real process kill mid-write (CRT-11/Phase 27 ACT-09) not empirically verified -- timing a SIGKILL inside WriteFileAtomic's few-ms temp-then-rename window is not reliably schedulable in this environment; the guarantee rests on unit tests (TestWriteFileAtomic_RemovesTempOnFailure, TestWriteFileAtomic_LeavesNoTempResidue) and os.Rename's OS-level atomicity, not a live kill test",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-08-15T02:39:08.862Z",
    "resolved_at": "2026-08-16T00:03:50.836Z"
  },
  {
    "id": 7,
    "kind": "unrun-verify",
    "phase": "26",
    "file": "internal/config/config.go",
    "line": null,
    "description": "26-02 Task 2: full OS quit-and-relaunch proof of RailSide persistence (no visible flash) not performed live -- verified instead via TestSetRailSide_Persists (same Load() path) and on-disk config.json readback, to avoid disrupting the shared wails dev process plans 26-03..05 depend on",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-15T13:54:20.843Z",
    "resolved_at": null
  },
  {
    "id": 8,
    "kind": "deviation",
    "phase": "27",
    "file": "internal/watch/watcher.go",
    "line": null,
    "description": "fsnotify Windows rename-release divergence (WATCH-03): fsnotify auto-removes a watch when the watched path itself is deleted or renamed on macOS/Linux, but its Windows backend does NOT remove the watcher on a rename of the watched directory (documented in fsnotify's own source, upstream behavior, not something internal/watch closes). If a user renames the catalog directory itself while StorCat is watching it outside of Settings, macOS/Linux drop the watch cleanly while Windows would keep a dangling watch on the old handle. applyWatchState's existing stop-and-restart-on-directory-change rule (the user-triggered path via Settings) is unaffected; the gap is specifically the watched directory vanishing out from under an active watch with no explicit directory-change action. No Windows machine was available to verify the resulting behavior. Sweep before v3.0.0 ships.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-16T00:04:19.774Z",
    "resolved_at": null
  },
  {
    "id": 9,
    "kind": "deviation",
    "phase": "27",
    "file": "internal/watch/watcher.go",
    "line": null,
    "description": "fsnotify Windows and Linux backends runtime-unverified (WATCH-02/WATCH-03): the ReadDirectoryChangesW (Windows) and inotify (Linux) backends compile for their targets but no Windows or Linux machine was available this session to run them; only the macOS kqueue backend was exercised live (dev-browser against wails dev, real touch/rm/cp-burst producing single debounced catalogs:changed events). Same class as existing entries 1, 2, 4, 5. Sweep before v3.0.0 ships.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-16T00:04:19.902Z",
    "resolved_at": null
  },
  {
    "id": 10,
    "kind": "deviation",
    "phase": "27",
    "file": "internal/osutil/trash.go",
    "line": null,
    "description": "wastebasket Windows and Linux backends runtime-unverified (ACT-04/ACT-05): the SHFileOperationW (Windows) and FreeDesktop-trash-spec (Linux) backends were read in full at research time and cross-build cleanly, but only the macOS osascript/Finder backend was actually exercised live this phase (real Trash move, real induced failure via chflags uchg, real retry). Runtime behavior on the other two platforms is unverified. Sweep before v3.0.0 ships.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-16T00:04:20.018Z",
    "resolved_at": null
  },
  {
    "id": 11,
    "kind": "deviation",
    "phase": "27",
    "file": "internal/catalog/atomicwrite.go",
    "line": null,
    "description": "WriteFileAtomic parent-directory fsync unsupported on Windows (ACT-09): the best-effort syncDir(filepath.Dir(path)) added in 27-02 works on Linux and macOS (confirmed live via the SIGKILL harness on this macOS host), but Windows does not expose a directory handle that can be fsync'd the same way, so the call is expected to fail there and its error is deliberately discarded (the temp file's own tmp.Sync() still provides the primary durability guarantee on Windows). The discard path itself is unverified on Windows -- no Windows machine was available this session. Sweep before v3.0.0 ships.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-16T00:04:20.134Z",
    "resolved_at": null
  },
  {
    "id": 12,
    "kind": "deviation",
    "phase": "27",
    "file": "internal/osutil/trash.go",
    "line": null,
    "description": "wastebasket's macOS AppleScript interpolation is an accepted, upstream-owned residual risk, not a defect to fix (ACT-04/ACT-05, threat T-27-01): the library's macOS backend builds an AppleScript string escaping only literal double-quote characters before osascript -e. This phase's mitigation is the osutil.ContainsPath containment gate applied to every path before the library is ever reached, which bounds the worst case but does not close the interpolation itself -- it is third-party code this project does not own. A future feature letting a user type an arbitrary filename that reaches the trash binding would need this revisited.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-16T00:04:20.253Z",
    "resolved_at": null
  },
  {
    "id": 13,
    "kind": "unmet-truth",
    "phase": "27",
    "file": "frontend/src/components/workspace/Menu.tsx",
    "line": null,
    "description": "Menu click-outside focus-restore does not reliably survive in this session's live re-test (matrix row 5, 27-07): 27-04-SUMMARY.md's coverage D1 claimed 'focus restore to the trigger button on every close path' verified live and passing, but re-testing this phase in the same Chromium-based dev-browser environment against wails dev :34115 shows the trigger DOES receive focus() momentarily during the close (confirmed via a focus()-call instrumentation log showing ws-details-overflow.focus() firing) but then loses it again to document.body by the time the click gesture's own native focus-follows-click default action completes, landing on <body> instead. Escape-driven close (a keyboard-only dismissal with no competing native click-focus event) restores focus reliably every time. Not fixed here -- Menu.tsx/useModalBehavior.ts are outside this plan's files_modified scope. Not confirmed whether this reproduces in the actual shipped app's WKWebView engine on macOS (Chromium and WebKit can differ on native focus-follows-click timing) -- host-OS GUI automation is prohibited by this project's standing constraint, so the real WKWebView window could not be driven to check. Needs a manual click-through in the built app before treating as a confirmed regression.",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-16T00:23:51.065Z",
    "resolved_at": null
  }
]
````
