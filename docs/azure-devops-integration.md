# Azure DevOps Integration Guide

vcli integrates with Azure DevOps Pipelines to provide automated configuration assurance for Veeam Backup & Replication environments. This guide covers what works today, what the complete integration looks like, and what features are needed to get there.

## Quick Start: Pipeline Templates

Ready-to-use pipeline templates are available in [`examples/pipelines/`](../examples/pipelines/):

| Template | Purpose |
|----------|---------|
| [detect-remediate.yml](../examples/pipelines/detect-remediate.yml) | Scheduled drift detection and auto-remediation |
| [pr-validation.yml](../examples/pipelines/pr-validation.yml) | PR validation gate with dry-run |
| [deployment.yml](../examples/pipelines/deployment.yml) | Multi-stage deployment with approval gates |
| [nightly-compliance.yml](../examples/pipelines/nightly-compliance.yml) | Compliance report generation |

See the [pipeline README](../examples/pipelines/README.md) for setup instructions.

## Current Integration (v0.10.0-beta1)

### What Works Today

vcli is already usable in Azure DevOps Pipelines with these capabilities:

| Capability | Status | Notes |
|------------|--------|-------|
| CLI invocation in pipeline steps | Ready | Standard `script` or `Bash@3` tasks |
| Exit codes for pipeline gates | Ready | `0`=clean, `3`=warning drift, `4`=critical drift, `1`=error |
| Environment variable auth | Ready | `VCLI_USERNAME`, `VCLI_PASSWORD`, `VCLI_URL` |
| Severity filtering | Ready | `--severity critical`, `--security-only` flags |
| Scheduled drift scans | Ready | Pipelines support cron schedules with `always: true` |
| State file in Git | Ready | `state.json` can be committed and tracked |
| Human-readable output | Ready | Text output in pipeline logs |
| Declarative job management | Ready | `job export` → `job apply` → `job diff` for backup jobs |
| Declarative management (all resources) | Ready | `export` → `apply` → `diff` for repos, SOBRs, encryption, KMS (see [pipeline templates](../examples/pipelines/)) |
| JSON output for diff commands | Not yet | Diff commands only output text |
| Markdown report generation | Not yet | No `--format markdown` flag |
| Webhook notifications | Not yet | Must use pipeline scripts for notifications |
| SARIF/JUnit output | Not yet | No standard security format output |

### Basic Pipeline: Scheduled Drift Scan

This is the simplest integration — a scheduled pipeline that runs drift checks and fails on critical findings.

```yaml
# azure-pipelines-drift-scan.yml
trigger: none

schedules:
  - cron: "0 6 * * 1-5"
    displayName: "Weekday 6AM UTC drift scan"
    branches:
      include:
        - master
    always: true  # Run even with no code changes

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: 'veeam-credentials'  # Contains VCLI_USERNAME, VCLI_PASSWORD, VCLI_URL

steps:
  - script: |
      curl -L https://github.com/shapedthought/vcli/releases/download/v0.10.0-beta1/vcli-linux-amd64 -o vcli
      chmod +x vcli
    displayName: 'Install vcli'

  - script: ./vcli init && ./vcli login
    displayName: 'Authenticate to VBR'
    env:
      VCLI_USERNAME: $(VCLI_USERNAME)
      VCLI_PASSWORD: $(VCLI_PASSWORD)
      VCLI_URL: $(VCLI_URL)

  - script: |
      ./vcli job diff --all --security-only
      JOB_EXIT=$?

      ./vcli repo diff --all --security-only
      REPO_EXIT=$?

      ./vcli repo sobr-diff --all --security-only
      SOBR_EXIT=$?

      ./vcli encryption diff --all --security-only
      ENC_EXIT=$?

      ./vcli encryption kms-diff --all --security-only
      KMS_EXIT=$?

      # Fail pipeline if any CRITICAL drift found (exit code 4)
      for CODE in $JOB_EXIT $REPO_EXIT $SOBR_EXIT $ENC_EXIT $KMS_EXIT; do
        if [ $CODE -eq 4 ]; then
          echo "##vso[task.logissue type=error]CRITICAL security drift detected"
          exit 1
        fi
      done

      # Warn if any non-critical drift found (exit code 3)
      for CODE in $JOB_EXIT $REPO_EXIT $SOBR_EXIT $ENC_EXIT $KMS_EXIT; do
        if [ $CODE -eq 3 ]; then
          echo "##vso[task.logissue type=warning]Security drift detected (non-critical)"
          echo "##vso[task.complete result=SucceededWithIssues;]Drift detected"
          exit 0
        fi
      done

      echo "No security drift detected"
    displayName: 'Run drift detection'
    env:
      VCLI_PASSWORD: $(VCLI_PASSWORD)
```

