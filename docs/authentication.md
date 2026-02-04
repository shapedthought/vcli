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

**New in v0.11.0:** Init is now non-interactive by default for GitOps/CI/CD compatibility.

Run the init command to create configuration files:

```bash
# Non-interactive init (default)
./vcli init

# With specific settings
./vcli init --insecure --creds-file --output-dir ~/.vcli/

# Legacy interactive mode (deprecated, will be removed in v0.12.0)
./vcli init --interactive
```

**Available Flags:**
- `--insecure` - Skip TLS verification
- `--creds-file` - Enable credentials file mode
- `--output-dir <path>` - Specify config file directory
- `--interactive` - Use legacy interactive prompts (deprecated)

**Files created:**
- `settings.json` - vcli settings and preferences
- `profiles.json` - API profiles for each Veeam product

**Output:** Init now outputs JSON to stdout for automation/piping:
```json
{
  "settings": {...},
  "profiles": [...],
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

**New in v0.11.0:** Profile commands now require explicit arguments (no interactive prompts).

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

### Profile Structure

Example profile (Enterprise Manager):

```json
{
  "name": "ent_man",
  "headers": {
    "accept": "application/json",
    "Content-type": "application/json",
    "x-api-version": ""
  },
  "url": ":9398/api/sessionMngr/?v=latest",
  "port": "9398",
  "api_version": "",
  "username": "administrator@",      // Only in creds file mode
  "address": "192.168.0.123"        // Only in creds file mode
}
```

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

## Authentication Modes

vcli supports two authentication modes. Choose based on your workflow.

### Environmental Mode (Recommended)

**Credentials from environment variables only.**

**Pros:**
- More secure (no credentials in files)
- Simple setup
- Works well with CI/CD
- Recommended for most users

**Cons:**
- Must set environment variables each session
- Slower when switching between products

**Setup:**
1. During `init`, omit `--creds-file` flag (defaults to environmental mode)
2. Set environment variables before each session

```bash
# Initialize in environmental mode (default)
./vcli init --output-dir ~/.vcli/
```

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
| `VCLI_USERNAME` | Yes | API username (may need `DOMAIN\user` format) |
| `VCLI_PASSWORD` | Yes | API password |
| `VCLI_URL` | Yes | Server hostname/IP (without `https://` or port) |
| `VCLI_SETTINGS_PATH` | No | Config file directory (default: current directory) |

### Creds File Mode

**Username and address stored in profiles.json, password from environment variable.**

**Pros:**
- Faster switching between products
- Username and address stored per profile

**Cons:**
- Username and address stored in plaintext in `profiles.json`
- Less secure
- Not recommended for shared systems

**Setup:**
1. During `init`, use `--creds-file` flag
2. Edit `profiles.json` and add `username` and `address` fields to each profile:

```bash
# Initialize in creds file mode
./vcli init --creds-file --output-dir ~/.vcli/
```

```json
{
  "name": "vbr",
  "username": "administrator@domain.com",
  "address": "vbr.example.com",
  // ... other fields
}
```

3. Set password environment variable:
```bash
export VCLI_PASSWORD="your-password"
```

4. Switch profiles:
```bash
./vcli profile --set vbr
./vcli login

./vcli profile --set vb365
./vcli login  # Uses vb365 credentials from profile
```

### Switching Modes

**Option 1: Re-initialize with different flag**

```bash
# Switch to environmental mode
./vcli init --output-dir ~/.vcli/

# Switch to creds file mode
./vcli init --creds-file --output-dir ~/.vcli/
```

**Option 2: Edit `settings.json` manually**

Edit `settings.json` and change `credsFileMode`:

```json
{
  "credsFileMode": false,    // Environmental mode
  "skipTLSVerify": false
}
```

```json
{
  "credsFileMode": true,     // Creds file mode
  "skipTLSVerify": false
}
```

## Logging In

After setting up credentials and selecting a profile, authenticate:

```bash
./vcli login
```

**On success:**
- Creates `headers.json` with authentication token
- Token is used for all subsequent API calls
- Token is valid until expiration or logout

**Important:** The authentication token is overwritten on each login. Switching between profiles requires re-login with each profile's credentials.

### Authentication Flow

**Environmental Mode:**
1. Set `VCLI_USERNAME`, `VCLI_PASSWORD`, `VCLI_URL`
2. Set profile: `./vcli profile --set vbr`
3. Login: `./vcli login`
4. API calls use token from `headers.json`

**Creds File Mode:**
1. Edit `profiles.json` with username and address
2. Set `VCLI_PASSWORD` environment variable
3. Set profile: `./vcli profile --set vbr`
4. Login: `./vcli login` (uses credentials from profile + env password)
5. API calls use token from `headers.json`

### Authentication Types by Product

**OAuth (most products):**
- VBR, VB365, VONE, AWS, Azure, GCP
- Uses Bearer token
- Token stored in `headers.json`

**Basic Auth (Enterprise Manager only):**
- Enterprise Manager uses session-based authentication
- Session token in `X-RestSvcSessionId` header
- Different URL pattern than other products

## Switching Between Products

### Method 1: Environmental Mode

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

### Method 2: Creds File Mode

```bash
# profiles.json contains credentials for all products
# Just switch profile and login

./vcli profile --set vbr
./vcli login
./vcli get jobs

./vcli profile --set vb365
./vcli login
./vcli get organizations
```

### Method 3: Multiple Config Directories

Use different directories for different environments:

```bash
# Production VBR
export VCLI_SETTINGS_PATH="$HOME/.vcli/prod/"
./vcli profile --set vbr
./vcli login

# Development VBR
export VCLI_SETTINGS_PATH="$HOME/.vcli/dev/"
./vcli profile --set vbr
./vcli login
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
mv settings.json profiles.json headers.json ~/.vcli/
export VCLI_SETTINGS_PATH="$HOME/.vcli/"
```

### Token Expired

**Problem:** API calls failing with authentication errors after working previously.

**Solution:** Re-login to get a fresh token:
```bash
./vcli login
```

### Multiple Profiles - Wrong Credentials Used

**Problem:** Creds file mode uses wrong username for current profile.

**Solution:** Verify each profile has correct credentials:

```bash
./vcli profile --profile vbr
# Check "username" and "address" fields

# Update profiles.json manually if needed
```

## Best Practices

### Security

1. **Use environmental mode** for better security
2. **Never commit `headers.json`** to version control
3. **Use secrets management** in CI/CD (Azure Key Vault, GitHub Secrets, etc.)
4. **Rotate credentials regularly**
5. **Use dedicated service accounts** for automation

### Organization

1. **Use custom settings path** to keep config centralized: `~/.vcli/`
2. **Name config directories** clearly for multiple environments:
   - `~/.vcli/prod/`
   - `~/.vcli/dev/`
   - `~/.vcli/customer-a/`
3. **Document which profile is for which product** in a README

### Automation

1. **Set credentials in CI/CD secrets**
2. **Use environmental mode** in automation
3. **Set `VCLI_SETTINGS_PATH` to build directory** to avoid conflicts
4. **Include profile setting in scripts**:
```bash
#!/bin/bash
export VCLI_SETTINGS_PATH="./vcli-config/"
./vcli init
./vcli profile --set vbr
./vcli login
./vcli get jobs
```

## See Also

- [Getting Started Guide](getting-started.md) - Complete setup walkthrough
- [Command Reference](command-reference.md) - Quick command reference
- [User Guide](../user_guide.md) - Full user documentation
