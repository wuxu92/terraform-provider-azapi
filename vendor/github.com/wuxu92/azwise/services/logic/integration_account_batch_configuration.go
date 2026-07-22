package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccountBatchConfiguration provides resource knowledge for
// Microsoft.Logic/integrationAccounts/batchConfigurations.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account_batch_configuration.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_batch_configuration_resource.go
//     (schema L46-217, CustomizeDiff L219-238, Create body L264-273)
//   - terraform-provider-azurerm internal/services/logic/validate/
//     integration_account_batch_configuration_name.go,
//     integration_account_batch_configuration_batch_group_name.go
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccountbatchconfigurations:
//     model_batchconfiguration.go, model_batchconfigurationproperties.go,
//     model_batchreleasecriteria.go, model_workflowtriggerrecurrence.go,
//     constants.go (PossibleValuesForRecurrenceFrequency/DayOfWeek/DaysOfWeek), id_batchconfiguration.go
//
// Notes:
//   - name (ForceNew), integration_account_name (ForceNew parent segment) and
//     resource_group_name are envelope-owned; only the batch-configuration-name regex is emitted.
//   - batch_group_name is ForceNew → properties.batchGroupName.
//   - recurrence.frequency maps to properties.releaseCriteria.recurrence.frequency; the full
//     RecurrenceFrequency SDK enum is emitted.
//   - recurrence.schedule hours/minutes/month_days/week_days are arrays of scalars
//     (properties.releaseCriteria.recurrence.schedule.hours[*] etc.); per array-element policy
//     their per-item int/enum bounds are not emitted as declarative rules.
//   - CustomizeDiff constrains week_days↔Week / month_days↔Month / monthly↔Month; this is a
//     cross-field semantic check over schedule sub-fields, not a single-path rule — belongs in
//     an azapin customizer if enforced, not a declarative rule.
type IntegrationAccountBatchConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccountBatchConfiguration)(nil)

// NewIntegrationAccountBatchConfiguration returns knowledge for the batchConfigurations child resource.
func NewIntegrationAccountBatchConfiguration() *IntegrationAccountBatchConfiguration {
	return &IntegrationAccountBatchConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts/batchConfigurations",
			ApiVersions:  []string{"2019-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.batchGroupName"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9]+$`,
					MaxLength:    20,
					Message:      "batch configuration name contains only letters and numbers, up to 20 characters",
				},
				{
					PropertyPath: "properties.batchGroupName",
					Regex:        `^[A-Za-z0-9-_().]+$`,
					MaxLength:    80,
					Message:      "batch_group_name contains only letters, numbers, underscores, dots, parentheses and hyphens, up to 80 characters",
				},
				{
					PropertyPath:  "properties.releaseCriteria.recurrence.frequency",
					AllowedValues: []string{"Day", "Hour", "Minute", "Month", "NotSpecified", "Second", "Week", "Year"},
					Message:       "recurrence frequency must be a valid RecurrenceFrequency",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.releaseCriteria.batchSize",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(83886080)),
				},
				{
					PropertyPath: "properties.releaseCriteria.messageCount",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(8000)),
				},
				{
					PropertyPath: "properties.releaseCriteria.recurrence.interval",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(100)),
				},
			},
			RequiredFields: []string{
				"properties.batchGroupName",
				"properties.releaseCriteria",
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccountBatchConfiguration()) }
