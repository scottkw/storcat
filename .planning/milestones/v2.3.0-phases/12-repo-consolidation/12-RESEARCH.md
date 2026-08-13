# Phase 12: Repo Consolidation - Research

**Researched:** 2026-03-27
**Domain:** Git repository management, WinGet manifest structure, Homebrew cask templating, GitHub CLI
**Confidence:** HIGH

## Summary

Phase 12 is a pure file-migration and repo-management phase — no new code, no new workflows. The goal is to move WinGet manifests and Homebrew packaging files from two satellite repos (`winget-storcat`, `homebrew-storcat`) into the main storcat repo under `packaging/`, update the satellite READMEs to reflect their new auto-managed status, and archive `winget-storcat`.

Both satellite repos have been fully audited. Their file contents are known, their current state is documented below, and the exact path mappings for the migration are clear. The `homebrew-storcat` repo must remain alive (it is the Homebrew tap target — `brew tap scottkw/storcat` resolves to `github.com/scottkw/homebrew-storcat`), but `winget-storcat` can be fully archived once its manifests are committed to main.

**Key gap discovered:** The winget manifests in `winget-storcat` only go up to v1.2.3 (the Electron version). There are no manifests for v2.0.0 or v2.1.0. The phase must create v2.1.0 manifests from scratch when populating `packaging/winget/manifests/` — and update the description text to reflect the Go/Wails rewrite (current text still says "Electron-based"). Similarly, the Homebrew cask still references v1.2.3.

**Primary recommendation:** Copy all files mechanically, fix the stale metadata, write a template for the cask, update both satellite READMEs, then archive `winget-storcat`. This phase has no automation — that is Phase 15.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| REPO-01 | WinGet manifests moved to `packaging/winget/` in main repo | All 6 version directories (1.1.1–1.2.3) plus new v2.1.0 manifests must be committed under `packaging/winget/manifests/s/scottkw/StorCat/` |
| REPO-02 | Homebrew cask template and update script moved to `packaging/homebrew/` in main repo | `Casks/storcat.rb` becomes `packaging/homebrew/storcat.rb.template` (with version/SHA placeholders); `update-tap.sh` moves from `homebrew-storcat` root to `packaging/homebrew/update-tap.sh` |
| REPO-03 | `winget-storcat` repo archived after migration verified | `gh repo archive scottkw/winget-storcat -y` — irreversible, do after verifying manifests are correct in main repo |
| REPO-04 | `homebrew-storcat` README updated to indicate auto-managed status | Update via `gh api` PATCH or a commit to homebrew-storcat; keep tap alive |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

- **GSD workflow enforcement**: All file-changing work must go through a GSD command (`/gsd:execute-phase` for this planned phase work).
- **No direct repo edits outside GSD workflow** unless user explicitly bypasses.
- **pnpm preferred** for Node; `uv` or `pip` for Python — not relevant to this phase (shell scripts only).
- **Python virtual environment** if any Python scripts are run — not applicable here.
- **Commit conventions**: Standard git commits; no `--no-verify` skipping.

## Current State Audit

### winget-storcat Repo (`scottkw/winget-storcat`)

**Status:** Public, not archived, default branch `main`

**Files present:**
```
.gitignore
README.md
manifests/s/scottkw/StorCat/
  1.1.1/  (3 YAML files)
  1.1.2/  (3 YAML files)
  1.2.0/  (3 YAML files)
  1.2.1/  (3 YAML files)
  1.2.2/  (3 YAML files)
  1.2.3/  (3 YAML files)
update-winget.sh
```

**YAML files per version (manifest schema 1.6.0):**
- `scottkw.StorCat.yaml` — version manifest
- `scottkw.StorCat.installer.yaml` — installer manifest (portable exe, x64 only)
- `scottkw.StorCat.locale.en-US.yaml` — locale/description manifest

