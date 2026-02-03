# User Guide

vcli is a CLI tool for interacting with Veeam APIs across multiple products. It provides both imperative (direct API access) and declarative (infrastructure-as-code) workflows.

## What is vcli?

vcli simplifies working with Veeam APIs by:
- Managing authentication across multiple products (VBR, Enterprise Manager, VB365, VONE, cloud products)
- Providing consistent command structure for all Veeam APIs
- Supporting both imperative commands (GET, POST, PUT) and declarative workflows (export, apply, diff)
- Enabling infrastructure-as-code with configuration overlays
- Detecting configuration drift with security-aware severity classification

## Getting Started

New to vcli? Start here:

1. **[Getting Started Guide](docs/getting-started.md)** - Complete setup walkthrough
2. **[Authentication Guide](docs/authentication.md)** - Profiles, credentials, and authentication modes
3. **[Command Reference](docs/command-reference.md)** - Quick reference for all commands

## Choose Your Workflow

### Imperative Mode (All Products)

Direct API access for quick operations, one-off tasks, and API exploration.

**📖 [Imperative Mode Guide](docs/imperative-mode.md)**

**Common commands:**
```bash
# Query data
vcli get jobs
vcli get backupInfrastructure/repositories

# Trigger operations
vcli post jobs/<id>/start
vcli post jobs/<id>/stop

# Update resources
vcli put jobs/<id> -f updated-job.json
```

**Best for:**
- Quick API queries
- Starting/stopping jobs
- Exploring API capabilities
- Working with products without declarative support (VB365, VONE, Enterprise Manager)

### Declarative Mode (VBR Only)

Infrastructure-as-code workflows for backup jobs, repositories, SOBRs, encryption passwords, and KMS servers.

**📖 [Declarative Mode Guide](docs/declarative-mode.md)**

**Common commands:**
```bash
# Export resources to YAML
vcli export --all -d ./jobs/
vcli repo export --all -d ./repos/

# Apply configurations with overlays
vcli job apply base.yaml -o prod-overlay.yaml
vcli repo apply repo.yaml --dry-run

# Detect configuration drift
vcli job diff --all --security-only
vcli repo diff --all
```

**Best for:**
- Infrastructure-as-code workflows
- Multi-environment deployments (dev/staging/prod)
- Drift detection and monitoring
- GitOps automation
- Configuration standardization

## Drift Detection & Security (VBR)

Monitor VBR infrastructure for unauthorized changes with security-aware severity classification.

**📖 [Drift Detection Guide](docs/drift-detection.md)**
**📖 [Security Alerting](docs/security-alerting.md)**

**Quick reference:**
```bash
# Snapshot resources to establish baseline
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all

# Detect drift
vcli job diff --all --security-only
vcli repo diff --all --severity critical
vcli encryption kms-diff --all

# Exit codes:
# 0 = No drift
# 3 = Drift detected (INFO/WARNING)
# 4 = Critical drift (security-impacting)
# 1 = Error
```

## CI/CD Integration

Automate vcli in CI/CD pipelines for continuous monitoring and deployment.

**📖 [Azure DevOps Integration](docs/azure-devops-integration.md)**

**Pipeline templates:**
- [Security Drift Monitor](examples/pipelines/security-drift-monitor.yaml) - Scheduled drift detection
- [Deploy VBR Job](examples/pipelines/deploy-vbr-job.yaml) - Multi-environment job deployment
- [Export to Git](examples/pipelines/export-to-git.yaml) - Daily configuration snapshots

## Reference Documentation

### API Versions

Current default API versions (as of October 2023):

| Product            | Version | API Version | Notes     |
| ------------------ | ------- | ----------- | --------- |
| VBR                | 13.0    | 1.3-rev1    |           |
| Enterprise Manager | 12.0    | -           | No Change |
| VONE               | 12.0    | v2.1        |           |
| VB365              | 7.0     | -           |           |
| VBAWS              | 5.0     | 1.4-rev0    |           |
| VBAZURE            | 5.0     | -           |           |
| VBGCP              | 1.0     | 1.2-rev0    |           |

To change API version, update the `api_version` or `x-api-version` field in `profiles.json`.

**Veeam API documentation:** https://www.veeam.com/documentation-guides-datasheets.html

### Profiles

vcli supports multiple Veeam products through profiles:

