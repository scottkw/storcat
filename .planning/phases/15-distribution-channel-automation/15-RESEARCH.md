# Phase 15: Distribution Channel Automation - Research

**Researched:** 2026-03-27
**Domain:** GitHub Actions, Homebrew cask automation, WinGet manifest automation
**Confidence:** HIGH

## Summary

Phase 15 automates two separate distribution channels: a Homebrew cask in the `homebrew-storcat` tap, and WinGet manifests in `microsoft/winget-pkgs`. Both channels are updated by a new workflow job that fires after the existing `release` job completes and the draft is manually published.

The central constraint from prior research is that SHA256 for the Homebrew cask must be computed locally (from the already-built artifact) rather than re-downloaded from the CDN. The release.yml already builds and uploads artifacts; Phase 15 adds two downstream jobs that consume those artifacts after publication.

The most important finding from this research phase is a **critical filename mismatch** between what Phase 14 produces and what the existing scripts/templates expect. This mismatch must be resolved before either distribution job can work. Additionally, `scottkw.StorCat` does not yet exist in `microsoft/winget-pkgs`, which blocks the `winget-releaser` action — the v2.1.0 stub must be submitted manually before automation can take over.

**Primary recommendation:** Build custom shell-script-based workflow jobs rather than relying on third-party marketplace actions. Third-party actions introduce supply-chain risk, require unclear version pinning, and the Homebrew job is simpler as a direct push script. Use `winget-releaser` for WinGet only after the first manual submission is done.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DIST-01 | Homebrew cask in `homebrew-storcat` auto-updated on release (SHA256 computed locally) | Custom workflow job: download DMG artifact, compute SHA256 with `shasum`, sed-replace template, push to tap repo via PAT |
| DIST-02 | WinGet manifest auto-submitted to `microsoft/winget-pkgs` on release | `winget-releaser` action (or `komac` CLI directly) after manual first-submission unblocks automation |
| DIST-03 | WinGet manifests in main repo auto-updated with new version on release | Additional step in same job: generate new version directory, commit back to storcat main repo |
</phase_requirements>

---

## Project Constraints (from CLAUDE.md)

- SHA-pin all third-party GitHub Actions to full 40-char commit SHAs (CICD-06, already established)
- `GITHUB_TOKEN` is insufficient for cross-repo operations — separate classic PATs required
- WinGet first submission to winget-pkgs must be manual — automation only works after package exists
- Homebrew SHA256 must be computed locally before release upload — never re-download from CDN
- No emojis in files
- GitHub Actions: `contents: write` permission is declared at workflow level; distribution jobs may need additional PAT scopes

---

## Critical Filename Mismatches (BLOCKING)

Before any automation works, two filename mismatches must be resolved:

### Mismatch 1: Homebrew DMG name

| Source | Value |
|--------|-------|
| Phase 14 release.yml produces | `StorCat-$VERSION-darwin-universal.dmg` |
| Cask template expects | `StorCat-#{version}-universal.dmg` (no `darwin-` segment) |

The cask URL pattern is: `https://github.com/scottkw/storcat/releases/download/#{version}/StorCat-#{version}-universal.dmg`

**Resolution options:**
1. Fix the cask template URL to use the `darwin-universal` filename (preferred — matches what CI already produces)
2. Rename the artifact in release.yml to drop `darwin-` (changes existing Phase 14 behavior)

Option 1 is lower risk. The template line becomes:
```
url "https://github.com/scottkw/storcat/releases/download/#{version}/StorCat-#{version}-darwin-universal.dmg"
```

### Mismatch 2: WinGet installer URL

| Source | Value |
|--------|-------|
| Phase 14 release.yml produces | `StorCat-$VERSION-windows-amd64-installer.exe` |
| 2.1.0 stub manifest expects | `StorCat.2.1.0.exe` (old v1 dot-notation format) |
| update-winget.sh generates | `StorCat.${VERSION}.exe` (also wrong) |

**Resolution:** Update manifests to use the hyphenated naming convention that Phase 14 actually produces.

---

## Standard Stack

