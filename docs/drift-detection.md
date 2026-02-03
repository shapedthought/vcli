# Drift Detection Guide

vcli provides drift detection across multiple VBR resource types. It compares the current live VBR configuration against a saved state snapshot to identify manual changes or unauthorized modifications.

## Supported Resource Types

| Resource Type | Snapshot Command | Diff Command |
|---------------|-----------------|--------------|
| Backup Jobs | `vcli job apply` | `vcli job diff` |
| Repositories | `vcli repo snapshot` | `vcli repo diff` |
| Scale-Out Repositories | `vcli repo sobr-snapshot` | `vcli repo sobr-diff` |
| Encryption Passwords | `vcli encryption snapshot` | `vcli encryption diff` |
| KMS Servers | `vcli encryption kms-snapshot` | `vcli encryption kms-diff` |

## How It Works

1. **Snapshot**: Capture the current VBR configuration into state (`state.json`)
2. **Drift Detection**: Compare the saved state against the live VBR configuration
3. **Classification**: Each drift is assigned a severity level (CRITICAL, WARNING, or INFO)
4. **Filtering**: Optionally filter output to show only security-relevant changes

### State File

State is stored in `state.json` at either `$VCLI_SETTINGS_PATH` or `~/.vcli/`.

```json
{
  "version": 1,
  "resources": {
    "Backup Job 1": {
      "type": "VBRJob",
      "id": "c07c7ea3-...",
      "name": "Backup Job 1",
      "lastApplied": "2026-02-01T14:30:00Z",
      "lastAppliedBy": "edwardhoward",
      "spec": { }
    }
  }
}
```

> **Note**: State files are operational tools for drift detection, not compliance-grade audit logs. For compliance, use Git commit history + CI/CD logs + VBR audit logs.

## Job Drift Detection

### Single Job

```bash
vcli job diff "Backup Job 1"
```

### All Jobs

```bash
vcli job diff --all
```

### Example Output

```
Checking drift for job: Backup Job 1

CRITICAL: 2 security-relevant changes detected

Drift detected:
  CRITICAL ~ isDisabled: false (state) -> true (VBR)
  CRITICAL ~ storage.retentionPolicy.quantity: 3 (state) -> 1 (VBR)
  INFO ~ schedule.runAutomatically: false (state) -> true (VBR)

Summary:
  - 3 drifts detected
  - Highest severity: CRITICAL
  - Last applied: 2026-02-01 17:17:25
  - Last applied by: edwardhoward
```

## Repository Drift Detection

### Snapshot Repositories

```bash
# Single repository
vcli repo snapshot "Default Backup Repository"

# All repositories
vcli repo snapshot --all
```

### Detect Drift

```bash
# Single repository
vcli repo diff "Default Backup Repository"

# All repositories
vcli repo diff --all
```

## Scale-Out Backup Repository (SOBR) Drift Detection

```bash
# Snapshot
vcli repo sobr-snapshot --all

# Detect drift
vcli repo sobr-diff --all
```

## Encryption Password Drift Detection

```bash
# Snapshot
vcli encryption snapshot --all

# Detect drift
vcli encryption diff --all
```

## KMS Server Drift Detection

```bash
# Snapshot
vcli encryption kms-snapshot --all

# Detect drift
vcli encryption kms-diff --all
```

## Severity Classification

Every drift is classified by security impact:

| Level | Meaning | Action |
|-------|---------|--------|
| **CRITICAL** | Directly weakens data protection or ransomware resilience | Immediate investigation |
| **WARNING** | Weakens defense-in-depth or monitoring | Investigate within 24 hours |
| **INFO** | Operational change with low direct security impact | Review during next audit |

See [Security Alerting](security-alerting.md) for the full severity reference and value-aware classification details.

## Severity Filtering

### Show Only Critical Drifts

```bash
vcli job diff --all --severity critical
```

### Show Security-Relevant Drifts (WARNING and above)

```bash
vcli job diff --all --security-only
```

These flags work on all diff commands (`job diff`, `repo diff`, `repo sobr-diff`, `encryption diff`, `encryption kms-diff`).

## Custom Severity Configuration

Override default severity levels by placing a `severity-config.json` in `$VCLI_SETTINGS_PATH` or `~/.vcli/`:

```json
{
  "job": {
    "isDisabled": "CRITICAL",
    "schedule": "WARNING"
  },
  "repository": {
    "type": "CRITICAL"
  },
  "sobr": {
    "isEnabled": "CRITICAL"
  },
  "encryption": {
    "hint": "WARNING"
  },
  "kms": {
    "type": "CRITICAL"
  }
}
```