### Credentials Management

Azure DevOps provides several options for storing VBR credentials securely:

**Variable Groups (recommended for most teams)**

1. Go to Pipelines > Library > Variable groups
2. Create a group called `veeam-credentials`
3. Add variables: `VCLI_USERNAME`, `VCLI_PASSWORD` (mark as secret), `VCLI_URL`
4. Reference in pipeline with `- group: 'veeam-credentials'`

**Azure Key Vault (recommended for enterprise)**

1. Store secrets in Azure Key Vault
2. Link a Variable Group to the Key Vault under Pipelines > Library
3. Secrets are fetched at pipeline runtime with full audit logging

```yaml
variables:
  - group: 'vcli-keyvault-linked'  # Linked to Azure Key Vault
```

**Secure Files (for settings.json and profiles.json)**

```yaml
steps:
  - task: DownloadSecureFile@1
    name: vcliSettings
    inputs:
      secureFile: 'vcli-settings.json'

  - task: DownloadSecureFile@1
    name: vcliProfiles
    inputs:
      secureFile: 'vcli-profiles.json'

  - script: |
      mkdir -p $HOME/.vcli
      cp $(vcliSettings.secureFilePath) $HOME/.vcli/settings.json
      cp $(vcliProfiles.secureFilePath) $HOME/.vcli/profiles.json
      ./vcli login
    env:
      VCLI_PASSWORD: $(VCLI_PASSWORD)
```

### Exit Codes and Pipeline Logic

vcli's exit codes map directly to Azure DevOps pipeline outcomes:

| vcli Exit Code | Pipeline Result | Action |
|----------------|----------------|--------|
| `0` | Succeeded | No drift — pipeline passes |
| `3` | SucceededWithIssues | INFO/WARNING drift — pipeline passes with warning |
| `4` | Failed | CRITICAL drift — pipeline fails, blocks deployment |
| `1` | Failed | Error — pipeline fails, needs investigation |

Use Azure DevOps logging commands to control how results appear:

```bash
# Mark step as succeeded with issues (yellow warning)
echo "##vso[task.complete result=SucceededWithIssues;]Non-critical drift found"

# Log a warning that appears in the pipeline summary
echo "##vso[task.logissue type=warning]3 WARNING-level drifts detected"

# Log an error
echo "##vso[task.logissue type=error]CRITICAL: Encryption disabled on production job"

# Tag the build for filtering
echo "##vso[build.addbuildtag]drift-detected"
echo "##vso[build.addbuildtag]critical-drift"
```

### Multi-Stage Pipeline: Scan, Report, Notify

