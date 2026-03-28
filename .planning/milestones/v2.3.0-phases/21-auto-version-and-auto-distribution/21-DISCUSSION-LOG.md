# Phase 21: Auto Version and Auto Distribution - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-28
**Phase:** 21-auto-version-and-auto-distribution
**Areas discussed:** Version bump strategy, Release trigger flow, Version file sync, Draft vs auto-publish

---

## Version Bump Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Conventional Commits | Commit messages like feat:, fix:, breaking: auto-determine major/minor/patch. Tools like release-please or semantic-release parse git log. | ✓ |
| Explicit flag on release | Developer chooses patch/minor/major when triggering the release (e.g., workflow_dispatch with dropdown). Simple and predictable. | |
| Always patch, manual override | Default to patch bump on every release. Developer manually sets minor/major when needed via commit tag or input. | |

**User's choice:** Conventional Commits
**Notes:** None

### Tool Selection

| Option | Description | Selected |
|--------|-------------|----------|
| release-please | Google's tool. Creates a release PR that accumulates changes, auto-generates CHANGELOG, bumps version. Well-suited for Go projects. | ✓ |
| semantic-release | Node.js ecosystem tool. Fully automated, requires Node in CI even for Go projects. | |
| Custom script | Shell/Go script that parses conventional commits and bumps version. Full control, more to maintain. | |

**User's choice:** release-please
**Notes:** None

### Changelog

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, auto-generate | release-please maintains CHANGELOG.md automatically from conventional commits. | ✓ |
| No, rely on GitHub release notes | release.yml already uses generate_release_notes. Skip CHANGELOG.md to avoid duplication. | |

**User's choice:** Yes, auto-generate
**Notes:** None

---

## Release Trigger Flow

| Option | Description | Selected |
|--------|-------------|----------|
| Merge release-please PR | release-please auto-creates a PR when releasable commits land on main. Merging that PR creates the git tag and GitHub release. | ✓ |
| workflow_dispatch manual trigger | Keep release-please for version calculation, but require a manual button press to cut the release. | |
| Auto-release on main merge | Every merge to main that has releasable commits auto-tags and releases. Fully hands-off. | |

**User's choice:** Merge release-please PR
**Notes:** None

### Branch Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| main only | release-please watches main branch. Standard setup for solo developer. | ✓ |
| main + release branches | Support release branches for hotfixes. More complex. | |

**User's choice:** main only
**Notes:** User mentioned splash screen version must auto-update — verified it already reads dynamically from Go backend.

---

## Version File Sync

| Option | Description | Selected |
|--------|-------------|----------|
| wails.json as single source | release-please updates wails.json only. frontend/package.json version is cosmetic. The Go binary embeds wails.json, so it's the real source of truth. | ✓ |
| release-please updates both | Configure release-please extra-files to bump both. Keeps them in sync but frontend version is unused at runtime. | |
| Remove version from package.json | Set frontend/package.json version to "0.0.0" and ignore it. | |

**User's choice:** wails.json as single source
**Notes:** None

---

## Draft vs Auto-Publish

| Option | Description | Selected |
|--------|-------------|----------|
| release-please creates release, release.yml builds | release-please creates the GitHub release. release.yml triggers on tag push, builds artifacts, uploads to existing release. distribute.yml triggers on release published. Seamless chain. | ✓ |
| Keep draft for review | release-please creates tag only. release.yml builds and creates a draft release. Manual publish after reviewing artifacts. | |
| Full auto-publish | Fully hands-off after PR merge. | |

**User's choice:** release-please creates release, release.yml builds
**Notes:** None

### Distribution Trigger

| Option | Description | Selected |
|--------|-------------|----------|
| Auto-publish after artifact upload | release.yml uploads artifacts and marks the release as published. distribute.yml triggers automatically. | ✓ |
| Manual publish after artifact review | release.yml uploads artifacts but keeps release as draft. Manual publish step. | |

**User's choice:** Auto-publish after artifact upload
**Notes:** None

## Claude's Discretion

- release-please configuration details (release-type, extra-files, versioning-strategy)
- How release.yml is refactored to upload to existing release instead of creating one
- Whether to use release-please GitHub Action or standalone CLI
- CHANGELOG.md formatting and sections

## Deferred Ideas

None — discussion stayed within phase scope