**Known defects in existing manifests (require fixing during migration):**
1. `scottkw.StorCat.installer.yaml` for v1.2.3 has a malformed SHA256 field — the `echo` stdout from the script leaked into the YAML value. The SHA256 line reads: `InstallerSha256: 📥 Downloading: https://...` followed by the actual hash on a new line. This is a pre-existing bug in the satellite repo. The migration should carry it forward as-is for historical versions (they are reference copies) but flag the defect in comments.
2. Locale manifest description still says "Electron-based" — acceptable for historical versions, but v2.1.0 manifest must use updated Go/Wails description.

**Version gap:** Manifests exist for v1.1.1 through v1.2.3 only. No manifests exist for v2.0.0 or v2.1.0. The current app version is 2.1.0 (`wails.json` `productVersion`). The GitHub releases only go up to v1.2.3 — v2.x has not been released yet. The planner must decide: create placeholder v2.1.0 manifests now (pointing to future release assets), or leave v2.x manifests for Phase 15. **Recommendation: create v2.1.0 manifest stubs now with correct structure and updated description text, leaving placeholder SHA256 values** (they will be real values when Phase 15 runs). This satisfies REPO-01 ("current version manifests committed").

### homebrew-storcat Repo (`scottkw/homebrew-storcat`)

**Status:** Public, not archived, default branch `main`

**Files present:**
```
.gitignore
README.md
Casks/
  storcat.rb      (current cask, v1.2.3)
update-tap.sh
```

**Current cask (`Casks/storcat.rb`):**
- Version: `1.2.3` (stale — app is now v2.1.0)
- SHA256: `812bcf43ac85cdbe591baa297f0e36855320fb04304b36261cc48da53aa51908`
- URL pattern: `https://github.com/scottkw/storcat/releases/download/#{version}/StorCat-#{version}-universal.dmg`
- Contains `app "StorCat.app"` and `zap trash:` block — keep in template

**`update-tap.sh` behavior:**
- Fetches latest release from GitHub API
- Downloads DMG, computes SHA256 via `shasum -a 256`
- Regenerates `Casks/storcat.rb` in-place (overwrites it)
- Commits and optionally pushes to GitHub
- Must be run from inside the `homebrew-storcat` repo root (uses relative paths)
- Path dependency: script assumes `$CASK_FILE="Casks/storcat.rb"` relative to CWD

**Path issue when migrating:** `update-tap.sh` is designed to run from `homebrew-storcat/`. When moved to `packaging/homebrew/` in the main repo, it still needs to push to `homebrew-storcat` (a different repo). The script currently does `git add "$CASK_FILE"` and `git commit` — this will not work from the main repo. The template approach decouples this: `storcat.rb.template` lives in main repo for reference; the actual push-to-tap script needs adaptation for cross-repo operation (that is Phase 15 work). For Phase 12, copy the script as-is and document the cross-repo limitation.

**Tap naming constraint (HIGH confidence):**
`brew tap scottkw/storcat` resolves to `github.com/scottkw/homebrew-storcat` via Homebrew's naming convention. The `homebrew-storcat` repo MUST remain alive and publicly accessible for this tap to continue working. Archiving or renaming it breaks all existing installations' `brew upgrade` command.

## Standard Stack

### Tools Used in This Phase
| Tool | Purpose | Version/Notes |
|------|---------|---------------|
| `gh` CLI | Archive repo, update README via API, list contents | Installed (`gh repo archive` confirmed available) |
| Git | Commit packaging files to main repo | Standard |
| Bash | Script adaptation when needed | System bash |

### WinGet Manifest Schema
**Schema version in use:** `1.6.0` (confirmed from existing YAML files)
**Schema URLs:**
- Version: `https://aka.ms/winget-manifest.version.1.6.0.schema.json`
- Installer: `https://aka.ms/winget-manifest.installer.1.6.0.schema.json`
- Locale: `https://aka.ms/winget-manifest.defaultLocale.1.6.0.schema.json`

