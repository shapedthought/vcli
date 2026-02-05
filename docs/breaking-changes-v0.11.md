# Breaking Changes in v0.11.0

**Version:** v0.11.0
**Type:** Major breaking changes (clean break)
**Impact:** Configuration files must be regenerated
**Upgrade Time:** ~2 minutes

---

## Overview

vcli v0.11.0 introduces comprehensive improvements to authentication, security, and automation workflows. To achieve these goals, we made a **clean break** from the v0.10.x configuration format.

**Key principle:** Rather than maintaining backward compatibility and complex migration logic, we opted for simplicity: regenerate your configuration files in ~10 seconds. This approach delivers:

- **Simpler upgrade:** `vcli init profiles` → `vcli login` (done!)
- **Cleaner codebase:** No legacy format support to maintain
- **Better UX:** Clear error messages, one simple fix
- **Faster delivery:** Ship improvements immediately

---

## Summary of Breaking Changes

| Change | Impact | Migration |
|--------|--------|-----------|
| Non-interactive init | Scripts work unchanged; interactive users add `--interactive` flag | Update scripts if using interactive mode |
| Profile commands | Must pass arguments instead of prompting | Add profile name as argument |
| Secure token storage | System keychain instead of plaintext files | Just re-login after upgrade |
| profiles.json v1.0 | Completely new structure | Regenerate with `vcli init profiles` |
| Removed CredsFileMode | Credentials always from environment variables | Set environment variables |

---

## Breaking Changes in Detail

### 1. Non-Interactive Init by Default

**What Changed:** `vcli init` is now non-interactive by default and outputs JSON for piping/scripting.

**Before (v0.10.x):**
```bash
$ vcli init
Use creds file mode? (y/N): _
Allow insecure TLS? (y/N): _
```

**After (v0.11.0):**
```bash
$ vcli init
{
  "version": "1.0",
  "settings": {
    "selectedProfile": "vbr",
    "apiNotSecure": false
  },
  "profiles": { ... },
  "files": {
    "settings": "/Users/you/.vcli/settings.json",
    "profiles": "/Users/you/.vcli/profiles.json"
  }
}

Initialized successfully (profiles v1.0)
Ensure environment variables are set: VCLI_USERNAME, VCLI_PASSWORD, VCLI_URL

Available profiles: [vbr vb365 aws azure gcp vone ent_man]
Current profile: vbr
```

**For Interactive Mode:**
```bash
$ vcli init --interactive
Allow insecure TLS? (y/N): _
```

**Why This Change:**
- **Automation-first:** Default behavior now works in CI/CD without modifications
- **Scriptable:** JSON output can be piped to `jq` or processed by scripts
- **Opt-in prompts:** Interactive mode available via explicit flag
- **Better defaults:** Secure settings by default (TLS verification on)

**Migration Steps:**
1. **CI/CD scripts:** No changes needed (already non-interactive)
2. **Interactive users:** Add `--interactive` flag if you prefer prompts
3. **Scripts expecting prompts:** Update to use flags: `vcli init --insecure`

**Example Script Update:**
```bash
# Before (v0.10.x)
echo "n" | vcli init  # Pipe answers to prompts

# After (v0.11.0)
vcli init  # Non-interactive by default
# or
vcli init --insecure  # For lab environments
```

---

### 2. Profile Commands Now Require Arguments

**What Changed:** Profile management commands now accept arguments instead of prompting interactively.

**Before (v0.10.x):**
```bash
$ vcli profile --set
Enter profile name: vbr

$ vcli profile --list
[vbr, vb365, aws, azure, gcp, vone, ent_man]
```

**After (v0.11.0):**
```bash
$ vcli profile --set vbr
Successfully switched to profile: vbr

$ vcli profile --list
{
  "currentProfile": "vbr",
  "availableProfiles": ["vbr", "vb365", "aws", "azure", "gcp", "vone", "ent_man"]
}

$ vcli profile --list --table
Current Profile: vbr

Available Profiles:
  vbr
  vb365
  aws
  azure
  gcp
  vone
  ent_man
```

**Why This Change:**
- **Scriptable:** Commands can be used in automation without `expect` or input piping
- **CI/CD ready:** Profile switching works in non-interactive environments
- **Consistent UX:** Matches standard CLI patterns (arguments over prompts)
- **JSON by default:** Structured output for scripting, `--table` for humans

**Migration Steps:**
1. **Scripts:** Add profile name as argument to `--set` command
2. **Human users:** No change needed (you already typed the name at the prompt)
3. **Parsing output:** Update to parse JSON or use `--table` flag

**Example Script Update:**
```bash
# Before (v0.10.x)
echo "aws" | vcli profile --set

# After (v0.11.0)
vcli profile --set aws

# Parse JSON output
CURRENT_PROFILE=$(vcli profile --list | jq -r '.currentProfile')
```

---

### 3. Secure Token Storage (System Keychain)

**What Changed:** Authentication tokens now stored in system keychain instead of plaintext `headers.json` file.

**Before (v0.10.x):**
- Tokens stored in `~/.vcli/headers.json` (plaintext)
- File permissions: `0644` (world-readable)
- Manual token management required
- Tokens persist across reboots but exposed on disk

**After (v0.11.0):**
- Tokens stored in system keychain:
  - **macOS:** Keychain Access
  - **Windows:** Credential Manager
  - **Linux:** Secret Service (GNOME Keyring, KWallet)
- File permissions: `0600` (owner-only) for config files
- Auto-authentication in CI/CD environments (non-TTY detection)
- Explicit token control via `VCLI_TOKEN` environment variable

**Authentication Priority:**
```
1. VCLI_TOKEN environment variable (highest priority)
   ↓
2. System keychain (interactive sessions)
   ↓
3. Auto-authenticate with VCLI_USERNAME/VCLI_PASSWORD/VCLI_URL
   (CI/CD environments, non-TTY sessions)
```

**CI/CD Behavior:**
- Detects non-interactive sessions automatically
- Authenticates on-demand using environment variables
- No keychain interaction in CI/CD pipelines
- Works in GitHub Actions, GitLab CI, Jenkins, etc.

**Why This Change:**
- **Security:** No plaintext tokens on disk
- **OS integration:** Uses secure, encrypted system storage
- **CI/CD friendly:** Auto-detects environment, no keychain on headless systems
- **Explicit control:** `VCLI_TOKEN` for custom token management

**Migration Steps:**

**Interactive Users:**
```bash
# After upgrading to v0.11.0, just re-login
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"
vcli login
# Token now stored in system keychain
```

**CI/CD Pipelines:**
```yaml
# No changes needed - works automatically
- name: Run vcli command
  env:
    VCLI_USERNAME: ${{ secrets.VCLI_USERNAME }}
    VCLI_PASSWORD: ${{ secrets.VCLI_PASSWORD }}
    VCLI_URL: ${{ secrets.VCLI_URL }}
  run: |
    vcli get jobs
    # Auto-authenticates using environment variables
```

**Explicit Token Control:**
```bash
# Use specific token (e.g., service account)
export VCLI_TOKEN="your-long-lived-token"
vcli get jobs
# Bypasses keychain and auto-auth
```

**File-Based Keyring (Fallback):**

If system keychain is unavailable, vcli falls back to encrypted file storage:

```bash
# Set encryption password via environment variable
export VCLI_FILE_KEY="your-secure-password"
vcli login

# Or enter password interactively when prompted
vcli login
Enter password for vcli file keyring: _
```

**Security Improvements:**
- Old `headers.json`: World-readable plaintext tokens
- New keychain: OS-encrypted, access-controlled storage
- New file backend: Encrypted, `0600` permissions

---

### 4. profiles.json v1.0 Format

**What Changed:** Complete restructure of `profiles.json` with versioned format and proper multi-profile support.

**Before (v0.10.x):**
```json
{
  "name": "vbr",
  "port": "9419",
  "url": ":9419/api/oauth2/token",
  "api_version": "1.1-rev0",
  "x_api_version": "1.1-rev0",
  "accept": "application/json",
  "content-type": "application/json",
  "authUrl": "/api/oauth2/token",
  "apiUrl": "/api/v1",
  "username": "",
  "address": ""
}
```

