---
phase: 12-repo-consolidation
plan: "02"
subsystem: infra
tags: [homebrew, packaging, cask, template, ruby, shell]

# Dependency graph
requires: []
provides:
  - "packaging/homebrew/storcat.rb.template — Homebrew cask template with {{VERSION}} and {{SHA256}} placeholders"
  - "packaging/homebrew/update-tap.sh — Script for pushing updated cask to homebrew-storcat tap (reference copy)"
affects:
  - phase-15-distribution-automation

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Homebrew cask template pattern: {{VERSION}}/{{SHA256}} shell placeholders + #{version} Ruby interpolation for URL"

key-files:
  created:
    - packaging/homebrew/storcat.rb.template
    - packaging/homebrew/update-tap.sh
  modified: []

key-decisions:
  - "Copy update-tap.sh as-is from satellite repo — cross-repo path adaptation deferred to Phase 15"
  - "Template uses {{VERSION}} and {{SHA256}} for substitution, preserves #{version} Ruby interpolation in URL line"
  - "Template does not contain hardcoded version 1.2.3 — all version/SHA references replaced with placeholders"

patterns-established:
  - "Cask template pattern: {{PLACEHOLDER}} for substitution targets, #{ruby_var} for Homebrew DSL interpolation"

requirements-completed:
  - REPO-02

# Metrics
duration: 2min
completed: 2026-03-27
---

# Phase 12 Plan 02: Homebrew Cask Template Summary

**Homebrew cask template with {{VERSION}}/{{SHA256}} placeholders migrated from scottkw/homebrew-storcat satellite repo into packaging/homebrew/**

## Performance

- **Duration:** 2 min
- **Started:** 2026-03-27T18:08:40Z
- **Completed:** 2026-03-27T18:10:00Z
- **Tasks:** 1 of 1
- **Files modified:** 2

## Accomplishments

- Created `packaging/homebrew/storcat.rb.template` from live `Casks/storcat.rb` in satellite repo, replacing hardcoded `1.2.3` version/SHA with `{{VERSION}}` and `{{SHA256}}` template placeholders
- Preserved `#{version}` Ruby DSL interpolation in URL line (Homebrew resolves this at install time, not substitution time)
- Copied `update-tap.sh` from satellite repo with execute permissions intact

## Task Commits

Each task was committed atomically:

1. **Task 1: Create Homebrew cask template and copy update script** - `93da56fc` (feat)

## Files Created/Modified

- `packaging/homebrew/storcat.rb.template` - Homebrew cask template with {{VERSION}} and {{SHA256}} placeholders; #{version} Ruby interpolation preserved in URL
- `packaging/homebrew/update-tap.sh` - Script that fetches latest release, computes SHA256, writes live cask, and commits to homebrew-storcat tap (reference copy — must be run from homebrew-storcat root)

## Decisions Made

- **Copy update-tap.sh as-is:** The script assumes it runs from the `homebrew-storcat` repo root (`CASK_FILE="Casks/storcat.rb"` relative path, `git add`/`git commit` operations). For Phase 12, it is copied verbatim as a reference. Cross-repo adaptation (making it work from the main repo and push to homebrew-storcat) is Phase 15 distribution automation work.
- **Template placeholder distinction:** `{{VERSION}}` and `{{SHA256}}` are shell-style substitution targets for the update script. `#{version}` is Homebrew Ruby DSL that the live cask uses at install time — these must not be conflated.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `packaging/homebrew/` is populated — REPO-02 satisfied
- Phase 15 (distribution automation) can reference `storcat.rb.template` and adapt `update-tap.sh` for cross-repo operation
- `homebrew-storcat` satellite repo remains live and unarchived (required for `brew tap scottkw/storcat` to continue working)

---
*Phase: 12-repo-consolidation*
*Completed: 2026-03-27*
