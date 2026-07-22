package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Image provides resource knowledge for Microsoft.Compute/images.
//
// Contributing Terraform resource: azurerm_image.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/image_resource.go
//     (schema L27-223, Create body L226-294, expand L371-454)
//   - go-azure-sdk resource-manager/compute/2022-03-01/images:
//     model_imageproperties.go, model_imagestorageprofile.go, model_imageosdisk.go,
//     model_imagedatadisk.go, constants.go (enums).
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; name is ForceNew but is not an
//     ARM-body property, so it is not emitted as a body ForceNew rule.
//   - os_disk is a MaxItems:1 block (properties.storageProfile.osDisk) marked ForceNew at the
//     block level, so the whole object path is emitted; its enum sub-fields resolve through
//     the single object and are validated below.
//   - data_disk is an array (properties.storageProfile.dataDisks[]); its per-element enum and
//     ForceNew constraints are array-element paths that azwise cannot lower — left as a note.
//   - os_disk.storage_type is Required only when the os_disk block is present (block itself is
//     optional and conflicts with source_virtual_machine_id) — value-conditional, not emitted
//     as a global RequiredField.
type Image struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Image)(nil)

// NewImage returns knowledge for the images resource.
func NewImage() *Image {
	return &Image{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/images",
			ApiVersions:  []string{"2022-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.hyperVGeneration"},
				{PropertyPath: "properties.storageProfile.zoneResilient"},
				{PropertyPath: "properties.storageProfile.osDisk"},
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.hyperVGeneration", AllowedValues: []string{
					"V1",
					"V2",
				}},
				{PropertyPath: "properties.storageProfile.osDisk.osType", AllowedValues: []string{
					"Linux",
					"Windows",
				}},
				{PropertyPath: "properties.storageProfile.osDisk.osState", AllowedValues: []string{
					"Generalized",
					"Specialized",
				}},
				{PropertyPath: "properties.storageProfile.osDisk.caching", AllowedValues: []string{
					"None",
					"ReadOnly",
					"ReadWrite",
				}},
				{PropertyPath: "properties.storageProfile.osDisk.storageAccountType", AllowedValues: []string{
					"Premium_LRS",
					"PremiumV2_LRS",
					"Premium_ZRS",
					"Standard_LRS",
					"StandardSSD_LRS",
					"StandardSSD_ZRS",
					"UltraSSD_LRS",
				}},
			},
			// source_virtual_machine_id conflicts with zone_resilient, os_disk and data_disk.
			ConflictsWith: []azwise.RelationalRule{
				{Paths: []string{"properties.sourceVirtualMachine.id", "properties.storageProfile.zoneResilient"}, Message: "source_virtual_machine_id conflicts with zone_resilient"},
				{Paths: []string{"properties.sourceVirtualMachine.id", "properties.storageProfile.osDisk"}, Message: "source_virtual_machine_id conflicts with os_disk"},
				{Paths: []string{"properties.sourceVirtualMachine.id", "properties.storageProfile.dataDisks"}, Message: "source_virtual_machine_id conflicts with data_disk"},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.hyperVGeneration", Value: "V1"},
				{PropertyPath: "properties.storageProfile.zoneResilient", Value: false},
				{PropertyPath: "properties.storageProfile.osDisk.caching", Value: "None"},
			},
			// Read-only server-populated properties returned by GET but never authored.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewImage()) }
