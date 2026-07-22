// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package mssqlmanagedinstance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StartStopSchedule provides resource knowledge for
// Microsoft.Sql/managedInstances/startStopSchedules
// (Terraform azurerm_mssql_managed_instance_start_stop_schedule).
//
// The ARM resource name is the constant "default". Body is
// {properties:{description, timeZoneId, scheduleList[]}} against the go-azure-sdk
// startstopmanagedinstanceschedules.StartStopManagedInstanceSchedule model.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssqlmanagedinstance/mssql_managed_instance_start_stop_schedule_resource.go:54-108
//     (Arguments: schedule list, timezone_id default UTC)
//   - .../mssql_managed_instance_start_stop_schedule_resource.go:124-177,301-315 (Create/expand → properties)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/startstopmanagedinstanceschedules:
//     model_startstopmanagedinstancescheduleproperties.go:6-12, model_scheduleitem.go:6-11,
//     constants.go (DayOfWeek)
//
// Intentionally not encoded (documented, not emitted):
//   - schedule block (Required, MinItems 1) maps to properties.scheduleList[*]
//     with start_day/stop_day enum (DayOfWeek: Monday..Sunday) and
//     start_time/stop_time non-empty strings. Array-element constraints are not
//     expressible as declarative rules, so they are documented here only.
//   - managed_instance_id (parent) is envelope-owned RequiresReplace.
type StartStopSchedule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StartStopSchedule)(nil)

// NewStartStopSchedule returns knowledge for Microsoft.Sql/managedInstances/startStopSchedules.
func NewStartStopSchedule() *StartStopSchedule {
	return &StartStopSchedule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/managedInstances/startStopSchedules",
			ApiVersions:  []string{"2023-08-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{},
			StringRules: []azwise.StringRule{
				// timezone_id → properties.timeZoneId (StringIsNotEmpty, default UTC).
				{PropertyPath: "properties.timeZoneId", MinLength: 1, Message: "timezone_id must not be empty"},
			},
			IntRules:        []azwise.IntRule{},
			FloatRules:      []azwise.FloatRule{},
			ArrayRules:      []azwise.ArrayRule{},
			SensitiveFields: []string{},
			ComputedFields:  []string{},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.timeZoneId", Value: "UTC"},
			},
			// scheduleList is Required (json:"scheduleList" without omitempty).
			RequiredFields: []string{"properties.scheduleList"},
		},
	}
}

func init() { azwise.Register(NewStartStopSchedule()) }
