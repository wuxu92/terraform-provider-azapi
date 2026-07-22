package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationWebhook provides resource knowledge for
// Microsoft.Automation/automationAccounts/webhooks.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_webhook_resource.go
//     schema (lines 50-107) + Create (lines 111-172); timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/automation/2015-10-31/webhook
//     WebhookCreateOrUpdateProperties: isEnabled/expiryTime/parameters/runbook/runOn/uri.
//   - validators: validate.RunbookName (^[0-9a-zA-Z][-_0-9a-zA-Z]{0,62}$),
//     validation.IsRFC3339Time (expiry_time), validation.IsURLWithHTTPorHTTPS (uri).
//
// name / automation_account_name live on the operational envelope + parent id
// (Required + RequiresReplace), so their ForceNew is not repeated here.
type AutomationWebhook struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationWebhook)(nil)

func NewAutomationWebhook() *AutomationWebhook {
	return &AutomationWebhook{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/webhooks",
			ApiVersions:  []string{"2015-10-31"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// expiry_time and uri are ForceNew in AzureRM; both live in the body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.expiryTime"},
				{PropertyPath: "properties.uri"},
			},
			// Required: expiry_time and runbook_name.
			RequiredFields: []string{
				"properties.expiryTime",
				"properties.runbook.name",
			},
			// enabled defaults to true (AzureRM schema Default: true).
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isEnabled", Value: true},
			},
			// uri is Optional+Computed+Sensitive; it is generated when omitted but
			// IS settable via the ARM Create model, so it is not a ComputedField.
			SensitiveFields: []string{
				"properties.uri",
			},
			StringRules: []azwise.StringRule{
				// name — validation.StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
				// expiry_time → properties.expiryTime — validation.IsRFC3339Time.
				{
					PropertyPath: "properties.expiryTime",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "must be a valid RFC3339 date/time",
				},
				// runbook_name → properties.runbook.name — validate.RunbookName.
				{
					PropertyPath: "properties.runbook.name",
					Regex:        `^[0-9a-zA-Z][-_0-9a-zA-Z]{0,62}$`,
					Message:      "runbook name may contain only letters, numbers, underscores and dashes, must begin with a letter, and be less than 64 characters",
				},
				// uri → properties.uri — validation.IsURLWithHTTPorHTTPS.
				{
					PropertyPath: "properties.uri",
					Regex:        `^https?://`,
					Message:      "must be a valid URL with http or https scheme",
				},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationWebhook()) }
