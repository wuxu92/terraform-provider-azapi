package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DiskEncryptionSet provides resource knowledge for Microsoft.Compute/diskEncryptionSets.
//
// Contributing Terraform resource: azurerm_disk_encryption_set.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/disk_encryption_set_resource.go
//     (schema L33-132, Create body L134-214)
//   - terraform-provider-azurerm internal/services/compute/validate/disk_encryption_set_name.go
//   - go-azure-sdk resource-manager/compute/2022-03-02/diskencryptionsets:
//     model_encryptionsetproperties.go, model_keyfordiskencryptionset.go, constants.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; name is ForceNew but is not an
//     ARM-body property, so it is not emitted as a body ForceNew rule.
//   - key_vault_key_id / managed_hsm_key_id are resolved to a key URL and written to
//     properties.activeKey.keyUrl (KeyForDiskEncryptionSet.keyUrl); that URL is the
//     required create field.
//   - identity type change userAssigned/systemAssignedUserAssigned -> systemAssigned is
//     ForceNew (CustomizeDiff, disk_encryption_set_resource.go L104-109); value-conditional,
//     left as a note.
//   - federated_client_id uses validation.IsUUID — a generic semantic validator that maps to
//     properties.federatedClientId; not expressible as a StringRule, left as a note for an
//     azapin UUID validator.
type DiskEncryptionSet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DiskEncryptionSet)(nil)

// NewDiskEncryptionSet returns knowledge for the diskEncryptionSets resource.
func NewDiskEncryptionSet() *DiskEncryptionSet {
	return &DiskEncryptionSet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/diskEncryptionSets",
			ApiVersions:  []string{"2022-03-02"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.encryptionType"},
			},
			StringRules: []azwise.StringRule{
				// Resource name validation (empty PropertyPath validates the name).
				{Regex: `^[^_\W][\w-._]{0,78}\w$`, MinLength: 1, MaxLength: 80, Message: "must be 1-80 characters, contain only a-z, A-Z, 0-9, _, . and -, and not start with _ or end with - or ."},
				{PropertyPath: "properties.encryptionType", AllowedValues: []string{
					"ConfidentialVmEncryptedWithCustomerKey",
					"EncryptionAtRestWithCustomerKey",
					"EncryptionAtRestWithPlatformAndCustomerKeys",
				}},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.rotationToLatestKeyVersionEnabled", Value: false},
				{PropertyPath: "properties.encryptionType", Value: "EncryptionAtRestWithCustomerKey"},
			},
			RequiredFields: []string{
				"properties.activeKey.keyUrl",
			},
			// Read-only server-populated properties returned by GET but never authored.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.lastKeyRotationTimestamp",
				"properties.autoKeyRotationError",
				"properties.previousKeys",
			},
		},
	}
}

func init() { azwise.Register(NewDiskEncryptionSet()) }
