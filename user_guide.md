# User Guide

## Table of Contents

- [API Versions](#api-versions)
- [Init](#init)
- [Profiles](#profiles)
- [Log in](#log-in)
  - [Environmental mode](#environmental-mode)
  - [Creds file mode](#creds-file-mode)
  - [Logging in](#logging-in)
  - [Change Modes](#change-modes)
- [API Commands overview](#api-commands-overview)
- [Get](#get)
- [Post / Put](#post--put)
- [Job](#job)
- [Utils](#utils)
  - [VBR Job JSON Converter](#vbr-job-json-converter)
  - [Check Version](#check-version)
- [Using with jq](#using-with-jq)
- [Using with Nushell](#using-with-nushell)
  - [Nu Modules](#nu-modules)
- [Drift Detection & Security Alerting (VBR)](#drift-detection--security-alerting-vbr)
- [Declarative Job Management (VBR)](#declarative-job-management-vbr)
  - [Overview](#overview)
  - [Key Concepts](#key-concepts)
  - [Export Command](#export-command)
  - [Creating Overlays](#creating-overlays)
  - [Job Plan Command](#job-plan-command)
  - [Job Apply Command](#job-apply-command)
  - [Environment Configuration (vcli.yaml)](#environment-configuration-vcliyaml)
  - [Strategic Merge Behavior](#strategic-merge-behavior)
  - [Complete Multi-Environment Example](#complete-multi-environment-example)
  - [Best Practices](#best-practices)
  - [Troubleshooting](#troubleshooting)
- [Tips and Tricks](#tips-and-tricks)

## API Versions

The default API versions are as follows:

| Product            | Version | API Version | Notes     |
| ------------------ | ------- | ----------- | --------- |
| VBR                | 13.0    | 1.3-rev1    |           |
| Enterprise Manager | 12.0    | -           | No Change |
| VONE               | 12.0    | v2.1        |           |
| V365               | 7.0     | -           |           |
| VBAWS              | 5.0     | 1.4-rev0    |           |
| VBAZURE            | 5.0     | -           |           |
| VBGCP              | 1.0     | 1.2-rev0    |           |

Updated 18/10/2023

If you wish to change API version, please update the profile in the profiles.json file.

https://www.veeam.com/documentation-guides-datasheets.html

## Init

To start using the app enter:

    ./vcli.exe init

This will create two files:

- settings.json - contains settings that can be adjusted
- profiles.json - profiles of each of the APIs

If you have set the VCLI_SETTINGS_PATH in the environmental variables before running this command, the files will be located in that directory. Otherwise they will be created in the directory you ran the command.

Note that the directory will not be automatically created for you, it will need to be in place before running the command.

Examples:

Windows

    "C:\User\UserName\.vcli\"

Linux

    "home/veeam/.vcli/"

## Profiles

The profiles.json file contains key information for each of the APIs, these mainly differ in terms of Port, and login URL.

The profiles currently are:

- vbr
- ent_man (Enterprise Manager)
- vb365
- vone
- aws
- azure
- gcp

To see a list of the profiles run

    ./vcli profile --list / -l

To see the current set profiles run

    ./vcli profiles --get / -g

To set a new profile run

    ./vcli profiles --set / -s

To see the details of a profile run

    ./vcli profiles --profile / -p

    {
    "name": "ent_man",
    "headers": {
        "accept": "application/json",
        "Content-type": "application/json",
        "x-api-version": ""
    },
    "url": ":9398/api/sessionMngr/?v=latest",
    "port": "9398",
    "api_version": "",
    "username": "administrator@",
    "address": "192.168.0.123"
    }

## Log in

_UPDATED_

There are now two modes to log in, the first is the "environmental" mode, and the second takes some of the credentials data from the profile.json file aka "creds file" mode.

### Environmental mode

When running the "init" command select N/No to use this mode.

Before logging in you will need to set the following environmental variables:

| Name               | Description                                                                        |
| ------------------ | ---------------------------------------------------------------------------------- |
| VCLI_USERNAME      | The username of the API you are logging into                                       |
| VCLI_PASSWORD      | The password of the API you are logging into                                       |
| VCLI_URL           | The address of the API (without the https:// at the start or the :port at the end) |
| VCLI_SETTINGS_PATH | Optional, sets the location for the settings and configuration files               |

As stated before if you have set the VCLI_SETTINGS_PATH before running "init" the files will be located there. If you set it after then you will need to manually move the files to that location before running further commands.

### Creds file mode

When running the "init" command it will ask if you wish to use the Creds file mode, if yes then the tool will read the username and address of the API from the profiles.json file.

You will need to locate and update the profiles with details before calling any of the APIs.

The password will still need to be set as an environmental variable VCLI_PASSWORD.

Doing this provides faster switching between APIs, but does expose the API username and address in the **clear** in the profiles.json file.

### Logging in

After setting up the credentials using either of the methods above and setting the required Profile, next you can simply run the following:

    ./vcli.exe login

If successful it will save a headers.json file which includes the API key that will be used for future calls.

NOTE: The API key is overwritten on each login so switching between profiles will require you to re-login. This may change in the future.

### Change Modes

Simply locate the settings.json file and update the "credsFileMode" to true or false:

    "credsFileMode":true

## API Commands overview

The tools has also been designed to allow you to output the responses to json and yaml formats. These allow you to then modify these responses using tools such as jq.

However, we have found that pairing vcli with "nu shell" provides an excellent user experience for manipulating API responses.

https://www.nushell.sh/

See the nushell section below.

## Get

With "get" pass the endpoint that you want to get data from after the API version number. The response is always json unless you pass the --yaml flag.

    vcli get jobs

### Example

To get all managed servers from VBR the full endpoint is:

    /api/v1/backupInfrastructure/managedServers

You would pass the following

    vcli get backupInfrastructure/managedServers

## Post / Put

With "post" if the endpoint does not need data sent, then simply enter the end of the URL

    vcli post jobs/57b3baab-6237-41bf-add7-db63d41d984c/start

If the endpoint requires data then use the -f flag with the path to a JSON file.

    vcli post jobs -f job_data.json

With "put" the endpoint must have a payload.

    vcli put jobs -f job_data.json

## Job

_new in 0.7.0_

The job command allows you to create jobs using templates.

See this guide for details.

[Job Guide](https://github.com/shapedthought/vcli/blob/master/job.md)

## Utils

_new in 0.5.0_

The utils command allows you to run a number of utility commands.

The current options are:

- VBR Job JSON Converter
- Check Version

### VBR Job JSON Converter

As the VBR API's GET job object is different from the POST job object, this utility allows you to convert a GET job object to a POST job object.

You will need to get a single job from the VBR API for this to work, it will not work on the "Get All Jobs" endpoint response object (something I might look at later).

### Example

Get a single job from the VBR API

    vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c > job.json

Run the utils command

    vcli utils

- Select "VBR Job JSON Converter"
- Enter the path to the job json file
- Enter the path to the output file e.g. job_updated.json
- Modify the job json file as required
- POST the job

```
vcli post/jobs -f job_updated.json
```

Note that VBR PUT job json/objects match the GET json/objects, so you can use the same file for PUT.

### Check Version

This command will check the version of the tool against the latest release on GitHub.

## Using with jq

[jq](https://stedolan.github.io/jq/) offers a lightweight way to navigate JSON date, simply pipe out the responses from the API and use JQ to manipulate the data.

Enterprise Manager Example:

     vcli get jobs?format=entity | jq '.Jobs[].JobInfo.BackupJobInfo.Includes.ObjectInJobs[] | .Name, .ObjectInJobId'

Useful tip: use jq "keys" to see the keys in a json object

    vcli get jobs | jq 'keys'

## Using with Nushell

https://www.nushell.sh/

![nushell](./assets/nushell.png)

Is designed to work around structured data in a better way that normal shells do so it is ideal for manipulating data from APIs.

Installation: https://www.nushell.sh/book/installation.html

Personally I use chocolatey.

Once installed you just need to enter the command:

    nu

Then if you have vcli installed you can set the environmental variables like so:

    let-env VCLI_USERNAME = "username"
    let-env VCLI_PASSWORD = "password"
    let-env VCLI_URL = "192.168.0.123"

Then login

    vcli login

As vcli prints json to the screen you can simply pipe the output into nushell

    vcli get jobs | from json

As most of the APIs hold the actual data under a "data" object you will need to "get" that data

    vcli get jobs | from json | get data

Nushell has a huge amount of methods to explore, filter and transform you data. One of my favorite is being about to pipe out to a different format.

    vcli get jobs | from json | get data | to yaml

You can also save it in a different format, though you need to use the --raw flag.

    vcli get jobs | from json | get data | to yaml | save jobs.yaml --raw

### Nu Modules

Nushell has its own module system which means you can define a series of methods which can then be brought into the shell's scope.

https://www.nushell.sh/book/modules.html

For example, create a file called v.nu and add the following:

    # v.nu

    export def vget [url: string] {
        vcli get $url | from json | get data
    }

You can then import the module by entering:

    use v.nu

You can then use that module like so:

    v vget jobs

Which does the same as if you did the longer version shown above.

Using modules means that you can easily create complex queries very easily and be able to recall them when needed.

It is also possible add the environmental variables to the module making it even easier to get going.

    #v.nu

    export-env {
        let-env VCLI_USERNAME = "username"
        let-env VCLI_PASSWORD = "password"
        let-env VCLI_URL = "192.168.0.123"
    }

    export def vlogin [] {
        vcli login
    }

    export def vget [url: string] {
        vcli get $url | from json | get data
    }

Nushell has it's own HTTP get and post options, which could be turned into a specific module for Veeam, however, vcli has been designed to do all that already.

There is also a plugin system that Nushell provides which might be something I look at in the future.

## Drift Detection & Security Alerting (VBR)

_new in 0.10.0-beta1_

vcli provides drift detection across multiple VBR resource types with security-aware severity classification. For full details see:

- [Drift Detection Guide](docs/drift-detection.md) - All resource types, filtering, CI/CD integration
- [Security Alerting](docs/security-alerting.md) - Value-aware severity, cross-resource validation, severity reference

**Quick reference:**
```bash
# Snapshot resources to establish baseline
vcli repo snapshot --all
vcli repo sobr-snapshot --all
vcli encryption snapshot --all

# Detect drift across all resource types
vcli job diff --all
vcli repo diff --all
vcli repo sobr-diff --all
vcli encryption diff --all
vcli encryption kms-diff --all

# Show only security-relevant drifts
vcli job diff --all --security-only
vcli job diff --all --severity critical
```

## Declarative Job Management (VBR)

_new in 0.9.0-beta1_

vcli now supports declarative, infrastructure-as-code style management for VBR backup jobs with configuration overlays for multi-environment deployments.

### Overview

The declarative workflow allows you to:
- Define backup jobs in YAML files
- Apply environment-specific overlays (prod, dev, staging)
- Preview merged configurations before applying
- Track configurations in version control
- Manage multiple environments from a single base template

### Key Concepts

#### 1. Base Configuration
A base YAML file defines common settings shared across all environments.

#### 2. Overlays
Overlay files contain environment-specific changes that merge with the base.

#### 3. Strategic Merge
vcli uses strategic merge to combine base + overlay:
- Maps are merged recursively (nested objects)
- Arrays are replaced (overlay replaces base)
- Labels/annotations are combined
- Base values preserved unless overridden

#### 4. Environment Configuration
The `vcli.yaml` file manages environment-specific settings and overlay mappings.

### Export Command

Export existing VBR jobs to YAML format for declarative management.

**Full Export (default):**
```bash
# Export complete job configuration (300+ fields)
vcli export 57b3baab-6237-41bf-add7-db63d41d984c -o my-job.yaml

# Export all jobs to directory
vcli export --all -d ./jobs/
```

**Simplified Export (legacy):**
```bash
# Export minimal configuration (20-30 fields)
vcli export 57b3baab-6237-41bf-add7-db63d41d984c -o my-job.yaml --simplified
```

The full export captures all VBR API fields ensuring complete job fidelity. Use simplified export only for basic job configurations.

### Creating Overlays

Overlays contain only the fields you want to override from the base configuration.

**Example Base Configuration:**
```yaml
# base-backup.yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
  labels:
    app: database
    managed-by: vcli
spec:
  type: VSphereBackup
  description: Database backup job
  repository: default-repo
  storage:
    compression: Optimal
    retention:
      type: Days
      quantity: 7
  schedule:
    enabled: true
    daily: "22:00"
    retry:
      enabled: true
      times: 3
      wait: 10
  objects:
    - type: VirtualMachine
      name: db-server
      hostName: 192.168.0.14
```

**Production Overlay:**
```yaml
# prod-overlay.yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
  labels:
    env: production
spec:
  description: Production database backup (30-day retention)
  repository: prod-repo
  storage:
    retention:
      quantity: 30  # Override: 7 days -> 30 days
  schedule:
    daily: "02:00"  # Override: 22:00 -> 02:00
```

**Development Overlay:**
```yaml
# dev-overlay.yaml
apiVersion: vcli.veeam.com/v1
kind: VBRJob
metadata:
  name: database-backup
  labels:
    env: development
spec:
  description: Development database backup (3-day retention)
  repository: dev-repo
  storage:
    retention:
      quantity: 3   # Override: 7 days -> 3 days
  schedule:
    daily: "23:00"  # Override: 22:00 -> 23:00
```

### Job Plan Command

Preview the merged configuration before applying.

**Basic Usage:**
```bash
# Preview base configuration (no overlay)
vcli job plan base-backup.yaml

# Preview with production overlay
vcli job plan base-backup.yaml -o prod-overlay.yaml

# Preview with development overlay
vcli job plan base-backup.yaml -o dev-overlay.yaml

# Show full merged YAML
vcli job plan base-backup.yaml -o prod-overlay.yaml --show-yaml
```

**Output:**
The plan command displays:
- Resource name and type
- Base configuration file
- Overlay file (if used)
- Labels (merged from base + overlay)
- Key configuration fields
- Storage settings
- Schedule settings
- Backup objects list

### Job Apply Command

Apply a job configuration with overlay support.

**Dry-Run Mode (recommended):**
```bash
# Preview what would be applied
vcli job apply base-backup.yaml -o prod-overlay.yaml --dry-run
```

**With Explicit Overlay:**
```bash
# Apply base + production overlay
vcli job apply base-backup.yaml -o prod-overlay.yaml

# Apply base + development overlay
vcli job apply base-backup.yaml -o dev-overlay.yaml
```

**With Environment:**
```bash
# Apply using environment from vcli.yaml
vcli job apply base-backup.yaml --env production
```

**Note:** Use `--dry-run` flag to preview changes before applying them to VBR.

### Environment Configuration (vcli.yaml)

Create a `vcli.yaml` file to manage environment-specific settings.

**Configuration Structure:**
```yaml
# vcli.yaml
currentEnvironment: production
defaultOverlayDir: ./overlays

environments:
  production:
    overlay: prod-overlay.yaml
    profile: vbr-prod
    labels:
      env: production
      managed-by: vcli

  development:
    overlay: dev-overlay.yaml
    profile: vbr-dev
    labels:
      env: development
      managed-by: vcli

  staging:
    overlay: staging-overlay.yaml
    profile: vbr-staging
    labels:
      env: staging
      managed-by: vcli
```

**Configuration File Locations:**
vcli searches for `vcli.yaml` in this order:
1. Path in `VCLI_CONFIG` environment variable
2. Current directory (`./vcli.yaml`)
3. Home directory (`~/.vcli/vcli.yaml`)

**Using Environment Configuration:**
```bash
# Apply using currentEnvironment (production)
vcli job plan base-backup.yaml

# Override with specific environment
vcli job plan base-backup.yaml --env development

# Explicit overlay takes precedence
vcli job plan base-backup.yaml -o custom-overlay.yaml
```

**Overlay Resolution Priority:**
1. `-o/--overlay` flag (highest priority)
2. `--env` flag (looks up in vcli.yaml)
3. `currentEnvironment` from vcli.yaml
4. No overlay (base config only)

### Strategic Merge Behavior

Understanding how overlays merge with base configurations:

**Scalar Values (strings, numbers, booleans):**
Overlay value replaces base value.
```yaml
# Base: quantity: 7
# Overlay: quantity: 30
# Result: quantity: 30
```

**Nested Objects (maps):**
Deep merge - overlays are applied recursively.
```yaml
# Base:
storage:
  compression: Optimal
  retention:
    type: Days
    quantity: 7

# Overlay:
storage:
  retention:
    quantity: 30

# Result:
storage:
  compression: Optimal      # Preserved from base
  retention:
    type: Days              # Preserved from base
    quantity: 30            # Overridden
```

**Arrays:**
Overlay array replaces base array completely.
```yaml
# Base:
objects:
  - name: vm1
  - name: vm2

# Overlay:
objects:
  - name: vm3

# Result:
objects:
  - name: vm3              # Base array replaced
```

**Labels and Annotations:**
Combined (merged) from base and overlay.
```yaml
# Base:
labels:
  app: database
  managed-by: vcli

# Overlay:
labels:
  env: production

# Result:
labels:
  app: database           # From base
  managed-by: vcli        # From base
  env: production         # From overlay
```

### Complete Multi-Environment Example

**Project Structure:**
```
my-backups/
├── vcli.yaml
├── base-backup.yaml
└── overlays/
    ├── prod-overlay.yaml
    ├── dev-overlay.yaml
    └── staging-overlay.yaml
```

**Workflow:**
```bash
# 1. Export existing job as base template
vcli export <job-id> -o base-backup.yaml

# 2. Create overlays directory
mkdir overlays

# 3. Create environment overlays
# (Edit overlays/prod-overlay.yaml, dev-overlay.yaml, etc.)

# 4. Create vcli.yaml configuration
cat > vcli.yaml <<EOF
currentEnvironment: production
defaultOverlayDir: ./overlays
environments:
  production:
    overlay: prod-overlay.yaml
  development:
    overlay: dev-overlay.yaml
  staging:
    overlay: staging-overlay.yaml
EOF

# 5. Preview configurations
vcli job plan base-backup.yaml              # Uses production (currentEnvironment)
vcli job plan base-backup.yaml --env dev    # Preview development
vcli job plan base-backup.yaml --env staging # Preview staging

# 6. Show full merged YAML
vcli job plan base-backup.yaml --show-yaml

# 7. Apply configurations
vcli job apply base-backup.yaml --env production --dry-run
vcli job apply base-backup.yaml --env development --dry-run

# 8. Commit to version control
git add .
git commit -m "Add multi-environment backup configuration"
git push
```

### Best Practices

1. **Keep base templates DRY**: Only include common settings in base files
2. **Small overlays**: Override only what differs per environment
3. **Use labels**: Tag resources with environment, app, and managed-by labels
4. **Version control**: Commit both base and overlays to Git
5. **Preview first**: Always use plan or --dry-run before applying
6. **Consistent naming**: Use clear, descriptive names for base files and overlays
7. **Document changes**: Use Git commit messages to explain configuration changes
8. **Test in dev first**: Apply to development environment before production

### Troubleshooting

**Problem:** Overlay not being applied
- Check overlay resolution priority (explicit -o > --env > currentEnvironment)
- Verify vcli.yaml exists and is in search path
- Confirm environment exists in vcli.yaml

**Problem:** Unexpected merge results
- Remember: arrays are replaced, not merged
- Use `--show-yaml` to see full merged result
- Check that overlay kind matches base kind

**Problem:** Labels not combining
- Ensure both base and overlay use `metadata.labels` field
- Labels should be at same level in both files

### Tips and Tricks

### Replacing a parameter in a JSON file

There is great tool (written in Rust 🦀) called sd which works like sed and allows you to replace strings in a file using string expressions and regex.

For example, if you wanted to replace the name of a job in a JSON file you could do the following:

Check the name using jq

    cat job.json | jq '.name'

The replace the name using sd

     sd '"name": "Backup Job 2"' '"name": "Backup Job 12"' .\job.json

You can also pipe the vcli output directly into sd to update a parameter.

    vcli get jobs/57b3baab-6237-41bf-add7-db63d41d984c | sd '"name": "Backup Job 2"' '"name": "Backup Job 12"' > job.json

I find this useful to make quick changes to a file without having to open it in a text editor.

sd tool: https://crates.io/crates/sd

Install using chocolatey

    choco install sd-cli

Install via Cargo (Rust package manager)

    cargo install sd
