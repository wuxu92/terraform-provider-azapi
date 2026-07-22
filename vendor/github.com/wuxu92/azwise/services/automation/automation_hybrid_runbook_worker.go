package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationHybridRunbookWorker provides resource knowledge for
// Microsoft.Automation/automationAccounts/hybridRunbookWorkerGroups/hybridRunbookWorkers.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_hybrid_runbook_worker_resource.go
//     schema (38-99), Create (109-154); timeouts 30m/5m/-/10m (no update).
//   - go-azure-sdk resource-manager/automation/2024-10-23/hybridrunbookworker
//     HybridRunbookWorkerCreateOrUpdateParameters.vmResourceId; HybridRunbookWorkerProperties
//     ip/lastSeenDateTime/registeredDateTime/workerName/workerType are read-only.
//   - validators: validation.IsUUID (worker_id = resource name), StringIsNotEmpty (vm_resource_id).
//
// worker_id is the ARM resource name (envelope); worker_group_name /
// automation_account_name are the parent id (ForceNew by construction).
type AutomationHybridRunbookWorker struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationHybridRunbookWorker)(nil)

func NewAutomationHybridRunbookWorker() *AutomationHybridRunbookWorker {
	return &AutomationHybridRunbookWorker{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/hybridRunbookWorkerGroups/hybridRunbookWorkers",
			ApiVersions:  []string{"2024-10-23"},
			// The resource has no Update; delete/read only besides create.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 10 * time.Minute,
			},
			// vm_resource_id is ForceNew and lives in the body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.vmResourceId"},
			},
			// vm_resource_id is Required.
			RequiredFields: []string{
				"properties.vmResourceId",
			},
			// ip/registeredDateTime/lastSeenDateTime/workerName/workerType are read-only
			// in HybridRunbookWorkerProperties and absent from the Create parameters.
			ComputedFields: []string{
				"properties.ip",
				"properties.registeredDateTime",
				"properties.lastSeenDateTime",
				"properties.workerName",
				"properties.workerType",
			},
			StringRules: []azwise.StringRule{
				// worker_id (resource name) — validation.IsUUID.
				{
					Regex:   `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message: "worker_id must be a valid UUID",
				},
				// vm_resource_id → properties.vmResourceId — StringIsNotEmpty.
				{PropertyPath: "properties.vmResourceId", MinLength: 1, Message: "vm_resource_id must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationHybridRunbookWorker()) }
