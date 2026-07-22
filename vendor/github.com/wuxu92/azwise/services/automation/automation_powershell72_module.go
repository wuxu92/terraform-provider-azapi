package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationPowerShell72Module provides resource knowledge for
// Microsoft.Automation/automationAccounts/powerShell72Modules.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_powershell72_module_resource.go
//     schema (43-92), Create (110-211); timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/automation/2023-11-01/module
//     ModuleCreateOrUpdateProperties.contentLink (contentHash{algorithm,value}/uri/version).
//   - validators: StringIsNotEmpty (name), module.ValidateAutomationAccountID (parent).
//
// name / automation_account_id are the envelope + parent id (ForceNew by
// construction). module_link.uri and module_link.hash are Required within the block.
type AutomationPowerShell72Module struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationPowerShell72Module)(nil)

func NewAutomationPowerShell72Module() *AutomationPowerShell72Module {
	return &AutomationPowerShell72Module{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/powerShell72Modules",
			ApiVersions:  []string{"2023-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// module_link (contentLink.uri) is Required.
			RequiredFields: []string{
				"properties.contentLink.uri",
			},
			StringRules: []azwise.StringRule{
				// name — StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
			},
			// module_link.uri/hash carry no ValidateFunc in AzureRM (Required only);
			// hash.algorithm/value map to properties.contentLink.contentHash.{algorithm,value}.
		},
	}
}

func init() { azwise.Register(NewAutomationPowerShell72Module()) }
