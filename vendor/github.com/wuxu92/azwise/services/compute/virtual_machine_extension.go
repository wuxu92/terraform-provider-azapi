package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineExtension provides resource knowledge for
// Microsoft.Compute/virtualMachines/extensions.
//
// Contributing TF resource:
//   - azurerm_virtual_machine_extension (internal/services/compute/virtual_machine_extension_resource.go)
//
// Sources:
//   - virtual_machine_extension_resource.go:30-130 (schema, timeouts, ForceNew)
//   - go-azure-sdk compute/2024-03-01/virtualmachineextensions/model_virtualmachineextensionproperties.go
type VirtualMachineExtension struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineExtension)(nil)

func NewVirtualMachineExtension() *VirtualMachineExtension {
	return &VirtualMachineExtension{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/virtualMachines/extensions",
			ApiVersions:  []string{"2024-03-01"},
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
			},
			SensitiveFields: []string{
				"properties.protectedSettings", // protected_settings (Sensitive)
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.suppressFailures", Value: false}, // failure_suppression_enabled
			},
			RequiredFields: []string{
				"properties.publisher",          // publisher
				"properties.type",               // type
				"properties.typeHandlerVersion", // type_handler_version
			},
			// protected_settings ConflictsWith protected_settings_from_key_vault.
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.protectedSettings", "properties.protectedSettingsFromKeyVault"},
					Message: "protected_settings cannot be combined with protected_settings_from_key_vault",
				},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachineExtension()) }
