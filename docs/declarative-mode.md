# Declarative Mode Guide

Declarative mode enables infrastructure-as-code workflows for Veeam Backup & Replication. Define configurations in YAML, track in Git, deploy to multiple environments, and detect drift.

## Table of Contents

- [Overview](#overview)
- [When to Use Declarative Mode](#when-to-use-declarative-mode)
- [Resource Types](#resource-types)
- [Key Concepts](#key-concepts)
- [Export Resources](#export-resources)
- [Apply Configurations](#apply-configurations)
- [State Management](#state-management)
- [Drift Detection](#drift-detection)
- [Configuration Overlays](#configuration-overlays)
- [Environment Configuration](#environment-configuration)
- [Strategic Merge Behavior](#strategic-merge-behavior)
- [Multi-Environment Workflow](#multi-environment-workflow)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Declarative mode allows you to:
- Define backup infrastructure as YAML configuration files
- Apply environment-specific overlays (prod, dev, staging)
- Track configurations in version control (Git)
- Detect configuration drift from desired state
- Manage multiple environments from a single base template
- Automate deployments with CI/CD pipelines

**When to use declarative mode:**
- Infrastructure-as-code workflows
- Multi-environment deployments (dev/staging/prod)
- Drift detection and monitoring
- GitOps automation
- Configuration standardization
- Disaster recovery planning

**When to use imperative mode:**
- Quick API queries
- Triggering one-off operations (start backup, stop job)
- Exploring API capabilities
- Working with products without declarative support

See [Imperative Mode Guide](imperative-mode.md) for imperative workflows.

## When to Use Declarative Mode

### Use Declarative Mode When:

✅ **Multi-environment deployments**
Deploy the same backup job to dev, staging, and production with environment-specific settings.

✅ **Configuration drift detection**
Monitor for unauthorized changes to backup infrastructure.

✅ **GitOps workflows**
Track all infrastructure changes in Git with full audit trail.

✅ **Standardization and compliance**
Enforce consistent backup policies across the organization.

✅ **Disaster recovery**
Quickly rebuild backup infrastructure from version-controlled configuration.

### Use Imperative Mode When:

❌ One-off operations (start/stop jobs)

❌ Quick API exploration

❌ Immediate troubleshooting

❌ Products without declarative support (VB365, VONE, Enterprise Manager)

## Resource Types

vcli supports declarative management for these VBR resource types:

| Resource | Commands | Create Support | Notes |
|----------|----------|----------------|-------|
| **Jobs** | `export`, `apply`, `diff` | ✅ Yes | Full create/update support |
| **Repositories** | `repo export`, `repo apply`, `repo snapshot`, `repo diff` | ❌ No | Update-only (create in VBR console) |
| **Scale-Out Repositories (SOBRs)** | `repo sobr-export`, `repo sobr-apply`, `repo sobr-snapshot`, `repo sobr-diff` | ❌ No | Update-only (create in VBR console) |
| **Encryption Passwords** | `encryption export`, `encryption snapshot`, `encryption diff` | ❌ No | Read-only (password values never exposed) |
| **KMS Servers** | `encryption kms-export`, `encryption kms-apply`, `encryption kms-snapshot`, `encryption kms-diff` | ❌ No | Update-only (create in VBR console) |

## Key Concepts

### 1. Base Configuration
A base YAML file defines common settings shared across all environments.

### 2. Overlays
Overlay files contain environment-specific changes that merge with the base.

### 3. Strategic Merge
vcli uses strategic merge to combine base + overlay:
- Maps are merged recursively (nested objects)
- Arrays are replaced (overlay replaces base)
- Labels/annotations are combined
- Base values preserved unless overridden

### 4. State Management
vcli maintains state in `state.json` to track:
- Last known configuration of each resource
- When snapshots were taken
- Configuration origin (applied vs adopted)

### 5. Drift Detection
Compare live VBR configuration against desired state to detect unauthorized changes with security-aware severity classification.

## Export Resources

Export existing VBR resources to YAML format for declarative management.

### Jobs

```bash
# Export single job by ID
vcli export 57b3baab-6237-41bf-add7-db63d41d984c -o my-job.yaml

# Export single job by name
vcli export "Database Backup" -o my-job.yaml

# Export all jobs to directory
vcli export --all -d ./jobs/

# Export as overlay (minimal changes only)
vcli export <id> --as-overlay -o overlay.yaml

# Export as overlay with base comparison
vcli export <id> --as-overlay --base base-job.yaml -o overlay.yaml

# Legacy simplified format (20-30 fields)
vcli export <id> --simplified -o job.yaml
```

**Full export (default):** Captures all VBR API fields (300+) ensuring complete job fidelity.

**Overlay export:** Exports minimal overlay format containing only fields that differ from base.

**Simplified export (legacy):** Minimal configuration (20-30 fields) for basic jobs only.

### Repositories

```bash
# Export single repository by name
vcli repo export "Default Backup Repository" -o repo.yaml

# Export all repositories
vcli repo export --all -d ./repos/
```

### Scale-Out Repositories (SOBRs)

```bash
# Export single SOBR by name
vcli repo sobr-export "SOBR-Production" -o sobr.yaml

# Export all SOBRs
vcli repo sobr-export --all -d ./sobrs/
```

### Encryption Passwords

```bash
# Export single encryption password by name
vcli encryption export "Production Encryption Key" -o enc.yaml

# Export all encryption passwords
vcli encryption export --all -d ./encryption/
```

**Note:** Password values are never exported. Only metadata (name, description, hint) is captured.

### KMS Servers

```bash
# Export single KMS server by name
vcli encryption kms-export "Azure Key Vault" -o kms.yaml

# Export all KMS servers
vcli encryption kms-export --all -d ./kms/
```

## Apply Configurations

Apply YAML configurations to VBR with optional dry-run preview.

### Jobs (Create or Update)

```bash
# Apply single job
vcli job apply my-job.yaml

# Preview changes without applying (recommended)
vcli job apply my-job.yaml --dry-run

# Apply with environment overlay
vcli job apply base-job.yaml -o prod-overlay.yaml

# Apply using environment from vcli.yaml
vcli job apply base-job.yaml --env production
```

**Note:** Jobs support both creation (POST) and updates (PUT).

### Repositories (Update-Only)

```bash
# Apply repository configuration
vcli repo apply repo.yaml

# Preview changes
vcli repo apply repo.yaml --dry-run

# Apply with overlay
vcli repo apply base-repo.yaml -o prod-repo-overlay.yaml
```

**Important:** Repositories must be created in VBR console first. Apply only updates existing repositories.

### Scale-Out Repositories (Update-Only)

```bash
# Apply SOBR configuration
vcli repo sobr-apply sobr.yaml

# Preview changes
vcli repo sobr-apply sobr.yaml --dry-run
```

**Important:** SOBRs must be created in VBR console first. Apply only updates existing SOBRs.

### KMS Servers (Update-Only)

```bash
# Apply KMS server configuration
vcli encryption kms-apply kms.yaml

# Preview changes
vcli encryption kms-apply kms.yaml --dry-run
```

**Important:** KMS servers must be created in VBR console first. Apply only updates existing KMS servers.

### Exit Codes

Apply commands return specific exit codes for automation:

| Code | Meaning | Action |
|------|---------|--------|
| `0` | Success | Continue |
| `1` | Error (API failure, invalid spec) | Fix and retry |
| `5` | Partial apply (some fields skipped) | Review skipped fields |
| `6` | Resource not found (update-only) | Create in VBR console first |

**Example usage in scripts:**
```bash
vcli repo apply repo.yaml
if [ $? -eq 6 ]; then
    echo "Repository doesn't exist - create it in VBR console first"
    exit 1
fi
```

## State Management

vcli maintains state in `state.json` to track resource configurations and enable drift detection.

### Snapshot Commands

Snapshot captures the current VBR configuration and saves it to state for drift detection.

**Repositories:**
```bash
# Snapshot single repository
vcli repo snapshot "Default Backup Repository"

# Snapshot all repositories
vcli repo snapshot --all
```

**SOBRs:**
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

**Note:** Jobs are snapshotted automatically when you apply them. Manual snapshots not needed for jobs.

### State File Format

`state.json` stores resource states:
```json
{
  "jobs": {
    "database-backup": {
      "id": "57b3baab-6237-41bf-add7-db63d41d984c",
      "config": { /* full job configuration */ },
      "lastSnapshot": "2024-01-15T10:30:00Z",
      "origin": "applied"
    }
  },
  "repositories": {
    "Default Backup Repository": {
      "id": "a1b2c3d4-...",
      "config": { /* full repo configuration */ },
      "lastSnapshot": "2024-01-15T10:31:00Z",
      "origin": "snapshot"
    }
  }
}
```

### Adopt Existing Resources

Adopt existing resources into declarative management without applying changes:

```bash
# Adopt repository (snapshot without changes)
vcli repo adopt "Default Backup Repository"

# Adopt all repositories
vcli repo adopt --all

# Adopt SOBR
vcli repo sobr-adopt "SOBR-Production"

# Adopt KMS server
vcli encryption kms-adopt "Azure Key Vault"
```

Adopt takes a snapshot and marks the resource as declaratively managed in `state.json` with `origin: "adopted"`.

## Drift Detection

Detect configuration drift from desired state with security-aware severity classification.

For full details see [Drift Detection Guide](drift-detection.md) and [Security Alerting](security-alerting.md).

### Detect Drift

```bash
# Single resource
vcli job diff "Database Backup"
vcli repo diff "Default Backup Repository"
vcli repo sobr-diff "SOBR-Production"
vcli encryption diff "Production Encryption Key"
vcli encryption kms-diff "Azure Key Vault"

# All resources of a type
vcli job diff --all
vcli repo diff --all
vcli repo sobr-diff --all
vcli encryption diff --all
vcli encryption kms-diff --all

# Filter by severity
vcli job diff --all --severity critical    # Only CRITICAL
vcli job diff --all --severity warning     # WARNING and above
vcli repo diff --all --security-only       # WARNING and above (alias)
```

### Severity Classification

Drifts are classified by security impact:

| Severity | Icon | Description | Exit Code |
|----------|------|-------------|-----------|
| **CRITICAL** | 🔴 | Security-impacting changes (encryption disabled, immutability off, GFS removed) | 4 |
| **WARNING** | ⚠️ | Important changes (retention reduced, compression changed, schedule modified) | 3 |
| **INFO** | ℹ️ | Minor changes (description, labels, non-security settings) | 3 |

### Exit Codes

Drift commands return specific exit codes:

| Code | Meaning | Action |
|------|---------|--------|
| `0` | No drift | Continue |
| `3` | Drift detected (INFO/WARNING) | Review and remediate if needed |
| `4` | Critical drift detected | Immediate remediation required |
| `1` | Error occurred | Check logs |

**Example usage in monitoring:**
```bash
vcli job diff --all --security-only
EXIT_CODE=$?

if [ $EXIT_CODE -eq 4 ]; then
    echo "CRITICAL security drift detected!"
    # Send alert, trigger remediation
    exit 1
elif [ $EXIT_CODE -eq 3 ]; then
    echo "WARNING: Drift detected, review required"
fi
```

## Configuration Overlays

Overlays enable multi-environment deployments from a single base template.

### Creating Overlays

Overlays contain only the fields you want to override from the base configuration.

**Example Base Configuration:**
```yaml
# base-backup.yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
  labels:
    app: database
    managed-by: vcli
spec:
  type: VSphereBackup
  description: Database backup job
  repository: default-repo
  storage:
    compression: Optimal
    retention:
      type: Days
      quantity: 7
  schedule:
    enabled: true
    daily: "22:00"
    retry:
      enabled: true
      times: 3
      wait: 10
  objects:
    - type: VirtualMachine
      name: db-server
      hostName: 192.168.0.14
```

**Production Overlay:**
```yaml
# overlays/prod-overlay.yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
  labels:
    env: production
spec:
  description: Production database backup (30-day retention)
  repository: prod-repo
  storage:
    retention:
      quantity: 30  # Override: 7 days -> 30 days
  schedule:
    daily: "02:00"  # Override: 22:00 -> 02:00
```

**Development Overlay:**
```yaml
# overlays/dev-overlay.yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
  labels:
    env: development
spec:
  description: Development database backup (3-day retention)
  repository: dev-repo
  storage:
    retention:
      quantity: 3   # Override: 7 days -> 3 days
  schedule:
    daily: "23:00"  # Override: 22:00 -> 23:00
```

### Preview Merged Configuration

Use the `plan` command to preview the merged configuration:

```bash
# Preview base configuration (no overlay)
vcli job plan base-backup.yaml

# Preview with production overlay
vcli job plan base-backup.yaml -o overlays/prod-overlay.yaml

# Preview with development overlay
vcli job plan base-backup.yaml -o overlays/dev-overlay.yaml

# Show full merged YAML
vcli job plan base-backup.yaml -o overlays/prod-overlay.yaml --show-yaml
```

**Output:**
The plan command displays:
- Resource name and type
- Base configuration file
- Overlay file (if used)
- Labels (merged from base + overlay)
- Key configuration fields
- Storage settings
- Schedule settings
- Backup objects list

### Apply with Overlay

```bash
# Preview changes (recommended)
vcli job apply base-backup.yaml -o overlays/prod-overlay.yaml --dry-run

# Apply to production
vcli job apply base-backup.yaml -o overlays/prod-overlay.yaml

# Apply to development
vcli job apply base-backup.yaml -o overlays/dev-overlay.yaml
```

## Environment Configuration

The `vcli.yaml` file manages environment-specific settings and overlay mappings.

### Configuration Structure

```yaml
# vcli.yaml
currentEnvironment: production
defaultOverlayDir: ./overlays

environments:
  production:
    overlay: prod-overlay.yaml
    profile: vbr-prod
    labels:
      env: production
      managed-by: vcli

  development:
    overlay: dev-overlay.yaml
    profile: vbr-dev
    labels:
      env: development
      managed-by: vcli

  staging:
    overlay: staging-overlay.yaml
    profile: vbr-staging
    labels:
      env: staging
      managed-by: vcli
```

### Configuration File Locations

vcli searches for `vcli.yaml` in this order:
1. Path in `VCLI_CONFIG` environment variable
2. Current directory (`./vcli.yaml`)
3. Home directory (`~/.vcli/vcli.yaml`)

### Using Environment Configuration

```bash
# Apply using currentEnvironment (production)
vcli job plan base-backup.yaml
vcli job apply base-backup.yaml

# Override with specific environment
vcli job plan base-backup.yaml --env development
vcli job apply base-backup.yaml --env development

# Explicit overlay takes precedence
vcli job plan base-backup.yaml -o custom-overlay.yaml
```

### Overlay Resolution Priority

1. `-o/--overlay` flag (highest priority)
2. `--env` flag (looks up in vcli.yaml)
3. `currentEnvironment` from vcli.yaml
4. No overlay (base config only)

## Strategic Merge Behavior

Understanding how overlays merge with base configurations:

### Scalar Values

Overlay value replaces base value.

```yaml
# Base
quantity: 7

# Overlay
quantity: 30

# Result
quantity: 30
```

### Nested Objects (Maps)

Deep merge - overlays are applied recursively.

```yaml
# Base
storage:
  compression: Optimal
  retention:
    type: Days
    quantity: 7

# Overlay
storage:
  retention:
    quantity: 30

# Result
storage:
  compression: Optimal      # Preserved from base
  retention:
    type: Days              # Preserved from base
    quantity: 30            # Overridden
```

### Arrays

Overlay array replaces base array completely.

```yaml
# Base
objects:
  - name: vm1
  - name: vm2

# Overlay
objects:
  - name: vm3

# Result
objects:
  - name: vm3              # Base array replaced
```

**Important:** Arrays are replaced, not merged. Include all desired items in the overlay array.

### Labels and Annotations

Combined (merged) from base and overlay.

```yaml
# Base
labels:
  app: database
  managed-by: vcli

# Overlay
labels:
  env: production

# Result
labels:
  app: database           # From base
  managed-by: vcli        # From base
  env: production         # From overlay
```

## Multi-Environment Workflow

### Project Structure

```
my-backups/
├── vcli.yaml
├── base-backup.yaml
├── overlays/
│   ├── prod-overlay.yaml
│   ├── dev-overlay.yaml
│   └── staging-overlay.yaml
└── specs/
    ├── jobs/
    ├── repos/
    ├── sobrs/
    └── kms/
```

### Complete Workflow Example

```bash
# 1. Export existing job as base template
vcli export <job-id> -o base-backup.yaml

# 2. Create overlays directory
mkdir overlays

# 3. Create environment overlays
# Edit overlays/prod-overlay.yaml, dev-overlay.yaml, etc.

# 4. Create vcli.yaml configuration
cat > vcli.yaml <<EOF
currentEnvironment: production
defaultOverlayDir: ./overlays
environments:
  production:
    overlay: prod-overlay.yaml
  development:
    overlay: dev-overlay.yaml
  staging:
    overlay: staging-overlay.yaml
EOF

# 5. Preview configurations
vcli job plan base-backup.yaml              # Uses production (currentEnvironment)
vcli job plan base-backup.yaml --env dev    # Preview development
vcli job plan base-backup.yaml --env staging # Preview staging

# 6. Show full merged YAML
vcli job plan base-backup.yaml --show-yaml

# 7. Apply configurations with dry-run first
vcli job apply base-backup.yaml --env production --dry-run
vcli job apply base-backup.yaml --env production

vcli job apply base-backup.yaml --env development --dry-run
vcli job apply base-backup.yaml --env development

# 8. Commit to version control
git add .
git commit -m "Add multi-environment backup configuration"
git push
```

### Bootstrap Declarative Management

Start managing existing VBR infrastructure declaratively:

```bash
# 1. Export current VBR state
vcli export --all -d specs/jobs/
vcli repo export --all -d specs/repos/
vcli repo sobr-export --all -d specs/sobrs/
vcli encryption kms-export --all -d specs/kms/

# 2. Snapshot current state
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all

# 3. Commit to Git (state.json and specs)
git add specs/ state.json
git commit -m "Bootstrap VBR declarative management"
git push

# 4. Verify no drift
vcli job diff --all
vcli repo diff --all
vcli repo sobr-diff --all
vcli encryption diff --all
vcli encryption kms-diff --all
```

## Best Practices

### 1. Keep Base Templates DRY
Only include common settings in base files. Environment-specific values go in overlays.

**Good:**
```yaml
# base-backup.yaml - common settings only
spec:
  type: VSphereBackup
  storage:
    compression: Optimal
```

**Bad:**
```yaml
# base-backup.yaml - don't include environment-specific values
spec:
  repository: prod-repo  # This varies by environment!
```

### 2. Small Overlays
Override only what differs per environment.

**Good:**
```yaml
# prod-overlay.yaml
spec:
  repository: prod-repo
  storage:
    retention:
      quantity: 30
```

**Bad:**
```yaml
# prod-overlay.yaml - duplicates too much from base
spec:
  type: VSphereBackup  # Unnecessary - same as base
  storage:
    compression: Optimal  # Unnecessary - same as base
    retention:
      type: Days  # Unnecessary - same as base
      quantity: 30  # This is the only change needed
```

### 3. Use Labels
Tag resources with environment, app, and managed-by labels.

```yaml
metadata:
  labels:
    app: database
    env: production
    managed-by: vcli
    team: infrastructure
```

### 4. Version Control
Commit both base and overlays to Git.

```bash
git add base-backup.yaml overlays/ vcli.yaml
git commit -m "Update production retention to 30 days"
git push
```

### 5. Preview First
Always use plan or --dry-run before applying.

```bash
# Preview merged configuration
vcli job plan base-backup.yaml -o prod-overlay.yaml --show-yaml

# Dry-run apply
vcli job apply base-backup.yaml -o prod-overlay.yaml --dry-run

# Apply for real
vcli job apply base-backup.yaml -o prod-overlay.yaml
```

### 6. Consistent Naming
Use clear, descriptive names for base files and overlays.

**Good:**
- `base-database-backup.yaml`
- `overlays/prod-database.yaml`
- `overlays/dev-database.yaml`

**Bad:**
- `backup1.yaml`
- `overlay.yaml`
- `prod.yaml`

### 7. Document Changes
Use Git commit messages to explain configuration changes.

```bash
git commit -m "Increase production retention to 30 days to meet compliance requirements"
```

### 8. Test in Dev First
Apply to development environment before production.

```bash
# 1. Test in development
vcli job apply base-backup.yaml --env development --dry-run
vcli job apply base-backup.yaml --env development

# 2. Verify in development
# Run test backup, verify results

# 3. Apply to production
vcli job apply base-backup.yaml --env production --dry-run
vcli job apply base-backup.yaml --env production
```

### 9. Regular Drift Scans
Schedule drift detection to catch unauthorized changes.

```bash
# Daily drift scan
vcli job diff --all --security-only
vcli repo diff --all --security-only
vcli repo sobr-diff --all --security-only
vcli encryption kms-diff --all --security-only
```

### 10. Snapshot Before Changes
Snapshot resources before making changes.

```bash
# Before applying changes
vcli repo snapshot --all
vcli repo sobr-snapshot --all

# Apply changes
vcli repo apply repo.yaml
vcli repo sobr-apply sobr.yaml

# Verify no unexpected drift
vcli repo diff --all
vcli repo sobr-diff --all
```

## Troubleshooting

### Overlay Not Being Applied

**Problem:** Overlay seems to be ignored.

**Solutions:**
1. Check overlay resolution priority:
   - Explicit `-o` flag has highest priority
   - `--env` flag looks up environment in vcli.yaml
   - `currentEnvironment` in vcli.yaml is used if no flags
2. Verify vcli.yaml exists and is in search path:
   - Check `VCLI_CONFIG` environment variable
   - Check current directory (`./vcli.yaml`)
   - Check home directory (`~/.vcli/vcli.yaml`)
3. Confirm environment exists in vcli.yaml
4. Use `--show-yaml` to see the actual merged result:
   ```bash
   vcli job plan base.yaml -o overlay.yaml --show-yaml
   ```

### Unexpected Merge Results

**Problem:** Merged configuration doesn't look right.

**Solutions:**
1. Remember: arrays are replaced, not merged
   ```yaml
   # If your overlay has an objects array, it replaces the entire base array
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

**Problem:** Labels from base not appearing in merged config.

**Solutions:**
1. Ensure both base and overlay use `metadata.labels` field
2. Labels should be at same level in both files:
   ```yaml
   # Both files should have:
   metadata:
     labels:
       key: value
   ```

### Resource Not Found (Exit Code 6)

**Problem:** Apply fails with "resource not found" error.

**Solution:**
Repositories, SOBRs, and KMS servers are update-only. Create them in VBR console first:
```bash
# This will fail if repository doesn't exist
vcli repo apply repo.yaml  # Exit code 6

# Create the repository in VBR console first, then:
vcli repo apply repo.yaml  # Now succeeds
```

### Drift Detected After Apply

**Problem:** Diff shows drift immediately after applying.

**Solutions:**
1. Some fields may not be updateable via API (skipped during apply)
2. Check for fields with exit code 5 (partial apply)
3. VBR may normalize certain values (e.g., schedule times)
4. Review the specific drifts - may be INFO severity and acceptable

### State File Out of Sync

**Problem:** State file doesn't match VBR.

**Solution:**
Re-snapshot resources to refresh state:
```bash
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all
vcli encryption kms-snapshot --all
```

### Encryption Password Apply Not Supported

**Problem:** Can't find `encryption apply` command.

**Explanation:**
Encryption passwords are read-only. Password values are never exposed by the VBR API, so updates aren't supported. Only metadata (name, description, hint) can be exported for documentation purposes.

## See Also

- [Command Reference](command-reference.md) - Quick command lookup
- [Drift Detection Guide](drift-detection.md) - Complete drift detection reference
- [Security Alerting](security-alerting.md) - Severity classification details
- [Azure DevOps Integration](azure-devops-integration.md) - CI/CD pipeline examples
- [Authentication Guide](authentication.md) - Setup and credentials
- [Getting Started](getting-started.md) - Complete setup guide
- [Pipeline Templates](../examples/pipelines/) - Ready-to-use automation
