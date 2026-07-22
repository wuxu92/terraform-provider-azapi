package costmanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CostManagementScheduledAction provides resource knowledge for
// Microsoft.CostManagement/scheduledActions.
//
// This is a MERGED file across a KIND variant: AzureRM exposes one ARM type
// (scheduledActions) as two typed Terraform resources distinguished only by the
// scheduledActions `kind` discriminator:
//   - azurerm_cost_management_scheduled_action -> kind = "Email"
//   - azurerm_cost_anomaly_alert               -> kind = "InsightAlert"
//
// (Verified: both resources use the go-azure-sdk scheduledactions package and
// scheduledactions.ValidateScopedScheduledActionID; cost_anomaly_alert builds a
// ScheduledAction with Kind=InsightAlert, so it is NOT a distinct ARM type.)
//
// Per the shared-type rules, only knowledge that is UNIVERSALLY true for BOTH
// kinds is unioned here. Kind-specific items are documented but intentionally
// excluded so they do not corrupt validation for the other kind:
//   - view_id ForceNew: only the Email variant exposes/replaces view_id
//     (properties.viewId); InsightAlert hardcodes ms:DailyAnomalyByResourceGroup.
//     Non-universal ForceNew -> excluded.
//   - notificationEmail Required: the Email variant requires email_address_sender
//     (properties.notificationEmail); InsightAlert makes notification_email
//     Optional/Computed. Kind-specific required -> excluded.
//   - name validation differs by kind (InsightAlert: ^[a-z0-9-]*$ lowercase +
//     digits + hyphens; Email: non-whitespace). Envelope-level and non-universal
//     -> no name StringRule emitted.
//   - email_subject max length differs (Email 1-50 in 5.0 / 1-70 before; InsightAlert
//     1-70). Only the universal minimum (1) is expressed on the subject rule.
//
// Universal field -> ARM body mapping:
//   - kind                -> kind (Email | InsightAlert)
//   - display_name        -> properties.displayName
//   - view_id / (fixed)   -> properties.viewId
//   - (status hardcoded)  -> properties.status (Enabled)
//   - email_subject       -> properties.notification.subject
//   - email_addresses     -> properties.notification.to
//   - message             -> properties.notification.message
//   - frequency           -> properties.schedule.frequency (ScheduleFrequency)
//   - start_date          -> properties.schedule.startDate
//   - end_date            -> properties.schedule.endDate
// Email-only value constraints (safe: they only fire when the field is present,
// which never happens on an InsightAlert body):
//   - hour_of_day         -> properties.schedule.hourOfDay (0-23)
//   - day_of_month        -> properties.schedule.dayOfMonth (1-31)
//   - days_of_week[*]     -> properties.schedule.daysOfWeek[*] (array element, not lowered)
//   - weeks_of_month[*]   -> properties.schedule.weeksOfMonth[*] (array element, not lowered)
//
// Sources:
//   - terraform-provider-azurerm internal/services/costmanagement/cost_management_scheduled_action_resource.go
//     (Arguments 27-134; Create 152-229: kind=Email, schedule/notification mapping; 30m/5m timeouts)
//   - internal/services/costmanagement/cost_anomaly_alert_resource.go
//     (Arguments 27-80; Create 94-160: kind=InsightAlert, status Enabled, hardcoded view/schedule)
//   - internal/services/costmanagement/validate/cost_anomaly_alert_name.go (name regex, InsightAlert only)
//   - go-azure-sdk resource-manager/costmanagement/2023-08-01/scheduledactions
//     (model_scheduledaction*.go, model_scheduleproperties.go, model_notificationproperties.go, constants.go)
type CostManagementScheduledAction struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CostManagementScheduledAction)(nil)

// NewCostManagementScheduledAction returns knowledge for the scheduledActions resource.
func NewCostManagementScheduledAction() *CostManagementScheduledAction {
	return &CostManagementScheduledAction{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CostManagement/scheduledActions",
			ApiVersions:  []string{"2023-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "kind",
					AllowedValues: []string{"Email", "InsightAlert"},
					Message:       "kind must be Email or InsightAlert",
				},
				{
					PropertyPath:  "properties.status",
					AllowedValues: []string{"Disabled", "Enabled", "Expired"},
					Message:       "status must be one of Disabled, Enabled, Expired",
				},
				{
					PropertyPath:  "properties.schedule.frequency",
					AllowedValues: []string{"Daily", "Monthly", "Weekly"},
					Message:       "schedule frequency must be one of Daily, Monthly, Weekly",
				},
				{
					PropertyPath: "properties.notification.subject",
					MinLength:    1,
					Message:      "email subject must not be empty",
				},
				{
					PropertyPath: "properties.notification.message",
					MinLength:    1,
					MaxLength:    250,
					Message:      "message must be 1-250 characters",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.schedule.hourOfDay",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(23)),
					Message:      "hour of day must be between 0 and 23",
				},
				{
					PropertyPath: "properties.schedule.dayOfMonth",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(31)),
					Message:      "day of month must be between 1 and 31",
				},
			},
			// Both kinds always populate status=Enabled.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.status", Value: "Enabled"},
			},
			// ARM body properties present/required in BOTH kind bodies.
			RequiredFields: []string{
				"properties.displayName",
				"properties.viewId",
				"properties.status",
				"properties.notification.subject",
				"properties.notification.to",
				"properties.schedule.frequency",
				"properties.schedule.startDate",
				"properties.schedule.endDate",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCostManagementScheduledAction()) }
