package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationSoftwareUpdateConfiguration provides resource knowledge for
// Microsoft.Automation/automationAccounts/softwareUpdateConfigurations.
//
// NOTE: AzureRM deprecates azurerm_automation_software_update_configuration
// (Azure Automation Update Management was shut down 2025-02-28). Knowledge is
// still generated for completeness of the ARM type surface.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_software_update_configuration_resource.go
//     schema (141-584), Attributes (586-598), Create (608-645),
//     expandUpdateConfig (1053-1232); create timeout 30m.
//   - go-azure-sdk resource-manager/automation/2019-06-01/softwareupdateconfiguration
//     SoftwareUpdateConfigurationProperties{scheduleInfo, updateConfiguration, tasks,
//     error, createdBy, creationTime, lastModifiedBy, lastModifiedTime, provisioningState};
//     UpdateConfiguration{operatingSystem, duration, azureVirtualMachines,
//     nonAzureComputerNames, linux, windows, targets}; SUCScheduleProperties;
//     LinuxProperties/WindowsProperties; TargetProperties; TaskProperties.
//   - validators: StringIsNotEmpty (name), ISO8601Duration (duration),
//     StringInSlice (reboot, frequency), IsRFC3339Time (start/expiry/next times).
//
// name / automation_account_id are the envelope + parent id (ForceNew by
// construction).
type AutomationSoftwareUpdateConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationSoftwareUpdateConfiguration)(nil)

func NewAutomationSoftwareUpdateConfiguration() *AutomationSoftwareUpdateConfiguration {
	return &AutomationSoftwareUpdateConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/softwareUpdateConfigurations",
			ApiVersions:  []string{"2019-06-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// schedule (frequency) is Required; operatingSystem is a required
			// non-omitempty field derived from the chosen linux/windows block.
			RequiredFields: []string{
				"properties.scheduleInfo.frequency",
				"properties.updateConfiguration.operatingSystem",
			},
			DefaultValues: []azwise.DefaultValue{
				// duration → properties.updateConfiguration.duration, AzureRM Default "PT2H".
				{PropertyPath: "properties.updateConfiguration.duration", Value: "PT2H"},
				// linux.reboot / windows.reboot default "IfRequired".
				{PropertyPath: "properties.updateConfiguration.linux.rebootSetting", Value: "IfRequired"},
				{PropertyPath: "properties.updateConfiguration.windows.rebootSetting", Value: "IfRequired"},
				// schedule.is_enabled default true.
				{PropertyPath: "properties.scheduleInfo.isEnabled", Value: true},
				// schedule.time_zone default "Etc/UTC".
				{PropertyPath: "properties.scheduleInfo.timeZone", Value: "Etc/UTC"},
				// schedule.start_time/expiry_time/next_run: O+C, server-computed if omitted.
				{PropertyPath: "properties.scheduleInfo.startTime", Value: nil},
				{PropertyPath: "properties.scheduleInfo.expiryTime", Value: nil},
				{PropertyPath: "properties.scheduleInfo.nextRun", Value: nil},
			},
			// Read-only fields present in the GET model but not set on Create.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.createdBy",
				"properties.creationTime",
				"properties.lastModifiedBy",
				"properties.lastModifiedTime",
				"properties.error",
			},
			StringRules: []azwise.StringRule{
				// name — StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
				// duration → properties.updateConfiguration.duration — ISO8601Duration.
				{
					PropertyPath: "properties.updateConfiguration.duration",
					Regex:        `^P(?:\d+Y)?(?:\d+M)?(?:\d+W)?(?:\d+D)?(?:T(?:\d+H)?(?:\d+M)?(?:\d+(?:\.\d+)?S)?)?$`,
					Message:      "duration must be a valid ISO 8601 duration (e.g. PT2H)",
				},
				// linux.reboot → properties.updateConfiguration.linux.rebootSetting — StringInSlice.
				{
					PropertyPath:  "properties.updateConfiguration.linux.rebootSetting",
					AllowedValues: []string{"Always", "IfRequired", "Never", "RebootOnly"},
					Message:       "reboot must be one of Always, IfRequired, Never, RebootOnly",
				},
				// windows.reboot → properties.updateConfiguration.windows.rebootSetting — StringInSlice.
				{
					PropertyPath:  "properties.updateConfiguration.windows.rebootSetting",
					AllowedValues: []string{"Always", "IfRequired", "Never", "RebootOnly"},
					Message:       "reboot must be one of Always, IfRequired, Never, RebootOnly",
				},
				// schedule.frequency → properties.scheduleInfo.frequency — StringInSlice.
				{
					PropertyPath:  "properties.scheduleInfo.frequency",
					AllowedValues: []string{"OneTime", "Hour", "Day", "Week", "Month"},
					Message:       "frequency must be one of OneTime, Hour, Day, Week, Month",
				},
				// schedule.start_time → properties.scheduleInfo.startTime — IsRFC3339Time.
				{
					PropertyPath: "properties.scheduleInfo.startTime",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "start_time must be a valid RFC3339 date/time",
				},
				// schedule.expiry_time → properties.scheduleInfo.expiryTime — IsRFC3339Time.
				{
					PropertyPath: "properties.scheduleInfo.expiryTime",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "expiry_time must be a valid RFC3339 date/time",
				},
				// schedule.next_run → properties.scheduleInfo.nextRun — IsRFC3339Time.
				{
					PropertyPath: "properties.scheduleInfo.nextRun",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "next_run must be a valid RFC3339 date/time",
				},
				// schedule.time_zone → properties.scheduleInfo.timeZone — StringIsNotEmpty.
				{PropertyPath: "properties.scheduleInfo.timeZone", MinLength: 1, Message: "time_zone must not be empty"},
			},
			// linux and windows blocks are mutually exclusive (AzureRM ExactlyOneOf).
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.updateConfiguration.linux", "properties.updateConfiguration.windows"},
					Message: "exactly one of linux or windows must be configured",
				},
			},
			// TODO: several AzureRM constraints cannot be expressed declaratively here:
			//   - classifications_included maps to linux.includedPackageClassifications /
			//     windows.includedUpdateClassifications, which ARM stores as a single
			//     comma-joined enum string (not a per-element enum field).
			//   - target.azure_query.tag_filter (Any/All) →
			//     ...targets.azureQueries[*].tagSettings.filterOperator is an array-element enum.
			//   - schedule.advanced_week_days (weekday enum), advanced_month_days
			//     (IntBetween 1-31) and monthly_occurrence.occurrence (IntInSlice 1,2,3,4,-1)
			//     live under scheduleInfo.advancedSchedule and are array elements.
			//   - virtual_machine_ids (ValidateVirtualMachineID) and target azure_query.scope
			//     (resource-group/subscription id) are per-element semantic ARM-ID validators.
		},
	}
}

func init() { azwise.Register(NewAutomationSoftwareUpdateConfiguration()) }
