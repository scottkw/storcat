---
schema_version: 1
open_count: 3
waived_count: 0
fixed_count: 0
total_count: 3
last_updated: 2026-08-15T00:33:49.624Z
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
  }
]
````