```yaml
# azure-pipelines-drift-full.yml
trigger: none

schedules:
  - cron: "0 6 * * *"
    displayName: "Daily 6AM UTC"
    branches:
      include:
        - master
    always: true

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: 'veeam-credentials'

stages:
  - stage: Scan
    displayName: 'Drift Detection'
    jobs:
      - job: DriftScan
        displayName: 'Scan all resources'
        steps:
          - script: |
              curl -sL https://github.com/shapedthought/vcli/releases/download/v0.10.0-beta1/vcli-linux-amd64 -o vcli
              chmod +x vcli
              ./vcli init && ./vcli login
            displayName: 'Setup vcli'
            env:
              VCLI_USERNAME: $(VCLI_USERNAME)
              VCLI_PASSWORD: $(VCLI_PASSWORD)
              VCLI_URL: $(VCLI_URL)

          - script: |
              CRITICAL=0
              OUTPUT=""

              for CMD in \
                "job diff --all" \
                "repo diff --all" \
                "repo sobr-diff --all" \
                "encryption diff --all" \
                "encryption kms-diff --all"
              do
                RESULT=$(./vcli $CMD 2>&1) || true
                EXIT=$?
                OUTPUT="$OUTPUT\n--- vcli $CMD ---\n$RESULT\n"
                if [ $EXIT -eq 4 ]; then
                  CRITICAL=1
                fi
              done

              echo -e "$OUTPUT" > $(Build.ArtifactStagingDirectory)/drift-report.txt
              echo "##vso[task.setvariable variable=hasCritical;isOutput=true]$CRITICAL"

              if [ $CRITICAL -eq 1 ]; then
                echo "##vso[task.logissue type=error]CRITICAL drift detected"
                echo "##vso[build.addbuildtag]critical-drift"
                exit 1
              fi
            displayName: 'Run drift scan'
            name: scan
            env:
              VCLI_PASSWORD: $(VCLI_PASSWORD)

          - publish: $(Build.ArtifactStagingDirectory)
            artifact: DriftReport
            condition: always()
            displayName: 'Publish drift report'

  - stage: Notify
    displayName: 'Notifications'
    dependsOn: Scan
    condition: failed()
    jobs:
      - job: NotifyTeam
        steps:
          - download: current
            artifact: DriftReport

          - script: |
              curl -X POST -H 'Content-type: application/json' \
                --data '{
                  "text": "CRITICAL: VBR security drift detected. Pipeline: $(Build.DefinitionName) Build: $(Build.BuildNumber). Investigate immediately."
                }' \
                $(SLACK_WEBHOOK_URL)
            displayName: 'Send Slack alert'
            env:
              SLACK_WEBHOOK_URL: $(SLACK_WEBHOOK_URL)
```

### PR Validation: Config File Changes

If VBR configuration files (YAML job definitions, severity config) are stored in Git, a PR validation pipeline can verify changes before merge:

```yaml
# azure-pipelines-pr-validation.yml
trigger: none

pr:
  branches:
    include:
      - master
  paths:
    include:
      - infrastructure/**
      - overlays/**

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: 'veeam-credentials'

steps:
  - script: |
      curl -sL https://github.com/shapedthought/vcli/releases/download/v0.10.0-beta1/vcli-linux-amd64 -o vcli
      chmod +x vcli
      ./vcli init && ./vcli login
    displayName: 'Setup vcli'
    env:
      VCLI_USERNAME: $(VCLI_USERNAME)
      VCLI_PASSWORD: $(VCLI_PASSWORD)
      VCLI_URL: $(VCLI_URL)

  - script: |
      # Validate all changed YAML files
      for file in $(git diff --name-only origin/master...HEAD -- '*.yaml' '*.yml'); do
        echo "Validating: $file"
        ./vcli job plan "$file" --show-yaml
        if [ $? -ne 0 ]; then
          echo "##vso[task.logissue type=error]Validation failed for $file"
          exit 1
        fi
      done
    displayName: 'Validate job configurations'
    env:
      VCLI_PASSWORD: $(VCLI_PASSWORD)
```

Set this as a **required build validation** under Repos > Branches > Branch Policies for `master` to block PRs that contain invalid configurations.

### Self-Hosted Agents

For environments where the VBR server is not internet-accessible, use a self-hosted agent on the same network:

```yaml
pool:
  name: 'VeeamInfraPool'  # Self-hosted agent pool
  demands:
    - Agent.OS -equals Linux
```

Self-hosted agents are installed on a machine within your network. This avoids exposing the VBR API to the internet. The agent communicates outbound to Azure DevOps (no inbound ports required).

---

## Complete Integration

Epic #42 is now complete, enabling full declarative management with automated remediation across all VBR resource types:

- **Epic #23** — Configuration Assurance (detection and alerting for all VBR resource types) ✅ Complete
- **Epic #42** — Declarative Resource Management and Remediation (export, snapshot, apply for all resource types) ✅ Complete

### Declarative Management Workflow

Every VBR resource type now supports the full lifecycle:

```
export → snapshot → apply → diff
```

**Day 1: Establish baseline**
```yaml
steps:
  - script: |
      # Export current VBR configuration to YAML specs
      ./vcli repo export --all -o infrastructure/repos/
      ./vcli repo sobr-export --all -o infrastructure/sobrs/
      ./vcli encryption export --all -o infrastructure/encryption/
      ./vcli encryption kms-export --all -o infrastructure/kms/

      # Snapshot resources as desired state (no VBR changes)
      ./vcli repo snapshot --all
      ./vcli repo sobr-snapshot --all
      ./vcli encryption snapshot --all
      ./vcli encryption kms-snapshot --all
    displayName: 'Bootstrap declarative management'
```

