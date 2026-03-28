# Credential Rotation Runbook

## Overview

This runbook documents all 9 code signing secrets stored in the GitHub `release` environment
for the `scottkw/storcat` repository. Use this guide when any certificate expires, is revoked,
or is compromised.

**Environment:** `release` in `scottkw/storcat`
**Total secrets:** 9 (5 Apple + 4 Windows)

**When to use this runbook:**
- A certificate is approaching expiry (see expiry dates below)
- A certificate or credential is compromised
- SSL.com eSigner account credentials change
- Apple Developer account credentials are rotated

---

## Section 1: Apple Credentials (5 secrets)

| Secret | What It Contains | Expiry | Renewal Steps |
|--------|-----------------|--------|---------------|
| `APPLE_CERTIFICATE` | base64-encoded .p12 (Developer ID Application cert + private key) | **2027-02-01** | Re-export from Keychain Access: login keychain > My Certificates > select cert + private key > Export 2 Items as .p12 > `base64 -i file.p12 \| pbcopy` > `gh secret set APPLE_CERTIFICATE --env release --repo scottkw/storcat` |
| `APPLE_CERTIFICATE_PASSWORD` | Password used during .p12 export | Same as APPLE_CERTIFICATE (re-created on each export) | Set new password during .p12 export > `gh secret set APPLE_CERTIFICATE_PASSWORD --env release --repo scottkw/storcat` |
| `APPLE_CERTIFICATE_NAME` | `Developer ID Application: Ken Scott (S2K7P43927)` | Same as certificate | Only changes if identity name changes (unlikely); verify with `security find-identity -v -p codesigning` |
| `APPLE_NOTARIZATION_PASSWORD` | App-specific password for Apple ID | Does not expire unless revoked | Generate at appleid.apple.com > Sign-In and Security > App-Specific Passwords > name it "StorCat CI" > `gh secret set APPLE_NOTARIZATION_PASSWORD --env release --repo scottkw/storcat` |
| `APPLE_TEAM_ID` | `S2K7P43927` | Does not expire | Only changes if Apple Developer account is recreated |

### Apple Certificate Export (Detailed Steps)

> **Critical:** Export BOTH the certificate and its private key, or codesign will fail with "no private key found."

1. Open **Keychain Access.app**
2. Select the **login** keychain, **My Certificates** category
3. Expand `Developer ID Application: Ken Scott (S2K7P43927)` to see the private key
4. Click the **certificate row** to select it
5. **Cmd+click** the private key row beneath it (select BOTH items — "Export 2 Items")
6. Right-click → **"Export 2 Items..."**
7. Save as `storcat-dev-id.p12`
8. Enter a strong password — this becomes `APPLE_CERTIFICATE_PASSWORD`

```bash
# Encode and store the certificate
gh secret set APPLE_CERTIFICATE --env release --repo scottkw/storcat \
  --body "$(base64 -i storcat-dev-id.p12)"

# Store the password (interactive prompt)
gh secret set APPLE_CERTIFICATE_PASSWORD --env release --repo scottkw/storcat

# Store the certificate name (static)
gh secret set APPLE_CERTIFICATE_NAME --env release --repo scottkw/storcat \
  --body "Developer ID Application: Ken Scott (S2K7P43927)"

# Store the Team ID (static)
gh secret set APPLE_TEAM_ID --env release --repo scottkw/storcat \
  --body "S2K7P43927"

# Store app-specific password (interactive prompt)
gh secret set APPLE_NOTARIZATION_PASSWORD --env release --repo scottkw/storcat
```

### Generating a New App-Specific Password (APPLE_NOTARIZATION_PASSWORD)

