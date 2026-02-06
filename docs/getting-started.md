# Getting Started with vcli

This guide walks you through installing vcli, setting up authentication, and choosing the right workflow for your needs.

## Prerequisites

- Access to a Veeam API endpoint (VBR, Enterprise Manager, VB365, VONE, or VB for cloud)
- API credentials (username and password)
- Network access to the Veeam server

## Installation

### Download

1. Go to the [Releases page](https://github.com/shapedthought/vcli/releases)
2. Download the appropriate binary for your platform:
   - `vcli-windows-amd64.exe` - Windows
   - `vcli-linux-amd64` - Linux
   - `vcli-darwin-amd64` - macOS (Intel)
   - `vcli-darwin-arm64` - macOS (Apple Silicon)

### Verify Checksum

**Windows (PowerShell):**
```powershell
Get-FileHash -Path vcli-windows-amd64.exe -Algorithm SHA256
```

**macOS:**
```bash
shasum -a 256 vcli-darwin-amd64
```

**Linux:**
```bash
sha256sum vcli-linux-amd64
```

Compare the output with the checksum in the release notes.

### Make Executable (macOS/Linux)

```bash
chmod +x vcli-darwin-amd64
# Optional: rename for convenience
mv vcli-darwin-amd64 vcli
```

### Add to PATH (Optional)

**macOS/Linux:**
```bash
# Move to a directory in your PATH
sudo mv vcli /usr/local/bin/

# Or add current directory to PATH
export PATH=$PATH:$(pwd)
```

**Windows:**
```powershell
# Add current directory to PATH (temporary)
$env:Path += ";$PWD"

# Or move to an existing PATH directory
Move-Item vcli-windows-amd64.exe C:\Windows\System32\vcli.exe
```

## First-Time Setup

### 1. Initialize vcli

Create the configuration files non-interactively:

```bash
# Basic init (creates files with defaults)
./vcli init

# With specific directory
./vcli init --output-dir ~/.vcli/

# With flags for specific settings
./vcli init --insecure
```

Init is non-interactive by default and outputs JSON to stdout. For legacy interactive mode, use `./vcli init --interactive`.

This creates:
- `settings.json` - vcli settings
- `profiles.json` - API profiles for each Veeam product (v1.0 format with all profiles)

**Configuration Location:**
- By default, files are created in the current directory
- Use `--output-dir` flag or set `VCLI_SETTINGS_PATH` environment variable to use a specific directory

**Available Flags:**
- `--insecure` - Skip TLS verification (sets `apiNotSecure: true`)
- `--output-dir <path>` - Specify where to write config files
- `--interactive` - Use legacy interactive prompts

**Subcommands:**
```bash
# Initialize only settings
./vcli init settings --insecure

# Initialize only profiles
./vcli init profiles
```

### 2. Set Credentials

vcli reads credentials from environment variables:

**Bash/Zsh (macOS/Linux):**
```bash
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"
```

**PowerShell (Windows):**
```powershell
$env:VCLI_USERNAME = "administrator"
$env:VCLI_PASSWORD = "your-password"
$env:VCLI_URL = "vbr.example.com"
```

**Notes:**
- `VCLI_URL` should be the hostname or IP without `https://` or port
- Domain users: Use format `DOMAIN\username` or `username@domain.com`
- Port is handled automatically based on the selected profile

### 3. Select Profile

Set the Veeam product you're connecting to:

```bash
# List available profiles
./vcli profile --list

# Set profile directly (no prompts)
./vcli profile --set vbr

# Get current profile (returns just "vbr")
./vcli profile --get
```

Profile commands require explicit arguments and return clean output for scripting.

**Available Profiles:**
- `vbr` - Veeam Backup & Replication (port 9419)
- `ent_man` - Enterprise Manager (port 9398)
- `vb365` - Veeam Backup for Microsoft 365 (port 4443)
- `vone` - Veeam ONE (port 1239)
- `aws` - Veeam Backup for AWS (port 11005)
- `azure` - Veeam Backup for Azure (port 443)
- `gcp` - Veeam Backup for GCP (port 13140)

### 4. Login

Authenticate with the Veeam API:

```bash
./vcli login
```

Tokens are stored securely:
- **Interactive sessions:** System keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **CI/CD pipelines:** Auto-authentication using environment variables (no keychain interaction)

**Troubleshooting:**
- **TLS certificate errors**: If using self-signed certificates, see [Troubleshooting](#troubleshooting) below
- **Authentication failed**: Verify username format and credentials
- **Connection refused**: Check firewall rules and that the REST API is enabled

## Choose Your Workflow

vcli supports two distinct modes of operation. Choose based on your use case.

### Imperative Mode

**Best for:**
- Quick API operations
- One-off tasks
- Exploring the API
- Products without declarative support (VB365, VONE, cloud products)

**Example: Start a backup job**
```bash
# List all jobs
./vcli get jobs

# Get specific job details
./vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c

# Start a job
./vcli post jobs/57b3baab-6237-41bf-add7-db63d41d984c/start

# Get job status
./vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c
```

**Key Commands:**
- `vcli get <endpoint>` - Retrieve data
- `vcli post <endpoint>` - Trigger operations (with optional `-f data.json`)
- `vcli put <endpoint> -f data.json` - Update resources

See the [User Guide](../user_guide.md) for complete imperative mode documentation.

### Declarative Mode (VBR Only)

**Best for:**
- Infrastructure-as-code
- Version control and GitOps
- Multi-environment deployments (dev/staging/prod)
- Drift detection and security monitoring
- Automated remediation in CI/CD

**Example: Manage backup jobs with Git**

#### 1. Export Existing Configuration

```bash
# Export a single job
./vcli export c07c7ea3-0471-43a6-af57-c03c0d82354a -o prod-backup.yaml

# Export all jobs
./vcli export --all -d jobs/

# Export repositories
./vcli repo export --all -d repos/

# Export SOBRs
./vcli repo sobr-export --all -d sobrs/

# Export KMS servers
./vcli encryption kms-export --all -d kms/
```

#### 2. Edit Configuration

Open `prod-backup.yaml` in your editor and make changes:

```yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: Production Database Backup
spec:
  description: Updated description
  storage:
    retentionPolicy:
      quantity: 30  # Changed from 7 to 30 days
  schedule:
    daily:
      localTime: "02:00"  # Changed from 22:00
```

#### 3. Preview Changes

```bash
# Dry-run to see what would change
./vcli job apply prod-backup.yaml --dry-run
```

#### 4. Apply Configuration

```bash
# Apply changes to VBR
./vcli job apply prod-backup.yaml
```

#### 5. Commit to Git

```bash
git add prod-backup.yaml
git commit -m "Update retention to 30 days for prod backup"
git push
```

#### 6. Detect Drift

Someone makes a manual change in VBR? vcli will detect it:

```bash
# Check for drift
./vcli job diff "Production Database Backup"

# Check all jobs for security-relevant drift
./vcli job diff --all --security-only
```

**Output:**
```
Checking drift for job: Production Database Backup

CRITICAL: 1 security-relevant changes detected

Drift detected:
  CRITICAL ~ storage.retentionPolicy.quantity: 30 (state) -> 7 (VBR)

Summary:
  - 1 drifts detected
  - Highest severity: CRITICAL
```

**Key Commands:**
- `vcli export <id>` - Export resource to YAML
- `vcli job apply <file>` - Apply job configuration
- `vcli job diff <name>` - Detect drift
- `vcli repo apply <file>` - Apply repository configuration
- `vcli repo sobr-apply <file>` - Apply SOBR configuration
- `vcli encryption kms-apply <file>` - Apply KMS server configuration

See [Drift Detection Guide](drift-detection.md) for complete documentation.

## Multi-Environment Workflow (Declarative)

Manage dev/staging/prod environments with configuration overlays.

### 1. Create Base Configuration

**base-backup.yaml:**
```yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
spec:
  type: VSphereBackup
  repository: default-repo
  storage:
    compression: Optimal
    retentionPolicy:
      type: Days
      quantity: 7
  schedule:
    daily:
      localTime: "22:00"
  objects:
    - type: VirtualMachine
      name: db-server
```

### 2. Create Environment Overlays

**overlays/prod.yaml:**
```yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
spec:
  repository: prod-repo
  storage:
    retentionPolicy:
      quantity: 30
  schedule:
    daily:
      localTime: "02:00"
```

**overlays/dev.yaml:**
```yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
spec:
  repository: dev-repo
  storage:
    retentionPolicy:
      quantity: 3
  schedule:
    daily:
      localTime: "23:00"
```

### 3. Apply with Overlays

```bash
# Apply production configuration
./vcli job apply base-backup.yaml -o overlays/prod.yaml

# Apply development configuration
./vcli job apply base-backup.yaml -o overlays/dev.yaml

# Preview merged result
./vcli job plan base-backup.yaml -o overlays/prod.yaml --show-yaml
```

### 4. Track in Git

```bash
git add base-backup.yaml overlays/
git commit -m "Add multi-environment backup configuration"
git push
```

## Next Steps

### Imperative Mode Users

- Read the [User Guide](../user_guide.md) for detailed command reference
- Learn about [API profiles](../user_guide.md#profiles) for multi-product management
- Explore [output formatting](../user_guide.md#using-with-nushell) with jq and Nushell

### Declarative Mode Users

- **Start here:** [GitOps Workflows Guide](gitops-workflows.md) - Comprehensive CI/CD integration
- Deep dive into [Drift Detection](drift-detection.md)
- Understand [Security Alerting](security-alerting.md) severity classification
- Set up [Azure DevOps Integration](azure-devops-integration.md) for CI/CD
- Review [pipeline templates](../examples/pipelines/) for automation

### Both

- Join the community on GitHub Discussions
- Report issues or request features on [GitHub Issues](https://github.com/shapedthought/vcli/issues)
- Check the [Changelog](../README.md#change-log) for latest updates

## Troubleshooting

### TLS Certificate Errors

**Problem:** `x509: certificate signed by unknown authority`

**Solution:** Trust the certificate or use a CA bundle.

**Option 1: Skip verification (NOT recommended for production)**

Set in `settings.json`:
```json
{
  "skipTLSVerify": true
}
```

**Option 2: Trust the CA certificate (recommended)**

```bash
# Export VBR certificate
openssl s_client -connect vbr.example.com:9419 -showcerts > vbr-cert.pem

# macOS: Add to keychain
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain vbr-cert.pem

# Linux: Add to CA bundle
sudo cp vbr-cert.pem /usr/local/share/ca-certificates/vbr.crt
sudo update-ca-certificates
```

### Authentication Failures

**Problem:** `401 Unauthorized` or `failed to login`

**Possible causes:**
1. Incorrect credentials
2. Wrong username format (try `DOMAIN\user` or `user@domain.com`)
3. Account locked or disabled
4. REST API not enabled on VBR

**Check REST API is enabled (VBR):**
1. Open VBR console
2. Menu → Options → Network Traffic Rules
3. Ensure "RESTful API service" is enabled

### Connection Refused

**Problem:** `connection refused` or `no route to host`

**Check:**
1. Firewall rules allow access to the API port
2. VBR/product is running
3. REST API service is started
4. Network connectivity: `ping vbr.example.com`

### Profile Not Set

**Problem:** `profile not set`

**Solution:**
```bash
./vcli profile --set vbr
```

### State File Corruption

**Problem:** `failed to parse state.json`

**Solution:**
```bash
# Backup current state
cp ~/.vcli/state.json ~/.vcli/state.json.backup

# Reset state (WARNING: loses drift detection history)
rm ~/.vcli/state.json

# Re-snapshot resources
./vcli repo snapshot --all
./vcli repo sobr-snapshot --all
./vcli encryption snapshot --all
./vcli encryption kms-snapshot --all
```

### Resource Not Found (Exit Code 6)

**Problem:** `vcli repo apply` returns exit code 6

**Explanation:** Repositories, SOBRs, and KMS servers are **update-only**. They must be created in the VBR console first, then managed via vcli.

**Solution:**
1. Create the resource in VBR console
2. Export it: `./vcli repo export "Repository Name" -o repo.yaml`
3. Now you can apply changes: `./vcli repo apply repo.yaml`

## Environment Variables Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `VCLI_USERNAME` | Yes | API username |
| `VCLI_PASSWORD` | Yes | API password |
| `VCLI_URL` | Yes | Veeam server hostname/IP (without https:// or port) |
| `VCLI_SETTINGS_PATH` | No | Directory for config files (default: current directory) |
| `VCLI_CONFIG` | No | Path to vcli.yaml (planned feature) |

## Exit Codes

### Apply Commands

| Code | Meaning |
|------|---------|
| `0` | Success - applied successfully |
| `1` | Error - API failure or invalid spec |
| `5` | Partial apply - some fields skipped (known immutable) |
| `6` | Resource not found - cannot apply (update-only resources) |

### Diff Commands

| Code | Meaning |
|------|---------|
| `0` | No drift detected |
| `3` | Drift detected (INFO or WARNING severity) |
| `4` | Critical drift detected (CRITICAL severity) |
| `1` | Error occurred |

Use exit codes in scripts:
```bash
./vcli job diff --all --security-only
if [ $? -eq 4 ]; then
    echo "CRITICAL security drift detected!"
    exit 1
fi
```
