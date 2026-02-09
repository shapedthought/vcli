# Configuration Overlays

This directory contains environment-specific overlays that extend base configurations for production, development, and other environments.

## Directory Structure

```
overlays/
├── prod/               # Production overlays
│   ├── database-backup-overlay.yaml
│   └── backup-repository-overlay.yaml
└── dev/                # Development overlays
    └── database-backup-overlay.yaml
```

## What are Overlays?

Overlays are YAML files containing only the fields you want to override from a base configuration. They enable:
- **DRY principle** - Define common settings once in base
- **Environment variations** - Customize for prod/dev/staging
- **Maintainability** - Update common settings in one place
- **Clarity** - Explicit differences between environments

## How Overlays Work

owlctl uses **strategic merge** to combine base + overlay:

```yaml
# Base configuration
storage:
  compression: Optimal
  retention:
    type: Days
    quantity: 7

# Overlay
storage:
  retention:
    quantity: 30

# Merged result
storage:
  compression: Optimal    # Preserved from base
  retention:
    type: Days            # Preserved from base
    quantity: 30          # Overridden by overlay
```

## Merge Behavior

### Scalar Values
Overlay replaces base:
```yaml
Base:    quantity: 7
Overlay: quantity: 30
Result:  quantity: 30
```

### Nested Objects (Maps)
Deep recursive merge:
```yaml
Base:
  storage:
    compression: Optimal
    retention:
      type: Days
      quantity: 7

Overlay:
  storage:
    retention:
      quantity: 30

Result:
  storage:
    compression: Optimal    # From base
    retention:
      type: Days           # From base
      quantity: 30          # From overlay
```

### Arrays
**Complete replacement** (not merged):
```yaml
Base:
  objects:
    - name: vm1
    - name: vm2

Overlay:
  objects:
    - name: vm3

Result:
  objects:
    - name: vm3    # Base array completely replaced
```

**Important:** When overriding arrays, include all desired items.

### Labels
Labels are combined (merged):
```yaml
Base:
  labels:
    app: database
    managed-by: owlctl

Overlay:
  labels:
    environment: production

Result:
  labels:
    app: database            # From base
    managed-by: owlctl        # From base
    environment: production  # From overlay
```

## Usage

### Apply with Overlay

```bash
# Explicit overlay file
owlctl job apply ../jobs/database-backup.yaml -o prod/database-backup-overlay.yaml

# Preview merged configuration
owlctl job plan ../jobs/database-backup.yaml -o prod/database-backup-overlay.yaml --show-yaml

# Dry-run before applying
owlctl job apply ../jobs/database-backup.yaml -o prod/database-backup-overlay.yaml --dry-run
```

### Groups and Overlays

Overlays work both standalone (via the `-o` flag) and as part of a group. When using groups, the overlay is automatically applied to every spec in the group.

Define groups in `owlctl.yaml`:

```yaml
# owlctl.yaml (in project root)
apiVersion: owlctl.veeam.com/v1
kind: Config

groups:
  database-prod:
    description: Production database backups
    profile: profiles/gold.yaml           # kind: Profile — base defaults
    overlay: prod/database-backup-overlay.yaml  # kind: Overlay — policy patch
    specs:
      - ../jobs/database-backup.yaml
```

Then use `--group`:
```bash
# Apply the entire group (profile + spec + overlay merged)
owlctl job apply --group database-prod

# Preview what would be applied
owlctl job apply --group database-prod --dry-run

# Drift check
owlctl job diff --group database-prod
```

For simpler cases without groups, use `-o` directly:
```bash
owlctl job apply database-backup.yaml -o prod/database-backup-overlay.yaml
```

