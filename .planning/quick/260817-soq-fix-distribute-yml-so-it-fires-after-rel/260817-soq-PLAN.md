---
phase: quick-260817-soq
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - .github/workflows/distribute.yml
autonomous: true
requirements: []
estimate:
  tokens: 24000
  raw_tokens: 12000
  tasks: 2
  confidence: low
must_haves:
  truths:
    - "Distribute fires automatically when the Release workflow completes, without any manual dispatch"
    - "A failed Release run never publishes to the Homebrew tap or commits WinGet manifests"
    - "TAG resolves to the release tag on all three trigger paths: workflow_dispatch, release, and workflow_run"
    - "The existing release/published and workflow_dispatch triggers keep working exactly as they do today"
    - "A future reader can tell from the file itself WHY workflow_run exists alongside the release trigger"
  artifacts:
    - path: ".github/workflows/distribute.yml"
      provides: "workflow_run trigger on the Release workflow, conclusion guards on every job, and a three-way TAG fallback"
      contains: "workflow_run"
  key_links:
    - from: ".github/workflows/release.yml (name: Release)"
      to: ".github/workflows/distribute.yml"
      via: "on.workflow_run.workflows: [\"Release\"] matched by workflow name string"
      pattern: "Release run completes -> Distribute is queued -> conclusion guard admits only successful runs"
    - from: "github.event.workflow_run.head_branch"
      to: "TAG shell variable in update-homebrew and update-winget-manifests"
      via: "GitHub Actions || fallback chain after inputs.tag and github.event.release.tag_name"
      pattern: "head_branch carries the tag name when the upstream run was triggered by a tag push"
---

<objective>
Make `.github/workflows/distribute.yml` fire automatically after `release.yml` finishes, by adding a `workflow_run` trigger instead of depending on the `release: published` event that GitHub never emits for our releases.

Purpose: v3.0.0 shipped on 2026-08-18 but the Homebrew tap stayed on 2.3.0 and no 3.0.0 WinGet manifests were generated, because `release.yml` publishes with the default `GITHUB_TOKEN` and GitHub suppresses workflow-triggering events for `GITHUB_TOKEN` actions. Distribute has not run automatically once — its only successful run (v2.3.0, 2026-03-28) was manually dispatched. Every future release repeats this failure until the trigger is fixed.

Output: A single modified workflow file with a `workflow_run` trigger, per-job conclusion guards, extended TAG resolution, and an in-file comment recording why the apparently-redundant trigger must not be removed.
</objective>

<execution_context>
@$HOME/.claude/gsd-core/workflows/execute-plan.md
@$HOME/.claude/gsd-core/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md

@.github/workflows/distribute.yml
@.github/workflows/release.yml
</context>

<background>
Confirmed root cause — do NOT re-investigate, this is settled:

1. Ken pushes tag `vX.Y.Z`. `release.yml` (workflow `name: Release`, confirmed at line 1) fires on `push: tags: ['v*.*.*']`.
2. `release.yml` line 385-390, step "Upload artifacts and publish release", uses `softprops/action-gh-release@v2.6.1` with **no `token:` input**, so it authenticates with the default `GITHUB_TOKEN`.
3. GitHub does not emit workflow-triggering events for actions taken with `GITHUB_TOKEN`. The `release: published` event is therefore never delivered.
4. `distribute.yml` triggers only on `release: [published]` and `workflow_dispatch`, so it never runs on its own.

Why `workflow_run` and not `push: tags`: a `push: tags` trigger would start Distribute at the same moment as `release.yml`, and the `curl` downloads of `StorCat-${TAG}-darwin-universal.dmg` and `...-windows-amd64-installer.exe` would 404 against a release whose assets are still building. `workflow_run` fires only after `release.yml` completes, so the assets exist. `workflow_run` is also not subject to the `GITHUB_TOKEN` event suppression.

**Reference facts about the current file** (read it, do not assume):
- `update-homebrew` computes TAG once, at line 24.
- `update-winget-manifests` computes TAG **twice** — once at line 90 ("Generate WinGet manifests for new version") and again at line 120 ("Commit manifests to main repo"). Both need the fallback.
- `update-winget` computes no TAG at all; it delegates version inference to the `vedantmgoyal9/winget-releaser` action.

