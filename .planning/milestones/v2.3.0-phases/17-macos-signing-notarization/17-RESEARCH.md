# Phase 17: macOS Signing & Notarization — Research

**Researched:** 2026-03-27
**Domain:** macOS code signing, notarization, Gatekeeper, GitHub Actions CI
**Confidence:** HIGH

## Summary

Phase 17 adds macOS Developer ID signing, notarization, and stapling to the existing `release.yml` `build-macos` job. The goal is that every DMG produced by a CI tag push is accepted by Gatekeeper on macOS 15+ without any user prompt.

The four mandatory signing steps must execute in strict order after `wails build` and before `upload-artifact`: (1) import certificate into an isolated keychain, (2) `codesign` the `.app` bundle with `--options runtime` and an entitlements plist, (3) `xcrun notarytool submit ... --wait` on the DMG, (4) `xcrun stapler staple` the DMG. A `spctl --assess` gate step validates the signed bundle.

The Wails-built StorCat app bundle is exceptionally lean — only three files (`Info.plist`, `Contents/MacOS/StorCat`, `Contents/Resources/iconfile.icns`) — with no nested frameworks, dylibs, or helper executables. This means a single `codesign` invocation on the `.app` bundle is sufficient; no bottom-up multi-step signing is required. The entitlements plist does not exist yet and must be created as part of this phase (SIGN-05).

**Primary recommendation:** Use `apple-actions/import-codesign-certs@v6` (SHA `b610f78`) for keychain management (it handles the `security set-key-partition-list` ACL call that prevents `codesign` hangs), then run `codesign` and `notarytool` as shell steps. No third-party notarization action is needed — `xcrun notarytool` is available on all macOS GitHub-hosted runners with Xcode installed.

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SIGN-01 | `release.yml` `build-macos` job signs .app bundle with `codesign --sign --options runtime` and entitlements | Confirmed: single `codesign` invocation suffices for StorCat's lean bundle structure. Command syntax fully documented below. |
| SIGN-02 | `release.yml` `build-macos` job submits DMG to Apple notarization via `xcrun notarytool` | Confirmed: `xcrun notarytool submit <dmg> --apple-id ... --team-id ... --password ... --wait` pattern documented. Use `store-credentials` keychain profile approach for CI. |
| SIGN-03 | `release.yml` `build-macos` job staples notarization ticket to DMG via `xcrun stapler` | Confirmed: `xcrun stapler staple <dmg>` after notarytool returns "Accepted". Must staple the DMG (not the .app) since DMG is the distribution artifact. |
| SIGN-04 | CI uses isolated temporary keychain, cleaned up after signing | Confirmed: `apple-actions/import-codesign-certs@v6` creates `signing_temp` keychain with auto-generated password and handles import. Manual `security delete-keychain` in a `if: always()` step provides cleanup. |
| SIGN-05 | Entitlements plist ported from Electron era and verified for Wails runtime | No Electron-era entitlements plist exists in the repo. Must create fresh `build/darwin/entitlements.plist`. Wails apps using WebKit require `com.apple.security.cs.allow-jit` and `com.apple.security.cs.allow-unsigned-executable-memory`. Full plist content documented below. |
| SIGN-06 | `spctl --assess` verification step confirms Gatekeeper acceptance | Confirmed: `spctl --assess --type exec build/bin/StorCat.app` returns exit code 0 ("accepted") when signed and notarized. Must run AFTER notarization (before notarization it will fail). Gate step uses `|| exit 1`. |
</phase_requirements>

---

## Standard Stack

### Core
| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| `apple-actions/import-codesign-certs` | v6.0.0 (SHA: `b610f78488812c1e56b20e6df63ec42d833f2d14`) | Import .p12 cert into isolated keychain with correct ACL | Handles `security set-key-partition-list` which prevents `codesign` hanging on CI; GitHub-hosted runners reset state between jobs so isolated keychain is ephemeral |
| `xcrun codesign` | macOS built-in (Xcode) | Sign `.app` bundle with Developer ID and hardened runtime | Apple's own tool — no alternative for notarization-eligible signing |
| `xcrun notarytool` | macOS built-in (Xcode 13+) | Submit to Apple notarization service | Replaced `altool` (decommissioned Nov 2023); the only current tool |
| `xcrun stapler` | macOS built-in (Xcode) | Staple notarization ticket to DMG | Required for offline Gatekeeper validation |
| `spctl` | macOS built-in | Gatekeeper assessment gate | CI verification that signed artifact is accepted |

