# State Management Guide

vcli maintains state in `state.json` to enable drift detection and track declarative resource management. This guide explains how state management works and best practices for using it.

## Table of Contents

- [Overview](#overview)
- [What is State?](#what-is-state)
- [State File Location](#state-file-location)
- [State File Format](#state-file-format)
- [Creating State](#creating-state)
- [State Origins](#state-origins)
- [Updating State](#updating-state)
- [Using State for Drift Detection](#using-state-for-drift-detection)
- [State in Git Workflows](#state-in-git-workflows)
- [State File Management](#state-file-management)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

## Overview

State management enables:
- **Drift detection** - Compare live VBR configuration against desired state
- **Change tracking** - Know when configurations were last captured
- **Origin tracking** - Distinguish between applied, adopted, and exported resources
- **GitOps workflows** - Version control your infrastructure state

## What is State?

State is vcli's record of the last known configuration for each managed resource. When you snapshot or apply a resource, vcli saves its configuration to `state.json`. Later, the diff command compares the current VBR configuration against this saved state to detect drift.

**Key concepts:**
- **Snapshot** - Captures current VBR configuration and saves to state
- **Apply** - Updates VBR and automatically snapshots the new configuration
- **Adopt** - Takes a snapshot without making changes (for existing resources)
- **Drift** - Differences between state and current VBR configuration

## State File Location

By default, `state.json` is created in the current directory. Use `VCLI_SETTINGS_PATH` to set a custom location:

```bash
# Set custom location
export VCLI_SETTINGS_PATH="$HOME/.vcli/"

# state.json will be created at:
# $HOME/.vcli/state.json
```

**Important:** The state file must be in the same directory as `settings.json` and `profiles.json`.

## State File Format

`state.json` stores resource configurations organized by type:

```json
{
  "jobs": {
    "Database Backup": {
      "id": "57b3baab-6237-41bf-add7-db63d41d984c",
      "config": {
        "name": "Database Backup",
        "type": "VSphereBackup",
        "repository": "prod-repo",
        "storage": {
          "compression": "Optimal",
          "retention": {
            "type": "Days",
            "quantity": 30
          }
        }
        // ... full job configuration
      },
      "lastSnapshot": "2024-01-15T10:30:00Z",
      "origin": "applied"
    }
  },
  "repositories": {
    "Default Backup Repository": {
      "id": "a1b2c3d4-5678-90ab-cdef-1234567890ab",
      "config": {
        "name": "Default Backup Repository",
        "type": "LinuxLocal",
        "path": "/backup",
        "immutabilityEnabled": true
        // ... full repository configuration
      },
      "lastSnapshot": "2024-01-15T10:31:00Z",
      "origin": "snapshot"
    }
  },
  "sobrs": {
    "SOBR-Production": {
      "id": "b2c3d4e5-6789-01bc-def1-234567890abc",
      "config": {
        "name": "SOBR-Production",
        "extents": [
          // ... extent configurations
        ]
      },
      "lastSnapshot": "2024-01-15T10:32:00Z",
      "origin": "adopted"
    }
  },
  "encryptionPasswords": {
    "Production Encryption Key": {
      "id": "c3d4e5f6-7890-12cd-ef12-34567890abcd",
      "config": {
        "description": "Production encryption password",
        "hint": "Company standard key"
        // Password value never stored
      },
      "lastSnapshot": "2024-01-15T10:33:00Z",
      "origin": "snapshot"
    }
  },
  "kmsServers": {
    "Azure Key Vault": {
      "id": "d4e5f6g7-8901-23de-f123-4567890abcde",
      "config": {
        "name": "Azure Key Vault",
        "type": "Azure",
        "endpoint": "https://vault.azure.net"
        // ... KMS configuration
      },
      "lastSnapshot": "2024-01-15T10:34:00Z",
      "origin": "applied"
    }
  }
}
```

**Field definitions:**
- **id** - VBR resource ID (UUID)
- **config** - Full resource configuration as JSON
- **lastSnapshot** - ISO 8601 timestamp of last snapshot
- **origin** - How the resource entered state management (see below)

## Creating State

### Snapshot Commands

Snapshot captures the current VBR configuration and saves it to state without making changes.

**Jobs:**
```bash
# Jobs are automatically snapshotted on apply
# Manual snapshot not needed for jobs
```

**Repositories:**
```bash
# Snapshot single repository
vcli repo snapshot "Default Backup Repository"

# Snapshot all repositories
vcli repo snapshot --all
```

**Scale-Out Repositories:**
```bash
# Snapshot single SOBR
vcli repo sobr-snapshot "SOBR-Production"

# Snapshot all SOBRs
vcli repo sobr-snapshot --all
```

**Encryption Passwords:**
```bash
# Snapshot single encryption password
vcli encryption snapshot "Production Encryption Key"

# Snapshot all encryption passwords
vcli encryption snapshot --all
```

**KMS Servers:**
```bash
# Snapshot single KMS server
vcli encryption kms-snapshot "Azure Key Vault"

# Snapshot all KMS servers
vcli encryption kms-snapshot --all
```

### Adopt Commands

Adopt takes a snapshot of an existing resource to bring it under declarative management without making any changes. This is useful when you want to start managing existing resources declaratively.

```bash
# Adopt repository
vcli repo adopt "Default Backup Repository"
vcli repo adopt --all

# Adopt SOBR
vcli repo sobr-adopt "SOBR-Production"
vcli repo sobr-adopt --all

# Adopt KMS server
vcli encryption kms-adopt "Azure Key Vault"
vcli encryption kms-adopt --all
```

**Adopt vs Snapshot:**
- Both capture current configuration and save to state
- Adopt marks `origin: "adopted"` in state
- Snapshot marks `origin: "snapshot"` in state
- Functionally identical, but origin tracking helps understand how resources were added

### Apply Commands

Apply automatically snapshots the configuration after successfully updating VBR.

```bash
# Apply job (automatically snapshots on success)
vcli job apply job.yaml

# Apply repository (automatically snapshots on success)
vcli repo apply repo.yaml

# Apply SOBR (automatically snapshots on success)
vcli repo sobr-apply sobr.yaml

# Apply KMS server (automatically snapshots on success)
vcli encryption kms-apply kms.yaml
```

**Note:** Jobs are only snapshotted via apply. There is no manual job snapshot command.

## State Origins

The `origin` field in state tracks how a resource entered declarative management:

| Origin | Meaning | How Created |
|--------|---------|-------------|
| `applied` | Configuration was applied via YAML | `vcli job apply`, `vcli repo apply`, etc. |
| `adopted` | Existing resource adopted into management | `vcli repo adopt`, `vcli repo sobr-adopt`, etc. |
| `snapshot` | Manual snapshot taken | `vcli repo snapshot`, `vcli encryption snapshot`, etc. |
| `exported` | Exported to YAML but not yet applied | `vcli export`, `vcli repo export`, etc. |

**Origin is informational only** - it doesn't affect drift detection or other operations. It helps you understand the history of each resource.

## Updating State

State is automatically updated when you:

1. **Apply a configuration** - Updates state with new configuration
2. **Take a snapshot** - Updates state with current VBR configuration
3. **Adopt a resource** - Adds resource to state with current configuration

**Manual state updates:**
To refresh state for all resources:
```bash
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all
```

**Important:** Jobs are only updated via apply. You cannot manually snapshot jobs.

## Using State for Drift Detection

Drift detection compares the current VBR configuration against state:

```bash
# Detect drift for single resource
vcli job diff "Database Backup"
vcli repo diff "Default Backup Repository"

# Detect drift for all resources
vcli job diff --all
vcli repo diff --all
vcli repo sobr-diff --all
vcli encryption diff --all
vcli encryption kms-diff --all
```

**How it works:**
1. vcli reads the resource from `state.json`
2. vcli fetches the current configuration from VBR API
3. vcli compares the two configurations field-by-field
4. vcli reports any differences with severity classification

**Prerequisites:**
- Resource must exist in state.json
- VBR must be accessible
- Profile must be set correctly

**If state doesn't exist:**
```bash
# Error: No state found for "Database Backup"
# Solution: Take a snapshot or apply a configuration first
vcli repo snapshot "Default Backup Repository"
```

See [Drift Detection Guide](drift-detection.md) for complete details.

## State in Git Workflows

### Committing State to Git

**Recommended approach:**
```bash
# 1. Export all configurations to YAML
vcli export --all -d specs/jobs/
vcli repo export --all -d specs/repos/
vcli repo sobr-export --all -d specs/sobrs/
vcli encryption kms-export --all -d specs/kms/

# 2. Snapshot all resources
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all

# 3. Commit both specs and state
git add specs/ state.json
git commit -m "Snapshot VBR configuration"
git push
```

**Benefits:**
- Full audit trail of infrastructure changes
- Point-in-time recovery of configurations
- Team visibility into infrastructure state
- Foundation for GitOps workflows

### GitOps Workflow Example

**Daily snapshot automation:**
```bash
#!/bin/bash
# daily-snapshot.sh

# Take snapshots
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all

# Commit if changes detected
if ! git diff --quiet state.json; then
    git add state.json
    git commit -m "Daily snapshot: $(date +%Y-%m-%d)"
    git push
fi
```

**CI/CD drift detection:**
```yaml
# Azure DevOps pipeline
- script: |
    vcli job diff --all --security-only
    EXIT_CODE=$?
    if [ $EXIT_CODE -eq 4 ]; then
      echo "##vso[task.logissue type=error]CRITICAL drift detected"
      exit 1
    fi
  displayName: 'Detect Security Drift'
```

### State File Merge Conflicts

If multiple people update state simultaneously, Git merge conflicts can occur.

**Resolution strategy:**
```bash
# 1. Accept the version with most recent timestamps
git checkout --theirs state.json
git add state.json

# 2. Re-snapshot to ensure state is current
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all

# 3. Commit the resolved state
git add state.json
git commit -m "Resolve state merge conflict with fresh snapshot"
```

**Prevention:**
- Use a single automation pipeline for state updates
- Coordinate manual state updates across team
- Consider separate state files per environment

## State File Management

### Backing Up State

**Manual backup:**
```bash
# Backup before major changes
cp state.json state.json.backup-$(date +%Y%m%d)
```

**Automated backup:**
```bash
# Keep last 7 days of state backups
#!/bin/bash
BACKUP_DIR="$HOME/.vcli/backups"
mkdir -p "$BACKUP_DIR"

# Create backup
cp state.json "$BACKUP_DIR/state-$(date +%Y%m%d-%H%M%S).json"

# Keep only last 7 days
find "$BACKUP_DIR" -name "state-*.json" -mtime +7 -delete
```

### Restoring State

**From backup:**
```bash
# Restore from backup
cp state.json.backup-20240115 state.json
```

**From Git:**
```bash
# Restore from specific commit
git show abc123:state.json > state.json
```

**Rebuild from VBR:**
```bash
# If state is lost, rebuild from current VBR configuration
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all
```

### Cleaning State

Remove resources that no longer exist in VBR:

**Manual cleanup:**
```bash
# Edit state.json and remove entries for deleted resources
vim state.json
```

**Automated cleanup (future feature):**
```bash
# Not yet implemented
# vcli state clean --remove-missing
```

### Migrating State

**Moving to a new environment:**
```bash
# 1. Export state from old environment
cp $HOME/.vcli/state.json state-old-env.json

# 2. Set up new environment
export VCLI_SETTINGS_PATH="$HOME/.vcli-new/"
vcli init

# 3. Take fresh snapshots in new environment
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all
```

**Don't copy state between environments** - Resource IDs differ between VBR servers.

## Troubleshooting

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

### State Shows Drift After Apply

**Problem:** Diff shows drift immediately after applying

**Possible causes:**
1. Some fields may not be updateable via API (VBR normalizes values)
2. VBR may add default values not in the spec
3. Timing issue (state captured before VBR fully updated)

**Solutions:**
```bash
# Re-snapshot after a short delay
sleep 5
vcli repo snapshot "Resource Name"

# Check if drift is INFO severity (may be acceptable)
vcli repo diff "Resource Name"
```

## Best Practices

### 1. Commit State to Git

Always version control your state file alongside configuration specs:
```bash
git add specs/ state.json
git commit -m "Update VBR configuration"
git push
```

### 2. Snapshot Regularly

Schedule regular snapshots to keep state current:
```bash
# Daily cron job
0 2 * * * cd /path/to/vcli && ./daily-snapshot.sh
```

### 3. Snapshot Before Changes

Take a snapshot before making manual changes in VBR console:
```bash
# Before making changes in VBR console
vcli repo snapshot --all

# Make changes in VBR console

# Check what changed
vcli repo diff --all
```

### 4. Use Adopt for Existing Resources

When starting declarative management, adopt existing resources:
```bash
# Adopt all existing resources
vcli repo adopt --all
vcli repo sobr-adopt --all
vcli encryption kms-adopt --all

# Commit initial state
git add state.json
git commit -m "Adopt existing VBR resources into declarative management"
git push
```

### 5. Separate State by Environment

Use different state files for different environments:
```bash
# Production
export VCLI_SETTINGS_PATH="$HOME/.vcli/prod/"
vcli repo snapshot --all

# Development
export VCLI_SETTINGS_PATH="$HOME/.vcli/dev/"
vcli repo snapshot --all
```

### 6. Backup State Before Major Changes

```bash
# Before major infrastructure changes
cp state.json state.json.backup-$(date +%Y%m%d)

# Make changes
vcli repo apply repo.yaml

# Verify
vcli repo diff --all
```

### 7. Don't Edit State Manually

Let vcli manage state.json. Manual edits can corrupt the file or introduce inconsistencies.

**Exception:** Removing entries for deleted resources is safe.

### 8. Keep State in Sync

If you make changes directly in VBR console, update state immediately:
```bash
# After console changes
vcli repo snapshot "Resource Name"
vcli job diff "Job Name"  # Verify no unexpected drift
```

### 9. Use State for Disaster Recovery

State combined with YAML specs enables full infrastructure recovery:
```bash
# 1. Store in Git
git clone https://github.com/company/vbr-config.git
cd vbr-config

# 2. Apply all configurations
for job in specs/jobs/*.yaml; do
    vcli job apply "$job"
done

for repo in specs/repos/*.yaml; do
    vcli repo apply "$repo"
done
```

### 10. Monitor State File Size

Large environments can produce large state files. Monitor and compress if needed:
```bash
# Check state file size
ls -lh state.json

# Compress for long-term storage
gzip < state.json > state-$(date +%Y%m%d).json.gz
```

## See Also

- [Declarative Mode Guide](declarative-mode.md) - Infrastructure-as-code workflows
- [Drift Detection Guide](drift-detection.md) - Configuration monitoring
- [Security Alerting](security-alerting.md) - Severity classification
- [Azure DevOps Integration](azure-devops-integration.md) - CI/CD automation
- [Getting Started](getting-started.md) - Initial setup
