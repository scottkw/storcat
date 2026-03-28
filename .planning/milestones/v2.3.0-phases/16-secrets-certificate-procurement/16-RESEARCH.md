# Phase 16: Secrets & Certificate Procurement - Research

**Researched:** 2026-03-27
**Domain:** Code signing credentials, GitHub Actions environments, Apple/Windows certificate management
**Confidence:** HIGH (macOS/Apple path), MEDIUM (Windows path — vendor decision required)

## Summary

Phase 16 is a pure credential-gathering and configuration phase. No code changes are made to the application or CI workflows — the sole output is that all 9 signing secrets exist in a GitHub `release` environment, and a rotation runbook document exists. This phase unblocks Phases 17 and 18, which are independent of each other and can proceed in parallel once complete.

The macOS side is straightforward: a valid Developer ID Application certificate for `Ken Scott (S2K7P43927)` already exists on the local keychain with expiry February 1, 2027. It needs to be exported as .p12, base64-encoded, and stored alongside 5 other Apple credentials in the `release` environment.

The Windows side requires a deliberate procurement decision. Post-June 2023, all new OV code signing certificates are non-exportable by CA/Browser Forum requirement — keys are stored on FIPS-compliant hardware tokens or cloud HSMs and cannot be placed in a .pfx file. Two viable paths exist: (1) SSL.com eSigner cloud signing (OV tier, ~$20/month, 20 signings), or (2) Azure Artifact Signing (formerly Trusted Signing, $9.99/month) — but Azure is currently restricted to US/Canadian businesses with 3+ years of history as of April 2025, making SSL.com eSigner the default recommendation for individual developers.

**Primary recommendation:** Use SSL.com OV eSigner for Windows (API-based, 4 secrets: `ES_USERNAME`, `ES_PASSWORD`, `CREDENTIAL_ID`, `ES_TOTP_SECRET`) and export the existing Apple Developer ID cert as .p12 for macOS (5 secrets: `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD`, `APPLE_CERTIFICATE_NAME`, `APPLE_NOTARIZATION_PASSWORD`, `APPLE_TEAM_ID`). Total = 9 secrets.

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CRED-01 | Locate or renew existing Apple Developer ID Application certificate | Certificate already present: `Developer ID Application: Ken Scott (S2K7P43927)`, expiry 2027-02-01 — no renewal needed, just verification |
| CRED-02 | Export Developer ID cert as .p12 and store base64-encoded as `APPLE_CERTIFICATE` GitHub secret | Step-by-step Keychain Access export process documented below; base64 encode command and `gh secret set` command provided |
| CRED-03 | Obtain Windows OV code signing certificate with RSA (not ECDSA) confirmed before purchase | Post-June 2023 no-export rule means SSL.com eSigner is recommended path; RSA vs ECDSA must be confirmed before purchase |
| CRED-04 | Windows PFX base64-encoded in `WINDOWS_CERTIFICATE` GitHub secret | With eSigner, no PFX is exported — instead 4 API secrets replace this single PFX secret; planner must decide naming scheme |
| CRED-05 | GitHub `release` environment exists with protection rules and all 9 signing secrets populated | Environment does not exist yet (`gh api` confirmed 0 environments); creation steps documented |
| CRED-06 | Credential rotation runbook document exists | Template content documented in this research; planner creates doc task |
</phase_requirements>

---

## Standard Stack

### Core Tools

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| `security` CLI | macOS built-in | Keychain operations, cert export, identity verification | Apple's own tool — no alternative |
| `base64` | macOS built-in | Encode .p12 for storage as GitHub secret | Shell standard |
| `gh` CLI | 2.x | Create environment, set secrets | Official GitHub CLI |
| Keychain Access.app | macOS built-in | GUI for .p12 export | Most reliable for cert+key export |

### Apple-Specific Secrets (macOS)

