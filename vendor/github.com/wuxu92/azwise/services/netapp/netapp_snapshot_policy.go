package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppSnapshotPolicy provides resource knowledge for
// Microsoft.NetApp/netAppAccounts/snapshotPolicies.
//
// Mirrors azurerm_netapp_snapshot_policy.
//
// Sources:
//   - internal/services/netapp/netapp_snapshot_policy_resource.go
//     (schema 32-217: name ForceNew+SnapshotName; enabled Required bool; hourly/daily/
//     weekly/monthly schedule blocks with snapshots_to_keep IntBetween(0,255),
//     hour IntBetween(0,23), minute IntBetween(0,59); timeouts C/U/D 30m R 5m;
//     create 219-265 → SnapshotPolicyProperties{Enabled, Hourly/Daily/Weekly/Monthly};
//     CustomizeDiff 199-215 ForceNew when a schedule block is removed).
//   - internal/services/netapp/validate/snapshot_name.go (regex ^[\da-zA-Z][-_\da-zA-Z]{3,63}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/snapshotpolicies:
//     model_snapshotpolicyproperties.go (provisioningState read-only),
//     model_{hourly,daily,weekly,monthly}schedule.go (snapshotsToKeep/hour/minute int64;
//     day/daysOfMonth are comma-joined strings; usedBytes read-only),
//     id_snapshotpolicy.go (type segment casing "snapshotPolicies").
type NetAppSnapshotPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppSnapshotPolicy)(nil)

// NewNetAppSnapshotPolicy returns knowledge for the snapshotPolicies resource.
func NewNetAppSnapshotPolicy() *NetAppSnapshotPolicy {
	return &NetAppSnapshotPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/snapshotPolicies",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			// enabled is Required.
			RequiredFields: []string{
				"properties.enabled",
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM uses SnapshotName: 4-64 chars, start alphanumeric.
					Regex:     `^[\da-zA-Z][-_\da-zA-Z]{3,63}$`,
					MinLength: 4,
					MaxLength: 64,
					Message:   "must be 4-64 characters, start with a letter or number, and contain only letters, numbers, underscores and hyphens",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.hourlySchedule.snapshotsToKeep", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(255))},
				{PropertyPath: "properties.hourlySchedule.minute", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(59))},
				{PropertyPath: "properties.dailySchedule.snapshotsToKeep", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(255))},
				{PropertyPath: "properties.dailySchedule.hour", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(23))},
				{PropertyPath: "properties.dailySchedule.minute", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(59))},
				{PropertyPath: "properties.weeklySchedule.snapshotsToKeep", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(255))},
				{PropertyPath: "properties.weeklySchedule.hour", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(23))},
				{PropertyPath: "properties.weeklySchedule.minute", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(59))},
				{PropertyPath: "properties.monthlySchedule.snapshotsToKeep", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(255))},
				{PropertyPath: "properties.monthlySchedule.hour", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(23))},
				{PropertyPath: "properties.monthlySchedule.minute", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(59))},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
			// NOTE: weekly_schedule.days_of_week (IsDayOfTheWeek) and monthly_schedule.
			// days_of_month (IntBetween 1-30) are TypeSet members serialized into single
			// comma-joined strings (properties.weeklySchedule.day / .monthlySchedule.
			// daysOfMonth), so their element-level validators cannot be expressed here.
			// NOTE: CustomizeDiff ForceNew fires when an existing schedule block is removed.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppSnapshotPolicy()) }