**After (v0.11.0):**
```json
{
  "version": "1.0",
  "currentProfile": "vbr",
  "profiles": {
    "vbr": {
      "product": "VeeamBackupReplication",
      "apiVersion": "1.3-rev1",
      "port": 9419,
      "endpoints": {
        "auth": "/api/oauth2/token",
        "apiPrefix": "/api/v1"
      },
      "authType": "oauth",
      "headers": {
        "accept": "application/json",
        "contentType": "application/x-www-form-urlencoded",
        "xAPIVersion": "1.3-rev1"
      }
    },
    "vb365": {
      "product": "VeeamBackupFor365",
      "apiVersion": "v8",
      "port": 4443,
      "endpoints": {
        "auth": "/v8/Token",
        "apiPrefix": "/v8"
      },
      "authType": "oauth",
      "headers": {
        "accept": "application/json",
        "contentType": "application/x-www-form-urlencoded",
        "xAPIVersion": ""
      }
    }
  }
}
```

**Key Differences:**

| Aspect | v0.10.x | v0.11.0 |
|--------|---------|---------|
| **Format** | Flat single profile | Versioned with all profiles |
| **Version field** | None | `"version": "1.0"` |
| **Profile storage** | Single profile only | Map of all profiles |
| **Profile switching** | Replace entire file | Switch `currentProfile` field |
| **Port type** | String `"9419"` | Number `9419` |
| **Endpoints** | Flat fields | Nested `endpoints` object |
| **Headers** | Flat fields | Nested `headers` object |
| **Credentials** | Optional in file | Never in file (env vars only) |

**Why This Change:**
- **Versioning:** `version` field enables future compatibility detection
- **Multi-profile:** All profiles in one file, switch instantly
- **Structure:** Logical grouping of endpoints and headers
- **Type safety:** Proper types (number for port, not string)
- **Extensibility:** Easy to add new fields without breaking parsers
- **Clarity:** Product name and auth type explicit

**Migration Steps:**

**Option 1: Full Regeneration (Recommended)**
```bash
# Backup old config (optional)
cp -r ~/.vcli ~/.vcli.old

# Regenerate profiles (creates v1.0 format)
vcli init profiles

# Output:
# {
#   "version": "1.0",
#   "profiles": {
#     "vbr": { ... },
#     "vb365": { ... },
#     ...
#   },
#   "file": "/Users/you/.vcli/profiles.json"
# }

# Re-login
vcli login
```

**Option 2: Separate Profile and Settings**
```bash
# Generate only profiles
vcli init profiles

# Generate only settings
vcli init settings --insecure  # For lab environments
```

**Detection of Old Format:**

vcli detects v0.10.x format and provides clear guidance:

```bash
$ vcli get jobs
Error: Invalid profiles.json format (legacy v0.10.x detected)

Run 'vcli init profiles' to regenerate (takes ~5 seconds)

See: https://github.com/shapedthought/vcli/blob/master/UPGRADING.md
```

---

### 5. Removed CredsFileMode

**What Changed:** The `--creds-file` mode is completely removed. Credentials now **always** come from environment variables.

**Before (v0.10.x):**
```bash
# Environmental mode (credentials from env vars)
vcli init
Use creds file mode? (y/N): n

# OR

# Creds file mode (username/address in profiles.json)
vcli init
Use creds file mode? (y/N): y
```

**After (v0.11.0):**
```bash
# Only one mode: credentials from environment variables
vcli init

# Always requires:
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"
```

**settings.json Changes:**

```json
// Before (v0.10.x)
{
  "selectedProfile": "vbr",
  "apiNotSecure": false,
  "credsFileMode": true
}

// After (v0.11.0)
{
  "selectedProfile": "vbr",
  "apiNotSecure": false
}
```

**Why This Change:**
- **Simpler architecture:** One credential source instead of two
- **Security:** No credentials in config files
- **12-factor app principles:** Configuration via environment
- **CI/CD standard:** All CI systems use environment variables
- **Consistent UX:** Same credential method for all environments

**Migration Steps:**

**If using environmental mode (already env vars):**
```bash
# No changes needed - you're already using env vars
vcli init profiles
vcli login
```

