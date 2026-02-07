# Repository Configuration Examples

This directory contains example VBR backup repository configurations demonstrating best practices for repository management.

## Available Examples

### backup-repository.yaml
Linux-based backup repository with:
- Immutability enabled (14-day protection)
- Concurrent task limit (8)
- Deduplication and compression
- Per-VM backup files
- Parallel processing enabled

**Use case:** Primary backup repository with ransomware protection via immutability.

## Important Notes

**⚠️ Repositories are Update-Only**

Repositories **cannot be created** via owlctl apply. They must be created in the VBR console first.

owlctl can only **update** existing repositories:
- Task limits
- Immutability settings
- Description
- Advanced settings

**Workflow:**
1. Create repository in VBR console
2. Configure basic settings (path, type, credentials)
3. Export to YAML: `owlctl repo export "Repository Name" -o repo.yaml`
4. Edit YAML for desired settings
5. Apply updates: `owlctl repo apply repo.yaml`

## Usage

### Snapshot Existing Repository

```bash
# Snapshot single repository
owlctl repo snapshot "Default Backup Repository"

# Snapshot all repositories
owlctl repo snapshot --all
```

### Export Repository Configuration

```bash
# Export by name
owlctl repo export "Default Backup Repository" -o repo.yaml

# Export all repositories
owlctl repo export --all -d ./all-repos/
```

### Apply Repository Updates

```bash
# Preview changes first (recommended)
owlctl repo apply backup-repository.yaml --dry-run

# Apply updates
owlctl repo apply backup-repository.yaml

# Apply with environment overlay
owlctl repo apply backup-repository.yaml -o ../overlays/prod/backup-repository-overlay.yaml
```

### Detect Configuration Drift

```bash
# Check single repository
owlctl repo diff "Default Backup Repository"

# Check all repositories
owlctl repo diff --all

# Check for security-relevant drift only
owlctl repo diff --all --security-only
```

## Customization

### Modifying Examples for Your Environment

1. **Update repository name** to match your VBR configuration
2. **Adjust task limits** based on workload and hardware
3. **Configure immutability** according to compliance requirements
4. **Set appropriate labels** for organization

### Updateable Fields

Via owlctl apply, you can update:
- Description
- Max concurrent task count
- Immutability period (days)
- Deduplication cache size
- Per-VM backup file settings
- Parallel processing settings

### Non-Updateable Fields

These must be set in VBR console:
- Repository type (Linux, Windows, etc.)
- Storage path
- Credentials
- Server/host

## Best Practices

1. **Enable immutability** for ransomware protection
2. **Set appropriate task limits** based on hardware capacity
3. **Use overlays** for environment-specific settings
4. **Regular drift detection** to catch unauthorized changes
5. **Snapshot before changes** to track configuration history
6. **Version control** repository configurations in Git
7. **Test updates with --dry-run** first

## Exit Codes

Repository apply commands return specific exit codes:

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Continue |
| 1 | Error (API failure) | Check logs and retry |
| 6 | Repository not found | Create in VBR console first |

## See Also

- [Declarative Mode Guide](../../docs/declarative-mode.md) - Complete repository management guide
- [State Management Guide](../../docs/state-management.md) - Understanding state and snapshots
- [Drift Detection Guide](../../docs/drift-detection.md) - Configuration monitoring
- [Security Alerting](../../docs/security-alerting.md) - Security-aware drift classification