**Already handled, out of scope:** the v3.0.0 backfill ran manually — commit `db834695` "chore: add WinGet manifests for 3.0.0" by `github-actions[bot]` on 2026-08-18 01:40 UTC. Do not dispatch workflows, push to the tap, or take any other outward-facing action.
</background>

<constraints>
- Scope is EXACTLY `.github/workflows/distribute.yml`. Do NOT touch `release.yml`, the release process, `packaging/homebrew/storcat.rb.template`, or anything under `packaging/winget/`.
- Keep the existing `release:` and `workflow_dispatch:` triggers exactly as they are. This change is additive.
- There is no unit test for a GitHub Actions workflow, and you must not invent a test harness. Verification is a YAML parse check plus structural assertions, both specified below.
- `actionlint` is NOT installed on this machine. PyYAML IS available via system `python3`. Do not install anything.
- **PyYAML gotcha you will hit:** under YAML 1.1, the bare key `on` parses as the boolean `True`, not the string `'on'`. `d['on']` raises `KeyError`. Use `d.get('on', d.get(True))`. This is a parser quirk, not a defect in the file — do not "fix" the file in response to it.
</constraints>

<tasks>

<task type="tracer">
  <name>Task 1: Wire the workflow_run trigger end-to-end through update-homebrew</name>
  <files>.github/workflows/distribute.yml</files>
  <action>
Prove the full path — trigger, guard, tag resolution, publish — on the one job that has historically succeeded (`update-homebrew`), before touching the other two jobs.

Three edits, all in `.github/workflows/distribute.yml`:

