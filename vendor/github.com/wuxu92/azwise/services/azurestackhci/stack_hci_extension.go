package azurestackhci

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StackHCIExtension provides resource knowledge for
// Microsoft.AzureStackHCI/clusters/arcSettings/extensions.
//
// Mirrors azurerm_stack_hci_extension. name lives on the operational envelope; arc_setting_id
// is the parent scope (envelope). Body properties nest under
// properties.extensionParameters.*.
//
// Sources:
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_extension_resource.go:51-118
//     (schema: ForceNew, validators, defaults, sensitive protected_settings)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_extension_resource.go:124-140
//     (CustomizeDiff: type_handler_version conflicts with automatic_upgrade_enabled=true)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_extension_resource.go:170-201
//     (create mapping to extensions.ExtensionParameters)
//   - go-azure-sdk resource-manager/azurestackhci/2024-01-01/extensions:
//     model_extensionproperties.go, model_extensionparameters.go (autoUpgradeMinorVersion,
//     enableAutomaticUpgrade, publisher, type, typeHandlerVersion, settings, protectedSettings)
//
// Not encoded (deliberate):
//   - name carries validation.All(StringIsNotEmpty, StringDoesNotContainAny("/")); the
//     no-slash rule maps cleanly to a regex on the name attribute (encoded below), but the
//     non-empty part is covered by MinLength.
//   - CustomizeDiff enforces that type_handler_version cannot be set when
//     automatic_upgrade_enabled is true. This is a value-conditional relation (the constraint
//     fires only when enableAutomaticUpgrade == true), not a presence relation, so it cannot
//     be expressed as a RelationalRule (those only test whether a path is set). Left to a
//     customizer/hook.
type StackHCIExtension struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StackHCIExtension)(nil)

// NewStackHCIExtension returns knowledge for the Azure Stack HCI extensions resource.
func NewStackHCIExtension() *StackHCIExtension {
	return &StackHCIExtension{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureStackHCI/clusters/arcSettings/extensions",
			ApiVersions:  []string{"2024-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.extensionParameters.publisher"},
				{PropertyPath: "properties.extensionParameters.type"},
				{PropertyPath: "properties.extensionParameters.autoUpgradeMinorVersion"},
			},
			RequiredFields: []string{
				"properties.extensionParameters.publisher",
				"properties.extensionParameters.type",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: non-empty and must not contain "/".
					Regex:     `^[^/]+$`,
					MinLength: 1,
					Message:   "name cannot be empty and must not contain '/'",
				},
			},
			SensitiveFields: []string{
				"properties.extensionParameters.protectedSettings",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.extensionParameters.autoUpgradeMinorVersion", Value: true},
				{PropertyPath: "properties.extensionParameters.enableAutomaticUpgrade", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewStackHCIExtension()) }
