---
schema_version: 1
open_count: 5
waived_count: 0
fixed_count: 0
total_count: 5
last_updated: 2026-08-15T00:48:29.765Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 23 | deviation | internal/osutil/reveal.go |  | Windows explorer /select,<path> argv shape (23-RESEARCH.md Assumption A1) unit-tested for structure only, not runtime-verified -- no Windows machine/VM available; sweep before v3.0.0 ships | open |  | 2026-08-14T02:17:12.499Z |  |
| 2 | 24 | deviation | frontend/src/components/workspace/WorkspaceShell.tsx |  | Ctrl+K open path (PLT-01, non-macOS) not runtime-verified -- the global keydown listener handles metaKey\|\|ctrlKey and macOS Cmd-K was confirmed in the real WKWebView window, but no Windows or Linux machine was available to press Ctrl+K in a native webview; sweep before v3.0.0 ships | open |  | 2026-08-14T15:23:41.035Z |  |
| 3 | 25 | unrun-verify | app.go |  | CRT-13 force-quit-mid-scan live verification not performed this session (wails dev not running at task start; executor instructed not to start a long-lived dev server) -- beforeClose's cancel-then-wait-then-requery sequence is unit-tested in isolation only, not proven end-to-end | open |  | 2026-08-15T00:33:49.624Z |  |
| 4 | 25 | deviation | internal/volumes/volumes_windows.go |  | Windows disk-space and drive-letter path: GetDiskFreeSpaceExW (via stdlib syscall.NewLazyDLL/LazyProc, not golang.org/x/sys) and GetLogicalDrives-based drive-letter enumeration compile under GOOS=windows GOARCH=amd64 (cross-build verified this session) but no Windows machine or VM was available to run them -- runtime behavior (correct byte totals, correct drive-letter set, permission/removed-media edge cases) is unverified and must be swept before v3.0.0 ships | open |  | 2026-08-15T00:48:22.173Z |  |
| 5 | 25 | deviation | internal/volumes/volumes_linux.go |  | Linux volume-enumeration heuristic: the /mnt and /media (including its per-user second level) removable-media roots, cross-checked against /proc/mounts, compile and are reasoned about statically only -- no Linux machine or VM was available this session to run them. This is a genuine, symmetrical unverified-platform risk 25-RESEARCH.md's Assumption A4 flagged as not previously tracked in this ledger (which began Windows-only); recorded here rather than left untracked. Must be swept before v3.0.0 ships, alongside entry 4's Windows gap | open |  | 2026-08-15T00:48:29.765Z |  |

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
    "status": "open",
    "reason": "",
    "recorded_at": "2026-08-15T00:33:49.624Z",
    "resolved_at": null
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
  }
]
````
