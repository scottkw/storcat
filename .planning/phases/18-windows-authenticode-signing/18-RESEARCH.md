# Phase 18: Windows Authenticode Signing — Research

**Researched:** 2026-03-28
**Domain:** Windows Authenticode signing, SSL.com eSigner, GitHub Actions CI, signtool.exe
**Confidence:** HIGH (workflow pattern), MEDIUM (eSigner integration — no Windows secrets exist yet to test)

## Summary

Phase 18 adds Windows Authenticode signing to the existing `release.yml` `build-windows` job. The goal is that every NSIS installer and portable .exe produced by a CI tag push is signed before upload, so SmartScreen warnings are suppressed as reputation builds and WinGet SHA256 manifests always reference signed binaries.

The Windows signing approach is locked: SSL.com eSigner OV via the `SSLcom/esigner-codesign` GitHub Action. This uses 4 API secrets (ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET) stored in the `release` GitHub Actions environment. As of Phase 16's Phase 03 plan execution, these 4 secrets were deferred — they do not yet exist in the environment. **Phase 18 is currently blocked on certificate procurement.** The plan must include a prerequisite task for the user to acquire the SSL.com OV cert and store the 4 secrets before the CI workflow change can be tested.

The Wails `-nsis` build produces two artifacts in `build/bin/`: `StorCat.exe` (portable) and `StorCat-amd64-installer.exe` (NSIS installer). Both must be signed before rename and upload. A known Wails issue (#3716) confirms that `wails build -nsis` does NOT sign the embedded executable — signing must be done as a post-build CI step on both files separately.

**Primary recommendation:** Sign both artifacts using `SSLcom/esigner-codesign@v1.3.2` (SHA `cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b`) before renaming and before `upload-artifact`. Add `environment: release` to the `build-windows` job to expose the signing secrets. Add `signtool verify /pa /v` as a CI gate step after signing. Signing must occur before renaming because the eSigner action uses file_path as its input — sign the canonical build output names, then rename for distribution.

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| WSIGN-01 | `release.yml` `build-windows` job signs NSIS installer with `signtool.exe` (or Azure Trusted Signing) | eSigner approach confirmed. Azure is ineligible (restricted to US/Canadian businesses). `SSLcom/esigner-codesign@v1.3.2` signs `.exe` files including NSIS installers. Both artifacts must be signed before rename. |
| WSIGN-02 | `release.yml` `build-windows` job signs portable .exe with same certificate | Same action invocation signs the portable `StorCat.exe`. Two separate eSigner action steps are needed (one per file), or use the `dir_path` batch signing approach. |
| WSIGN-03 | Signing occurs before `upload-artifact` so WinGet SHA256 is computed from signed binary | Confirmed critical. `distribute.yml` downloads the release asset and runs `sha256sum` — if a pre-signing artifact was uploaded, the WinGet manifest would reference the wrong hash. Signing → rename → upload is the mandatory order. |
| WSIGN-04 | `signtool verify` step confirms valid signature before upload | `signtool verify /pa /v` is the standard verification command. Available on all `windows-2022` runners via Windows SDK. Must run after signing and before `upload-artifact`. |
</phase_requirements>

---

## Project Constraints (from CLAUDE.md)

| Directive | Impact on This Phase |
|-----------|---------------------|
| GSD workflow enforcement: use `/gsd:execute-phase` | All tasks executed via GSD workflow |
| No Electron dependencies: use Go/Wails patterns | Not applicable — pure CI workflow change |
| GitHub Actions: pin actions to SHAs | `SSLcom/esigner-codesign` must use pinned SHA `cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b` (v1.3.2), not `@develop` |
| Branch: work on `feature/go-refactor-2.0.0-clean`, merge to `main` when complete | Verify current branch before editing `release.yml` |
| `pnpm` preferred for Node | Not applicable to this phase |

---

## Standard Stack

### Core
| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| `SSLcom/esigner-codesign` | v1.3.2 (SHA: `cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b`) | Sign .exe files via SSL.com cloud HSM | Post-June 2023 CA/Browser Forum requirement makes PFX export from new OV certs impossible; eSigner API is the only CI-compatible approach for new OV certs |
| `signtool.exe` | Windows SDK built-in (windows-2022 runner) | Verify Authenticode signature after signing | Available on all windows-2022 runners via Windows Kits — no installation needed |

### Supporting
| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| `signtool verify` | Windows SDK built-in | Gate step: verify signature before upload | After eSigner signs both files, before `upload-artifact` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `SSLcom/esigner-codesign` | Azure Artifact Signing | Azure is restricted to US/Canadian businesses as of April 2025 — ineligible for this project |
| `SSLcom/esigner-codesign` | Traditional PFX + `signtool sign /f` | Cannot purchase new OV certs with PFX export capability post-June 2023; pre-2023 certs can use this path but user has no existing Windows cert |
| `SSLcom/esigner-codesign` | `dlemstra/code-sign-action` | ARCHIVED October 2025 — do not use |
| Two separate eSigner action steps | `dir_path` batch signing | `dir_path` signs all supported files in a directory — works but is less explicit; two separate steps make the plan clearer and easier to debug |

**Installation:** No npm/pip packages needed. `SSLcom/esigner-codesign` is a GitHub Action; `signtool.exe` is pre-installed on `windows-2022` runners.

---

## Architecture Patterns

### Signing Step Sequence (Mandatory Order)

```
wails build -nsis → sign StorCat.exe → sign StorCat-amd64-installer.exe → verify both → rename → upload-artifact
```

This order cannot be changed:
- eSigner must sign before rename (signs using the original Wails output filename)
- Verify must happen after signing (confirms both files carry valid signatures)
- Upload must happen after rename (rename step creates distribution-ready filenames)
- distribute.yml downloads from the release URL and runs sha256sum — the uploaded artifact must be the signed binary

### Recommended Project Structure

No new project files are created by this phase. The only change is to `.github/workflows/release.yml`.

```
.github/
└── workflows/
    └── release.yml    # MODIFIED: build-windows job gains environment: release + signing steps
```

### Pattern 1: Add `environment: release` to build-windows job

**What:** The `release` GitHub environment holds all signing secrets. Without `environment: release`, secrets are not exposed.

**When to use:** Required for any job that reads ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, or ES_TOTP_SECRET.

```yaml
# Source: Phase 16 design + Phase 17 pattern (build-macos already has this)
build-windows:
  runs-on: windows-2022
  environment: release
  steps:
    # ... existing steps ...
```

### Pattern 2: Sign with SSL.com eSigner (per-file approach)

**What:** Two separate eSigner action invocations — one for the portable .exe, one for the NSIS installer.

**When to use:** After `wails build` completes, before renaming artifacts.

```yaml
# Source: SSLcom/esigner-codesign v1.3.2 README + ssl.com/how-to/cloud-code-signing-integration-with-github-actions
- name: Sign portable exe
  uses: SSLcom/esigner-codesign@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b  # v1.3.2
  with:
    command: sign
    username: ${{ secrets.ES_USERNAME }}
    password: ${{ secrets.ES_PASSWORD }}
    credential_id: ${{ secrets.CREDENTIAL_ID }}
    totp_secret: ${{ secrets.ES_TOTP_SECRET }}
    file_path: build\bin\StorCat.exe
    malware_block: false
    override: true

- name: Sign NSIS installer
  uses: SSLcom/esigner-codesign@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b  # v1.3.2
  with:
    command: sign
    username: ${{ secrets.ES_USERNAME }}
    password: ${{ secrets.ES_PASSWORD }}
    credential_id: ${{ secrets.CREDENTIAL_ID }}
    totp_secret: ${{ secrets.ES_TOTP_SECRET }}
    file_path: build\bin\StorCat-amd64-installer.exe
    malware_block: false
    override: true
```

**Note on `malware_block`:** Set to `false` for CI. The malware scanning feature is a paid add-on and will cause failures if not enabled on the account.

**Note on `override`:** Set to `true` to overwrite the input file in-place (no separate output path needed since rename happens after signing).

### Pattern 3: Verify with signtool

**What:** After signing both files, verify with `signtool verify /pa /v` as a CI gate.

**When to use:** After both eSigner sign steps, before rename and upload.

```yaml
# Source: Microsoft signtool documentation + signmycode.com/resources/how-to-create-verify-a-windows-authenticode-signature-using-signtool
- name: Verify Authenticode signatures
  shell: cmd
  run: |
    signtool verify /pa /v build\bin\StorCat.exe
    signtool verify /pa /v build\bin\StorCat-amd64-installer.exe
```

**Note on shell:** Use `shell: cmd` or `shell: pwsh` for Windows commands. `signtool` is on PATH on windows-2022 runners via Windows SDK. Exit code 0 = success; exit code 1 = failure. The `/pa` flag validates using Authenticode policy; `/v` gives verbose output useful for CI logs.

### Pattern 4: Sign before rename, then upload

**What:** The current workflow renames after build. Signing must be inserted between build and rename.

**When to use:** Always — current step order in `build-windows` is: build → rename → upload. New order: build → sign → verify → rename → upload.

```yaml
# Existing steps (no change):
- name: Build Windows with NSIS installer
  run: wails build -clean -platform windows/amd64 -windowsconsole -nsis

# New steps (inserted after build):
- name: Sign portable exe
  uses: SSLcom/esigner-codesign@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b
  with: ...

- name: Sign NSIS installer
  uses: SSLcom/esigner-codesign@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b
  with: ...

- name: Verify Authenticode signatures
  shell: cmd
  run: |
    signtool verify /pa /v build\bin\StorCat.exe
    signtool verify /pa /v build\bin\StorCat-amd64-installer.exe

# Existing steps (no change — signing already done):
- name: Rename Windows binary
  shell: bash
  run: mv build/bin/StorCat.exe build/bin/StorCat-${{ steps.version.outputs.VERSION }}-windows-amd64.exe

- name: Rename NSIS installer
  shell: bash
  run: mv build/bin/StorCat-amd64-installer.exe build/bin/StorCat-${{ steps.version.outputs.VERSION }}-windows-amd64-installer.exe

- name: Upload Windows artifact
  uses: actions/upload-artifact@...
```

### Anti-Patterns to Avoid

- **Signing after rename:** The eSigner action takes `file_path` as input. If the files have been renamed to their version-tagged names, the file_path reference breaks. Sign using Wails' canonical output names, then rename.
- **Signing the installer only:** WSIGN-02 requires both files. SmartScreen evaluates both the portable .exe and the installed binary. Sign both.
- **Using `@develop` tag for eSigner action:** `@develop` is a floating reference; use the pinned SHA `cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b` per CLAUDE.md requirement.
- **Using `dlemstra/code-sign-action`:** Archived October 2025. Do not use.
- **Missing `environment: release` on build-windows job:** Without this, ES_USERNAME etc. are not exposed. Phase 17's macOS job already has `environment: release`; the Windows job currently does not.
- **Uploading before verifying:** If signing fails silently (eSigner network error, bad credentials), the unsigned binary would be uploaded. Always gate with `signtool verify` before upload.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cloud HSM signing in CI | Custom signtool + PFX decode shell script | `SSLcom/esigner-codesign@v1.3.2` | Post-June 2023 OV certs cannot be exported as PFX; the private key lives on SSL.com's cloud HSM and signing requires their API |
| TOTP generation | Custom TOTP script | Built into eSigner action | eSigner action handles OAuth TOTP generation from the ES_TOTP_SECRET seed internally |
| Signature verification | Custom PE parsing | `signtool verify /pa /v` | Windows SDK tool, available on all runners, authoritative exit code |

**Key insight:** The entire Windows signing stack is cloud-HSM-mediated since the CA/Browser Forum June 2023 rule change. No local key material exists to sign with. All approaches must go through an API (SSL.com eSigner, Azure Trusted Signing, etc.).

---

## Critical Prerequisite: Certificates Not Yet Acquired

**Phase 18 is currently blocked.** As established in Phase 16-03:
- 0 Windows secrets exist in the `release` GitHub environment
- SSL.com eSigner OV certificate has not been purchased
- The 4 secrets (ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET) are not stored

**Current environment secrets (verified):** APPLE_CERTIFICATE, APPLE_CERTIFICATE_NAME, APPLE_CERTIFICATE_PASSWORD, APPLE_ID, APPLE_NOTARIZATION_PASSWORD, APPLE_TEAM_ID — all 6 Apple secrets. No Windows secrets.

The plan must include Wave 0 tasks for:
1. User purchases SSL.com OV code signing certificate
2. User enrolls in eSigner and saves the ES_TOTP_SECRET shown during QR scan
3. User stores 4 secrets in GitHub `release` environment
4. Phase 18 CI workflow changes can then be implemented and tested

The workflow changes themselves are straightforward YAML edits; the blocker is purely credential procurement.

---

## Wails NSIS Signing Architecture

### What `wails build -nsis` produces

Wails `build -nsis` produces two distinct files in `build/bin/`:
1. `StorCat.exe` — the portable application executable (directly runnable)
2. `StorCat-amd64-installer.exe` — the NSIS installer that wraps and installs `StorCat.exe`

### Known Issue: NSIS does not sign the embedded exe

Wails issue #3716 (open as of 2026-03) confirms that `wails build -nsis` does NOT sign the application binary embedded in the installer. After installation, `C:\Program Files\StorCat\StorCat.exe` is unsigned.

**Solution for Phase 18:** Sign `build/bin/StorCat.exe` BEFORE the NSIS installer would normally embed it. However, since `wails build -nsis` runs the full build in one step (including NSIS installer creation), the recommended approach is:

- Sign `StorCat.exe` first (even though it's already embedded in the installer at this point)
- Sign `StorCat-amd64-installer.exe` separately

This means the portable .exe download will be signed, and the NSIS installer wrapper will be signed. However, if a user runs the NSIS installer, the installed `StorCat.exe` in `Program Files` will be the version that was embedded during `wails build -nsis` — which is unsigned at the NSIS build stage.

**Workaround options:**
1. Accept that the installed binary is the same signed binary (since we sign `StorCat.exe` before renaming) — but the NSIS installer already contains an unsigned version at embed time. This is the current state.
2. Use a two-step build: build without `-nsis` to get `StorCat.exe`, sign it, then build the NSIS installer separately. However, Wails does not support this workflow natively.
3. Use Wails' NSIS `!finalize` signing directive in the custom NSIS script (referenced but commented out in Wails templates).

**For Phase 18:** Sign both output files in CI as post-build steps. The NSIS installer itself will be signed (satisfying SmartScreen for the download). The embedded binary is a known limitation documented in Wails issue #3716. The planner should note this as acceptable scope for v2.3.0 — WSIGN-01 and WSIGN-02 require CI signing, not per-extracted-binary verification.

---

## Common Pitfalls

### Pitfall 1: ES_TOTP_SECRET Shown Only Once

**What goes wrong:** The TOTP secret seed is displayed one time during SSL.com eSigner enrollment alongside the QR code. If the page is closed without recording the base32 seed, re-enrollment is required.
**Why it happens:** Standard TOTP security design — the seed enables offline token generation and must be kept secret.
**How to avoid:** During enrollment, extract the `secret=XXXXXX` value from the `otpauth://totp/...?secret=XXXXXX` QR URI. Save immediately as `ES_TOTP_SECRET`. The runbook at `docs/runbooks/credential-rotation.md` documents this.
**Warning signs:** eSigner action returns "Invalid TOTP" or authentication failure.

### Pitfall 2: `environment: release` Missing from build-windows Job

**What goes wrong:** Secrets ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, and ES_TOTP_SECRET are not exposed to the job. The eSigner action receives empty credentials and fails.
**Why it happens:** GitHub Actions environment secrets require explicit opt-in via `environment: release` on the job. The `build-macos` job has this; `build-windows` currently does not.
**How to avoid:** Add `environment: release` to the `build-windows` job definition. This is the single-line addition that unlocks all Windows signing secrets.
**Warning signs:** eSigner action fails with "invalid credentials" even though secrets appear to be set.

### Pitfall 3: Signing After Rename

**What goes wrong:** The eSigner `file_path` points to `build\bin\StorCat-v2.3.0-windows-amd64.exe` (the renamed file), but the action cannot find the file because the rename happened before signing.
**Why it happens:** The current workflow renames the binary immediately after build. If signing steps are inserted after rename, the file_path must use the renamed name.
**How to avoid:** Insert signing steps between the build step and the rename steps. Use Wails' canonical output filenames (`build\bin\StorCat.exe`, `build\bin\StorCat-amd64-installer.exe`) as `file_path`.
**Warning signs:** eSigner action fails with "file not found" or similar.

### Pitfall 4: eSigner `malware_block` Default Behavior

**What goes wrong:** The eSigner action has a `malware_block` input. If set to `true` (or if the default is `true`), SSL.com's malware scanning service is invoked. If the account does not have this add-on, the signing fails.
**Why it happens:** `malware_block` is an optional SSL.com service add-on.
**How to avoid:** Explicitly set `malware_block: false` in the action inputs.
**Warning signs:** Signing step fails with SSL.com API error about malware scanning subscription.

### Pitfall 5: `signtool verify` Not on PATH

**What goes wrong:** `signtool: command not found` or similar error in the verify step.
**Why it happens:** `signtool.exe` is installed via Windows SDK in the `windows-2022` runner, but may not be on PATH by default depending on the SDK version installed.
**How to avoid:** Use `signtool verify /pa /v ...` with `shell: cmd`. If signtool is not on PATH, locate it explicitly: `"C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x86\signtool.exe"`. Alternatively, add a step using the `lukesampson/signtool` or `vedantmgoyal2009/winget-releaser` pattern to locate and add signtool to PATH first.
**Warning signs:** Step fails with "signtool is not recognized as an internal or external command."

### Pitfall 6: distribute.yml SHA256 from Pre-Signing Upload

**What goes wrong:** If the NSIS installer is uploaded before signing, `distribute.yml` downloads the unsigned installer from the release, computes its SHA256, and commits that hash to the WinGet manifest. When a user runs `winget install StorCat`, WinGet downloads the signed installer (which has a different SHA256) and the install fails with a hash mismatch.
**Why it happens:** SHA256 of a file changes after signing — the signature bytes are appended to the PE file.
**How to avoid:** Signing MUST happen before `upload-artifact`. The distribute.yml job runs after release is published (which happens after all artifacts are uploaded), so if artifacts are signed before upload, distribute.yml will always compute from signed binaries.
**Warning signs:** WinGet install fails with "Hash does not match" after a release.

### Pitfall 7: SmartScreen Reputation Not Instant

**What goes wrong:** Even after signing with an OV certificate, SmartScreen shows a warning on first installs.
**Why it happens:** Since August 2024, Microsoft removed EV certificate immediate-reputation advantage. Both EV and OV certificates now build reputation organically through download volume. This is expected behavior, not a signing failure.
**How to avoid:** No workaround — reputation builds over time with downloads. Document this for the user so they understand SmartScreen warnings will appear initially.
**Warning signs:** Users report SmartScreen warnings even though signing succeeds in CI.

---

## Code Examples

Verified patterns from official sources:

### Complete build-windows Job (modified sections)

```yaml
# Source: SSLcom/esigner-codesign v1.3.2 + ssl.com/how-to/cloud-code-signing-integration-with-github-actions
build-windows:
  runs-on: windows-2022
  environment: release   # NEW: exposes ES_* secrets
  steps:
    - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2

    - name: Set up Go
      uses: actions/setup-go@0aaccfd150d50ccaeb58ebd88d36e91967a5f35b  # v5.4.0
      with:
        go-version-file: go.mod

    - name: Set up Node
      uses: actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020  # v4.4.0
      with:
        node-version: '20'

    - name: Install Wails
      run: go install github.com/wailsapp/wails/v2/cmd/wails@v2.10.2

    - name: Get version from tag
      id: version
      shell: bash
      run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

    - name: Build Windows with NSIS installer
      run: wails build -clean -platform windows/amd64 -windowsconsole -nsis

    # NEW: Sign portable exe
    - name: Sign portable exe
      uses: SSLcom/esigner-codesign@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b  # v1.3.2
      with:
        command: sign
        username: ${{ secrets.ES_USERNAME }}
        password: ${{ secrets.ES_PASSWORD }}
        credential_id: ${{ secrets.CREDENTIAL_ID }}
        totp_secret: ${{ secrets.ES_TOTP_SECRET }}
        file_path: build\bin\StorCat.exe
        malware_block: false
        override: true

    # NEW: Sign NSIS installer
    - name: Sign NSIS installer
      uses: SSLcom/esigner-codesign@cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b  # v1.3.2
      with:
        command: sign
        username: ${{ secrets.ES_USERNAME }}
        password: ${{ secrets.ES_PASSWORD }}
        credential_id: ${{ secrets.CREDENTIAL_ID }}
        totp_secret: ${{ secrets.ES_TOTP_SECRET }}
        file_path: build\bin\StorCat-amd64-installer.exe
        malware_block: false
        override: true

    # NEW: Verify signatures (CI gate)
    - name: Verify Authenticode signatures
      shell: cmd
      run: |
        signtool verify /pa /v build\bin\StorCat.exe
        signtool verify /pa /v build\bin\StorCat-amd64-installer.exe

    - name: Rename Windows binary
      shell: bash
      run: mv build/bin/StorCat.exe build/bin/StorCat-${{ steps.version.outputs.VERSION }}-windows-amd64.exe

    - name: Rename NSIS installer
      shell: bash
      run: mv build/bin/StorCat-amd64-installer.exe build/bin/StorCat-${{ steps.version.outputs.VERSION }}-windows-amd64-installer.exe

    - name: Upload Windows artifact
      uses: actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02  # v4.6.2
      with:
        name: StorCat-windows-amd64
        path: |
          build/bin/StorCat-*-windows-amd64.exe
          build/bin/StorCat-*-windows-amd64-installer.exe
        retention-days: 1
```

### Secret Storage Commands (Wave 0 prerequisite — user action)

```bash
# After SSL.com eSigner enrollment — store 4 secrets:
gh secret set ES_USERNAME --env release --repo scottkw/storcat
gh secret set ES_PASSWORD --env release --repo scottkw/storcat
gh secret set CREDENTIAL_ID --env release --repo scottkw/storcat
gh secret set ES_TOTP_SECRET --env release --repo scottkw/storcat

# Verify all 10 secrets now present:
gh secret list --env release --repo scottkw/storcat
```

### Signature Verification (local, for debugging)

```powershell
# Source: Microsoft signtool documentation
# On a Windows machine with Windows SDK:
signtool verify /pa /v StorCat.exe
# Expected output includes: "Signature verified."
# Exit code 0 = valid, 1 = invalid
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| OV cert exported as .pfx, base64 in GitHub secret | Cloud HSM via SSL.com eSigner API (no export) | June 1, 2023 (CA/Browser Forum) | New OV certs cannot be exported — must use API-based signing |
| EV certificate for instant SmartScreen | OV + organic reputation building | August 2024 (Microsoft policy) | EV no longer grants instant reputation; OV and EV are equivalent for SmartScreen |
| `dlemstra/code-sign-action` | ARCHIVED — do not use | October 2025 | Cannot reference in any plans |
| Azure Trusted Signing for individual devs | Restricted to US/Canadian businesses | April 2025 | Individual developers cannot use Azure Artifact Signing |
| `actions-codesigner` (SSL.com deprecated repo) | `esigner-codesign` (active repo) | 2023 | `actions-codesigner` explicitly deprecated; use `esigner-codesign` |

**Deprecated/outdated:**
- `dlemstra/code-sign-action`: Archived October 2025 — any workflow using it will fail
- `SSLcom/actions-codesigner`: Explicitly deprecated — README says to use `esigner-codesign` instead
- EV certificates: No SmartScreen advantage over OV since August 2024

---

## Open Questions

1. **`override: true` behavior in esigner-codesign v1.3.2**
   - What we know: The action has an `override` input. Setting `true` should sign in-place, overwriting the source file.
   - What's unclear: Whether the default (without `override`) places the signed file in `output_path`. If `output_path` is omitted and `override` is false, the signed file location may be ambiguous.
   - Recommendation: Explicitly set `override: true` and omit `output_path` to sign in-place. This is the simplest path for a rename-after-sign workflow. Verify this behavior during Wave 1 implementation.

2. **signtool.exe PATH on windows-2022 runners**
   - What we know: `signtool.exe` is installed on windows-2022 runners via Windows SDK. The exact PATH varies by SDK version (e.g., `C:\Program Files (x86)\Windows Kits\10\bin\10.0.22621.0\x86\signtool.exe`).
   - What's unclear: Whether `signtool` is reliably on PATH without explicit setup, or whether a `where signtool` or Add Signtool Action step is needed.
   - Recommendation: Use `shell: cmd` and reference `signtool` without full path first. If the verify step fails with "not recognized," add a preceding step to locate and add to PATH. The plan should include a fallback note.

3. **NSIS installer embedded binary signing**
   - What we know: Wails issue #3716 confirms the embedded `StorCat.exe` inside the NSIS installer is unsigned at build time. The installed binary in `Program Files` is the unsigned version.
   - What's unclear: Whether this matters for SmartScreen — SmartScreen evaluates the installer's signature (which we do sign), not the installed binary's signature at install time for initial reputation building.
   - Recommendation: Accept as known limitation for Phase 18. Sign both output files in CI. Document in plan that installed binary signature state is out of scope (requires Wails NSIS template changes, tracked in Wails issue #3716).

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `windows-2022` GitHub-hosted runner | All WSIGN-* | Yes | Runner image 20250322+ | — |
| `signtool.exe` | WSIGN-04 | Yes (pre-installed on runner via Windows SDK) | 10.0.22621.0 or newer | Locate via `where signtool` at runtime |
| `SSLcom/esigner-codesign` action | WSIGN-01, WSIGN-02 | Yes (GitHub Marketplace, active) | v1.3.2 (SHA cf5f6c1d) | — |
| GitHub `release` environment | WSIGN-01, WSIGN-02 | Yes (exists — verified Phase 16) | — | — |
| `ES_USERNAME` GitHub secret | WSIGN-01, WSIGN-02 | **NOT PRESENT** | — | **No fallback — must acquire SSL.com cert** |
| `ES_PASSWORD` GitHub secret | WSIGN-01, WSIGN-02 | **NOT PRESENT** | — | **No fallback — must acquire SSL.com cert** |
| `CREDENTIAL_ID` GitHub secret | WSIGN-01, WSIGN-02 | **NOT PRESENT** | — | **No fallback — must acquire SSL.com cert** |
| `ES_TOTP_SECRET` GitHub secret | WSIGN-01, WSIGN-02 | **NOT PRESENT** | — | **No fallback — must acquire SSL.com cert** |

**Missing dependencies with no fallback:**
- All 4 SSL.com eSigner secrets — blocked on certificate procurement. User must purchase SSL.com OV code signing certificate, enroll in eSigner, and store the 4 secrets before the CI workflow can be tested.

**Missing dependencies with fallback:**
- None — signtool and runner are available; only the secrets are missing.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | None (CI workflow verification — no unit test framework applicable) |
| Config file | `.github/workflows/release.yml` (the artifact under test) |
| Quick run command | `yamllint .github/workflows/release.yml` (YAML syntax check) |
| Full suite command | Push a `v*.*.*` tag to trigger full CI pipeline |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| WSIGN-01 | NSIS installer signed with Authenticode | integration | `signtool verify /pa /v build\bin\StorCat-amd64-installer.exe` in CI | N/A — CI step |
| WSIGN-02 | Portable .exe signed with Authenticode | integration | `signtool verify /pa /v build\bin\StorCat.exe` in CI | N/A — CI step |
| WSIGN-03 | Signing occurs before upload-artifact | structural | Inspect workflow YAML — verify sign steps precede upload step | N/A — YAML review |
| WSIGN-04 | `signtool verify` gate step in workflow | smoke | `grep "signtool verify /pa" .github/workflows/release.yml` | N/A — YAML review |

### Sampling Rate

- **Per task commit:** `yamllint .github/workflows/release.yml` — confirm YAML is valid
- **Per wave merge:** Full CI run via tag push on a test tag (e.g., `v99.0.0-test`) after secrets are stored — then delete the draft release
- **Phase gate:** Full CI green + `signtool verify /pa /v` passes in CI logs + `distribute.yml` WinGet manifest SHA256 matches signed binary

### Wave 0 Gaps

**Credential prerequisite (user action — not a code gap):**
- [ ] User must purchase SSL.com OV code signing certificate
- [ ] User must enroll in eSigner and save ES_TOTP_SECRET during QR scan
- [ ] User must store ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET in GitHub `release` environment

**Code gaps:**
- None — no new test files needed; this is a CI workflow change phase

*(Existing `.github/workflows/release.yml` is the file to modify; no new test infrastructure required)*

---

## Sources

### Primary (HIGH confidence)
- `gh api repos/scottkw/storcat/environments/release/secrets` — confirmed 6 Apple secrets, 0 Windows secrets (2026-03-28)
- [SSLcom/esigner-codesign releases page](https://github.com/SSLcom/esigner-codesign/releases) — v1.3.2 confirmed, SHA `cf5f6c1d38ad10f47e3ed9aca873f429b1a8d85b`, December 22, 2025
- [SSLcom/actions-codesigner](https://github.com/SSLcom/actions-codesigner) — confirmed deprecated; README says use esigner-codesign instead
- [SSL.com: Cloud Code Signing Integration with GitHub Actions](https://www.ssl.com/how-to/cloud-code-signing-integration-with-github-actions/) — current secret names: ES_USERNAME, ES_PASSWORD, CREDENTIAL_ID, ES_TOTP_SECRET; action usage pattern
- Existing `.github/workflows/release.yml` — confirmed build-windows job structure, rename order, upload-artifact targets
- Phase 16-03-SUMMARY.md — confirmed Windows signing deferred, 0 secrets stored, phases 18/20 blocked
- `gh api repos/scottkw/storcat/environments/release/secrets` (live query) — 6 secrets confirmed: APPLE_CERTIFICATE, APPLE_CERTIFICATE_NAME, APPLE_CERTIFICATE_PASSWORD, APPLE_ID, APPLE_NOTARIZATION_PASSWORD, APPLE_TEAM_ID

### Secondary (MEDIUM confidence)
- [Wails issue #3716](https://github.com/wailsapp/wails/issues/3716) — confirmed wails -nsis does not sign embedded exe; signing must be done as CI post-build step; issue is open and in v3 backlog
- [Microsoft Q&A: Reputation with OV certificates](https://learn.microsoft.com/en-us/answers/questions/417016/reputation-with-ov-certificates-and-are-ev-certifi) — OV reputation builds organically; EV advantage removed August 2024
- [BigBinary Blog: EV code sign Windows application using ssl.com](https://www.bigbinary.com/blog/ev-code-sign-windows-application-ssl-com) — eSigner pattern confirmed with same 4 secret names
- [signmycode.com: How to Create & Verify Windows Authenticode Signature](https://signmycode.com/resources/how-to-create-verify-a-windows-authenticode-signature-using-signtool) — `signtool verify /pa /v` syntax confirmed

### Tertiary (LOW confidence)
- [Federico Terzi: Automatic Code-signing on Windows using GitHub Actions](https://federicoterzi.com/blog/automatic-codesigning-on-windows-using-github-actions/) — traditional PFX approach (not applicable post-June 2023); signtool flags are consistent with Microsoft docs
- signtool.exe PATH availability on windows-2022 runners — assumed based on Windows SDK pre-install; verify during CI run

---

## Metadata

**Confidence breakdown:**
- Standard stack (action + SHA): HIGH — v1.3.2 SHA confirmed from releases page; action is actively maintained
- Workflow pattern: HIGH — derived from existing release.yml structure + Phase 17 macOS patterns + SSL.com docs
- eSigner action inputs: MEDIUM — based on SSL.com official guide + action README; exact `override: true` behavior needs runtime validation
- signtool verify PATH: MEDIUM — pre-installed on windows-2022 runners but exact PATH varies by SDK version
- Pitfalls: HIGH — sourced from official docs, Phase 16 research, and Wails issue tracker

**Research date:** 2026-03-28
**Valid until:** 2026-06-28 (SSL.com eSigner API and action inputs are stable; Windows runner images update quarterly)
