package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationJobSchedule provides resource knowledge for
// Microsoft.Automation/automationAccounts/jobSchedules.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_job_schedule_resource.go
//     schema (29-106), Create (108-181); timeouts 30m/5m/-/30m (no update).
//   - go-azure-sdk resource-manager/automation/2024-10-23/jobschedule
//     JobScheduleCreateProperties: schedule(name)/runbook(name)/parameters/runOn.
//   - validators: validate.ScheduleName (schedule_name), validate.RunbookName (runbook_name),
//     validate.ParameterNames (parameters), validation.IsUUID (job_schedule_id = resource name).
//
// job_schedule_id is the ARM resource name (envelope); runbook_name /
// schedule_name / automation_account_name form the parent id + associations.
// Everything is ForceNew (the resource has no Update).
type AutomationJobSchedule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationJobSchedule)(nil)

func NewAutomationJobSchedule() *AutomationJobSchedule {
	return &AutomationJobSchedule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/jobSchedules",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// schedule/runbook associations, parameters and run_on are all ForceNew
			// and live in the body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.schedule.name"},
				{PropertyPath: "properties.runbook.name"},
				{PropertyPath: "properties.parameters"},
				{PropertyPath: "properties.runOn"},
			},
			// schedule_name and runbook_name are Required.
			RequiredFields: []string{
				"properties.schedule.name",
				"properties.runbook.name",
			},
			StringRules: []azwise.StringRule{
				// job_schedule_id (resource name) — validation.IsUUID.
				{
					Regex:   `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message: "job_schedule_id must be a valid UUID",
				},
				// schedule_name → properties.schedule.name — validate.ScheduleName.
				{
					PropertyPath: "properties.schedule.name",
					Regex:        `^[^<>*%&:\\?.+/]{0,127}[^<>*%&:\\?.+/\s]$`,
					Message:      "schedule_name must be 1-128 characters and cannot contain < > * % & : \\ ? . + /",
				},
				// runbook_name → properties.runbook.name — validate.RunbookName.
				{
					PropertyPath: "properties.runbook.name",
					Regex:        `^[0-9a-zA-Z][-_0-9a-zA-Z]{0,62}$`,
					Message:      "runbook_name may contain only letters, numbers, underscores and dashes, must begin with a letter, and be less than 64 characters",
				},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationJobSchedule()) }
