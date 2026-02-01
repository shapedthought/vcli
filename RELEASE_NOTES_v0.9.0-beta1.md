# Release Notes: v0.9.0-beta1

**Release Date:** February 1, 2026
**Branch:** feature/overlay-system-phase1 → master
**Tag:** v0.9.0-beta1

## Overview

This release introduces the **Configuration Overlay System**, enabling declarative, multi-environment VBR job management through YAML configuration files with strategic overlay support.

## 🎯 Key Features

### 1. Strategic Merge Engine
Deep merge YAML configurations while preserving base values:
- Recursive map merging (nested objects)
- Strategic array replacement
- Label/annotation combining
- Configurable merge behavior

**Example:**
```yaml
# Base: 7-day retention, Optimal compression
# Overlay: 30-day retention
# Result: 30-day retention + Optimal compression (preserved)
```

### 2. Environment Configuration (vcli.yaml)
Manage multiple environments with a single configuration file:
```yaml
currentEnvironment: production
defaultOverlayDir: ./overlays
environments:
  production:
    overlay: prod-overlay.yaml
    profile: vbr-prod
  development:
    overlay: dev-overlay.yaml
    profile: vbr-dev
```

### 3. Enhanced Export
Export command now captures complete job configuration:
- **Before:** 21 lines (minimal fields)
- **After:** 300+ lines (all VBR API fields)
- Use `--simplified` flag for legacy format

### 4. Job Apply with Overlays
```bash
# Apply with explicit overlay
vcli job apply base-job.yaml -o prod-overlay.yaml --dry-run

# Apply with environment
vcli job apply base-job.yaml --env production

# Auto-apply current environment overlay
vcli job apply base-job.yaml
```

### 5. Job Plan with Rich Preview
```bash
vcli job plan base-job.yaml -o prod-overlay.yaml
```
Shows:
- Merged configuration preview
- Storage and schedule settings
- Backup objects list
- Optional full YAML output (`--show-yaml`)

## 📦 What's Included

### New Files
- `resources/merge.go` - Strategic merge engine (254 lines)
- `resources/merge_test.go` - Unit tests (230 lines)
- `resources/merge_example_test.go` - Integration tests (154 lines)
- `config/vcli_config.go` - Configuration management (205 lines)
- `config/vcli_config_test.go` - Config tests (147 lines)
- `config/README.md` - Configuration documentation
- `examples/overlay-*.yaml` - Example overlay files
- `examples/vcli.yaml` - Example configuration
- `examples/README.md` - Usage guide
- `PHASE1_COMPLETE.md` - Detailed completion summary

### Modified Files
- `cmd/apply.go` - Added overlay support (347 lines)
- `cmd/plan.go` - Enhanced with overlay support (306 lines)
- `cmd/export.go` - Full export by default (+56 lines)
- `models/vbrJobsmodels.go` - Fixed API v1.3 structure
- `.gitignore` - Allow example YAML files

### Removed Files
- `cmd/diff.go` - Removed incomplete implementation

## 🧪 Testing

**Test Results:**
- Config tests: 4/4 passing ✅
- Merge tests: 13/13 passing ✅
- Integration test: 1/1 passing ✅
- **Total: 18/18 tests passing** ✅

**Manual Testing:**
- Export command with live VBR v1.3 ✅
- Plan command with overlays ✅
- Apply command dry-run mode ✅
- Label merging ✅
- Deep nested object merging ✅

## 📊 Statistics

- **8 commits** merged from feature branch
- **2,188 lines added**
- **579 lines deleted**
- **23 files changed**
- **9 new files created**

## 🚀 Usage Examples

### Export Existing Job
```bash
vcli export 57b3baab-6237-41bf-add7-db63d41d984c -o my-job.yaml
```

### Create Environment Overlays
```bash
# Production overlay
cat > prod-overlay.yaml <<EOF
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: my-job
spec:
  storage:
    retention:
      quantity: 30
  repository: prod-repo
  schedule:
    daily: "02:00"
EOF

# Development overlay
cat > dev-overlay.yaml <<EOF
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: my-job
spec:
  storage:
    retention:
      quantity: 3
  repository: dev-repo
  schedule:
    daily: "23:00"
EOF
```

### Preview Merged Configuration
```bash
# Preview with production overlay
vcli job plan my-job.yaml -o prod-overlay.yaml

# Preview with development overlay
vcli job plan my-job.yaml -o dev-overlay.yaml

# Show full YAML
vcli job plan my-job.yaml -o prod-overlay.yaml --show-yaml
```

### Apply Configuration (Dry-run)
```bash
vcli job apply my-job.yaml -o prod-overlay.yaml --dry-run
```

## 🔄 Workflow Examples

### Pattern 1: Export-Modify-Plan-Apply
```bash
# 1. Export existing job
vcli export job-123 -o base.yaml

# 2. Create overlay for changes
echo "spec:
  storage:
    retention:
      quantity: 14" > changes.yaml

# 3. Preview merged result
vcli job plan base.yaml -o changes.yaml

# 4. Apply (when ready)
vcli job apply base.yaml -o changes.yaml
```

### Pattern 2: Multi-Environment Management
```bash
# 1. Configure environments
cat > vcli.yaml <<EOF
currentEnvironment: production
defaultOverlayDir: ./overlays
environments:
  production:
    overlay: prod.yaml
  development:
    overlay: dev.yaml
EOF

# 2. Create base job template
cat > base-job.yaml <<EOF
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
spec:
  type: VSphereBackup
  objects:
    - type: VirtualMachine
      name: db-server
  repository: default-repo
  storage:
    compression: Optimal
    retention:
      type: Days
      quantity: 7
EOF

# 3. Switch environments and apply
vcli config set-env production
vcli job plan base-job.yaml    # Uses prod overlay

vcli config set-env development
vcli job plan base-job.yaml    # Uses dev overlay
```

## ⚠️ Breaking Changes

None. All existing vcli commands continue to work unchanged.

## 📝 Known Limitations

1. **Apply command** - Dry-run mode only in this release
   - Actual job creation/update deferred to next release
   - Use for validation and planning workflows

2. **State management** - Not yet implemented
   - No tracking of created/updated jobs
   - Coming in next release

3. **Drift detection** - Limited
   - No comparison against current VBR state
   - Plan shows merged config only

4. **Repository resolution** - Not implemented
   - Must use repository IDs, not names
   - Coming in next release

## 🔜 What's Next

**Phase 2 priorities:**
1. Implement actual job creation in apply command
2. Add state management (`.vcli-state.json`)
3. Implement job updates (PUT operations)
4. Add repository name → ID resolution
5. Full drift detection (compare state vs VBR)

## 🐛 Bug Fixes

- Fixed VBR Job model to match actual API v1.3-rev1 response format
- Removed incorrect `inventoryObject` nesting from includes
- Export command now correctly populates objects array

## 📚 Documentation

New documentation added:
- `PHASE1_COMPLETE.md` - Comprehensive Phase 1 summary
- `config/README.md` - Configuration system guide
- `examples/README.md` - Overlay examples and patterns
- `RELEASE_NOTES_v0.9.0-beta1.md` - This file

## 🙏 Acknowledgments

This release was developed with assistance from Claude Sonnet 4.5.

## 📞 Feedback

Report issues at: https://github.com/shapedthought/vcli/issues

---

**Full Changelog:** https://github.com/shapedthought/vcli/compare/v0.8.0-beta1...v0.9.0-beta1
