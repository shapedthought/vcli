# KMS Server Configuration Examples

This directory contains example VBR Key Management Server (KMS) configurations for enterprise encryption key management.

## Available Examples

### azure-key-vault.yaml
Azure Key Vault integration:
- Enterprise key management
- Certificate-based authentication
- Health monitoring enabled
- Production compliance settings

**Use case:** Enterprise encryption using Azure Key Vault for centralized key management and compliance.

## Important Notes

**⚠️ KMS Servers are Update-Only**

KMS servers **cannot be created** via vcli apply. They must be configured in the VBR console first.

vcli can **update** KMS server settings:
- Description
- Monitoring settings
- Health check intervals
- Alert configurations

**Most KMS configuration is done in VBR console:**
- KMS provider selection
- Connection settings
- Authentication/credentials
- Certificate configuration
- Firewall rules

**Workflow:**
1. Configure KMS server in VBR console
2. Test connection and permissions
3. Export to YAML: `vcli encryption kms-export "KMS Name" -o kms.yaml`
4. Edit description/monitoring settings if needed
5. Apply updates: `vcli encryption kms-apply kms.yaml`

## Usage

### Snapshot Existing KMS Server

```bash
# Snapshot single KMS server
vcli encryption kms-snapshot "Azure Key Vault Production"

# Snapshot all KMS servers
vcli encryption kms-snapshot --all
```

### Export KMS Configuration

```bash
# Export by name
vcli encryption kms-export "Azure Key Vault Production" -o kms.yaml

# Export all KMS servers
vcli encryption kms-export --all -d ./all-kms/
```

### Apply KMS Updates

```bash
# Preview changes first (recommended)
vcli encryption kms-apply azure-key-vault.yaml --dry-run

# Apply updates
vcli encryption kms-apply azure-key-vault.yaml

# Apply with environment overlay
vcli encryption kms-apply azure-key-vault.yaml -o ../overlays/prod/kms-overlay.yaml
```

### Detect Configuration Drift

```bash
# Check single KMS server
vcli encryption kms-diff "Azure Key Vault Production"

# Check all KMS servers
vcli encryption kms-diff --all

# Check for security-relevant drift only
vcli encryption kms-diff --all --security-only
```

## Supported KMS Providers

VBR supports multiple KMS providers:

### Azure Key Vault
- **Type:** AzureKeyVault
- **Authentication:** Service Principal + Certificate
- **Use case:** Azure-native environments

### AWS KMS
- **Type:** AWSKMS
- **Authentication:** IAM credentials
- **Use case:** AWS-native environments

### Enterprise KMS
- **Type:** KMIP (various vendors)
- **Authentication:** Certificate-based
- **Use case:** On-premises enterprise key management

## Customization

### Modifying Examples for Your Environment

1. **Update KMS name** to match your VBR configuration
2. **Adjust description** for clarity and documentation
3. **Configure monitoring** based on operational requirements
4. **Set appropriate labels** for organization

### Updateable Fields (Limited)

Via vcli apply, you can typically update:
- Description
- Monitoring enabled/disabled
- Health check intervals
- Alert settings

### Non-Updateable Fields

These must be configured in VBR console:
- KMS provider type
- Connection endpoints
- Authentication credentials
- Certificates
- Firewall/network settings
- Service principal/IAM settings

## Azure Key Vault Setup

**Prerequisites:**
1. Azure Key Vault created with appropriate SKU
2. Service principal created with key permissions
3. Certificate created and added to VBR
4. Network access configured (firewall rules)

**Required Permissions:**
- Get key
- List keys
- Unwrap key
- Wrap key

**VBR Console Configuration:**
1. Backup Infrastructure → Managed Servers → KMS
2. Add KMS Server → Azure Key Vault
3. Provide vault URL and tenant ID
4. Select certificate for authentication
5. Test connection

## Best Practices

1. **Use certificate-based authentication** for better security
2. **Enable monitoring and alerts** for KMS availability
3. **Test KMS connection** after any configuration changes
4. **Regular drift detection** to catch unauthorized changes
5. **Document firewall rules** required for KMS access
6. **Maintain certificate validity** and renewal schedules
7. **Version control** KMS configurations in Git
8. **Keep credentials secure** (never commit to Git)

## Security Considerations

### Network Access
- Configure firewall to allow VBR → KMS communication
- Use private endpoints where possible (Azure)
- Restrict KMS access to known VBR server IPs

### Authentication
- Use service principals with minimal required permissions
- Rotate certificates before expiration
- Use separate service principals per environment

### Monitoring
- Enable health checks to detect connectivity issues
- Alert on KMS connection failures
- Monitor key usage and access logs

## Troubleshooting

### Connection Failures
1. Verify firewall rules allow VBR → KMS
2. Check certificate validity and trust
3. Confirm service principal permissions
4. Test DNS resolution to KMS endpoint

### Authentication Errors
1. Verify certificate is installed and trusted
2. Check service principal has required permissions
3. Confirm tenant ID is correct
4. Review KMS audit logs for access denials

### Certificate Expiration
1. Monitor certificate expiration dates
2. Renew certificates before expiration
3. Update certificate in VBR console after renewal
4. Test KMS connection after certificate update

## Exit Codes

KMS apply commands return specific exit codes:

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Continue |
| 1 | Error (API failure) | Check logs and retry |
| 6 | KMS server not found | Create in VBR console first |

## See Also

- [Declarative Mode Guide](../../docs/declarative-mode.md) - Complete KMS management guide
- [State Management Guide](../../docs/state-management.md) - Understanding state and snapshots
- [Drift Detection Guide](../../docs/drift-detection.md) - Configuration monitoring
- [Security Alerting](../../docs/security-alerting.md) - Security-aware drift classification
- [Troubleshooting Guide](../../docs/troubleshooting.md) - KMS connection issues