**Required fields for installer manifest (portable type):**
```yaml
PackageIdentifier: scottkw.StorCat
PackageVersion: 2.1.0
InstallerType: portable
Commands:
- storcat
Installers:
- Architecture: x64
  InstallerUrl: https://github.com/scottkw/storcat/releases/download/2.1.0/StorCat.2.1.0.exe
  InstallerSha256: <sha256>
ManifestType: installer
ManifestVersion: 1.6.0
```

**Note on sha256 placeholder:** WinGet validates SHA256 format. For stub manifests, use a 64-character zero string: `0000000000000000000000000000000000000000000000000000000000000000`. This is clearly a placeholder and won't validate against a real installer, but the file structure is correct.

### Homebrew Cask Template Format
The cask is in Ruby DSL format. Template uses shell-style `{{VERSION}}` and `{{SHA256}}` placeholders (standard in update scripts):

```ruby
cask "storcat" do
  version "{{VERSION}}"
  sha256 "{{SHA256}}"

  url "https://github.com/scottkw/storcat/releases/download/#{version}/StorCat-#{version}-universal.dmg"

  name "StorCat"
  desc "Directory Catalog Manager - Create, search, and browse directory catalogs"
  homepage "https://github.com/scottkw/storcat"

  livecheck do
    url :url
    strategy :github_latest
  end

  app "StorCat.app"

  zap trash: [
    "~/Library/Application Support/StorCat",
    "~/Library/Preferences/com.kenscott.storcat.plist",
    "~/Library/Saved Application State/com.kenscott.storcat.savedState",
    "~/Library/Caches/com.kenscott.storcat",
  ]
end
```

**Note:** In Homebrew Ruby DSL, `#{version}` is Ruby string interpolation (works in live cask). In a `.template` file, `#{version}` should remain as-is (the update script substitutes `{{VERSION}}` before writing the live cask). Double-check the existing `update-tap.sh` — it writes the cask with literal `$VERSION` shell variable, not `#{version}` Ruby interpolation. The live cask uses `#{version}` which is Homebrew's own interpolation. The template approach needs to preserve that distinction.

## Architecture Patterns

### Resulting Directory Structure

```
storcat/                                  # Main repo
├── packaging/
│   ├── homebrew/
│   │   ├── storcat.rb.template           # Cask template with {{VERSION}} and {{SHA256}} placeholders
│   │   └── update-tap.sh                 # Script that pushes updated cask to homebrew-storcat
│   └── winget/
│       ├── manifests/
│       │   └── s/scottkw/StorCat/
│       │       ├── 1.1.1/               # Historical (copied as-is from winget-storcat)
│       │       ├── 1.1.2/
│       │       ├── 1.2.0/
│       │       ├── 1.2.1/
│       │       ├── 1.2.2/
│       │       ├── 1.2.3/
│       │       └── 2.1.0/               # New — stub manifests for current app version
│       └── update-winget.sh              # Script that generates new version manifests
├── .github/
│   └── workflows/
│       └── build.yml                     # Existing — not modified in Phase 12
└── ...
```

### Migration Pattern: Copy-Then-Adapt

Phase 12 is copy-first, adapt-second:
1. Copy manifests verbatim (no edits to historical versions)
2. Create new v2.1.0 manifests with updated content
3. Convert live cask to template format
4. Copy update scripts (adapt paths only if needed for local execution)
5. Update satellite READMEs
6. Archive `winget-storcat`

## Don't Hand-Roll

| Problem | Don't Build | Use Instead |
|---------|-------------|-------------|
| Archiving a GitHub repo | Custom API call | `gh repo archive scottkw/winget-storcat -y` |
| Updating remote README | Manual API JSON construction | `gh api --method PUT repos/scottkw/homebrew-storcat/contents/README.md` with base64 content, or commit to a local clone |
| Computing SHA256 | Custom hash utility | `shasum -a 256 file` (macOS) / `sha256sum file` (Linux) |

## Common Pitfalls

