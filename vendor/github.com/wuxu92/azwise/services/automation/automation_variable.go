package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationVariable provides resource knowledge for
// Microsoft.Automation/automationAccounts/variables.
//
// Contributing Terraform resources (all map to the same ARM type; the value's
// wire encoding differs per type but the ARM body shape is identical):
//   - azurerm_automation_variable_bool
//   - azurerm_automation_variable_datetime
//   - azurerm_automation_variable_int
//   - azurerm_automation_variable_object
//   - azurerm_automation_variable_string
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_variable.go
//     (resourceAutomationVariableCommonSchema L64-99, Create L156-202, formatAutomationVariableValue L134-154)
//   - per-type resource files automation_variable_{bool,datetime,int,object,string}_resource.go
//   - go-azure-sdk resource-manager/automation/2024-10-23/variable:
//     model_variablecreateorupdateproperties.go (Create body: description, isEncrypted, value)
//     model_variableproperties.go (GET-only: creationTime, lastModifiedTime).
//
// Notes:
//   - properties.value is always an ARM string, but AzureRM transforms the user value before
//     sending (bool→"true", int→"5", datetime→"\/Date(ms)\/", string→quoted JSON, object→raw
//     JSON). The AzureRM value ValidateFuncs (StringIsNotEmpty / IsRFC3339Time) validate the
//     pre-transform Terraform input, NOT the ARM body value, so they are not portable to a
//     declarative StringRule on properties.value and are intentionally omitted.
type AutomationVariable struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationVariable)(nil)

// NewAutomationVariable returns knowledge for the automationAccounts/variables resource.
func NewAutomationVariable() *AutomationVariable {
	return &AutomationVariable{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/variables",
			ApiVersions:  []string{"2024-10-23"},
			// name & automation_account_name are envelope-owned (Required+ForceNew).
			// No body-path ForceNew.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.StringIsNotEmpty
				{
					MinLength: 1,
					Message:   "must not be empty",
				},
			},
			// encrypted Default false → properties.isEncrypted.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isEncrypted", Value: false},
			},
			// GET-only server timestamps, absent from the Create/Update model.
			ComputedFields: []string{
				"properties.creationTime",
				"properties.lastModifiedTime",
			},
		},
	}
}

func init() { azwise.Register(NewAutomationVariable()) }
