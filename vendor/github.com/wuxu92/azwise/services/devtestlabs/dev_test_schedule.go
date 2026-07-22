package devtestlabs

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevTestLabSchedule provides resource knowledge for Microsoft.DevTestLab/labs/schedules.
//
// Contributing Terraform resource: azurerm_dev_test_schedule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devtestlabs/dev_test_schedule_resource.go
//     (schema L51-190, Create body L194-268, recurrence/notification expanders L341-460)
//   - go-azure-sdk resource-manager/devtestlab/2018-09-15/schedules:
//     model_scheduleproperties.go, model_notificationsettings.go, model_daydetails.go,
//     model_hourdetails.go, model_weekdetails.go, constants.go (EnableStatus),
//     id_labschedule.go.
//
// Notes:
//   - name/lab_name are envelope-owned (ARM name segment + parent labs/{labName}) and
//     ForceNew; name validation is StringIsNotEmpty (MinLength 1, empty PropertyPath).
//   - time_zone_id uses computeValidate.VirtualMachineTimeZoneCaseInsensitive() — a semantic
//     timezone-membership check that cannot be expressed as a StringRule; left as a note for
//     an azapin service validator on properties.timeZoneId.
//   - weekly_recurrence.week_days is an enum (Monday..Sunday) over an array of strings
//     (properties.weeklyRecurrence.weekdays[*]); array-element paths are unsupported, so it
//     is documented rather than emitted.
type DevTestLabSchedule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevTestLabSchedule)(nil)

func NewDevTestLabSchedule() *DevTestLabSchedule {
	return &DevTestLabSchedule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevTestLab/labs/schedules",
			ApiVersions:  []string{"2018-09-15"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name validation (empty PropertyPath validates the name).
				{MinLength: 1, Message: "must not be empty"},
				{PropertyPath: "properties.status", AllowedValues: []string{"Disabled", "Enabled"}},
				{PropertyPath: "properties.notificationSettings.status", AllowedValues: []string{"Disabled", "Enabled"}},
				{PropertyPath: "properties.weeklyRecurrence.time", Regex: `^(0[0-9]|1[0-9]|2[0-3]|[0-9])[0-5][0-9]$`, Message: "Time of day must match the format HHmm where HH is 00-23 and mm is 00-59"},
				{PropertyPath: "properties.dailyRecurrence.time", Regex: `^(0[0-9]|1[0-9]|2[0-3]|[0-9])[0-5][0-9]$`, Message: "Time of day must match the format HHmm where HH is 00-23 and mm is 00-59"},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.hourlyRecurrence.minute", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(59))},
				{PropertyPath: "properties.notificationSettings.timeInMinutes", MinValue: azwise.Ptr(int64(1))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.status", Value: "Disabled"},
				{PropertyPath: "properties.notificationSettings.status", Value: "Disabled"},
			},
			RequiredFields: []string{
				"properties.taskType",
				"properties.timeZoneId",
				"properties.notificationSettings",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.uniqueIdentifier",
				"properties.createdDate",
			},
		},
	}
}

func init() { azwise.Register(NewDevTestLabSchedule()) }
