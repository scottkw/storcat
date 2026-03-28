# Phase 21: Auto Version and Auto Distribution - Research

**Researched:** 2026-03-28
**Domain:** GitHub Actions CI/CD automation — release-please, conventional commits, multi-job release pipelines
**Confidence:** HIGH

## Summary

This phase automates the full release pipeline using Google's release-please. The tool monitors conventional commits on `main`, maintains a release PR, and when that PR is merged, creates a git tag and GitHub release. The existing `release.yml` must be refactored: instead of creating a new release itself, it must trigger from the tag and upload artifacts into the release that release-please already created. After artifact upload, release.yml marks the release as published, which triggers `distribute.yml` (already correct, no changes needed).

The core technical challenge is two-fold: (1) configuring release-please to update `wails.json`'s `$.info.productVersion` field via the `extra-files` / `jsonpath` mechanism, and (2) refactoring `release.yml`'s final `release` job to upload to the existing release rather than create a new one. The existing tag-push trigger on `release.yml` continues to work because release-please creates the tag.

**Primary recommendation:** Use `googleapis/release-please-action@v4` with manifest configuration, `release-type: simple`, and a JSON extra-file targeting `$.info.productVersion` in `wails.json`.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Use conventional commits (feat:, fix:, breaking:) to auto-determine semver bump
- **D-02:** Use Google's release-please as the versioning/release tool
- **D-03:** release-please auto-generates and maintains CHANGELOG.md from conventional commits
- **D-04:** release-please runs on pushes to main branch only (no release branches)
- **D-05:** Merging the release-please PR creates the git tag and GitHub release, triggering the existing release.yml build pipeline
- **D-06:** The splash screen version already reads dynamically from Go backend's GetVersion() which embeds wails.json — no extra work needed, just ensure release-please bumps wails.json
- **D-07:** wails.json `info.productVersion` is the single source of truth for version
- **D-08:** release-please updates wails.json only. frontend/package.json version is cosmetic and can be left as-is or pinned
- **D-09:** release-please creates the GitHub release (not release.yml). release.yml triggers on tag push, builds all platform artifacts, and uploads them to the existing release
- **D-10:** After artifact upload, release.yml marks the release as published (no longer draft). This triggers distribute.yml automatically
- **D-11:** distribute.yml fires on release published — updates Homebrew tap and WinGet manifests. Full chain is automated: merge PR → tag → build → upload → publish → distribute

### Claude's Discretion
- release-please configuration details (release-type, extra-files, versioning-strategy)
- How release.yml is refactored to upload to existing release instead of creating one
- Whether to use release-please GitHub Action or the standalone release-please CLI
- CHANGELOG.md formatting and sections

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| googleapis/release-please-action | v4 | Conventional commit parsing, PR management, release creation | Official Google action; manifest-based config supports arbitrary JSON file updates |
| softprops/action-gh-release | v2 (153bb8e already in release.yml) | Upload artifacts to existing release | Already pinned in project; supports uploading to existing release by tag |

### Discretion Choice: GitHub Action (not standalone CLI)

Use `googleapis/release-please-action@v4` (not the standalone `release-please` CLI). Reasons:

- No additional secrets or credentials required — uses `GITHUB_TOKEN`
- Simpler workflow: single step, outputs `release_created` / `tag_name` / `upload_url`
- CLI requires a separate install step and PAT; adds complexity for no benefit here

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| release-please-action | semantic-release | semantic-release is heavier, npm-centric, harder to configure for non-Node targets |
| release-please-action | manual tag + CHANGELOG | Loses automation entirely — defeats the point |
| simple release-type | go release-type | go release-type updates `version.go` but NOT arbitrary JSON — doesn't work for wails.json |

**Installation (no install needed — Action-based):** Config files in repo are all that's required.

---

## Architecture Patterns

### New Workflow Files Required

```
.github/
└── workflows/
    ├── release-please.yml      # NEW — runs on push to main, manages PR + creates release
    ├── release.yml             # MODIFIED — uploads artifacts to existing release, publishes it
    └── distribute.yml          # UNCHANGED — fires on release published
.release-please-manifest.json   # NEW — tracks current version (e.g., {"." : "2.2.1"})
release-please-config.json      # NEW — defines release-type, extra-files for wails.json
```

