# Scale-Out Repository (SOBR) Configuration Examples

This directory contains example VBR Scale-Out Backup Repository configurations demonstrating capacity tier and archive tier configurations.

## Available Examples

### scale-out-repository.yaml
Production SOBR with capacity tier:
- Performance tier with per-VM backup files
- S3 capacity tier for cost optimization
- 30-day move policy
- 7-day operational restore window
- Encryption enabled for cloud storage

**Use case:** Multi-tier backup storage with automatic data movement to cloud for long-term retention.

## Important Notes

**⚠️ SOBRs are Update-Only**

SOBRs **cannot be created** via owlctl apply. They must be created in the VBR console first, including:
- SOBR creation
- Adding extent repositories
- Basic capacity/archive tier setup

owlctl can **update** SOBR settings:
- Capacity tier move policies
- Operational restore windows
- Encryption settings
- Archive tier configuration
- Data placement policies

**Workflow:**
1. Create SOBR in VBR console
2. Add extent repositories
3. Configure basic tier settings
4. Export to YAML: `owlctl repo sobr-export "SOBR Name" -o sobr.yaml`
5. Edit YAML for desired settings
6. Apply updates: `owlctl repo sobr-apply sobr.yaml`

## Usage

### Snapshot Existing SOBR

```bash
# Snapshot single SOBR
owlctl repo sobr-snapshot "Scale-out Repository 1"

# Snapshot all SOBRs
owlctl repo sobr-snapshot --all
```

### Export SOBR Configuration

```bash
# Export by name
owlctl repo sobr-export "Scale-out Repository 1" -o sobr.yaml

# Export all SOBRs
owlctl repo sobr-export --all -d ./all-sobrs/
```

### Apply SOBR Updates

```bash
# Preview changes first (recommended)
owlctl repo sobr-apply scale-out-repository.yaml --dry-run

# Apply updates
owlctl repo sobr-apply scale-out-repository.yaml

# Apply with environment overlay
owlctl repo sobr-apply scale-out-repository.yaml -o ../overlays/prod/sobr-overlay.yaml
```

### Detect Configuration Drift

```bash
# Check single SOBR
owlctl repo sobr-diff "Scale-out Repository 1"

# Check all SOBRs
owlctl repo sobr-diff --all

# Check for security-relevant drift only
owlctl repo sobr-diff --all --security-only
```

## Customization

### Modifying Examples for Your Environment

1. **Update SOBR name** to match your VBR configuration
2. **Configure capacity tier** move policies based on requirements
3. **Set operational restore window** according to RTO objectives
4. **Enable/configure archive tier** for compliance retention
5. **Adjust encryption settings** based on security requirements

### Updateable Fields

Via owlctl apply, you can update:
- Policy type (PerformanceTier vs DataLocality)
- Capacity tier settings:
  - Days to move
  - Operational restore window
  - Copy vs Move mode
  - Encryption settings
- Archive tier settings:
  - Days to archive
  - Cost-optimized archive mode
- Per-VM backup file settings

### Non-Updateable Fields

These must be managed in VBR console:
- SOBR creation
- Extent repository assignments
- Extent repository order
- Cloud/object storage credentials

## Capacity Tier Strategies

### Move Policy
Removes backups from performance tier after moving to cloud:
- **Pros:** Frees performance tier space immediately
- **Cons:** Restore from cloud may be slower
- **Use case:** Space-constrained performance tier

```yaml
capacityTier:
  daysToMove: 30
  copyMode: Move
  operationalRestoreWindowDays: 7
```

### Copy Policy
Keeps copy on performance tier after moving to cloud:
- **Pros:** Faster restores from performance tier
- **Cons:** Requires more performance tier space
- **Use case:** Fast restore requirements

```yaml
capacityTier:
  daysToMove: 30
  copyMode: Copy
  operationalRestoreWindowDays: 30
```

## Best Practices

1. **Set appropriate move policies** based on restore requirements
2. **Enable encryption** for data moved to cloud
3. **Configure operational restore window** to balance cost and RTO
4. **Use archive tier** for compliance/long-term retention
5. **Regular drift detection** to catch unauthorized changes
6. **Test restores** from capacity and archive tiers
7. **Monitor cloud storage costs** and adjust policies accordingly
8. **Version control** SOBR configurations in Git

## Common Configurations

### Cost-Optimized Setup
```yaml
capacityTier:
  enabled: true
  daysToMove: 14
  copyMode: Move
  operationalRestoreWindowDays: 3
```

### Performance-Optimized Setup
```yaml
capacityTier:
  enabled: true
  daysToMove: 60
  copyMode: Copy
  operationalRestoreWindowDays: 30
```

### Compliance Setup with Archive Tier
```yaml
capacityTier:
  enabled: true
  daysToMove: 30
  copyMode: Move

archiveTier:
  enabled: true
  daysToArchive: 365
  costOptimizedArchiveMode: true
```

## Exit Codes

SOBR apply commands return specific exit codes:

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Continue |
| 1 | Error (API failure) | Check logs and retry |
| 6 | SOBR not found | Create in VBR console first |

## See Also

- [Declarative Mode Guide](../../docs/declarative-mode.md) - Complete SOBR management guide
- [State Management Guide](../../docs/state-management.md) - Understanding state and snapshots
- [Drift Detection Guide](../../docs/drift-detection.md) - Configuration monitoring
- [Security Alerting](../../docs/security-alerting.md) - Security-aware drift classification
