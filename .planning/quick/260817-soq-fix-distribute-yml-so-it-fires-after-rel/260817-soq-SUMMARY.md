---
phase: quick-260817-soq
plan: 01
subsystem: infra
tags: [github-actions, ci-cd, workflow_run, homebrew, winget]

# Dependency graph
requires: []
provides:
  - workflow_run trigger on .github/workflows/distribute.yml, bound to the "Release" workflow's completion
  - job-level conclusion guard on all three distribute.yml jobs (update-homebrew, update-winget, update-winget-manifests)
  - three-operand TAG fallback chain (inputs.tag || release.tag_name || workflow_run.head_branch) on all three TAG computation sites
affects: [release process, homebrew tap publishing, winget manifest generation]

actuals:
  tokens: 1163
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "workflow_run trigger + job-level conclusion guard, used to react to a downstream workflow's completion when the upstream publish step suppresses its own event (default GITHUB_TOKEN)"

key-files:
  created: []
  modified:
    - .github/workflows/distribute.yml

key-decisions:
  - "Used workflow_run (not push: tags) because push: tags would race release.yml's asset upload, causing 404s on the DMG/installer downloads"
  - "Guard placed at job level (if: on runs-on's sibling), not step level, so a failed Release run skips checkout/download entirely rather than letting early steps run and fail messily"
  - "TAG fallback ordered inputs.tag -> release.tag_name -> workflow_run.head_branch so a manual dispatch always wins over a possibly-stale payload field"
  - "update-winget guarded identically to the other two jobs but left otherwise untouched — its version-inference limitation under workflow_run is documented in-file, not engineered around, since the submission step is currently inert (package not yet in winget-pkgs)"

patterns-established:
  - "Chesterton's Fence comment block above a trigger that looks redundant next to an existing one, explaining why both must stay"

requirements-completed: []

coverage:
  - id: D1
    description: "distribute.yml fires automatically via workflow_run when the Release workflow completes, without manual dispatch"
    verification:
      - kind: other
        ref: "python3 yaml parse assertion (Task 1 <verify>) confirming workflows: [Release], types: [completed] on the workflow_run trigger"
        status: pass
    human_judgment: true
    rationale: "No CI run of the actual trigger firing end-to-end was performed (would require a real Release workflow run, which the plan's constraints prohibit dispatching) — YAML structure is proven, live firing is not"
  - id: D2
    description: "A failed Release run cannot publish to Homebrew tap or commit WinGet manifests"
    verification:
      - kind: other
        ref: "python3 assertion (Task 2 <verify>) confirming 'workflow_run.conclusion' present in every job's if: and absent from every step's if:"
        status: pass
    human_judgment: false
  - id: D3
    description: "TAG resolves correctly on all three trigger paths (workflow_dispatch, release, workflow_run) across all three TAG computation sites"
    verification:
      - kind: other
        ref: "grep count assertion: exactly 3 non-comment lines carry the full inputs.tag || release.tag_name || workflow_run.head_branch chain"
        status: pass
    human_judgment: false

duration: ~10min
completed: 2026-08-18
status: complete
---

# Quick Task 260817-soq: Fix distribute.yml trigger Summary

**Added a `workflow_run` trigger to `.github/workflows/distribute.yml` bound to the "Release" workflow's completion, with job-level conclusion guards and an extended TAG fallback chain, so distribution no longer silently depends on a `release: published` event GitHub never emits for `GITHUB_TOKEN`-published releases.**

## Performance

- **Duration:** ~10 min
- **Completed:** 2026-08-18T01:46:11Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- `on:` block gained a `workflow_run` trigger (`workflows: ["Release"]`, `types: [completed]`) preceded by a multi-line comment explaining why it is not redundant with `release: published`
- All three jobs (`update-homebrew`, `update-winget`, `update-winget-manifests`) gained an identical job-level `if:` guard admitting the job only when the trigger isn't `workflow_run`, or when it is and the upstream Release run's conclusion was `success`
- All three TAG computation sites (one in `update-homebrew`, two in `update-winget-manifests`) extended with a third fallback operand, `github.event.workflow_run.head_branch`, appended after the existing `inputs.tag || github.event.release.tag_name` chain
- `update-winget` documented in-file: `winget-releaser` cannot resolve a version under `workflow_run`, currently inert since the package isn't yet in the winget-pkgs registry

## Task Commits

Each task was committed atomically:

1. **Task 1: Wire the workflow_run trigger end-to-end through update-homebrew** - `e81c78bf` (fix)
2. **Task 2: Extend guards and TAG resolution to the two WinGet jobs** - `d518ddd0` (fix)

_Tracer task (Task 1) verified via its own automated `<verify>` before proceeding to Task 2, per the tracer feedback gate — verification passed, no halt._

## Files Created/Modified
- `.github/workflows/distribute.yml` - Added `workflow_run` trigger with WHY comment, job-level conclusion guards on all three jobs, extended TAG fallback on all three computation sites, and an in-file limitation note on `update-winget`

## Decisions Made
- `workflow_run` chosen over `push: tags` specifically to avoid a race with `release.yml`'s asset uploads (documented rationale from the plan's `<background>`, carried into the in-file comment)
- Guard expression form: `github.event_name != 'workflow_run' || github.event.workflow_run.conclusion == 'success'` — placed at job level so unguarded early steps (checkout, download) never run after a failed Release
- No changes made to `release.yml`, `packaging/homebrew/`, or `packaging/winget/` — scope held exactly to `distribute.yml` per the plan's constraints

## Deviations from Plan

None - plan executed exactly as written. Both tasks' automated `<verify>` blocks and the whole-file verification checks (YAML parse, TAG operand order, additive-only diff, job-level guard placement) all passed on first attempt.

## Known Limitations (accepted, not open bugs)

Recorded verbatim from the plan's `<known_limitations>`:

1. **`release.yml` run via its own `workflow_dispatch`.** `github.event.workflow_run.head_branch` then holds the branch name (`main`), not a tag, so TAG resolves to `main`, the DMG URL 404s, and `update-homebrew` fails loudly rather than publishing something wrong. Distribute's own `workflow_dispatch` with an explicit `tag` input remains the manual escape hatch. Accepted — not engineered around.

2. **`update-winget` cannot function under `workflow_run`.** `vedantmgoyal9/winget-releaser` infers its version from the release event payload, which is absent on this trigger. Currently inert: `scottkw.StorCat` is not in the winget-pkgs registry, so the check step short-circuits and the submission step never runs on any trigger. This becomes a real gap only once the package is accepted upstream, at which point the job needs its own follow-up. Not in scope here.

3. **The `workflows: ["Release"]` binding is by display name.** Renaming `release.yml`'s `name:` field silently disconnects Distribute. There is no way to bind by file path in the `workflow_run` trigger.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required. No workflow was dispatched, no commit was pushed to the tap repo, and no outward-facing action was taken during this task, per the plan's constraints.

## Next Phase Readiness
- `distribute.yml` is ready to fire automatically on the next tag push once `release.yml` completes successfully
- Live end-to-end firing (an actual Release run triggering Distribute) has not been observed — will be proven organically on the next real release, or could be verified retroactively via `gh run list` after that release ships
- No further work queued by this task; the three known limitations above are accepted, not blockers

---
*Task: quick-260817-soq*
*Completed: 2026-08-18*

## Self-Check: PASSED

- FOUND: `.github/workflows/distribute.yml`
- FOUND: `.planning/quick/260817-soq-fix-distribute-yml-so-it-fires-after-rel/260817-soq-SUMMARY.md`
- FOUND commit: `e81c78bf`
- FOUND commit: `d518ddd0`
