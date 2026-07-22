package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationDscNodeConfiguration provides resource knowledge for
// Microsoft.Automation/automationAccounts/nodeConfigurations.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_dsc_nodeconfiguration_resource.go
//     schema (24-72), Create (74-117); timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/dscnodeconfiguration
//     DscNodeConfigurationCreateOrUpdateParametersProperties.source (ContentSource{
//     type/value}) and .configuration (name). AzureRM hardcodes source.type to
//     "embeddedContent" and derives configuration.name from the node config name prefix.
//   - validators: StringIsNotEmpty (name, content_embedded).
//
// name / automation_account_name are the envelope + parent id (ForceNew by
// construction).
type AutomationDscNodeConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationDscNodeConfiguration)(nil)

func NewAutomationDscNodeConfiguration() *AutomationDscNodeConfiguration {
	return &AutomationDscNodeConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/nodeConfigurations",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// content_embedded is Required; AzureRM also hardcodes the ContentSource type
			// and the configuration association name (derived from the resource name).
			RequiredFields: []string{
				"properties.source.value",
				"properties.configuration.name",
			},
			// AzureRM hardcodes the content source type to "embeddedContent".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.source.type", Value: "embeddedContent"},
			},
			StringRules: []azwise.StringRule{
				// name — StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
				// content_embedded → properties.source.value — StringIsNotEmpty.
				{PropertyPath: "properties.source.value", MinLength: 1, Message: "content_embedded must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationDscNodeConfiguration()) }