**If using creds file mode (username/address in profiles.json):**
```bash
# 1. Extract credentials from old profiles.json
OLD_USERNAME=$(jq -r '.username' ~/.vcli.old/profiles.json)
OLD_ADDRESS=$(jq -r '.address' ~/.vcli.old/profiles.json)

# 2. Set environment variables
export VCLI_USERNAME="$OLD_USERNAME"
export VCLI_URL="$OLD_ADDRESS"
export VCLI_PASSWORD="your-password"  # You already had this

# 3. Regenerate configs
vcli init profiles
vcli login

# 4. Add to shell profile for persistence
echo 'export VCLI_USERNAME="administrator"' >> ~/.bashrc
echo 'export VCLI_URL="vbr.example.com"' >> ~/.bashrc
# Don't add password to .bashrc - use a secrets manager
```

**Best Practices for Credentials:**

**Development/Testing:**
```bash
# In terminal
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr-lab.local"

# Or in shell profile
echo 'export VCLI_USERNAME="administrator"' >> ~/.bashrc
echo 'export VCLI_URL="vbr-lab.local"' >> ~/.bashrc
# Password from pass, 1Password, etc.
```

**Production/CI/CD:**
```yaml
# GitHub Actions
env:
  VCLI_USERNAME: ${{ secrets.VCLI_USERNAME }}
  VCLI_PASSWORD: ${{ secrets.VCLI_PASSWORD }}
  VCLI_URL: ${{ secrets.VCLI_URL }}

# GitLab CI
variables:
  VCLI_USERNAME: $VCLI_USERNAME  # From CI/CD variables
  VCLI_PASSWORD: $VCLI_PASSWORD
  VCLI_URL: $VCLI_URL

# Azure DevOps
env:
  VCLI_USERNAME: $(VCLI_USERNAME)  # From pipeline variables
  VCLI_PASSWORD: $(VCLI_PASSWORD)
  VCLI_URL: $(VCLI_URL)
```

---

## Quick Upgrade Guide

### Prerequisites
- vcli v0.11.0 installed
- Access to VBR/product credentials

### Upgrade Steps (2 minutes)

#### 1. Optional: Backup Old Configs
```bash
cp -r ~/.vcli ~/.vcli.old
```

#### 2. Regenerate Profiles
```bash
vcli init profiles
```

Output:
```json
{
  "version": "1.0",
  "profiles": {
    "vbr": { ... },
    "vb365": { ... },
    ...
  },
  "file": "/Users/you/.vcli/profiles.json"
}
```

#### 3. Regenerate Settings (if needed)
```bash
vcli init settings          # Default settings
vcli init settings --insecure  # For lab environments
```

#### 4. Set Environment Variables
```bash
export VCLI_USERNAME="your-username"
export VCLI_PASSWORD="your-password"
export VCLI_URL="your-vbr-server"
```

#### 5. Re-Login
```bash
vcli login
```

#### 6. Test
```bash
vcli get jobs
vcli profile --list
```

Done! 🎉

---

## CI/CD Pipeline Migration

**Good news:** Most CI/CD pipelines require minimal or no changes.

### What Works Unchanged

✅ **Environment variables** - Already the standard in CI/CD
✅ **`vcli init`** - Already non-interactive by default
✅ **Commands** - All `vcli get/post/put` commands work identically
✅ **Exit codes** - No changes to success/failure codes

### What Needs Updating

⚠️ **Profile commands** - Add arguments instead of piping input

**Before (v0.10.x):**
```yaml
- name: Switch profile
  run: echo "aws" | vcli profile --set
```

**After (v0.11.0):**
```yaml
- name: Switch profile
  run: vcli profile --set aws
```

### Example Pipeline Updates

**GitHub Actions:**
```yaml
# No changes needed to this workflow
name: VBR Drift Detection
on:
  schedule:
    - cron: '0 */6 * * *'

jobs:
  drift-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Download vcli
        run: |
          wget https://github.com/shapedthought/vcli/releases/download/v0.11.0/vcli-linux
          chmod +x vcli-linux
          mv vcli-linux vcli

      - name: Check drift
        env:
          VCLI_USERNAME: ${{ secrets.VCLI_USERNAME }}
          VCLI_PASSWORD: ${{ secrets.VCLI_PASSWORD }}
          VCLI_URL: ${{ secrets.VCLI_URL }}
        run: |
          ./vcli job diff --all --security-only
          # Auto-authenticates using env vars
```

**GitLab CI:**
```yaml
# No changes needed
drift-check:
  stage: compliance
  script:
    - vcli job diff --all --security-only
    - vcli repo diff --all
  variables:
    VCLI_USERNAME: $VCLI_USERNAME
    VCLI_PASSWORD: $VCLI_PASSWORD
    VCLI_URL: $VCLI_URL
```

