# Phase 19: Homebrew CLI PATH - Research

**Researched:** 2026-03-28
**Domain:** Homebrew Cask `binary` stanza, macOS app bundle structure, unified binary CLI/GUI dispatch
**Confidence:** HIGH

## Summary

Phase 19 is a single-file change: add a `binary` stanza to `packaging/homebrew/storcat.rb.template`. The stanza symlinks `StorCat.app/Contents/MacOS/StorCat` into `$(brew --prefix)/bin/storcat`, which Homebrew manages automatically on install and removes on uninstall.

The StorCat binary already handles the CLI/GUI split correctly: when invoked with a recognized subcommand (`version`, `create`, `search`, etc.), it runs in CLI mode and exits. When invoked with no arguments, it launches the GUI. This was implemented in Phase 8 and is confirmed in `main.go`. No Go code changes are needed for this phase.

The Electron-app `binary` stanza failure mode (helper app crash) does not apply to Wails apps. Wails bundles a single Go binary with no separate helper processes — the binary runs identically whether launched from within the app bundle or via a symlink.

**Primary recommendation:** Add `binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"` to `storcat.rb.template`, then verify `update-tap.sh` and `distribute.yml` preserve this stanza during template rendering.

## Project Constraints (from CLAUDE.md)

- pnpm preferred for Node; `uv`/`pip` for Python; `go mod` for Go
- Python: always use venv, never global installs
- GSD workflow: use `/gsd:execute-phase` for planned phase work — do not make direct repo edits outside GSD unless explicitly asked
- No Electron dependencies — all fixes must use Go/Wails patterns
- Chesterton's Fence: articulate why existing code exists before removing it

## Standard Stack

### Core
| Library / Tool | Version | Purpose | Why Standard |
|----------------|---------|---------|--------------|
| Homebrew Cask `binary` stanza | N/A (DSL keyword) | Symlinks app-bundle binary to `$(brew --prefix)/bin` | Official Homebrew mechanism for CLI tools bundled in `.app` |
| Homebrew `appdir` interpolation | N/A | References `/Applications` or user app dir at install time | Required because install path may vary per user |

### Architecture Facts

- `$(brew --prefix)/bin` is `/opt/homebrew/bin` on Apple Silicon, `/usr/local/bin` on Intel. It is on PATH by default after Homebrew install.
- Homebrew `binary` stanzas create symlinks (not copies). Uninstall removes the symlink cleanly.
- The symlink target uses `#{appdir}` at install time, resolving to the actual installed app location (typically `/Applications`).
- `target:` parameter renames the symlink. Without `target:`, the symlink takes the source filename (`StorCat`). Since we want lowercase `storcat`, `target: "storcat"` is required.

**No installation command needed** — this is a template file edit only.

## Architecture Patterns

### Binary Stanza Syntax (Verified: Homebrew official docs)

```ruby
binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"
```

This places the stanza after the `app` stanza and before `zap`. Full updated template:

```ruby
cask "storcat" do
  version "{{VERSION}}"
  sha256 "{{SHA256}}"

  url "https://github.com/scottkw/storcat/releases/download/v#{version}/StorCat-v#{version}-darwin-universal.dmg"

  name "StorCat"
  desc "Directory Catalog Manager - Create, search, and browse directory catalogs"
  homepage "https://github.com/scottkw/storcat"

  livecheck do
    url :url
    strategy :github_latest
  end

  app "StorCat.app"
  binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"

  zap trash: [
    "~/Library/Application Support/StorCat",
    "~/Library/Preferences/com.kenscott.storcat.plist",
    "~/Library/Saved Application State/com.kenscott.storcat.savedState",
    "~/Library/Caches/com.kenscott.storcat",
  ]
end
```

### How `distribute.yml` Renders the Template

The CI workflow uses `sed` substitution of only `{{VERSION}}` and `{{SHA256}}` tokens:

```bash
sed -e "s/{{VERSION}}/${{ steps.meta.outputs.version }}/g" \
    -e "s/{{SHA256}}/${{ steps.meta.outputs.sha256 }}/g" \
    packaging/homebrew/storcat.rb.template \
    > tap/Casks/storcat.rb
```

The `binary` stanza contains no template tokens, so it passes through untouched. No CI changes are needed.

### How `update-tap.sh` Renders the Template

The legacy `update-tap.sh` script writes the cask directly via heredoc (`cat > "$CASK_FILE" << EOF`) and does NOT read from `storcat.rb.template`. It will miss the `binary` stanza unless updated. However, this script is the manual fallback — `distribute.yml` is the canonical path.

The planner should decide whether to update `update-tap.sh` to match. If left as-is, manual runs of the shell script would produce a cask missing the `binary` stanza.

### Binary Path Inside the App Bundle

Confirmed on this machine:

```
StorCat.app/
  Contents/
    Info.plist
    MacOS/
      StorCat          ← the unified CLI+GUI binary
    Resources/
```

