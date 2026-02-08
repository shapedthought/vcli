# owlctl Configuration Examples

This directory contains example configuration files demonstrating owlctl's declarative management features for all VBR resource types.

## Overlay System

The overlay system allows you to define a base template with common settings and then apply environment-specific overlays. This enables DRY (Don't Repeat Yourself) configuration management across multiple environments.

### Example: Multi-Environment Database Backup

#### Base Template (`overlay-base.yaml`)
The base template defines common settings shared across all environments:
- Backup type and VM selection
- Default repository
- Base retention policy (7 days)
- Schedule with retry configuration

#### Environment Overlays

**Production Overlay (`overlay-prod.yaml`)**
- Extends retention to 30 days
- Changes schedule to 02:00
- Uses production repository
- Adds production label

**Development Overlay (`overlay-dev.yaml`)**
- Reduces retention to 3 days
- Changes schedule to 23:00
- Uses development repository
- Adds development label

### How Merging Works

The strategic merge engine:
1. **Preserves base values** - Fields not mentioned in the overlay keep their base values
2. **Deep merges maps** - Nested objects are merged recursively
3. **Replaces arrays** - Arrays in the overlay completely replace base arrays
4. **Merges labels** - Labels and annotations are combined

Example:
```yaml
# Base
spec:
  storage:
    compression: Optimal
    retention:
      type: Days
      quantity: 7

# Overlay
spec:
  storage:
    retention:
      quantity: 30

# Merged Result
spec:
  storage:
    compression: Optimal  # Preserved from base
    retention:
      type: Days          # Preserved from base
      quantity: 30        # Updated from overlay
```

### Usage

```bash
# Create production job
owlctl job apply overlay-base.yaml --overlay overlay-prod.yaml

# Create development job
owlctl job apply overlay-base.yaml --overlay overlay-dev.yaml

# Plan changes before applying
owlctl job plan overlay-base.yaml --overlay overlay-prod.yaml
```

### Benefits

- **Consistency**: All environments share the same base configuration
- **Maintainability**: Update common settings in one place
- **Clarity**: Environment-specific differences are explicit
- **Safety**: Base template is version-controlled, overlays show what's different
- **Scalability**: Add new environments by creating new overlay files

---

## Repository Management

The overlay system works for all VBR resource types, including repositories.

### Export and Apply Repositories

```bash
# Export existing repository to YAML
owlctl repo export "Default Backup Repository" -o base-repo.yaml

# Apply repository configuration
owlctl repo apply base-repo.yaml

# Apply with environment overlay
owlctl repo apply base-repo.yaml -o prod-repo-overlay.yaml

# Preview changes before applying
owlctl repo apply base-repo.yaml -o prod-repo-overlay.yaml --dry-run
```

### Example: Multi-Environment Repository

**Base Repository (`base-repo.yaml`):**
```yaml
apiVersion: owlctl.veeam.com/v1
kind: VBRRepository
metadata:
  name: Primary Backup Repository
spec:
  description: Primary backup storage
  type: LinuxLocal
  maxTaskCount: 4
  makeRecentBackupsImmutableDays: 7
```

**Production Overlay (`prod-repo-overlay.yaml`):**
```yaml
apiVersion: owlctl.veeam.com/v1
kind: VBRRepository
spec:
  maxTaskCount: 8
  makeRecentBackupsImmutableDays: 14
```

**Apply:**
```bash
owlctl repo apply base-repo.yaml -o prod-repo-overlay.yaml
```

---

## Scale-Out Backup Repository (SOBR) Management

### Export and Apply SOBRs

```bash
# Export existing SOBR to YAML
owlctl repo sobr-export "Scale-out Repository 1" -o base-sobr.yaml

# Apply SOBR configuration
owlctl repo sobr-apply base-sobr.yaml

# Apply with environment overlay
owlctl repo sobr-apply base-sobr.yaml -o prod-sobr-overlay.yaml

# Preview changes before applying
owlctl repo sobr-apply base-sobr.yaml -o prod-sobr-overlay.yaml --dry-run
```

### Example: Multi-Environment SOBR

**Base SOBR (`base-sobr.yaml`):**
```yaml
apiVersion: owlctl.veeam.com/v1
kind: VBRSOBR
metadata:
  name: Scale-out Repository
spec:
  description: Scale-out backup storage
  policyType: PerformanceTier
  usePerVMBackupFiles: true
```

**Production Overlay (`prod-sobr-overlay.yaml`):**
```yaml
apiVersion: owlctl.veeam.com/v1
kind: VBRSOBR
spec:
  description: Production scale-out backup storage
  capacityTier:
    enabled: true
    daysToMove: 30
```

---

## KMS Server Management

### Export and Apply KMS Servers

```bash
# Export existing KMS server to YAML
owlctl encryption kms-export "Azure Key Vault" -o base-kms.yaml

# Apply KMS configuration
owlctl encryption kms-apply base-kms.yaml

# Apply with environment overlay
owlctl encryption kms-apply base-kms.yaml -o prod-kms-overlay.yaml

# Preview changes before applying
owlctl encryption kms-apply base-kms.yaml -o prod-kms-overlay.yaml --dry-run
```

### Example: Multi-Environment KMS

**Base KMS (`base-kms.yaml`):**
```yaml
apiVersion: owlctl.veeam.com/v1
kind: VBRKmsServer
metadata:
  name: Company Key Vault
spec:
  description: Enterprise key management
  type: AzureKeyVault
```

**Production Overlay (`prod-kms-overlay.yaml`):**
```yaml
apiVersion: owlctl.veeam.com/v1
kind: VBRKmsServer
spec:
  description: Production enterprise key management - 24x7 monitoring
```

---

## Groups, Profiles, and Targets

Groups are the recommended way to organize and deploy specs. A group bundles multiple specs with a shared profile (defaults) and overlay (policy patch).

### Example `owlctl.yaml`

```yaml
apiVersion: v1
kind: Config

groups:
  sql-tier:
    description: SQL Server backup group
    profile: profiles/gold.yaml
    overlay: overlays/compliance.yaml
    specs:
      - specs/jobs/sql-vm-01.yaml
      - specs/jobs/sql-vm-02.yaml

  web-tier:
    description: Web server backups
    profile: profiles/standard.yaml
    specs:
      - specs/jobs/web-frontend.yaml

targets:
  primary:
    url: https://vbr-prod.example.com
    description: Production VBR
  dr:
    url: https://vbr-dr.example.com
    description: DR site
```

### Group Commands

```bash
# List all groups
owlctl group list

# Show group details
owlctl group show sql-tier

# List targets
owlctl target list
```

### Apply a Group

```bash
# Dry-run
owlctl job apply --group sql-tier --dry-run

# Apply to production VBR
owlctl job apply --group sql-tier --target primary

# Apply to DR site
owlctl job apply --group sql-tier --target dr

# Drift check
owlctl job diff --group sql-tier --target primary
```

---

## Complete Multi-Target Workflow

### Group-Based Approach (Recommended for Jobs)

```bash
# 1. Export all resources from VBR
owlctl export --all -d ./specs/jobs/
owlctl repo export --all -d ./specs/repos/
owlctl repo sobr-export --all -d ./specs/sobrs/
owlctl encryption kms-export --all -d ./specs/kms/

# 2. Create profiles and overlays
# profiles/gold.yaml (kind: Profile) — retention, compression defaults
# overlays/compliance.yaml (kind: Overlay) — policy overrides

# 3. Define groups in owlctl.yaml (see example above)

# 4. Apply job groups
owlctl job apply --group sql-tier --target primary --dry-run
owlctl job apply --group sql-tier --target primary

owlctl job apply --group web-tier --target primary --dry-run
owlctl job apply --group web-tier --target primary

# 5. Apply non-job resources individually
for repo in specs/repos/*.yaml; do
  owlctl repo apply "$repo"
done

for sobr in specs/sobrs/*.yaml; do
  owlctl repo sobr-apply "$sobr"
done

for kms in specs/kms/*.yaml; do
  owlctl encryption kms-apply "$kms"
done

# 6. Verify no drift after applying
owlctl job diff --group sql-tier --target primary
owlctl job diff --group web-tier --target primary
owlctl repo diff --all
owlctl repo sobr-diff --all
owlctl encryption kms-diff --all
```

### Simpler Alternative: Single-File Overlay

For setups that don't need groups, apply individual files with `-o`:

```bash
for job in specs/jobs/*.yaml; do
  owlctl job apply "$job" -o overlays/prod/$(basename "$job")
done

owlctl job diff --all
```

---

## Next Steps

- See [Azure DevOps Pipeline Templates](pipelines/) for automated CI/CD workflows
- Read [Drift Detection Guide](../docs/drift-detection.md) for monitoring configuration changes
- Check [Security Alerting](../docs/security-alerting.md) for security-aware drift classification