---

## Troubleshooting

### Error: Invalid profiles.json format

**Symptom:**
```
Error: Invalid profiles.json format (legacy v0.10.x detected)
Run 'vcli init profiles' to regenerate
```

**Solution:**
```bash
vcli init profiles
vcli login
```

---

### Error: Authentication failed

**Symptom:**
```
Authentication failed: VCLI_USERNAME or VCLI_URL environment variable not set
```

**Solution:**
```bash
# Check environment variables are set
echo $VCLI_USERNAME
echo $VCLI_URL
# (Don't echo password for security)

# If not set, export them
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"

# Then retry
vcli login
```

---

### Error: File keyring password required

**Symptom:**
```
Error: VCLI_FILE_KEY environment variable required for file keyring in non-interactive mode
```

**Context:** File keyring is used when system keychain is unavailable (e.g., headless Linux servers).

**Solution:**

**Interactive systems:**
```bash
# vcli will prompt for password
vcli login
Enter password for vcli file keyring: _
```

**CI/CD systems:**
```bash
# Set file keyring password via environment variable
export VCLI_FILE_KEY="your-secure-password"
vcli login
```

---

### Working directory issues

**Symptom:**
```
Error: settings.json not found
```

**Solution:**

Use `VCLI_SETTINGS_PATH` environment variable:

```bash
export VCLI_SETTINGS_PATH="$HOME/.vcli/"
vcli get jobs
```

Or run from directory containing config files:

```bash
cd ~/.vcli
vcli get jobs
```

---

### Profile commands return JSON instead of text

**Symptom:**
```bash
$ vcli profile --list
{"currentProfile":"vbr","availableProfiles":["vbr","vb365",...]}
```

**Solution:**

This is expected - JSON is now the default for scripting. For human-readable output:

```bash
vcli profile --list --table
Current Profile: vbr

Available Profiles:
  vbr
  vb365
  ...
```

---

## Why Clean Break?

We chose a clean break over backward compatibility for these reasons:

### Simpler for Users
- **10-second regeneration** vs understanding migration logic
- **Clear error messages** that tell you exactly what to do
- **One solution** that always works: regenerate configs

### Cleaner Codebase
- **No legacy format support** to maintain forever
- **No migration code** with edge case bugs
- **Simpler testing** - only one format to validate

### Better UX
- **Immediate error detection** - know right away if config is old
- **Clear path forward** - single fix that always works
- **No hidden state** - fresh start with best practices

### Faster Delivery
- **Ship improvements immediately** instead of delaying for migration logic
- **Focus on features** instead of backward compatibility
- **Iterate quickly** with version field enabling future upgrades

### Industry Standard
- Major tools (Terraform, Kubernetes, etc.) make clean breaks for major versions
- Users expect breaking changes in major version bumps (v0.10 → v0.11)
- Clear upgrade path is better than complex automatic migration

---

## API Version Updates

v0.11.0 also includes updated API versions for Veeam products:

| Product | v0.10.x | v0.11.0 |
|---------|---------|---------|
| VBR | 1.1-rev0 | 1.3-rev1 |
| VB365 | v6/v7 (mixed) | v8 |
| VB for AWS | 1.3-rev0 | 1.4-rev0 |
| VB for Azure | v4 | v5 |
| VB for GCP | 1.1-rev0 | 1.2-rev0 |
| VONE | 1.0-rev1 | 1.0-rev2 |

These updates are automatically included when you run `vcli init profiles`.

---

## Getting Help

- **Documentation:** [https://github.com/shapedthought/vcli](https://github.com/shapedthought/vcli)
- **Upgrade Guide:** [UPGRADING.md](../UPGRADING.md)
- **User Guide:** [user_guide.md](../user_guide.md)
- **Issues:** [GitHub Issues](https://github.com/shapedthought/vcli/issues)

---

## Summary

v0.11.0 delivers major improvements to security, automation, and usability through a clean break approach. The upgrade process is straightforward:

1. Regenerate configs: `vcli init profiles`
2. Re-login: `vcli login`
3. Done!

This approach prioritizes simplicity and clarity over complex migration logic, resulting in a better experience for all users.