| Secret Name | Contents | How to Obtain |
|-------------|----------|---------------|
| `APPLE_CERTIFICATE` | base64-encoded .p12 file | Export from Keychain Access → base64 encode |
| `APPLE_CERTIFICATE_PASSWORD` | Password set during .p12 export | User-defined at export time |
| `APPLE_CERTIFICATE_NAME` | Full identity string from `security find-identity` | Already known: `Developer ID Application: Ken Scott (S2K7P43927)` |
| `APPLE_NOTARIZATION_PASSWORD` | App-specific password for Apple ID | Generate at appleid.apple.com → Sign-In & Security → App-Specific Passwords |
| `APPLE_TEAM_ID` | 10-char Team ID from Apple Developer portal | Already known: `S2K7P43927` |

**5 Apple secrets total.**

### Windows Secrets via SSL.com eSigner (Recommended)

| Secret Name | Contents | How to Obtain |
|-------------|----------|---------------|
| `ES_USERNAME` | SSL.com account email/username | SSL.com account creation |
| `ES_PASSWORD` | SSL.com account password | SSL.com account creation |
| `CREDENTIAL_ID` | Signing credential ID from SSL.com dashboard | Generated after certificate enrollment |
| `ES_TOTP_SECRET` | TOTP secret for eSigner MFA automation | Shown once during eSigner enrollment; save during setup |

**4 Windows secrets total. Grand total = 9 secrets.**

### Alternatives Considered

| Windows Approach | Tradeoff | Recommended? |
|-----------------|----------|--------------|
| SSL.com eSigner OV | $20/month, 20 signings, no hardware token, full CI automation | YES — default |
| Azure Artifact Signing | $9.99/month, instant SmartScreen reputation, BUT restricted to US/Canadian businesses 3+ years old as of April 2025 | NO — eligibility risk |
| Traditional PFX (pre-June 2023 cert) | Only possible if user already owns a pre-June 2023 certificate — cannot purchase new ones with export capability | NO — not purchasable |
| Certum SimplySign | Non-exportable; SimplySign desktop app required; harder to automate | NO — friction |

---

## Architecture Patterns

### GitHub Environment Architecture

```
GitHub Repository: scottkw/storcat
└── Environment: release
    ├── Protection Rules
    │   ├── Required Reviewers: [scottkw] (1 person)
    │   └── Deployment Branch: Tag pattern v*.*.*
    └── Secrets (9 total)
        ├── APPLE_CERTIFICATE            ← base64 .p12
        ├── APPLE_CERTIFICATE_PASSWORD   ← .p12 export password
        ├── APPLE_CERTIFICATE_NAME       ← "Developer ID Application: Ken Scott (S2K7P43927)"
        ├── APPLE_NOTARIZATION_PASSWORD  ← app-specific password
        ├── APPLE_TEAM_ID               ← S2K7P43927
        ├── ES_USERNAME                  ← SSL.com username
        ├── ES_PASSWORD                  ← SSL.com password
        ├── CREDENTIAL_ID               ← SSL.com credential ID
        └── ES_TOTP_SECRET              ← SSL.com TOTP secret
```

### How Downstream Phases Consume Secrets

Phase 17 (macOS Signing) will reference:
```yaml
environment: release
env:
  APPLE_CERTIFICATE: ${{ secrets.APPLE_CERTIFICATE }}
  APPLE_CERTIFICATE_PASSWORD: ${{ secrets.APPLE_CERTIFICATE_PASSWORD }}
  APPLE_CERTIFICATE_NAME: ${{ secrets.APPLE_CERTIFICATE_NAME }}
  APPLE_NOTARIZATION_PASSWORD: ${{ secrets.APPLE_NOTARIZATION_PASSWORD }}
  APPLE_TEAM_ID: ${{ secrets.APPLE_TEAM_ID }}
```

Phase 18 (Windows Signing) will reference:
```yaml
environment: release
env:
  ES_USERNAME: ${{ secrets.ES_USERNAME }}
  ES_PASSWORD: ${{ secrets.ES_PASSWORD }}
  CREDENTIAL_ID: ${{ secrets.CREDENTIAL_ID }}
  ES_TOTP_SECRET: ${{ secrets.ES_TOTP_SECRET }}
```

### Anti-Patterns to Avoid

