# vcli

![nu_demo](./assets/main.png)
**Formatting provided by [Nushell.](https://www.nushell.sh/)** vcli only gets the data!

See the [User guide](https://github.com/shapedthought/vcli/blob/master/user_guide.md) for more information.

NOTE:

- This is not an official Veeam tool and is provided under the MIT license.
- This tool is still in development so there maybe breaking changes in the future.

## What is it?

The vcli is a powerful CLI tool for interacting with Veeam APIs, featuring both imperative API operations and declarative infrastructure management.

**Supported Products:**
- VBR (Veeam Backup & Replication)
- Enterprise Manager
- VB365 (Veeam Backup for Microsoft 365)
- VONE (Veeam ONE)
- VB for Azure
- VB for AWS
- VB for GCP

**Key Features:**
- **Imperative Mode**: Direct API interactions (GET, POST, PUT)
- **Declarative Mode**: Terraform-style infrastructure management for VBR jobs
- Version control friendly YAML configurations
- GitOps workflows with drift detection
- State management and conflict prevention

You can also add new endpoints by updating the profiles.json file.

## Why?

The main aim here is to make using Veeam APIs more accessible by handling many of the barriers to entry such as authentication.

If you are already a power API user with your own tools, then fantastic, please carry on with what you are doing!

The benefits include:

- simple to install
- simple syntax
- run anywhere
- run on anything

VBR and VB365 have powerful PowerShell cmdlets which I encourage you to use if that is your preference. vcli is not designed as a replacement, just a compliment.

However, products such as VB for AWS/Azure/GCP do not have a command line interface, this is where the vcli can really help.

## Commands

### Imperative Commands
- `login` - Authenticate with Veeam APIs
- `get` - Retrieve information from the API
- `post` - Send POST requests with optional data payload
- `put` - Send PUT requests with data payload
- `profile` - Manage API profiles (get, list, set)
- `utils` - Additional tools for working with Veeam APIs
- `job` - Create jobs using templates (legacy)

### Declarative Commands (VBR Only)
- `export` - Export existing jobs to declarative YAML format (full 300+ field export)
- `job apply` - Create or update jobs from YAML configuration with overlay support
- `job plan` - Preview merged configurations and changes
- `diff` - Detect configuration drift between state and VBR (coming soon)

## How to use

Please see the [user guide](https://github.com/shapedthought/vcli/blob/master/user_guide.md) for more information on imperative commands.

### Declarative Job Management (VBR)

vcli supports declarative infrastructure management for VBR backup jobs with **configuration overlays** for multi-environment deployments, enabling GitOps workflows and version control.

#### Quick Start

**1. Export existing job to YAML (full configuration):**
```bash
vcli export <job-id> -o my-backup.yaml
# Exports complete job with all 300+ VBR API fields
```

**2. Create environment-specific overlays:**
```yaml
# base-backup.yaml - Common configuration
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
  labels:
    app: database
spec:
  type: VSphereBackup
  repository: default-repo
  storage:
    compression: Optimal
    retention:
      type: Days
      quantity: 7
  objects:
    - type: VirtualMachine
      name: db-server

# prod-overlay.yaml - Production overrides
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  labels:
    env: production
spec:
  description: "Production database backup (30-day retention)"
  repository: prod-repo
  storage:
    retention:
      quantity: 30
  schedule:
    daily: "02:00"

# dev-overlay.yaml - Development overrides
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  labels:
    env: development
spec:
  description: "Development database backup (3-day retention)"
  repository: dev-repo
  storage:
    retention:
      quantity: 3
  schedule:
    daily: "23:00"
```

**3. Preview merged configuration:**
```bash
# Preview production configuration
vcli job plan base-backup.yaml -o prod-overlay.yaml

# Preview development configuration
vcli job plan base-backup.yaml -o dev-overlay.yaml

# Show full merged YAML
vcli job plan base-backup.yaml -o prod-overlay.yaml --show-yaml
```

**4. Apply configuration (dry-run mode):**
```bash
# Preview what would be applied
vcli job apply base-backup.yaml -o prod-overlay.yaml --dry-run

# Apply when ready (note: actual job creation coming in next release)
vcli job apply base-backup.yaml -o prod-overlay.yaml
```

**5. Configure environments (optional):**
```yaml
# vcli.yaml - Environment configuration
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

Then simply:
```bash
vcli job plan base-backup.yaml  # Uses production overlay automatically
vcli job apply base-backup.yaml --env development  # Use specific environment
```

#### Key Benefits

- **Multi-Environment Support**: Single base config with environment-specific overlays
- **DRY Configuration**: Define common settings once, override only what differs
- **Strategic Merge**: Deep merge preserves base values while applying overrides
- **Version Control**: Track job configurations in Git
- **GitOps Ready**: Automated deployments via CI/CD
- **Rich Previews**: See merged configurations before applying
- **Environment Awareness**: vcli.yaml manages environment-specific settings
- **Full API Fidelity**: Export captures all 300+ VBR API fields

#### Workflow Examples

**Single Environment Workflow:**
```bash
# Export existing job
vcli export <job-id> -o prod-backup.yaml

# Make changes to YAML file
vim prod-backup.yaml

# Preview changes
vcli job plan prod-backup.yaml

# Apply if satisfied
vcli job apply prod-backup.yaml --dry-run

# Commit to version control
git add prod-backup.yaml
git commit -m "Update backup schedule"
git push
```

**Multi-Environment Workflow:**
```bash
# 1. Create base template and overlays
mkdir -p configs/overlays
vcli export <job-id> -o configs/base-backup.yaml

# 2. Create environment overlays
cat > configs/overlays/prod.yaml <<EOF
spec:
  storage:
    retention:
      quantity: 30
  schedule:
    daily: "02:00"
EOF

cat > configs/overlays/dev.yaml <<EOF
spec:
  storage:
    retention:
      quantity: 3
  schedule:
    daily: "23:00"
EOF

# 3. Configure environments
cat > vcli.yaml <<EOF
currentEnvironment: production
defaultOverlayDir: ./configs/overlays
environments:
  production:
    overlay: prod.yaml
  development:
    overlay: dev.yaml
EOF

# 4. Preview and apply
vcli job plan configs/base-backup.yaml  # Uses prod overlay (currentEnvironment)
vcli job plan configs/base-backup.yaml --env development  # Preview dev

# 5. Commit everything
git add configs/ vcli.yaml
git commit -m "Add multi-environment backup configuration"
git push

# 6. Deploy to different environments
vcli job apply configs/base-backup.yaml --env production
vcli job apply configs/base-backup.yaml --env development
```

#### Commands Reference

| Command | Description | Example |
|---------|-------------|---------|
| `export <job-id>` | Export job to YAML (full config) | `vcli export abc-123 -o job.yaml` |
| `export --all` | Export all jobs to directory | `vcli export --all -d ./configs/` |
| `export --simplified` | Export minimal format (legacy) | `vcli export abc-123 -o job.yaml --simplified` |
| `job apply` | Apply configuration with overlay | `vcli job apply base.yaml -o prod.yaml --dry-run` |
| `job apply --env` | Apply using environment overlay | `vcli job apply base.yaml --env production` |
| `job plan` | Preview merged configuration | `vcli job plan base.yaml -o prod.yaml` |
| `job plan --show-yaml` | Show full merged YAML | `vcli job plan base.yaml -o prod.yaml --show-yaml` |

**Overlay Resolution Priority:**
1. Explicit `-o/--overlay` flag (highest)
2. `--env` flag (looks up in vcli.yaml)
3. `currentEnvironment` from vcli.yaml
4. No overlay (base config only)

## Installing 🛠️

<b>IMPORTANT The only trusted source of this tool is directly from the release page, or building from source.</b>

<b>Please also check the checksum of the downloaded file before running.</b>

vcli runs in place without needing to be installed on a computer.

It can be added to your system's path to allow system wide access.

To download please go to the releases tab.

windows

    Get-FileHash -Path <file path> -Algorithm SHA256

Mac

    shasum -a 256 <file path>

Linux - distributions may vary

    sha256sum <file path>

## Compiling from source

If you wish to compile the tool from source, please clone this repo, install Golang, and run the following in the root directory.

Windows:

    go build -o vcli.exe

Mac/ Linux

    go build -o vcli

## Docker 🐋

NOTE: at time of writing there will be <b>no official docker image.</b>

To run vcli in an isolated environment you can use a Docker container.

    docker run --rm -it ubuntu bash

    wget <URL of download>

    ./vcli init

Persisting the init fils can be done with a bindmount, but note that this does open a potential security hole.

You can of course create your own docker image with vcli installed which can be built from source.

Example using local git clone:

    FROM golang:1.18 as build

    WORKDIR /usr/src/app

    COPY go.mod go.sum ./

    RUN go mod download && go mod verify

    COPY . .

    RUN go build -v -o /usr/local/bin/vcli

    FROM ubuntu:latest

    WORKDIR /home/vcli

    RUN apt-get update && apt-get upgrade -y && apt-get install vim -y

    COPY --from=build /usr/local/bin/vcli ./

Then exec into the container

    docker run --rm -it YOUR_ACCOUNT/vcli:0.2 bash

    cd /home/vcli

    ./vcli init

With the --rm flag the container will deleted immediately after use.

Even when downloading the vcli into a Docker container, ensure you check the checksum!

## Why Go?

The main reason for using Go, after careful consideration, was that it compiles to a single binary with all decencies included.

Python was a close second, and is a great language, but some of the complexities of dependency management can make it more difficult to get started.

If anyone knows me I love RUST, but for this decided that Go was a better choice.

If you prefer other languages, feel free to take this implementation as an example and build your own.

## Contribution 🤝

If you think something is missing, or you think you can make it better, feel free to send me a pull request.

## Issues and comments ☝️

If you have any issues or would like to see a feature added please raise an issue.

### Change Log 🪵

| Version     | Changes                                                             |
| ----------- | ------------------------------------------------------------------- |
| 0.1.0-beta1 | First beta                                                          |
| 0.2.0-beta1 | Added ability change settings files location                        |
| 0.3.0-beta1 | Added Enterprise Manager support                                    |
| 0.4.0-beta1 | Added POST command and added new credentials management option      |
| 0.5.0-beta1 | Added Utils and PUT commands                                        |
| 0.6.0-beta1 | Added version check and updated VBR and VB365 to latest the version |
| 0.7.0-beta1 | Added job template feature                                          |
| 0.8.0-beta1 | Bumped API versions in the init command                             |
| 0.9.0-beta1 | **Configuration Overlay System**: Strategic merge engine for multi-environment deployments. Enhanced export (300+ fields), vcli.yaml environment config, overlay support in apply/plan commands. Deep merge preserves base values. Full GitOps support. See [RELEASE_NOTES_v0.9.0-beta1.md](RELEASE_NOTES_v0.9.0-beta1.md) |