The binary name matches the app name exactly. Case matters on case-sensitive filesystems.

### CLI Dispatch Confirmation (`main.go`)

```go
// CLI dispatch: known subcommand -> CLI mode, exit before GUI
if len(args) > 0 {
    switch args[0] {
    case "version", "create", "search", "list", "show", "open", "help", "--help", "-h":
        os.Exit(cli.Run(args, Version))
    }
}
// GUI mode: no args, or unrecognized args fall through
```

`storcat version` routes to CLI mode. No Go changes needed.

### Anti-Patterns to Avoid

- **Using bare filename without `target:`**: Creates symlink named `StorCat` (uppercase), but users expect `storcat` (lowercase) for a CLI tool. Always use `target: "storcat"`.
- **Linking to a wrapper script instead of the binary**: Not needed here — the Wails binary handles CLI natively. Wrapper scripts add complexity with no benefit.
- **Using absolute path instead of `#{appdir}`**: Hardcoding `/Applications` breaks for users who install to `~/Applications`. Use `#{appdir}`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| PATH registration | Shell profile modification, custom installer script | Homebrew `binary` stanza | Homebrew manages symlink lifecycle (install, uninstall, upgrade) automatically |
| Binary naming/aliasing | Shell alias in docs, separate wrapper script | `target:` parameter in `binary` stanza | Single source of truth in cask formula |

## Common Pitfalls

### Pitfall 1: Electron Helper App Crash
**What goes wrong:** Cask `binary` stanza causes app to crash when launched via symlink because framework cannot find its helper processes.
**Why it happens:** Electron resolves helper processes relative to the symlink path, not the actual binary location inside the bundle.
**How to avoid:** StorCat uses Wails (Go binary, no helper processes). This pitfall does not apply. The `binary` stanza is safe.
**Warning signs:** Would only appear if StorCat were ever migrated back to Electron.

### Pitfall 2: Template Not Passed Through by CI
**What goes wrong:** `sed` substitution accidentally corrupts or strips the `binary` stanza.
**Why it happens:** Only possible if the stanza contained `{{` or `}}` characters matching the token patterns.
**How to avoid:** The `binary` stanza contains no template tokens. `sed` only replaces `{{VERSION}}` and `{{SHA256}}`. Verify final output in `tap/Casks/storcat.rb` after CI runs.
**Warning signs:** `brew audit tap/Casks/storcat.rb` fails with syntax error.

### Pitfall 3: Stanza Missing from `update-tap.sh`
**What goes wrong:** Manual tap updates via `update-tap.sh` produce a cask without the `binary` stanza.
**Why it happens:** The script writes the cask via heredoc, not by reading `storcat.rb.template`.
**How to avoid:** Either update `update-tap.sh` heredoc to include the `binary` stanza, or deprecate the script in favor of the CI workflow.
**Warning signs:** Tap cask has `app` stanza but no `binary` stanza after a manual update run.

### Pitfall 4: Gatekeeper Blocks Symlinked Binary
**What goes wrong:** `storcat version` fails with Gatekeeper error when invoked via the symlink.
**Why it happens:** macOS Gatekeeper checks the code signature at the symlink target. If the `.app` bundle is unsigned, Gatekeeper blocks execution on macOS 15+.
**How to avoid:** Phase 17 (macOS signing) is a prerequisite and is complete. The app bundle is already signed, notarized, and stapled. This phase depends on Phase 17 correctly.
**Warning signs:** `spctl --assess -v /opt/homebrew/bin/storcat` returns "rejected".

### Pitfall 5: `appdir` Resolves Before App Is Installed
**What goes wrong:** `binary` stanza fails with "symlink source not found" during install.
**Why it happens:** Homebrew evaluates the `binary` stanza after installing the `app` stanza. If the path is wrong, the symlink creation fails.
**How to avoid:** Use `#{appdir}/StorCat.app/Contents/MacOS/StorCat` exactly — match app name and binary name case precisely. The binary name is `StorCat` (capital S and C), matching `wails.json` `"outputfilename": "StorCat"`.
**Warning signs:** `brew install --cask storcat` prints "It seems the symlink source ... is not there".

## Code Examples

### Complete Updated Template (Verified Pattern)
```ruby
# Source: https://docs.brew.sh/Cask-Cookbook#stanza-binary
binary "#{appdir}/StorCat.app/Contents/MacOS/StorCat", target: "storcat"
```

### Real-World Precedent (mpv cask)
```ruby
# Source: https://github.com/orgs/Homebrew/discussions/5128
binary "#{appdir}/mpv.app/Contents/MacOS/mpv"
```

### Verifying the Symlink After Install (Manual Test)
```bash
# After brew install --cask storcat:
which storcat          # should be /opt/homebrew/bin/storcat or /usr/local/bin/storcat
storcat version        # should print version string
ls -la $(which storcat) # should show symlink to .../StorCat.app/Contents/MacOS/StorCat
```

