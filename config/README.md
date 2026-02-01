# vcli Configuration System

The vcli configuration system enables environment-aware overlay management through a `vcli.yaml` configuration file.

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

## Usage Patterns

### 1. Environment-Based Workflows

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

### 2. Explicit Environment Override

```bash
# Apply with specific environment (ignores currentEnvironment)
vcli job apply base-job.yaml --env staging
```

### 3. Custom Overlay Override

```bash
# Override configured overlay with custom file
vcli job apply base-job.yaml -o custom-overlay.yaml
```

### 4. View Configuration

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

## Implementation Notes

- Overlay paths can be absolute or relative
- Relative paths are resolved from `defaultOverlayDir`
- Configuration is optional (vcli works without vcli.yaml)
- Environment labels are NOT automatically applied (future feature)
- Profile switching based on environment is NOT implemented yet (future feature)
