// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package mssqlmanagedinstance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedInstanceEncryptionProtector provides resource knowledge for
// Microsoft.Sql/managedInstances/encryptionProtector
// (Terraform azurerm_mssql_managed_instance_transparent_data_encryption).
//
// The ARM resource name is the constant "current". Body is
// {properties:{serverKeyType, serverKeyName, autoRotationEnabled, uri}} against the
// go-azure-sdk managedinstanceencryptionprotectors.ManagedInstanceEncryptionProtector
// model.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssqlmanagedinstance/mssql_managed_instance_transparent_data_encryption_resource.go:48-117
//     (Schema: key_vault_key_id, auto_rotation_enabled default false)
//   - .../mssql_managed_instance_transparent_data_encryption_resource.go:122-213 (CreateUpdate → properties)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/managedinstanceencryptionprotectors:
//     model_managedinstanceencryptionprotectorproperties.go:6-12, constants.go:12-24
//     (ServerKeyType: AzureKeyVault, ServiceManaged)
//
// Intentionally not encoded (documented, not emitted):
//   - key_vault_key_id / managed_hsm_key_id do NOT map to a single encryptionProtector
//     body field. AzureRM first writes a separate
//     Microsoft.Sql/managedInstances/keys resource (SDK managedinstancekeys) named
//     "{vault}_{key}_{version}" from the key URI, then sets
//     properties.serverKeyName to that derived name and
//     properties.serverKeyType = AzureKeyVault. So the KV key validator
//     (keyvault.ValidateNestedItemID) has no single ARM path on THIS resource and
//     is left to the customizer / the managedinstances/keys knowledge file.
//   - managed_hsm_key_id ConflictsWith key_vault_key_id (4.x only): both collapse
//     into the same serverKeyName derivation, so it is not an ARM-body
//     ConflictsWith and is not emitted as a RelationalRule.
//   - properties.serverKeyType default: AzureRM sends ServiceManaged when no key is
//     given, AzureKeyVault when a key is given — a conditional default, so no static
//     DefaultValue is emitted (documented instead).
//   - managed_instance_id (parent) is envelope-owned RequiresReplace.
type ManagedInstanceEncryptionProtector struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedInstanceEncryptionProtector)(nil)

// NewManagedInstanceEncryptionProtector returns knowledge for
// Microsoft.Sql/managedInstances/encryptionProtector.
func NewManagedInstanceEncryptionProtector() *ManagedInstanceEncryptionProtector {
	return &ManagedInstanceEncryptionProtector{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/managedInstances/encryptionProtector",
			ApiVersions:  []string{"2023-08-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{},
			StringRules: []azwise.StringRule{
				// properties.serverKeyType — full SDK enum set.
				{
					PropertyPath:  "properties.serverKeyType",
					AllowedValues: []string{"AzureKeyVault", "ServiceManaged"},
					Message:       "must be one of AzureKeyVault or ServiceManaged",
				},
			},
			IntRules:        []azwise.IntRule{},
			FloatRules:      []azwise.FloatRule{},
			ArrayRules:      []azwise.ArrayRule{},
			SensitiveFields: []string{},
			ComputedFields:  []string{},
			DefaultValues: []azwise.DefaultValue{
				// auto_rotation_enabled default false.
				{PropertyPath: "properties.autoRotationEnabled", Value: false},
			},
			// serverKeyType is Required (json:"serverKeyType" without omitempty).
			RequiredFields: []string{"properties.serverKeyType"},
		},
	}
}

func init() { azwise.Register(NewManagedInstanceEncryptionProtector()) }