### Supporting
| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| `security` CLI | macOS built-in | Keychain cleanup (`security delete-keychain`) | In `if: always()` cleanup step after signing |
| `hdiutil` | macOS built-in | DMG creation (already in workflow) | Existing step; no changes needed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `apple-actions/import-codesign-certs@v6` | Manual `security create-keychain` + `security import` shell steps | Manual approach works but requires correct `security set-key-partition-list` invocation to prevent codesign hangs — easy to misconfigure |
| `xcrun notarytool` directly | `GuillaumeFalourd/notary-tools@v1` action | Action is a thin wrapper; direct `xcrun notarytool` gives more control and fewer action dependencies |
| `codesign --deep` flag | Manual bottom-up signing of nested content | `--deep` is unreliable for complex bundles; not needed for StorCat (no nested frameworks) |

**Installation:** No npm/pip packages needed. All tools are macOS built-ins on GitHub-hosted `macos-14` runners.

---

## Architecture Patterns

### Signing Step Sequence (Mandatory Order)

```
wails build → codesign .app → hdiutil (create DMG) → notarytool submit → stapler staple → spctl assess
```

This order cannot be changed:
- `codesign` must happen before DMG creation (sign the app, not the DMG)
- `notarytool` requires the DMG to exist
- `stapler` requires successful notarization
- `spctl` requires stapled DMG

### Recommended Project Structure

```
build/
├── darwin/
│   ├── entitlements.plist    # NEW: created in this phase (SIGN-05)
│   ├── Info.dev.plist        # existing Wails template
│   └── Info.plist            # existing Wails template
└── bin/
    └── StorCat.app/          # Wails output (3 files only)
```

### Pattern 1: Certificate Import with Isolated Keychain

**What:** Use `apple-actions/import-codesign-certs` to import the .p12, then clean up the keychain in an `if: always()` step.

**When to use:** Every CI signing job — ensures the keychain never persists to the next run.

```yaml
# Source: https://github.com/Apple-Actions/import-codesign-certs (v6.0.0)
- name: Import Apple certificate
  uses: apple-actions/import-codesign-certs@b610f78488812c1e56b20e6df63ec42d833f2d14
  with:
    p12-file-base64: ${{ secrets.APPLE_CERTIFICATE }}
    p12-password: ${{ secrets.APPLE_CERTIFICATE_PASSWORD }}

# ... signing steps ...

- name: Clean up keychain
  if: always()
  run: security delete-keychain signing_temp.keychain-db
```

### Pattern 2: Sign .app with Hardened Runtime + Entitlements

**What:** One `codesign` invocation signs the entire `.app` bundle.

**When to use:** After `wails build`, before DMG creation.

```bash
# Source: Apple codesign documentation + verified pattern
/usr/bin/codesign \
  --sign "${{ secrets.APPLE_CERTIFICATE_NAME }}" \
  --options runtime \
  --entitlements build/darwin/entitlements.plist \
  --force \
  --timestamp \
  --verbose \
  build/bin/StorCat.app
```

Note: No `--deep` flag. StorCat has no nested frameworks or helpers — `--deep` is not needed and can cause issues with complex bundles (not applicable here, but avoid as a convention).

### Pattern 3: Notarize DMG via notarytool

**What:** Store credentials in keychain profile, submit DMG, wait for result.

**When to use:** After DMG creation.

```bash
# Source: Official Apple notarytool documentation + federicoterzi.com verified pattern
xcrun notarytool store-credentials "notarytool-profile" \
  --apple-id "${{ secrets.APPLE_NOTARIZATION_APPLE_ID }}" \
  --team-id "${{ secrets.APPLE_TEAM_ID }}" \
  --password "${{ secrets.APPLE_NOTARIZATION_PASSWORD }}"

xcrun notarytool submit \
  "build/bin/StorCat-${{ steps.version.outputs.VERSION }}-darwin-universal.dmg" \
  --keychain-profile "notarytool-profile" \
  --wait
```

**Note on APPLE_NOTARIZATION_APPLE_ID:** Phase 16 research documented `APPLE_NOTARIZATION_PASSWORD` (app-specific password) and `APPLE_TEAM_ID`, but `notarytool` also requires the Apple ID email (`--apple-id`). The Phase 16 secrets list did not include an explicit `APPLE_NOTARIZATION_APPLE_ID` secret — it was assumed to be embedded. The plan must add this secret or inline the Apple ID as a CI variable.

