---
status: partial
phase: 18-windows-authenticode-signing
source: [18-VERIFICATION.md]
started: 2026-03-28T06:00:00Z
updated: 2026-03-28T06:00:00Z
---

## Current Test

[awaiting human testing]

## Tests

### 1. Store eSigner secrets in GitHub release environment
expected: `gh secret list --env release --repo scottkw/storcat` returns 10 entries including ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET
result: [pending]

### 2. Live CI signing run
expected: Push a release tag; both Sign portable exe and Sign NSIS installer steps exit 0; Verify Authenticode signatures step passes
result: [pending]

### 3. Signed binary verification on Windows
expected: `signtool verify /pa /v` reports "Successfully verified" with SSL.com OV certificate chain for both artifacts
result: [pending]

## Summary

total: 3
passed: 0
issues: 0
pending: 3
skipped: 0
blocked: 0

## Gaps
