# Azure DevOps Pipeline Templates

Ready-to-use Azure DevOps pipeline templates for VBR configuration management with owlctl.

## Available Templates

| Template | Purpose | Trigger |
|----------|---------|---------|
| [bootstrap.yml](bootstrap.yml) | Capture current VBR state into Git (run once) | Manual |
| [retention-change.yml](retention-change.yml) | Declarative retention change walkthrough | Manual |
| [retention-change-gitops.yml](retention-change-gitops.yml) | GitOps retention change (reads/commits to Git) | Manual |
| [detect-remediate.yml](detect-remediate.yml) | Detect drift and auto-remediate | Scheduled (weekday 6AM UTC) |
| [pr-validation.yml](pr-validation.yml) | Validate spec changes before merge | Pull requests |
| [deployment.yml](deployment.yml) | Multi-stage deployment with approval gates | Push to master / manual |
| [nightly-compliance.yml](nightly-compliance.yml) | Generate compliance reports | Scheduled (nightly 2AM UTC) |

## Prerequisites

### 1. Variable Group

Create a Variable Group named `veeam-credentials` in your Azure DevOps project:

1. Go to **Pipelines → Library → Variable groups**
2. Click **+ Variable group**
3. Name: `veeam-credentials`
4. Add variables:
   - `OWLCTL_USERNAME` - VBR API username
   - `OWLCTL_PASSWORD` - VBR API password (mark as **secret**)
   - `OWLCTL_URL` - VBR server URL (e.g., `https://vbr.example.com:9419`)
5. (Optional) `SLACK_WEBHOOK_URL` for notifications

**Security note:** Use Azure Key Vault linked Variable Groups for production environments.

### 2. Infrastructure Specs in Git

Organize your VBR specs in a directory structure:

```
infrastructure/
├── jobs/
│   ├── daily-backup.yaml
│   └── weekly-full.yaml
├── repos/
│   ├── default-repo.yaml
│   └── linux-repo.yaml
├── sobrs/
│   └── performance-tier.yaml
└── kms/
    └── azure-kms.yaml
```

Bootstrap from existing VBR configuration:

**Option 1: Use the bootstrap pipeline (recommended)**

Run [bootstrap.yml](bootstrap.yml) once — it exports all jobs, snapshots all resources, and commits everything to Git automatically.

**Option 2: Manual bootstrap**

```bash
# Export current configuration as YAML specs
owlctl export --all -d infrastructure/jobs/

# Snapshot current state (records baseline in state.json)
owlctl repo snapshot --all
owlctl repo sobr-snapshot --all
owlctl encryption snapshot --all
owlctl encryption kms-snapshot --all

# Commit specs and state to Git
git add infrastructure/ state.json
git commit -m "Bootstrap VBR declarative management"
```

### 3. Environment with Approvals (deployment.yml only)

For the deployment pipeline, create an Environment with approval gates:

1. Go to **Pipelines → Environments**
2. Click **New environment**
3. Name: `vbr-production`
4. Click the environment, then **⋮ → Approvals and checks**
5. Add **Approvals** and select approvers

### 4. Self-Hosted Agent (if needed)

If VBR is not internet-accessible, use a self-hosted agent:

1. **Pipelines → Agent pools → Add pool**
2. Install agent on a machine with network access to VBR
3. Update pipeline `pool` section:

```yaml
pool:
  name: 'VeeamInfraPool'  # Your agent pool name
```

## Quick Start

### Option 1: Copy and customize

1. Copy the template files to your repository
2. Update paths and variable names as needed
3. Create the pipeline in Azure DevOps

### Option 2: Import directly

1. Go to **Pipelines → New pipeline**
2. Select your repository
3. Choose **Existing Azure Pipelines YAML file**
4. Select the template path (e.g., `examples/pipelines/detect-remediate.yml`)

## Exit Code Reference

owlctl exit codes determine pipeline outcomes:

| Exit Code | Meaning | Pipeline Behavior |
|-----------|---------|-------------------|
| `0` | Success / No drift | Passes |
| `3` | Warning-level drift | Succeeds with issues |
| `4` | Critical drift | Fails (triggers remediation) |
| `5` | Partial apply | Succeeds with issues |
| `6` | Resource not found | Fails (manual recreation needed) |
| `1` | Error | Fails |

## Template Details

### bootstrap.yml

**Purpose:** Run-once pipeline to capture the current VBR environment into Git

This pipeline exports all backup jobs as YAML files and snapshots all resource types (repositories, SOBRs, encryption passwords, KMS servers) into `state.json`. Run this once to establish the GitOps baseline before using other pipelines.

**Single stage:**
1. **Bootstrap** - Build owlctl, export all jobs, snapshot all resources, commit to Git

**Result in repo:**
```
infrastructure/
└── jobs/
    ├── backup-job-1.yaml
    ├── weekly-full.yaml
    └── ...
state.json    ← contains state for ALL resource types
```

**Requirements:**
- Self-hosted agent with Go installed
- Git repo with write permissions (uses `persistCredentials: true`)

**Notes:**
- Commits with `[skip ci]` to prevent pipeline trigger loops
- Safe to re-run: if no changes are detected, the commit is skipped
- Only jobs have YAML export; repos, SOBRs, encryption, and KMS are captured in `state.json` only

### retention-change.yml

**Purpose:** Step-by-step walkthrough of declarative job management

This pipeline demonstrates the complete GitOps workflow for changing a backup job's retention policy. It is designed as a learning resource and starting point for teams adopting declarative configuration management.

