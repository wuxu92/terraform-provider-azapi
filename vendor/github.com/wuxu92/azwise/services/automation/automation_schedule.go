package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationSchedule provides resource knowledge for
// Microsoft.Automation/automationAccounts/schedules.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_schedule_resource.go
//     schema (lines 50-181), CustomizeDiff (183-211), Create (215-292),
//     expandArmAutomationScheduleAdvanced (458-502); timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/schedule
//     ScheduleCreateOrUpdateProperties: description/frequency/interval/startTime/
//     expiryTime/timeZone/advancedSchedule(weekDays/monthDays/monthlyOccurrences).
//   - validators: validate.ScheduleName (name), StringInSlice (frequency, week_days,
//     monthly_occurrence.day), IntBetween (interval 1-100, month_days -1..31 excl 0,
//     occurrence -1..5 excl 0), IsRFC3339Time (start_time/expiry_time),
//     azvalidate.AzureTimeZoneString (timezone).
//
// name / automation_account_name are the envelope + parent id (ForceNew by
// construction). week_days/month_days/monthly_occurrence are mutually exclusive
// and are only valid for the matching frequency (enforced by CustomizeDiff);
// captured here as ConflictsWith.
type AutomationSchedule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationSchedule)(nil)

func NewAutomationSchedule() *AutomationSchedule {
	return &AutomationSchedule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/schedules",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// frequency is Required.
			RequiredFields: []string{
				"properties.frequency",
			},
			DefaultValues: []azwise.DefaultValue{
				// timezone → properties.timeZone, AzureRM schema Default "Etc/UTC".
				{PropertyPath: "properties.timeZone", Value: "Etc/UTC"},
				// interval → properties.interval: O+C, AzureRM defaults it to 1 unless
				// frequency is OneTime (in which case it is omitted). Dynamic → nil.
				{PropertyPath: "properties.interval", Value: nil},
				// start_time → properties.startTime: O+C, AzureRM defaults to now+7m. Dynamic.
				{PropertyPath: "properties.startTime", Value: nil},
				// expiry_time → properties.expiryTime: O+C, server-computed for recurring.
				{PropertyPath: "properties.expiryTime", Value: nil},
			},
			IntRules: []azwise.IntRule{
				// interval → properties.interval — validation.IntBetween(1, 100).
				{
					PropertyPath: "properties.interval",
					MinValue:     azwise.Ptr[int64](1),
					MaxValue:     azwise.Ptr[int64](100),
					Message:      "interval must be between 1 and 100",
				},
			},
			StringRules: []azwise.StringRule{
				// name — validate.ScheduleName: 1-128 chars, no < > * % & : \ ? . + /,
				// cannot end with whitespace.
				{
					Regex:   `^[^<>*%&:\\?.+/]{0,127}[^<>*%&:\\?.+/\s]$`,
					Message: "name must be 1-128 characters, cannot contain < > * % & : \\ ? . + / and cannot end with whitespace",
				},
				// frequency → properties.frequency — StringInSlice.
				{
					PropertyPath:  "properties.frequency",
					AllowedValues: []string{"Day", "Hour", "Month", "OneTime", "Week"},
					Message:       "frequency must be one of Day, Hour, Month, OneTime, Week",
				},
				// start_time → properties.startTime — validation.IsRFC3339Time.
				{
					PropertyPath: "properties.startTime",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "must be a valid RFC3339 date/time",
				},
				// expiry_time → properties.expiryTime — validation.IsRFC3339Time.
				{
					PropertyPath: "properties.expiryTime",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "must be a valid RFC3339 date/time",
				},
				// monthly_occurrence.day → properties.advancedSchedule.monthlyOccurrences[*].day.
				// StringInSlice weekday enum. Array-element path; validated per element.
				{
					PropertyPath:  "properties.advancedSchedule.monthlyOccurrences[*].day",
					AllowedValues: []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
					Message:       "day must be a valid day of week",
				},
				// week_days → properties.advancedSchedule.weekDays[*] — StringInSlice weekday enum.
				{
					PropertyPath:  "properties.advancedSchedule.weekDays[*]",
					AllowedValues: []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
					Message:       "week_days must be valid days of week",
				},
			},
			// AzureRM ConflictsWith: week_days / month_days / monthly_occurrence are
			// mutually exclusive (also frequency-gated by CustomizeDiff).
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.advancedSchedule.weekDays",
						"properties.advancedSchedule.monthDays",
						"properties.advancedSchedule.monthlyOccurrences",
					},
					Message: "week_days, month_days and monthly_occurrence are mutually exclusive",
				},
				{
					Paths: []string{
						"properties.advancedSchedule.monthDays",
						"properties.advancedSchedule.weekDays",
						"properties.advancedSchedule.monthlyOccurrences",
					},
					Message: "week_days, month_days and monthly_occurrence are mutually exclusive",
				},
				{
					Paths: []string{
						"properties.advancedSchedule.monthlyOccurrences",
						"properties.advancedSchedule.weekDays",
						"properties.advancedSchedule.monthDays",
					},
					Message: "week_days, month_days and monthly_occurrence are mutually exclusive",
				},
			},
			// TODO: month_days (properties.advancedSchedule.monthDays[*]) and
			// monthly_occurrence.occurrence carry IntBetween(-1,31)/IntBetween(-1,5)
			// excluding 0; these are array-element int constraints not expressible
			// via IntRule (which targets a single scalar path), so they are omitted.
		},
	}
}

func init() { azwise.Register(NewAutomationSchedule()) }
