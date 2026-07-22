package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationWatcher provides resource knowledge for
// Microsoft.Automation/automationAccounts/watchers.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_watcher_resource.go
//     schema (38-107), Create (117-166); timeouts 30m/5m/10m/10m.
//   - go-azure-sdk resource-manager/automation/2020-01-13-preview/watcher
//     WatcherProperties: description/executionFrequencyInSeconds/scriptName/
//     scriptParameters/scriptRunOn; status/creationTime are read-only.
//   - validators: StringIsNotEmpty (name, script_name, script_run_on, description,
//     etag), IntAtLeast(0) (execution_frequency_in_seconds).
//
// name / automation_account_id are the envelope + parent id (ForceNew by
// construction).
type AutomationWatcher struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationWatcher)(nil)

func NewAutomationWatcher() *AutomationWatcher {
	return &AutomationWatcher{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/watchers",
			ApiVersions:  []string{"2020-01-13-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			// script_name and script_parameters are ForceNew and live in the body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.scriptName"},
				{PropertyPath: "properties.scriptParameters"},
			},
			RequiredFields: []string{
				"properties.executionFrequencyInSeconds",
				"properties.scriptName",
				"properties.scriptRunOn",
			},
			// status is read-only (server-computed) in WatcherProperties.
			ComputedFields: []string{
				"properties.status",
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.executionFrequencyInSeconds",
					MinValue:     azwise.Ptr[int64](0),
					Message:      "execution_frequency_in_seconds must be at least 0",
				},
			},
			StringRules: []azwise.StringRule{
				// name — StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
				// script_name → properties.scriptName — StringIsNotEmpty.
				{PropertyPath: "properties.scriptName", MinLength: 1, Message: "script_name must not be empty"},
				// script_run_on → properties.scriptRunOn — StringIsNotEmpty.
				{PropertyPath: "properties.scriptRunOn", MinLength: 1, Message: "script_run_on must not be empty"},
				// description → properties.description — StringIsNotEmpty.
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
				// etag → etag (top-level) — StringIsNotEmpty.
				{PropertyPath: "etag", MinLength: 1, Message: "etag must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationWatcher()) }
