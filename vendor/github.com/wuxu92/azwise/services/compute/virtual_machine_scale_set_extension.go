package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineScaleSetExtension provides resource knowledge for
// Microsoft.Compute/virtualMachineScaleSets/extensions.
//
// Contributing TF resource:
//   - azurerm_virtual_machine_scale_set_extension (internal/services/compute/virtual_machine_scale_set_extension_resource.go)
//
// Sources:
//   - virtual_machine_scale_set_extension_resource.go:28-133 (schema, timeouts, ForceNew)
//   - go-azure-sdk compute/2024-03-01/virtualmachinescalesetextensions/model_virtualmachinescalesetextensionproperties.go
type VirtualMachineScaleSetExtension struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineScaleSetExtension)(nil)

func NewVirtualMachineScaleSetExtension() *VirtualMachineScaleSetExtension {
	return &VirtualMachineScaleSetExtension{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/virtualMachineScaleSets/extensions",
			ApiVersions:  []string{"2024-03-01", "2025-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.publisher"}, // publisher
				{PropertyPath: "properties.type"},      // type
			},
			SensitiveFields: []string{
				"properties.protectedSettings", // protected_settings (Sensitive)
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.autoUpgradeMinorVersion", Value: true}, // auto_upgrade_minor_version
				{PropertyPath: "properties.suppressFailures", Value: false},       // failure_suppression_enabled
			},
			RequiredFields: []string{
				"properties.publisher",          // publisher
				"properties.type",               // type
				"properties.typeHandlerVersion", // type_handler_version
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachineScaleSetExtension()) }