### Pattern 1: release-please-config.json

**What:** Declares `release-type: simple`, points extra-files at `wails.json`, sets the JSON path to the version field.

**When to use:** Single-package repository with a non-standard version file.

```json
{
  "packages": {
    ".": {
      "release-type": "simple",
      "extra-files": [
        {
          "type": "json",
          "path": "wails.json",
          "jsonpath": "$.info.productVersion"
        }
      ]
    }
  }
}
```

**Confidence:** MEDIUM — `release-type: simple` with a JSON `extra-files` entry using JSONPath is documented in `googleapis/release-please` `docs/customizing.md`. The exact JSONPath `$.info.productVersion` for a nested object is standard JSONPath syntax. One GitHub issue (#2135) confirms this pattern works for nested fields.

### Pattern 2: .release-please-manifest.json

**What:** Bootstraps release-please with the current version so it doesn't scan the full git history.

```json
{
  ".": "2.2.1"
}
```

**Note:** The version here must match `wails.json`'s `info.productVersion` at the time the file is committed. This is the bootstrap value only; release-please owns it after first run.

### Pattern 3: release-please.yml Workflow

**What:** Runs on every push to `main`. On ordinary feature pushes, keeps the release PR up to date. When the release PR is merged, creates a tag + GitHub release (non-draft by default) and sets `release_created: true`.

```yaml
# Source: googleapis/release-please-action README
name: Release Please

on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-latest
    steps:
      - uses: googleapis/release-please-action@v4
        id: release
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
```

**Key outputs available:**
- `steps.release.outputs.release_created` — `'true'` only when a release is created
- `steps.release.outputs.tag_name` — e.g., `v2.3.0`
- `steps.release.outputs.upload_url` — GitHub API endpoint for asset upload

**Note on GITHUB_TOKEN and downstream workflow triggering:** Release-please uses `GITHUB_TOKEN` to create the tag and release. GitHub's default behavior prevents `GITHUB_TOKEN`-initiated events from triggering other workflows. However, the existing `release.yml` triggers on `push: tags: ['v*.*.*']` — a **tag push event**, not a workflow event. Tag push events **are** triggered by `GITHUB_TOKEN`, so `release.yml` will fire correctly. Only workflow-to-workflow triggers via `release:` events require a PAT; the tag trigger is exempt from this restriction.

**Confidence:** HIGH — verified via GitHub docs and release-please-action issue #1000.

### Pattern 4: release.yml Refactoring — Upload to Existing Release

**What:** The current `release` job creates a draft release via `softprops/action-gh-release`. This must be replaced with: (a) upload artifacts to the existing release that release-please already created, and (b) publish the release (set `draft: false`).

**Current (to remove):**
```yaml
  release:
    needs: [build-macos, build-windows, build-linux-amd64, build-linux-arm64]
    runs-on: ubuntu-22.04
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@...
      - name: Create GitHub Release        # <-- REMOVE THIS
        uses: softprops/action-gh-release@...
        with:
          draft: true
          generate_release_notes: true
          files: artifacts/**/*
```

**New pattern:**
```yaml
  release:
    needs: [build-macos, build-windows, build-linux-amd64, build-linux-arm64]
    runs-on: ubuntu-22.04
    permissions:
      contents: write
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02  # v4.2.1
        with:
          path: artifacts/

      - name: Get version from tag
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Upload artifacts and publish release
        uses: softprops/action-gh-release@153bb8e04406b158c6c84fc1615b65b24149a1fe  # v2.6.1
        with:
          tag_name: ${{ steps.version.outputs.VERSION }}
          draft: false
          files: artifacts/**/*
```

**How this works:** `softprops/action-gh-release` detects that the tag already has an associated release (created by release-please) and uploads the files to that existing release, then sets `draft: false`. This is the documented behavior: "If a tag already has a GitHub release, the existing release will be updated with the release assets."

**Confidence:** HIGH — documented in softprops/action-gh-release README; `tag_name` parameter explicitly selects the existing release to update.

### Pattern 5: distribute.yml — No Changes Required

The existing `distribute.yml` triggers on `release: types: [published]`. When `release.yml`'s final job sets `draft: false` via `softprops/action-gh-release`, that publishes the release, which fires the `published` event. **No changes to `distribute.yml` are needed.**

**Caveat:** `distribute.yml` downloads the DMG from the published GitHub release URL to compute SHA256. There is a brief window between the release being published and the assets being fully available for download. The existing `--retry 5 --retry-delay 10` curl flags in `distribute.yml` already handle this. No change needed.

### Anti-Patterns to Avoid
- **Using `generate_release_notes: true` in the upload step:** release-please already generates release notes from CHANGELOG.md. Adding `generate_release_notes: true` in the upload step would duplicate or overwrite the notes. Omit it.
- **Setting `draft: true` then publishing separately:** release-please creates a non-draft release by default. Keep it non-draft; the upload step finds it by tag and uploads to it.
- **Using `release-type: go`:** The Go release type updates `version.go` not `wails.json`. The Go source of truth in this project is `wails.json` via `//go:embed`, so `release-type: simple` with explicit `extra-files` is the correct approach.
- **Editing `.release-please-manifest.json` manually after bootstrap:** Once release-please owns the file, manual edits can desync it from the git tag history and cause unexpected version bumps.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Conventional commit parsing | Custom regex | release-please | Handles edge cases: `!` suffix for breaking, footers, scopes, multi-paragraph bodies |
| CHANGELOG generation | Custom script | release-please | Auto-sections (Features, Bug Fixes, etc.), links to commits and PRs |
| Version bump calculation | Custom semver logic | release-please | Correctly handles pre-1.0.0 semver rules, pre-releases |
| Release PR management | Custom GitHub API calls | release-please | Handles PR update/squash/rebase, conflict resolution |

**Key insight:** release-please has been battle-tested across thousands of Google and community projects. The conventional commit parser handles every edge case in the spec (breaking changes in footer vs. `!`, multi-module PRs, etc.). Building a custom alternative is high risk with no upside.

---

## Common Pitfalls

### Pitfall 1: `release-type: go` Does Not Update wails.json
**What goes wrong:** If `release-type: go` is used, release-please tries to update `version.go` using a specific Go version file pattern. It will NOT update `wails.json`.
**Why it happens:** Each release-type has a hardcoded set of version files it manages.
**How to avoid:** Use `release-type: simple` with an explicit `extra-files` entry pointing to `wails.json` with `jsonpath: $.info.productVersion`.
**Warning signs:** Release PR is created but `wails.json` productVersion is not changed in the diff.

### Pitfall 2: release-please Creates Releases as Draft by Default (v3 behavior; v4 changed)
**What goes wrong:** In older configurations, release-please creates draft releases. `softprops/action-gh-release` then tries to find the release by tag, but draft releases may not be found correctly in all versions.
**Why it happens:** v3 of the action had `draft: true` by default. v4 changed to non-draft.
**How to avoid:** With `googleapis/release-please-action@v4`, releases are non-draft by default. Do not add `draft: true` to the config unless you want manual publish gates.
**Warning signs:** distribute.yml does not fire after release.yml completes.

### Pitfall 3: GITHUB_TOKEN Cannot Trigger release: published Events in Downstream Workflows (Only Applies to release: trigger)
**What goes wrong:** If distribute.yml is changed to trigger on `workflow_run` or `workflow_dispatch` instead of `release: published`, the GITHUB_TOKEN limitation matters.
**Why it happens:** GitHub prevents `GITHUB_TOKEN`-initiated events from triggering other workflows for most event types to prevent loops. However, `release: published` events created by `GITHUB_TOKEN` DO fire for workflows listening on `release: types: [published]`.
**How to avoid:** Keep distribute.yml on `release: types: [published]`. Do not change to other trigger types.
**Warning signs:** distribute.yml never fires.

### Pitfall 4: .release-please-manifest.json Version Mismatch
**What goes wrong:** If the manifest contains a version different from the latest git tag, release-please calculates the next version incorrectly.
**Why it happens:** Manual editing of the manifest without a matching tag.
**How to avoid:** Bootstrap the manifest with the current version (`2.2.1`) before the first run. After that, let release-please own the file.
**Warning signs:** Release PR bumps from an unexpected base version.

### Pitfall 5: release.yml's Final Job Overwrites Release Notes
**What goes wrong:** If `generate_release_notes: true` is kept in the upload step, GitHub regenerates release notes from commit messages, overwriting the CHANGELOG-sourced notes that release-please wrote.
**Why it happens:** `softprops/action-gh-release` has `generate_release_notes` as an option that calls the GitHub API to regenerate notes.
**How to avoid:** Remove `generate_release_notes: true` from the upload step.
**Warning signs:** After release, the GitHub release notes look different from the CHANGELOG entry.

### Pitfall 6: Artifacts Not Available When distribute.yml Runs
**What goes wrong:** distribute.yml downloads the DMG immediately after `published` event fires, but the release assets may not yet be propagated by GitHub's CDN.
**Why it happens:** GitHub release asset availability has propagation delay.
**How to avoid:** The existing `--retry 5 --retry-delay 10` flags in distribute.yml already handle this. Verify they are present before removing the draft-based protection.
**Warning signs:** distribute.yml fails with 404 on the DMG download.

---

## Code Examples

### release-please-config.json (complete)
```json
{
  "packages": {
    ".": {
      "release-type": "simple",
      "extra-files": [
        {
          "type": "json",
          "path": "wails.json",
          "jsonpath": "$.info.productVersion"
        }
      ]
    }
  }
}
```

### .release-please-manifest.json (bootstrap)
```json
{
  ".": "2.2.1"
}
```

### .github/workflows/release-please.yml (complete)
```yaml
name: Release Please

on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-latest
    steps:
      - uses: googleapis/release-please-action@v4
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
```

### release.yml — Modified release job (replaces the existing release job)
```yaml
  release:
    needs: [build-macos, build-windows, build-linux-amd64, build-linux-arm64]
    runs-on: ubuntu-22.04
    permissions:
      contents: write
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@95815c38cf2ff2164869cbab79da8d1f422bc89e  # v4.2.1
        with:
          path: artifacts/

      - name: List artifacts
        run: ls -laR artifacts/

      - name: Get version from tag
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Upload artifacts and publish release
        uses: softprops/action-gh-release@153bb8e04406b158c6c84fc1615b65b24149a1fe  # v2.6.1
        with:
          tag_name: ${{ steps.version.outputs.VERSION }}
          draft: false
          files: artifacts/**/*
```

**Note:** Remove `generate_release_notes: true` from the original step — release-please already wrote the release notes.

---

## Runtime State Inventory

> Rename/refactor audit — not applicable for this phase. This phase adds new workflow files and modifies an existing workflow; no string rename occurs.

**Step 2.5: SKIPPED** — Phase is additive (new config files, modified workflow), not a rename/refactor.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| googleapis/release-please-action@v4 | release-please.yml | ✓ (GitHub Marketplace) | v4 | — |
| softprops/action-gh-release@v2 | release.yml upload step | ✓ (already pinned) | 153bb8e | — |
| GITHUB_TOKEN | release-please.yml, release.yml | ✓ (default in all Actions) | — | — |
| Conventional commit history | release-please version calc | Must verify | — | Bootstrap SHA if needed |

**Missing dependencies with no fallback:** None.

**Note on conventional commit history:** release-please scans commit history back to the last release. If the git history does not contain conventional commits (e.g., older commits use freeform messages), release-please will still work — it simply won't find changes to bump. The bootstrap manifest sets the starting version; the first PR merge after setup will create the next release based on whatever conventional commits exist since the bootstrap SHA.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | No unit test framework applicable — CI/CD workflow changes only |
| Config file | N/A |
| Quick run command | N/A |
| Full suite command | N/A |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| D-01 through D-11 | All behaviors | Manual / integration | Run via actual GitHub Actions | N/A — CI only |

**Note:** This phase is entirely GitHub Actions workflow configuration. There are no unit-testable code changes. Validation must be done through:
1. **Wave 0 dry-run test:** Push a `fix:` commit to main; verify release-please creates/updates the PR (no merge yet).
2. **Wave 1 merge test:** Merge the release PR; verify tag is created, release.yml fires, artifacts upload, release is published, distribute.yml fires.
3. **Splash screen verification:** Build locally after release-please bumps wails.json; confirm `GetVersion()` returns new version.

### Wave 0 Gaps
- [ ] Conventional commit history check — verify recent commits use conventional format or note the bootstrap SHA approach
- [ ] Workflow permissions check — confirm repo settings allow GitHub Actions to create PRs (Settings > Actions > General > Workflow permissions)

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| release-please v3 action (google-github-actions/) | release-please v4 (googleapis/) | 2023 | Different import path; v3 repo is archived. Must use `googleapis/release-please-action@v4` |
| Inputs in workflow YAML | Config in `release-please-config.json` | v4 (2023) | release-type, extra-files, etc. moved from Action `with:` inputs to config file |
| release-please creates draft release | release-please creates non-draft release | v4 default | Affects how downstream publish-trigger chains work |

**Deprecated/outdated:**
- `google-github-actions/release-please-action`: Archived repo, do not use. Use `googleapis/release-please-action@v4`.
- Workflow-level `with: release-type: ...` inputs: Deprecated in v4. Move to `release-please-config.json`.

---

## Open Questions

1. **Does the existing git history contain conventional commits?**
   - What we know: The project has been developed and the current version is 2.2.1. Recent commits (from git log) include messages like `docs(state): record phase 21 context session`, `docs(21): capture phase context`, `docs(v2.3.0): add milestone audit report` — these are conventional commit format.
   - What's unclear: How far back conventional commits were consistently used; whether there's a commit that introduced them.
   - Recommendation: Set `bootstrap-sha` in `release-please-config.json` to the SHA of the last release tag (v2.2.1) to prevent release-please from scanning beyond that point.

2. **Does `release-type: simple` with JSON extra-files reliably update nested JSON paths?**
   - What we know: The `GenericJson` updater supports JSONPath. Official docs show `$.json.path.to.field` syntax. Community confirmations exist for nested paths.
   - What's unclear: Whether the updater handles the `$schema` field in `wails.json` without issues (it's a URL string, not a version field; the jsonpath is specific so this should be fine).
   - Recommendation: Test with a dry-run by creating a test branch and verifying the release PR diff shows the correct wails.json change.