The specs are committed to Git. State records each resource with `origin: "applied"`, meaning vcli knows what the configuration *should* look like and can remediate drift.

### Automated Detect and Remediate

The key pipeline enabled by Epic #42. Detects drift, auto-remediates what VBR allows, and reports what it can't fix.

```yaml
# azure-pipelines-remediate.yml
trigger: none

schedules:
  - cron: "0 6 * * 1-5"
    displayName: "Weekday 6AM UTC"
    branches:
      include:
        - master
    always: true

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: 'veeam-credentials'

stages:
  - stage: Detect
    displayName: 'Detect Drift'
    jobs:
      - job: DriftScan
        steps:
          - script: |
              curl -sL https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o vcli
              chmod +x vcli && ./vcli init && ./vcli login
            displayName: 'Setup'
            env:
              VCLI_USERNAME: $(VCLI_USERNAME)
              VCLI_PASSWORD: $(VCLI_PASSWORD)
              VCLI_URL: $(VCLI_URL)

          - script: |
              CRITICAL=0
              for CMD in \
                "job diff --all --security-only" \
                "repo diff --all --security-only" \
                "repo sobr-diff --all --security-only" \
                "encryption diff --all --security-only" \
                "encryption kms-diff --all --security-only"
              do
                ./vcli $CMD 2>&1 || true
                if [ $? -eq 4 ]; then CRITICAL=1; fi
              done
              echo "##vso[task.setvariable variable=hasCritical;isOutput=true]$CRITICAL"
              if [ $CRITICAL -eq 1 ]; then
                echo "##vso[build.addbuildtag]critical-drift"
              fi
            displayName: 'Scan for drift'
            name: scan
            env:
              VCLI_PASSWORD: $(VCLI_PASSWORD)

  - stage: Remediate
    displayName: 'Auto-Remediate'
    dependsOn: Detect
    condition: eq(dependencies.Detect.outputs['DriftScan.scan.hasCritical'], '1')
    jobs:
      - job: ApplySpecs
        steps:
          - checkout: self

          - script: |
              curl -sL https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o vcli
              chmod +x vcli && ./vcli init && ./vcli login
            displayName: 'Setup'
            env:
              VCLI_USERNAME: $(VCLI_USERNAME)
              VCLI_PASSWORD: $(VCLI_PASSWORD)
              VCLI_URL: $(VCLI_URL)

          - script: |
              PARTIAL=0

              # Apply all specs from Git
              for spec in infrastructure/repos/*.yaml; do
                ./vcli repo apply "$spec"
                if [ $? -eq 5 ]; then PARTIAL=1; fi
              done

              for spec in infrastructure/sobrs/*.yaml; do
                ./vcli repo sobr-apply "$spec"
                if [ $? -eq 5 ]; then PARTIAL=1; fi
              done

              for spec in infrastructure/jobs/*.yaml; do
                ./vcli job apply "$spec"
                if [ $? -eq 5 ]; then PARTIAL=1; fi
              done

              if [ $PARTIAL -eq 1 ]; then
                echo "##vso[task.logissue type=warning]Partial remediation — some fields rejected by VBR"
                echo "##vso[task.complete result=SucceededWithIssues;]Partial remediation"
              fi
            displayName: 'Apply desired state'
            env:
              VCLI_PASSWORD: $(VCLI_PASSWORD)

  - stage: Verify
    displayName: 'Post-Remediation Check'
    dependsOn: Remediate
    jobs:
      - job: VerifyClean
        steps:
          - script: |
              curl -sL https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o vcli
              chmod +x vcli && ./vcli init && ./vcli login
            displayName: 'Setup'
            env:
              VCLI_USERNAME: $(VCLI_USERNAME)
              VCLI_PASSWORD: $(VCLI_PASSWORD)
              VCLI_URL: $(VCLI_URL)

          - script: |
              ./vcli job diff --all
              ./vcli repo diff --all
              ./vcli repo sobr-diff --all
              ./vcli encryption diff --all
              ./vcli encryption kms-diff --all
            displayName: 'Verify no remaining drift'
            env:
              VCLI_PASSWORD: $(VCLI_PASSWORD)

  - stage: Notify
    displayName: 'Notifications'
    dependsOn:
      - Remediate
      - Verify
    condition: |
      or(
        failed(),
        eq(dependencies.Remediate.result, 'SucceededWithIssues')
      )
    jobs:
      - job: Alert
        steps:
          - script: |
              curl -X POST -H 'Content-type: application/json' \
                --data '{
                  "text": "VBR drift detected and remediation attempted. Some fields may require manual intervention. Pipeline: $(Build.DefinitionName) Build: $(Build.BuildNumber)"
                }' \
                $(SLACK_WEBHOOK_URL)
            displayName: 'Notify team'
            env:
              SLACK_WEBHOOK_URL: $(SLACK_WEBHOOK_URL)
```