Both short field names (`isDisabled`) and full dotted paths (`storage.retentionPolicy.quantity`) are supported.

## Exit Codes

All commands return structured exit codes for CI/CD integration:

### Diff Commands

| Code | Meaning |
|------|---------|
| `0` | No drift detected |
| `3` | Drift detected (INFO or WARNING severity) |
| `4` | Critical drift detected |
| `1` | Error occurred |

### Apply Commands

| Code | Meaning |
|------|---------|
| `0` | Applied successfully |
| `1` | Error occurred (API failure, invalid spec) |
| `5` | Partial apply — some resources succeeded, some failed (batch operations) |
| `6` | Resource not found — cannot apply (update-only resources like repos, SOBRs, KMS) |

### CI/CD Example (Diff)

```bash
#!/bin/bash
vcli job diff --all --security-only
EXIT_CODE=$?

if [ $EXIT_CODE -eq 4 ]; then
    echo "CRITICAL security drift detected!"
    # Send alert to security team
    curl -X POST "$SLACK_WEBHOOK" \
      -d '{"text":"CRITICAL: VBR security drift detected. Immediate investigation required."}'
    exit 1
elif [ $EXIT_CODE -eq 3 ]; then
    echo "WARNING: Security drift detected, review required."
fi
```

### CI/CD Example (Apply)

```bash
#!/bin/bash
# Single resource apply
vcli repo apply repos/default-repo.yaml
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "Repository applied successfully"
elif [ $EXIT_CODE -eq 6 ]; then
    echo "Resource not found - must be created in VBR console first"
    exit 1
else
    echo "Apply failed with error"
    exit 1
fi
```

```bash
#!/bin/bash
# Batch apply (multiple files) - exit code 5 possible
for spec in repos/*.yaml; do
    vcli repo apply "$spec"
    RESULT=$?
    if [ $RESULT -ne 0 ]; then
        FAILED=$((FAILED + 1))
    else
        SUCCESS=$((SUCCESS + 1))
    fi
done

if [ $FAILED -gt 0 ] && [ $SUCCESS -gt 0 ]; then
    echo "Partial success: $SUCCESS applied, $FAILED failed"
    exit 5  # Partial apply
elif [ $FAILED -gt 0 ]; then
    echo "All applies failed"
    exit 1
fi
```

## Dry-Run Mode

All apply commands support `--dry-run` to preview changes without making any modifications:

```bash
# Preview repository changes
vcli repo apply repos/default-repo.yaml --dry-run

# Preview SOBR changes
vcli repo sobr-apply sobrs/sobr1.yaml --dry-run

# Preview KMS server changes
vcli encryption kms-apply kms/my-kms.yaml --dry-run
```

### Example Output

```
=== Dry Run Mode ===
Resource: Default Backup Repository (VBRRepository)
Action: Would UPDATE existing resource

Changes that would be applied:
  ~ description: "Created by Veeam Backup" -> "Updated description"
  ~ repository.maxTaskCount: 4 -> 8
  ~ repository.makeRecentBackupsImmutableDays: 7 -> 14

3 field(s) would be changed.

=== End Dry Run ===
No changes made. Remove --dry-run flag to apply.
```

### CI/CD Validation Stage

Use dry-run for safe PR validation before applying changes:

```yaml
# Azure DevOps Pipeline Example
stages:
  - stage: Validate
    jobs:
      - job: DryRun
        steps:
          - script: |
              for spec in repos/*.yaml; do
                ./vcli repo apply "$spec" --dry-run
                if [ $? -ne 0 ]; then
                  echo "Validation failed for $spec"
                  exit 1
                fi
              done
            displayName: 'Preview Changes (Dry Run)'

  - stage: Apply
    dependsOn: Validate
    condition: succeeded()
    jobs:
      - job: ApplyChanges
        steps:
          - script: ./vcli repo apply repos/*.yaml
            displayName: 'Apply Configuration'
```

### Dry-Run Behavior

- **No API modifications**: Dry-run fetches current state from VBR (read-only) but makes no changes
- **No state updates**: State file is not modified in dry-run mode
- **Same exit codes**: Returns the same exit codes as regular apply (e.g., `6` if resource not found)
- **Full change preview**: Shows exactly what would change if applied

## Ignored Fields

Each resource type has fields that are excluded from drift detection because they are read-only or frequently changing (e.g., `lastRun`, `nextRun`, `statistics`, `id`). These are defined internally and cannot be overridden.
