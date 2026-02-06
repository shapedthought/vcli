# Authentication and Setup

This guide covers initializing vcli, managing profiles, and setting up authentication for all Veeam products.

## Table of Contents

- [First-Time Setup](#first-time-setup)
- [Profiles](#profiles)
- [Authentication Modes](#authentication-modes)
- [Logging In](#logging-in)
- [Switching Between Products](#switching-between-products)
- [Troubleshooting](#troubleshooting)

## First-Time Setup

### Initialize vcli

Init is non-interactive by default for GitOps/CI/CD compatibility.

Run the init command to create configuration files:

```bash
# Non-interactive init (default)
./vcli init

# With specific settings
./vcli init --insecure --output-dir ~/.vcli/

# Legacy interactive mode
./vcli init --interactive
```

**Available Flags:**
- `--insecure` - Skip TLS verification (sets `apiNotSecure: true`)
- `--output-dir <path>` - Specify config file directory
- `--interactive` - Use legacy interactive prompts

**Files created:**
- `settings.json` - vcli settings and preferences
- `profiles.json` - API profiles for each Veeam product (v1.0 format)

**Output:** Init now outputs JSON to stdout for automation/piping:
```json
{
  "version": "1.0",
  "settings": {...},
  "profiles": {...},
  "files": {
    "settings": "path/to/settings.json",
    "profiles": "path/to/profiles.json"
  }
}
```

### Configuration File Location

By default, files are created in the current directory. To use a custom location:

**Set before running init:**
```bash
# Bash/Zsh
export VCLI_SETTINGS_PATH="$HOME/.vcli/"
mkdir -p ~/.vcli
./vcli init

# PowerShell
$env:VCLI_SETTINGS_PATH = "$HOME\.vcli\"
New-Item -ItemType Directory -Path $HOME\.vcli -Force
.\vcli init
```

**Example paths:**
- Windows: `C:\Users\UserName\.vcli\`
- macOS/Linux: `/home/username/.vcli/`

**Note:** The directory must exist before running `init`.

## Profiles

Profiles define connection settings for each Veeam product. Each profile contains:
- API version headers
- Port numbers
- Authentication URL patterns
- Optional username and address (creds file mode only)

### Available Profiles

| Profile | Product | Port | Notes |
|---------|---------|------|-------|
| `vbr` | Veeam Backup & Replication | 9419 | OAuth authentication |
| `ent_man` | Enterprise Manager | 9398 | Basic authentication |
| `vb365` | Veeam Backup for Microsoft 365 | 4443 | OAuth authentication |
| `vone` | Veeam ONE | 1239 | OAuth authentication |
| `aws` | Veeam Backup for AWS | 11005 | OAuth authentication |
| `azure` | Veeam Backup for Azure | 443 | OAuth authentication |
| `gcp` | Veeam Backup for GCP | 13140 | OAuth authentication |

### Profile Commands

Profile commands require explicit arguments (no interactive prompts).

```bash
# List all available profiles
./vcli profile --list
./vcli profile -l

# Get current active profile (returns clean output: "vbr")
./vcli profile --get
./vcli profile -g

# Set active profile (requires explicit argument)
./vcli profile --set vbr
./vcli profile -s vbr

# View profile details
./vcli profile --profile vbr
./vcli profile -p vbr
```

**Breaking Change:** `./vcli profile --set` now requires an argument. Previously it prompted interactively.

### Profile Structure (v1.0 Format)

profiles.json uses versioned format with all profiles in one file.

Example structure:

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
    "ent_man": {
      "product": "EnterpriseManager",
      "apiVersion": "",
      "port": 9398,
      "endpoints": {
        "auth": "/api/sessionMngr/?v=latest",
        "apiPrefix": "/api"
      },
      "authType": "basic",
      "headers": {
        "accept": "application/json",
        "contentType": "application/json",
        "xAPIVersion": ""
      }
    }
  }
}
```

**Note:** Credentials are never stored in profiles.json. Always use environment variables.

### API Versions

Current default versions (as of October 2023):

| Product | Version | API Version |
|---------|---------|-------------|
| VBR | 13.0 | 1.3-rev1 |
| Enterprise Manager | 12.0 | - |
| VONE | 12.0 | v2.1 |
| VB365 | 7.0 | - |
| VB for AWS | 5.0 | 1.4-rev0 |
| VB for Azure | 5.0 | - |
| VB for GCP | 1.0 | 1.2-rev0 |

**To change API version:** Edit the `profiles.json` file and update the `api_version` or `x-api-version` field.

**Veeam API documentation:** https://www.veeam.com/documentation-guides-datasheets.html

## Credentials and Token Storage

Credentials always come from environment variables. Tokens are stored securely in the system keychain.

### Setting Credentials

vcli requires three environment variables for authentication:

**Bash/Zsh (macOS/Linux):**
```bash
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"
export VCLI_SETTINGS_PATH="$HOME/.vcli/"  # Optional
```

**PowerShell (Windows):**
```powershell
$env:VCLI_USERNAME = "administrator"
$env:VCLI_PASSWORD = "your-password"
$env:VCLI_URL = "vbr.example.com"
$env:VCLI_SETTINGS_PATH = "$HOME\.vcli\"  # Optional
```

**Environment Variables:**

| Variable | Required | Description |
|----------|----------|-------------|
| `VCLI_USERNAME` | Yes* | API username (may need `DOMAIN\user` format) |
| `VCLI_PASSWORD` | Yes* | API password |
| `VCLI_URL` | Yes* | Server hostname/IP (without `https://` or port) |
| `VCLI_TOKEN` | No | Explicit authentication token (bypasses auto-auth) |
| `VCLI_SETTINGS_PATH` | No | Config file directory (default: current directory) |
| `VCLI_FILE_KEY` | No | File keyring password (for non-interactive systems) |

*Not required if `VCLI_TOKEN` is set

### Secure Token Storage

Authentication tokens are stored in your system's secure keychain instead of plaintext files.

#### Token Resolution Priority

vcli attempts authentication in this order:

1. **`VCLI_TOKEN` environment variable** (highest priority)
   - For explicit token control
   - Useful for service accounts or long-lived tokens
   ```bash
   export VCLI_TOKEN="your-long-lived-token"
   ./vcli get jobs
   ```

2. **System keychain** (interactive sessions only)
   - **macOS:** Keychain Access
   - **Windows:** Credential Manager
   - **Linux:** Secret Service (GNOME Keyring, KWallet)
   - Token stored encrypted by OS
   - Persists across vcli sessions
   ```bash
   ./vcli login  # Token stored in keychain
   ./vcli get jobs  # Uses token from keychain
   ```

3. **Auto-authenticate** (CI/CD and non-TTY environments)
   - Detects non-interactive sessions automatically
   - Authenticates on-demand using `VCLI_USERNAME`/`VCLI_PASSWORD`/`VCLI_URL`
   - No keychain interaction on headless systems
   ```bash
   # GitHub Actions, GitLab CI, Jenkins, etc.
   ./vcli get jobs  # Auto-authenticates using env vars
   ```

#### File-Based Keyring (Fallback)

If system keychain is unavailable, vcli uses encrypted file storage (`~/.vcli/vcli-keyring`):

**Interactive systems:**
```bash
./vcli login
# Prompts: Enter password for vcli file keyring: _
```

**CI/CD systems:**
```bash
export VCLI_FILE_KEY="your-secure-password"
./vcli login  # Uses VCLI_FILE_KEY for encryption
```

### Authentication Flow

**Interactive Session (Local Development):**
```bash
# 1. Set credentials
export VCLI_USERNAME="admin"
export VCLI_PASSWORD="pass"
export VCLI_URL="vbr.local"

# 2. Set profile
./vcli profile --set vbr

# 3. Login (stores token in keychain)
./vcli login
# Token stored in system keychain

# 4. Make API calls (uses keychain token)
./vcli get jobs
./vcli post jobs/<id>/start
```

**CI/CD Pipeline:**
```yaml
# GitHub Actions example
- name: Check VBR drift
  env:
    VCLI_USERNAME: ${{ secrets.VBR_USERNAME }}
    VCLI_PASSWORD: ${{ secrets.VBR_PASSWORD }}
    VCLI_URL: ${{ secrets.VBR_URL }}
  run: |
    ./vcli profile --set vbr
    ./vcli job diff --all --security-only
    # Auto-authenticates, no keychain interaction
```

**Explicit Token Control:**
```bash
# Use specific token
export VCLI_TOKEN="your-service-account-token"
./vcli get jobs  # Bypasses keychain and auto-auth
```

## Logging In

After setting up credentials and selecting a profile, authenticate:

```bash
./vcli login
```

**On success:**
- Stores authentication token in system keychain (interactive) or authenticates on-demand (CI/CD)
- Token used for all subsequent API calls
- Token valid until expiration

**Interactive Sessions:**
- Token stored in system keychain
- Persists across terminal sessions
- Encrypted by operating system

**CI/CD Environments:**
- Auto-authenticates using environment variables
- No keychain storage (non-TTY detection)
- Authenticates on-demand for each command

### Authentication Types by Product

**OAuth (most products):**
- VBR, VB365, VONE, AWS, Azure, GCP
- Uses Bearer token
- Token stored in system keychain

**Basic Auth (Enterprise Manager only):**
- Enterprise Manager uses session-based authentication
- Session token in `X-RestSvcSessionId` header
- Different URL pattern than other products

## Switching Between Products

### Method 1: Switch Credentials and Profile

```bash
# Connect to VBR
export VCLI_USERNAME="admin" VCLI_PASSWORD="pass1" VCLI_URL="vbr.example.com"
./vcli profile --set vbr
./vcli login
./vcli get jobs

# Switch to VB365
export VCLI_USERNAME="admin" VCLI_PASSWORD="pass2" VCLI_URL="vb365.example.com"
./vcli profile --set vb365
./vcli login
./vcli get organizations
```

Each profile gets its own token in the keychain (keyed by profile name).

### Method 2: Multiple Config Directories

Use separate directories for different environments:

```bash
# Production VBR
export VCLI_SETTINGS_PATH="$HOME/.vcli/prod/"
export VCLI_USERNAME="prod-admin" VCLI_PASSWORD="prod-pass" VCLI_URL="vbr-prod.local"
./vcli profile --set vbr
./vcli login

# Development VBR
export VCLI_SETTINGS_PATH="$HOME/.vcli/dev/"
export VCLI_USERNAME="dev-admin" VCLI_PASSWORD="dev-pass" VCLI_URL="vbr-dev.local"
./vcli profile --set vbr
./vcli login
```

### Method 3: Shell Functions for Quick Switching

```bash
# Add to ~/.bashrc or ~/.zshrc
vbr-prod() {
    export VCLI_USERNAME="prod-admin"
    export VCLI_PASSWORD="$VBR_PROD_PASS"
    export VCLI_URL="vbr-prod.local"
    ./vcli profile --set vbr
    ./vcli login
}

vbr-dev() {
    export VCLI_USERNAME="dev-admin"
    export VCLI_PASSWORD="$VBR_DEV_PASS"
    export VCLI_URL="vbr-dev.local"
    ./vcli profile --set vbr
    ./vcli login
}

# Usage
vbr-prod
./vcli get jobs

vbr-dev
./vcli get jobs
```

## Troubleshooting

### "Profile not set" Error

**Problem:** You tried to login without setting a profile.

**Solution:**
```bash
./vcli profile --set vbr
./vcli login
```

### "Authentication failed" or 401 Unauthorized

**Possible causes:**
1. Incorrect username or password
2. Wrong username format
3. Account locked or disabled
4. REST API not enabled

**Solutions:**

**Try different username formats:**
```bash
# Standard
export VCLI_USERNAME="administrator"

# Domain user (format 1)
export VCLI_USERNAME="DOMAIN\\administrator"

# Domain user (format 2)
export VCLI_USERNAME="administrator@domain.com"
```

**Check VBR REST API is enabled:**
1. Open VBR Console
2. Menu → Options → Network Traffic Rules
3. Ensure "RESTful API service" is checked

### "Connection refused" or "Dial tcp" Errors

**Possible causes:**
1. Wrong hostname/IP
2. Firewall blocking port
3. REST API service not running
4. Wrong profile selected

**Solutions:**

**Verify connectivity:**
```bash
# Test network connection
ping vbr.example.com

# Test port is open (requires telnet/nc)
telnet vbr.example.com 9419
# or
nc -zv vbr.example.com 9419
```

**Verify profile and URL:**
```bash
# Check current profile
./vcli profile --get

# Verify VCLI_URL doesn't include protocol or port
echo $VCLI_URL
# Should be: vbr.example.com
# NOT: https://vbr.example.com:9419
```

### TLS Certificate Errors

**Problem:** `x509: certificate signed by unknown authority`

**Option 1: Skip TLS verification (NOT recommended for production)**

Edit `settings.json`:
```json
{
  "skipTLSVerify": true
}
```

**Option 2: Trust the CA certificate (recommended)**

See [Getting Started Guide - Troubleshooting](getting-started.md#tls-certificate-errors) for detailed instructions.

### Files Created in Wrong Directory

**Problem:** `settings.json` and `profiles.json` not in expected location.

**Solution:**

1. Set `VCLI_SETTINGS_PATH` before running `init`:
```bash
export VCLI_SETTINGS_PATH="$HOME/.vcli/"
mkdir -p ~/.vcli
./vcli init
```

2. Or move files manually:
```bash
mkdir -p ~/.vcli
mv settings.json profiles.json ~/.vcli/
export VCLI_SETTINGS_PATH="$HOME/.vcli/"
```

### Token Expired

**Problem:** API calls failing with authentication errors after working previously.

**Solution:** Re-login to refresh the token in keychain:
```bash
./vcli login
```

The keychain token is automatically refreshed.

### Token Storage Issues

**Problem:** `failed to open keyring` or `no authentication method available`

**Context:** System keychain unavailable (headless Linux, Docker, etc.)

**Solution:**

**Option 1: Use explicit token**
```bash
export VCLI_TOKEN="your-long-lived-token"
./vcli get jobs
```

**Option 2: Set file keyring password**
```bash
# CI/CD environments
export VCLI_FILE_KEY="your-secure-password"
./vcli login

# Interactive systems
./vcli login
# Prompts for file keyring password
```

### CI/CD Auto-Authentication Not Working

**Problem:** Commands fail in CI/CD pipeline

**Cause:** Missing environment variables

**Solution:** Verify all required variables are set:
```yaml
# GitHub Actions
env:
  VCLI_USERNAME: ${{ secrets.VBR_USERNAME }}
  VCLI_PASSWORD: ${{ secrets.VBR_PASSWORD }}
  VCLI_URL: ${{ secrets.VBR_URL }}
```

Check environment variables in pipeline:
```bash
echo "Username: $VCLI_USERNAME"
echo "URL: $VCLI_URL"
# Don't echo password for security
```

## Best Practices

### Security

1. **Environment variables for credentials** - Never store credentials in configuration files
2. **Use secrets management** in CI/CD - Azure Key Vault, GitHub Secrets, AWS Secrets Manager, etc.
3. **Rotate credentials regularly** - Especially for service accounts
4. **Use dedicated service accounts** for automation - Separate accounts for CI/CD vs interactive use
5. **Trust system keychain** - Let vcli manage token storage securely
6. **Use `VCLI_TOKEN` for long-lived tokens** - Service accounts with explicit token control

**What to commit to Git:**
- ✅ `settings.json` (no secrets)
- ✅ `profiles.json` (no credentials in v1.0 format)
- ✅ Configuration YAML files
- ❌ Never commit tokens or credentials
- ❌ state.json (optional - depends on workflow)

### Organization

1. **Use custom settings path** to keep config centralized:
   ```bash
   export VCLI_SETTINGS_PATH="$HOME/.vcli/"
   ```

2. **Name config directories** clearly for multiple environments:
   - `~/.vcli/prod/` - Production environment
   - `~/.vcli/dev/` - Development environment
   - `~/.vcli/customer-a/` - Customer-specific config

3. **Document environment setup** in project README:
   ```markdown
   # Setup
   export VCLI_SETTINGS_PATH="$HOME/.vcli/prod/"
   export VCLI_USERNAME="prod-admin"
   export VCLI_URL="vbr-prod.example.com"
   ```

### Automation

1. **Set credentials in CI/CD secrets** - Use platform-native secret stores
2. **Let auto-authentication work** - vcli detects CI/CD environments automatically
3. **Set `VCLI_SETTINGS_PATH` to build directory** to avoid conflicts
4. **Pass profile as argument** (not interactive):
```bash
#!/bin/bash
# CI/CD automation script
export VCLI_SETTINGS_PATH="./vcli-config/"

# Non-interactive init
./vcli init --output-dir ./vcli-config/

# Set profile with argument (not interactive)
./vcli profile --set vbr

# Auto-authenticates using VCLI_USERNAME/PASSWORD/URL from secrets
./vcli job diff --all --security-only
```

5. **Handle exit codes properly** for drift detection:
```bash
./vcli job diff --all --security-only
EXIT_CODE=$?
if [ $EXIT_CODE -eq 4 ]; then
    echo "CRITICAL drift detected"
    exit 1
fi
```

## See Also

- [Getting Started Guide](getting-started.md) - Complete setup walkthrough
- [Command Reference](command-reference.md) - Quick command reference
- [User Guide](../user_guide.md) - Full user documentation