### Apply Exit Codes

With Epic #42, apply commands return additional exit codes that pipelines can use:

| Exit Code | Command | Meaning | Pipeline Action |
|-----------|---------|---------|-----------------|
| `0` | diff | No drift | Pass |
| `0` | apply | All fields accepted by VBR | Pass |
| `3` | diff | INFO/WARNING drift | SucceededWithIssues |
| `4` | diff | CRITICAL drift | Fail — trigger remediation |
| `5` | apply | Partial — some fields accepted, some rejected by VBR | SucceededWithIssues — notify team |
| `6` | apply | Resource destroyed — deleted from VBR | Fail — manual recreation needed |
| `1` | any | Error | Fail — investigate |

### API-First Remediation

vcli follows the Kubernetes/Terraform model for remediation: **send all changes to VBR and let the API decide what's mutable**. Apply sends the desired state from the Git-tracked YAML spec, and VBR accepts or rejects each field. vcli reports the outcome:

```
Applying: Default Backup Repository
  Applied: maxTaskCount: 1 -> 4
  Applied: description: "old" -> "updated"
  Rejected by VBR: type: WinLocal -> LinuxHardened
    (field is immutable after creation)

2 fields applied, 1 field rejected.
```

This avoids maintaining field-level mutability maps that must stay in sync across VBR versions. The API is the authority on what can be changed. An optional `remediation-config.yaml` can pre-filter fields to suppress known failures, but it's not required.

### Full Resource Coverage

With Epic #23 Phases 2-4 complete, drift scans and remediation cover the entire VBR security surface:

```yaml
steps:
  - script: |
      # Current (v0.10.0-beta1)
      ./vcli job diff --all --security-only
      ./vcli repo diff --all --security-only
      ./vcli repo sobr-diff --all --security-only
      ./vcli encryption diff --all --security-only
      ./vcli encryption kms-diff --all --security-only

      # Epic #23 Phase 2 (planned — #29, #30, #31, #32)
      ./vcli rbac diff --all --security-only
      ./vcli infra servers diff --all --security-only
      ./vcli credentials diff --all --security-only
      ./vcli traffic diff --all --security-only

      # Epic #23 Phase 3 (planned — #33, #34, #35)
      ./vcli malware diff --all --security-only
      ./vcli config-backup diff --all --security-only
      ./vcli notifications diff --all --security-only
    displayName: 'Full security scan'
```

### Deployment Gate with Remediation

A multi-stage pipeline that validates before deploying, applies changes, and verifies afterward:

```yaml
stages:
  - stage: PreDeployCheck
    displayName: 'Pre-Deploy Compliance Check'
    jobs:
      - job: DriftGate
        steps:
          - template: templates/drift-scan.yml
            parameters:
              severity: 'critical'
              failOnDrift: true

  - stage: Deploy
    displayName: 'Deploy Configuration'
    dependsOn: PreDeployCheck
    condition: succeeded()
    jobs:
      - job: ApplyConfig
        steps:
          - script: |
              # Apply all resource specs from Git
              for spec in infrastructure/jobs/*.yaml; do
                ./vcli job apply "$spec"
              done
              for spec in infrastructure/repos/*.yaml; do
                ./vcli repo apply "$spec"
              done
            displayName: 'Apply VBR configuration'
            env:
              VCLI_PASSWORD: $(VCLI_PASSWORD)

  - stage: PostDeployVerify
    displayName: 'Post-Deploy Verification'
    dependsOn: Deploy
    jobs:
      - job: VerifyNoDrift
        steps:
          - template: templates/drift-scan.yml
            parameters:
              severity: 'warning'
              failOnDrift: false
```

### JSON Output for Machine Consumption

