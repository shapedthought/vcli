# Changelog

All notable changes to vcli will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Documentation
- Complete documentation restructure and improvements
- New focused guides: Authentication, Imperative Mode, Declarative Mode, State Management, Troubleshooting
- Comprehensive examples for jobs, repositories, SOBRs, KMS servers, and overlays
- Added Getting Started guide for new users
- Created Command Reference quick lookup guide
- Transformed user_guide.md into navigable index page

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
