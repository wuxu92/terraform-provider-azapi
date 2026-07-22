package elasticsan

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ElasticSanVolumeGroup provides resource knowledge for
// Microsoft.ElasticSan/elasticSans/volumeGroups.
//
// Mirrors azurerm_elastic_san_volume_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/elasticsan/elastic_san_volume_group_resource.go
//     schema Arguments (67-150: name ForceNew, encryption_type StringInSlice default
//     EncryptionAtRestWithPlatformKey, encryption block CMK settings, network_rule block
//     with action StringInSlice default Allow, protocol_type StringInSlice(Iscsi) default
//     Iscsi, identity), CustomizeDiff (152-172: encryption block requires
//     encryption_type == EncryptionAtRestWithCustomerManagedKey), Expand/Flatten
//     encryption + network rules (371-473); timeouts Create/Update/Delete 30m, Read 5m.
//   - internal/services/elasticsan/validate/elastic_san_volume_group_name.go (name regex 3-63).
//   - go-azure-sdk resource-manager/elasticsan/2023-01-01/volumegroups:
//     VolumeGroupProperties (encryption/encryptionProperties/networkAcls/protocolType/
//     provisioningState), EncryptionProperties{Identity,KeyVaultProperties},
//     KeyVaultProperties (currentVersionedKey*/lastKeyRotationTimestamp read-only),
//     constants.go (EncryptionType, StorageTargetType Iscsi/None, Action Allow),
//     id_volumegroup.go Segments (type segment casing "volumeGroups").
type ElasticSanVolumeGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ElasticSanVolumeGroup)(nil)

// NewElasticSanVolumeGroup returns knowledge for the volumeGroups resource.
func NewElasticSanVolumeGroup() *ElasticSanVolumeGroup {
	return &ElasticSanVolumeGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ElasticSan/elasticSans/volumeGroups",
			ApiVersions:  []string{"2023-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// elastic_san_id is the parent reference (envelope), not a body field.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM ElasticSanVolumeGroupName:
					// 3-63 chars, lowercase alphanumeric plus '_'/'-', start/end alphanumeric.
					Regex:     `^[a-z0-9][a-z0-9_-]{1,61}[a-z0-9]$`,
					MinLength: 3,
					MaxLength: 63,
					Message:   "must be 3-63 characters, lowercase letters, numbers, underscores and hyphens only, and start/end with a lowercase letter or number",
				},
				{
					PropertyPath:  "properties.encryption",
					AllowedValues: []string{"EncryptionAtRestWithCustomerManagedKey", "EncryptionAtRestWithPlatformKey"},
				},
				{
					// AzureRM restricts to Iscsi, but the ARM SDK enum also allows None;
					// AzAPI sends raw ARM values so both are permitted.
					PropertyPath:  "properties.protocolType",
					AllowedValues: []string{"Iscsi", "None"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.encryption", Value: "EncryptionAtRestWithPlatformKey"},
				{PropertyPath: "properties.protocolType", Value: "Iscsi"},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.encryptionProperties.keyVaultProperties.currentVersionedKeyExpirationTimestamp",
				"properties.encryptionProperties.keyVaultProperties.currentVersionedKeyIdentifier",
				"properties.encryptionProperties.keyVaultProperties.lastKeyRotationTimestamp",
				"properties.provisioningState",
			},
			// NOTE: several constraints cannot be represented declaratively:
			//   - CustomizeDiff couples the encryptionProperties block to
			//     encryption == EncryptionAtRestWithCustomerManagedKey (value-conditional
			//     required-with, not expressible as a RelationalRule).
			//   - key_vault_key_id expands one-to-many into keyVaultProperties.keyName /
			//     keyVersion / keyVaultUri (composite, no single ARM body path).
			//   - network_rule.action StringInSlice(Allow) and its default Allow live at
			//     properties.networkAcls.virtualNetworkRules[*].action, an array-element
			//     path that the generator cannot lower — skipped.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewElasticSanVolumeGroup()) }
