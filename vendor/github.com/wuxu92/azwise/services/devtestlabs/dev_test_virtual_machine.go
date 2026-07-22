package devtestlabs

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevTestVirtualMachine provides resource knowledge for Microsoft.DevTestLab/labs/virtualMachines.
//
// MERGED OS-variant file. AzureRM splits this single ARM type into two typed TF resources that
// share the same ARM body (virtualmachines.LabVirtualMachine / LabVirtualMachineProperties):
//   - azurerm_dev_test_linux_virtual_machine   (osType = "Linux")
//   - azurerm_dev_test_windows_virtual_machine (osType = "Windows")
//
// Only knowledge universal to BOTH bodies is unioned here. OS-specific differences are noted
// but NOT emitted as rules (they would corrupt validation for the other variant):
//   - name length: Linux max 62, Windows max 15 (validate.DevTestVirtualMachineName). The shared
//     regex is emitted; the differing MaxLength is documented only.
//   - ssh_key (properties.sshKey): Linux-only, ForceNew. Windows never sets it. Included as a
//     ForceNew rule since it only ever diffs on Linux bodies (never present on Windows).
//   - password (properties.password): Linux optional, Windows required — ForceNew in both, so
//     ForceNew is unioned, but it is NOT added to RequiredFields (not universal).
//
// Sources:
//   - terraform-provider-azurerm internal/services/devtestlabs/dev_test_linux_virtual_machine_resource.go
//     (schema L51-156, Create body L205-232)
//   - terraform-provider-azurerm internal/services/devtestlabs/dev_test_windows_virtual_machine_resource.go
//     (schema L51-149, Create body L197-222)
//   - terraform-provider-azurerm internal/services/devtestlabs/helpers.go
//     (gallery_image_reference schema L89-119, inbound_nat_rule L14-44)
//   - terraform-provider-azurerm internal/services/devtestlabs/validate/devtest.go
//     (DevTestVirtualMachineName L22-54)
//   - go-azure-sdk resource-manager/devtestlab/2018-09-15/virtualmachines:
//     model_labvirtualmachineproperties.go, model_galleryimagereference.go, id_virtualmachine.go.
//
// Notes:
//   - name/lab_name/location/resource_group_name are envelope-owned (name + parent labs/{labName})
//     and ForceNew; only the shared name regex is emitted (empty PropertyPath).
//   - storage_type maps to properties.storageType (enum Standard/Premium, ForceNew).
//   - gallery_image_reference (Required) maps to properties.galleryImageReference; its offer/
//     publisher/sku/version sub-fields are Required + ForceNew in both variants.
type DevTestVirtualMachine struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevTestVirtualMachine)(nil)

func NewDevTestVirtualMachine() *DevTestVirtualMachine {
	return &DevTestVirtualMachine{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevTestLab/labs/virtualMachines",
			ApiVersions:  []string{"2018-09-15"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.size"},
				{PropertyPath: "properties.userName"},
				{PropertyPath: "properties.storageType"},
				{PropertyPath: "properties.labSubnetName"},
				{PropertyPath: "properties.labVirtualNetworkId"},
				{PropertyPath: "properties.disallowPublicIpAddress"},
				{PropertyPath: "properties.password"},
				{PropertyPath: "properties.sshKey"}, // Linux-only, but only ever diffs on Linux bodies
				{PropertyPath: "properties.galleryImageReference.offer"},
				{PropertyPath: "properties.galleryImageReference.publisher"},
				{PropertyPath: "properties.galleryImageReference.sku"},
				{PropertyPath: "properties.galleryImageReference.version"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Shared name regex (empty PropertyPath validates the name). Length limit differs
				// by OS (Linux 62 / Windows 15) so it is documented, not emitted.
				{Regex: `^([a-zA-Z0-9]{1})([a-zA-Z0-9-]+)([a-zA-Z0-9]{1})$`, Message: "may contain letters, numbers, or '-', must begin and end with a letter or number, and cannot be all numbers"},
				{PropertyPath: "properties.storageType", AllowedValues: []string{"Standard", "Premium"}},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.allowClaim", Value: true},
			},
			RequiredFields: []string{
				"properties.size",
				"properties.userName",
				"properties.storageType",
				"properties.labSubnetName",
				"properties.labVirtualNetworkId",
				"properties.osType",
				"properties.galleryImageReference",
				"properties.galleryImageReference.offer",
				"properties.galleryImageReference.publisher",
				"properties.galleryImageReference.sku",
				"properties.galleryImageReference.version",
			},
			ComputedFields: []string{
				"properties.fqdn",
				"properties.uniqueIdentifier",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewDevTestVirtualMachine()) }
