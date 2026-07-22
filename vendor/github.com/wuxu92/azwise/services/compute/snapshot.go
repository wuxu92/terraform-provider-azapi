package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Snapshot provides resource knowledge for Microsoft.Compute/snapshots.
//
// Contributing Terraform resource: azurerm_snapshot.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/snapshot_resource.go
//     (schema L30-145, Create body L148-228)
//   - terraform-provider-azurerm internal/services/compute/validate/snapshot_name.go
//   - go-azure-sdk resource-manager/compute/2022-03-02/snapshots:
//     model_snapshotproperties.go, model_creationdata.go, constants.go (enums).
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; name is ForceNew but is not an
//     ARM-body property, so it is not emitted as a body ForceNew rule.
//   - encryption_settings (properties.encryptionSettingsCollection) is ForceNew only when
//     removed once set (CustomizeDiff ForceNewIfChange, snapshot_resource.go L140-144);
//     value-conditional, left as a note.
//   - create_option is Required and updatable (not ForceNew) for a snapshot.
//   - trusted_launch_enabled is Computed-only here (read from securityProfile.securityType).
type Snapshot struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Snapshot)(nil)

// NewSnapshot returns knowledge for the snapshots resource.
func NewSnapshot() *Snapshot {
	return &Snapshot{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/snapshots",
			ApiVersions:  []string{"2022-03-02"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.incremental"},
				{PropertyPath: "properties.creationData.sourceUri"},
				{PropertyPath: "properties.creationData.sourceResourceId"},
				{PropertyPath: "properties.creationData.storageAccountId"},
			},
			StringRules: []azwise.StringRule{
				// Resource name validation (empty PropertyPath validates the name).
				{Regex: `^[A-Za-z0-9_-]+$`, MaxLength: 80, Message: "must be up to 80 alphanumeric characters, underscores and hyphens"},
				{PropertyPath: "properties.creationData.createOption", AllowedValues: []string{
					"Attach",
					"Copy",
					"CopyStart",
					"Empty",
					"FromImage",
					"Import",
					"ImportSecure",
					"Restore",
					"Upload",
					"UploadPreparedSecure",
				}},
				{PropertyPath: "properties.networkAccessPolicy", AllowedValues: []string{
					"AllowAll",
					"AllowPrivate",
					"DenyAll",
				}},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.incremental", Value: false},
				{PropertyPath: "properties.networkAccessPolicy", Value: "AllowAll"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			RequiredFields: []string{
				"properties.creationData.createOption",
			},
			// Read-only server-populated properties returned by GET but never authored.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.diskState",
				"properties.diskSizeBytes",
				"properties.timeCreated",
				"properties.uniqueId",
				"properties.completionPercent",
				"properties.copyCompletionError",
			},
		},
	}
}

func init() { azwise.Register(NewSnapshot()) }
