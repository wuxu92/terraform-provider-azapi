package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SharedImageGallery provides resource knowledge for
// Microsoft.Compute/galleries.
//
// Contributing Terraform resource: azurerm_shared_image_gallery.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/shared_image_gallery_resource.go
//     (schema L50-133: name ForceNew, sharing block ForceNew with permission enum +
//     community_gallery eula/prefix/publisher_email/publisher_uri; description Optional;
//     unique_name Computed; timeouts 30m/5m/30m/30m; Create body L164-171,
//     expand sharing L315-367)
//   - terraform-provider-azurerm internal/services/compute/validate/compute.go
//     (SharedImageGalleryName L15-29: `^[A-Za-z0-9._]+$`, max 80) and
//     shared_image_gallery_public_name_prefix.go (SharedImageGalleryPrefix
//     L11: `^[A-Za-z0-9]{5,16}$`)
//   - go-azure-sdk resource-manager/compute/2022-03-03/galleries:
//     model_galleryproperties.go, model_sharingprofile.go, model_communitygalleryinfo.go,
//     model_galleryidentifier.go, constants.go (GallerySharingPermissionTypes).
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; no body ForceNew for them.
//   - sharing and its whole sub-tree are ForceNew, mapped to properties.sharingProfile.*.
//   - sharing.community_gallery.name is Computed but maps to the read-only
//     communityGalleryInfo.publicNames array (no single body path) — not encoded.
type SharedImageGallery struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SharedImageGallery)(nil)

func NewSharedImageGallery() *SharedImageGallery {
	return &SharedImageGallery{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/galleries",
			ApiVersions:  []string{"2022-03-03"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.sharingProfile.permissions"},
				{PropertyPath: "properties.sharingProfile.communityGalleryInfo.eula"},
				{PropertyPath: "properties.sharingProfile.communityGalleryInfo.publicNamePrefix"},
				{PropertyPath: "properties.sharingProfile.communityGalleryInfo.publisherContact"},
				{PropertyPath: "properties.sharingProfile.communityGalleryInfo.publisherUri"},
			},
			StringRules: []azwise.StringRule{
				// Resource name — validate.SharedImageGalleryName.
				{
					Regex:     `^[A-Za-z0-9._]+$`,
					MaxLength: 80,
					Message:   "can only contain alphanumeric, full stops and underscores (up to 80 characters)",
				},
				// sharing.permission → properties.sharingProfile.permissions.
				{
					PropertyPath:  "properties.sharingProfile.permissions",
					AllowedValues: []string{"Community", "Groups", "Private"},
					Message:       "must be one of Community, Groups, or Private",
				},
				// community_gallery.prefix → publicNamePrefix — validate.SharedImageGalleryPrefix.
				{
					PropertyPath: "properties.sharingProfile.communityGalleryInfo.publicNamePrefix",
					Regex:        `^[A-Za-z0-9]{5,16}$`,
					MinLength:    5,
					MaxLength:    16,
					Message:      "must be 5-16 alphanumeric characters",
				},
				// community_gallery.eula → eula (StringIsNotEmpty).
				{
					PropertyPath: "properties.sharingProfile.communityGalleryInfo.eula",
					MinLength:    1,
					Message:      "must not be empty",
				},
				// community_gallery.publisher_email → publisherContact (StringIsNotEmpty).
				{
					PropertyPath: "properties.sharingProfile.communityGalleryInfo.publisherContact",
					MinLength:    1,
					Message:      "must not be empty",
				},
				// community_gallery.publisher_uri → publisherUri (StringIsNotEmpty).
				{
					PropertyPath: "properties.sharingProfile.communityGalleryInfo.publisherUri",
					MinLength:    1,
					Message:      "must not be empty",
				},
			},
			// unique_name → properties.identifier.uniqueName is server-assigned (read-only).
			ComputedFields: []string{"properties.identifier.uniqueName"},
		},
	}
}

func init() { azwise.Register(NewSharedImageGallery()) }
