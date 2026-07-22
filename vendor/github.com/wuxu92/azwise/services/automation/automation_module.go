package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationModule provides resource knowledge for
// Microsoft.Automation/automationAccounts/modules.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_module_resource.go
//     schema (23-91), Create (94-134), expandModuleLink (230-251); timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/module
//     ModuleCreateOrUpdateProperties.contentLink (contentHash{algorithm,value}/uri/version).
//   - validators: StringIsNotEmpty (name), validate.AutomationAccount (parent).
//
// name / automation_account_name are the envelope + parent id (ForceNew by
// construction). module_link.uri and module_link.hash are Required within the block.
type AutomationModule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationModule)(nil)

func NewAutomationModule() *AutomationModule {
	return &AutomationModule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/modules",
			ApiVersions:  []string{"2024-10-23"},
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

func init() { azwise.Register(NewAutomationModule()) }
