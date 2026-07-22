package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SharedImageVersion provides resource knowledge for
// Microsoft.Compute/galleries/images/versions.
//
// Contributing Terraform resource: azurerm_shared_image_version.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/shared_image_version_resource.go
//     (schema L33-203: target_region Required; blob_uri/storage_account_id/
//     os_disk_snapshot_id/managed_image_id/replication_mode/
//     deletion_of_replicated_locations_enabled ForceNew; ExactlyOneOf source group;
//     blob_uri<->storage_account_id RequiredWith; timeouts 30m/5m/30m/30m;
//     Create body L232-283; CustomizeDiff conditional ForceNew for end_of_life_date L198-202)
//   - terraform-provider-azurerm internal/services/compute/validate/compute.go
//     (SharedImageVersionName L68-76: `1.2.3` | latest | recent)
//   - go-azure-sdk resource-manager/compute/2023-07-03/galleryimageversions:
//     model_galleryimageversionproperties.go, model_galleryartifactpublishingprofilebase.go,
//     model_galleryimageversionstorageprofile.go, model_gallerydiskimage.go,
//     model_gallerydiskimagesource.go, model_galleryartifactversionfullsource.go,
//     model_galleryimageversionsafetyprofile.go, model_targetregion.go,
//     constants.go (ReplicationMode, StorageAccountType).
//
// Notes:
//   - name/gallery_name/image_name/resource_group_name/location are envelope/parent-owned.
//   - managed_image_id expands one-to-many: a VM id → storageProfile.source.virtualMachineId,
//     else → storageProfile.source.id; both encoded ForceNew.
//   - target_region is a list; its regional_replica_count / disk_encryption_set_id /
//     exclude_from_latest / storage_account_type live under
//     properties.publishingProfile.targetRegions[] array elements (StringRule uses [*]).
type SharedImageVersion struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SharedImageVersion)(nil)

func NewSharedImageVersion() *SharedImageVersion {
	return &SharedImageVersion{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/galleries/images/versions",
			ApiVersions:  []string{"2023-07-03"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.storageProfile.osDiskImage.source.uri"},              // blob_uri
				{PropertyPath: "properties.storageProfile.osDiskImage.source.storageAccountId"}, // storage_account_id
				{PropertyPath: "properties.storageProfile.osDiskImage.source.id"},               // os_disk_snapshot_id
				{PropertyPath: "properties.storageProfile.source.id"},                           // managed_image_id (image/snapshot id)
				{PropertyPath: "properties.storageProfile.source.virtualMachineId"},             // managed_image_id (vm id)
				{PropertyPath: "properties.publishingProfile.replicationMode"},
				{PropertyPath: "properties.safetyProfile.allowDeletionOfReplicatedLocations"}, // deletion_of_replicated_locations_enabled
				// end_of_life_date (publishingProfile.endOfLifeDate) is ForceNew only when
				// cleared (CustomizeDiff ForceNewIfChange) — not encoded as unconditional.
			},
			StringRules: []azwise.StringRule{
				// Resource name — validate.SharedImageVersionName.
				{
					Regex:   `^([0-9]{1,10}\.[0-9]{1,10}\.[0-9]{1,10}|latest|recent)$`,
					Message: "must be in the format `1.2.3`, `latest`, or `recent`",
				},
				{
					PropertyPath:  "properties.publishingProfile.replicationMode",
					AllowedValues: []string{"Full", "Shallow"},
					Message:       "must be one of Full or Shallow",
				},
				// target_region.storage_account_type → targetRegions[] element enum.
				{
					PropertyPath:  "properties.publishingProfile.targetRegions[*].storageAccountType",
					AllowedValues: []string{"Premium_LRS", "Standard_LRS", "Standard_ZRS"},
					Message:       "must be one of Premium_LRS, Standard_LRS, or Standard_ZRS",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publishingProfile.replicationMode", Value: "Full"},
				{PropertyPath: "properties.publishingProfile.excludeFromLatest", Value: false},
				{PropertyPath: "properties.safetyProfile.allowDeletionOfReplicatedLocations", Value: false},
			},
			RequiredFields: []string{
				"properties.publishingProfile.targetRegions",
				"properties.storageProfile",
			},
			// blob_uri / os_disk_snapshot_id / managed_image_id — ExactlyOneOf source.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.storageProfile.osDiskImage.source.uri",
						"properties.storageProfile.osDiskImage.source.id",
						"properties.storageProfile.source.id",
					},
					Message: "exactly one of blob_uri, os_disk_snapshot_id, or managed_image_id must be set",
				},
			},
			// blob_uri <-> storage_account_id are mutually required.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.storageProfile.osDiskImage.source.uri",
						"properties.storageProfile.osDiskImage.source.storageAccountId",
					},
					Message: "storage_account_id is required when blob_uri is set",
				},
				{
					Paths: []string{
						"properties.storageProfile.osDiskImage.source.storageAccountId",
						"properties.storageProfile.osDiskImage.source.uri",
					},
					Message: "blob_uri is required when storage_account_id is set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSharedImageVersion()) }
