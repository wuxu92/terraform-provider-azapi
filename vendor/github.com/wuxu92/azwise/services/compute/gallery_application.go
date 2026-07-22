package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// GalleryApplication provides resource knowledge for
// Microsoft.Compute/galleries/applications.
//
// Contributing Terraform resource: azurerm_gallery_application.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/gallery_application_resource.go
//     (schema L52-108: name/gallery_id/supported_os_type ForceNew;
//     description/end_of_life_date/eula/privacy_statement_uri/release_note_uri Optional;
//     Create body L155-182; timeout 30m; CustomizeDiff conditional ForceNew L350-375)
//   - terraform-provider-azurerm internal/services/compute/validate/compute.go
//     (GalleryApplicationName L78-91: `^[A-Za-z0-9._-]+$`, max 80)
//   - go-azure-sdk resource-manager/compute/2022-03-03/galleryapplications:
//     model_galleryapplicationproperties.go, constants.go (OperatingSystemTypes).
//
// Notes:
//   - name/gallery_id/location are envelope/parent-owned.
//   - end_of_life_date/privacy_statement_uri/release_note_uri are ForceNew only when
//     cleared (CustomizeDiff) — encoded as ForceNew with that caveat noted.
type GalleryApplication struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*GalleryApplication)(nil)

func NewGalleryApplication() *GalleryApplication {
	return &GalleryApplication{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/galleries/applications",
			ApiVersions:  []string{"2022-03-03"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.supportedOSType"},
				// The following are ForceNew only when their value is removed
				// (CustomizeDiff ForceNew-on-clear), not on every change.
				{PropertyPath: "properties.endOfLifeDate"},
				{PropertyPath: "properties.privacyStatementUri"},
				{PropertyPath: "properties.releaseNoteUri"},
			},
			StringRules: []azwise.StringRule{
				// Resource name — validate.GalleryApplicationName.
				{
					Regex:     `^[A-Za-z0-9._-]+$`,
					MaxLength: 80,
					Message:   "can only contain alphanumeric, full stops, dashes and underscores (up to 80 characters)",
				},
				{
					PropertyPath:  "properties.supportedOSType",
					AllowedValues: []string{"Linux", "Windows"},
					Message:       "must be one of Linux or Windows",
				},
			},
			RequiredFields: []string{
				"properties.supportedOSType",
			},
		},
	}
}

func init() { azwise.Register(NewGalleryApplication()) }
