# Azure DevOps Pipeline Templates

Ready-to-use Azure DevOps pipeline templates for VBR configuration management with vcli.

## Available Templates

| Template | Purpose | Trigger |
|----------|---------|---------|
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
   - `VCLI_USERNAME` - VBR API username
   - `VCLI_PASSWORD` - VBR API password (mark as **secret**)
   - `VCLI_URL` - VBR server URL (e.g., `https://vbr.example.com:9419`)
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

```bash
# Export current configuration
vcli job export --all -o infrastructure/jobs/
vcli repo export --all -o infrastructure/repos/
vcli repo sobr-export --all -o infrastructure/sobrs/
vcli encryption kms-export --all -o infrastructure/kms/

# Adopt as desired state
for spec in infrastructure/**/*.yaml; do
  vcli adopt "$spec"
done

# Commit to Git
git add infrastructure/
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

vcli exit codes determine pipeline outcomes:

| Exit Code | Meaning | Pipeline Behavior |
|-----------|---------|-------------------|
| `0` | Success / No drift | Passes |
| `3` | Warning-level drift | Succeeds with issues |
| `4` | Critical drift | Fails (triggers remediation) |
| `5` | Partial apply | Succeeds with issues |
| `6` | Resource not found | Fails (manual recreation needed) |
| `1` | Error | Fails |

## Template Details

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
  - name: vcliVersion
    value: 'v1.0.0'  # Pin to specific version
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

- Verify `VCLI_URL` includes port (e.g., `https://vbr:9419`)
- Check username format (may need `DOMAIN\user`)
- Ensure VBR REST API is enabled

### Resource Not Found (Exit 6)

```
Error: resource 'X' not found in VBR (update-only mode)
```

- Resource was deleted from VBR
- Must recreate in VBR console, then re-adopt

### TLS Certificate Errors

For self-signed certificates, the pipelines trust all certificates by default. If needed, add:

```yaml
- script: |
    export VCLI_SKIP_TLS_VERIFY=true
    ./vcli login
```

### Pipeline Permissions

If wiki update fails:
1. Go to **Project Settings → Repositories → Security**
2. Grant **Contribute** permission to build service account

## See Also

- [Drift Detection Guide](../../docs/drift-detection.md)
- [Azure DevOps Integration](../../docs/azure-devops-integration.md)
- [Security Alerting](../../docs/security-alerting.md)
