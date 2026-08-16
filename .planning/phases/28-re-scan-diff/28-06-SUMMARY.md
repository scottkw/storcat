---
phase: 28-re-scan-diff
plan: 06
subsystem: infra
tags: [ci, github-actions, wails, release-please, broken-windows-ledger]

requires:
  - phase: 28-re-scan-diff
    plan: 05
    provides: "STATE-03 complete; the milestone's code frozen and ready to publish"
provides:
  - "A real, observed green build.yml run (all three legs individually success) cited by URL and commit SHA as COMPAT-06's build-leg evidence"
  - "local main published to origin/main (329 commits, Phases 22-28) -- GSD worktrees can now fork from a current origin"
  - "WINDOWS.md ledger honestly swept: entry #7 closed on real live evidence, #4/#5 reworded to native-runner compile evidence (still open), #10/#11 annotated with new call sites (still open), ten entries remain open"
affects: []

actuals:
  tokens: 7199
  tasks: 3
  commits: 1

tech-stack:
  added: []
  patterns:
    - "COMPAT-06 evidence recorded as trailing prose AFTER the WINDOWS.md ledger's closing JSON fence (not between the header and table), because gsd-tools windows CLI's writeLedgerAtomic only preserves content after the fence on every future append/waive/fixed write -- content placed anywhere else would be silently destroyed by the next CLI-driven ledger edit"

key-files:
  created: []
  modified:
    - .planning/WINDOWS.md
    - .planning/REQUIREMENTS.md
    - .planning/ROADMAP.md
    - .planning/STATE.md

key-decisions:
  - "COMPAT-06 requirement checkbox left UNCHECKED in REQUIREMENTS.md, and its Traceability row set to 'Partial -- build proven, sign/notarize/release open' rather than 'Complete'. The requirement's own literal text is 'builds, signs, notarizes, and releases' -- this plan's evidence proves only the build leg. Marking it fully Complete would be exactly the over-claim this plan's honesty_constraints prohibit; the standard execute-plan step of blindly running `requirements mark-complete` on every frontmatter requirement ID was deliberately skipped for COMPAT-06 for this reason."
  - "Original railSide=Right restored via the real SetRailSide binding after the live entry-#7 check, rather than left at the test's Left value or hand-edited on disk -- the check's own binding is the only legitimate way to mutate the user's real config.json."

requirements-completed: []

coverage:
  - id: D1
    description: "build.yml has actually run against this milestone's code (Phases 22-28) and is green on all three legs -- macOS universal, Windows amd64, Linux amd64 -- individually, with an artifact each, cited by run URL and commit SHA"
    requirement: "COMPAT-06"
    verification:
      - kind: other
        ref: "gh run view 31976677486 --json jobs: build-macos/build-linux/build-windows all conclusion=success, status=completed; gh api .../artifacts confirms StorCat-macOS (8175257B), StorCat-Windows-amd64 (5121734B), StorCat-Linux-amd64 (4285484B) all non-expired; git merge-base --is-ancestor 2715a42d HEAD succeeds"
        status: pass
    human_judgment: false
  - id: D2
    description: "The record states plainly what the green run proves (native compilation + tsc type-checking on three real runners) and what it does not (no signing, no notarization, no binary execution) -- signing/notarization recorded as an open ledger item, not claimed"
    requirement: "COMPAT-06"
    verification:
      - kind: other
        ref: ".planning/WINDOWS.md 'COMPAT-06 CI Evidence' section, both-halves statement plus explicit CRED-04/CRED-05 open citation; grep -ci 'signing|notariz' returns 2"
        status: pass
    human_judgment: false
  - id: D3
    description: "WINDOWS.md entry #7 (RailSide quit-and-relaunch) closed only after being performed live -- real Settings-UI toggle, real window.runtime.Quit(), real wails dev relaunch, zero flash of the wrong side observed"
    requirement: null
    verification:
      - kind: e2e
        ref: "dev-browser live session against wails dev :34115 -- toggled Right to Left via the real SegmentedControl click, persisted immediately to config.json; window.runtime.Quit() (process confirmed exited via ps/lsof); relaunched wails dev; 40 samples of .ws-root's data-rail-side at 50ms intervals from first paint all read 'Left', zero 'Right'; config.json read railSide:Left both pre- and post-quit"
        status: pass
    human_judgment: false
  - id: D4
    description: "Exactly one entry (#7) closed; #4/#5 reworded without closing; #10/#11 annotated without closing; ten entries remain open -- no other status changed"
    requirement: null
    verification:
      - kind: other
        ref: "grep -c '| open |' .planning/WINDOWS.md returns 10; gsd-tools windows status confirms open:10 fixed:4 waived:0 total:14 (fixed_count was 3 before this plan, now 4 -- only #7 moved)"
        status: pass
    human_judgment: false

duration: ~55min
completed: 2026-08-16
status: complete
---

# Phase 28 Plan 06: CI Proof & Honest Ledger Sweep Summary

