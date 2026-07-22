package elasticsan

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ElasticSanVolume provides resource knowledge for
// Microsoft.ElasticSan/elasticSans/volumeGroups/volumes.
//
// Mirrors azurerm_elastic_san_volume.
//
// Sources:
//   - terraform-provider-azurerm internal/services/elasticsan/elastic_san_volume_resource.go
//     schema Arguments (61-113: name ForceNew, size_in_gib IntBetween(1,65536),
//     create_source block ForceNew with source_id (resource-ID validation.Any) and
//     source_type StringInSlice), Attributes (115-137: target_iqn/target_portal_hostname/
//     target_portal_port/volume_id computed), CustomizeDiff (139-150: size_in_gib may only
//     grow), Expand/Flatten create source (299-322); timeouts Create/Update/Delete 30m, Read 5m.
//   - internal/services/elasticsan/validate/elastic_san_volume_name.go (name regex 3-63).
//   - go-azure-sdk resource-manager/elasticsan/2023-01-01/volumes:
//     VolumeProperties (sizeGiB required; creationData; storageTarget/volumeId/
//     provisioningState read-only), SourceCreationData{CreateSource,SourceId},
//     IscsiTargetInfo (targetIqn/targetPortalHostname/targetPortalPort read-only),
//     constants.go (VolumeCreateOption Disk/DiskRestorePoint/DiskSnapshot/None/VolumeSnapshot),
//     id_volume.go Segments (type segment casing "volumeGroups"/"volumes").
type ElasticSanVolume struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ElasticSanVolume)(nil)

// NewElasticSanVolume returns knowledge for the volumes resource.
func NewElasticSanVolume() *ElasticSanVolume {
	return &ElasticSanVolume{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ElasticSan/elasticSans/volumeGroups/volumes",
			ApiVersions:  []string{"2023-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// volume_group_id is the parent reference (envelope), not a body field.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.creationData"},
				{PropertyPath: "properties.creationData.sourceId"},
				{PropertyPath: "properties.creationData.createSource"},
			},
			// sizeGiB is required (no omitempty in the ARM model).
			RequiredFields: []string{
				"properties.sizeGiB",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM ElasticSanVolumeName:
					// 3-63 chars, lowercase alphanumeric plus '_'/'-', start/end alphanumeric.
					Regex:     `^[a-z0-9][a-z0-9_-]{1,61}[a-z0-9]$`,
					MinLength: 3,
					MaxLength: 63,
					Message:   "must be 3-63 characters, lowercase letters, numbers, underscores and hyphens only, and start/end with a lowercase letter or number",
				},
				{
					// AzureRM restricts to Disk/DiskRestorePoint/DiskSnapshot/VolumeSnapshot,
					// but the ARM SDK enum also allows None; AzAPI sends raw ARM values.
					PropertyPath:  "properties.creationData.createSource",
					AllowedValues: []string{"Disk", "DiskRestorePoint", "DiskSnapshot", "None", "VolumeSnapshot"},
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.sizeGiB",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(65536)),
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.storageTarget.targetIqn",
				"properties.storageTarget.targetPortalHostname",
				"properties.storageTarget.targetPortalPort",
				"properties.volumeId",
				"properties.provisioningState",
			},
			// NOTE: create_source.source_id uses validation.Any over four resource-ID
			// validators (managed disk / restore point / disk snapshot / volume snapshot);
			// this is a semantic resource-ID check with no single declarative form and is
			// not expressed here. CustomizeDiff also enforces that sizeGiB may only grow.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewElasticSanVolume()) }
