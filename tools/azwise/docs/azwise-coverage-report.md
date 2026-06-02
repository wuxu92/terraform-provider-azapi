# Azwise Coverage Comparison Report

Generated from `terraform-provider-azurerm` source. Compares what azurerm encodes
per resource against what `internal/azure/azwise/` currently covers.

## Currently Covered by Azwise

### Microsoft.KeyVault/vaults
**Azwise:** Name validation (3-24 chars, alphanumeric+hyphens), soft-delete retention (IntRule 7-90 days), 6 property rules (sku.name, networkAcls.defaultAction, etc.), timeouts 30m/5m/30m/30m.
**AzureRM has:** 1 ForceNew, 0 Sensitive, 0 ForceNewIfChange, CustomizeDiff=True, SoftDelete=True

### Microsoft.Storage/storageAccounts
**Azwise:** Name validation (3-24 lowercase alphanumeric), sku.name / kind / accessTier property rules, conditional ForceNew for zone-redundant SKU migration, sensitive access keys, timeouts 60m/5m/60m/60m.
**AzureRM has:** 10 ForceNew, 6 Sensitive, 1 ForceNewIfChange, CustomizeDiff=True, SoftDelete=True

## Priority Resources NOT Yet Covered

Ranked by complexity score = ForceNew + ConditionalForceNew×3 + Sensitive + CustomizeDiff×3 + SoftDelete×2.

### 1. Microsoft.ContainerService/managedClusters (score: 84)

| Attribute | Value |
|---|---|
| ForceNew fields | 23 |
| ForceNewIfChange | 13 |
| Timeouts (C/R/U/D) | 90m/5m/90m/90m |
| Sensitive fields | 19 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 12 regex patterns |

### 2. Microsoft.DocumentDB/databaseAccounts (score: 37)

| Attribute | Value |
|---|---|
| ForceNew fields | 16 |
| ForceNewIfChange | 2 |
| Timeouts (C/R/U/D) | 180m/5m/180m/300m |
| Sensitive fields | 12 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 3 regex patterns |

### 3. Microsoft.Compute/virtualMachines (Linux) (score: 20)

| Attribute | Value |
|---|---|
| ForceNew fields | 15 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | 45m/5m/45m/45m |
| Sensitive fields | 2 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 23 regex patterns |

### 4. Microsoft.Compute/disks (score: 20)

| Attribute | Value |
|---|---|
| ForceNew fields | 14 |
| ForceNewIfChange | 1 |
| Timeouts (C/R/U/D) | 30m/5m/30m/30m |
| Sensitive fields | 0 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 23 regex patterns |

### 5. Microsoft.Sql/servers/databases (score: 19)

| Attribute | Value |
|---|---|
| ForceNew fields | 7 |
| ForceNewIfChange | 2 |
| Timeouts (C/R/U/D) | 60m/5m/60m/60m |
| Sensitive fields | 3 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 7 regex patterns |

### 6. Microsoft.Compute/virtualMachines (Windows) (score: 19)

| Attribute | Value |
|---|---|
| ForceNew fields | 17 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | 45m/5m/45m/45m |
| Sensitive fields | 2 |
| CustomizeDiff | False |
| Soft-delete | False |
| Naming validators | 23 regex patterns |

### 7. Microsoft.Network/publicIPAddresses (score: 12)

| Attribute | Value |
|---|---|
| ForceNew fields | 6 |
| ForceNewIfChange | 1 |
| Timeouts (C/R/U/D) | 30m/5m/30m/30m |
| Sensitive fields | 0 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 26 regex patterns |

### 8. Microsoft.KeyVault/vaults/keys (score: 12)

| Attribute | Value |
|---|---|
| ForceNew fields | 4 |
| ForceNewIfChange | 1 |
| Timeouts (C/R/U/D) | 30m/30m/30m/30m |
| Sensitive fields | 0 |
| CustomizeDiff | True |
| Soft-delete | True |
| Naming validators | 3 regex patterns |

### 9. Microsoft.Network/applicationGateways (score: 9)

| Attribute | Value |
|---|---|
| ForceNew fields | 1 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | 90m/5m/90m/90m |
| Sensitive fields | 5 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 26 regex patterns |

### 10. Microsoft.Sql/servers/elasticPools (score: 8)

| Attribute | Value |
|---|---|
| ForceNew fields | 2 |
| ForceNewIfChange | 1 |
| Timeouts (C/R/U/D) | 30m/5m/30m/30m |
| Sensitive fields | 0 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 7 regex patterns |

### 11. Microsoft.Sql/servers (score: 8)

| Attribute | Value |
|---|---|
| ForceNew fields | 3 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | 60m/5m/60m/60m |
| Sensitive fields | 2 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 7 regex patterns |

### 12. Microsoft.Web/sites/functions (Linux) (score: 6)

| Attribute | Value |
|---|---|
| ForceNew fields | 1 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | defaults |
| Sensitive fields | 2 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 11 regex patterns |

### 13. Microsoft.ContainerRegistry/registries (score: 6)

| Attribute | Value |
|---|---|
| ForceNew fields | 2 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | 30m/5m/30m/30m |
| Sensitive fields | 1 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 12 regex patterns |

### 14. Microsoft.Web/sites (Windows) (score: 5)

| Attribute | Value |
|---|---|
| ForceNew fields | 1 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | defaults |
| Sensitive fields | 1 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 11 regex patterns |

### 15. Microsoft.Web/sites (Linux) (score: 5)

| Attribute | Value |
|---|---|
| ForceNew fields | 1 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | defaults |
| Sensitive fields | 1 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 11 regex patterns |

### 16. Microsoft.Web/serverfarms (score: 5)

| Attribute | Value |
|---|---|
| ForceNew fields | 2 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | defaults |
| Sensitive fields | 0 |
| CustomizeDiff | True |
| Soft-delete | False |
| Naming validators | 11 regex patterns |

### 17. Microsoft.KeyVault/vaults/secrets (score: 4)

| Attribute | Value |
|---|---|
| ForceNew fields | 1 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | 30m/30m/30m/30m |
| Sensitive fields | 1 |
| CustomizeDiff | False |
| Soft-delete | True |
| Naming validators | 3 regex patterns |

### 18. Microsoft.Network/virtualNetworks (score: 1)

| Attribute | Value |
|---|---|
| ForceNew fields | 1 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | 30m/5m/30m/30m |
| Sensitive fields | 0 |
| CustomizeDiff | False |
| Soft-delete | False |
| Naming validators | 26 regex patterns |

### 19. Microsoft.Network/networkSecurityGroups (score: 1)

| Attribute | Value |
|---|---|
| ForceNew fields | 1 |
| ForceNewIfChange | 0 |
| Timeouts (C/R/U/D) | 30m/5m/30m/30m |
| Sensitive fields | 0 |
| CustomizeDiff | False |
| Soft-delete | False |
| Naming validators | 26 regex patterns |
