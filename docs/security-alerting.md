# Security Alerting

owlctl provides enhanced security alerting for configuration drift that goes beyond simple field-level comparison. It uses **value-aware severity classification** and **cross-resource validation** to assess the actual security impact of each change.

## Value-Aware Severity

Standard drift detection classifies severity based on which field changed. Value-aware severity also considers the **direction** of the change.

For example, `isDisabled` changing to `true` (job disabled) is CRITICAL, but changing to `false` (job re-enabled) is only WARNING.

### Value-Aware Rules

| Field | VBR Value | Severity | Rationale |
|-------|-----------|----------|-----------|
| `isDisabled` | `true` | CRITICAL | Job was disabled - backups will stop |
| `isDisabled` | `false` | WARNING | Job was re-enabled - investigate why it was disabled |
| `storage.advancedSettings.storageData.encryption.isEnabled` | `false` | CRITICAL | Encryption was disabled - new backups will be unencrypted |
| `storage.advancedSettings.storageData.encryption.isEnabled` | `true` | INFO | Encryption was enabled - positive security change |
| `storage.retentionPolicy.quantity` | Reduced | CRITICAL | Retention was shortened - clean backups will expire sooner |
| `storage.retentionPolicy.quantity` | Increased | WARNING | Retention was extended - review storage impact |
| `guestProcessing.appAwareProcessing.isEnabled` | `false` | WARNING | App-aware processing disabled - databases may not be recoverable |
| `guestProcessing.appAwareProcessing.isEnabled` | `true` | INFO | App-aware processing enabled - positive change |
| `schedule.daily.isEnabled` | `false` | WARNING | Daily schedule disabled |
| `schedule.daily.isEnabled` | `true` | INFO | Daily schedule enabled |
| `schedule.runAutomatically` | `false` | WARNING | Automatic scheduling disabled |
| `schedule.runAutomatically` | `true` | INFO | Automatic scheduling enabled |

## Cross-Resource Repository Validation

When a job's `storage.backupRepositoryId` changes, owlctl cross-references the repository state to determine if the job was moved off a **hardened repository** (Linux Hardened type).

### What Gets Detected

If the old repository was `LinuxHardened` and the new repository is not, owlctl:

1. Ensures the drift is classified as CRITICAL
2. Adds a descriptive synthetic drift entry:

```
CRITICAL ~ storage.backupRepositoryId: Moved from hardened repository "Default Backup Repository" to non-hardened "Temp Local Repository"
```

### Requirements

For cross-resource validation to work, the repositories must be snapshotted in state:

```bash
# Snapshot all repositories first
owlctl repo snapshot --all

# Then job diffs will cross-reference repository data
owlctl job diff --all
```

If the repositories are not in state, the standard severity classification still applies.

## Security Summary Header

When security-relevant drifts (WARNING or higher) are detected, a summary header is printed before the drift list:

```
CRITICAL: 2 security-relevant changes detected
```

or:

```
WARNING: 3 security-relevant changes detected
```

The severity level of the header reflects the highest severity drift found. The count includes all WARNING and CRITICAL drifts.

## Full Severity Reference

### Job Fields

| Field Path | Default Severity |
|-----------|-----------------|
| `isDisabled` | CRITICAL (value-aware) |
| `retentionPolicy` | CRITICAL |
| `retainCycles` | CRITICAL |
| `gfsPolicy` | CRITICAL |
| `backupRepositoryId` | CRITICAL |
| `storage.advancedSettings.storageData.encryption.isEnabled` | CRITICAL (value-aware) |
| `storage.retentionPolicy.quantity` | CRITICAL (value-aware) |
| `storage.backupRepositoryId` | CRITICAL (cross-resource) |
| `guestProcessing` | WARNING |
| `schedule` | WARNING |
| `encryption` | WARNING |
| `healthCheck` | WARNING |
| `storage.retentionPolicy.type` | WARNING |
| `guestProcessing.appAwareProcessing.isEnabled` | WARNING (value-aware) |
| `schedule.daily.isEnabled` | WARNING (value-aware) |
| `schedule.runAutomatically` | WARNING (value-aware) |
| All other fields | INFO |

### Repository Fields

| Field Path | Default Severity |
|-----------|-----------------|
| `type` | CRITICAL |
| `repository.makeRecentBackupsImmutableDays` | CRITICAL |
| `repository.advancedSettings.decompressBeforeStoring` | CRITICAL |
| `repository.advancedSettings.perVmBackup` | CRITICAL |
| `path` | WARNING |
| `maxTaskCount` | WARNING |
| `repository.advancedSettings.alignDataBlocks` | WARNING |
| `repository.readWriteLimitEnabled` | WARNING |
| `repository.readWriteRate` | WARNING |
| `repository.taskLimitEnabled` | WARNING |
| All other fields | INFO |

### Scale-Out Repository (SOBR) Fields

| Field Path | Default Severity |
|-----------|-----------------|
| `isEnabled` | CRITICAL |
| `immutabilityMode` | CRITICAL |
| `type` | CRITICAL |
| `enforceStrictPlacementPolicy` | CRITICAL |
| `capacityTier.encryption` | CRITICAL |
| `capacityTier.encryption.isEnabled` | CRITICAL |
| `movePolicyEnabled` | WARNING |
| `copyPolicyEnabled` | WARNING |
| `daysCount` | WARNING |
| `performanceExtents` | WARNING |
| `extents` | WARNING |
| `capacityTier.backupHealth` | WARNING |
| `capacityTier.backupHealth.isEnabled` | WARNING |
| All other fields | INFO |

### Encryption Password Fields

| Field Path | Default Severity |
|-----------|-----------------|
| `hint` | WARNING |
| All other fields | INFO |

### KMS Server Fields

| Field Path | Default Severity |
|-----------|-----------------|
| `type` | CRITICAL |
| `name` | WARNING |
| `description` | WARNING |
| All other fields | INFO |

## Example: Complete Security Workflow

```bash
# 1. Snapshot all resources to establish baseline
owlctl repo snapshot --all
owlctl repo sobr-snapshot --all
owlctl encryption snapshot --all
owlctl encryption kms-snapshot --all

# 2. Apply job configurations (creates job state)
owlctl job apply prod-backup.yaml

# 3. Periodic security check (run via cron or CI/CD)
owlctl job diff --all --security-only
owlctl repo diff --all --security-only
owlctl repo sobr-diff --all --security-only
owlctl encryption diff --all --security-only

# 4. Full audit check
owlctl job diff --all
owlctl repo diff --all
owlctl repo sobr-diff --all
owlctl encryption diff --all
owlctl encryption kms-diff --all
```

### Automated Security Gate

```bash
#!/bin/bash
# security-check.sh - Run as scheduled CI/CD job

CRITICAL=0

for CMD in \
    "owlctl job diff --all --severity critical" \
    "owlctl repo diff --all --severity critical" \
    "owlctl repo sobr-diff --all --severity critical" \
    "owlctl encryption diff --all --severity critical"
do
    eval $CMD
    if [ $? -eq 4 ]; then
        CRITICAL=1
    fi
done

if [ $CRITICAL -eq 1 ]; then
    echo "CRITICAL security drift detected across VBR environment"
    # Alert security team, fail pipeline, etc.
    exit 1
fi

echo "No critical security drift detected"
exit 0
```