### Pattern 4: Staple and Verify

**What:** Staple ticket to the DMG, then assess with spctl.

**When to use:** After successful notarization.

```bash
# Source: Apple xcrun documentation + gist.github.com/rsms verified pattern
xcrun stapler staple \
  "build/bin/StorCat-${{ steps.version.outputs.VERSION }}-darwin-universal.dmg"

# Gate step: fails CI if not accepted
spctl --assess --type exec build/bin/StorCat.app || exit 1
```

**Note on spctl target:** `spctl --assess --type exec` operates on the `.app` bundle, not the DMG. The DMG is assessed with `--type open`. Use the `.app` for exec assessment per SIGN-06 requirements.

### Anti-Patterns to Avoid

- **Using `--deep` on codesign:** Not needed for StorCat (no nested content), and unreliable for bundles with frameworks.
- **Signing the DMG before notarization:** DMG signature is not required; signing the `.app` inside is what matters.
- **Running `spctl --assess` before notarization:** Will return "rejected" even for correctly signed apps — spctl checks notarization status.
- **Notarizing the .app directly:** Apple accepts .app in a zip, but the DMG is already available — submit the DMG directly.
- **Using `altool` or `gon`:** Both are deprecated/dead. `altool` was decommissioned November 2023. `gon` depended on `altool`.
- **Storing credentials in-line:** Always use `notarytool store-credentials` keychain profile approach for CI to avoid credential exposure in logs.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Keychain ACL setup | Custom `security set-key-partition-list` shell script | `apple-actions/import-codesign-certs@v6` | ACL setup is easy to get wrong — missing it causes `codesign` to hang indefinitely waiting for keychain permission prompts |
| Notarization polling | Custom polling loop with `notarytool history` | `xcrun notarytool submit --wait` | `--wait` handles polling internally with correct backoff; Apple's server can take 1-15 minutes |
| DMG creation | Custom dmgcanvas or create-dmg | `hdiutil` (already in workflow) | Already working; no reason to change |

**Key insight:** The four Apple tools (`codesign`, `notarytool`, `stapler`, `spctl`) are the canonical path. Third-party wrappers add complexity without solving anything for a simple Go/Wails binary with no nested frameworks.

---

## Entitlements Plist (SIGN-05)

### Wails Runtime Entitlement Requirements

Wails uses WebKit (WKWebView) which requires JIT compilation for JavaScript. Without `com.apple.security.cs.allow-jit`, the WebKit renderer cannot execute JavaScript, breaking the entire UI.

The Electron-era entitlements plist (referenced in CLAUDE.md as `build/entitlements.mac.plist`) does not exist in the Go/Wails repo. It must be created fresh.

**Required entitlements for StorCat (Wails v2 + WebKit):**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <!-- Required: WebKit uses JIT for JS execution -->
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <!-- Required: WebKit needs writable+executable memory -->
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <!-- Required: StorCat reads arbitrary user-selected directories -->
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
    <!-- Required: network for Apple notarization validation (Gatekeeper) -->
    <key>com.apple.security.network.client</key>
    <true/>
