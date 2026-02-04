# Init Command Examples (v0.11.0+)

Quick reference for the new non-interactive `init` command introduced in v0.11.0.

## Basic Usage

### Simple Init (Defaults)

```bash
# Non-interactive, creates files with defaults
./vcli init
```

**Output:**
```json
{
  "settings": {
    "selectedProfile": "vbr",
    "apiNotSecure": false,
    "credsFileMode": false
  },
  "profiles": [...],
  "files": {
    "settings": "settings.json",
    "profiles": "profiles.json"
  }
}
```

### Init with Custom Directory

```bash
# Write files to specific directory
./vcli init --output-dir ~/.vcli/

# Write to build directory (CI/CD)
./vcli init --output-dir $CI_PROJECT_DIR/.vcli/

# Write to relative path
./vcli init --output-dir ./config/
```

### Init with Configuration Flags

```bash
# Skip TLS verification (insecure)
./vcli init --insecure

# Enable credentials file mode
./vcli init --creds-file

# Combine flags
./vcli init --insecure --creds-file --output-dir ~/.vcli/
```

## Subcommands

### Initialize Settings Only

```bash
# Create only settings.json
./vcli init settings

# With flags
./vcli init settings --insecure --creds-file --output-dir ~/.vcli/
```

**Output:**
```json
{
  "settings": {
    "selectedProfile": "vbr",
    "apiNotSecure": true,
    "credsFileMode": false
  },
  "file": "/home/user/.vcli/settings.json"
}
```

### Initialize Profiles Only

```bash
# Create only profiles.json
./vcli init profiles

# With output directory
./vcli init profiles --output-dir ~/.vcli/
```

**Output:**
```json
{
  "profiles": [
    {
      "name": "vbr",
      "port": "9419",
      ...
    },
    ...
  ],
  "file": "/home/user/.vcli/profiles.json"
}
```

## JSON Output and Piping

### Extract Specific Fields

```bash
# Get settings only
./vcli init | jq '.settings'

# Get profiles only
./vcli init | jq '.profiles'

# Get file paths
./vcli init | jq '.files'

# Count profiles
./vcli init | jq '.profiles | length'
# Output: 7
```

### Save to Custom Locations

```bash
# Split settings and profiles into different files
./vcli init | jq '.settings' > /opt/vcli/settings.json
./vcli init | jq '.profiles' > /opt/vcli/profiles.json

# Save to environment-specific directories
ENV=prod
./vcli init --insecure | jq '.settings' > ~/.vcli/$ENV/settings.json
./vcli init | jq '.profiles' > ~/.vcli/$ENV/profiles.json
```

### Conditional Init

```bash
# Only init if files don't exist
if [ ! -f ~/.vcli/settings.json ]; then
    ./vcli init --output-dir ~/.vcli/
fi

# Check if init was successful
if ./vcli init --output-dir ~/.vcli/ > /dev/null 2>&1; then
    echo "Init successful"
else
    echo "Init failed"
    exit 1
fi
```

## CI/CD Examples

### GitHub Actions

```yaml
name: VBR Drift Detection
on: [push]

jobs:
  check-drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Download vcli
        run: |
          curl -L https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o vcli
          chmod +x vcli

      - name: Initialize vcli
        run: |
          ./vcli init --output-dir .vcli/ --insecure
          ./vcli profile --set vbr
        env:
          VCLI_SETTINGS_PATH: ${{ github.workspace }}/.vcli/

      - name: Login and check drift
        run: |
          ./vcli login
          ./vcli job diff --all --security-only
        env:
          VCLI_USERNAME: ${{ secrets.VBR_USERNAME }}
          VCLI_PASSWORD: ${{ secrets.VBR_PASSWORD }}
          VCLI_URL: ${{ secrets.VBR_URL }}
          VCLI_SETTINGS_PATH: ${{ github.workspace }}/.vcli/
```

### Azure DevOps

