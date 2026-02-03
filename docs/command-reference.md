# Command Quick Reference

Fast reference for common vcli commands. See full documentation in [User Guide](../user_guide.md) and [Drift Detection Guide](drift-detection.md).

## Table of Contents

- [Setup & Authentication](#setup--authentication)
- [Imperative Commands (All Products)](#imperative-commands-all-products)
- [Declarative Commands (VBR Only)](#declarative-commands-vbr-only)
- [Common Flags](#common-flags)
- [Exit Codes](#exit-codes)

---

## Setup & Authentication

```bash
# Initialize (first time only)
vcli init

# Set profile
vcli profile --set vbr                    # VBR
vcli profile --set ent_man                # Enterprise Manager
vcli profile --set vb365                  # VB365
vcli profile --list                       # Show all profiles

# Set credentials (environment variables)
export VCLI_USERNAME="administrator"
export VCLI_PASSWORD="your-password"
export VCLI_URL="vbr.example.com"

# Login
vcli login
```

---

## Imperative Commands (All Products)

### GET - Retrieve Data

```bash
# List all resources
vcli get jobs
vcli get repositories
vcli get backupInfrastructure/managedServers

# Get specific resource
vcli get jobs/<job-id>
vcli get repositories/<repo-id>

# Output formats
vcli get jobs                             # JSON (default)
vcli get jobs --yaml                      # YAML
vcli get jobs | jq '.data'                # Parse with jq
```

### POST - Trigger Operations

```bash
# Without payload
vcli post jobs/<job-id>/start
vcli post jobs/<job-id>/stop
vcli post jobs/<job-id>/retry

# With payload
vcli post jobs -f job-data.json
vcli post repositories -f repo-data.json
```

### PUT - Update Resources

```bash
# Update resource (requires payload)
vcli put jobs/<job-id> -f updated-job.json
vcli put repositories/<repo-id> -f updated-repo.json
```

### Common Endpoints (VBR)

```bash
# Jobs
vcli get jobs                             # All jobs
vcli get jobs/<id>                        # Job details
vcli post jobs/<id>/start                 # Start job
vcli post jobs/<id>/stop                  # Stop job

# Repositories
vcli get backupInfrastructure/repositories
vcli get backupInfrastructure/scaleOutRepositories

# Infrastructure
vcli get backupInfrastructure/managedServers
vcli get backupInfrastructure/credentials

# Sessions
vcli get sessions                         # Recent sessions
vcli get sessions/<id>                    # Session details
```

---

## Declarative Commands (VBR Only)

### Export Resources

```bash
# Jobs
vcli export <job-id> -o job.yaml          # Export single job
vcli export --all -d jobs/                # Export all jobs
vcli export <id> --as-overlay -o ov.yaml  # Export as overlay
vcli export <id> --simplified -o job.yaml # Minimal format (legacy)

# Repositories
vcli repo export <name> -o repo.yaml
vcli repo export --all -d repos/

# Scale-Out Repositories (SOBRs)
vcli repo sobr-export <name> -o sobr.yaml
vcli repo sobr-export --all -d sobrs/

# Encryption Passwords
vcli encryption export <name> -o enc.yaml
vcli encryption export --all -d encryption/

# KMS Servers
vcli encryption kms-export <name> -o kms.yaml
vcli encryption kms-export --all -d kms/
```

### Apply Configurations

```bash
# Jobs (create or update)
vcli job apply job.yaml
vcli job apply job.yaml --dry-run         # Preview changes
vcli job apply base.yaml -o prod.yaml     # Apply with overlay

# Repositories (update-only)
vcli repo apply repo.yaml
vcli repo apply repo.yaml --dry-run

# SOBRs (update-only)
vcli repo sobr-apply sobr.yaml
vcli repo sobr-apply sobr.yaml --dry-run

# KMS Servers (update-only)
vcli encryption kms-apply kms.yaml
vcli encryption kms-apply kms.yaml --dry-run
```

**Note:** Repos, SOBRs, and KMS are update-only. Create them in VBR console first.

### Snapshot State

```bash
# Repositories
vcli repo snapshot <name>
vcli repo snapshot --all

# SOBRs
vcli repo sobr-snapshot <name>
vcli repo sobr-snapshot --all

# Encryption Passwords
vcli encryption snapshot <name>
vcli encryption snapshot --all

# KMS Servers
vcli encryption kms-snapshot <name>
vcli encryption kms-snapshot --all
```

**Note:** Jobs are snapshotted automatically on apply.

### Detect Drift

```bash
# Single resource
vcli job diff "Job Name"
vcli repo diff "Repository Name"
vcli repo sobr-diff "SOBR Name"
vcli encryption diff "Password Name"
vcli encryption kms-diff "KMS Server Name"

# All resources
vcli job diff --all
vcli repo diff --all
vcli repo sobr-diff --all
vcli encryption diff --all
vcli encryption kms-diff --all

# Severity filtering
vcli job diff --all --severity critical   # Only CRITICAL
vcli job diff --all --security-only       # WARNING and above
vcli repo diff --all --severity warning   # WARNING and above
```

### Plan (Preview)

```bash
# Preview merged configuration
vcli job plan base.yaml
vcli job plan base.yaml -o prod.yaml
vcli job plan base.yaml -o prod.yaml --show-yaml
```

---

## Common Flags

### Apply Commands

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview changes without applying |
| `-o, --overlay <file>` | Apply with configuration overlay |
| `--env <name>` | Use environment from vcli.yaml (planned) |

### Diff Commands

| Flag | Description |
|------|-------------|
| `--all` | Check all resources |
| `--severity <level>` | Filter by severity: `critical`, `warning`, or `info` |
| `--security-only` | Show only WARNING and CRITICAL drifts |

### Export Commands

| Flag | Description |
|------|-------------|
| `-o, --output <file>` | Output file path |
| `-d, --directory <dir>` | Output directory (for --all) |
| `--all` | Export all resources |
| `--as-overlay` | Export as minimal overlay (jobs only) |
| `--base <file>` | Base file for overlay comparison (with --as-overlay) |
| `--simplified` | Minimal format - legacy (jobs only) |

### Global Flags

| Flag | Description |
|------|-------------|
| `--yaml` | Output in YAML format (default: JSON) |
| `-h, --help` | Show help |

---

## Exit Codes

### Apply Commands

| Code | Meaning | Action |
|------|---------|--------|
| `0` | Success | Continue |
| `1` | Error (API failure, invalid spec) | Fix and retry |
| `5` | Partial apply (some fields skipped) | Review skipped fields |
| `6` | Resource not found (update-only) | Create in VBR console first |

**Example:**
```bash
vcli repo apply repo.yaml
if [ $? -eq 6 ]; then
    echo "Repository doesn't exist - create it in VBR console first"
    exit 1
fi
```

### Diff Commands

| Code | Meaning | Action |
|------|---------|--------|
| `0` | No drift | Continue |
| `3` | Drift detected (INFO/WARNING) | Review and remediate if needed |
| `4` | Critical drift detected | Immediate remediation required |
| `1` | Error occurred | Check logs |

**Example:**
```bash
vcli job diff --all --security-only
EXIT_CODE=$?

if [ $EXIT_CODE -eq 4 ]; then
    echo "CRITICAL security drift detected!"
    # Send alert
    exit 1
elif [ $EXIT_CODE -eq 3 ]; then
    echo "WARNING: Drift detected, review required"
fi
```

---

## Common Workflows

### Quick API Check

```bash
# 1. Login
export VCLI_USERNAME="admin" VCLI_PASSWORD="pass" VCLI_URL="vbr.local"
vcli profile --set vbr
vcli login

# 2. Query
vcli get jobs | jq '.data[] | {name: .name, type: .type}'
```

### Export All Configurations

```bash
# Export everything to Git
vcli export --all -d specs/jobs/
vcli repo export --all -d specs/repos/
vcli repo sobr-export --all -d specs/sobrs/
vcli encryption kms-export --all -d specs/kms/

# Commit to Git
git add specs/
git commit -m "Snapshot VBR configuration"
git push
```

### Multi-Environment Deployment

```bash
# Preview production changes
vcli job plan base-backup.yaml -o overlays/prod.yaml --show-yaml

# Apply production (dry-run first)
vcli job apply base-backup.yaml -o overlays/prod.yaml --dry-run
vcli job apply base-backup.yaml -o overlays/prod.yaml

# Apply development
vcli job apply base-backup.yaml -o overlays/dev.yaml
```

### Drift Detection Scan

```bash
#!/bin/bash
# Full security drift scan

CRITICAL=0

for cmd in \
    "job diff --all --severity critical" \
    "repo diff --all --severity critical" \
    "repo sobr-diff --all --severity critical" \
    "encryption diff --all --severity critical" \
    "encryption kms-diff --all --severity critical"
do
    vcli $cmd
    [ $? -eq 4 ] && CRITICAL=1
done

if [ $CRITICAL -eq 1 ]; then
    echo "CRITICAL drift detected across VBR environment"
    exit 1
fi

echo "No critical drift detected"
```

### Bootstrap Declarative Management

```bash
# 1. Export current VBR state
vcli export --all -d infrastructure/jobs/
vcli repo export --all -d infrastructure/repos/
vcli repo sobr-export --all -d infrastructure/sobrs/
vcli encryption kms-export --all -d infrastructure/kms/

# 2. Snapshot current state
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all

# 3. Commit to Git (state.json and specs)
git add infrastructure/ state.json
git commit -m "Bootstrap VBR declarative management"
git push

# 4. Verify no drift
vcli job diff --all
vcli repo diff --all
```

---

## Tips

### Output Parsing

```bash
# JSON (default)
vcli get jobs | jq '.data[0].name'

# YAML
vcli get jobs --yaml | yq '.data[0].name'

# With Nushell
vcli get jobs | from json | get data | where isDisabled == false
```

### Credential Management

```bash
# Temporary credentials (current shell only)
export VCLI_USERNAME="admin" VCLI_PASSWORD="pass" VCLI_URL="vbr.local"

# Persistent credentials (bash/zsh)
echo 'export VCLI_USERNAME="admin"' >> ~/.bashrc
echo 'export VCLI_PASSWORD="pass"' >> ~/.bashrc
echo 'export VCLI_URL="vbr.local"' >> ~/.bashrc

# Use .env file (with direnv or similar)
# .env
VCLI_USERNAME=admin
VCLI_PASSWORD=pass
VCLI_URL=vbr.local
```

### Configuration Directory

```bash
# Set custom config location
export VCLI_SETTINGS_PATH="$HOME/.vcli/"

# Initialize in custom location
mkdir -p ~/.vcli
cd ~/.vcli
vcli init
```

---

## See Also

- [Getting Started Guide](getting-started.md) - Complete setup walkthrough
- [User Guide](../user_guide.md) - Full imperative mode documentation
- [Drift Detection Guide](drift-detection.md) - Complete drift detection reference
- [Security Alerting](security-alerting.md) - Severity classification details
- [Azure DevOps Integration](azure-devops-integration.md) - CI/CD pipeline examples
- [Pipeline Templates](../examples/pipelines/) - Ready-to-use automation
