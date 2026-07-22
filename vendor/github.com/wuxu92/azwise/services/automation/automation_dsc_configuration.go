package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationDscConfiguration provides resource knowledge for
// Microsoft.Automation/automationAccounts/configurations.
//
// Contributing Terraform resource: azurerm_automation_dsc_configuration.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_dsc_configuration_resource.go
//     (schema L26-92, Create L94-135)
//   - go-azure-sdk resource-manager/automation/2024-10-23/dscconfiguration:
//     model_dscconfigurationcreateorupdateproperties.go (Create body: description, logProgress,
//     logVerbose, parameters, source) and model_dscconfigurationproperties.go (GET-only:
//     creationTime, jobCount, lastModifiedTime, nodeConfigurationCount, provisioningState, state).
//
// Notes:
//   - content_embedded maps into properties.source.value (with properties.source.type hardcoded
//     to "embeddedContent" by AzureRM).
type AutomationDscConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationDscConfiguration)(nil)

// NewAutomationDscConfiguration returns knowledge for the automationAccounts/configurations resource.
func NewAutomationDscConfiguration() *AutomationDscConfiguration {
	return &AutomationDscConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/configurations",
			ApiVersions:  []string{"2024-10-23"},
			// name & automation_account_name are envelope-owned (Required+ForceNew).
			// commonschema.Location() is ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// content_embedded is Required → properties.source.value.
			RequiredFields: []string{
				"properties.source.value",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.StringMatch
				{
					Regex:     `^[a-zA-Z0-9_]{1,64}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters and contain only letters, numbers and underscores",
				},
				// ── content_embedded → properties.source.value ── validation.StringIsNotEmpty
				{
					PropertyPath: "properties.source.value",
					MinLength:    1,
					Message:      "must not be empty",
				},
			},
			// log_verbose Default false → properties.logVerbose.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.logVerbose", Value: false},
			},
			// GET-only fields, absent from the Create/Update model.
			ComputedFields: []string{
				"properties.creationTime",
				"properties.jobCount",
				"properties.lastModifiedTime",
				"properties.nodeConfigurationCount",
				"properties.provisioningState",
				"properties.state",
			},
		},
	}
}

func init() { azwise.Register(NewAutomationDscConfiguration()) }
