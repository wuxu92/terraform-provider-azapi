package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationRunbook provides resource knowledge for
// Microsoft.Automation/automationAccounts/runbooks.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_runbook_resource.go
//     contentLinkSchema (30-79), schema (100-286), CreateUpdate (290-383),
//     expandContentLink (500-529), expandDraft (531-566); timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/runbook
//     RunbookCreateOrUpdateProperties: runbookType/logVerbose/logProgress/
//     description/logActivityTrace/runtimeEnvironment/publishContentLink/draft.
//   - runbook.RunbookTypeEnum constants (constants.go:96-107).
//   - validators: validate.RunbookName (name), StringInSlice (runbook_type),
//     StringIsNotEmpty (runtime_environment_name, content, publish_content_link.*),
//     IntAtLeast(0) (log_activity_trace_level), IsURLWithScheme (publish_content_link.uri).
//
// name / automation_account_name are the envelope + parent id (ForceNew by
// construction). content and job_schedule are managed via the data-plane draft/
// publish flow and the separate jobSchedules resource respectively, so they carry
// no runbook body property path.
type AutomationRunbook struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationRunbook)(nil)

func NewAutomationRunbook() *AutomationRunbook {
	return &AutomationRunbook{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/runbooks",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// runbook_type is ForceNew and lives in the body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.runbookType"},
			},
			// runbook_type, log_progress and log_verbose are Required.
			RequiredFields: []string{
				"properties.runbookType",
				"properties.logProgress",
				"properties.logVerbose",
			},
			IntRules: []azwise.IntRule{
				// log_activity_trace_level → properties.logActivityTrace — IntAtLeast(0).
				{
					PropertyPath: "properties.logActivityTrace",
					MinValue:     azwise.Ptr[int64](0),
					Message:      "log_activity_trace_level must be at least 0",
				},
			},
			StringRules: []azwise.StringRule{
				// name — validate.RunbookName.
				{
					Regex:   `^[0-9a-zA-Z][-_0-9a-zA-Z]{0,62}$`,
					Message: "runbook name may contain only letters, numbers, underscores and dashes, must begin with a letter, and be less than 64 characters",
				},
				// runbook_type → properties.runbookType — StringInSlice (full SDK enum).
				{
					PropertyPath: "properties.runbookType",
					AllowedValues: []string{
						"Graph", "GraphPowerShell", "GraphPowerShellWorkflow",
						"PowerShell", "PowerShell72", "PowerShellWorkflow",
						"Python", "Python3", "Python2", "Script",
					},
					Message: "runbook_type must be a valid runbook type",
				},
				// runtime_environment_name → properties.runtimeEnvironment — StringIsNotEmpty.
				{
					PropertyPath: "properties.runtimeEnvironment",
					MinLength:    1,
					Message:      "runtime_environment_name must not be empty",
				},
				// publish_content_link.uri → properties.publishContentLink.uri —
				// IsURLWithScheme(http/https) OR empty.
				{
					PropertyPath: "properties.publishContentLink.uri",
					Regex:        `^(https?://.*)?$`,
					Message:      "publish_content_link uri must be an http/https URL or empty",
				},
				// publish_content_link.version → properties.publishContentLink.version — StringIsNotEmpty.
				{
					PropertyPath: "properties.publishContentLink.version",
					MinLength:    1,
					Message:      "publish_content_link version must not be empty",
				},
				// publish_content_link.hash.algorithm → properties.publishContentLink.contentHash.algorithm.
				{
					PropertyPath: "properties.publishContentLink.contentHash.algorithm",
					MinLength:    1,
					Message:      "publish_content_link hash algorithm must not be empty",
				},
				// publish_content_link.hash.value → properties.publishContentLink.contentHash.value.
				{
					PropertyPath: "properties.publishContentLink.contentHash.value",
					MinLength:    1,
					Message:      "publish_content_link hash value must not be empty",
				},
			},
			// TODO: AzureRM AtLeastOneOf{content, publish_content_link, draft} cannot be
			// captured as a RelationalRule because `content` is applied via the data-plane
			// draft/publish flow and has no runbook body property path.
		},
	}
}

func init() { azwise.Register(NewAutomationRunbook()) }
