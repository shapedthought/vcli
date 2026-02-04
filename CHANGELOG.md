# Changelog

All notable changes to vcli will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.11.0-beta1] - 2026-02-04

### Epic #66: Modernized Authentication & Automation

This release delivers major improvements to security, automation workflows, and CI/CD integration through a **clean break** from v0.10.x configuration format.

⚠️ **BREAKING CHANGES** - See [Breaking Changes](docs/breaking-changes-v0.11.md) | [Upgrade Guide](UPGRADING.md)

**Quick upgrade:**
```bash
vcli init-profiles  # Regenerate configs
vcli login          # Done!
```

### Breaking Changes

#### 1. Non-Interactive Init by Default
- `vcli init` now non-interactive by default (outputs JSON for scripting)
- Interactive mode available via `--interactive` flag
- Automation-first design for CI/CD workflows
- **Migration:** CI/CD scripts work unchanged; interactive users add `--interactive` flag

#### 2. Profile Commands Take Arguments
- Profile management commands now require arguments instead of prompting
- `vcli profile --set vbr` (was: interactive prompt)
- JSON output by default; `--table` flag for human-readable format
- **Migration:** Add profile name as argument in scripts

#### 3. Secure Token Storage
- Tokens stored in system keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- File-based fallback with encryption for systems without keychain
- Auto-authentication in CI/CD environments (non-TTY detection)
- Explicit token control via `VCLI_TOKEN` environment variable
- **Security:** No more plaintext tokens in `headers.json`
- **Migration:** Just `vcli login` after upgrade

#### 4. profiles.json v1.0 Format
- Completely restructured configuration format with versioning
- All profiles in single file (easy switching)
- Logical grouping of endpoints and headers
- Proper types (port as number, not string)
- **Migration:** Regenerate with `vcli init-profiles`

#### 5. Removed CredsFileMode
- Credentials now **always** from environment variables
- No more `--creds-file` option
- Consistent with 12-factor app principles and CI/CD standards
- **Migration:** Set `VCLI_USERNAME`, `VCLI_PASSWORD`, `VCLI_URL` environment variables

### Added

#### Authentication Improvements
- System keychain integration for secure token storage
- Hybrid token resolution: `VCLI_TOKEN` → keychain → auto-auth
- CI/CD environment detection (skips keychain on headless systems)
- File-based keyring with encryption as fallback
- Password prompting for file keyring in interactive sessions
- `VCLI_FILE_KEY` environment variable for non-interactive file keyring

#### Configuration Management
- Versioned profiles.json format (`"version": "1.0"`)
- Multi-profile support in single configuration file
- Structured endpoint and header configuration
- `init-profiles` command for profile-only initialization
- `init-settings` command for settings-only initialization

#### Commands
- `vcli init` - Non-interactive with JSON output by default
- `vcli init --interactive` - Legacy interactive mode
- `vcli init profiles` - Generate only profiles.json
- `vcli init settings` - Generate only settings.json
- `vcli profile --list --table` - Human-readable profile list
- `vcli profile --set <name>` - Argument-based profile switching

### Changed

#### Behavior Changes
- Default init mode is now non-interactive (JSON output)
- Profile commands require arguments instead of prompting
- Tokens never stored in plaintext files
- Configuration file permissions: `0600` (owner-only)
- Authentication flows through system keychain in interactive sessions
- Auto-authentication in CI/CD without keychain interaction

#### API Version Updates
- VBR: `1.1-rev0` → `1.3-rev1`
- VB365: `v6/v7` → `v8` (consistent version)
- VB for AWS: `1.3-rev0` → `1.4-rev0`
- VB for Azure: `v4` → `v5`
- VB for GCP: `1.1-rev0` → `1.2-rev0`
- VONE: `1.0-rev1` → `1.0-rev2`

### Removed
- `--creds-file` flag and CredsFileMode support
- Plaintext token storage in headers.json
- Interactive prompts from default init behavior
- Legacy profiles.json format (pre-v1.0)