**Local main (329 commits, the full v3.0.0 milestone) published to origin/main per the user's pre-made push-main decision; build.yml run 31976677486 observed green on all three native runners individually; WINDOWS.md swept to close exactly one entry (#7, live-verified) while leaving ten honestly open**

## Performance

- **Duration:** ~55 min
- **Started:** 2026-08-16T22:33:00Z (approx)
- **Completed:** 2026-08-16T23:15:00Z (approx)
- **Tasks:** 3 (Task 1 a pre-decided checkpoint, Tasks 2-3 auto)
- **Files modified:** 4 (`.planning/WINDOWS.md`, `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`)

## Accomplishments

- **Task 1 (pre-decided):** `git push origin main` executed once -- no `--force`, no `--no-verify`, no tag cut. 329 commits (`f9c7f024..2715a42d`) published, covering the entire Go/Wails v3.0.0 rewrite (Phases 22-28). A pre-push scan of `git log --stat origin/main..HEAD` for credential-like filenames found nothing (the sole grep hit was a benign `import.meta.env.DEV` code comment).
- **Task 2:** Watched `build.yml` run [31976677486](https://github.com/scottkw/storcat/actions/runs/31976677486) (push-triggered, commit `2715a42d`) to completion via `gh run watch`. All three jobs individually reported `success` and each uploaded a non-expired artifact (macOS 8,175,257B, Windows 5,121,734B, Linux 4,285,484B). Confirmed exactly one `build.yml` run exists for this SHA (no PR was opened, so no PR-triggered run to conflate). Confirmed `release.yml` was NOT triggered (no run newer than 2026-03-28) and `release-please.yml` opened/updated PR #3 (`chore(main): release 2.4.0`) -- the correct, tag-free precursor behavior. Recorded a "COMPAT-06 CI Evidence" section in `WINDOWS.md` stating both halves of the claim: what's proven (native compilation + `tsc` type-checking on three real runners) and what is not (signing, notarization, binary execution -- all tag-gated in `release.yml`, unfired by this push).
- **Task 3:** Swept the ledger. Performed entry #7's live quit-and-relaunch check via dev-browser (see below) and closed it -- the only entry closed. Reworded #4/#5 to cite the new native-runner compile evidence from this run while keeping both OPEN (runtime behavior on real Windows/Linux hardware remains unverified -- `build.yml` never executes the binary). Annotated #10/#11 with the new call sites Phase 28 added (remove-from-library -> delete-to-Trash; overwrite -> the atomic write) without changing their open status. Ten entries remain open, exactly as the plan's own success criteria call for as the honest outcome.

## Task Commits

1. **Task 1: Push local main to origin/main** - no commit (the push itself publishes existing commits; no local diff)
2. **Task 2 + Task 3: Record COMPAT-06 evidence and sweep WINDOWS.md** - `96282130` (docs) -- both tasks landed in one commit since they modify the same file's interleaved sections (evidence section, entry rewording, entry closure); each is independently described in the commit body

**Plan metadata:** (this commit)

## Files Created/Modified

- `.planning/WINDOWS.md` - COMPAT-06 CI evidence section added (trailing prose after the JSON fence); entry #7 closed; entries #4/#5 reworded; entries #10/#11 annotated
- `.planning/REQUIREMENTS.md` - COMPAT-06 checkbox left unchecked with a build-leg-only note; Traceability row set to "Partial"
- `.planning/ROADMAP.md` - Phase 28 plan checklist, plan count, and progress table updated
- `.planning/STATE.md` - position, decisions, metrics updated

## Decisions Made

- **COMPAT-06 not marked Complete in REQUIREMENTS.md.** The requirement's literal text is "builds, signs, notarizes, and releases" -- this plan's evidence proves only the build leg (a real, green, three-runner `build.yml` run). Marking it fully Complete would itself be the over-claim this plan's `<prohibitions>` explicitly forbid making about `WINDOWS.md`; the same honesty standard was applied to `REQUIREMENTS.md` even though the plan's text didn't name that file directly. Recorded instead as "Partial -- build proven, sign/notarize/release open" with a pointer to the WINDOWS.md evidence.
- **CI evidence recorded as trailing prose after WINDOWS.md's closing JSON fence**, not inserted between the header and the table. `gsd-tools windows`'s `writeLedgerAtomic` reconstructs frontmatter + fixed header + table + JSON block on every write and only explicitly preserves content *after* the closing fence (issue #2893) -- anything placed between the header and the table would be silently destroyed by the next `windows fixed`/`waive`/`append` call.
- **Original `railSide: Right` restored** after the live entry-#7 quit-and-relaunch check, via the real `SetRailSide` binding (not a raw file edit), so the user's actual config.json was left exactly as found.
- **Single commit for Task 2 + Task 3.** Both tasks' edits landed in the same file with interleaved sections (the evidence section is new trailing content; the entry rewording/closure edits are inline). Splitting into two commits would have required artificial hunk surgery on a single logical ledger update; the commit body describes both tasks distinctly instead.

## Deviations from Plan

### Auto-fixed Issues

None -- no bugs, missing functionality, or blocking issues encountered in the plan's own scope.

**Process note (not a deviation from the plan's instructions, a tooling gotcha worth recording):** the first `gsd_run windows fixed 7` invocation silently no-op'd because `$GSD_TOOLS`/`gsd_run` had been defined in an earlier Bash tool call and this tool's shell state does not persist between calls -- `node ""` with an empty path produced no output and no error. Caught immediately by checking `windows status` afterward and seeing `open_count` unchanged; re-ran with `GSD_TOOLS` explicitly re-set in the same call, which then correctly transitioned entry #7 from `open` to `fixed`. No incorrect ledger state was ever written to disk.

---

**Total deviations:** 0 auto-fixed. One tooling gotcha caught and corrected before any incorrect write.

## Issues Encountered

None that affected the outcome. The `wails dev` session for entry #7's live check started and quit cleanly on both cycles; no stale process was found on `:34115` or `:5173` at any point.

## Live Verification (dev-browser, wails dev on :34115) -- WINDOWS.md entry #7

Per this repo's CLAUDE.md dev-browser mandate and the plan's `<live_verification_for_entry_7>` constraints:

1. Confirmed no stale process on `:34115`/`:5173` before starting.
2. Started `wails dev`; probed `Object.keys(window.go.main.App)` immediately (35 bindings present, including `SetRailSide`/`GetConfig`) for binding freshness before recording anything.
3. Read on-disk `config.json`: `railSide: "Right"` (the pre-existing real value).
4. Opened the real Settings dialog via a UI click on the settings trigger button (found by `aria-label`/`title` matching "settings"), confirmed `.ws-rail-seg` rendered.
5. Clicked the real "Left" segment inside `.ws-rail-seg` (a genuine CDP-trusted click on the actual control, not a synthetic binding call) -- live config binding immediately reflected `railSide: "Left"`, and on-disk `config.json` matched within the same check.
6. Issued `window.runtime.Quit()` -- the app's own quit path. Confirmed the process fully exited (`ps aux` empty for `StorCat.app`/`wails dev`; `:34115` freed; the dev log printed "Development mode exited").
7. Re-read on-disk `config.json` post-quit: still `railSide: "Left"` -- proves the persistence survived the real quit, not just an in-memory write.
8. Relaunched `wails dev`. Navigated a fresh page and sampled `.ws-root`'s `data-rail-side` attribute 40 times at 50ms intervals starting immediately after `page.goto` -- every sample read `"Left"`, zero occurrences of `"Right"`, confirming no visible flash of the wrong side on startup.
9. Confirmed fresh bindings post-relaunch (`SetRailSide` present) and live config `railSide: "Left"`.
10. Restored the original `railSide: "Right"` via the real `SetRailSide` binding (not a file edit), confirmed on disk.
11. Cleanup: killed `wails dev`/`StorCat`/`vite` (confirmed via `ps`/`lsof` on both ports), reverted `frontend/wailsjs/runtime/{package.json,runtime.d.ts,runtime.js}` file-mode noise via `git checkout --`, confirmed `git status --short` clean before proceeding.

## User Setup Required

None -- no external service configuration required. (CRED-04/CRED-05 -- Windows eSigner signing credentials -- remain an existing, previously-recorded deferred item; this plan did not attempt to provision them, per the plan's own scope.)

## Next Phase Readiness

- **This is the last plan of the last phase of v3.0.0.** `local main` is now published to `origin/main`; a release-please PR (#3, `chore(main): release 2.4.0`) is open and awaiting the user's own decision on when/whether to cut a release tag.
- **COMPAT-06 is genuinely partial, not complete.** Cutting a `v*.*.*` tag and observing the resulting `release.yml` run (macOS sign+notarize; Windows signing will SKIP with a warning until CRED-04/CRED-05 are provisioned) is the only way to close the remaining half of COMPAT-06 -- that is a deliberate product decision outside this plan's scope, not an oversight.
- **Ten WINDOWS.md entries remain open**, all requiring real Windows or Linux hardware this project has never had, or are accepted upstream-owned residuals (entry #12) or an accepted judgment-call candidate for reframing (entry #8, left to the user per the plan's own instruction not to reframe it unilaterally).
- **GSD worktrees can now fork from a current `origin/main`** (`workflow.use_worktrees` was deliberately left `false` all milestone specifically because local `main` was far ahead of `origin/main`; that constraint is now resolved by this plan's push, though STATE.md's guidance note about *why* it was off should stay for historical context unless the user wants it trimmed).

---
*Phase: 28-re-scan-diff*
*Completed: 2026-08-16*

## Self-Check: PASSED

Files verified present: `.planning/WINDOWS.md`, `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, this SUMMARY -- all present on disk. Commit `96282130` verified present in `git log --oneline -3`. `gsd-tools windows status` confirms ledger parses cleanly (open:10, fixed:4, waived:0, total:14). `grep -c '| open |' .planning/WINDOWS.md` returns 10. Working tree clean except for the four planning files this plan intentionally modifies (STATE.md/ROADMAP.md updates follow in the plan-completion commit). No process listening on :34115 or :5173. `frontend/wailsjs/runtime/*` file-mode noise reverted.
