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
- `export` - Export existing jobs to declarative YAML format
- `apply` - Create or update jobs from YAML configuration
- `plan` - Preview changes without applying them
- `diff` - Detect configuration drift between state and VBR

## How to use

Please see the [user guide](https://github.com/shapedthought/vcli/blob/master/user_guide.md) for more information on imperative commands.

### Declarative Job Management (VBR)

vcli now supports declarative infrastructure management for VBR backup jobs, enabling GitOps workflows and version control.

#### Quick Start

**1. Export existing job to YAML:**
```bash
vcli export <job-id> -o my-backup.yaml
```

**2. Edit the configuration:**
```yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: prod-db-backup
spec:
  type: Backup
  description: "Production database backup"
  repository: "Default Backup Repository"
  schedule:
    enabled: true
    daily: "22:00"
  objects:
    - type: VM
      name: "PROD-SQL-01"
```

**3. Preview changes:**
```bash
vcli plan my-backup.yaml
```

**4. Apply configuration:**
```bash
vcli apply my-backup.yaml
```

**5. Detect drift:**
```bash
vcli diff
```

#### Key Benefits

- **Version Control**: Track job configurations in Git
- **GitOps Ready**: Automated deployments via CI/CD
- **Drift Detection**: Identify manual changes made in VBR UI
- **Idempotent**: Safe to run repeatedly, only applies needed changes
- **Preview Changes**: See what will change before applying
- **State Management**: Prevents concurrent modifications

#### Workflow Example

```bash
# Export all jobs
vcli export --all -d ./jobs/

# Make changes to YAML files
vim jobs/prod-backup.yaml

# Preview changes
vcli plan jobs/prod-backup.yaml

# Apply if satisfied
vcli apply jobs/prod-backup.yaml

# Commit to version control
git add jobs/
git commit -m "Update backup schedule"
git push

# Later, detect drift
vcli diff  # Detects manual VBR UI changes
```

#### Commands Reference

| Command | Description | Example |
|---------|-------------|---------|
| `export` | Generate YAML from existing jobs | `vcli export <job-id> -o backup.yaml` |
| `export --all` | Export all jobs to directory | `vcli export --all -d ./configs/` |
| `apply` | Create/update jobs from YAML | `vcli apply backup.yaml` |
| `apply --dry-run` | Preview without applying | `vcli apply backup.yaml --dry-run` |
| `plan` | Show planned changes | `vcli plan backup.yaml` |
| `diff` | Check for configuration drift | `vcli diff` |
| `diff <name>` | Check specific resource | `vcli diff prod-backup` |

All declarative commands support `--json` output for CI/CD integration.

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
| 0.9.0-beta1 | **Declarative Job Management**: Added export, apply, plan, and diff commands for VBR. Terraform-style workflows with state management, drift detection, and GitOps support. Updated dependencies. |
