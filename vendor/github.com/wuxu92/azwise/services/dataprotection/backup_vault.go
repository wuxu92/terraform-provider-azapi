package dataprotection

import (
	"strings"
	"time"

	"github.com/wuxu92/azwise"
)

// BackupVault provides resource knowledge for
// Microsoft.DataProtection/backupVaults (azurerm_data_protection_backup_vault).
//
// The value-conditional ForceNew cases AzureRM enforces via CustomizeDiff
// (cross_region_restore_enabled cannot be disabled, immutability cannot leave
// Locked, soft_delete cannot leave AlwaysOn) are represented in the CheckForceNew
// override below, since the declarative ForceNew list cannot express "only when
// the value changes from X".
//
// This file also encodes the customer-managed-key knowledge from the separate
// azurerm_data_protection_backup_vault_customer_managed_key resource. That TF
// resource has no distinct ARM type — it PATCHes
// properties.securitySettings.encryptionSettings onto the SAME backupVaults body —
// so its constraints are merged here rather than into a new file.
//
// Sources:
//   - AzureRM internal/services/dataprotection/data_protection_backup_vault_resource.go
//     :40-45   (timeouts create/update/delete 30m, read 5m)
//     :52-118  (schema: name regex, datastore_type, redundancy, soft_delete,
//     immutability, retention_duration_in_days, cross_region_restore_enabled)
//     :120-145 (CustomizeDiff: ForceNewIfChange cross_region_restore/immutability/
//     soft_delete; GeoRedundant requirement for cross_region_restore)
//     :179-214 (expand: storageSettings, securitySettings.softDelete/immutability,
//     featureSettings.crossRegionRestore)
//   - AzureRM internal/services/dataprotection/data_protection_backup_vault_customer_managed_key_resource.go
//     :52-67   (schema: data_protection_backup_vault_id, key_vault_key_id)
//     :117-126 (expand: securitySettings.encryptionSettings state/keyVaultProperties/kekIdentity)
//   - go-azure-sdk resource-manager/dataprotection/2025-07-01/backupvaultresources:
//     model_backupvault.go, model_storagesetting.go, model_securitysettings.go,
//     model_softdeletesettings.go, model_immutabilitysettings.go,
//     model_featuresettings.go, model_crossregionrestoresettings.go,
//     model_encryptionsettings.go, model_cmkkeyvaultproperties.go,
//     model_cmkkekidentity.go, constants.go (enum PossibleValuesFor* value sets)
//
// Not encoded (deliberate):
//   - datastore_type (storageSettings[*].datastoreType) and redundancy
//     (storageSettings[*].type) are ForceNew enum fields, but storageSettings is an
//     ARRAY ([]StorageSetting) and azwise/azapin cannot lower or resolve a rule
//     through an array element, so their ForceNew + enum rules are skipped.
//   - key_vault_key_id maps to
//     properties.securitySettings.encryptionSettings.keyVaultProperties.keyUri and is
//     validated by a semantic Key Vault nested-item-ID validator, not a declarative
//     regex/enum — left to a residual customizer validator.
//   - identity (top-level SystemAssigned/UserAssigned) is an envelope field, not a
//     body property under properties.
type BackupVault struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BackupVault)(nil)

// CheckForceNew extends BaseKnowledge with AzureRM's value-conditional
// replacements (data_protection_backup_vault_resource.go:123-134): a backup vault
// must be replaced when cross-region restore is turned off, when immutability
// leaves the Locked state, or when soft-delete leaves the AlwaysOn state.
func (s *BackupVault) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	oldCRR := azwise.ExtractStringValue(oldBody, "properties.featureSettings.crossRegionRestoreSettings.state")
	newCRR := azwise.ExtractStringValue(newBody, "properties.featureSettings.crossRegionRestoreSettings.state")
	if strings.EqualFold(oldCRR, "Enabled") && !strings.EqualFold(newCRR, "Enabled") {
		return true
	}

	oldImm := azwise.ExtractStringValue(oldBody, "properties.securitySettings.immutabilitySettings.state")
	newImm := azwise.ExtractStringValue(newBody, "properties.securitySettings.immutabilitySettings.state")
	if strings.EqualFold(oldImm, "Locked") && !strings.EqualFold(newImm, "Locked") {
		return true
	}

	oldSD := azwise.ExtractStringValue(oldBody, "properties.securitySettings.softDeleteSettings.state")
	newSD := azwise.ExtractStringValue(newBody, "properties.securitySettings.softDeleteSettings.state")
	return strings.EqualFold(oldSD, "AlwaysOn") && !strings.EqualFold(newSD, "AlwaysOn")
}

// NewBackupVault returns knowledge for the backupVaults resource.
func NewBackupVault() *BackupVault {
	return &BackupVault{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataProtection/backupVaults",
			ApiVersions:  []string{"2025-07-01"},
			// The vault itself supports soft-delete of its protected data, but the
			// vault resource is deleted synchronously; no purge/recover flow.
			SoftDelete: false,
			// name and resource_group_name are envelope-owned. location
			// (commonschema.Location) is ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: 2-50 chars, letters, numbers, hyphens.
					Regex:     `^[-a-zA-Z0-9]{2,50}$`,
					MinLength: 2,
					MaxLength: 50,
					Message:   "DataProtection BackupVault name must be 2-50 characters long and contain only letters, numbers and hyphens",
				},
				{
					PropertyPath:  "properties.securitySettings.softDeleteSettings.state",
					AllowedValues: []string{"AlwaysOn", "Off", "On"},
					Message:       "must be one of AlwaysOn, Off or On",
				},
				{
					PropertyPath:  "properties.securitySettings.immutabilitySettings.state",
					AllowedValues: []string{"Disabled", "Locked", "Unlocked"},
					Message:       "must be one of Disabled, Locked or Unlocked",
				},
				{
					PropertyPath:  "properties.featureSettings.crossRegionRestoreSettings.state",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Disabled or Enabled",
				},
				// Customer-managed-key knowledge (from the CMK sub-resource).
				{
					PropertyPath:  "properties.securitySettings.encryptionSettings.state",
					AllowedValues: []string{"Disabled", "Enabled", "Inconsistent"},
					Message:       "must be one of Disabled, Enabled or Inconsistent",
				},
				{
					PropertyPath:  "properties.securitySettings.encryptionSettings.infrastructureEncryption",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.securitySettings.encryptionSettings.kekIdentity.identityType",
					AllowedValues: []string{"SystemAssigned", "UserAssigned"},
					Message:       "must be SystemAssigned or UserAssigned",
				},
			},
			FloatRules: []azwise.FloatRule{
				{
					PropertyPath: "properties.securitySettings.softDeleteSettings.retentionDurationInDays",
					MinValue:     azwise.Ptr(float64(14)),
					MaxValue:     azwise.Ptr(float64(180)),
					Message:      "soft-delete retention must be between 14 and 180 days",
				},
			},
			// Response-only properties Azure populates in the GET body.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.resourceMoveState",
				"properties.resourceMoveDetails",
				"properties.secureScore",
				"properties.bcdrSecurityLevel",
				"properties.isVaultProtectedByResourceGuard",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.securitySettings.softDeleteSettings.state", Value: "On"},
				{PropertyPath: "properties.securitySettings.softDeleteSettings.retentionDurationInDays", Value: float64(14)},
				{PropertyPath: "properties.securitySettings.immutabilitySettings.state", Value: "Disabled"},
			},
			RequiredFields: []string{
				"properties.storageSettings",
			},
		},
	}
}

func init() { azwise.Register(NewBackupVault()) }
