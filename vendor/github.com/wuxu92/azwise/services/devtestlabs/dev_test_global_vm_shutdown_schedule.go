package devtestlabs

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevTestGlobalVMShutdownSchedule provides resource knowledge for Microsoft.DevTestLab/schedules.
//
// This is the global (subscription-level) DevTest schedule type — note the ARM type has NO
// labs/ parent segment (globalschedules SDK: id_schedule.go segments are
// providers/Microsoft.DevTestLab/schedules/{scheduleName}).
//
// Contributing Terraform resource: azurerm_dev_test_global_vm_shutdown_schedule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devtestlabs/dev_test_global_vm_shutdown_schedule_resource.go
//     (schema L43-103, Create body L107-177, expanders L239-300)
//   - go-azure-sdk resource-manager/devtestlab/2018-09-15/globalschedules:
//     model_scheduleproperties.go, model_notificationsettings.go, constants.go (EnableStatus),
//     id_schedule.go.
//
// Notes:
//   - The ARM resource name is derived by AzureRM ("shutdown-computevm-<vmName>") and is not a
//     user-settable field, so no name rule is emitted.
//   - virtual_machine_id maps to properties.targetResourceId (ForceNew).
//   - properties.taskType is hardcoded to "ComputeVmShutdownTask" by AzureRM/the API.
//   - timezone uses computeValidate.VirtualMachineTimeZoneCaseInsensitive() — a semantic
//     timezone-membership check that cannot be expressed as a StringRule; left as a note for an
//     azapin service validator on properties.timeZoneId.
//   - enabled (default true) maps to properties.status ("Enabled"/"Disabled").
type DevTestGlobalVMShutdownSchedule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevTestGlobalVMShutdownSchedule)(nil)

func NewDevTestGlobalVMShutdownSchedule() *DevTestGlobalVMShutdownSchedule {
	return &DevTestGlobalVMShutdownSchedule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevTestLab/schedules",
			ApiVersions:  []string{"2018-09-15"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.targetResourceId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.dailyRecurrence.time", Regex: `^(0[0-9]|1[0-9]|2[0-3]|[0-9])[0-5][0-9]$`, Message: "Time of day must match the format HHmm where HH is 00-23 and mm is 00-59"},
				{PropertyPath: "properties.notificationSettings.status", AllowedValues: []string{"Disabled", "Enabled"}},
				{PropertyPath: "properties.status", AllowedValues: []string{"Disabled", "Enabled"}},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.notificationSettings.timeInMinutes", MinValue: azwise.Ptr(int64(15)), MaxValue: azwise.Ptr(int64(120))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.status", Value: "Enabled"},
				{PropertyPath: "properties.taskType", Value: "ComputeVmShutdownTask"},
				{PropertyPath: "properties.notificationSettings.timeInMinutes", Value: 30},
			},
			RequiredFields: []string{
				"properties.targetResourceId",
				"properties.taskType",
				"properties.timeZoneId",
				"properties.dailyRecurrence.time",
				"properties.notificationSettings",
				"properties.notificationSettings.status",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.uniqueIdentifier",
				"properties.createdDate",
			},
		},
	}
}

func init() { azwise.Register(NewDevTestGlobalVMShutdownSchedule()) }
