# Job Configuration Examples

This directory contains example VBR backup job configurations demonstrating common use cases and best practices.

## Available Examples

### database-backup.yaml
Production database server backup with:
- Application-aware processing (SQL transaction logs)
- 30-day retention policy
- Encryption optional
- Nightly schedule with retry logic
- Email notifications

**Use case:** Database servers requiring transaction log handling and longer retention.

### web-tier-backup.yaml
Web server tier backup with:
- Image-level backup only (no app-aware processing)
- 14-day retention policy
- Multiple web servers
- Simplified processing for stateless servers

**Use case:** Stateless web servers that don't require application-aware processing.

## Usage

### Apply a Job Configuration

```bash
# Apply directly
vcli job apply database-backup.yaml

# Preview changes first (dry-run)
vcli job apply database-backup.yaml --dry-run

# Apply with environment overlay
vcli job apply database-backup.yaml -o ../overlays/prod/database-backup-overlay.yaml
```

### Export Existing Job

```bash
# Export by job ID
vcli export 57b3baab-6237-41bf-add7-db63d41d984c -o my-job.yaml

# Export by job name
vcli export "My Backup Job" -o my-job.yaml

# Export all jobs
vcli export --all -d ./all-jobs/
```

### Create Environment-Specific Jobs

```bash
# Production
vcli job apply database-backup.yaml -o ../overlays/prod/database-backup-overlay.yaml

# Development
vcli job apply database-backup.yaml -o ../overlays/dev/database-backup-overlay.yaml

# Preview merged configuration
vcli job plan database-backup.yaml -o ../overlays/prod/database-backup-overlay.yaml --show-yaml
```

## Customization

### Modifying Examples for Your Environment

1. **Update VM names and hosts** in the `objects` section
2. **Set repository name** to match your VBR repositories
3. **Adjust retention** based on your requirements
4. **Configure encryption** if needed
5. **Update notification recipients**
6. **Modify schedule** to fit your backup windows

### Required Fields

All job configurations must include:
- `apiVersion: vcli.veeam.com/v1`
- `kind: VBRJob`
- `metadata.name` - Unique job name
- `spec.type` - Job type (VSphereBackup, HyperVBackup, etc.)
- `spec.repository` - Target repository name
- `spec.storage.retention` - Retention policy
- `spec.objects` - VMs or objects to backup

## Best Practices

1. **Use overlays** for environment-specific settings
2. **Version control** all job configurations in Git
3. **Test with --dry-run** before applying
4. **Use meaningful labels** for organization
5. **Document customizations** in YAML comments
6. **Export existing jobs** as templates for new jobs
7. **Regular drift detection** to catch unauthorized changes

## See Also

- [Declarative Mode Guide](../../docs/declarative-mode.md) - Complete guide to job management
- [Overlay Examples](../overlays/) - Environment-specific overlays
- [Command Reference](../../docs/command-reference.md) - Quick command lookup
