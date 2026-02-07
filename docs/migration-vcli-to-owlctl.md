# Migration Guide: vcli to owlctl

This guide covers the steps to migrate from `vcli` to `owlctl`.

## 1. Rename the binary

Replace `vcli` (or `vcli.exe`) with `owlctl` (or `owlctl.exe`). Download the new binary from the [Releases](https://github.com/shapedthought/owlctl/releases) page.

## 2. Update environment variables

| Old | New |
|-----|-----|
| `VCLI_USERNAME` | `OWLCTL_USERNAME` |
| `VCLI_PASSWORD` | `OWLCTL_PASSWORD` |
| `VCLI_URL` | `OWLCTL_URL` |
| `VCLI_SETTINGS_PATH` | `OWLCTL_SETTINGS_PATH` |
| `VCLI_TOKEN` | `OWLCTL_TOKEN` |
| `VCLI_CONFIG` | `OWLCTL_CONFIG` |
| `VCLI_FILE_KEY` | `OWLCTL_FILE_KEY` |

Update these in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) and in any CI/CD variable groups.

## 3. Rename the config directory

```bash
mv ~/.vcli ~/.owlctl
```

## 4. Rename the config file

If you use a project-level config file:

```bash
mv vcli.yaml owlctl.yaml
```

## 5. Update apiVersion in YAML specs

In all your exported job/repo/SOBR/KMS YAML files, update the `apiVersion` field:

```yaml
# Old
apiVersion: vcli.veeam.com/v1

# New
apiVersion: owlctl.veeam.com/v1
```

A quick sed one-liner:

```bash
find . -name "*.yaml" -exec sed -i 's|vcli.veeam.com/v1|owlctl.veeam.com/v1|g' {} +
```

## 6. Re-login

The keyring service name changed, so stored tokens won't carry over. Run:

```bash
owlctl login
```

## 7. Update CI/CD pipelines

- Change binary references from `./vcli` to `./owlctl`
- Update build commands from `go build -o vcli` to `go build -o owlctl`
- Update environment variable names (see step 2)
- Update any Git clone URLs from `shapedthought/vcli` to `shapedthought/owlctl`