</dict>
</plist>
```

**Entitlements NOT included (and why):**

| Entitlement | Reason excluded |
|-------------|-----------------|
| `com.apple.security.app-sandbox` | StorCat is NOT sandboxed — requires access to arbitrary filesystem paths chosen by user. Sandbox would require entitlement exceptions for every path. |
| `com.apple.security.cs.disable-library-validation` | Not needed — StorCat loads no third-party dylibs dynamically. |
| `com.apple.security.network.server` | StorCat does not run a local server. |
| `com.apple.security.files.downloads.read-write` | Not needed; `user-selected.read-write` covers user-chosen directories. |

**File location:** `build/darwin/entitlements.plist` (matches Wails Mac App Store guide convention for the `build/darwin/` directory).

**Confidence:** MEDIUM — `allow-jit` and `allow-unsigned-executable-memory` are confirmed required for WebKit/WKWebView. The exact minimal set for StorCat's filesystem access pattern may need runtime verification. If notarization fails with an entitlement error, the log from `xcrun notarytool log` will identify which entitlements are missing or disallowed.

---

## Common Pitfalls

### Pitfall 1: Missing Apple ID Secret for notarytool

**What goes wrong:** `xcrun notarytool store-credentials` requires three inputs: `--apple-id`, `--team-id`, `--password`. Phase 16 documented secrets as `APPLE_NOTARIZATION_PASSWORD` and `APPLE_TEAM_ID`, but did not explicitly name a secret for the Apple ID email address.
**Why it happens:** Phase 16 research assumed the Apple ID was known and might be hardcoded or named differently.
**How to avoid:** Add `APPLE_ID` (the Apple account email) as a 10th secret in the `release` environment, OR confirm the Apple ID (`kenscott@gmail.com` or similar) can be hardcoded in the workflow (not a secret — it's not sensitive). Plan must address this gap.
**Warning signs:** `notarytool store-credentials` fails with "invalid arguments" because `--apple-id` is missing.

### Pitfall 2: codesign Hangs Without Keychain ACL

**What goes wrong:** `codesign` hangs indefinitely waiting for a user to grant keychain access via a macOS security dialog — which never appears in CI.
**Why it happens:** After `security import`, the key is imported but `codesign` is not yet on the ACL for that key. The `security set-key-partition-list` command must be run to grant `codesign` access.
**How to avoid:** Use `apple-actions/import-codesign-certs@v6` which runs this step automatically. If using manual keychain setup, run: `security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$KEYCHAIN_PASSWORD" signing_temp.keychain-db`
**Warning signs:** The `codesign` step runs but never completes; job times out.

### Pitfall 3: notarytool submit Returns Status 0 on Rejection

**What goes wrong:** `xcrun notarytool submit --wait` exits with status 0 even when Apple rejects the submission. The failure is in the stdout output, not the exit code.
**Why it happens:** `notarytool` treats "request submitted successfully" as success at the process level; the notarization result (Accepted/Invalid) is in the text output.
**How to avoid:** After `notarytool submit --wait`, grep the output for "status: Accepted" OR use `notarytool log <submission-id>` to retrieve the full log when status is not Accepted. Add: `xcrun notarytool submit ... --wait | tee /tmp/notary.log && grep -q "status: Accepted" /tmp/notary.log`
**Warning signs:** CI step succeeds (green) but `spctl --assess` later returns "rejected."

### Pitfall 4: stapling DMG Before .app Is Signed

**What goes wrong:** `xcrun stapler staple` on a DMG that was created from an unsigned `.app` results in a DMG that passes notarization but Gatekeeper rejects the `.app` inside.
**Why it happens:** The DMG wraps the `.app`; if the `.app` was unsigned when packed into the DMG, the DMG contains an unsigned binary. Notarization validates the DMG, but Gatekeeper evaluates the extracted `.app`.
**How to avoid:** Always sign `.app` before calling `hdiutil create`. The current workflow must be reordered: codesign → hdiutil → notarytool → stapler.
**Warning signs:** `spctl --assess --type exec StorCat.app` fails after extraction from a notarized DMG.

### Pitfall 5: spctl Assessment Before Notarization

**What goes wrong:** Running `spctl --assess --type exec StorCat.app` as a gate step immediately after codesign (before notarization) returns "rejected" even for a correctly-signed app.
**Why it happens:** `spctl` checks Apple's notarization database. A freshly-signed but not-yet-notarized app is not in the database.
**How to avoid:** Place the `spctl --assess` step AFTER `xcrun stapler staple` completes. The stapled ticket embeds the notarization result and allows offline Gatekeeper verification.
**Warning signs:** `spctl` step fails immediately after codesign; notarization was never attempted.

### Pitfall 6: entitlements.plist Missing → notarytool Rejection

**What goes wrong:** `xcrun notarytool submit` returns "Invalid" with an error about unsigned executable memory or JIT not being allowed.
**Why it happens:** Wails apps using WKWebView require `com.apple.security.cs.allow-jit`. Without the entitlements plist, these capabilities are not granted, and the hardened runtime blocks WebKit.
**How to avoid:** Create `build/darwin/entitlements.plist` with the required entitlements before running `codesign`. Use the plist content documented in the Entitlements section above.
**Warning signs:** Notarization returns "Invalid"; `xcrun notarytool log <id>` shows "The executable ... was not signed with the required entitlements."

### Pitfall 7: environment: release Missing from build-macos Job