### Security
- **Critical:** Hardcoded file keyring password replaced with `VCLI_FILE_KEY` env var + interactive prompt
- **Critical:** Configuration file permissions changed from `0644` to `0600`
- **Critical:** Tokens stored in OS-encrypted keychain instead of plaintext files
- Tilde expansion bug fixed in file keyring path resolution
- Token validation improved (JWT detection + length checks)

### Fixed
- Tilde expansion in file keyring directory path
- Token validation logic (flawed OR condition)
- Flag description inconsistencies (`apiNotSecure` vs `skipTLSVerify`)
- VB365 API version consistency (mixed v6/v7 → v8)
- Removed deprecated `authenticate()` method
- Removed unused code and comments

### Documentation
- Complete documentation restructure and improvements
- New focused guides: Authentication, Imperative Mode, Declarative Mode, State Management, Troubleshooting
- Comprehensive examples for jobs, repositories, SOBRs, KMS servers, and overlays
- Added Getting Started guide for new users
- Created Command Reference quick lookup guide
- Transformed user_guide.md into navigable index page
- **New:** Breaking changes documentation (docs/breaking-changes-v0.11.md)
- **New:** Upgrade guide (UPGRADING.md)
- CI/CD pipeline migration guide
- Troubleshooting common upgrade issues

### Dependencies
- Updated `jose2go` from v1.5.0 to v1.7.0 (fixes Snyk vulnerabilities)

### Design Philosophy: Why Clean Break?

We chose a clean break over backward compatibility:
- **Simpler:** 2-minute config regeneration vs complex migration logic
- **Cleaner:** No legacy format support to maintain
- **Better UX:** Clear errors, single fix that always works
- **Faster delivery:** Ship improvements immediately
- **Industry standard:** Major tools (Terraform, Kubernetes) make clean breaks

### Upgrade Notes

**For all users:**
1. Backup configs (optional): `cp -r ~/.vcli ~/.vcli.old`
2. Regenerate: `vcli init-profiles`
3. Re-login: `vcli login`
4. Test: `vcli get jobs`

**For CI/CD:**
- Environment variables work unchanged
- Update profile commands to pass arguments: `vcli profile --set vbr`
- See [Azure DevOps Integration Guide](docs/azure-devops-integration.md)

**Detailed guidance:**
- [Breaking Changes Documentation](docs/breaking-changes-v0.11.md)
- [Upgrade Guide](UPGRADING.md)

## [0.10.0-beta1] - 2024-01-15

### Epic #42: Declarative Resource Management & Remediation

Complete infrastructure-as-code workflow for VBR with drift detection and security alerting.

### Added

#### Core Features
- **Multi-Resource Declarative Management**
  - Jobs, repositories, SOBRs, encryption passwords, and KMS servers
  - Export resources to YAML format
  - Apply configurations with dry-run support
  - State management for drift detection

#### Export Commands
- `vcli export <job-id|name>` - Export jobs to YAML (full, simplified, or overlay format)
- `vcli export --all -d <dir>` - Export all jobs to directory
- `vcli repo export <name>` - Export repository configuration
- `vcli repo sobr-export <name>` - Export SOBR configuration
- `vcli encryption export <name>` - Export encryption password metadata
- `vcli encryption kms-export <name>` - Export KMS server configuration
- Export flags: `--as-overlay`, `--base`, `--simplified`

#### State Management
- `vcli repo snapshot <name>` - Capture current repository configuration
- `vcli repo sobr-snapshot <name>` - Capture current SOBR configuration
- `vcli encryption snapshot <name>` - Capture current encryption password
- `vcli encryption kms-snapshot <name>` - Capture current KMS server
- `--all` flag for bulk snapshots
- State file (`state.json`) tracks configurations and origins
- Adopt commands for bringing existing resources under management

#### Apply Commands
- `vcli job apply <file>` - Apply job configuration (create or update)
- `vcli repo apply <file>` - Apply repository configuration (update-only)
- `vcli repo sobr-apply <file>` - Apply SOBR configuration (update-only)
- `vcli encryption kms-apply <file>` - Apply KMS configuration (update-only)
- Apply flags: `--dry-run`, `-o/--overlay`, `--env`
- Support for configuration overlays (base + environment-specific)

