package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SharedImage provides resource knowledge for
// Microsoft.Compute/galleries/images.
//
// Contributing Terraform resource: azurerm_shared_image.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/shared_image_resource.go
//     (schema L35-287: architecture/os_type/hyper_v_generation/identifier/eula/
//     purchase_plan/privacy_statement_uri/specialized ForceNew; recommended vCPU/memory
//     IntBetween; disk_types_not_allowed enum; feature toggles; timeouts 30m/5m/30m/30m;
//     Create body L316-349, expand helpers L636-813; CustomizeDiff conditional ForceNew
//     for end_of_life_date L282-286)
//   - terraform-provider-azurerm internal/services/compute/validate/compute.go
//     (SharedImageName L31-45: `^[A-Za-z0-9._-]+$`, max 80;
//     SharedImageIdentifierAttribute L47-66: max length + no trailing dot)
//   - go-azure-sdk resource-manager/compute/2022-03-03/galleryimages:
//     model_galleryimageproperties.go, model_galleryimageidentifier.go,
//     model_imagepurchaseplan.go, model_recommendedmachineconfiguration.go,
//     model_resourcerange.go, model_disallowed.go, constants.go (Architecture,
//     OperatingSystemTypes, OperatingSystemStateTypes, HyperVGeneration).
//
// Notes:
//   - name/gallery_name/resource_group_name/location are envelope/parent-owned.
//   - specialized (bool) is mapped by AzureRM to the required properties.osState enum
//     (Specialized/Generalized); encoded as a body enum + required field.
//   - The trusted_launch/confidential_vm/accelerated_network/hibernation/
//     disk_controller feature toggles (all ForceNew) and their mutual ConflictsWith
//     are expanded into properties.features[] array elements (Name/Value pairs) with no
//     single body path, so their ForceNew and relational constraints are not encoded.
type SharedImage struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SharedImage)(nil)

func NewSharedImage() *SharedImage {
	return &SharedImage{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/galleries/images",
			ApiVersions:  []string{"2022-03-03"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.architecture"},
				{PropertyPath: "properties.osType"},
				{PropertyPath: "properties.osState"}, // from specialized (bool)
				{PropertyPath: "properties.hyperVGeneration"},
				{PropertyPath: "properties.identifier.publisher"},
				{PropertyPath: "properties.identifier.offer"},
				{PropertyPath: "properties.identifier.sku"},
				{PropertyPath: "properties.eula"},
				{PropertyPath: "properties.privacyStatementUri"},
				{PropertyPath: "properties.purchasePlan.name"},
				{PropertyPath: "properties.purchasePlan.publisher"},
				{PropertyPath: "properties.purchasePlan.product"},
				// end_of_life_date (properties.endOfLifeDate) is ForceNew only when
				// cleared (CustomizeDiff ForceNewIfChange) — not encoded as unconditional.
			},
			StringRules: []azwise.StringRule{
				// Resource name — validate.SharedImageName.
				{
					Regex:     `^[A-Za-z0-9._-]+$`,
					MaxLength: 80,
					Message:   "can only contain alphanumeric, full stops, dashes and underscores (up to 80 characters)",
				},
				{
					PropertyPath:  "properties.architecture",
					AllowedValues: []string{"Arm64", "x64"},
					Message:       "must be one of Arm64 or x64",
				},
				{
					PropertyPath:  "properties.osType",
					AllowedValues: []string{"Linux", "Windows"},
					Message:       "must be one of Linux or Windows",
				},
				{
					PropertyPath:  "properties.osState",
					AllowedValues: []string{"Generalized", "Specialized"},
					Message:       "must be one of Generalized or Specialized",
				},
				{
					PropertyPath:  "properties.hyperVGeneration",
					AllowedValues: []string{"V1", "V2"},
					Message:       "must be one of V1 or V2",
				},
				// identifier.{publisher,offer,sku} — validate.SharedImageIdentifierAttribute.
				{
					PropertyPath: "properties.identifier.publisher",
					MaxLength:    128,
					Message:      "can be up to 128 characters",
				},
				{
					PropertyPath: "properties.identifier.offer",
					MaxLength:    64,
					Message:      "can be up to 64 characters",
				},
				{
					PropertyPath: "properties.identifier.sku",
					MaxLength:    64,
					Message:      "can be up to 64 characters",
				},
				// disk_types_not_allowed → properties.disallowed.diskTypes[] element enum.
				{
					PropertyPath:  "properties.disallowed.diskTypes[*]",
					AllowedValues: []string{"Standard_LRS", "Premium_LRS"},
					Message:       "must be one of Standard_LRS or Premium_LRS",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.recommended.vCPUs.max",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(80)),
					Message:      "max_recommended_vcpu_count must be between 1 and 80",
				},
				{
					PropertyPath: "properties.recommended.vCPUs.min",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(80)),
					Message:      "min_recommended_vcpu_count must be between 1 and 80",
				},
				{
					PropertyPath: "properties.recommended.memory.max",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(640)),
					Message:      "max_recommended_memory_in_gb must be between 1 and 640",
				},
				{
					PropertyPath: "properties.recommended.memory.min",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(640)),
					Message:      "min_recommended_memory_in_gb must be between 1 and 640",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.architecture", Value: "x64"},
				{PropertyPath: "properties.hyperVGeneration", Value: "V1"},
			},
			RequiredFields: []string{
				"properties.osType",
				"properties.osState",
				"properties.identifier.publisher",
				"properties.identifier.offer",
				"properties.identifier.sku",
			},
		},
	}
}

func init() { azwise.Register(NewSharedImage()) }