- **Repository-level secrets instead of environment secrets:** Secrets in the `release` environment are only accessible to jobs that declare `environment: release`, adding a protection layer. Storing them at repo level bypasses this gate.
- **Committing the .p12 file to git:** Even if encrypted, this is unnecessary risk — base64 in GitHub secrets is the correct storage.
- **Using the same app-specific password for multiple services:** Generate a dedicated app-specific password named "StorCat CI" for traceability.
- **Skipping the TOTP secret backup:** The `ES_TOTP_SECRET` is shown only once during SSL.com eSigner enrollment. Store it immediately or enrollment must be repeated.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Keychain import in CI | Custom shell script to manage keychains | `apple-actions/import-codesign-certs@v6` (SHA: `b610f78`) | Handles ACL setup that prevents codesign hangs — easy to miss |
| Windows signing | Raw signtool PowerShell | `SSLcom/esigner-codesign` GitHub Action | Handles TOTP generation, CodeSignTool lifecycle |
| Secret validation | Manual checks | `security find-identity -v -p codesigning` gate step | Fails fast if cert not loaded correctly |

---

## Common Pitfalls

### Pitfall 1: Exporting Only the Certificate, Not the Private Key

**What goes wrong:** Export produces a .p12 that imports correctly but `codesign` fails with "no private key found."
**Why it happens:** Keychain Access export only grabs the certificate leaf if the private key is not selected in the same export operation.
**How to avoid:** In Keychain Access, select BOTH the `Developer ID Application: Ken Scott (S2K7P43927)` certificate AND the private key item below it (usually labeled with the key name), then right-click → "Export 2 Items…" → choose .p12 format.
**Warning signs:** `security find-identity -v -p codesigning` shows the identity, but `codesign` fails with "code object is not signed" or "no identity found."

### Pitfall 2: App-Specific Password vs. Apple ID Password

**What goes wrong:** `xcrun notarytool` rejects the password because the main Apple ID password is used instead of an app-specific password.
**Why it happens:** Apple ID accounts with two-factor authentication require app-specific passwords for non-Apple tools.
**How to avoid:** Go to appleid.apple.com → Sign-In and Security → App-Specific Passwords → Generate. Name it "StorCat CI" for traceability.
**Warning signs:** `notarytool` returns "Unable to authenticate" or "Password verification failed."

### Pitfall 3: ECDSA Certificate Purchase

**What goes wrong:** Windows binary is signed, but Microsoft SmartScreen still warns users because ECDSA certificates have poor reputation history with signtool.
**Why it happens:** SmartScreen reputation is algorithm-sensitive; RSA has broader legacy trust. Resellers sometimes sell ECDSA certificates without flagging this.
**How to avoid:** Confirm "RSA" explicitly with the vendor before purchase. For SSL.com: OV code signing certificates use RSA by default.
**Warning signs:** signtool signs successfully but `signtool verify` reports algorithm issues; SmartScreen warnings persist even with valid signature.

### Pitfall 4: TOTP Secret Lost After Enrollment

**What goes wrong:** The `ES_TOTP_SECRET` is the base32 seed used to generate time-based tokens. It is displayed only once during SSL.com eSigner enrollment.
**Why it happens:** Standard TOTP design — the seed is shown at setup time to load into an authenticator app. If the page is closed without recording it, re-enrollment is required.
**How to avoid:** During eSigner enrollment, the QR code contains a `otpauth://totp/...?secret=XXXXXX` URI. Extract and save `XXXXXX` as `ES_TOTP_SECRET` immediately.
**Warning signs:** eSigner automation returns "Invalid TOTP" or similar authentication failure in CI.

### Pitfall 5: GitHub Environment Protection Blocking CI

**What goes wrong:** The `release` environment with "Required Reviewers" blocks the CI workflow from running automatically on tags.
**Why it happens:** Required Reviewer protection requires a human to approve before secrets are exposed. This is correct security behavior but must be accounted for.
**How to avoid:** For a solo developer, set yourself as the required reviewer. Alternatively, use "Deployment branch" restriction only (require tag pattern `v*.*.*`) without reviewer requirement.
**Warning signs:** CI starts, reaches a job using `environment: release`, then pauses with "Waiting for review."

### Pitfall 6: Azure Trusted Signing (Now Azure Artifact Signing) Eligibility

