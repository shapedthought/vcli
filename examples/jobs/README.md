# Standalone Job Configuration Examples

These are **full, self-contained** job specs that can be applied directly without a profile or group. Every field is specified in the file itself.

For the recommended groups workflow with thin specs and profiles, see the [examples README](../README.md) and the [`specs/`](../specs/) directory.

## Available Examples

### database-backup.yaml

Production database server backup with:
- Application-aware processing (VSS + SQL transaction logs)
- 30-day retention
- Nightly schedule at 02:00 with retry logic
- Email notifications

**Use case:** Database servers requiring transaction log handling and longer retention.

### web-tier-backup.yaml

Web server tier backup with:
- Image-level backup only (no guest processing)
- 14-day retention
- Multiple web servers
- Simplified processing for stateless servers

**Use case:** Stateless web servers that don't require application-aware processing.

## Usage

```bash
# Apply directly (standalone — no profile or group needed)
owlctl job apply database-backup.yaml

# Preview changes first
owlctl job apply database-backup.yaml --dry-run

# Apply with a policy overlay
owlctl job apply database-backup.yaml -o ../overlays/retention-30d.yaml

# Preview merged result
owlctl job plan database-backup.yaml -o ../overlays/enable-encryption.yaml --show-yaml
```

## Exporting Existing Jobs

```bash
# Export by job ID
owlctl export 57b3baab-6237-41bf-add7-db63d41d984c -o my-job.yaml

# Export all jobs
owlctl export --all -d ./all-jobs/
```

## Required Fields

All job configurations must include:
- `apiVersion: owlctl.veeam.com/v1`
- `kind: VBRJob`
- `metadata.name` — unique job name
- `spec.type` — job type (`VSphereBackup`, `HyperVBackup`, etc.)
- `spec.repository` — target repository name
- `spec.storage.retention` — retention policy
- `spec.objects` — VMs or objects to back up

## See Also

- [Groups Workflow](../README.md) — recommended approach using profiles and thin specs
- [Overlay Examples](../overlays/) — policy-focused overlays
- [Drift Detection Guide](../../docs/drift-detection.md)
