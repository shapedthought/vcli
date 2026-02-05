# Troubleshooting Guide

Common problems and solutions for vcli. This guide covers authentication, connection, API, state management, and operational issues.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Authentication Issues](#authentication-issues)
- [Connection Issues](#connection-issues)
- [TLS Certificate Issues](#tls-certificate-issues)
- [Profile Issues](#profile-issues)
- [API Issues](#api-issues)
- [State Management Issues](#state-management-issues)
- [Drift Detection Issues](#drift-detection-issues)
- [Apply Issues](#apply-issues)
- [Export Issues](#export-issues)
- [Overlay Issues](#overlay-issues)
- [Environment Variable Issues](#environment-variable-issues)
- [Platform-Specific Issues](#platform-specific-issues)

## Quick Diagnostics

Start troubleshooting with these quick checks:

```bash
# 1. Check vcli version
vcli utils
# Select "Check Version"

# 2. Verify environment variables
echo $VCLI_USERNAME
echo $VCLI_URL
echo $VCLI_SETTINGS_PATH

# 3. Check current profile
vcli profile --get

# 4. Test connectivity
ping vbr.example.com
nc -zv vbr.example.com 9419

# 5. Verify authentication
vcli login
```

## Authentication Issues

### 401 Unauthorized

**Problem:** Login fails with "Authentication failed" or 401 Unauthorized.

**Possible causes:**
1. Incorrect username or password
2. Wrong username format
3. Account locked or disabled
4. REST API not enabled (VBR)

**Solutions:**

**Try different username formats:**
```bash
# Standard
export VCLI_USERNAME="administrator"

# Domain user (format 1)
export VCLI_USERNAME="DOMAIN\\\\administrator"

# Domain user (format 2)
export VCLI_USERNAME="administrator@domain.com"

# Enterprise Manager format
export VCLI_USERNAME="administrator@"
```

**Check VBR REST API is enabled:**
1. Open VBR Console
2. Menu → Options → Network Traffic Rules
3. Ensure "RESTful API service" is checked
4. If unchecked, enable it and try again

**Verify credentials:**
```bash
# Test credentials manually
curl -k -u "username:password" https://vbr.example.com:9419/api/v1/jobs
```

**Check account status:**
- Verify account isn't locked in Active Directory
- Confirm account has appropriate VBR permissions
- Test login with Veeam Console to verify credentials

### Token Expired

**Problem:** API calls failing with authentication errors after working previously.

**Solution:** Re-login to get a fresh token:
```bash
vcli login
```

**Note:** Tokens expire after a period of inactivity (varies by product).

### Profile Not Set

**Problem:** "Profile not set" error when trying to login.

**Solution:**
```bash
# Set the profile first
vcli profile --set vbr

# Then login
vcli login
```

### Token Storage Issues (v0.11.0+)

**Problem:** `failed to open keyring` or `no authentication method available`

**Context:** System keychain unavailable (headless Linux, Docker, restricted environments)

**Solutions:**

**Option 1: Use explicit token (recommended for CI/CD)**
```bash
export VCLI_TOKEN="your-long-lived-token"
vcli get jobs
```

**Option 2: File-based keyring with password**
```bash
# CI/CD environments
export VCLI_FILE_KEY="your-secure-password"
vcli login

# Interactive systems
vcli login
# Prompts: Enter password for vcli file keyring: _
```

**Option 3: Auto-authenticate (requires credentials)**
```bash
# Set all three credentials
export VCLI_USERNAME="admin"
export VCLI_PASSWORD="pass"
export VCLI_URL="vbr.local"

# vcli auto-authenticates when keychain unavailable
vcli get jobs
```

### CI/CD Auto-Authentication Not Working

**Problem:** Commands fail in pipeline with "no authentication method available"

**Cause:** Missing required environment variables

**Solution:**

Check all required variables are set in CI/CD secrets:
```yaml
# GitHub Actions example
env:
  VCLI_USERNAME: ${{ secrets.VBR_USERNAME }}
  VCLI_PASSWORD: ${{ secrets.VBR_PASSWORD }}
  VCLI_URL: ${{ secrets.VBR_URL }}
```

Verify variables in pipeline:
```bash
echo "Username: $VCLI_USERNAME"
echo "URL: $VCLI_URL"
# Don't echo password for security

# Test authentication
vcli login
vcli get jobs
```

### Profile Command Requires Argument (v0.11.0+)

**Problem:** `vcli profile --set` hangs or returns "profile name required" error

**Breaking Change:** Profile commands now require explicit arguments (no interactive prompts)

**Old command (v0.10.x):**
```bash
vcli profile --set
# Prompted: Enter profile name: _
```

**New command (v0.11.0+):**
```bash
vcli profile --set vbr  # Provide argument
vcli profile -s vbr     # Short form
```

**In scripts:**
```bash
# Update CI/CD pipelines
- script: vcli profile --set vbr  # Add profile name
  displayName: 'Set Profile'
```

## Connection Issues

### Connection Refused

**Problem:** "Connection refused" or "dial tcp: connection refused"

**Possible causes:**
1. Wrong hostname/IP
2. Firewall blocking port
3. REST API service not running
4. Wrong profile selected (wrong port)

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

**Verify URL format:**
```bash
# Check VCLI_URL doesn't include protocol or port
echo $VCLI_URL
# Should be: vbr.example.com
# NOT: https://vbr.example.com:9419
```

**Check profile and port:**
```bash
# Verify correct profile is set
vcli profile --get

# Check profile port matches product
vcli profile --profile vbr
# VBR should be port 9419
# Enterprise Manager should be port 9398
```

**Check firewall rules:**
- Verify firewall allows connections to VBR port
- Test from another machine to isolate network issues
- Check Windows Firewall or iptables rules

### Timeout

**Problem:** Requests timeout without response.

**Possible causes:**
1. Network latency
2. VBR server overloaded
3. Large dataset causing slow response

**Solutions:**
```bash
# Test with a simple endpoint
vcli get jobs | jq '.data | length'

# Check VBR server performance in console
# Look for high CPU or memory usage

# Try connecting from VBR server itself (local)
# to eliminate network issues
```

### DNS Resolution Failed

**Problem:** "no such host" error

**Solutions:**
```bash
# Test DNS resolution
nslookup vbr.example.com

# Use IP address instead of hostname
export VCLI_URL="192.168.0.123"
vcli login

# Check /etc/hosts or C:\Windows\System32\drivers\etc\hosts
```

## TLS Certificate Issues

### Certificate Signed by Unknown Authority

**Problem:** `x509: certificate signed by unknown authority`

**Option 1: Skip TLS verification (NOT recommended for production)**

Edit `settings.json`:
```json
{
  "skipTLSVerify": true
}
```

**Option 2: Trust the CA certificate (recommended)**

**Windows:**
1. Export certificate from VBR Console
2. Double-click certificate file
3. Install to "Trusted Root Certification Authorities"

**Linux:**
```bash
# Get certificate
echo | openssl s_client -connect vbr.example.com:9419 | \
    openssl x509 > vbr.crt

# Install certificate (Ubuntu/Debian)
sudo cp vbr.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Install certificate (RHEL/CentOS)
sudo cp vbr.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust
```

**macOS:**
```bash
# Get certificate
echo | openssl s_client -connect vbr.example.com:9419 | \
    openssl x509 > vbr.crt

# Install to keychain
sudo security add-trusted-cert -d -r trustRoot \
    -k /Library/Keychains/System.keychain vbr.crt
```

### Certificate Hostname Mismatch

**Problem:** "certificate is valid for X, not Y"

**Causes:**
- Using IP instead of hostname in VCLI_URL
- Certificate issued for different hostname

**Solutions:**
```bash
# Use the hostname from the certificate
export VCLI_URL="vbr.company.local"  # Match certificate CN

# Or skip TLS verification (not recommended for production)
```

## Profile Issues

### Files Created in Wrong Directory

**Problem:** `settings.json` and `profiles.json` not in expected location.

**Solution:**

1. Set `VCLI_SETTINGS_PATH` before running `init`:
```bash
export VCLI_SETTINGS_PATH="$HOME/.vcli/"
mkdir -p ~/.vcli
vcli init
```

2. Or move files manually:
```bash
mkdir -p ~/.vcli
mv settings.json profiles.json ~/.vcli/
export VCLI_SETTINGS_PATH="$HOME/.vcli/"

# Note: headers.json no longer exists in v0.11.0+
# Tokens stored in system keychain instead
```

### Profile Not Found

**Problem:** vcli can't find the profile you specified.

**Solution:**
```bash
# List available profiles
vcli profile --list

# Use exact profile name
vcli profile --set vbr  # Not "VBR" or "vbr-prod"
```

### API Version Mismatch

**Problem:** API calls fail with version-related errors.

**Solution:**
Update API version in profiles.json:
```bash
# Edit profiles.json
vim $VCLI_SETTINGS_PATH/profiles.json

# Find the profile and update api_version or x-api-version
# VBR 12: "api_version": "1.1-rev2"
# VBR 13: "api_version": "1.3-rev1"
```

## API Issues

### Empty Response

**Problem:** Command succeeds but returns empty `{}`

**Causes:**
- Resource doesn't exist
- Wrong endpoint
- Permissions issue

**Solutions:**
```bash
# Verify endpoint works
vcli get jobs  # Should show list

# Check profile
vcli profile --get

# Verify authentication
vcli login

# Test with known good endpoint
vcli get backupInfrastructure/repositories
```

### JSON Parse Error

**Problem:** `invalid character '<' looking for beginning of value`

**Causes:**
- VBR returned HTML error page
- Wrong endpoint
- API service not running

**Solution:**
```bash
# Save response to see error
vcli get jobs > response.txt
cat response.txt  # Check if HTML error page

# Common causes:
# - Wrong API version in URL
# - Service not running
# - Incorrect endpoint
```

### No Such File or Directory (with -f flag)

**Problem:** `vcli post jobs -f data.json` fails

**Solutions:**
```bash
# Use absolute path
vcli post jobs -f /full/path/to/data.json

# Or relative from current directory
vcli post jobs -f ./data.json

# Verify file exists
ls -la data.json
```

### Rate Limiting

**Problem:** API calls fail with 429 Too Many Requests

**Solution:**
```bash
# Add delays between requests
for id in $(vcli get jobs | jq -r '.data[].id'); do
    vcli post jobs/$id/start
    sleep 2  # Wait 2 seconds between requests
done
```

## State Management Issues

### State File Not Found

**Problem:** vcli can't find state.json

**Solutions:**
```bash
# Check VCLI_SETTINGS_PATH
echo $VCLI_SETTINGS_PATH

# Create state with snapshots
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all

# Verify state file was created
ls -la $VCLI_SETTINGS_PATH/state.json
```

### Resource Not in State

**Problem:** Diff command reports "No state found for resource"

**Solution:**
```bash
# Snapshot the resource first
vcli repo snapshot "Resource Name"

# Or adopt existing resource
vcli repo adopt "Resource Name"

# Then diff will work
vcli repo diff "Resource Name"
```

### Corrupt State File

**Problem:** state.json is malformed or corrupt

**Solutions:**
```bash
# Validate JSON syntax
cat state.json | jq '.'

# Restore from backup
cp state.json.backup state.json

# Or rebuild from VBR
rm state.json
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all
```

### State Out of Sync

**Problem:** State doesn't match current VBR configuration

**Solution:**
```bash
# Re-snapshot to refresh state
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all
```

## Drift Detection Issues

### Drift Detected After Apply

**Problem:** Diff shows drift immediately after applying

**Possible causes:**
1. Some fields may not be updateable via API
2. VBR may add default values not in the spec
3. VBR may normalize certain values (e.g., schedule times)

**Solutions:**
```bash
# Re-snapshot after a short delay
sleep 5
vcli repo snapshot "Resource Name"

# Check drift severity (may be INFO and acceptable)
vcli repo diff "Resource Name"

# Compare full configurations
vcli repo export "Resource Name" -o current.yaml
diff applied.yaml current.yaml
```

### False Positive Drift

**Problem:** Diff reports drift but configurations look identical

**Causes:**
- Field order differences (not actual drift)
- Type differences (string "30" vs int 30)
- VBR adds computed fields not in spec

**Solutions:**
```bash
# Check severity - may be INFO (non-critical)
vcli job diff "Job Name"

# Use --security-only to filter out INFO drifts
vcli job diff --all --security-only

# Compare with exported current config
vcli export "Job Name" -o current.yaml
diff spec.yaml current.yaml
```

### No Drift Detection for Jobs

**Problem:** Can't manually snapshot jobs

**Explanation:**
Jobs are automatically snapshotted when you apply them. Manual job snapshots are not supported.

**Solution:**
```bash
# Apply the job to snapshot it
vcli job apply job.yaml

# Then diff will work
vcli job diff "Job Name"
```

## Apply Issues

### Resource Not Found (Exit Code 6)

**Problem:** Apply fails with "resource not found"

**Causes:**
Repositories, SOBRs, and KMS servers are update-only. They must be created in VBR console first.

**Solution:**
```bash
# 1. Create the resource in VBR console
# 2. Then apply can update it
vcli repo apply repo.yaml  # Now succeeds

# Jobs support creation via apply
vcli job apply job.yaml  # Creates if doesn't exist
```

### Partial Apply (Exit Code 5)

**Problem:** Apply succeeds but some fields were skipped

**Causes:**
Some fields may not be updateable via the VBR API.

**Solution:**
Review the output to see which fields were skipped. If they're important, you may need to update them manually in VBR console.

```bash
# Check which fields were skipped
vcli repo apply repo.yaml --dry-run

# Update non-updateable fields manually in VBR console
```

### Apply Fails with Validation Error

**Problem:** Apply fails with "invalid configuration" or validation error

**Solutions:**
```bash
# Validate YAML syntax
cat job.yaml | yq '.'

# Check for required fields
vcli job plan job.yaml --show-yaml

# Compare with exported working job
vcli export <working-job-id> -o working.yaml
diff job.yaml working.yaml

# Common issues:
# - Missing required fields
# - Invalid field values
# - Wrong resource type
```

### Apply Succeeds but Changes Not Visible

**Problem:** Apply reports success but VBR console doesn't show changes

**Solutions:**
```bash
# Refresh VBR console
# Press F5 or close/reopen

# Verify via API
vcli get jobs/<job-id>

# Check if change is actually supported
# Some fields may be display-only
```

## Export Issues

### Export Returns Empty Configuration

**Problem:** Export command succeeds but produces minimal/empty YAML

**Causes:**
- Resource doesn't exist
- Wrong resource ID or name
- Permissions issue

**Solutions:**
```bash
# List all resources first
vcli get jobs | jq '.data[] | {id, name}'

# Export by exact name
vcli export "Exact Job Name" -o job.yaml

# Export by ID
vcli export 57b3baab-6237-41bf-add7-db63d41d984c -o job.yaml

# Verify resource exists
vcli get jobs/<job-id>
```

### Export Fails for Specific Resource

**Problem:** Export works for some resources but not others

**Solutions:**
```bash
# Check resource type
vcli get jobs/<id> | jq '.type'

# Some job types may have export limitations
# Try simplified export for basic jobs
vcli export <id> --simplified -o job.yaml

# Check VBR logs for API errors
```

## Overlay Issues

### Overlay Not Being Applied

**Problem:** Overlay seems to be ignored

**Solutions:**

1. Check overlay resolution priority:
```bash
# Explicit -o flag has highest priority
vcli job apply base.yaml -o overlay.yaml

# --env flag looks up environment in vcli.yaml
vcli job apply base.yaml --env production

# currentEnvironment in vcli.yaml is used if no flags
vcli job apply base.yaml
```

2. Verify vcli.yaml exists and is in search path:
```bash
# Check VCLI_CONFIG
echo $VCLI_CONFIG

# Check current directory
ls -la vcli.yaml

# Check home directory
ls -la ~/.vcli/vcli.yaml
```

3. Use `--show-yaml` to see the actual merged result:
```bash
vcli job plan base.yaml -o overlay.yaml --show-yaml
```

### Unexpected Merge Results

**Problem:** Merged configuration doesn't look right

**Solutions:**

1. Remember: arrays are replaced, not merged
```yaml
# If your overlay has an objects array,
# it replaces the entire base array
```

2. Use `--show-yaml` to see full merged result:
```bash
vcli job plan base.yaml -o overlay.yaml --show-yaml
```

3. Check that overlay `kind` matches base `kind`:
```yaml
# Base and overlay must both be: kind: VBRJob
```

### Labels Not Combining

**Problem:** Labels from base not appearing in merged config

**Solutions:**

1. Ensure both base and overlay use `metadata.labels` field
2. Labels should be at same level in both files:
```yaml
# Both files should have:
metadata:
  labels:
    key: value
```

## Environment Variable Issues

### Environment Variables Not Set

**Problem:** vcli can't find credentials

**Solutions:**

**Bash/Zsh:**
```bash
# Set variables
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="password"
export VCLI_URL="vbr.example.com"

# Verify they're set
echo $VCLI_USERNAME
echo $VCLI_URL

# Make permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export VCLI_USERNAME="administrator"' >> ~/.bashrc
```

**PowerShell:**
```powershell
# Set variables
$env:VCLI_USERNAME = "administrator"
$env:VCLI_PASSWORD = "password"
$env:VCLI_URL = "vbr.example.com"

# Verify they're set
$env:VCLI_USERNAME
$env:VCLI_URL

# Make permanent
[Environment]::SetEnvironmentVariable("VCLI_USERNAME", "administrator", "User")
```

### Variables Set But Not Working

**Problem:** Variables are set but vcli doesn't see them

**Solutions:**
```bash
# Check for typos
env | grep VCLI

# Ensure no spaces around =
export VCLI_USERNAME="admin"  # Correct
export VCLI_USERNAME = "admin"  # Wrong

# Restart shell after setting variables
```

### Settings Path Not Found

**Problem:** vcli can't find settings files

**Solutions:**
```bash
# Check VCLI_SETTINGS_PATH
echo $VCLI_SETTINGS_PATH

# Ensure directory exists
mkdir -p ~/.vcli

# Set before running init
export VCLI_SETTINGS_PATH="$HOME/.vcli/"
vcli init

# Verify files are in correct location
ls -la $VCLI_SETTINGS_PATH/
```

## Platform-Specific Issues

### Windows PowerShell Escaping

**Problem:** Commands with special characters fail

**Solutions:**
```powershell
# Use double quotes for paths with spaces
vcli export <id> -o "C:\Users\Name\My Documents\job.yaml"

# Escape backslashes in paths
vcli export <id> -o "C:\\Users\\Name\\job.yaml"

# Use forward slashes (PowerShell accepts them)
vcli export <id> -o "C:/Users/Name/job.yaml"
```

### Linux Permissions

**Problem:** Permission denied errors

**Solutions:**
```bash
# Make vcli executable
chmod +x vcli

# Run with sudo if needed (not recommended)
sudo ./vcli login

# Better: Fix ownership
sudo chown $USER:$USER vcli
chmod +x vcli
```

### macOS Gatekeeper

**Problem:** "vcli cannot be opened because the developer cannot be verified"

**Solutions:**
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine vcli

# Or allow in System Preferences
# System Preferences → Security & Privacy → Allow anyway
```

### Path Issues (All Platforms)

**Problem:** `vcli: command not found`

**Solutions:**

**Use explicit path:**
```bash
./vcli login
/path/to/vcli login
```

**Add to PATH:**
```bash
# Linux/macOS
export PATH="$PATH:/path/to/vcli"
echo 'export PATH="$PATH:/path/to/vcli"' >> ~/.bashrc

# Windows PowerShell
$env:PATH += ";C:\path\to\vcli"
```

## Getting Help

If you can't find a solution here:

1. **Check existing issues:** https://github.com/shapedthought/vcli/issues
2. **Create a new issue:** https://github.com/shapedthought/vcli/issues/new
3. **Include diagnostics:**
   - vcli version (`vcli utils` → Check Version)
   - Operating system
   - Veeam product and version
   - Error messages (full text)
   - Steps to reproduce

**Useful diagnostic commands:**
```bash
# Version info
vcli utils  # Select "Check Version"

# Environment
env | grep VCLI

# Profile info
vcli profile --get
vcli profile --profile vbr

# Connectivity test
ping vbr.example.com
nc -zv vbr.example.com 9419

# API test
curl -k https://vbr.example.com:9419/api/v1/jobs
```

## See Also

- [Getting Started Guide](getting-started.md) - Initial setup
- [Authentication Guide](authentication.md) - Authentication troubleshooting
- [Imperative Mode Guide](imperative-mode.md) - API command troubleshooting
- [Declarative Mode Guide](declarative-mode.md) - Apply/export troubleshooting
- [State Management Guide](state-management.md) - State troubleshooting
- [Drift Detection Guide](drift-detection.md) - Drift detection issues
