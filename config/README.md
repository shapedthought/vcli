# vcli Configuration System

> **⚠️ Status: Planned Feature - Not Yet Implemented**
>
> The environment configuration system described in this document is **designed but not yet implemented**. This document serves as a specification for future development.
>
> **What works today:**
> ```bash
> # ✅ Explicit overlay specification works
> vcli job apply base.yaml -o prod-overlay.yaml
> vcli repo apply base.yaml -o prod-overlay.yaml
> ```
>
> **What doesn't work yet:**
> ```bash
> # ❌ Environment-based overlay resolution not implemented
> vcli job apply base.yaml --env production
>
> # ❌ Config commands don't exist yet
> vcli config set-env production
> vcli config show
> ```
>
> **Current workaround:** Use the `-o/--overlay` flag to explicitly specify overlay files for each command.

---

The vcli configuration system will enable environment-aware overlay management through a `vcli.yaml` configuration file.

## Configuration File Format

The `vcli.yaml` file defines:
- **Current environment**: Which environment is active
- **Environment configurations**: Overlay files, profiles, and labels for each environment
- **Default overlay directory**: Base path for relative overlay file paths

## File Locations

vcli searches for configuration in this order:
1. Path specified in `VCLI_CONFIG` environment variable
2. `vcli.yaml` in current directory
3. `~/.vcli/vcli.yaml` in home directory

## Configuration Structure

```yaml
# Current active environment
currentEnvironment: production

# Default directory for overlay files
defaultOverlayDir: ./overlays

# Environment-specific configurations
environments:
  production:
    overlay: prod-overlay.yaml      # Overlay file for this environment
    profile: vbr-prod                # VBR profile to use
    labels:                          # Labels applied to all resources
      env: production
      managed-by: vcli

  development:
    overlay: dev-overlay.yaml
    profile: vbr-dev
    labels:
      env: development
```

## Planned Usage Patterns

When implemented, the following workflows will be available:

### 1. Environment-Based Workflows (Planned)

```bash
# Set current environment
vcli config set-env production

# Apply job (automatically uses production overlay)
vcli job apply base-job.yaml

# Switch to development
vcli config set-env development

# Same command now uses development overlay
vcli job apply base-job.yaml
```

### 2. Explicit Environment Override (Planned)

```bash
# Apply with specific environment (ignores currentEnvironment)
vcli job apply base-job.yaml --env staging
```

### 3. Custom Overlay Override (Works Today)

```bash
# Override configured overlay with custom file
vcli job apply base-job.yaml -o custom-overlay.yaml
```

### 4. View Configuration (Planned)

```bash
# Show current configuration
vcli config show

# List available environments
vcli config list-envs

# Show environment details
vcli config show-env production
```

## Benefits

1. **DRY Configuration**: Define overlays once, reference by environment name
2. **Team Consistency**: Version-controlled vcli.yaml ensures team uses same overlays
3. **Safety**: Current environment is explicit, reducing accidental prod changes
4. **Flexibility**: Can still override with custom overlays when needed
5. **Multi-Project**: Different projects can have different vcli.yaml files

## Example Workflows

### MSP Managing Multiple Customers

```yaml
# acme-corp/vcli.yaml
currentEnvironment: customer-acme
defaultOverlayDir: ../overlays

environments:
  customer-acme:
    overlay: acme-overlay.yaml
    profile: vbr-acme
    labels:
      customer: acme-corp
      tier: premium
```

### Enterprise with Multiple Environments

```yaml
# company-backups/vcli.yaml
currentEnvironment: production
defaultOverlayDir: ./config/overlays

environments:
  production:
    overlay: prod.yaml
    profile: vbr-prod-datacenter1
    labels:
      env: production
      datacenter: dc1
      compliance: required

  production-dc2:
    overlay: prod-dc2.yaml
    profile: vbr-prod-datacenter2
    labels:
      env: production
      datacenter: dc2
      compliance: required

  staging:
    overlay: staging.yaml
    profile: vbr-staging
    labels:
      env: staging
```

## Implementation Status

**Not Yet Implemented:**
- `vcli.yaml` configuration file parsing
- `vcli config` command and subcommands
- `--env` flag for automatic overlay resolution
- `currentEnvironment` awareness
- Automatic profile switching based on environment
- Automatic label application from environment config

**Currently Available:**
- Manual overlay specification via `-o/--overlay` flag
- Overlay merge functionality (works when explicitly specified)
- All resource types support overlays when using `-o` flag

**Tracking:** This feature is planned for a future release. Follow the project roadmap for updates.