### Pitfall 1: Homebrew tap naming — must keep homebrew-storcat alive
**What goes wrong:** Archiving `homebrew-storcat` or renaming it breaks `brew tap scottkw/storcat` for all existing users. Homebrew resolves `brew tap <user>/<name>` to `github.com/<user>/homebrew-<name>`. There is no workaround after archival — GitHub archived repos are read-only but still accessible for git clone, so `brew tap` actually continues to work on archived repos. However, the update script can no longer push to an archived repo.
**How to avoid:** Archive `winget-storcat` only. Keep `homebrew-storcat` live (unarchived) — it must accept push commits for Phase 15 automation.

### Pitfall 2: WinGet SHA256 format in malformed v1.2.3 manifest
**What goes wrong:** The `scottkw.StorCat.installer.yaml` for v1.2.3 has a malformed SHA256 value (echo output leaked into the YAML). If this is copied verbatim, the manifest file in main repo is also malformed. This is acceptable for historical reference copies (they are not submitted to `microsoft/winget-pkgs`), but should be noted in a comment.
**How to avoid:** Copy as-is but add a YAML comment noting the defect; clean it up when creating the v2.1.0 manifests.

### Pitfall 3: `update-tap.sh` path assumption — runs from homebrew-storcat root
**What goes wrong:** The script uses `CASK_FILE="Casks/storcat.rb"` relative path and does `git add`/`git commit`. If run from `packaging/homebrew/` in the main repo, it will try to commit to the main repo instead of `homebrew-storcat`.
**How to avoid:** For Phase 12, copy the script as-is and document that it must be run from a local clone of `homebrew-storcat`. Full cross-repo adaptation is Phase 15 work.

### Pitfall 4: v2.1.0 manifests reference non-existent release assets
**What goes wrong:** v2.0.0 and v2.1.0 have not been published as GitHub releases yet. Any v2.1.0 WinGet manifests will have placeholder SHA256 values and will not actually install via `winget`.
**How to avoid:** Create manifests with clearly marked placeholder SHA256 (`0000...0000`) and add a comment. REPO-01 requires "current version manifests committed" — stub manifests satisfy this structurally. Real values come in Phase 15 when release assets exist.

### Pitfall 5: Description text still says "Electron-based"
**What goes wrong:** All existing WinGet locale manifests describe StorCat as "a modern Electron-based directory catalog management application." This was accurate for v1.x but is wrong for v2.x (Go/Wails).
**How to avoid:** Update the description in the v2.1.0 locale manifest. Leave historical manifests unchanged.

### Pitfall 6: `gh repo archive` is irreversible via CLI
**What goes wrong:** Archiving via `gh repo archive` cannot be undone via CLI — requires GitHub web UI to unarchive.
**How to avoid:** Verify manifests are correct in main repo before archiving. Run the archive step last.

## Code Examples

### Create v2.1.0 version manifest
```yaml
# yaml-language-server: $schema=https://aka.ms/winget-manifest.version.1.6.0.schema.json

PackageIdentifier: scottkw.StorCat
PackageVersion: 2.1.0
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.6.0
```

### Create v2.1.0 installer manifest
```yaml
# yaml-language-server: $schema=https://aka.ms/winget-manifest.installer.1.6.0.schema.json

PackageIdentifier: scottkw.StorCat
PackageVersion: 2.1.0
InstallerType: portable
Commands:
- storcat
Installers:
- Architecture: x64
  InstallerUrl: https://github.com/scottkw/storcat/releases/download/2.1.0/StorCat.2.1.0.exe
  InstallerSha256: 0000000000000000000000000000000000000000000000000000000000000000
ManifestType: installer
ManifestVersion: 1.6.0
```