**Required feature:** `--output json` flag on all diff commands (see [Features Needed](#features-needed))

With JSON output, pipelines can parse results programmatically:

```yaml
steps:
  - script: |
      ./vcli job diff --all --output json > job-drift.json
      ./vcli repo diff --all --output json > repo-drift.json

      # Parse with jq for conditional logic
      CRITICAL_COUNT=$(jq '[.drifts[] | select(.severity == "CRITICAL")] | length' job-drift.json)
      echo "##vso[task.setvariable variable=criticalCount]$CRITICAL_COUNT"
    displayName: 'Structured drift scan'
```

### Markdown Reports Attached to Build Summary

**Required feature:** `--output markdown` flag or `vcli report` command (#38)

Azure DevOps can render Markdown reports directly in the build summary tab:

```yaml
steps:
  - script: |
      ./vcli report --format markdown > $(System.DefaultWorkingDirectory)/drift-summary.md
      echo "##vso[task.uploadsummary]$(System.DefaultWorkingDirectory)/drift-summary.md"
    displayName: 'Generate compliance report'
```

### Multi-Resource Correlation

**Required feature:** Correlation engine (#36)

```yaml
steps:
  - script: |
      ./vcli correlate --all --output json > correlation.json
      PATTERNS=$(jq '.patterns | length' correlation.json)
      if [ $PATTERNS -gt 0 ]; then
        echo "##vso[task.logissue type=error]$PATTERNS coordinated change patterns detected"
        echo "##vso[build.addbuildtag]correlated-drift"
        exit 1
      fi
    displayName: 'Check for coordinated changes'
```

### Dashboard Integration

Azure DevOps dashboards can display drift trends using several approaches:

**Build History Widget** — Shows pass/fail trend for the drift scan pipeline. A failing build = drift detected.

**Test Results Trend** — If vcli outputs JUnit XML format (see [Features Needed](#features-needed)), the Test Results Trend widget shows drift counts over time.

**Wiki Auto-Update** — Pipeline publishes the latest report to an Azure DevOps wiki page:

```yaml
steps:
  - script: |
      REPORT=$(./vcli report --format markdown)
      az devops configure --defaults \
        organization=$(System.CollectionUri) \
        project="$(System.TeamProject)"
      echo "$REPORT" | az repos wiki page update \
        --wiki "project-wiki" \
        --path "/Compliance/VBR-Drift-Report" \
        --content "$REPORT"
    displayName: 'Update wiki dashboard'
    env:
      AZURE_DEVOPS_EXT_PAT: $(System.AccessToken)
```

---

## Features Needed

These vcli features would complete the Azure DevOps integration story. They are ordered by impact.

### 1. Declarative Management for All Resources (High Impact)

**Status:** ✅ Complete (Epic #42)

The `export` → `snapshot` → `apply` → `diff` lifecycle is now implemented for all resource types (jobs, repositories, SOBRs, encryption passwords, KMS servers). This enables automated remediation pipelines. See [pipeline templates](../examples/pipelines/) for ready-to-use examples.

### 2. JSON Output for Diff Commands (High Impact)

**Status:** Not implemented
**Related issue:** Part of #37 (structured output)

Add `--output json` flag to all diff commands. The JSON schema should include:

```json
{
  "timestamp": "2026-02-01T06:00:00Z",
  "resource": "Backup Job 1",
  "resourceType": "VBRJob",
  "origin": "applied",
  "driftCount": 3,
  "highestSeverity": "CRITICAL",
  "securityDriftCount": 2,
  "drifts": [
    {
      "path": "isDisabled",
      "action": "modified",
      "stateValue": false,
      "vbrValue": true,
      "severity": "CRITICAL"
    }
  ],
  "metadata": {
    "lastApplied": "2026-02-01T14:30:00Z",
    "lastAppliedBy": "edwardhoward"
  }
}
```

**Why it matters:** Enables pipeline scripts to parse results with `jq`, build conditional logic, and forward structured data to external systems. The current text output requires fragile regex parsing.

### 3. Markdown Report Generation (High Impact)

**Status:** Planned as #38 (compliance report generation)

A `vcli report` command (or `--output markdown` on diff commands) that produces a formatted Markdown report. Azure DevOps renders Markdown in build summaries via the `##vso[task.uploadsummary]` logging command, making drift results visible without downloading artifacts.

### 4. Unified Scan Command (Medium Impact)

**Status:** Not planned

A single command that runs all diff checks and produces a combined report:

```bash
vcli scan --all --security-only --output json
```

This simplifies pipeline definitions from 5+ sequential commands to one. It could also enable the correlation engine (#36) to analyse cross-resource patterns.

### 5. JUnit XML Output (Medium Impact)

**Status:** Not planned

Azure DevOps has a built-in "Tests" tab that displays JUnit XML results with pass/fail counts, trends, and drill-down. If vcli outputs drift results in JUnit format, each drift becomes a "test case":

```xml
<testsuites>
  <testsuite name="Job Drift" tests="3" failures="2">
    <testcase name="Backup Job 1 - isDisabled" classname="job.drift">
      <failure message="CRITICAL: false (state) -> true (VBR)" />
    </testcase>
  </testsuite>
</testsuites>
```

This gives free trending, history, and dashboard widgets via the `PublishTestResults@2` task.

### 6. SARIF Output (Medium Impact)

**Status:** Not planned

The Static Analysis Results Interchange Format (SARIF) is a standard for security tool output. Azure DevOps has a "Scans" tab (via the SARIF SAST extension) that displays SARIF results alongside code. If drift results map to config files tracked in Git, SARIF output would show drifts annotated on the relevant lines.

### 7. Azure DevOps Logging Command Support (Low Impact)

**Status:** Not planned

An `--ado` flag that wraps output with Azure DevOps logging commands automatically:

```bash
./vcli job diff --all --ado
```

Would output:
```
##vso[task.logissue type=error]CRITICAL ~ isDisabled: false (state) -> true (VBR)
##vso[task.logissue type=warning]WARNING ~ schedule: modified
##vso[task.complete result=SucceededWithIssues;]2 security-relevant changes detected
```

This is a convenience — the same result can be achieved with shell scripting around the current text output, but native support reduces boilerplate.

### 8. Non-Interactive Mode Flag (Low Impact)

**Status:** Partially implemented (vcli already uses env vars for auth)

An explicit `--non-interactive` or `--ci` flag that:
- Suppresses any interactive prompts
- Ensures clean stdout (no progress spinners or ANSI codes)
- Guarantees exit codes are set correctly
- Reads all configuration from environment variables or flags only

vcli largely behaves this way already, but an explicit flag documents the intent and prevents future regressions.

---

## Feature Priority Matrix

| Feature | Azure DevOps Impact | Effort | Dependency |
|---------|-------------------|--------|------------|
| Declarative management (all resources) | Unlocks automated remediation pipelines | High | Epic #42 |
| JSON output for diffs | Unlocks programmatic parsing, webhook forwarding, dashboard data | Medium | None |
| Markdown report | Build summary integration, wiki dashboards, compliance artifacts | Medium | #38 |
| Unified scan command | Simplifies pipelines, enables correlation | Medium | #36 |
| JUnit XML output | Free trending/history via Tests tab | Low | JSON output |
| SARIF output | Security scan tab integration | Low | JSON output |
| ADO logging commands | Convenience for pipeline authors | Low | None |
| Non-interactive flag | CI/CD safety net | Low | None |

---

## Reference: Pipeline Templates

Complete, ready-to-use pipeline templates are available in [`examples/pipelines/`](../examples/pipelines/). The templates below show reusable steps that can be embedded in your own pipelines.

### Template: Reusable Drift Scan Step

Create a template that can be included in any pipeline:

```yaml
# templates/drift-scan.yml
parameters:
  - name: severity
    type: string
    default: 'critical'
  - name: failOnDrift
    type: boolean
    default: true

steps:
  - script: |
      CRITICAL=0

      for CMD in \
        "job diff --all --severity ${{ parameters.severity }}" \
        "repo diff --all --severity ${{ parameters.severity }}" \
        "repo sobr-diff --all --severity ${{ parameters.severity }}" \
        "encryption diff --all --severity ${{ parameters.severity }}" \
        "encryption kms-diff --all --severity ${{ parameters.severity }}"
      do
        ./vcli $CMD 2>&1 || true
        if [ $? -eq 4 ]; then
          CRITICAL=1
        fi
      done

      if [ $CRITICAL -eq 1 ] && [ "${{ parameters.failOnDrift }}" = "True" ]; then
        echo "##vso[task.logissue type=error]Critical drift detected"
        exit 1
      fi
    displayName: 'vcli drift scan'
    env:
      VCLI_PASSWORD: $(VCLI_PASSWORD)
```

Use in a pipeline:

```yaml
steps:
  - template: templates/drift-scan.yml
    parameters:
      severity: 'warning'
      failOnDrift: true
```

### Template: Remediate from Git Specs

A reusable step that applies all YAML specs from a directory structure:

```yaml
# templates/remediate.yml
parameters:
  - name: specsPath
    type: string
    default: 'infrastructure'

steps:
  - script: |
      PARTIAL=0
      DESTROYED=0

      # Apply job specs
      for spec in ${{ parameters.specsPath }}/jobs/*.yaml; do
        [ -f "$spec" ] || continue
        ./vcli job apply "$spec"
        EXIT=$?
        if [ $EXIT -eq 5 ]; then PARTIAL=1; fi
        if [ $EXIT -eq 6 ]; then DESTROYED=1; fi
      done

      # Apply repo specs
      for spec in ${{ parameters.specsPath }}/repos/*.yaml; do
        [ -f "$spec" ] || continue
        ./vcli repo apply "$spec"
        EXIT=$?
        if [ $EXIT -eq 5 ]; then PARTIAL=1; fi
        if [ $EXIT -eq 6 ]; then DESTROYED=1; fi
      done

      # Apply SOBR specs
      for spec in ${{ parameters.specsPath }}/sobrs/*.yaml; do
        [ -f "$spec" ] || continue
        ./vcli repo sobr-apply "$spec"
        EXIT=$?
        if [ $EXIT -eq 5 ]; then PARTIAL=1; fi
        if [ $EXIT -eq 6 ]; then DESTROYED=1; fi
      done

      if [ $DESTROYED -eq 1 ]; then
        echo "##vso[task.logissue type=error]One or more resources deleted from VBR — manual recreation required"
        exit 1
      fi

      if [ $PARTIAL -eq 1 ]; then
        echo "##vso[task.logissue type=warning]Partial remediation — some fields rejected by VBR API"
        echo "##vso[task.complete result=SucceededWithIssues;]Partial remediation"
      fi
    displayName: 'Apply desired state from Git'
    env:
      VCLI_PASSWORD: $(VCLI_PASSWORD)
```

### Template: Nightly Compliance Report

```yaml
# azure-pipelines-nightly-compliance.yml
trigger: none

schedules:
  - cron: "0 2 * * *"
    displayName: "Nightly 2AM UTC"
    branches:
      include:
        - master
    always: true

pool:
  vmImage: 'ubuntu-latest'

variables:
  - group: 'veeam-credentials'

steps:
  - script: |
      curl -sL https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o vcli
      chmod +x vcli
      ./vcli init && ./vcli login
    displayName: 'Setup'
    env:
      VCLI_USERNAME: $(VCLI_USERNAME)
      VCLI_PASSWORD: $(VCLI_PASSWORD)
      VCLI_URL: $(VCLI_URL)

  - script: |
      # Capture all drift output
      {
        echo "# VBR Compliance Report"
        echo "**Generated:** $(date -u '+%Y-%m-%d %H:%M UTC')"
        echo "**Pipeline:** $(Build.DefinitionName) #$(Build.BuildNumber)"
        echo ""

        echo "## Job Configuration"
        ./vcli job diff --all 2>&1 || true
        echo ""

        echo "## Repository Configuration"
        ./vcli repo diff --all 2>&1 || true
        echo ""

        echo "## Scale-Out Repositories"
        ./vcli repo sobr-diff --all 2>&1 || true
        echo ""

        echo "## Encryption Inventory"
        ./vcli encryption diff --all 2>&1 || true
        echo ""

        echo "## KMS Servers"
        ./vcli encryption kms-diff --all 2>&1 || true
      } > $(Build.ArtifactStagingDirectory)/compliance-report.md

      echo "##vso[task.uploadsummary]$(Build.ArtifactStagingDirectory)/compliance-report.md"
    displayName: 'Generate compliance report'
    env:
      VCLI_PASSWORD: $(VCLI_PASSWORD)

  - publish: $(Build.ArtifactStagingDirectory)
    artifact: ComplianceReport
    displayName: 'Publish report'
    condition: always()
```