#### Drift Detection
- `vcli job diff <name>` - Detect job configuration drift
- `vcli repo diff <name>` - Detect repository configuration drift
- `vcli repo sobr-diff <name>` - Detect SOBR configuration drift
- `vcli encryption diff <name>` - Detect encryption password drift
- `vcli encryption kms-diff <name>` - Detect KMS server drift
- `--all` flag for bulk drift detection
- `--severity <level>` filtering (critical, warning, info)
- `--security-only` flag for security-relevant drifts

#### Security-Aware Severity Classification
- **CRITICAL** severity for security-impacting changes:
  - Encryption disabled
  - Immutability disabled or reduced
  - GFS retention removed
  - Backup chain truncated
  - Application-aware processing disabled
  - Malware detection disabled
- **WARNING** severity for important changes:
  - Retention reduced
  - Compression changed
  - Schedule modified
  - Repository changed
  - Immutability period shortened
- **INFO** severity for minor changes:
  - Description updated
  - Labels changed
  - Non-security settings modified

#### Configuration Overlays
- Strategic merge system for base + overlay configurations
- Support for multiple environments (prod, dev, staging)
- `vcli.yaml` environment configuration file
- Overlay resolution priority: explicit `-o` > `--env` > `currentEnvironment`
- `vcli job plan <file>` - Preview merged configuration
- `--show-yaml` flag for full merged output

#### Exit Codes
- **Apply commands:**
  - 0 = Success
  - 1 = Error (API failure, invalid spec)
  - 5 = Partial apply (some fields skipped)
  - 6 = Resource not found (update-only resources)
- **Diff commands:**
  - 0 = No drift
  - 3 = Drift detected (INFO/WARNING)
  - 4 = Critical drift detected
  - 1 = Error occurred

### Changed
- Job export now produces full API fidelity by default (300+ fields)
- Simplified export available via `--simplified` flag for basic jobs
- State management integrated into apply workflow (automatic snapshots)

### Documentation
- New comprehensive drift detection guide
- Security alerting documentation with severity reference
- State management guide
- Azure DevOps integration examples
- Pipeline templates for drift monitoring and deployment

## [0.9.0-beta1] - 2023-10-15

### Added
- Initial declarative job management (jobs only)
- Job export command
- Job apply command with overlay support
- Job plan command
- Basic configuration overlay system
- Environment configuration via vcli.yaml

## [0.7.0] - 2023-08-01

### Added
- Job template system for creating jobs from templates
- Multi-file job templates with folder support
- Template creation from existing jobs

## [0.5.0] - 2023-05-01

### Added
- Utils command with multiple utilities
- VBR Job JSON Converter (GET to POST format)
- Check Version utility

### Changed
- Improved error handling and user feedback

## [0.4.0] - 2023-03-01

### Added
- PUT command for updating resources
- File-based payload support for POST/PUT
- YAML output format option

## [0.3.0] - 2023-01-01

### Added
- Profile system for multiple Veeam products
- Support for Enterprise Manager, VB365, VONE, cloud products
- Profile management commands (list, get, set, view)
- Creds file mode for faster profile switching

## [0.2.0] - 2022-11-01

### Added
- POST command for triggering operations and creating resources
- Environmental mode for credential management
- TLS certificate verification skip option

## [0.1.0] - 2022-09-01

### Added
- Initial release
- GET command for retrieving data from Veeam APIs
- Basic authentication for VBR
- Login command with token management
- Profile system for VBR
- JSON output format
- Settings and profiles configuration files

### Supported Products
- Veeam Backup & Replication (VBR)
- Enterprise Manager
- Veeam Backup for Microsoft 365
- Veeam ONE
- Veeam Backup for AWS
- Veeam Backup for Azure
- Veeam Backup for GCP

## Version Numbering

vcli follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backward compatible manner
- **PATCH** version for backward compatible bug fixes
- **Beta** releases (e.g., 0.10.0-beta1) for preview features

## See Also

- [GitHub Releases](https://github.com/shapedthought/vcli/releases) - Binary downloads and release notes
- [User Guide](user_guide.md) - Complete documentation
- [Getting Started](docs/getting-started.md) - Setup guide
- [Documentation](docs/) - All guides and references
