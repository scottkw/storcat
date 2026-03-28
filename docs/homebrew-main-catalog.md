# Submitting StorCat to the Main Homebrew Catalog

This document covers the full process for getting StorCat accepted into `Homebrew/homebrew-cask` (the main catalog) so users can install with just `brew install --cask storcat` — no tap required.

## Current State

StorCat is distributed via a custom tap:
```bash
brew tap scottkw/storcat
brew install --cask storcat
```

The goal is to reach:
```bash
brew install --cask storcat
```

## Notability Requirements

Homebrew-cask enforces minimum GitHub metrics before accepting new casks. These thresholds are non-negotiable and checked during PR review.

### Thresholds

| Submitter Type | Stars | Forks | Watchers | All Required? |
|----------------|-------|-------|----------|---------------|
| **Third-party** (someone else submits) | 75+ | 30+ | 30+ | Yes, all three |
| **Self-submitted** (repo owner submits) | 225+ | 90+ | 90+ | Yes, all three |

**StorCat's current metrics: 0 stars, 0 forks, 0 watchers.**

### Exceptions (Rare)

Homebrew maintainers may waive notability for:
- Apps with their own website where GitHub is only used for binary hosting (misleading repo stats)
- Software submitted by a Homebrew maintainer or prolific contributor
- Apps that have gone viral (Hacker News front page, trending on Twitter/X, multiple premature third-party submissions)

These exceptions are not guaranteed and require maintainer discretion.

## Meeting Notability: Strategy

The easiest path is the **third-party threshold (75 stars, 30 forks, 30 watchers)** since it's roughly 3x lower than self-submission.

### High-Impact Actions

**1. Hacker News "Show HN" Post**

The Electron-to-Go/Wails migration story is the strongest hook. The HN audience loves:
- Concrete before/after performance numbers (93% smaller, 5x faster, 80% faster startup)
- Electron replacement stories
- Go projects
- Tools they can actually use

Suggested title: `Show HN: StorCat – I rewrote my Electron app in Go/Wails (150MB → 10MB)`

Best posting times: Tuesday-Thursday, 8-10am EST.

A successful Show HN post can generate 50-200+ stars in a single day.

**2. Reddit Posts**

Target subreddits with relevant audiences:

| Subreddit | Angle |
|-----------|-------|
| r/golang | Go/Wails desktop app with CLI, stdlib-only CLI framework |
| r/commandline | CLI catalog tool, `storcat create/search/list/show/open` |
| r/selfhosted | Catalog external drives and network storage |
| r/DataHoarder | Catalog large media collections across drives |
| r/electronjs | "I migrated from Electron to Go/Wails — here's what changed" |
| r/programming | General interest in the migration story |

**3. Blog Post: "Why I Replaced Electron with Go/Wails"**

Publish on personal blog, Medium, or dev.to. Include:
- The problem (Electron bloat, V8/ARM64 issues, memory usage)
- Why Go/Wails (native webview, small binaries, fast startup)
- Concrete metrics (bundle size, memory, startup time, search speed)
- Architecture comparison (IPC vs Wails bindings)
- What stayed the same (React frontend, Ant Design, catalog format)
- Lessons learned

Cross-post the link to HN, Reddit, Twitter/X, and Lobsters.

**4. Migration Case Study for Wails**

Reach out to the Wails project to feature StorCat as a migration case study. The Wails team actively promotes apps built with their framework. This gets exposure to their audience and a backlink from wails.io.

### Lower-Effort Actions

