# Release Notes: v0.10.0-beta1

**Release Date:** February 2026
**Branch:** master

## Overview

This release delivers **comprehensive security drift detection** for VBR environments. Building on the declarative job management from v0.9.0, this phase adds state management for repositories, scale-out repositories, encryption passwords, and KMS servers, along with security-aware severity classification across all resource types.

## Key Features

### 1. Multi-Resource Drift Detection

Drift detection now covers the full VBR security surface:

| Resource | Snapshot | Diff | Severity |
|----------|----------|------|----------|
| Backup Jobs | `job apply` | `job diff` | CRITICAL/WARNING/INFO |
| Repositories | `repo snapshot` | `repo diff` | CRITICAL/WARNING/INFO |
| Scale-Out Repos | `repo sobr-snapshot` | `repo sobr-diff` | CRITICAL/WARNING/INFO |
| Encryption Passwords | `encryption snapshot` | `encryption diff` | WARNING/INFO |
| KMS Servers | `encryption kms-snapshot` | `encryption kms-diff` | CRITICAL/WARNING/INFO |

### 2. Security Severity Classification

Every detected drift is classified by security impact:

- **CRITICAL** - Directly weakens data protection (e.g., encryption disabled, retention reduced, job disabled, repository type changed)
- **WARNING** - Weakens defense-in-depth (e.g., schedule modified, guest processing disabled, KMS description changed)
- **INFO** - Operational change with low security impact

### 3. Value-Aware Severity (NEW in this release)

Severity classification now considers the **direction** of change, not just the field:

| Change | Severity | Reason |
|--------|----------|--------|
| Job disabled (`isDisabled: true`) | CRITICAL | Backups will stop |
| Job re-enabled (`isDisabled: false`) | WARNING | Investigate why it was disabled |
| Encryption disabled | CRITICAL | New backups will be unencrypted |
| Encryption enabled | INFO | Positive security change |
| Retention reduced (e.g., 30 -> 3 days) | CRITICAL | Clean backups will expire sooner |
| Retention increased (e.g., 3 -> 30 days) | WARNING | Review storage impact |
| Schedule/app-aware disabled | WARNING | Defense-in-depth weakened |
| Schedule/app-aware enabled | INFO | Positive change |

### 4. Cross-Resource Repository Validation (NEW in this release)

When a job's target repository changes, vcli cross-references the repository state to detect if the job was moved **off a hardened repository**:

```
CRITICAL ~ storage.backupRepositoryId: Moved from hardened repository "Default Backup Repository" to non-hardened "Temp Local Repository"
```

This requires repositories to be snapshotted in state (`vcli repo snapshot --all`).

### 5. Security Summary Header (NEW in this release)

When security-relevant drifts exist, a summary header is displayed before the drift list:

```
CRITICAL: 2 security-relevant changes detected
```

### 6. Severity Filtering

```bash
# Show only CRITICAL drifts
vcli job diff --all --severity critical

# Show WARNING and above
vcli job diff --all --security-only
```

### 7. Customizable Severity

Override default classifications via `severity-config.json`:

```json
{
  "job": { "isDisabled": "CRITICAL", "schedule": "WARNING" },
  "repository": { "type": "CRITICAL" },
  "sobr": { "isEnabled": "CRITICAL" }
}
```

### 8. CI/CD Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No drift detected |
| `3` | Drift detected (INFO or WARNING) |
| `4` | Critical security drift detected |
| `1` | Error occurred |

## New Commands

| Command | Description |
|---------|-------------|
| `vcli repo snapshot [name\|--all]` | Snapshot repository configuration to state |
| `vcli repo diff [name\|--all]` | Detect repository configuration drift |
| `vcli repo sobr-snapshot [name\|--all]` | Snapshot scale-out repository to state |
| `vcli repo sobr-diff [name\|--all]` | Detect scale-out repository drift |
| `vcli encryption snapshot [name\|--all]` | Snapshot encryption password inventory |
| `vcli encryption diff [name\|--all]` | Detect encryption password drift |
| `vcli encryption kms-snapshot [name\|--all]` | Snapshot KMS server configuration |
| `vcli encryption kms-diff [name\|--all]` | Detect KMS server drift |

All diff commands support `--severity` and `--security-only` flags.

## New Files

| File | Purpose |
|------|---------|
| `cmd/drift.go` | Shared drift detection engine, severity classification, filtering |
| `cmd/job_security.go` | Value-aware severity, cross-resource validation, security summary |
| `cmd/repo.go` | Repository and SOBR state management and drift detection |
| `cmd/encryption.go` | Encryption password and KMS server state management |
| `cmd/severity_config.go` | Customizable severity overrides via JSON config |
| `state/manager.go` | State file management (load, save, atomic writes) |
| `state/models.go` | State data model (resources, specs) |
| `state/lock.go` | State file locking |
| `docs/drift-detection.md` | Comprehensive drift detection guide |
| `docs/security-alerting.md` | Security alerting reference |

## Testing

Tested against live VBR v1.3-rev1 environment:

- Job drift detection with value-aware severity
- Retention reduction correctly classified as CRITICAL
- Job disable correctly classified as CRITICAL
- Schedule enable correctly classified as INFO (downgraded from WARNING)
- `--severity critical` filter correctly excludes INFO drifts
- `--security-only` filter correctly shows WARNING+ drifts
- Security summary header displays correct counts
- Repository, SOBR, encryption, and KMS drift detection verified
- No regressions in existing commands

## Breaking Changes

None. All existing vcli commands continue to work unchanged.

## Documentation

New:
- `docs/drift-detection.md` - Full drift detection guide covering all resource types
- `docs/security-alerting.md` - Value-aware severity reference and CI/CD integration

Updated:
- `README.md` - Updated commands reference, drift detection section, and changelog

## What's Next

Potential future enhancements:
- Correlation engine to detect multi-resource attack patterns
- Webhook/SIEM integration for pushing alerts
- RBAC and credential inventory tracking
- Malware detection and compliance analyzer monitoring