**Stages:**
1. **Setup** - Build owlctl and verify VBR connectivity
2. **Export** - Capture current job config as base YAML and snapshot state
3. **Plan** - Preview the retention change with dry-run (no changes made)
4. **Apply** - Apply the retention overlay to VBR
5. **Verify** - Confirm VBR matches desired state (drift check)

**Customization:**
```yaml
variables:
  # Change these to match your environment
  - name: JOB_NAME
    value: 'My Backup Job'        # Name of the job to modify
  - name: DESIRED_RETENTION_DAYS
    value: '30'                    # Target retention in days
```

**What It Demonstrates:**
- Exporting live VBR configuration as YAML
- Creating minimal overlays (only the fields you want to change)
- Dry-run to preview changes before applying
- Apply with overlay merge (base + overlay = desired state)
- Post-apply verification with drift detection
- Exit code handling for CI/CD integration

### retention-change-gitops.yml

**Purpose:** GitOps deployment pipeline that reads from and commits back to Git

Unlike `retention-change.yml` (which is ephemeral), this pipeline treats Git as the single source of truth. It reads job YAML and `state.json` from the repo, applies changes to VBR, then commits the updated state and re-exported job spec back to Git.

**Prerequisites:**
- Run `bootstrap.yml` first to populate `infrastructure/` and `state.json`
- Self-hosted agent with Go installed

**Stages:**
1. **Plan** - Validate prerequisites, create overlay, dry-run (read-only)
2. **Apply and Commit** - Apply to VBR, re-export job YAML, commit updated state to Git
3. **Verify** - Pull latest commit and confirm zero drift

**Customization:**
```yaml
variables:
  - name: JOB_NAME
    value: 'My Backup Job'        # Must match a bootstrapped job
  - name: DESIRED_RETENTION_DAYS
    value: '30'                    # Target retention in days
```

**What It Demonstrates:**
- Reading job specs and state from Git (not live VBR export)
- Deriving job filenames using the same sanitization as `owlctl export`
- Committing updated state and job YAML back to Git with `[skip ci]`
- Re-exporting from VBR after apply to capture authoritative state
- End-to-end verification that VBR matches the committed Git state

### detect-remediate.yml

**Purpose:** Automated drift detection and remediation

**Stages:**
1. **Detect** - Scan all resources for security drift
2. **Remediate** - Apply specs from Git (only if drift found)
3. **Verify** - Confirm remediation was successful
4. **Notify** - Alert team if manual intervention needed

**Customization:**
```yaml
variables:
  - name: specsPath
    value: 'my-custom-path'  # Change infrastructure path
  - name: owlctlVersion
    value: 'v1.1.0'  # Pin to specific version
```

### pr-validation.yml

**Purpose:** Gate PRs that modify infrastructure specs

**Stages:**
1. **Validate** - Dry-run all changed YAML files
2. **DriftCheck** - Check for unexpected drift in existing resources
3. **Summary** - Report validation results

**Branch Policy Setup:**
1. Go to **Repos → Branches → master → ⋮ → Branch policies**
2. Under **Build Validation**, click **+**
3. Select the PR validation pipeline
4. Enable **Require** and **Immediately when updated**

### deployment.yml

**Purpose:** Controlled deployment with approval gates

**Stages:**
1. **PreDeploy** - Validate all specs with dry-run
2. **Deploy** - Apply configuration (with optional approval)
3. **PostDeploy** - Verify changes were applied

**Manual Trigger:**
```yaml
trigger: none  # Change to enable auto-trigger on push
```

### nightly-compliance.yml

**Purpose:** Generate compliance reports for audit

**Output:**
- Markdown report attached to build summary
- Artifact: `ComplianceReport/compliance-report.md`
- Optional: Wiki page update

**Enable Wiki Update:**
```yaml
variables:
  - name: updateWiki
    value: 'true'
  - name: wikiPath
    value: '/Compliance/VBR-Drift-Report'
```

## Troubleshooting

### Authentication Failures

```
Error: failed to login
```

- Verify `OWLCTL_URL` includes port (e.g., `https://vbr:9419`)
- Check username format (may need `DOMAIN\user`)
- Ensure VBR REST API is enabled

### Resource Not Found (Exit 6)

```
Error: resource 'X' not found in VBR (update-only mode)
```

- Resource was deleted from VBR
- Must recreate in VBR console, then re-adopt

### TLS Certificate Errors

For self-signed certificates, you have two options:

**Option 1: Install CA certificate on agent (recommended)**

Install the VBR server's CA certificate in the agent's system trust store. owlctl will automatically trust certificates validated by the system trust store.

**Linux agent:**
```bash
# Ubuntu/Debian
sudo cp vbr-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# RHEL/CentOS
sudo cp vbr-ca.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust
```

**Windows agent:**
```powershell
# Import to Trusted Root Certification Authorities
Import-Certificate -FilePath vbr-ca.crt -CertStoreLocation Cert:\LocalMachine\Root
```

**Option 2: Skip TLS verification (not recommended for production)**

Set `skipTLSVerify: true` in settings.json. This bypasses certificate validation entirely and should only be used for testing.

```json
{
  "skipTLSVerify": true,
  "credsFileMode": false
}
```

**Warning:** Disabling TLS verification exposes credentials to man-in-the-middle attacks and should never be used in production environments.

### Pipeline Permissions

If wiki update fails:
1. Go to **Project Settings → Repositories → Security**
2. Grant **Contribute** permission to build service account

## See Also

- [Drift Detection Guide](../../docs/drift-detection.md)
- [Azure DevOps Integration](../../docs/azure-devops-integration.md)
- [Security Alerting](../../docs/security-alerting.md)