### Core
| Library/Tool | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `vedantmgoyal9/winget-releaser` | v2 (pin to SHA) | Auto-submit PR to microsoft/winget-pkgs | Uses Komac under the hood; cross-platform (runs on ubuntu); actively maintained (280+ commits) |
| Bash + `shasum` | macOS built-in | Compute SHA256 of DMG locally | Available on every macOS/Linux runner; no extra install |
| `gh` CLI | pre-installed on runners | Git operations, API calls | Native on all GitHub-hosted runners |

### Supporting
| Library/Tool | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `actions/checkout` | SHA-pinned v4 | Check out tap repo | Required for direct-push Homebrew approach |
| `softprops/action-gh-release` | already in use | Trigger marker for distribution jobs | Existing dependency |
| `sed` | system | Template variable substitution in cask file | Simple, reliable; already used in update-tap.sh |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom shell Homebrew job | `macauley/action-homebrew-bump-cask` | Bump-cask creates a PR against official homebrew/cask, not a custom tap; limited maintenance; not suitable here |
| `winget-releaser` | `komac` CLI directly | winget-releaser wraps komac with GH Actions integration; either works; winget-releaser is simpler |
| `on: release types: [published]` trigger | `on: push tags` | Tags fire on draft creation; `release: published` fires only when draft is manually published — matches our existing draft: true pattern |

---

## Architecture Patterns

### Workflow Trigger Pattern

The existing release.yml creates a **draft release** (`draft: true`). Distribution automation must fire **after the draft is published**, not on tag push. The correct trigger is:

```yaml
on:
  release:
    types: [published]
```

This is a SEPARATE workflow file (e.g., `.github/workflows/distribute.yml`) or additional jobs added to `release.yml` with a `if: github.event_name == 'release'` guard.

**Recommended:** Add distribution jobs to a new `distribute.yml` that triggers on `release: published`. This keeps release.yml focused on building and keeps distribution separate.

### Pattern 1: Homebrew Cask Direct Push

**What:** Workflow downloads the already-uploaded DMG from the GitHub release, computes SHA256, applies it to the cask template, and direct-pushes to the `homebrew-storcat` tap repo.

**When to use:** Custom private/personal taps where you control the repo and don't need a PR review gate.

**Why direct push, not PR:** The `homebrew-storcat` tap is owned by the same user (`scottkw`). A PR to your own repo is unnecessary ceremony.

**Flow:**
1. Trigger: `release: published`
2. Download the DMG asset from the published release URL using `curl` (release is now public)
3. Compute SHA256 with `shasum -a 256`
4. Check out `homebrew-storcat` repo using the Homebrew PAT
5. Apply `sed` replacements on the template to produce `Casks/storcat.rb`
6. Commit and push

```yaml
# Source: pattern derived from josh.fail article + STATE.md decision
- name: Compute SHA256
  run: |
    VERSION="${{ github.event.release.tag_name }}"
    DMG_URL="https://github.com/scottkw/storcat/releases/download/${VERSION}/StorCat-${VERSION}-darwin-universal.dmg"
    SHA256=$(curl -sL "$DMG_URL" | shasum -a 256 | awk '{print $1}')
    echo "SHA256=$SHA256" >> $GITHUB_ENV
    echo "VERSION_CLEAN=${VERSION#v}" >> $GITHUB_ENV

- name: Update cask
  run: |
    sed -e "s/{{VERSION}}/$VERSION_CLEAN/" \
        -e "s/{{SHA256}}/$SHA256/" \
        packaging/homebrew/storcat.rb.template > /tmp/storcat.rb
```

**Note:** `github.event.release.tag_name` provides the version with the `v` prefix. The cask uses version without `v` prefix (e.g., `2.2.0` not `v2.2.0`), so strip the prefix.

### Pattern 2: WinGet PR Submission via winget-releaser

**What:** Triggers Komac to compute SHA256 from the installer URL, generate manifests, fork `microsoft/winget-pkgs`, and submit a PR.

**When to use:** After at least one version of `scottkw.StorCat` exists in the official winget-pkgs.

```yaml
# Source: vedantmgoyal9/winget-releaser documentation
- uses: vedantmgoyal9/winget-releaser@PINNED_SHA
  with:
    identifier: scottkw.StorCat
    token: ${{ secrets.WINGET_TOKEN }}
    installers-regex: 'StorCat-.*-windows-amd64-installer\.exe'
```

