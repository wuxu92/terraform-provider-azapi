package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RecoveryServicesVault provides resource knowledge for
// Microsoft.RecoveryServices/vaults.
//
// Contributing TF resource:
//   - azurerm_recovery_services_vault
//
// Sources:
//   - AzureRM internal/services/recoveryservices/recovery_services_vault_resource.go
//     :49-53 (timeouts create 120m / read 5m / update 60m / delete 30m).
//     :57-61 (name Required+ForceNew, RecoveryServicesVaultName validator).
//     :111-115 (immutability enum), :121-127 (sku enum), :130-137 (storage_mode_type
//     enum + default GeoRedundant), :101-104 (public_network_access default true),
//     :141-144 (cross_region_restore_enabled default false), :80-93 (encryption block).
//     :192-198 (conditional ForceNew for cross_region_restore_enabled / immutability).
//   - internal/services/recoveryservices/validate/recovery_services_vault_name.go:12-17
//     (name regex ^[a-zA-Z][-a-zA-Z0-9]{1,49}$, 2-50 chars).
//   - go-azure-sdk resource-manager/recoveryservices/2025-08-01/vaults:
//     model_vault.go (Sku, Properties), model_vaultproperties.go,
//     model_vaultpropertiesredundancysettings.go, model_vaultpropertiesencryption.go,
//     model_cmkkekidentity.go, model_sku.go, constants.go (enum values).
//
// Not encoded (deliberate):
//   - ForceNew on `name` is the resource-name envelope field, not a body path.
//   - `cross_region_restore_enabled` -> properties.redundancySettings.crossRegionRestore
//     and `immutability` -> properties.securitySettings.immutabilitySettings.state are
//     only conditionally/directionally ForceNew (cross-region restore replaces only when
//     switching Enabled->Disabled; immutability replaces only when the current state is
//     Locked). Encoding them as unconditional ForceNew would wrongly flag valid in-place
//     changes, so they are documented here but not emitted as ForceNewRule.
//   - `classic_vmware_replication_enabled` (schema ForceNew) is not a vault body property;
//     it drives a separate replicationVaultSetting sub-API, so no body path applies.
type RecoveryServicesVault struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RecoveryServicesVault)(nil)

// NewRecoveryServicesVault returns knowledge for the vaults resource.
func NewRecoveryServicesVault() *RecoveryServicesVault {
	return &RecoveryServicesVault{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults",
			ApiVersions:  []string{"2023-02-01", "2024-04-01", "2025-08-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 120 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"sku.name",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.redundancySettings.standardTierStorageRedundancy", Value: "GeoRedundant"},
				{PropertyPath: "properties.redundancySettings.crossRegionRestore", Value: "Disabled"},
				{PropertyPath: "properties.encryption.kekIdentity.useSystemAssignedIdentity", Value: true},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name.
					Regex:     "^[a-zA-Z][-a-zA-Z0-9]{1,49}$",
					MinLength: 2,
					MaxLength: 50,
					Message:   "Recovery Services Vault name must be 2-50 characters long, start with a letter, and contain only letters, numbers and hyphens",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"RS0", "Standard"},
					Message:       "sku.name must be one of RS0 or Standard",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "properties.publicNetworkAccess must be Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.redundancySettings.standardTierStorageRedundancy",
					AllowedValues: []string{"GeoRedundant", "LocallyRedundant", "ZoneRedundant"},
					Message:       "properties.redundancySettings.standardTierStorageRedundancy must be one of GeoRedundant, LocallyRedundant or ZoneRedundant",
				},
				{
					PropertyPath:  "properties.redundancySettings.crossRegionRestore",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "properties.redundancySettings.crossRegionRestore must be Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.securitySettings.immutabilitySettings.state",
					AllowedValues: []string{"Disabled", "Locked", "Unlocked"},
					Message:       "properties.securitySettings.immutabilitySettings.state must be one of Disabled, Locked or Unlocked",
				},
				{
					PropertyPath:  "properties.encryption.infrastructureEncryption",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "properties.encryption.infrastructureEncryption must be Enabled or Disabled",
				},
			},
		},
	}
}

func init() { azwise.Register(NewRecoveryServicesVault()) }
