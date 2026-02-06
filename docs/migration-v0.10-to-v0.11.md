# Migration Guide: v0.10.x to v0.11.0

**Impact:** Configuration files must be regenerated (clean break)
**Upgrade Time:** ~2 minutes

---

## Quick Upgrade

### 1. Backup Old Configs (Optional)
```bash
cp -r ~/.vcli ~/.vcli.old
```

### 2. Regenerate Profiles
```bash
vcli init profiles
```

### 3. Regenerate Settings (if needed)
```bash
vcli init settings          # Default: secure settings
vcli init settings --insecure  # For lab environments
```

### 4. Set Environment Variables
```bash
export VCLI_USERNAME="your-username"
export VCLI_PASSWORD="your-password"
export VCLI_URL="your-vbr-server"
```

### 5. Re-Login
```bash
vcli login
```

### 6. Test
```bash
vcli get jobs
vcli profile --list
```

---

## Breaking Changes Summary

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

`vcli init` is now non-interactive by default and outputs JSON for piping/scripting.

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
  "settings": { ... },
  "profiles": { ... },
  "files": { ... }
}
```

**For interactive mode:**
```bash
$ vcli init --interactive
```

**Script update example:**
```bash
# Before (v0.10.x)
echo "n" | vcli init  # Pipe answers to prompts

# After (v0.11.0)
vcli init  # Non-interactive by default
# or
vcli init --insecure  # For lab environments
```

### 2. Profile Commands Require Arguments

**Before (v0.10.x):**
```bash
$ vcli profile --set
Enter profile name: vbr
```

**After (v0.11.0):**
```bash
$ vcli profile --set vbr

$ vcli profile --list
{"currentProfile":"vbr","availableProfiles":["vbr","vb365",...]}

$ vcli profile --list --table   # Human-readable format
```

### 3. Secure Token Storage (System Keychain)

Tokens now stored in system keychain instead of plaintext `headers.json`:
- **macOS:** Keychain Access
- **Windows:** Credential Manager
- **Linux:** Secret Service (GNOME Keyring, KWallet)

**Authentication priority:**
```
1. VCLI_TOKEN environment variable (highest priority)
2. System keychain (interactive sessions)
3. Auto-authenticate with VCLI_USERNAME/VCLI_PASSWORD/VCLI_URL (CI/CD)
```

**File-based keyring fallback** (for systems without keychain):
```bash
# CI/CD environments
export VCLI_FILE_KEY="your-secure-password"
vcli login

# Interactive systems
vcli login
# Prompts: Enter password for vcli file keyring: _
```

### 4. profiles.json v1.0 Format

Complete restructure with versioned format and multi-profile support.

**Before (v0.10.x):** Flat single-profile file
**After (v0.11.0):** Versioned map of all profiles

```json
{
  "version": "1.0",
  "currentProfile": "vbr",
  "profiles": {
    "vbr": {
      "product": "VeeamBackupReplication",
      "apiVersion": "1.3-rev1",
      "port": 9419,
      "endpoints": { "auth": "/api/oauth2/token", "apiPrefix": "/api/v1" },
      "authType": "oauth",
      "headers": { ... }
    }
  }
}
```

**Migration:** `vcli init profiles`

### 5. Removed CredsFileMode

Credentials now always come from environment variables. No more `--creds-file` option.

```bash
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"
```

---

## Upgrade Scenarios

### Interactive User (Local Development)

```bash
cp -r ~/.vcli ~/.vcli.old   # Backup
vcli init profiles
vcli init settings
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr-lab.local"
vcli login
vcli get jobs  # Test
```

### CI/CD Pipeline

Most pipelines need minimal or no changes.

**No changes needed:**
- Environment variables (already standard)
- `vcli init` (already non-interactive)
- `vcli get/post/put` commands

**Update profile commands:**
```yaml
# Before
- run: echo "aws" | vcli profile --set

# After
- run: vcli profile --set aws
```

**GitHub Actions example:**
```yaml
name: Drift Detection
on:
  schedule:
    - cron: '0 */6 * * *'

jobs:
  check-drift:
    runs-on: ubuntu-latest
    steps:
      - name: Download vcli
        run: |
          wget https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64
          chmod +x vcli-linux-amd64
          mv vcli-linux-amd64 vcli

      - name: Check drift
        env:
          VCLI_USERNAME: ${{ secrets.VCLI_USERNAME }}
          VCLI_PASSWORD: ${{ secrets.VCLI_PASSWORD }}
          VCLI_URL: ${{ secrets.VCLI_URL }}
        run: ./vcli job diff --all --security-only
```

### Docker Container

```dockerfile
FROM ubuntu:latest
RUN wget https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 && \
    chmod +x vcli-linux-amd64 && \
    mv vcli-linux-amd64 /usr/local/bin/vcli
RUN vcli init profiles
# At runtime, pass credentials via environment variables
```

---

## API Version Updates

v0.11.0 includes updated API versions:

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

## Testing Your Migration

### Step 1: Test Locally

```bash
./vcli init --output-dir /tmp/vcli-test/
ls -la /tmp/vcli-test/
VCLI_SETTINGS_PATH=/tmp/vcli-test/ ./vcli profile --set vbr
VCLI_SETTINGS_PATH=/tmp/vcli-test/ ./vcli profile --get
rm -rf /tmp/vcli-test/
```

### Step 2: Test in CI/CD

```yaml
- job: Test_Migration
  steps:
    - script: |
        ./vcli init --output-dir .vcli/
        TEST_PROFILE=$(VCLI_SETTINGS_PATH=.vcli/ ./vcli profile --set vbr && VCLI_SETTINGS_PATH=.vcli/ ./vcli profile --get)
        if [ "$TEST_PROFILE" != "vbr" ]; then
          echo "Migration test failed"
          exit 1
        fi
        echo "Migration test passed"
```

---

## Rollback Plan

### Option 1: Use Legacy Mode

```bash
./vcli init --interactive
```

### Option 2: Downgrade to v0.10.x

```bash
cp -r ~/.vcli.old ~/.vcli
wget https://github.com/shapedthought/vcli/releases/download/v0.10.0-beta1/vcli-linux
chmod +x vcli-linux
./vcli-linux get jobs
```

**Warning:** v0.10.x is no longer maintained.

---

## Troubleshooting

### Error: Invalid profiles.json format

```
Error: Invalid profiles.json format (legacy v0.10.x detected)
Run 'vcli init profiles' to regenerate
```

**Fix:** `vcli init profiles && vcli login`

### Error: Authentication failed

```
Authentication failed: VCLI_USERNAME or VCLI_URL environment variable not set
```

**Fix:**
```bash
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"
vcli login
```

### File keyring password prompt (Linux/headless)

```
Enter password for vcli file keyring: _
```

**CI/CD fix:** `export VCLI_FILE_KEY="your-secure-password"`

### Profile commands return JSON

This is expected. For human-readable output: `vcli profile --list --table`

---

## Getting Help

- **Documentation:** [GitHub Repository](https://github.com/shapedthought/vcli)
- **User Guide:** [user_guide.md](../user_guide.md)
- **Issues:** [GitHub Issues](https://github.com/shapedthought/vcli/issues)