**(a) Add the `workflow_run` trigger to the `on:` block**, placed above the existing `release:` key. Keep `release:` and `workflow_dispatch:` byte-identical to what is there now. The new trigger is:

  workflows list containing the single string `Release` (this must match `release.yml`'s `name:` field exactly — confirmed as `Release`), and a types list containing the single value `completed`.

Use `types: [completed]`, not `[success]` — `workflow_run` has no `success` type; conclusion is filtered by the job guard in (b).

**(b) Add a WHY comment immediately above the `workflow_run` key.** This is a Chesterton's Fence the file must document itself: the trigger looks redundant next to `release: published`, and a future reader will delete it without this note. The comment must state, in prose:
  - that release/published does not fire for this repo's releases,
  - that the cause is release.yml publishing with the default token, and that GitHub suppresses workflow-triggering events for such actions,
  - that the trigger below is therefore NOT redundant and must not be removed,
  - that it additionally guarantees release assets have finished uploading before the download steps run,
  - the accepted known limitation: when release.yml is started through its own manual dispatch rather than a tag push, `head_branch` holds the branch name (main) rather than a tag, and this workflow's own manual dispatch input is the escape hatch for that case.

Write it as normal YAML `#` comment lines indented two spaces to sit inside the `on:` block. Do not paraphrase this into one terse line — the whole point is that the next reader needs the reasoning, not a label.

**(c) Guard the `update-homebrew` job and extend its TAG resolution.**
  - Add a job-level `if:` key to `update-homebrew`, as a sibling of `runs-on:` and `steps:` (job level, NOT on any individual step). Its expression admits the job when `github.event_name` is not equal to the string `workflow_run`, OR when `github.event.workflow_run.conclusion` equals the string `success`. That shape leaves the release and manual-dispatch paths completely unaffected while blocking publication after a failed Release run.
  - In the "Get version and compute SHA256" step (currently line 24), extend the TAG assignment with a third fallback operand, `github.event.workflow_run.head_branch`, appended after the existing `inputs.tag` and `github.event.release.tag_name` operands using the same `||` chaining already present. Order matters: manual input wins, then the release payload, then the upstream run's head ref. Leave `VERSION_CLEAN`, `DMG_URL`, and every other line in that step untouched.
  </action>
  <verify>
    <automated>python3 -c "
import yaml
d = yaml.safe_load(open('.github/workflows/distribute.yml'))
on = d.get('on', d.get(True))
wr = on['workflow_run']
assert wr['workflows'] == ['Release'], wr
assert wr['types'] == ['completed'], wr
assert on['release']['types'] == ['published'], on['release']
assert on['workflow_dispatch']['inputs']['tag']['required'] is True, on['workflow_dispatch']
g = str(d['jobs']['update-homebrew'].get('if', ''))
assert 'workflow_run.conclusion' in g and 'success' in g, g
print('OK trigger+guard')
" && test $(grep -v '^[[:space:]]*#' .github/workflows/distribute.yml | grep -c 'github.event.workflow_run.head_branch') -eq 1</automated>
  </verify>
  <done>`on:` carries a `workflow_run` trigger on the `Release` workflow with `types: [completed]`; the pre-existing `release` and `workflow_dispatch` triggers parse unchanged; `update-homebrew` has a job-level conclusion guard; exactly one non-comment line references `head_branch`; a multi-line comment above the trigger explains the token-suppression cause and the manual-dispatch limitation.</done>
</task>

<task type="auto">
  <name>Task 2: Extend guards and TAG resolution to the two WinGet jobs</name>
  <files>.github/workflows/distribute.yml</files>
  <action>
Apply the same treatment established in Task 1 to `update-winget` and `update-winget-manifests`.

**(a) `update-winget-manifests`** — add the identical job-level `if:` guard used in Task 1, as a sibling of `runs-on:`/`permissions:`/`steps:`. Then extend the TAG assignment with the `github.event.workflow_run.head_branch` fallback in **both** places this job computes it: the "Generate WinGet manifests for new version" step (currently line 90) and the "Commit manifests to main repo" step (currently line 120). Missing the second one would generate correct manifests and then commit them under a path built from an empty version string, so check both. Change nothing else in either step — the 2.1.0 template-baseline `sed` logic and its explanatory comment stay exactly as written.

**(b) `update-winget`** — add the identical job-level `if:` guard. Do NOT add any other conditional and do NOT modify its steps. Under `workflow_run` this job runs its "Check if package exists in winget-pkgs" step, emits its existing warning, sets `exists=false`, and skips the submission step — behaviorally identical to what it does on every other trigger today, so the guard alone preserves its behavior on all paths.

**(c)** Add a short comment above `update-winget`'s `steps:` recording the known limitation: the `vedantmgoyal9/winget-releaser` action infers its version from the release event payload and cannot resolve a version under `workflow_run`. Note that this is currently inert because `scottkw.StorCat` is not yet in the winget-pkgs registry, so the submission step never executes, and that the limitation becomes live only if and when the package is accepted upstream. Do NOT engineer around this — no extra step conditions, no version plumbing into the action. The comment is the deliverable.
  </action>
  <verify>
    <automated>python3 -c "
import yaml
d = yaml.safe_load(open('.github/workflows/distribute.yml'))
jobs = d['jobs']
assert set(jobs) == {'update-homebrew','update-winget','update-winget-manifests'}, set(jobs)
bad = [n for n,j in jobs.items() if 'workflow_run.conclusion' not in str(j.get('if',''))]
assert not bad, 'unguarded jobs: %s' % bad
for n,j in jobs.items():
    for s in j['steps']:
        assert 'workflow_run.conclusion' not in str(s.get('if','')), (n, s.get('name'))
print('OK all jobs guarded at job level')
" && test $(grep -v '^[[:space:]]*#' .github/workflows/distribute.yml | grep -c 'github.event.workflow_run.head_branch') -eq 3 && test $(grep -v '^[[:space:]]*#' .github/workflows/distribute.yml | grep -c 'inputs.tag || github.event.release.tag_name') -eq 3</automated>
  </verify>
  <done>All three jobs carry the conclusion guard at job level (never at step level); exactly three non-comment lines carry the full three-operand TAG fallback chain; the `update-winget` limitation is documented in-file; no step logic outside the TAG assignments changed.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| upstream workflow run -> Distribute | `workflow_run` grants Distribute a read/write `GITHUB_TOKEN` plus `HOMEBREW_TAP_TOKEN` in the base-repo context, keyed off another workflow's completion |
| GitHub Releases CDN -> runner | `curl` downloads a DMG and an .exe whose URL is built from a tag string sourced from event payload |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-soq-01 | Elevation of Privilege | `on.workflow_run` trigger | medium | accept | `workflow_run` is the documented "pwn request" vector when the upstream workflow is itself triggered by untrusted input. Not applicable here: `release.yml` triggers only on `push: tags` and `workflow_dispatch`, neither of which any non-maintainer can reach on this repo. No `pull_request` or `pull_request_target` trigger exists upstream. Accepted, with the constraint that adding a PR trigger to `release.yml` would invalidate this analysis. |
| T-soq-02 | Tampering | tap push / manifest commit after a failed Release | high | mitigate | Job-level `if: github.event_name != 'workflow_run' \|\| github.event.workflow_run.conclusion == 'success'` on all three jobs (Task 1c, Task 2a, Task 2b). A cancelled, failed, or skipped Release run cannot reach the `HOMEBREW_TAP_TOKEN` push or the manifest commit. Verified structurally by the Task 2 parse assertion over every job. |
| T-soq-03 | Tampering | `TAG` interpolated into `DMG_URL` / `INSTALLER_URL` / `MANIFEST_DIR` | low | accept | `head_branch` is a git ref name on this repo, settable only by someone who can already push a tag here (i.e. a maintainer, who can already run the workflow directly). Pre-existing exposure via `inputs.tag`, unchanged in kind by this plan. `curl -fsSL` fails closed on a bad URL rather than publishing a wrong artifact. |
| T-soq-04 | Spoofing | `workflows: ["Release"]` name match | low | accept | The trigger binds by workflow display name, not path, so renaming `release.yml`'s `name:` silently breaks this wiring. Only repo writers can rename it. Recorded here so the coupling is discoverable; the in-file comment from Task 1b is the operational mitigation. |

No package-manager installs are introduced by this plan, so no package-legitimacy gate applies.
</threat_model>

<verification>
Whole-file checks after both tasks:

1. **Parses as valid YAML with the expected shape:**
   `python3 -c "import yaml; d=yaml.safe_load(open('.github/workflows/distribute.yml')); on=d.get('on',d.get(True)); print(sorted(on)); print(sorted(d['jobs']))"`
   Expect the trigger keys `['release','workflow_dispatch','workflow_run']` and the job names `['update-homebrew','update-winget','update-winget-manifests']`.

2. **Read the three TAG lines aloud and confirm operand order** is manual input, then release payload, then `workflow_run.head_branch` — a wrong order would let an empty payload field win over a real manual input:
   `grep -n 'TAG=' .github/workflows/distribute.yml`

3. **Confirm the diff is additive only.** `git diff .github/workflows/distribute.yml` must show no removed lines except the three TAG lines being replaced by their extended versions. Nothing under `packaging/` and no other workflow file may appear in `git status`.

4. **Confirm the guard sits at job level, not step level** — a step-level guard would let the checkout and download steps run after a failed Release. Covered by the Task 2 automated assertion; re-read the diff to confirm visually.
</verification>

<success_criteria>
- `.github/workflows/distribute.yml` parses as valid YAML and is the only modified file.
- The `on:` block contains all three triggers, with `workflow_run` bound to workflow name `Release` and `types: [completed]`.
- All three jobs carry the conclusion guard at job level.
- All three TAG assignments carry the three-operand fallback chain.
- The `workflow_run` trigger is preceded by a comment explaining the token-suppression cause, the asset-race benefit, and the manual-dispatch `head_branch` limitation.
- `update-winget` carries a comment recording that `winget-releaser` cannot resolve a version under `workflow_run` and that this is currently inert.
- No workflow was dispatched, no commit was pushed to the tap repo, and no outward-facing action was taken.
</success_criteria>

<known_limitations>
Record these verbatim in the SUMMARY as accepted, not as open bugs:

1. **`release.yml` run via its own `workflow_dispatch`.** `github.event.workflow_run.head_branch` then holds the branch name (`main`), not a tag, so TAG resolves to `main`, the DMG URL 404s, and `update-homebrew` fails loudly rather than publishing something wrong. Distribute's own `workflow_dispatch` with an explicit `tag` input remains the manual escape hatch. Accepted — not engineered around.

2. **`update-winget` cannot function under `workflow_run`.** `vedantmgoyal9/winget-releaser` infers its version from the release event payload, which is absent on this trigger. Currently inert: `scottkw.StorCat` is not in the winget-pkgs registry, so the check step short-circuits and the submission step never runs on any trigger. This becomes a real gap only once the package is accepted upstream, at which point the job needs its own follow-up. Not in scope here.

3. **The `workflows: ["Release"]` binding is by display name.** Renaming `release.yml`'s `name:` field silently disconnects Distribute. There is no way to bind by file path in the `workflow_run` trigger.
</known_limitations>

<output>
Create `.planning/quick/260817-soq-fix-distribute-yml-so-it-fires-after-rel/260817-soq-SUMMARY.md` when done.
</output>