**What goes wrong:** Developer signs up for Azure Trusted Signing expecting individual dev access, but is denied at identity verification because accounts are restricted to US/Canadian businesses with 3+ year history.
**Why it happens:** As of April 2025, Microsoft restricted Azure Artifact Signing to businesses only (previously included individual developers during public preview).
**How to avoid:** Use SSL.com eSigner as the default Windows path. Revisit Azure Artifact Signing only if business eligibility is confirmed.

---

## Code Examples

### Verify Existing Apple Certificate

```bash
# Source: Apple security CLI
security find-identity -v -p codesigning
# Expected: "Developer ID Application: Ken Scott (S2K7P43927)"

# Check expiry
security find-certificate -c "Developer ID Application: Ken Scott" -p | \
  openssl x509 -noout -dates
# Current cert: notBefore=Jul 29 21:50:59 2025 GMT, notAfter=Feb  1 22:12:15 2027 GMT
```

### Export .p12 from Keychain (GUI Method — Recommended)

```
1. Open Keychain Access.app
2. Select "login" keychain, "My Certificates" category
3. Expand "Developer ID Application: Ken Scott (S2K7P43927)" to see private key
4. Click the certificate row to select it
5. Cmd+click the private key row beneath it (select BOTH items)
6. Right-click → "Export 2 Items..."
7. Save as "storcat-dev-id.p12"
8. Enter a strong password (record it — becomes APPLE_CERTIFICATE_PASSWORD)
```

### Encode .p12 and Store as GitHub Secret

```bash
# Source: macOS base64 CLI + gh CLI
base64 -i storcat-dev-id.p12 | pbcopy
# Paste value → APPLE_CERTIFICATE

# Or store directly:
gh secret set APPLE_CERTIFICATE --env release --repo scottkw/storcat \
  --body "$(base64 -i storcat-dev-id.p12)"

gh secret set APPLE_CERTIFICATE_PASSWORD --env release --repo scottkw/storcat
gh secret set APPLE_CERTIFICATE_NAME --env release --repo scottkw/storcat \
  --body "Developer ID Application: Ken Scott (S2K7P43927)"
gh secret set APPLE_TEAM_ID --env release --repo scottkw/storcat \
  --body "S2K7P43927"
gh secret set APPLE_NOTARIZATION_PASSWORD --env release --repo scottkw/storcat
```

### Create GitHub release Environment

```bash
# Source: GitHub REST API via gh CLI
gh api --method PUT repos/scottkw/storcat/environments/release \
  --field wait_timer=0 \
  --field prevent_self_review=false \
  --field reviewers='[]' \
  --field deployment_branch_policy='{"protected_branches":false,"custom_branch_policies":true}'

# Add tag deployment policy (v*.*.* only)
gh api --method POST \
  repos/scottkw/storcat/environments/release/deployment-branch-policies \
  --field name="v*.*.*" \
  --field type="tag"
```

### SSL.com eSigner GitHub Action Usage (Phase 18 Preview)