**What goes wrong:** The signing secrets (`APPLE_CERTIFICATE`, etc.) are stored in the `release` GitHub Actions environment, not as repo-level secrets. If the `build-macos` job does not declare `environment: release`, secrets are not exposed.
**Why it happens:** Phase 16 stored secrets in the `release` environment with deployment protection. Jobs must explicitly opt-in to environment secrets.
**How to avoid:** Add `environment: release` to the `build-macos` job definition in `release.yml`.
**Warning signs:** `APPLE_CERTIFICATE` is empty/null; `apple-actions/import-codesign-certs` fails with "no certificate provided."

---

## Code Examples

Verified patterns from official and cross-referenced sources:

### Complete Signing Block (build-macos job additions)

```yaml
# Source: apple-actions/import-codesign-certs@v6 README + federicoterzi.com + Apple docs
- name: Import Apple certificate
  uses: apple-actions/import-codesign-certs@b610f78488812c1e56b20e6df63ec42d833f2d14  # v6.0.0
  with:
    p12-file-base64: ${{ secrets.APPLE_CERTIFICATE }}
    p12-password: ${{ secrets.APPLE_CERTIFICATE_PASSWORD }}

- name: Sign .app bundle
  run: |
    /usr/bin/codesign \
      --sign "${{ secrets.APPLE_CERTIFICATE_NAME }}" \
      --options runtime \
      --entitlements build/darwin/entitlements.plist \
      --force \
      --timestamp \
      --verbose \
      build/bin/StorCat.app
```

### Notarization Block (after hdiutil DMG creation)

```yaml
# Source: xcrun notarytool documentation + federicoterzi.com verified pattern
- name: Notarize DMG
  env:
    APPLE_NOTARIZATION_APPLE_ID: ${{ secrets.APPLE_ID }}
    APPLE_NOTARIZATION_PASSWORD: ${{ secrets.APPLE_NOTARIZATION_PASSWORD }}
    APPLE_TEAM_ID: ${{ secrets.APPLE_TEAM_ID }}
  run: |
    xcrun notarytool store-credentials "notarytool-profile" \
      --apple-id "$APPLE_NOTARIZATION_APPLE_ID" \
      --team-id "$APPLE_TEAM_ID" \
      --password "$APPLE_NOTARIZATION_PASSWORD"

    xcrun notarytool submit \
      "build/bin/StorCat-${{ steps.version.outputs.VERSION }}-darwin-universal.dmg" \
      --keychain-profile "notarytool-profile" \
      --wait \
      | tee /tmp/notary-output.log

    grep -q "status: Accepted" /tmp/notary-output.log || {
      echo "Notarization failed. Fetching log..."
      SUBMISSION_ID=$(grep "  id: " /tmp/notary-output.log | head -1 | awk '{print $2}')
      xcrun notarytool log "$SUBMISSION_ID" --keychain-profile "notarytool-profile"
      exit 1
    }
```

### Staple and Verify Block

```yaml
# Source: Apple xcrun stapler documentation + gist.github.com/rsms
- name: Staple notarization ticket
  run: |
    xcrun stapler staple \
      "build/bin/StorCat-${{ steps.version.outputs.VERSION }}-darwin-universal.dmg"

- name: Verify Gatekeeper acceptance
  run: spctl --assess --type exec build/bin/StorCat.app || exit 1

- name: Clean up keychain
  if: always()
  run: security delete-keychain signing_temp.keychain-db 2>/dev/null || true
```

### entitlements.plist (SIGN-05 artifact)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.cs.allow-jit</key>
    <true/>
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
    <key>com.apple.security.network.client</key>
    <true/>
</dict>
</plist>
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `xcrun altool --notarize-app` | `xcrun notarytool submit` | November 2023 (altool decommissioned) | Must use notarytool; altool calls fail |
| `gon` tool for sign + notarize | Direct `codesign` + `notarytool` | 2023 (gon depended on altool) | gon is unusable; use Apple tools directly |
| `--deep` flag for recursive signing | Sign bottom-up or per-component | Community consensus 2022+ | `--deep` unreliable for frameworks; not needed for simple bundles |
| App-specific password in notarytool | App Store Connect API key (SIGN-07) | Available now; SIGN-07 is post-v2.3.x | App-specific password is valid for v2.3.0; API key is more robust |
| `dlemstra/code-sign-action` | Archived; use eSigner or direct tools | October 2025 | Do not reference; out of scope for macOS |

**Deprecated/outdated:**
- `altool`: Decommissioned November 2023 — any workflow using it will fail
- `gon`: Unmaintained; depended on altool — do not use
- `--deep` on codesign: Works for simple bundles (StorCat) but is not recommended practice

