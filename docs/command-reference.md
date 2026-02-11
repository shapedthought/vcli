# Command Quick Reference

Fast reference for common owlctl commands. See full documentation in [User Guide](../user_guide.md) and [Drift Detection Guide](drift-detection.md).

## Table of Contents

- [Setup & Authentication](#setup--authentication)
- [Imperative Commands (All Products)](#imperative-commands-all-products)
- [Declarative Commands (VBR Only)](#declarative-commands-vbr-only)
- [Group Commands](#group-commands)
- [Instance Commands](#instance-commands)
- [Target Commands](#target-commands)
- [Common Flags](#common-flags)
- [Exit Codes](#exit-codes)

---

## Setup & Authentication

```bash
# Initialize (first time only)
owlctl init

# Set profile
owlctl profile --set vbr                    # VBR
owlctl profile --set ent_man                # Enterprise Manager
owlctl profile --set vb365                  # VB365
owlctl profile --list                       # Show all profiles

# Set credentials (environment variables)
export OWLCTL_USERNAME="administrator"
export OWLCTL_PASSWORD="your-password"
export OWLCTL_URL="vbr.example.com"

# Login
owlctl login
```

---

## Imperative Commands (All Products)

### GET - Retrieve Data

```bash
# List all resources
owlctl get jobs
owlctl get repositories
owlctl get backupInfrastructure/managedServers

# Get specific resource
owlctl get jobs/<job-id>
owlctl get repositories/<repo-id>

# Output formats
owlctl get jobs                             # JSON (default)
owlctl get jobs --yaml                      # YAML
owlctl get jobs | jq '.data'                # Parse with jq
```

### POST - Trigger Operations

```bash
# Without payload
owlctl post jobs/<job-id>/start
owlctl post jobs/<job-id>/stop
owlctl post jobs/<job-id>/retry

# With payload
owlctl post jobs -f job-data.json
owlctl post repositories -f repo-data.json
```

### PUT - Update Resources

```bash
# Update resource (requires payload)
owlctl put jobs/<job-id> -f updated-job.json
owlctl put repositories/<repo-id> -f updated-repo.json
```

### Common Endpoints (VBR)

```bash
# Jobs
owlctl get jobs                             # All jobs
owlctl get jobs/<id>                        # Job details
owlctl post jobs/<id>/start                 # Start job
owlctl post jobs/<id>/stop                  # Stop job

# Repositories
owlctl get backupInfrastructure/repositories
owlctl get backupInfrastructure/scaleOutRepositories

# Infrastructure
owlctl get backupInfrastructure/managedServers
owlctl get backupInfrastructure/credentials

# Sessions
owlctl get sessions                         # Recent sessions
owlctl get sessions/<id>                    # Session details
```

---

## Declarative Commands (VBR Only)

### Export Resources

```bash
# Jobs
owlctl export <job-id> -o job.yaml          # Export single job
owlctl export --all -d jobs/                # Export all jobs
owlctl export <id> --as-overlay -o ov.yaml  # Export as overlay
owlctl export <id> --simplified -o job.yaml # Minimal format (legacy)

# Repositories
owlctl repo export <name> -o repo.yaml
owlctl repo export --all -d repos/
owlctl repo export <name> --as-overlay --base base-repo.yaml -o overlay.yaml

# Scale-Out Repositories (SOBRs)
owlctl repo sobr-export <name> -o sobr.yaml
owlctl repo sobr-export --all -d sobrs/
owlctl repo sobr-export <name> --as-overlay --base base-sobr.yaml -o overlay.yaml

# Encryption Passwords (read-only, no overlay support)
owlctl encryption export <name> -o enc.yaml
owlctl encryption export --all -d encryption/

# KMS Servers
owlctl encryption kms-export <name> -o kms.yaml
owlctl encryption kms-export --all -d kms/
owlctl encryption kms-export <name> --as-overlay --base base-kms.yaml -o overlay.yaml
```

### Apply Configurations

```bash
# Jobs (create or update)
owlctl job apply job.yaml
owlctl job apply job.yaml --dry-run         # Preview changes
owlctl job apply base.yaml -o prod.yaml     # Apply with overlay

# Repositories (update-only)
owlctl repo apply repo.yaml
owlctl repo apply repo.yaml --dry-run

# SOBRs (update-only)
owlctl repo sobr-apply sobr.yaml
owlctl repo sobr-apply sobr.yaml --dry-run

# KMS Servers (update-only)
owlctl encryption kms-apply kms.yaml
owlctl encryption kms-apply kms.yaml --dry-run
```

**Note:** Repos, SOBRs, and KMS are update-only. Create them in VBR console first.

### Snapshot State

```bash
# Repositories
owlctl repo snapshot <name>
owlctl repo snapshot --all

# SOBRs
owlctl repo sobr-snapshot <name>
owlctl repo sobr-snapshot --all

# Encryption Passwords
owlctl encryption snapshot <name>
owlctl encryption snapshot --all

# KMS Servers
owlctl encryption kms-snapshot <name>
owlctl encryption kms-snapshot --all
```

**Note:** Jobs are snapshotted automatically on apply.

### Detect Drift

```bash
# Single resource
owlctl job diff "Job Name"
owlctl repo diff "Repository Name"
owlctl repo sobr-diff "SOBR Name"
owlctl encryption diff "Password Name"
owlctl encryption kms-diff "KMS Server Name"

# All resources
owlctl job diff --all
owlctl repo diff --all
owlctl repo sobr-diff --all
owlctl encryption diff --all
owlctl encryption kms-diff --all

# Severity filtering
owlctl job diff --all --severity critical   # Only CRITICAL
owlctl job diff --all --security-only       # WARNING and above
owlctl repo diff --all --severity warning   # WARNING and above
```

### Plan (Preview)

```bash
# Preview merged configuration
owlctl job plan base.yaml
owlctl job plan base.yaml -o prod.yaml
owlctl job plan base.yaml -o prod.yaml --show-yaml
```

---

## Group Commands

Groups bundle specs with a shared profile, overlay, and optional instance for batch operations. Defined in `owlctl.yaml`.

```bash
# List all groups (shows instance, profile, overlay, spec count)
owlctl group list

# Show group details (resolved paths, spec count, instance)
owlctl group show sql-tier
```

### Apply with Group

```bash
# Apply all specs in a group (profile + spec + overlay merge)
owlctl job apply --group sql-tier

# Dry-run first
owlctl job apply --group sql-tier --dry-run
```

### Diff with Group

```bash
# Drift-check all specs in a group against live VBR
owlctl job diff --group sql-tier

# Group diff does NOT require state.json — the group definition is the source of truth
```

**Note:** `--group` is mutually exclusive with positional file args, `-o/--overlay`, `--env`, and `--all`.

---

## Instance Commands

Instances define named server connections with product type, credentials, and TLS settings. They replace `--target` for multi-server workflows. Defined in `owlctl.yaml`.

```bash
# List all instances
owlctl instance list

# Show instance details (product, URL, credential ref)
owlctl instance show vbr-prod
```

### Using --instance

```bash
# Run any command against a named instance
owlctl --instance vbr-prod get jobs
owlctl --instance vbr-prod login

# Apply group (group can also specify instance in owlctl.yaml)
owlctl job apply --group sql-tier --instance vbr-prod
```

---

## Target Commands (Deprecated)

> **Deprecated:** Use `--instance` instead. Targets only set the URL; instances also handle product type, credentials, and per-instance token caching.

Targets define named VBR server connections in `owlctl.yaml`. Use `--target` to switch between servers.

```bash
# List all targets
owlctl target list

# List as JSON (for scripting)
owlctl target list --json
```

### Multi-Target Workflow

```bash
# Apply to production VBR (deprecated — use --instance instead)
owlctl job apply --group sql-tier --target primary

# Apply same group to DR site
owlctl job apply --group sql-tier --target dr
```

---

## Common Flags

### Apply Commands

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview changes without applying |
| `-o, --overlay <file>` | Apply with configuration overlay |
| `--group <name>` | Apply all specs in named group (from `owlctl.yaml`) |
| `--env <name>` | Legacy flag; supported for backwards compatibility. Prefer `--group`. |

### Diff Commands

| Flag | Description |
|------|-------------|
| `--all` | Check all resources |
| `--group <name>` | Check drift for all specs in named group |
| `--severity <level>` | Filter by severity: `critical`, `warning`, or `info` |
| `--security-only` | Show only WARNING and CRITICAL drifts |

### Export Commands

| Flag | Description |
|------|-------------|
| `-o, --output <file>` | Output file path |
| `-d, --directory <dir>` | Output directory (for --all) |
| `--all` | Export all resources |
| `--as-overlay` | Export as minimal overlay (jobs, repos, SOBRs, KMS) |
| `--base <file>` | Base file for overlay comparison (with --as-overlay) |
| `--simplified` | Minimal format - legacy (jobs only) |

### Global Flags

| Flag | Description |
|------|-------------|
| `--yaml` | Output in YAML format (default: JSON) |
| `--instance <name>` | Use named instance from `owlctl.yaml` (sets URL, credentials, product) |
| `--target <name>` | Use named VBR target from `owlctl.yaml` (deprecated, use `--instance`) |
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
owlctl repo apply repo.yaml
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
owlctl job diff --all --security-only
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
export OWLCTL_USERNAME="admin" OWLCTL_PASSWORD="pass" OWLCTL_URL="vbr.local"
owlctl profile --set vbr
owlctl login

# 2. Query
owlctl get jobs | jq '.data[] | {name: .name, type: .type}'
```

### Export All Configurations

```bash
# Export everything to Git
owlctl export --all -d specs/jobs/
owlctl repo export --all -d specs/repos/
owlctl repo sobr-export --all -d specs/sobrs/
owlctl encryption kms-export --all -d specs/kms/

# Commit to Git
git add specs/
git commit -m "Snapshot VBR configuration"
git push
```

### Group-Based Deployment

```bash
# List available groups
owlctl group list

# Preview a group's configuration
owlctl group show sql-tier

# Dry-run, then apply
owlctl job apply --group sql-tier --dry-run
owlctl job apply --group sql-tier

# Check drift for the group
owlctl job diff --group sql-tier
```

### Multi-Instance Deployment

```bash
# Apply group to production VBR
owlctl job apply --group sql-tier --instance vbr-prod

# Apply same group to DR site
owlctl job apply --group sql-tier --instance vbr-dr

# Drift check across both instances
owlctl job diff --group sql-tier --instance vbr-prod
owlctl job diff --group sql-tier --instance vbr-dr

# Or define the instance on the group in owlctl.yaml:
# groups:
#   prod-jobs:
#     instance: vbr-prod
#     specsDir: specs/jobs/
owlctl job apply --group prod-jobs    # uses vbr-prod automatically
```

### Single-File Overlay (Simpler Alternative)

```bash
# Preview production changes
owlctl job plan base-backup.yaml -o overlays/prod.yaml --show-yaml

# Apply production (dry-run first)
owlctl job apply base-backup.yaml -o overlays/prod.yaml --dry-run
owlctl job apply base-backup.yaml -o overlays/prod.yaml
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
    owlctl $cmd
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
owlctl export --all -d infrastructure/jobs/
owlctl repo export --all -d infrastructure/repos/
owlctl repo sobr-export --all -d infrastructure/sobrs/
owlctl encryption kms-export --all -d infrastructure/kms/

# 2. Snapshot current state
owlctl repo snapshot --all
owlctl repo sobr-snapshot --all
owlctl encryption snapshot --all
owlctl encryption kms-snapshot --all

# 3. Commit to Git (state.json and specs)
git add infrastructure/ state.json
git commit -m "Bootstrap VBR declarative management"
git push

# 4. Verify no drift
owlctl job diff --all
owlctl repo diff --all
```

---

## Tips

### Output Parsing

```bash
# JSON (default)
owlctl get jobs | jq '.data[0].name'

# YAML
owlctl get jobs --yaml | yq '.data[0].name'

# With Nushell
owlctl get jobs | from json | get data | where isDisabled == false
```

### Credential Management

```bash
# Temporary credentials (current shell only)
export OWLCTL_USERNAME="admin" OWLCTL_PASSWORD="pass" OWLCTL_URL="vbr.local"

# Persistent credentials (bash/zsh)
echo 'export OWLCTL_USERNAME="admin"' >> ~/.bashrc
echo 'export OWLCTL_PASSWORD="pass"' >> ~/.bashrc
echo 'export OWLCTL_URL="vbr.local"' >> ~/.bashrc

# Use .env file (with direnv or similar)
# .env
OWLCTL_USERNAME=admin
OWLCTL_PASSWORD=pass
OWLCTL_URL=vbr.local
```

### Configuration Directory

```bash
# Set custom config location
export OWLCTL_SETTINGS_PATH="$HOME/.owlctl/"

# Initialize in custom location
mkdir -p ~/.owlctl
cd ~/.owlctl
owlctl init
```

---

## See Also

- [Getting Started Guide](getting-started.md) - Complete setup walkthrough
- [User Guide](../user_guide.md) - Full imperative mode documentation
- [Drift Detection Guide](drift-detection.md) - Complete drift detection reference
- [Security Alerting](security-alerting.md) - Severity classification details
- [Azure DevOps Integration](azure-devops-integration.md) - CI/CD pipeline examples
- [Pipeline Templates](../examples/pipelines/) - Ready-to-use automation
