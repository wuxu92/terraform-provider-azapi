package azurestackhci

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StackHCIMarketplaceGalleryImage provides resource knowledge for
// Microsoft.AzureStackHCI/marketplaceGalleryImages.
//
// Mirrors azurerm_stack_hci_marketplace_gallery_image. name / resource_group_name / location
// live on the operational envelope; location is surfaced here as ForceNew (commonschema.Location).
// Every argument on this resource is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_marketplace_gallery_image_resource.go:63-142
//     (schema: ForceNew, validators, enums, MaxItems)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_marketplace_gallery_image_resource.go:172-192,308-320
//     (create/expand mapping: identifier{offer,publisher,sku}, osType, hyperVGeneration,
//     version.name, containerId, extendedLocation)
//   - go-azure-sdk resource-manager/azurestackhci/2024-01-01/marketplacegalleryimages:
//     model_marketplacegalleryimageproperties.go, model_galleryimageidentifier.go,
//     model_galleryimageversion.go, constants.go:93-96 (HyperVGeneration V1/V2),
//     134-137 (OperatingSystemTypes Linux/Windows)
//
// Not encoded (deliberate):
//   - custom_location_id and storage_path_id carry resource-ID semantic validators
//     (customlocations / storagecontainers) that belong in an azapin customizer
//     (validators.AzureResourceID) attached to extendedLocation.name / properties.containerId,
//     not a StringRule.
//   - identifier is a MaxItems:1 block that AzureRM expands into a single
//     properties.identifier object (not an array), so its offer/publisher/sku paths are direct
//     and their Required flags are captured in RequiredFields below.
type StackHCIMarketplaceGalleryImage struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StackHCIMarketplaceGalleryImage)(nil)

// NewStackHCIMarketplaceGalleryImage returns knowledge for the Azure Stack HCI
// marketplaceGalleryImages resource.
func NewStackHCIMarketplaceGalleryImage() *StackHCIMarketplaceGalleryImage {
	return &StackHCIMarketplaceGalleryImage{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureStackHCI/marketplaceGalleryImages",
			ApiVersions:  []string{"2024-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "properties.hyperVGeneration"},
				{PropertyPath: "properties.identifier"},
				{PropertyPath: "properties.osType"},
				{PropertyPath: "properties.version"},
				{PropertyPath: "properties.containerId"},
			},
			RequiredFields: []string{
				"properties.hyperVGeneration",
				"properties.osType",
				"properties.version.name",
				"properties.identifier.offer",
				"properties.identifier.publisher",
				"properties.identifier.sku",
				// AzureRM always sends the custom location as the extended location.
				"extendedLocation.name",
				"extendedLocation.type",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 3 * time.Hour,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: begin/end alphanumeric, 2-80 chars, alphanumeric + - . _ inside.
					Regex:   `^[a-zA-Z0-9][\-\.\_a-zA-Z0-9]{0,78}[a-zA-Z0-9]$`,
					Message: "name must begin and end with an alphanumeric character, be between 2 and 80 characters in length and can only contain alphanumeric characters, hyphens, periods or underscores",
				},
				{
					PropertyPath:  "properties.hyperVGeneration",
					AllowedValues: []string{"V1", "V2"},
					Message:       "must be V1 or V2",
				},
				{
					PropertyPath:  "properties.osType",
					AllowedValues: []string{"Linux", "Windows"},
					Message:       "must be Linux or Windows",
				},
			},
		},
	}
}

func init() { azwise.Register(NewStackHCIMarketplaceGalleryImage()) }