```yaml
trigger:
  - master

pool:
  vmImage: 'ubuntu-latest'

steps:
  - script: |
      curl -L https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o vcli
      chmod +x vcli
    displayName: 'Download vcli'

  - script: |
      ./vcli init --output-dir $(Build.SourcesDirectory)/.vcli/ --insecure
      ./vcli profile --set vbr
    displayName: 'Initialize vcli'
    env:
      VCLI_SETTINGS_PATH: $(Build.SourcesDirectory)/.vcli/

  - script: |
      ./vcli login
      ./vcli job diff --all --security-only
    displayName: 'Check for drift'
    env:
      VCLI_USERNAME: $(VCLI_USERNAME)
      VCLI_PASSWORD: $(VCLI_PASSWORD)
      VCLI_URL: $(VCLI_URL)
      VCLI_SETTINGS_PATH: $(Build.SourcesDirectory)/.vcli/
```

### GitLab CI

```yaml
vbr-drift-check:
  stage: test
  script:
    - curl -L https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o vcli
    - chmod +x vcli
    - ./vcli init --output-dir .vcli/ --insecure
    - ./vcli profile --set vbr
    - ./vcli login
    - ./vcli job diff --all --security-only
  variables:
    VCLI_SETTINGS_PATH: $CI_PROJECT_DIR/.vcli/
    VCLI_USERNAME: $VBR_USERNAME
    VCLI_PASSWORD: $VBR_PASSWORD
    VCLI_URL: $VBR_URL
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any

    environment {
        VCLI_SETTINGS_PATH = "${WORKSPACE}/.vcli/"
        VCLI_USERNAME = credentials('vbr-username')
        VCLI_PASSWORD = credentials('vbr-password')
        VCLI_URL = credentials('vbr-url')
    }

    stages {
        stage('Setup vcli') {
            steps {
                sh '''
                    curl -L https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o vcli
                    chmod +x vcli
                    ./vcli init --output-dir ${VCLI_SETTINGS_PATH} --insecure
                    ./vcli profile --set vbr
                '''
            }
        }

        stage('Check Drift') {
            steps {
                sh '''
                    ./vcli login
                    ./vcli job diff --all --security-only
                '''
            }
        }
    }
}
```

## Docker Examples

### Dockerfile with Init

```dockerfile
FROM alpine:latest

# Install dependencies
RUN apk add --no-cache bash curl jq

# Download vcli
RUN curl -L https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o /usr/local/bin/vcli \
    && chmod +x /usr/local/bin/vcli

# Create config directory
RUN mkdir -p /vcli-config

# Initialize vcli non-interactively
RUN vcli init --output-dir /vcli-config/ --insecure

# Set environment variable
ENV VCLI_SETTINGS_PATH=/vcli-config/

# Default profile
RUN vcli profile --set vbr

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
```

### Docker Entrypoint

```bash
#!/bin/bash
# entrypoint.sh
set -e

# Login if credentials provided
if [ -n "$VCLI_USERNAME" ] && [ -n "$VCLI_PASSWORD" ] && [ -n "$VCLI_URL" ]; then
    echo "Logging in to VBR..."
    vcli login
fi

# Execute command
exec "$@"
```

### Docker Compose

```yaml
version: '3.8'

services:
  vcli:
    build: .
    environment:
      - VCLI_USERNAME=${VBR_USERNAME}
      - VCLI_PASSWORD=${VBR_PASSWORD}
      - VCLI_URL=${VBR_URL}
      - VCLI_SETTINGS_PATH=/vcli-config/
    volumes:
      - ./specs:/specs
    command: vcli job diff --all --security-only
```

## Script Examples

### Bash Setup Script