| Action | Expected Impact |
|--------|----------------|
| Add a demo GIF to the README | Makes the repo more shareable |
| Add GitHub topics: `go`, `wails`, `desktop-app`, `catalog`, `electron-alternative` | Improves discoverability in GitHub search |
| List on [awesome-go](https://github.com/avelino/awesome-go) | Steady trickle of stars from curated list visitors |
| List on [awesome-wails](https://github.com/nickarellano/awesome-wails) | Targeted Wails community exposure |
| Star/watch your own repo | Does not count toward thresholds (GitHub filters owner activity) |

### Timeline Estimates

| Scenario | Time to 75 Stars |
|----------|-----------------|
| Successful Show HN (front page) | 1-3 days |
| Viral Reddit post | 1-7 days |
| Blog post shared widely | 1-2 weeks |
| Organic growth from awesome lists + topics | 2-6 months |
| No promotion | Unlikely to reach threshold |

## Technical Preparation

Everything below can be done now, before meeting the notability threshold.

### Cask Formula for Main Repo

The cask format is identical to your tap, but the main repo requires additional stanzas. Here's what StorCat's cask would look like:

```ruby
cask "storcat" do
  version "2.3.0"
  sha256 "ACTUAL_SHA256_HERE"

  url "https://github.com/scottkw/storcat/releases/download/v#{version}/StorCat-v#{version}-darwin-universal.dmg",
      verified: "github.com/scottkw/storcat/"
  name "StorCat"
  desc "Create, browse, and search directory catalogs"
  homepage "https://github.com/scottkw/storcat"

  livecheck do
    url :url
    strategy :github_latest
  end

  depends_on macos: ">= :catalina"

  app "StorCat.app"
  binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"

  zap trash: [
    "~/Library/Application Support/storcat",
    "~/Library/Caches/com.kenscott.storcat",
    "~/Library/Preferences/com.kenscott.storcat.plist",
    "~/Library/Saved Application State/com.kenscott.storcat.savedState",
  ]
end
```

**Key differences from your tap cask:**
- `verified:` parameter on the URL (required for GitHub-hosted downloads)
- `livecheck` block (enables Homebrew's autobump bot to detect new releases)
- `zap` stanza (cleanup paths for complete uninstall — verify these paths are correct)
- `depends_on macos:` (minimum macOS version)
- `desc` must be concise (under 80 characters, no "A" or "An" prefix)

### Pre-Submission Audit Commands

Run these locally before opening the PR:

```bash
# Create the cask file locally for testing
brew tap --force homebrew/cask
cp storcat.rb "$(brew --repository homebrew/cask)/Casks/s/storcat.rb"

# Audit for new cask (strictest checks)
brew audit --cask --new storcat

# Style check and auto-fix
brew style --fix "$(brew --repository homebrew/cask)/Casks/s/storcat.rb"

# Full online audit
brew audit --cask --online storcat

# Test install (bypasses API cache to test the actual cask file)
HOMEBREW_NO_INSTALL_FROM_API=1 brew install --cask storcat

# Test uninstall
brew uninstall --cask storcat
```

All commands must pass with zero errors before submitting.

### Verify zap Paths

The `zap` stanza must list every file/directory StorCat writes outside the app bundle. To find them:

```bash
# Install StorCat, run it, use all features, then check:
find ~/Library -name "*storcat*" -o -name "*StorCat*" -o -name "*com.kenscott.storcat*" 2>/dev/null
```

Incorrect `zap` paths are a common rejection reason, especially if AI-generated (reviewers check this).

## Submission Process

### Step 1: Fork and Branch

```bash
# Fork Homebrew/homebrew-cask on GitHub, then:
git clone https://github.com/YOUR_FORK/homebrew-cask.git
cd homebrew-cask
git checkout -b storcat-new-cask
```

### Step 2: Add Cask File

```bash
# Create the cask file
cp storcat.rb Casks/s/storcat.rb

# Commit with required message format
git add Casks/s/storcat.rb
git commit -m "storcat 2.3.0 (new cask)"
git push origin storcat-new-cask
```

The commit message format `<token> <version> (new cask)` is required for new submissions.

### Step 3: Open PR

Open a PR against `Homebrew/homebrew-cask:master` and fill out the template:

**Required checkboxes:**
- [ ] Cask is for a stable version
- [ ] `brew audit --cask --online storcat` is error-free
- [ ] `brew style --fix storcat` reports no offenses
- [ ] Named according to the [token reference](https://docs.brew.sh/Cask-Cookbook#token-reference)
- [ ] Cask was not [already refused](https://github.com/search?q=repo:Homebrew/homebrew-cask+is:closed+is:unmerged+storcat)
- [ ] `brew audit --cask --new storcat` passes
- [ ] `HOMEBREW_NO_INSTALL_FROM_API=1 brew install --cask storcat` works
- [ ] `brew uninstall --cask storcat` works

**AI disclosure (required):**
- [ ] Disclose if AI was used to generate any part of the PR (especially `zap` paths — reviewers verify these manually)

### Step 4: Review

- **Automated CI** runs immediately (audit, style, install tests)
- **Human review** typically takes 3-14 days for new casks
- If changes are requested, push additional commits (do NOT squash/rebase)
- The PR may show as "Closed" instead of "Merged" — this is normal for Homebrew's merge process

## After Acceptance

### Automatic Updates

Once in the main repo, Homebrew's **autobump bot (BrewTestBot)** checks for new releases every 3 hours using the `livecheck` block. When it detects a new GitHub release, it automatically opens a version bump PR. You don't need to do anything.

### Your Tap

Your tap (`scottkw/homebrew-storcat`) keeps working. Options:

1. **Keep both** — users who already tap'd keep working, new users use the main catalog
2. **Add a deprecation notice** to the tap README pointing users to `brew install --cask storcat`
3. **Archive the tap** after confirming the main cask is stable

### Distribution Workflow

After acceptance, the `update-homebrew` job in `distribute.yml` becomes optional for main-catalog users. You may want to keep it running for the tap as a fallback, or remove it and rely entirely on Homebrew's autobump.

## Checklist

### Before You Can Submit
- [ ] Reach 75+ stars (third-party submission) or 225+ (self-submission)
- [ ] Reach 30+ forks (third-party) or 90+ (self-submission)
- [ ] Reach 30+ watchers (third-party) or 90+ (self-submission)

### Technical Readiness (Already Done)
- [x] macOS DMG with stable download URL on GitHub Releases
- [x] Developer ID code signing
- [x] Apple notarization + stapling
- [x] GUI application (not CLI-only)
- [x] No auto-update mechanism that bypasses Homebrew

### Before Submitting the PR
- [ ] Verify `zap` paths by running StorCat and checking `~/Library`
- [ ] Confirm minimum macOS version in `depends_on`
- [ ] Run `brew audit --cask --new storcat` (zero errors)
- [ ] Run `brew style --fix storcat` (zero offenses)
- [ ] Test full install/uninstall cycle
- [ ] Search [closed PRs](https://github.com/search?q=repo:Homebrew/homebrew-cask+is:closed+is:unmerged+storcat) to confirm not already refused
- [ ] Prepare honest AI disclosure for the PR template