---

## Open Questions

1. **Apple ID secret for notarytool**
   - What we know: `notarytool store-credentials` requires `--apple-id` (the Apple account email). Phase 16 provisioned 5 Apple secrets but did not include an explicit `APPLE_ID` secret for the email.
   - What's unclear: Was the Apple ID intended to be hardcoded in the workflow, or stored as a secret? The email is not sensitive — it can be hardcoded.
   - Recommendation: Plan should add `APPLE_ID` as a 10th secret in the `release` environment, OR the plan task should document that the Apple ID email is hardcoded in the workflow. Either is acceptable; confirm with user which approach to use.

2. **Entitlements plist — minimal set correctness**
   - What we know: `com.apple.security.cs.allow-jit` is required for WebKit/WKWebView. `user-selected.read-write` covers StorCat's filesystem access pattern.
   - What's unclear: Whether Wails v2.10.2's specific WebKit embedding requires any additional entitlements not covered by the plist above.
   - Recommendation: Use the plist documented above. If notarization fails with entitlement errors, `xcrun notarytool log` will identify the specific gap. Include a verification task in the plan that checks notarization output.

3. **Verify codesign after signing**
   - What we know: `codesign --verify --verbose build/bin/StorCat.app` can confirm signing before DMG creation.
   - What's unclear: Whether adding an explicit verify step is worth the time vs. letting `notarytool` surface issues.
   - Recommendation: Include `codesign --verify --verbose build/bin/StorCat.app` as a step between signing and DMG creation. Fails fast and reduces debug loop time.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `codesign` | SIGN-01 | Yes (macos-14 runner) | Xcode built-in | — |
| `xcrun notarytool` | SIGN-02 | Yes (Xcode 13+, macos-14 runner) | Current | — |
| `xcrun stapler` | SIGN-03 | Yes (macos-14 runner) | Current | — |
| `spctl` | SIGN-06 | Yes (macos-14 runner) | macOS built-in | — |
| `apple-actions/import-codesign-certs` | SIGN-04 | Yes (GitHub Marketplace) | v6.0.0 / SHA b610f78 | Manual `security` shell steps (more error-prone) |
| `APPLE_CERTIFICATE` GitHub secret | SIGN-04 | Yes (Phase 16 complete: CRED-05 done) | — | — |
| `APPLE_CERTIFICATE_NAME` secret | SIGN-01 | Yes | — | — |
| `APPLE_NOTARIZATION_PASSWORD` secret | SIGN-02 | Yes | — | — |
| `APPLE_TEAM_ID` secret | SIGN-02 | Yes | — | — |
| `APPLE_ID` secret (Apple account email) | SIGN-02 | Unknown — not explicitly in Phase 16 secrets list | — | Hardcode in workflow (not sensitive) |
| `build/darwin/entitlements.plist` | SIGN-05 | Not present — must be created | — | — |

**Missing dependencies with no fallback:**
- `build/darwin/entitlements.plist` — must be created as part of this phase (Wave 0 / task 1)

**Missing dependencies with fallback:**
- `APPLE_ID` secret — can be hardcoded in workflow if not added as a secret

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | None (CI workflow verification — no unit test framework applicable) |
| Config file | `.github/workflows/release.yml` (the artifact under test) |
| Quick run command | `codesign --verify --verbose build/bin/StorCat.app` (local smoke check) |
| Full suite command | Push a `v*.*.*` tag to trigger full CI pipeline |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SIGN-01 | .app signed with runtime entitlements | smoke | `codesign --verify --verbose build/bin/StorCat.app && codesign -d --entitlements :- build/bin/StorCat.app` | N/A — CI step |
| SIGN-02 | DMG passes Apple notarization | integration | `xcrun notarytool submit ... --wait` in CI | N/A — CI step |
| SIGN-03 | Notarization ticket stapled to DMG | smoke | `xcrun stapler validate build/bin/StorCat-*.dmg` | N/A — CI step |
| SIGN-04 | Isolated keychain cleaned up | integration | `security list-keychains` after signing — `signing_temp` should be absent | N/A — CI step |
| SIGN-05 | entitlements.plist exists with correct content | smoke | `ls build/darwin/entitlements.plist && plutil -lint build/darwin/entitlements.plist` | ❌ Wave 0 |
| SIGN-06 | spctl assessment returns "accepted" | smoke | `spctl --assess --type exec build/bin/StorCat.app` | N/A — CI step |