See [Declarative Mode Guide — Groups](../../docs/declarative-mode.md#groups) for the full groups reference.

## Available Overlays

### Production (`prod/`)

Production overlays typically include:
- Extended retention periods
- Encryption enabled
- Production repositories
- Enhanced monitoring/notifications
- Higher task limits
- Earlier backup schedules
- Production-specific labels

### Development (`dev/`)

Development overlays typically include:
- Shorter retention periods
- Encryption disabled (optional)
- Development repositories
- Reduced notifications
- Lower task limits
- Later backup schedules
- Development-specific labels

## Creating New Overlays

### 1. Start with Base Configuration

```bash
# Export existing resource as base
owlctl export <job-id> -o base-job.yaml
```

### 2. Create Minimal Overlay

Only include fields that differ from base:

```yaml
# staging-overlay.yaml
apiVersion: owlctl.veeam.com/v1
kind: VBRJob
metadata:
  name: Database Backup
  labels:
    environment: staging
spec:
  repository: Staging Repository
  storage:
    retention:
      quantity: 14  # Different from base
  schedule:
    daily: "03:00"  # Different from base
```

### 3. Test the Merge

```bash
# Preview merged result
owlctl job plan base-job.yaml -o staging-overlay.yaml --show-yaml
```

### 4. Apply

```bash
# Dry-run first
owlctl job apply base-job.yaml -o staging-overlay.yaml --dry-run

# Apply
owlctl job apply base-job.yaml -o staging-overlay.yaml
```

## Best Practices

### 1. Keep Overlays Small
Only override what's different:
```yaml
# Good - minimal overlay
spec:
  repository: Production Repository
  storage:
    retention:
      quantity: 60

# Bad - duplicates base unnecessarily
spec:
  type: VSphereBackup          # Same as base, not needed
  repository: Production Repository
  storage:
    compression: Optimal        # Same as base, not needed
    retention:
      type: Days               # Same as base, not needed
      quantity: 60             # Different - keep this
```

### 2. Use Consistent Naming

```
base-database-backup.yaml
prod-database-backup-overlay.yaml
dev-database-backup-overlay.yaml
staging-database-backup-overlay.yaml
```

### 3. Document Overrides

```yaml
spec:
  storage:
    retention:
      quantity: 60  # Extended for compliance requirements
```

### 4. Version Control Everything

```bash
git add base-*.yaml overlays/
git commit -m "Add staging environment overlay"
git push
```

### 5. Test with --show-yaml

Always preview the full merged result:
```bash
owlctl job plan base.yaml -o overlay.yaml --show-yaml
```

### 6. Watch for Array Replacement

Remember arrays are replaced, not merged:
```yaml
# Base has 2 VMs, overlay adds 1 more
# Wrong approach - this replaces entirely:
objects:
  - name: new-vm-03

# Right approach - include all VMs:
objects:
  - name: vm-01  # From base
  - name: vm-02  # From base
  - name: vm-03  # New one
```

### 7. Use Labels Effectively

Labels are merged, so use them for classification:
```yaml
# Base
labels:
  app: database
  team: infrastructure
  managed-by: owlctl

# Overlay
labels:
  environment: production  # Adds without removing base labels
  sla: tier-1
```

## Common Patterns

### Repository Override

```yaml
# Override repository for environment
spec:
  repository: Production Backup Repository
```

### Retention Extension

```yaml
# Extend retention for compliance
spec:
  storage:
    retention:
      quantity: 365  # 1 year for production
```

### Schedule Adjustment

```yaml
# Earlier schedule for production
spec:
  schedule:
    daily: "01:00"
```

### VM List Override

```yaml
# Different VMs per environment
spec:
  objects:
    - type: VirtualMachine
      name: app-server-01-prod
      hostName: esxi-prod-01.company.local
```

### Encryption Toggle

```yaml
# Enable encryption for production
spec:
  storage:
    encryption:
      enabled: true
      encryptionKey: "Production Encryption Key"
```

## Troubleshooting

### Overlay Not Applied

Check resolution priority and file paths:
```bash
# Verify file exists
ls -la overlays/prod/overlay.yaml

# Use absolute path
owlctl job apply base.yaml -o $PWD/overlays/prod/overlay.yaml

# Check owlctl.yaml if using --env
cat owlctl.yaml
```

### Unexpected Merge Results

Use `--show-yaml` to see actual result:
```bash
owlctl job plan base.yaml -o overlay.yaml --show-yaml
```

### Array Not Merged as Expected

Arrays are replaced, not merged. Include all items:
```yaml
# Overlay must include complete array
objects:
  - name: vm1  # Must repeat from base
  - name: vm2  # Must repeat from base
  - name: vm3  # New item
```

## See Also

- [Declarative Mode Guide](../../docs/declarative-mode.md) - Complete overlay documentation
- [Job Examples](../jobs/) - Base job configurations
- [Repository Examples](../repos/) - Base repository configurations
- [Command Reference](../../docs/command-reference.md) - Apply and plan commands
