---
status: accepted
phase: 28-re-scan-diff
source: [28-VERIFICATION.md]
started: 2026-08-16T23:20:00Z
updated: 2026-08-16T23:20:00Z
---

## Current Test

number: 1
name: Volume enumeration in RescanDialog step 1 is fast enough that no loading state is needed
expected: |
  wailsAPI.listVolumes() resolves quickly enough in practice — on real removable media,
  not just this dev machine — that the absence of a loading spinner in step 1 never reads
  as a frozen dialog.
awaiting: none — user accepted backstop items 2026-08-16

## Tests

### 1. Volume enumeration needs no dedicated loading state

expected: `wailsAPI.listVolumes()` resolves quickly enough on real removable media (external
drives, optical discs, network volumes) that step 1's lack of a loading spinner never reads as
a frozen dialog.

why_human: 28-01-PLAN.md marks this `verification: backstop` — a timing/UX judgment about real
removable-media latency, observable only against real hardware, not inferable from source and
deliberately not pinned by a timing assertion.

how_to_test: Open Re-scan with a slow external or optical volume mounted. Watch step 1 render.
If it ever looks frozen rather than instant, this fails.

result: [accepted-deferred] user accepted; observe in real use

### 2. The 0.6 similarity ratio and 20-entry floor flag a wrong disc without false positives

expected: Over real usage, the wrong-disc banner fires when you actually inserted the wrong
disc, and stays quiet for legitimate large changes.

why_human: 28-03-PLAN.md marks this `verification: backstop` — `similarityThreshold = 0.6` and
`similarityMinEntries = 20` in `internal/catalog/diff.go` are untuned defaults the plan itself
flagged as needing real-world observation. Not inferable from source or a unit test.

how_to_test: Over normal use, note any re-scan where the banner appeared but the volume was
correct (false positive), or where you did swap discs and it stayed quiet (false negative).
Both constants are named package-level and greppable for tuning.

result: [accepted-deferred] user accepted; observe in real use

## Summary

total: 2
passed: 0
issues: 0
pending: 0
accepted_deferred: 2
skipped: 0
blocked: 0

## Gaps

None. Verification found no gaps — 13/15 must-haves verified against the codebase. These two
items are deferred observations the plans themselves designated as backstop, not defects.