### Sampling Rate

- **Per task commit:** Lint the YAML with `yamllint .github/workflows/release.yml`; verify `entitlements.plist` with `plutil -lint`
- **Per wave merge:** Full CI run via tag push on a test tag (e.g., `v99.0.0-test`) — then delete the release
- **Phase gate:** Full CI green + DMG opens on macOS 15 without Gatekeeper prompt (UAT)

### Wave 0 Gaps

- [ ] `build/darwin/entitlements.plist` — covers SIGN-05; must exist before signing step can run

*(No test framework installation needed — this is a CI workflow + file creation phase)*

---

## Project Constraints (from CLAUDE.md)

| Directive | Impact on This Phase |
|-----------|---------------------|
| GSD workflow enforcement: use `/gsd:execute-phase` | All tasks executed via GSD workflow |
| No Electron dependencies: use Go/Wails patterns | Entitlements plist is Wails-native; no Electron-builder config involved |
| GitHub Actions: pin actions to SHAs | `apple-actions/import-codesign-certs` must use pinned SHA (`b610f78`) — already documented |
| Branch: `feature/go-refactor-2.0.0-clean`, merge to `main` | Verify current branch before editing `release.yml` |
| `pnpm` preferred for Node | Not applicable to this phase |

---

## Sources

### Primary (HIGH confidence)
- `apple-actions/import-codesign-certs` releases page — v6.0.0 SHA `b610f78488812c1e56b20e6df63ec42d833f2d14` confirmed, December 2, 2024
- [GitHub Docs: Installing an Apple certificate on macOS runners](https://docs.github.com/en/actions/use-cases-and-examples/deploying/installing-an-apple-certificate-on-macos-runners-for-xcode-development) — official keychain workflow pattern
- [GitHub Marketplace: Import Code-Signing Certificates](https://github.com/marketplace/actions/import-code-signing-certificates) — confirmed input names: `p12-file-base64`, `p12-password`
- `find /Users/ken/dev/storcat/build/bin/StorCat.app -type f` — confirmed 3-file bundle structure (no nested frameworks)
- Wails `wails.json` — confirmed `com.wails.StorCat` bundle ID, v2.10.2 targeted

### Secondary (MEDIUM confidence)
- [Federico Terzi: Automatic Code-signing and Notarization for macOS apps using GitHub Actions](https://federicoterzi.com/blog/automatic-code-signing-and-notarization-for-macos-apps-using-github-actions/) — complete workflow verified against official docs; notarytool pattern with `store-credentials` + `--keychain-profile`
- [rsms macOS distribution gist](https://gist.github.com/rsms/929c9c2fec231f0cf843a1a746a416f5) — bottom-up signing, `ditto` for archive, `spctl` assessment
- [tonygo.tech: Complete Guide to Notarizing macOS Apps with notarytool](https://tonygo.tech/blog/2023/notarization-for-macos-app-with-notarytool) — `xcrun notarytool submit ... --wait` pattern; spctl assessment flags
- [DoltHub: How to Publish a Mac Desktop App Outside the App Store](https://www.dolthub.com/blog/2024-10-22-how-to-publish-a-mac-desktop-app-outside-the-app-store/) — entitlements plist content for non-sandboxed apps with filesystem access

### Tertiary (LOW confidence)
- WebSearch results on Wails + entitlements: Wails documentation (403 on direct fetch) is confirmed to require `allow-jit` for WKWebView. Indirect confirmation via multiple community sources and Apple Developer Forums. Exact minimal entitlement set may need runtime validation.

---

## Metadata

**Confidence breakdown:**
- Standard stack (tools + SHA): HIGH — all tools are macOS built-ins; SHA confirmed from releases page
- Signing workflow pattern: HIGH — verified against GitHub official docs + cross-referenced with three independent sources
- Entitlements plist content: MEDIUM — `allow-jit` confirmed required for WebKit; `user-selected.read-write` assumed correct for filesystem access; may need adjustment after first notarization attempt
- Pitfalls: HIGH — all pitfalls sourced from official docs or well-known community patterns
- Notarytool rejection detection: MEDIUM — exit code 0 on rejection is documented in multiple places; `grep "status: Accepted"` pattern is practical but should be validated

**Research date:** 2026-03-27
**Valid until:** 2026-06-27 (Apple toolchain is stable; GitHub Actions runner image updates are infrequent for macOS tooling)