```bash
#!/bin/bash
# setup-vcli.sh
set -e

VCLI_DIR="${HOME}/.vcli"
VCLI_BIN="${HOME}/.local/bin/vcli"

# Download vcli if not exists
if [ ! -f "$VCLI_BIN" ]; then
    echo "Downloading vcli..."
    curl -L https://github.com/shapedthought/vcli/releases/latest/download/vcli-linux-amd64 -o "$VCLI_BIN"
    chmod +x "$VCLI_BIN"
fi

# Initialize vcli
echo "Initializing vcli..."
"$VCLI_BIN" init --output-dir "$VCLI_DIR/" --insecure

# Set profile
echo "Setting profile to VBR..."
VCLI_SETTINGS_PATH="$VCLI_DIR/" "$VCLI_BIN" profile --set vbr

echo "Setup complete. Config files at: $VCLI_DIR"
echo "Set environment variables:"
echo "  export VCLI_USERNAME='your-username'"
echo "  export VCLI_PASSWORD='your-password'"
echo "  export VCLI_URL='vbr.example.com'"
echo "  export VCLI_SETTINGS_PATH='$VCLI_DIR/'"
```

### PowerShell Setup Script

```powershell
# setup-vcli.ps1
$ErrorActionPreference = "Stop"

$VcliDir = "$HOME\.vcli"
$VcliBin = "$HOME\.vcli\vcli.exe"

# Download vcli if not exists
if (-not (Test-Path $VcliBin)) {
    Write-Host "Downloading vcli..."
    Invoke-WebRequest -Uri "https://github.com/shapedthought/vcli/releases/latest/download/vcli-windows-amd64.exe" -OutFile $VcliBin
}

# Initialize vcli
Write-Host "Initializing vcli..."
& $VcliBin init --output-dir $VcliDir --insecure

# Set profile
Write-Host "Setting profile to VBR..."
$env:VCLI_SETTINGS_PATH = $VcliDir
& $VcliBin profile --set vbr

Write-Host "Setup complete. Config files at: $VcliDir"
Write-Host "Set environment variables:"
Write-Host "  `$env:VCLI_USERNAME = 'your-username'"
Write-Host "  `$env:VCLI_PASSWORD = 'your-password'"
Write-Host "  `$env:VCLI_URL = 'vbr.example.com'"
Write-Host "  `$env:VCLI_SETTINGS_PATH = '$VcliDir'"
```

### Multi-Environment Setup

```bash
#!/bin/bash
# multi-env-setup.sh
set -e

ENVIRONMENTS=("prod" "dev" "staging")

for ENV in "${ENVIRONMENTS[@]}"; do
    echo "Setting up $ENV environment..."

    # Create directory
    mkdir -p ~/.vcli/$ENV

    # Initialize
    ./vcli init --output-dir ~/.vcli/$ENV/ --insecure

    # Set profile
    VCLI_SETTINGS_PATH=~/.vcli/$ENV/ ./vcli profile --set vbr

    echo "$ENV setup complete"
done

echo "All environments configured"
echo "Usage: VCLI_SETTINGS_PATH=~/.vcli/prod/ ./vcli login"
```

## Legacy Interactive Mode

For backward compatibility during migration period:

```bash
# Use legacy interactive mode (deprecated)
./vcli init --interactive
```

**Output:**
```
⚠️  WARNING: Interactive mode is deprecated and will be removed in v0.12.0
   Use 'vcli init --interactive' to explicitly enable interactive mode
   Or use non-interactive mode: 'vcli init --insecure --creds-file'

Allow insecure TLS? (y/N): n
Use Creds file mode? (y/N): n
Initialized, ensure all environment variables are set.
```

**Note:** Interactive mode will be removed in v0.12.0. Update scripts to use non-interactive mode.

## Command Reference

```bash
# Main command
vcli init [flags]
vcli init [command]

# Subcommands
vcli init settings [flags]     # Initialize only settings.json
vcli init profiles [flags]     # Initialize only profiles.json

# Flags
--insecure                     # Skip TLS verification
--creds-file                   # Enable credentials file mode
--output-dir string            # Directory for config files
--interactive                  # Legacy interactive mode (deprecated)

# Global flags
-h, --help                     # Help for init command
```

## See Also

- [v0.11 Migration Guide](../docs/v0.11-migration-guide.md) - Migrating from v0.10.x
- [Getting Started](../docs/getting-started.md) - Complete setup guide
- [Authentication Guide](../docs/authentication.md) - Authentication workflows
- [Azure DevOps Integration](../docs/azure-devops-integration.md) - CI/CD pipeline templates