**Note:** `installers-regex` is critical — the release contains multiple files. Without it, the action may pick up wrong assets (AppImage, DMG, etc.).

### Pattern 3: WinGet Manifests in Main Repo (DIST-03)

**What:** In addition to submitting to winget-pkgs, the manifests in `packaging/winget/manifests/s/scottkw/StorCat/$VERSION/` are auto-generated and committed back to the main repo.

**How:** Add a step in the distribute.yml that:
1. Computes SHA256 of the NSIS installer (same `curl | shasum` approach)
2. Creates new directory `packaging/winget/manifests/s/scottkw/StorCat/$VERSION_CLEAN/`
3. Generates the three YAML files from the existing 2.1.0 stubs as templates
4. Commits and pushes to the main storcat repo

This step runs with `GITHUB_TOKEN` (same-repo write) — no PAT needed.

### Recommended Project Structure (new files)
```
.github/workflows/
├── release.yml          # existing — builds, packages, creates draft release
└── distribute.yml       # NEW — triggers on release:published, updates channels

packaging/homebrew/
├── storcat.rb.template  # existing — fix URL to use darwin-universal filename
└── update-tap.sh        # existing — updated for CI use (non-interactive)
```

### Anti-Patterns to Avoid
- **Re-downloading from CDN for SHA256:** Never fetch from `releases/download/` CDN for SHA256 after publish. Always compute from the locally available artifact (download from release URL immediately after publish is fine since the release is now public, but note the CDN may have propagation delay — use `--retry` on curl).
- **Using `GITHUB_TOKEN` for cross-repo push:** `GITHUB_TOKEN` cannot write to `homebrew-storcat`. A PAT with `repo` scope on that repo is required.
- **Using fine-grained PAT for WinGet:** `winget-releaser`/Komac explicitly require classic PATs with `public_repo` scope. Fine-grained tokens do not work for PR creation on `microsoft/winget-pkgs`.
- **Triggering on `push: tags`:** The release is a draft on tag push. Distribution must wait for the release to be published.
- **Relying on `workflow_call` or tag triggers for the distribution job:** The release event `published` type is the correct and only reliable trigger.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| WinGet PR submission | Custom fork+PR bash script | `winget-releaser` / Komac | Komac handles manifest schema validation, multi-arch URL detection, PR creation — complex to replicate |
| SHA256 of release asset | Download twice | `curl ... \| shasum -a 256` inline | Zero dependencies; asset is already public after release publish |

**Key insight:** For WinGet, let Komac handle the manifest complexity (schema compliance, architecture mapping). For Homebrew, the problem is simple enough (one file, one SHA256) that a direct shell approach in the workflow is preferable to any third-party action.

---

## Common Pitfalls

### Pitfall 1: Package Not in winget-pkgs Yet
**What goes wrong:** `winget-releaser` will fail with "package not found" or similar if `scottkw.StorCat` has never been accepted into `microsoft/winget-pkgs`.
**Why it happens:** Komac requires an existing package to base new manifests on.
**How to avoid:** The v2.1.0 manifests in `packaging/winget/` must be manually submitted as a PR to `microsoft/winget-pkgs` before Phase 15 is considered done. This is a prerequisite human step.
**Warning signs:** winget-releaser action fails with "no existing manifest found" type errors.

### Pitfall 2: CDN Propagation Delay for SHA256
**What goes wrong:** The distribute.yml job fires immediately after release publish. The DMG download via `curl` may 404 or return a partial file if CDN hasn't propagated yet.
**Why it happens:** GitHub release assets are served via CDN with propagation delay after publishing.
**How to avoid:** Add `--retry 5 --retry-delay 10` to the curl command that downloads the DMG for SHA256 computation.
**Warning signs:** `shasum` returns empty hash or job fails with curl error.

### Pitfall 3: Version Prefix Mismatch (v2.2.0 vs 2.2.0)
**What goes wrong:** The cask `version` field must not have the `v` prefix. The WinGet `PackageVersion` field also must not have the `v` prefix. `github.event.release.tag_name` includes the `v` prefix.
**Why it happens:** Git tags use `v*.*.*` convention; package registries use bare version numbers.
**How to avoid:** Always strip the `v` prefix: `VERSION_CLEAN="${TAG#v}"`. This is already done in Phase 14's deb build steps.
**Warning signs:** `brew info storcat` shows `storcat vv2.2.0`; WinGet manifest validation fails on version format.