```yaml
# Source: https://github.com/SSLcom/esigner-codesign
- name: Sign Windows binary
  uses: SSLcom/esigner-codesign@develop
  with:
    command: sign
    username: ${{ secrets.ES_USERNAME }}
    password: ${{ secrets.ES_PASSWORD }}
    credential_id: ${{ secrets.CREDENTIAL_ID }}
    totp_secret: ${{ secrets.ES_TOTP_SECRET }}
    file_path: build/bin/StorCat.exe
    output_path: build/bin/
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Export OV cert as .pfx, store in GitHub secret | Cloud HSM / eSigner API (no export) | June 1, 2023 (CA/Browser Forum requirement) | Windows signing now requires API-based cloud service, not base64 PFX |
| Apple `altool` for notarization | `xcrun notarytool` | Nov 2023 (altool decommissioned) | Must use notarytool; altool is dead |
| `dlemstra/code-sign-action` for Windows | ARCHIVED — do not use | October 2025 | Use SSL.com eSigner or direct signtool |
| EV certificate for immediate SmartScreen | OV + reputation building (EV lost advantage Aug 2024) | August 2024 | EV no longer provides instant SmartScreen trust over OV; OV is sufficient |
| Azure Trusted Signing for individual devs | Restricted to US/Canadian businesses | April 2025 | Individual developers cannot use Azure Artifact Signing |

**Deprecated/outdated:**
- `gon`: Apple decommissioned `altool` which `gon` depended on — replaced by `xcrun notarytool` directly
- `dlemstra/code-sign-action`: Archived October 2025 — do not reference in any plans
- EV certificates: No advantage over OV since August 2024 per Microsoft policy change; hardware token requirement is incompatible with CI
- PFX export from OV certs purchased after June 1, 2023: Impossible per CA/Browser Forum baseline requirements

---

## Open Questions

1. **Windows certificate vendor decision**
   - What we know: SSL.com eSigner OV at $20/month is the recommended path; Azure Artifact Signing is blocked for individual developers
   - What's unclear: User preference between SSL.com eSigner vs. accepting a SmartScreen warning period (sign but no cert = big warning; OV cert = reduced warning after reputation builds)
   - Recommendation: Plan for SSL.com eSigner as the default. Include a decision point task in the plan: "Confirm Windows cert vendor and enroll before proceeding."

2. **Windows PFX secret naming (CRED-04 vs. eSigner secrets)**
   - What we know: CRED-04 specifies `WINDOWS_CERTIFICATE` as a base64-encoded PFX. With eSigner, there is no PFX — instead there are 4 API secrets.
   - What's unclear: Requirements document was written with traditional PFX in mind; eSigner changes the credential set entirely.
   - Recommendation: Plan should treat CRED-04 as "Windows signing credentials stored in GitHub" and use the 4 eSigner secrets. Phase 18 plan will consume these 4 secrets, not a PFX.

3. **Apple notarization: Apple ID vs. App Store Connect API Key**
   - What we know: Apple ID + app-specific password works today. REQUIREMENTS.md lists `SIGN-07` (App Store Connect API key) as a post-v2.3.x improvement.
   - What's unclear: No blocker; app-specific password is the v2.3.0 approach.
   - Recommendation: Use Apple ID + app-specific password for this phase. Document App Store Connect API key path in runbook.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `security` CLI | CRED-01, CRED-02 | Yes (macOS built-in) | macOS 25.3.0 | — |
| `base64` CLI | CRED-02 | Yes (macOS built-in) | — | — |
| `gh` CLI | CRED-05 | Yes | authenticated as `scottkw` | — |
| Apple Developer Account | CRED-01, CRED-02 | Yes (Team ID `S2K7P43927` confirmed) | — | — |
| SSL.com eSigner account | CRED-03, CRED-04 | Not yet (requires signup) | — | Azure Artifact Signing (blocked for individual devs) |
| GitHub `release` environment | CRED-05 | Not yet (0 environments found) | — | — |
| appleid.apple.com access | CRED-02 (notarization password) | Assumed yes | — | — |

**Missing dependencies with no fallback:**
- SSL.com eSigner account: Requires new account creation and identity verification before credentials exist

**Missing dependencies with fallback:**
- GitHub `release` environment: Does not exist, but creation is a 2-command `gh api` operation (no fallback needed)

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | None detected (this is a credential/configuration phase, not a code phase) |
| Config file | n/a |
| Quick run command | `security find-identity -v -p codesigning` (spot check) |
| Full suite command | Manual verification checklist (see below) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CRED-01 | Developer ID cert present and valid | smoke | `security find-identity -v -p codesigning \| grep "Developer ID Application: Ken Scott (S2K7P43927)"` | n/a — shell command |
| CRED-02 | `APPLE_CERTIFICATE` secret exists in release env | smoke | `gh secret list --env release --repo scottkw/storcat \| grep APPLE_CERTIFICATE` | n/a — shell command |
| CRED-03 | Windows cert vendor selected, RSA confirmed | manual | Manual vendor confirmation before purchase | n/a |
| CRED-04 | Windows signing secrets exist in release env | smoke | `gh secret list --env release --repo scottkw/storcat \| grep ES_USERNAME` | n/a — shell command |
| CRED-05 | All 9 secrets present in release environment | smoke | `gh secret list --env release --repo scottkw/storcat \| wc -l` (expect 9) | n/a — shell command |
| CRED-06 | Rotation runbook document exists | smoke | `ls docs/runbooks/credential-rotation.md` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** Run the smoke command for that task's requirement
- **Per wave merge:** `gh secret list --env release --repo scottkw/storcat` — confirm all 9 secrets listed
- **Phase gate:** All 9 secrets present + rotation runbook file exists

### Wave 0 Gaps

- [ ] `docs/runbooks/credential-rotation.md` — covers CRED-06 (created as part of this phase, not pre-existing)

Note: This phase creates no source code. Test validation is smoke-check CLI commands and file existence checks, not automated test suites.

---

## Project Constraints (from CLAUDE.md)

| Directive | Impact on This Phase |
|-----------|---------------------|
| GSD workflow enforcement: use `/gsd:execute-phase` entry point | All tasks must be executed via GSD workflow |
| No Electron dependencies: use Go/Wails patterns | Not applicable to credential phase |
| Package manager preference: `pnpm` for Node | Not applicable to credential phase |
| Python venv required | Not applicable |
| GitHub Actions: pin actions to SHAs | Any new GitHub Actions added in runbook or config should use pinned SHAs |

---

## Sources

### Primary (HIGH confidence)

- `security find-identity -v -p codesigning` — confirmed Developer ID Application: Ken Scott (S2K7P43927), expiry 2027-02-01
- `gh api repos/scottkw/storcat/environments` — confirmed 0 environments (release environment does not exist)
- [GitHub Docs: Installing an Apple certificate on macOS runners](https://docs.github.com/en/actions/use-cases-and-examples/deploying/installing-an-apple-certificate-on-macos-runners-for-xcode-development) — official secret names and process
- [Apple-Actions/import-codesign-certs](https://github.com/Apple-Actions/import-codesign-certs) — required inputs: `p12-file-base64`, `p12-password`

### Secondary (MEDIUM confidence)

- [Federico Terzi: Automatic Code-signing and Notarization for macOS apps using GitHub Actions](https://federicoterzi.com/blog/automatic-code-signing-and-notarization-for-macos-apps-using-github-actions/) — 7-secret macOS pattern verified against official docs
- [SSL.com: Cloud Code Signing Integration with GitHub Actions](https://www.ssl.com/how-to/cloud-code-signing-integration-with-github-actions/) — eSigner secret names: `ES_USERNAME`, `ES_PASSWORD`, `CREDENTIAL_ID`, `ES_TOTP_SECRET`
- [SSL.com: eSigner Pricing](https://www.ssl.com/guide/esigner-pricing-for-code-signing/) — OV Tier 1: $20/month, 20 signings
- [Melatonin: Code signing on Windows with Azure Trusted Signing](https://melatonin.dev/blog/code-signing-on-windows-with-azure-trusted-signing/) — Azure restriction to US/Canadian businesses confirmed
- [Hendrik Erz: Code Signing With Azure Trusted Signing on GitHub Actions](https://hendrik-erz.de/post/code-signing-with-azure-trusted-signing-on-github-actions) — individual dev identity verification requirements

### Tertiary (LOW confidence)

- [SSLcom/esigner-codesign GitHub Action](https://github.com/SSLcom/esigner-codesign) — action inputs confirmed but exact action version not pinned here (Phase 18 will verify current SHA)

---

## Metadata

**Confidence breakdown:**
- macOS certificate (CRED-01, CRED-02): HIGH — cert confirmed on keychain, official Apple docs for export process
- Apple notarization secrets (part of CRED-05): HIGH — official GitHub docs confirm secret names
- Windows cert vendor (CRED-03): MEDIUM — landscape confirmed, but final vendor selection is a user decision
- Windows eSigner secrets (CRED-04, part of CRED-05): MEDIUM — SSL.com docs confirm secret names; action version for Phase 18 not yet pinned
- GitHub environment creation (CRED-05): HIGH — API confirmed, environment absent, creation commands standard
- Rotation runbook (CRED-06): HIGH — content is documentation; structure is well-understood

**Research date:** 2026-03-27
**Valid until:** 2026-06-27 (stable domain — Apple cert processes and GitHub environments are not fast-moving; Windows vendor landscape could shift if Azure relaxes restrictions)