| Profile | Product | Port | Authentication |
|---------|---------|------|----------------|
| `vbr` | Veeam Backup & Replication | 9419 | OAuth |
| `ent_man` | Enterprise Manager | 9398 | Basic Auth |
| `vb365` | Veeam Backup for Microsoft 365 | 4443 | OAuth |
| `vone` | Veeam ONE | 1239 | OAuth |
| `aws` | Veeam Backup for AWS | 11005 | OAuth |
| `azure` | Veeam Backup for Azure | 443 | OAuth |
| `gcp` | Veeam Backup for GCP | 13140 | OAuth |

```bash
# List all profiles
vcli profile --list

# Get current profile
vcli profile --get

# Set active profile
vcli profile --set vbr

# View profile details
vcli profile --profile vbr
```

## Tips and Tricks

### Using with jq

[jq](https://stedolan.github.io/jq/) is excellent for parsing vcli JSON output:

```bash
# Get all job names
vcli get jobs | jq '.data[].name'

# Filter disabled jobs
vcli get jobs | jq '.data[] | select(.isDisabled == true)'

# Get specific fields
vcli get jobs | jq '.data[] | {name: .name, type: .type}'

# Explore object structure
vcli get jobs | jq 'keys'
vcli get jobs | jq '.data[0] | keys'
```

### Using with Nushell

[Nushell](https://www.nushell.sh/) provides structured data handling with intuitive syntax:

```bash
# Parse JSON and explore
vcli get jobs | from json | get data

# Filter by criteria
vcli get jobs | from json | get data | where isDisabled == false

# Select specific columns
vcli get jobs | from json | get data | select name type

# Convert formats
vcli get jobs | from json | get data | to yaml
```

**Create reusable module (v.nu):**
```nu
export def vget [url: string] {
    vcli get $url | from json | get data
}

export-env {
    let-env VCLI_USERNAME = "administrator"
    let-env VCLI_PASSWORD = "password"
    let-env VCLI_URL = "vbr.example.com"
}
```

### Replacing Parameters with sd

The [sd](https://crates.io/crates/sd) tool (written in Rust 🦀) works like sed and allows you to replace strings using regex:

```bash
# Check current value
cat job.json | jq '.name'

# Replace value in file
sd '"name": "Backup Job 2"' '"name": "Backup Job 12"' ./job.json

# Pipe vcli output directly to sd
vcli get jobs/<id> | sd '"name": "Old Name"' '"name": "New Name"' > job.json
```

**Installation:**
```bash
# Windows (Chocolatey)
choco install sd-cli

# Rust package manager
cargo install sd

# macOS (Homebrew)
brew install sd
```

This is useful for quick changes without opening a text editor.

## Examples

The `examples/` directory contains ready-to-use configuration examples:

- **[examples/jobs/](examples/jobs/)** - Job configurations with overlays
- **[examples/repos/](examples/repos/)** - Repository configurations
- **[examples/sobrs/](examples/sobrs/)** - Scale-out repository configurations
- **[examples/pipelines/](examples/pipelines/)** - Azure DevOps pipeline templates
- **[examples/README.md](examples/README.md)** - Complete examples guide

## Documentation Index

### Setup and Configuration
- [Getting Started Guide](docs/getting-started.md) - Installation, initialization, first steps
- [Authentication Guide](docs/authentication.md) - Profiles, credentials, authentication modes
- [Command Reference](docs/command-reference.md) - Quick command lookup

### Usage Guides
- [Imperative Mode Guide](docs/imperative-mode.md) - GET, POST, PUT commands for all products
- [Declarative Mode Guide](docs/declarative-mode.md) - Infrastructure-as-code workflows for VBR

### Advanced Features
- [Drift Detection Guide](docs/drift-detection.md) - Configuration monitoring and drift detection
- [Security Alerting](docs/security-alerting.md) - Security-aware severity classification
- [Azure DevOps Integration](docs/azure-devops-integration.md) - CI/CD pipelines and automation

### Additional Resources
- [Examples](examples/README.md) - Configuration examples and templates
- [Pipeline Templates](examples/pipelines/) - Ready-to-use Azure DevOps pipelines
- [CLAUDE.md](CLAUDE.md) - AI assistance file (for development)

## Support

- **Issues:** https://github.com/shapedthought/vcli/issues
- **Documentation:** https://github.com/shapedthought/vcli/tree/master/docs

## Version

vcli follows semantic versioning. Use `vcli utils` and select "Check Version" to compare against the latest GitHub release.
