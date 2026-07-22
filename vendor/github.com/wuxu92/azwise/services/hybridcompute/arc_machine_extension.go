package hybridcompute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ArcMachineExtension provides resource knowledge for
// Microsoft.HybridCompute/machines/extensions.
//
// Contributing TF resource: azurerm_arc_machine_extension.
//
// Sources:
//   - AzureRM internal/services/hybridcompute/arc_machine_extension_resource.go
//     (model :24-36, schema :54-135, Create body :169-209)
//   - go-azure-sdk .../hybridcompute/2022-11-10/machineextensions:
//     model_machineextension.go, model_machineextensionproperties.go
//     (ARM body paths, provisioningState/instanceView read-only)
type ArcMachineExtension struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ArcMachineExtension)(nil)

func NewArcMachineExtension() *ArcMachineExtension {
	return &ArcMachineExtension{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HybridCompute/machines/extensions",
			ApiVersions:  []string{"2022-11-10"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.publisher"},
				{PropertyPath: "properties.type"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: validation.All(StringIsNotEmpty, StringDoesNotContainAny("/")).
					Regex:     `^[^/]+$`,
					MinLength: 1,
					Message:   "must not be empty and must not contain '/'",
				},
				{
					// force_update_tag: validation.StringIsNotEmpty.
					PropertyPath: "properties.forceUpdateTag",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					// publisher: validation.StringIsNotEmpty.
					PropertyPath: "properties.publisher",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					// type: validation.StringIsNotEmpty.
					PropertyPath: "properties.type",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					// type_handler_version: validation.StringIsNotEmpty.
					PropertyPath: "properties.typeHandlerVersion",
					MinLength:    1,
					Message:      "must not be empty",
				},
				// protected_settings / settings: validation.StringIsJSON — a semantic
				// JSON check over a map[string]interface{} body property; not expressible
				// as a declarative StringRule, so intentionally omitted here.
			},
			DefaultValues: []azwise.DefaultValue{
				// automatic_upgrade_enabled schema Default:true.
				{PropertyPath: "properties.enableAutomaticUpgrade", Value: true},
			},
			RequiredFields: []string{
				"properties.publisher",
				"properties.type",
			},
			SensitiveFields: []string{
				"properties.protectedSettings",
			},
			// provisioningState and instanceView are Azure-populated read-only fields
			// (present on GET, absent from the Create/Update body).
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.instanceView",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewArcMachineExtension()) }
