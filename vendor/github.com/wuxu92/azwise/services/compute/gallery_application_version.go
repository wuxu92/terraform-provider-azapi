package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// GalleryApplicationVersion provides resource knowledge for
// Microsoft.Compute/galleries/applications/versions.
//
// Contributing Terraform resource: azurerm_gallery_application_version.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/gallery_application_version_resource.go
//     (schema L72-208: config_file/package_file/manage_action(install/remove/update)/
//     source(media_link/default_configuration_link) ForceNew; enable_health_check/
//     exclude_from_latest defaults; manage_action StringLenBetween(1,4096);
//     target_region Required with regional_replica_count IntBetween(1,10) + storage
//     enum; Create body L255-291, expand helpers L500-571; CustomizeDiff conditional
//     ForceNew for end_of_life_date L485-498; SafetyProfile.AllowDeletion hardcoded true)
//   - terraform-provider-azurerm internal/services/compute/validate/compute.go
//     (GalleryApplicationVersionName L93-101: `1.2.3` | latest | recent)
//   - go-azure-sdk resource-manager/compute/2022-03-03/galleryapplicationversions:
//     model_galleryapplicationversionproperties.go,
//     model_galleryapplicationversionpublishingprofile.go, model_userartifactmanage.go,
//     model_userartifactsource.go, model_userartifactsettings.go, model_targetregion.go,
//     model_galleryartifactsafetyprofilebase.go, constants.go (StorageAccountType).
//
// Notes:
//   - name/gallery_application_id/location are envelope/parent-owned.
//   - target_region.regional_replica_count (IntBetween 1-10) lives under the
//     targetRegions[] array element (publishingProfile.targetRegions[*].regionalReplicaCount);
//     azwise IntRules do not resolve through array elements, so it is not encoded.
//   - target_region.storage_account_type is encoded as a [*] StringRule.
type GalleryApplicationVersion struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*GalleryApplicationVersion)(nil)

func NewGalleryApplicationVersion() *GalleryApplicationVersion {
	return &GalleryApplicationVersion{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/galleries/applications/versions",
			ApiVersions:  []string{"2022-03-03"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.publishingProfile.settings.configFileName"},  // config_file
				{PropertyPath: "properties.publishingProfile.settings.packageFileName"}, // package_file
				{PropertyPath: "properties.publishingProfile.manageActions.install"},
				{PropertyPath: "properties.publishingProfile.manageActions.remove"},
				{PropertyPath: "properties.publishingProfile.manageActions.update"},
				{PropertyPath: "properties.publishingProfile.source.mediaLink"},
				{PropertyPath: "properties.publishingProfile.source.defaultConfigurationLink"},
				// end_of_life_date (publishingProfile.endOfLifeDate) is ForceNew only when
				// cleared (CustomizeDiff ForceNewIfChange) — not encoded as unconditional.
			},
			StringRules: []azwise.StringRule{
				// Resource name — validate.GalleryApplicationVersionName.
				{
					Regex:   `^([0-9]{1,10}\.[0-9]{1,10}\.[0-9]{1,10}|latest|recent)$`,
					Message: "must be in the format `1.2.3`, `latest`, or `recent`",
				},
				{
					PropertyPath: "properties.publishingProfile.manageActions.install",
					MinLength:    1,
					MaxLength:    4096,
					Message:      "must be between 1 and 4096 characters",
				},
				{
					PropertyPath: "properties.publishingProfile.manageActions.remove",
					MinLength:    1,
					MaxLength:    4096,
					Message:      "must be between 1 and 4096 characters",
				},
				{
					PropertyPath: "properties.publishingProfile.manageActions.update",
					MinLength:    1,
					MaxLength:    4096,
					Message:      "must be between 1 and 4096 characters",
				},
				// target_region.storage_account_type → targetRegions[] element enum.
				{
					PropertyPath:  "properties.publishingProfile.targetRegions[*].storageAccountType",
					AllowedValues: []string{"Premium_LRS", "Standard_LRS", "Standard_ZRS"},
					Message:       "must be one of Premium_LRS, Standard_LRS, or Standard_ZRS",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publishingProfile.enableHealthCheck", Value: false},
				{PropertyPath: "properties.publishingProfile.excludeFromLatest", Value: false},
				// AzureRM hardcodes allowDeletionOfReplicatedLocations = true.
				{PropertyPath: "properties.safetyProfile.allowDeletionOfReplicatedLocations", Value: true},
			},
			RequiredFields: []string{
				"properties.publishingProfile.manageActions.install",
				"properties.publishingProfile.manageActions.remove",
				"properties.publishingProfile.source.mediaLink",
				"properties.publishingProfile.targetRegions",
			},
		},
	}
}

func init() { azwise.Register(NewGalleryApplicationVersion()) }