### Create v2.1.0 locale manifest (updated description)
```yaml
# yaml-language-server: $schema=https://aka.ms/winget-manifest.defaultLocale.1.6.0.schema.json

PackageIdentifier: scottkw.StorCat
PackageVersion: 2.1.0
PackageLocale: en-US
Publisher: Ken Scott
PublisherUrl: https://github.com/scottkw
PublisherSupportUrl: https://github.com/scottkw/storcat/issues
PackageName: StorCat
PackageUrl: https://github.com/scottkw/storcat
License: MIT
LicenseUrl: https://github.com/scottkw/storcat/blob/main/LICENSE
ShortDescription: Directory Catalog Manager - Create, search, and browse directory catalogs
Description: |-
  StorCat is a cross-platform desktop application and CLI tool for creating, browsing,
  and searching directory catalogs. Built with Go and Wails for fast, lightweight operation.
  Generates JSON and HTML representations of directory trees.

  Features:
  - Create comprehensive directory catalogs with detailed file information
  - Fast full-text search across multiple catalog files
  - Interactive catalog browser with modern table interface
  - Dark/Light mode with complete theme support (11 built-in themes)
  - CLI subcommands: create, search, list, show, open, version
  - Cross-platform: macOS (universal), Windows (x64/arm64), Linux (x64/arm64)
  - 100% compatible with existing bash script catalog files (v1 format)
Moniker: storcat
Tags:
- catalog
- directory
- file-manager
- search
- go
- wails
- documentation
- json
- html
ReleaseNotesUrl: https://github.com/scottkw/storcat/releases/tag/2.1.0
Documentations:
- DocumentLabel: User Guide
  DocumentUrl: https://github.com/scottkw/storcat#readme
ManifestType: defaultLocale
ManifestVersion: 1.6.0
```

### Archive winget-storcat repo
```bash
gh repo archive scottkw/winget-storcat -y
```

### Update homebrew-storcat README via local clone
```bash
# Clone homebrew-storcat locally, edit README, commit, push
git clone https://github.com/scottkw/homebrew-storcat /tmp/homebrew-storcat-update
cd /tmp/homebrew-storcat-update
# Edit README.md
git add README.md
git commit -m "docs: mark tap as auto-managed, link to main repo"
git push origin main
```

### Homebrew cask template (packaging/homebrew/storcat.rb.template)
The template replaces hardcoded version/SHA with `{{VERSION}}` and `{{SHA256}}` placeholders for the update-tap.sh script to substitute at release time.

## Runtime State Inventory

> This section is included because Phase 12 involves renaming/migrating files across repos.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — no databases involved | None |
| Live service config | `homebrew-storcat/Casks/storcat.rb` — live Homebrew tap cask, currently v1.2.3 | Update to v2.1.0 as part of README update commit (or leave stale — Phase 15 will overwrite it anyway) |
| OS-registered state | None | None |
| Secrets/env vars | None — this phase has no automation requiring PATs | None |
| Build artifacts | None | None |

**homebrew-storcat cask staleness:** The cask is pinned at v1.2.3. Phase 12 does not update the live cask (that is Phase 15 distribution automation). The README update should note this explicitly so users are not confused by the stale cask.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `gh` CLI | Archive repo, API calls | Yes | (confirmed present) | Manual GitHub web UI |
| Git | Commit packaging files | Yes | system | — |
| Bash | Run/validate scripts | Yes | system | — |
| GitHub API access | Read satellite repo contents | Yes | Verified — can list both repos | — |