### Auditing a Private Tap Cask Locally
```bash
# From tap directory:
brew audit --cask Casks/storcat.rb --strict
brew style Casks/storcat.rb
```

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Homebrew | Manual testing of cask audit | Yes | 5.1.1 | — |
| `brew audit` | Cask syntax validation | Yes | included with Homebrew | — |

No missing dependencies. This phase makes no runtime calls to external services — it is a template file edit.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` + manual Homebrew audit |
| Config file | none (no test config file for cask validation) |
| Quick run command | `brew audit --cask packaging/homebrew/storcat.rb.template --strict` (if tap installed) or `brew style packaging/homebrew/storcat.rb.template` |
| Full suite command | `go test ./...` (unit tests unchanged by this phase) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PKG-01 | `binary` stanza present in template | content inspection | `grep "binary" packaging/homebrew/storcat.rb.template` | Verified after edit |
| PKG-01 | Stanza passes through `sed` unchanged | content inspection | `sed -e "s/{{VERSION}}/2.3.0/g" -e "s/{{SHA256}}/abc123/g" packaging/homebrew/storcat.rb.template \| grep binary` | Verified after edit |
| PKG-03 | `storcat version` works via symlink | manual smoke | Run on a machine with cask installed | Manual only — requires real install |

### Sampling Rate
- **Per task commit:** `grep "binary" packaging/homebrew/storcat.rb.template`
- **Per wave merge:** `brew style packaging/homebrew/storcat.rb.template` (local syntax check)
- **Phase gate:** Manual verification — install cask, run `storcat version` in new terminal

### Wave 0 Gaps
None — existing test infrastructure is unchanged. The `binary` stanza edit requires no new test files. The smoke test (PKG-03) is manual-only because it requires a real cask installation.

## State of the Art

| Old Approach | Current Approach | Impact |
|--------------|------------------|--------|
| Document PATH setup in README | Homebrew `binary` stanza | Zero-config for Homebrew users |
| Separate CLI binary from GUI app | Unified Wails binary with CLI dispatch | Single artifact, simpler distribution |

## Open Questions

1. **Should `update-tap.sh` be updated or deprecated?**
   - What we know: The script writes the cask via heredoc, missing the `binary` stanza
   - What's unclear: Whether the manual update script is still used or superseded by CI
   - Recommendation: Update the heredoc in `update-tap.sh` to match the template (adds `binary` line), low effort, prevents confusion

2. **Should the `binary` stanza be tested via CI?**
   - What we know: `brew audit` can syntax-check a cask file, but requires Homebrew and a tap context
   - What's unclear: Whether the project's CI (ubuntu-22.04 runner) has Homebrew available for audit
   - Recommendation: Add a `brew audit` step to `distribute.yml` after template rendering, using the tap checkout that already exists

## Sources

### Primary (HIGH confidence)
- [Homebrew Cask Cookbook](https://docs.brew.sh/Cask-Cookbook#stanza-binary) — `binary` stanza syntax, `appdir` interpolation, `target:` parameter
- `packaging/homebrew/storcat.rb.template` — existing template structure (read directly)
- `.github/workflows/distribute.yml` — `sed` substitution logic (read directly)
- `main.go` — CLI/GUI dispatch logic (read directly)
- `build/bin/StorCat.app/Contents/MacOS/StorCat` — confirmed binary name and path (verified with `ls`)

### Secondary (MEDIUM confidence)
- [Homebrew/homebrew-cask issue #252423](https://github.com/Homebrew/homebrew-cask/issues/252423) — Electron helper app crash with `binary` stanza; confirmed does not apply to Wails
- [Wails discussions #3098](https://github.com/wailsapp/wails/discussions/3098) — unified binary CLI/GUI approach for Wails distribution
- [Homebrew discussions #5128](https://github.com/orgs/Homebrew/discussions/5128) — mpv cask `binary "#{appdir}/mpv.app/Contents/MacOS/mpv"` real-world precedent

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — official Homebrew docs confirm `binary` stanza syntax; verified `appdir` pattern
- Architecture: HIGH — binary path confirmed via `ls`; CLI dispatch confirmed in `main.go`; `sed` passthrough confirmed by reading `distribute.yml`
- Pitfalls: HIGH — Electron issue documented and confirmed not applicable; Gatekeeper dependency on Phase 17 confirmed in STATE.md

**Research date:** 2026-03-28
**Valid until:** 2026-09-28 (stable Homebrew Cask DSL; unlikely to change)

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PKG-01 | Homebrew cask `binary` stanza puts `storcat` on PATH after `brew install --cask storcat` | `binary` stanza with `target: "storcat"` creates symlink in `$(brew --prefix)/bin` — verified via Homebrew docs and `mpv` cask precedent |
| PKG-03 | `storcat version` works from any new terminal after Homebrew install (macOS) | CLI dispatch in `main.go` routes `version` subcommand to `cli.Run()` before GUI launch; symlink via `binary` stanza makes command available in PATH |
</phase_requirements>