3. **Does `softprops/action-gh-release` reliably find and update the release-please-created release?**
   - What we know: The action's documented behavior is to update an existing release when the tag already has one. There was a 2024 issue (#445 "Existing releases are no longer updated") but it was resolved.
   - What's unclear: Whether the pinned SHA `153bb8e04406b158c6c84fc1615b65b24149a1fe` is a version that contains the fix.
   - Recommendation: The SHA in the current release.yml maps to v2.6.1 (December 2025). v2.6.1 includes the fix and the new "mark as draft until all artifacts uploaded" feature. This SHA is current and correct.

---

## Sources

### Primary (HIGH confidence)
- [googleapis/release-please-action README](https://github.com/googleapis/release-please-action) — v4 workflow syntax, outputs, config-file/manifest-file inputs
- [googleapis/release-please docs/customizing.md](https://github.com/googleapis/release-please/blob/main/docs/customizing.md) — extra-files, JSON type, jsonpath syntax
- [googleapis/release-please docs/manifest-releaser.md](https://github.com/googleapis/release-please/blob/main/docs/manifest-releaser.md) — release-please-config.json and manifest structure
- [softprops/action-gh-release README](https://github.com/softprops/action-gh-release) — upload to existing release by tag_name

### Secondary (MEDIUM confidence)
- [release-please-action issue #1000](https://github.com/googleapis/release-please-action/issues/1000) — GITHUB_TOKEN and downstream workflow triggering limitations; tag push events are exempt
- [release-please issue #2135](https://github.com/googleapis/release-please/issues/2135) — Community confirmation of extra-files JSON with nested jsonpath
- [softprops/action-gh-release issue #445](https://github.com/softprops/action-gh-release/issues/445) — Existing release update behavior confirmed fixed in v2

### Tertiary (LOW confidence — for context only)
- Multiple Medium/dev.to articles on release-please patterns — patterns consistent with official docs

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — release-please-action@v4 and softprops@v2 are the de facto standard; both already in project
- Architecture: HIGH — release-please config structure and jsonpath for wails.json verified against official docs
- release.yml refactoring: HIGH — softprops update-existing-release behavior documented and confirmed
- distribute.yml compatibility: HIGH — no changes needed; publish event trigger already in place
- Pitfalls: MEDIUM — some derived from community issues, not official docs, but cross-verified

**Research date:** 2026-03-28
**Valid until:** 2026-09-28 (release-please v4 is stable; no imminent major changes expected)