### Pitfall 4: Wrong Installer File Selected by winget-releaser
**What goes wrong:** Without `installers-regex`, winget-releaser may try to process all release assets including the bare `.exe`, AppImage, `.deb`, etc.
**Why it happens:** Komac auto-detects installer types from URLs; multiple `.exe` files cause ambiguity.
**How to avoid:** Set `installers-regex: 'StorCat-.*-windows-amd64-installer\.exe'` to exactly match only the NSIS installer.
**Warning signs:** Action uploads wrong file or creates manifest with wrong installer type.

### Pitfall 5: Homebrew Tap Cask File Path
**What goes wrong:** The tap repo (`homebrew-storcat`) must have the cask at `Casks/storcat.rb`. If the repo structure differs, `brew install` will fail.
**Why it happens:** Homebrew expects casks in a `Casks/` directory for custom taps.
**How to avoid:** Verify `homebrew-storcat` repo structure before implementing workflow. The update-tap.sh already hardcodes `CASK_FILE="Casks/storcat.rb"`.

### Pitfall 6: Existing update-tap.sh is Interactive
**What goes wrong:** The current `update-tap.sh` contains `read -p "Push changes? (y/N):"` which will hang in CI.
**Why it happens:** Script was written for local manual use.
**How to avoid:** CI workflow should NOT call `update-tap.sh`. Instead, inline the equivalent non-interactive steps directly in the workflow YAML.

---

## Code Examples

### Complete distribute.yml skeleton (Homebrew job)
```yaml
# Source: adapted from josh.fail pattern + STATE.md SHA256 constraint
name: Distribute

on:
  release:
    types: [published]

jobs:
  update-homebrew:
    runs-on: ubuntu-22.04
    steps:
      - uses: actions/checkout@PINNED_SHA  # check out storcat main repo (for template)

      - name: Get version and compute SHA256
        id: sha
        run: |
          VERSION="${{ github.event.release.tag_name }}"
          VERSION_CLEAN="${VERSION#v}"
          DMG_URL="https://github.com/scottkw/storcat/releases/download/${VERSION}/StorCat-${VERSION}-darwin-universal.dmg"
          SHA256=$(curl -sL --retry 5 --retry-delay 10 "$DMG_URL" | shasum -a 256 | awk '{print $1}')
          echo "version=$VERSION_CLEAN" >> $GITHUB_OUTPUT
          echo "sha256=$SHA256" >> $GITHUB_OUTPUT

      - name: Checkout homebrew-storcat tap
        uses: actions/checkout@PINNED_SHA
        with:
          repository: scottkw/homebrew-storcat
          token: ${{ secrets.HOMEBREW_TAP_TOKEN }}
          path: tap

      - name: Update cask
        run: |
          sed -e 's/{{VERSION}}/${{ steps.sha.outputs.version }}/g' \
              -e 's/{{SHA256}}/${{ steps.sha.outputs.sha256 }}/g' \
              $GITHUB_WORKSPACE/packaging/homebrew/storcat.rb.template \
              > tap/Casks/storcat.rb

      - name: Commit and push
        working-directory: tap
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add Casks/storcat.rb
          git commit -m "Update StorCat to ${{ steps.sha.outputs.version }}"
          git push
```

### Complete distribute.yml skeleton (WinGet job)
```yaml
  update-winget:
    runs-on: ubuntu-22.04
    steps:
      - uses: vedantmgoyal9/winget-releaser@PINNED_SHA
        with:
          identifier: scottkw.StorCat
          token: ${{ secrets.WINGET_TOKEN }}
          installers-regex: 'StorCat-.*-windows-amd64-installer\.exe'
```

