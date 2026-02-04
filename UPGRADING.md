# Upgrading vcli

This guide covers upgrading between major versions of vcli.

---

## v0.10.x → v0.11.0 (Breaking Changes)

⚠️ **Clean break:** Old configurations are not compatible with v0.11.0.

### Quick Upgrade (~2 minutes)

#### 1. Backup Old Configs (Optional)
```bash
cp -r ~/.vcli ~/.vcli.old
```

#### 2. Regenerate Profiles
```bash
vcli init-profiles
```

**Output:**
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
vcli init-settings          # Default: secure settings
vcli init-settings --insecure  # For lab environments
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

**Output:**
```
Successfully authenticated
Token stored in system keychain
Profile: vbr
```

#### 6. Test
```bash
vcli get jobs
vcli profile --list
```

✅ **Done!**

---

## What Changed in v0.11.0?

### Summary of Breaking Changes

| Change | Impact |
|--------|--------|
| **Non-interactive init** | `vcli init` is non-interactive by default (outputs JSON) |
| **Profile commands** | Now take arguments: `vcli profile --set vbr` |
| **Secure token storage** | Tokens in system keychain instead of plaintext files |
| **profiles.json v1.0** | Completely new structure (must regenerate) |
| **Removed CredsFileMode** | Credentials always from environment variables |

### Why Clean Break?

- **Simpler:** Regenerate configs in 10 seconds vs complex migration
- **Cleaner:** No legacy format support to maintain
- **Better UX:** Clear errors, one simple fix
- **Faster:** Ship improvements immediately

---

## Upgrade Scenarios

### Interactive User (Local Development)

```bash
# You're working locally and run vcli manually
cd ~/.vcli
cp -r . ../vcli.old  # Backup

vcli init-profiles
vcli init-settings

# Set credentials (add to ~/.bashrc for persistence)
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr-lab.local"

vcli login
vcli get jobs  # Test
```

---

### CI/CD Pipeline

**Good news:** Most pipelines need minimal or no changes.

✅ **No changes needed:**
- Environment variables (already standard)
- `vcli init` (already non-interactive)
- `vcli get/post/put` commands

⚠️ **Update profile commands:**

**Before:**
```yaml
- run: echo "aws" | vcli profile --set
```

**After:**
```yaml
- run: vcli profile --set aws
```

**Full Example (GitHub Actions):**
```yaml
name: Drift Detection
on:
  schedule:
    - cron: '0 */6 * * *'

jobs:
  check-drift:
    runs-on: ubuntu-latest
    steps:
      - name: Download vcli v0.11.0
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
          # Auto-authenticates using environment variables
```

---

### Docker Container

```dockerfile
FROM ubuntu:latest

# Install vcli v0.11.0
RUN wget https://github.com/shapedthought/vcli/releases/download/v0.11.0/vcli-linux && \
    chmod +x vcli-linux && \
    mv vcli-linux /usr/local/bin/vcli

# Initialize profiles (one-time setup)
RUN vcli init-profiles

# At runtime, pass credentials via environment variables
# docker run -e VCLI_USERNAME=admin -e VCLI_PASSWORD=pass -e VCLI_URL=vbr.local ...
```

---

## Troubleshooting

### Error: Invalid profiles.json format

```
Error: Invalid profiles.json format (legacy v0.10.x detected)
Run 'vcli init-profiles' to regenerate
```

**Fix:**
```bash
vcli init-profiles
vcli login
```

---

### Error: Authentication failed

```
Authentication failed: VCLI_USERNAME or VCLI_URL environment variable not set
```

**Fix:**
```bash
# Check environment variables
echo $VCLI_USERNAME
echo $VCLI_URL

# Set if missing
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"

# Retry
vcli login
```

---

### File keyring password prompt (Linux/headless)

```
Enter password for vcli file keyring: _
```

**Context:** File keyring is used when system keychain unavailable.

**Interactive fix:**
```bash
# Enter password at prompt
vcli login
Enter password for vcli file keyring: your-secure-password
```

**CI/CD fix:**
```bash
# Set file keyring password via environment variable
export VCLI_FILE_KEY="your-secure-password"
vcli login
```

---

### Profile commands return JSON

```bash
$ vcli profile --list
{"currentProfile":"vbr","availableProfiles":["vbr","vb365",...]}
```

**Expected behavior:** JSON is now default for scripting.

**For human-readable output:**
```bash
vcli profile --list --table
Current Profile: vbr

Available Profiles:
  vbr
  vb365
  aws
  ...
```

---

## Detailed Documentation

For comprehensive information about breaking changes, see:

📖 [Breaking Changes in v0.11.0](docs/breaking-changes-v0.11.md)

Topics covered:
- Detailed explanation of each breaking change
- Before/after examples
- Migration strategies
- CI/CD pipeline updates
- Why clean break was chosen
- API version updates

---

## Downgrading (Not Recommended)

If you need to temporarily use v0.10.x:

```bash
# Restore old configs
cp -r ~/.vcli.old ~/.vcli

# Download v0.10.0
wget https://github.com/shapedthought/vcli/releases/download/v0.10.0-beta1/vcli-linux
chmod +x vcli-linux

# Use old version
./vcli-linux get jobs
```

**Warning:** v0.10.x is no longer maintained. Upgrade to v0.11.0 as soon as possible.

---

## Getting Help

- **Documentation:** [GitHub Repository](https://github.com/shapedthought/vcli)
- **Breaking Changes:** [docs/breaking-changes-v0.11.md](docs/breaking-changes-v0.11.md)
- **User Guide:** [user_guide.md](user_guide.md)
- **Issues:** [GitHub Issues](https://github.com/shapedthought/vcli/issues)
- **Examples:** [docs/](docs/) directory

---

## Next Steps

After upgrading:

1. ✅ Test basic operations: `vcli get jobs`
2. ✅ Update any scripts that use profile commands
3. ✅ Update CI/CD pipelines if needed
4. ✅ Review [breaking changes documentation](docs/breaking-changes-v0.11.md)
5. ✅ Explore new features (secure token storage, improved automation)

Welcome to v0.11.0! 🎉