**No missing dependencies.** All tools required for Phase 12 are available.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | None (no test framework applicable — this phase is file migration + git operations) |
| Config file | none |
| Quick run command | `ls packaging/winget/manifests/s/scottkw/StorCat/ && ls packaging/homebrew/` |
| Full suite command | `gh repo view scottkw/winget-storcat --json isArchived` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| REPO-01 | `packaging/winget/manifests/` exists with version dirs including 2.1.0 | smoke | `ls packaging/winget/manifests/s/scottkw/StorCat/` | ❌ Wave 0 — directory doesn't exist yet |
| REPO-02 | `packaging/homebrew/storcat.rb.template` and `update-tap.sh` exist | smoke | `test -f packaging/homebrew/storcat.rb.template && test -f packaging/homebrew/update-tap.sh` | ❌ Wave 0 |
| REPO-03 | `winget-storcat` repo is archived | smoke | `gh repo view scottkw/winget-storcat --json isArchived \| python3 -c "import sys,json; d=json.load(sys.stdin); exit(0 if d['isArchived'] else 1)"` | ❌ Wave 0 — repo not yet archived |
| REPO-04 | `homebrew-storcat` README indicates auto-managed | smoke | `gh api repos/scottkw/homebrew-storcat/contents/README.md \| python3 -c "import sys,json,base64; d=json.load(sys.stdin); content=base64.b64decode(d['content']).decode(); exit(0 if 'auto-managed' in content.lower() or 'main repo' in content.lower() else 1)"` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `ls packaging/winget/manifests/s/scottkw/StorCat/ && ls packaging/homebrew/`
- **Per wave merge:** Full suite (all 4 smoke commands above)
- **Phase gate:** All 4 smoke commands green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `packaging/winget/manifests/s/scottkw/StorCat/` directory — covers REPO-01
- [ ] `packaging/homebrew/storcat.rb.template` — covers REPO-02
- [ ] `packaging/homebrew/update-tap.sh` — covers REPO-02

*(No test framework install needed — all validation is shell smoke tests)*

## Open Questions

1. **Should v2.1.0 WinGet manifests be stubs or deferred entirely?**
   - What we know: REPO-01 says "current version manifests committed" — current version is 2.1.0
   - What's unclear: Whether placeholder SHA256 satisfies the requirement
   - Recommendation: Create stub manifests with `0000...0000` SHA256, clearly commented. This completes the structural migration; Phase 15 fills in real values.

2. **Should the live cask in homebrew-storcat be updated to v2.1.0 in Phase 12?**
   - What we know: Current cask is v1.2.3 (stale). No v2.1.0 release assets exist yet.
   - What's unclear: Whether leaving the cask at v1.2.3 confuses users
   - Recommendation: Leave cask at v1.2.3 for now. The README update (REPO-04) will explain that the tap is auto-managed and link to the main repo. Phase 15 will update the live cask when real release assets exist.

3. **Should historical WinGet manifests (v1.1.1–v1.2.3) be copied into main repo?**
   - What we know: They are the complete history of the satellite repo
   - What's unclear: Whether there is value in carrying forward Electron-era manifests into the Go-era repo
   - Recommendation: Copy all historical versions for completeness. They are small YAML files and preserve audit history. Storage cost is negligible.

## Sources

### Primary (HIGH confidence)
- Direct `gh api` calls to `scottkw/winget-storcat` and `scottkw/homebrew-storcat` — all file contents verified
- `gh repo archive --help` — archive command confirmed available
- `wails.json` in main repo — current version confirmed 2.1.0
- `gh release list --repo scottkw/storcat` — v2.x releases not yet published confirmed

### Secondary (MEDIUM confidence)
- Homebrew naming convention (`brew tap <user>/<name>` → `github.com/<user>/homebrew-<name>`) — well-known documented behavior, verified by existing working tap
- WinGet manifest schema 1.6.0 — inferred from existing YAML schema comment URLs

### Tertiary (LOW confidence)
- GitHub archived repo behavior with `brew tap` — archived repos are read-only but still accessible via git clone; Homebrew can tap from them. Not explicitly tested but consistent with GitHub's documented archive behavior.

## Metadata

**Confidence breakdown:**
- Current repo state: HIGH — all files directly fetched via gh API and curl
- WinGet manifest structure: HIGH — verified from existing live manifests
- Homebrew template approach: HIGH — update-tap.sh source code read directly
- Archive command: HIGH — gh CLI help text confirmed
- Stub SHA256 approach: MEDIUM — structurally correct but not tested against winget validation

**Research date:** 2026-03-27
**Valid until:** 2026-04-27 (stable domain — satellite repos won't change without action)