### WinGet manifest update in main repo (DIST-03)
```yaml
  update-winget-manifests:
    runs-on: ubuntu-22.04
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@PINNED_SHA

      - name: Compute installer SHA256 and generate manifests
        run: |
          VERSION="${{ github.event.release.tag_name }}"
          VERSION_CLEAN="${VERSION#v}"
          INSTALLER_URL="https://github.com/scottkw/storcat/releases/download/${VERSION}/StorCat-${VERSION}-windows-amd64-installer.exe"
          SHA256=$(curl -sL --retry 5 --retry-delay 10 "$INSTALLER_URL" | shasum -a 256 | awk '{print $1}')

          MANIFEST_DIR="packaging/winget/manifests/s/scottkw/StorCat/$VERSION_CLEAN"
          PREV_DIR="packaging/winget/manifests/s/scottkw/StorCat/2.1.0"
          mkdir -p "$MANIFEST_DIR"

          # Generate from 2.1.0 templates with sed
          sed "s/2\.1\.0/$VERSION_CLEAN/g" "$PREV_DIR/scottkw.StorCat.yaml" > "$MANIFEST_DIR/scottkw.StorCat.yaml"
          sed -e "s/2\.1\.0/$VERSION_CLEAN/g" \
              -e "s|StorCat\.2\.1\.0\.exe|StorCat-${VERSION}-windows-amd64-installer.exe|g" \
              -e "s/0000000000000000000000000000000000000000000000000000000000000000/$SHA256/g" \
              "$PREV_DIR/scottkw.StorCat.installer.yaml" > "$MANIFEST_DIR/scottkw.StorCat.installer.yaml"
          sed "s/2\.1\.0/$VERSION_CLEAN/g" "$PREV_DIR/scottkw.StorCat.locale.en-US.yaml" > "$MANIFEST_DIR/scottkw.StorCat.locale.en-US.yaml"

          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add "$MANIFEST_DIR"
          git commit -m "chore: add WinGet manifests for $VERSION_CLEAN"
          git push
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `brew bump-cask-pr` (creates PR to homebrew/core) | Direct push to custom tap | 2023+ | Simpler; no review gate needed for personal taps |
| YamlCreate (PowerShell, Windows-only) | Komac (Rust, cross-platform) | 2023 | WinGet automation now works on Linux runners |
| `on: push tags` for distribution | `on: release types: [published]` | Always correct | Prevents distributing before release assets are finalized |

**Deprecated/outdated:**
- `winget-create submit`: Older tool, largely superseded by Komac for automation workflows
- `brew cask audit --strict`: Removed in recent Homebrew versions; just `brew audit`

---

## Open Questions

1. **Does `scottkw.StorCat` need to be in winget-pkgs before Phase 15 can be called complete?**
   - What we know: `winget-releaser` requires at least one existing version in the repo. The 2.1.0 stubs exist in `packaging/winget/` but have never been submitted.
   - What's unclear: Whether the Phase 15 success criteria considers the initial PR submission as part of this phase or a prerequisite.
   - Recommendation: Treat the manual first-submission of 2.1.0 stubs (with correct SHA256) as Task 1 in the plan. Block automation on that being merged.

2. **Which version should be the first manual WinGet submission: 2.1.0 (already built) or 2.2.0 (built by the new pipeline)?**
   - What we know: The 2.1.0 stubs have placeholder SHA256 values.
   - What's unclear: Whether submitting 2.2.0 manifests (the first real release) as both the initial submission and the auto-submitted version is acceptable.
   - Recommendation: Submit 2.2.0 as the first version when Phase 15 is executed. Skip 2.1.0 for WinGet purposes.

3. **PAT token names and repository secrets that need to be created**
   - What we know: `HOMEBREW_TAP_TOKEN` (repo write to homebrew-storcat) and `WINGET_TOKEN` (classic PAT, public_repo scope, for winget-releaser) are both needed.
   - What's unclear: Whether these secrets already exist in the repository settings.
   - Recommendation: Document as "must create before workflow runs" in the plan.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `gh` CLI | WinGet/Homebrew git ops | Yes (runner pre-installed) | pre-installed | — |
| `shasum` | SHA256 computation | Yes (macOS/Linux built-in) | built-in | `sha256sum` on Linux (already present) |
| `curl` | Asset download | Yes (runner pre-installed) | pre-installed | — |
| `sed` | Template substitution | Yes (runner pre-installed) | built-in | — |
| `winget` CLI | Local WinGet validation | No (not on Linux runner) | — | Skip validation step; Komac handles validation |
| `brew` | Local cask audit | No (not on ubuntu runner) | — | Skip audit step |
| `HOMEBREW_TAP_TOKEN` secret | Cross-repo push | Unknown | — | Must be created |
| `WINGET_TOKEN` secret | winget-releaser | Unknown | — | Must be created |

**Missing dependencies with no fallback:**
- `HOMEBREW_TAP_TOKEN` — classic or fine-grained PAT with `contents: write` on `scottkw/homebrew-storcat`. Must be created in repo secrets before workflow runs.
- `WINGET_TOKEN` — classic PAT with `public_repo` scope. Must be created and added to repo secrets. Fine-grained tokens do not work.

**Missing dependencies with fallback:**
- `winget` validation — not available on Linux runners. Komac performs its own manifest validation before submission; skip explicit local validation.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | GitHub Actions workflow integration test (no unit test framework) |
| Config file | `.github/workflows/distribute.yml` (to be created) |
| Quick run command | `gh workflow run distribute.yml --ref main` (manual trigger via `workflow_dispatch` for testing) |
| Full suite command | Push a test release tag, publish the draft, observe distribute.yml run |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DIST-01 | Homebrew cask updated in homebrew-storcat after release publish | integration | Observe distribute.yml job `update-homebrew` completes green | No — Wave 0 |
| DIST-01 | SHA256 in cask matches actual DMG file | integration | `brew fetch --cask storcat` passes audit | No — Wave 0 |
| DIST-02 | PR submitted to microsoft/winget-pkgs with new version | integration | PR visible in microsoft/winget-pkgs after distribute.yml runs | No — Wave 0 |
| DIST-03 | packaging/winget/ manifests committed to main repo with new version | integration | `git log --oneline packaging/winget/` shows auto-commit | No — Wave 0 |

### Wave 0 Gaps
- [ ] `.github/workflows/distribute.yml` — the entire workflow to be created
- [ ] `packaging/homebrew/storcat.rb.template` URL fix — `darwin-universal` filename
- [ ] `packaging/winget/manifests/s/scottkw/StorCat/2.1.0/scottkw.StorCat.installer.yaml` — fix placeholder SHA256 and installer URL to correct format
- [ ] GitHub repository secrets: `HOMEBREW_TAP_TOKEN`, `WINGET_TOKEN`
- [ ] Manual first-submission of storcat to microsoft/winget-pkgs (human prerequisite)

---

## Sources

### Primary (HIGH confidence)
- `release.yml` local file — exact artifact filenames produced by Phase 14
- `packaging/homebrew/storcat.rb.template` local file — current cask URL pattern (mismatch documented above)
- `packaging/winget/manifests/s/scottkw/StorCat/2.1.0/` local files — current stub format
- `.planning/STATE.md` — key decisions: SHA256 locally, GITHUB_TOKEN insufficient for cross-repo, first WinGet submission must be manual
- [vedantmgoyal9/winget-releaser README](https://github.com/vedantmgoyal9/winget-releaser) — inputs, PAT requirements, fork requirements, pre-existence requirement
- [Komac README](https://github.com/russellbanks/Komac/blob/main/README.md) — CLI usage, token requirements

### Secondary (MEDIUM confidence)
- [josh.fail — Automate Homebrew formulae with GitHub Actions](https://josh.fail/2023/automate-updating-custom-homebrew-formulae-with-github-actions/) — cross-repo trigger + push pattern
- GitHub community discussion on release trigger types — `published` vs `released` vs draft behavior
- WebSearch results confirming `on: release types: [released]` fires when draft is published

### Tertiary (LOW confidence)
- WebSearch results on `InstallerType: nullsoft` vs `exe` for NSIS — not yet validated against actual Komac behavior for this specific installer

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified against official action documentation and STATE.md prior decisions
- Architecture: HIGH — filename mismatches verified against actual local files; trigger pattern verified against GitHub docs
- Pitfalls: HIGH — CDN delay, interactive script, PAT type all verified; WinGet pre-existence confirmed by gh API query
- WinGet installer type: MEDIUM — NSIS type handling assumed correct from general winget docs; not tested with Komac specifically

**Research date:** 2026-03-27
**Valid until:** 2026-04-27 (stable tooling, but verify winget-releaser SHA before use)
