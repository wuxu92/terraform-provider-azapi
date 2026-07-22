package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationHybridRunbookWorkerGroup provides resource knowledge for
// Microsoft.Automation/automationAccounts/hybridRunbookWorkerGroups.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_hybrid_runbook_worker_group_resource.go
//     schema (32-53), Create (67-110); timeouts 30m/5m/10m/10m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/hybridrunbookworkergroup
//     HybridRunbookWorkerGroupCreateOrUpdateProperties.credential (name).
//   - validators: StringIsNotEmpty (name, credential_name).
//
// name / automation_account_name are the envelope + parent id (ForceNew by
// construction).
type AutomationHybridRunbookWorkerGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationHybridRunbookWorkerGroup)(nil)

func NewAutomationHybridRunbookWorkerGroup() *AutomationHybridRunbookWorkerGroup {
	return &AutomationHybridRunbookWorkerGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/hybridRunbookWorkerGroups",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name — StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
				// credential_name → properties.credential.name — StringIsNotEmpty.
				{PropertyPath: "properties.credential.name", MinLength: 1, Message: "credential_name must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationHybridRunbookWorkerGroup()) }
