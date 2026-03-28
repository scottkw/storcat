---
status: partial
phase: 20-windows-cli-path-via-nsis
source: [20-VERIFICATION.md]
started: 2026-03-28T07:00:00Z
updated: 2026-03-28T07:00:00Z
---

## Current Test

[awaiting human testing]

## Tests

### 1. Install StorCat and run `storcat version` from new terminal
expected: Command prints version info without manual PATH changes
result: [pending]

### 2. Check System PATH in Environment Variables UI
expected: StorCat install directory (e.g. C:\Program Files\kenscott\StorCat) appears in System PATH
result: [pending]

### 3. Uninstall StorCat and verify PATH cleanup
expected: StorCat install directory removed from System PATH after uninstall
result: [pending]

### 4. CI NSIS compilation succeeds on Windows runner
expected: `wails build -clean -platform windows/amd64 -windowsconsole -nsis` completes with no NSIS errors; makensis finds EnVar.dll
result: [pending]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps
