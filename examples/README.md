# vcli Configuration Examples

This directory contains example configuration files demonstrating vcli's declarative job management features.

## Overlay System

The overlay system allows you to define a base template with common settings and then apply environment-specific overlays. This enables DRY (Don't Repeat Yourself) configuration management across multiple environments.

### Example: Multi-Environment Database Backup

#### Base Template (`overlay-base.yaml`)
The base template defines common settings shared across all environments:
- Backup type and VM selection
- Default repository
- Base retention policy (7 days)
- Schedule with retry configuration

#### Environment Overlays

**Production Overlay (`overlay-prod.yaml`)**
- Extends retention to 30 days
- Changes schedule to 02:00
- Uses production repository
- Adds production label

**Development Overlay (`overlay-dev.yaml`)**
- Reduces retention to 3 days
- Changes schedule to 23:00
- Uses development repository
- Adds development label

### How Merging Works

The strategic merge engine:
1. **Preserves base values** - Fields not mentioned in the overlay keep their base values
2. **Deep merges maps** - Nested objects are merged recursively
3. **Replaces arrays** - Arrays in the overlay completely replace base arrays
4. **Merges labels** - Labels and annotations are combined

Example:
```yaml
# Base
spec:
  storage:
    compression: Optimal
    retention:
      type: Days
      quantity: 7

# Overlay
spec:
  storage:
    retention:
      quantity: 30

# Merged Result
spec:
  storage:
    compression: Optimal  # Preserved from base
    retention:
      type: Days          # Preserved from base
      quantity: 30        # Updated from overlay
```

### Usage (Future)

Once the `--overlay` flag is implemented:

```bash
# Create production job
vcli job apply overlay-base.yaml --overlay overlay-prod.yaml

# Create development job
vcli job apply overlay-base.yaml --overlay overlay-dev.yaml

# Plan changes before applying
vcli job plan overlay-base.yaml --overlay overlay-prod.yaml
```

### Benefits

- **Consistency**: All environments share the same base configuration
- **Maintainability**: Update common settings in one place
- **Clarity**: Environment-specific differences are explicit
- **Safety**: Base template is version-controlled, overlays show what's different
- **Scalability**: Add new environments by creating new overlay files
