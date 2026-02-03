# Imperative Mode Guide

Imperative mode provides direct API access to all Veeam products. Use it for quick operations, one-off tasks, and API exploration.

## Table of Contents

- [Overview](#overview)
- [GET Command](#get-command)
- [POST Command](#post-command)
- [PUT Command](#put-command)
- [Utils Commands](#utils-commands)
- [Output Formats](#output-formats)
- [Using with jq](#using-with-jq)
- [Using with Nushell](#using-with-nushell)
- [Common Endpoints](#common-endpoints)
- [Best Practices](#best-practices)

## Overview

Imperative mode executes direct API operations against Veeam products. Each command maps to an HTTP method:

| Command | HTTP Method | Purpose |
|---------|-------------|---------|
| `vcli get` | GET | Retrieve data |
| `vcli post` | POST | Create resources, trigger operations |
| `vcli put` | PUT | Update resources |

**When to use imperative mode:**
- Quick API queries
- Triggering one-off operations (start backup, stop job)
- Exploring API capabilities
- Working with products without declarative support (VB365, VONE, cloud)

**When to use declarative mode:**
- Infrastructure-as-code workflows
- Multi-environment deployments
- Drift detection and monitoring
- GitOps automation

See [Declarative Mode Guide](declarative-mode.md) for declarative workflows.

## GET Command

Retrieve data from the API.

### Basic Syntax

```bash
vcli get <endpoint>
```

The endpoint is the path after the API version. vcli adds the version prefix automatically based on your profile.

### Examples

```bash
# VBR - Get all jobs
vcli get jobs

# VBR - Get specific job
vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c

# VBR - Get all repositories
vcli get backupInfrastructure/repositories

# VBR - Get managed servers
vcli get backupInfrastructure/managedServers

# Enterprise Manager - Get jobs (with query parameter)
vcli get jobs?format=entity

# VB365 - Get organizations
vcli get organizations

# VONE - Get alarms
vcli get alarms
```

### Understanding Endpoints

**Full VBR API endpoint:**
```
https://vbr.example.com:9419/api/v1/backupInfrastructure/managedServers
```

**vcli command:**
```bash
vcli get backupInfrastructure/managedServers
```

vcli automatically adds:
- Protocol (`https://`)
- Hostname (from `VCLI_URL` or profile)
- Port (from profile)
- API version prefix (`/api/v1/`)

You only specify the resource path.

### Output

Default output is JSON:

```bash
vcli get jobs
```

```json
{
  "data": [
    {
      "id": "c07c7ea3-0471-43a6-af57-c03c0d82354a",
      "name": "Backup Job 1",
      "type": "Backup",
      "isDisabled": false,
      ...
    }
  ]
}
```

### Output Formats

```bash
# JSON (default)
vcli get jobs

# YAML
vcli get jobs --yaml
```

## POST Command

Create resources or trigger operations.

### Basic Syntax

```bash
# Without payload
vcli post <endpoint>

# With payload
vcli post <endpoint> -f <file.json>
```

### Examples - Operations (No Payload)

```bash
# Start a backup job
vcli post jobs/57b3baab-6237-41bf-add7-db63d41d984c/start

# Stop a backup job
vcli post jobs/57b3baab-6237-41bf-add7-db63d41d984c/stop

# Retry a failed job
vcli post jobs/57b3baab-6237-41bf-add7-db63d41d984c/retry

# Enable a job
vcli post jobs/57b3baab-6237-41bf-add7-db63d41d984c/enable

# Disable a job
vcli post jobs/57b3baab-6237-41bf-add7-db63d41d984c/disable
```

### Examples - Creating Resources (With Payload)

```bash
# Create a backup job
vcli post jobs -f new-job.json

# Create a repository
vcli post backupInfrastructure/repositories -f repository.json

# Create credentials
vcli post backupInfrastructure/credentials -f creds.json
```

**Example payload (new-job.json):**
```json
{
  "name": "New Backup Job",
  "type": "Backup",
  "repository": {
    "id": "a1b2c3d4-..."
  },
  "schedule": {
    "enabled": true
  }
}
```

## PUT Command

Update existing resources.

### Basic Syntax

```bash
vcli put <endpoint> -f <file.json>
```

PUT requires a payload file.

### Examples

```bash
# Update a backup job
vcli put jobs/57b3baab-6237-41bf-add7-db63d41d984c -f updated-job.json

# Update a repository
vcli put backupInfrastructure/repositories/<id> -f updated-repo.json

# Update credentials
vcli put backupInfrastructure/credentials/<id> -f updated-creds.json
```

### GET vs PUT Object Differences (VBR Jobs)

**Important:** VBR GET and POST job objects have different structures. PUT uses the same structure as GET.

**Workflow:**
1. GET the job to see current configuration
2. Modify the JSON (GET structure)
3. PUT the modified JSON back

```bash
# 1. Get current job
vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c > job.json

# 2. Edit job.json (change retention, schedule, etc.)

# 3. Put updated job back
vcli put jobs/57b3baab-6237-41bf-add7-db63d41d984c -f job.json
```

**For creating new jobs (POST)**, you need to convert GET structure to POST structure. See [Utils - Job JSON Converter](#vbr-job-json-converter).

## Utils Commands

Utility commands for common tasks.

### VBR Job JSON Converter

Converts VBR GET job JSON to POST job JSON format.

**Why needed:** VBR API uses different JSON structures for GET (retrieve) vs POST (create) operations.

**Usage:**

```bash
# 1. Get a job
vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c > job-get.json

# 2. Run utils converter
vcli utils

# 3. Select "VBR Job JSON Converter"
# 4. Enter input file: job-get.json
# 5. Enter output file: job-post.json

# 6. Modify job-post.json as needed

# 7. Create new job
vcli post jobs -f job-post.json
```

**Note:** PUT operations use GET structure, so this converter is only needed for POST (creating new jobs).

### Check Version

Check if you're running the latest vcli release.

```bash
vcli utils
# Select "Check Version"
```

Compares your version against the latest GitHub release.

## Output Formats

vcli supports JSON and YAML output formats.

### JSON (Default)

```bash
vcli get jobs
```

```json
{
  "data": [
    {
      "id": "...",
      "name": "Backup Job 1"
    }
  ]
}
```

### YAML

```bash
vcli get jobs --yaml
```

```yaml
data:
  - id: "..."
    name: Backup Job 1
```

### Saving Output

```bash
# Save to file
vcli get jobs > jobs.json
vcli get jobs --yaml > jobs.yaml

# Pipe to tools
vcli get jobs | jq '.data[0]'
vcli get jobs --yaml | yq '.data[0]'
```

## Using with jq

[jq](https://stedolan.github.io/jq/) is a lightweight JSON processor perfect for parsing vcli output.

### Installation

```bash
# macOS
brew install jq

# Linux (Debian/Ubuntu)
sudo apt-get install jq

# Windows (Chocolatey)
choco install jq
```

### Basic Usage

```bash
# Get all job names
vcli get jobs | jq '.data[].name'

# Get jobs that are disabled
vcli get jobs | jq '.data[] | select(.isDisabled == true)'

# Get first job's ID
vcli get jobs | jq '.data[0].id'

# Filter by job type
vcli get jobs | jq '.data[] | select(.type == "Backup")'

# Get specific fields
vcli get jobs | jq '.data[] | {name: .name, type: .type, disabled: .isDisabled}'
```

### Useful jq Patterns

```bash
# See keys in an object (useful for exploring)
vcli get jobs | jq 'keys'
vcli get jobs | jq '.data[0] | keys'

# Count jobs
vcli get jobs | jq '.data | length'

# Pretty print
vcli get jobs | jq '.'

# Compact output
vcli get jobs | jq -c '.data[]'

# Raw string output (no quotes)
vcli get jobs | jq -r '.data[].name'

# Filter and format
vcli get jobs | jq '.data[] | "\(.name) (\(.type))"'
```

### Complex Example (Enterprise Manager)

```bash
# Get VM names and IDs from all jobs
vcli get jobs?format=entity | jq '.Jobs[].JobInfo.BackupJobInfo.Includes.ObjectInJobs[] | .Name, .ObjectInJobId'
```

## Using with Nushell

[Nushell](https://www.nushell.sh/) is a modern shell designed for structured data. It provides an excellent experience for working with API responses.

### Installation

```bash
# macOS
brew install nushell

# Linux
cargo install nu

# Windows
winget install nushell
# or
choco install nushell
```

Website: https://www.nushell.sh/book/installation.html

### Basic Usage

```bash
# Start Nushell
nu

# Set credentials
let-env VCLI_USERNAME = "administrator"
let-env VCLI_PASSWORD = "password"
let-env VCLI_URL = "vbr.example.com"

# Login
vcli login

# Get jobs and parse JSON automatically
vcli get jobs | from json

# Access the data array
vcli get jobs | from json | get data

# Filter disabled jobs
vcli get jobs | from json | get data | where isDisabled == true

# Select specific columns
vcli get jobs | from json | get data | select name type isDisabled

# Convert to YAML
vcli get jobs | from json | get data | to yaml

# Save as YAML
vcli get jobs | from json | get data | to yaml | save jobs.yaml --raw
```

### Nushell Advantages

- Automatic table formatting for structured data
- Powerful filtering with `where`
- Easy column selection with `select`
- Format conversion built-in (`to yaml`, `to json`, `to csv`)
- Type-aware operations

### Nushell Modules

Create reusable command modules:

**v.nu:**
```nu
# Simplified vcli commands
export def vget [url: string] {
    vcli get $url | from json | get data
}

export def vlogin [] {
    vcli login
}

# With environment variables included
export-env {
    let-env VCLI_USERNAME = "administrator"
    let-env VCLI_PASSWORD = "password"
    let-env VCLI_URL = "vbr.example.com"
}
```

**Usage:**
```bash
# Import module
use v.nu

# Use simplified commands
v vlogin
v vget jobs
v vget jobs | where isDisabled == false
```

### Nushell vs jq

| Feature | jq | Nushell |
|---------|-----|---------|
| JSON parsing | Excellent | Excellent |
| Tables | Manual formatting | Automatic |
| Filtering | `select()` expressions | `where` clause (more readable) |
| Learning curve | Steeper | Gentler |
| Speed | Very fast | Fast |
| Use case | Quick scripts | Interactive exploration |

## Common Endpoints

### VBR (Veeam Backup & Replication)

```bash
# Jobs
vcli get jobs
vcli get jobs/<id>
vcli post jobs/<id>/start
vcli post jobs/<id>/stop

# Infrastructure
vcli get backupInfrastructure/repositories
vcli get backupInfrastructure/scaleOutRepositories
vcli get backupInfrastructure/managedServers
vcli get backupInfrastructure/credentials

# Sessions
vcli get sessions
vcli get sessions/<id>

# Restore points
vcli get restorePoints

# Configuration
vcli get configuration/serverSettings
```

### Enterprise Manager

```bash
# Jobs
vcli get jobs?format=entity
vcli get jobs/<id>

# Repositories
vcli get repositories

# Backup servers
vcli get backupServers
```

### VB365

```bash
# Organizations
vcli get organizations
vcli get organizations/<id>

# Jobs
vcli get jobs
vcli post jobs/<id>/start

# Backup repositories
vcli get backupRepositories

# Proxies
vcli get proxies
```

### VONE

```bash
# Alarms
vcli get alarms

# Reports
vcli get reports

# Infrastructure
vcli get infrastructure
```

## Best Practices

### 1. Save API Responses

```bash
# Save for reference
vcli get jobs > jobs-backup-$(date +%Y%m%d).json

# Save before making changes
vcli get jobs/57b3baab > job-before-change.json
vcli put jobs/57b3baab -f updated-job.json
```

### 2. Use Descriptive File Names

```bash
# Bad
vcli post jobs -f data.json

# Good
vcli post jobs -f new-sql-backup-job.json
```

### 3. Preview Before Modifying

```bash
# Get current config
vcli get jobs/57b3baab | jq '.'

# Make changes in editor, then apply
vcli put jobs/57b3baab -f modified-job.json
```

### 4. Script Common Operations

```bash
#!/bin/bash
# start-all-jobs.sh

for job_id in $(vcli get jobs | jq -r '.data[].id'); do
    echo "Starting job: $job_id"
    vcli post jobs/$job_id/start
done
```

### 5. Use jq for Filtering

```bash
# Get only enabled jobs
vcli get jobs | jq '.data[] | select(.isDisabled == false)'

# Get job IDs only
vcli get jobs | jq -r '.data[].id'
```

### 6. Combine with Other Tools

```bash
# Count jobs by type
vcli get jobs | jq -r '.data[].type' | sort | uniq -c

# Get job names in CSV
vcli get jobs | jq -r '.data[] | [.name, .type] | @csv' > jobs.csv

# Monitor job status in loop
watch 'vcli get jobs | jq ".data[] | {name: .name, status: .status}"'
```

## Troubleshooting

### Empty Response

**Problem:** Command succeeds but returns empty `{}`

**Causes:**
- Resource doesn't exist
- Wrong endpoint
- Permissions issue

**Solution:**
```bash
# Verify endpoint
vcli get jobs  # Should show list

# Check profile
vcli profile --get

# Verify authentication
vcli login
```

### JSON Parse Error

**Problem:** `invalid character '<' looking for beginning of value`

**Causes:**
- VBR returned HTML error page
- Wrong endpoint
- API service not running

**Solution:**
```bash
# Save response to see error
vcli get jobs > response.txt
cat response.txt  # Check if HTML error page
```

### "No such file or directory" with -f Flag

**Problem:** `vcli post jobs -f data.json` fails

**Solution:**
```bash
# Use absolute path
vcli post jobs -f /full/path/to/data.json

# Or relative from current directory
vcli post jobs -f ./data.json
```

## See Also

- [Command Reference](command-reference.md) - Quick command lookup
- [Authentication Guide](authentication.md) - Setup and credentials
- [Declarative Mode Guide](declarative-mode.md) - Infrastructure-as-code workflows
- [Getting Started](getting-started.md) - Complete setup guide
