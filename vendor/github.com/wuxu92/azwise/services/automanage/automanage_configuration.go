package automanage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomanageConfiguration provides resource knowledge for
// Microsoft.Automanage/configurationProfiles.
//
// Mirrors azurerm_automanage_configuration. The entire AzureRM configuration is
// flattened into a single free-form ARM object at properties.configuration, whose
// keys are slash-delimited strings (e.g. "Antimalware/ScanType"). Because azwise
// dot-path traversal splits only on ".", each slash-keyed leaf is addressable as
// properties.configuration.<Slash/Delimited/Key>, so the AzureRM schema validators
// map cleanly onto those paths.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automanage/automanage_configuration_resource.go
//     schema (Arguments 118-466), expandConfigurationProfile (652-752) for the exact
//     properties.configuration key names; timeouts 30m/5m/30m/30m (Create 472, Update 519,
//     Read 552, Delete 623).
//   - go-azure-sdk resource-manager/automanage/2022-05-04/configurationprofiles
//     ConfigurationProfileProperties.Configuration (*interface{} json:"configuration").
type AutomanageConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomanageConfiguration)(nil)

// NewAutomanageConfiguration returns knowledge for the configurationProfiles resource.
func NewAutomanageConfiguration() *AutomanageConfiguration {
	return &AutomanageConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automanage/configurationProfiles",
			ApiVersions:  []string{"2022-05-04"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// antimalware.scheduled_scan_type -> StringInSlice(Quick, Full).
				{
					PropertyPath:  "properties.configuration.Antimalware/ScanType",
					AllowedValues: []string{"Quick", "Full"},
					Message:       "antimalware scheduled scan type must be Quick or Full",
				},
				// azure_security_baseline.assignment_type -> StringInSlice(...).
				{
					PropertyPath:  "properties.configuration.AzureSecurityBaseline/AssignmentType",
					AllowedValues: []string{"ApplyAndAutoCorrect", "ApplyAndMonitor", "Audit", "DeployAndAutoCorrect"},
					Message:       "assignment type must be ApplyAndAutoCorrect, ApplyAndMonitor, Audit, or DeployAndAutoCorrect",
				},
				// backup.policy_name -> StringMatch(^[a-zA-Z0-9][a-zA-Z0-9-]{2,149}$).
				{
					PropertyPath: "properties.configuration.Backup/PolicyName",
					Regex:        "^[a-zA-Z0-9][a-zA-Z0-9-]{2,149}$",
					Message:      "backup policy name must be 3-150 characters, begin with an alphanumeric character, and contain only alphanumeric characters and hyphens",
				},
				// backup.schedule_policy.schedule_run_frequency -> StringInSlice(Daily, Weekly).
				{
					PropertyPath:  "properties.configuration.Backup/SchedulePolicy/ScheduleRunFrequency",
					AllowedValues: []string{"Daily", "Weekly"},
					Message:       "schedule run frequency must be Daily or Weekly",
				},
				// backup.schedule_policy.schedule_policy_type -> StringInSlice(SimpleSchedulePolicy).
				{
					PropertyPath:  "properties.configuration.Backup/SchedulePolicy/SchedulePolicyType",
					AllowedValues: []string{"SimpleSchedulePolicy"},
					Message:       "schedule policy type must be SimpleSchedulePolicy",
				},
				// backup.retention_policy.retention_policy_type -> StringInSlice(LongTermRetentionPolicy).
				{
					PropertyPath:  "properties.configuration.Backup/RetentionPolicy/RetentionPolicyType",
					AllowedValues: []string{"LongTermRetentionPolicy"},
					Message:       "retention policy type must be LongTermRetentionPolicy",
				},
				// backup.retention_policy.daily_schedule.retention_duration.duration_type -> StringInSlice(Days).
				{
					PropertyPath:  "properties.configuration.Backup/RetentionPolicy/DailySchedule/RetentionDuration/DurationType",
					AllowedValues: []string{"Days"},
					Message:       "daily retention duration type must be Days",
				},
				// backup.retention_policy.weekly_schedule.retention_duration.duration_type -> StringInSlice(Weeks).
				{
					PropertyPath:  "properties.configuration.Backup/RetentionPolicy/WeeklySchedule/RetentionDuration/DurationType",
					AllowedValues: []string{"Weeks"},
					Message:       "weekly retention duration type must be Weeks",
				},
			},
			IntRules: []azwise.IntRule{
				// antimalware.scheduled_scan_day -> IntInSlice(0..8) (contiguous 0-8).
				{
					PropertyPath: "properties.configuration.Antimalware/ScanDay",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(8)),
					Message:      "antimalware scheduled scan day must be between 0 (daily) and 8 (disabled)",
				},
				// antimalware.scheduled_scan_time_in_minutes -> IntBetween(0, 1439).
				{
					PropertyPath: "properties.configuration.Antimalware/ScanTimeInMinutes",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(1439)),
					Message:      "antimalware scheduled scan time must be between 0 and 1439 minutes",
				},
				// backup.instant_rp_retention_range_in_days -> IntBetween(1, 5).
				{
					PropertyPath: "properties.configuration.Backup/InstantRpRetentionRangeInDays",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(5)),
					Message:      "instant RP retention range must be between 1 and 5 days",
				},
				// backup.retention_policy.daily_schedule.retention_duration.count -> IntBetween(7, 9999).
				{
					PropertyPath: "properties.configuration.Backup/RetentionPolicy/DailySchedule/RetentionDuration/Count",
					MinValue:     azwise.Ptr(int64(7)),
					MaxValue:     azwise.Ptr(int64(9999)),
					Message:      "daily retention duration count must be between 7 and 9999",
				},
				// backup.retention_policy.weekly_schedule.retention_duration.count -> IntBetween(1, 5163).
				{
					PropertyPath: "properties.configuration.Backup/RetentionPolicy/WeeklySchedule/RetentionDuration/Count",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(5163)),
					Message:      "weekly retention duration count must be between 1 and 5163",
				},
			},
			// NOTE: AzureRM schema defaults inside properties.configuration (e.g.
			// Antimalware/ScanType="Quick", Backup/TimeZone="UTC") are only emitted when
			// their parent block is present, so they are conditional and not declared as
			// static DefaultValues here.
		},
	}
}

func init() { azwise.Register(NewAutomanageConfiguration()) }