1. Go to [appleid.apple.com](https://appleid.apple.com)
2. Sign in → **Sign-In and Security** → **App-Specific Passwords**
3. Click **+** → name it **"StorCat CI"** for traceability
4. Copy the generated password immediately (shown only once)
5. Run: `gh secret set APPLE_NOTARIZATION_PASSWORD --env release --repo scottkw/storcat`

---

## Section 2: Windows Credentials (4 secrets, SSL.com eSigner)

| Secret | What It Contains | Expiry | Renewal Steps |
|--------|-----------------|--------|---------------|
| `ES_USERNAME` | SSL.com account email | Does not expire | Update if SSL.com account email changes; `gh secret set ES_USERNAME --env release --repo scottkw/storcat` |
| `ES_PASSWORD` | SSL.com account password | Per SSL.com password policy | Update in both SSL.com and `gh secret set ES_PASSWORD --env release --repo scottkw/storcat` |
| `CREDENTIAL_ID` | Signing credential ID from SSL.com dashboard | When OV certificate expires (typically 1-3 years) | Purchase new OV cert > enroll in eSigner > copy new credential ID > `gh secret set CREDENTIAL_ID --env release --repo scottkw/storcat` |
| `ES_TOTP_SECRET` | TOTP seed for eSigner MFA | Same as CREDENTIAL_ID (re-enrollment generates new TOTP) | **CRITICAL:** shown only once during eSigner enrollment — save immediately or must re-enroll. During QR code display, extract the `secret=XXXXXX` parameter from the `otpauth://` URI. `gh secret set ES_TOTP_SECRET --env release --repo scottkw/storcat` |

### SSL.com eSigner Initial Enrollment

1. Create an account at [ssl.com](https://www.ssl.com)
2. Purchase an OV Code Signing certificate (confirm **RSA** — not ECDSA — before purchase)
3. Enroll in eSigner in the SSL.com dashboard
4. **During enrollment, when the QR code appears:**
   - Look for a text link to view the URI: `otpauth://totp/...?secret=XXXXXX`
   - Copy `XXXXXX` immediately — this is `ES_TOTP_SECRET`
   - This is shown **only once**; if lost, you must re-enroll
5. Copy the **Credential ID** from the SSL.com dashboard

```bash
# Store eSigner credentials
gh secret set ES_USERNAME --env release --repo scottkw/storcat
gh secret set ES_PASSWORD --env release --repo scottkw/storcat
gh secret set CREDENTIAL_ID --env release --repo scottkw/storcat
gh secret set ES_TOTP_SECRET --env release --repo scottkw/storcat
```

---

## Section 3: Verification Commands

```bash
# Check Apple cert validity on local machine
security find-identity -v -p codesigning | grep "Developer ID Application: Ken Scott"

# Check Apple cert expiry
security find-certificate -c "Developer ID Application: Ken Scott" -p | \
  openssl x509 -noout -dates
# Expected: notAfter=Feb  1 22:12:15 2027 GMT

# List all secrets in release environment (expect 9)
gh secret list --env release --repo scottkw/storcat

# Verify specific secrets exist (names only — values are hidden)
gh secret list --env release --repo scottkw/storcat | grep APPLE_CERTIFICATE
gh secret list --env release --repo scottkw/storcat | grep ES_USERNAME
gh secret list --env release --repo scottkw/storcat | grep CREDENTIAL_ID
gh secret list --env release --repo scottkw/storcat | grep ES_TOTP_SECRET

# Verify environment exists with correct tag policy
gh api repos/scottkw/storcat/environments/release --jq '.name'
gh api repos/scottkw/storcat/environments/release/deployment-branch-policies \
  --jq '.branch_policies[0].name'
```

---

## Section 4: Emergency Procedures

### Apple Certificate Compromised

1. Revoke the certificate at [developer.apple.com](https://developer.apple.com) → Certificates
2. Request a new Developer ID Application certificate
3. Download and install the new certificate in Keychain Access
4. Export as .p12 following the steps in Section 1
5. Update both `APPLE_CERTIFICATE` and `APPLE_CERTIFICATE_PASSWORD` in GitHub secrets
6. Verify: `security find-identity -v -p codesigning`

### SSL.com Account or Certificate Compromised

1. Log into the SSL.com dashboard immediately
2. Revoke the certificate
3. Contact SSL.com support if the account is compromised
4. Once resolved, re-enroll in eSigner and update all 4 `ES_*` secrets:
   - `ES_USERNAME`, `ES_PASSWORD`, `CREDENTIAL_ID`, `ES_TOTP_SECRET`

### Secret Accidentally Exposed in Logs

If any secret value appears in GitHub Actions logs:

1. **Immediately** rotate the exposed credential using the renewal steps in Sections 1 or 2
2. Review GitHub Actions logs for all recent runs — ensure the value is redacted (GitHub auto-redacts registered secrets, but an unregistered value would not be redacted)
3. Check if the value appeared in any artifacts or releases
4. Update the GitHub secret with the new value

---

## Section 5: Future Improvements

- **SIGN-07:** Migrate to App Store Connect API key for notarization (more robust than app-specific password; supports key-based auth without 2FA dependency)
- **WSIGN-05:** Evaluate Azure Artifact Signing if business eligibility is confirmed (currently restricted to US/Canadian businesses with 3+ years of history as of April 2025)
- **SIGN-09:** Scheduled GitHub Actions workflow to monitor certificate expiry dates and alert before they lapse

---

## Quick Reference: All 9 Secrets

| # | Secret Name | Platform | Expires |
|---|-------------|----------|---------|
| 1 | `APPLE_CERTIFICATE` | macOS | 2027-02-01 |
| 2 | `APPLE_CERTIFICATE_PASSWORD` | macOS | 2027-02-01 |
| 3 | `APPLE_CERTIFICATE_NAME` | macOS | 2027-02-01 |
| 4 | `APPLE_NOTARIZATION_PASSWORD` | macOS | No expiry (revocable) |
| 5 | `APPLE_TEAM_ID` | macOS | No expiry |
| 6 | `ES_USERNAME` | Windows | No expiry |
| 7 | `ES_PASSWORD` | Windows | Per SSL.com policy |
| 8 | `CREDENTIAL_ID` | Windows | With OV cert (1-3 years) |
| 9 | `ES_TOTP_SECRET` | Windows | With OV cert (1-3 years) |
