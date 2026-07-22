package desktopvirtualization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualDesktopScalingPlan provides resource knowledge for
// Microsoft.DesktopVirtualization/scalingPlans.
//
// Sources:
//   - internal/services/desktopvirtualization/virtual_desktop_scaling_plan_resource.go
//     (schema + CRUD, lines 30-297)
//   - vendor/.../desktopvirtualization/2025-10-10/scalingplan/model_scalingplanproperties.go
//   - vendor/.../desktopvirtualization/2025-10-10/scalingplan/model_scalingschedule.go
//   - vendor/.../desktopvirtualization/2025-10-10/scalingplan/constants.go
//
// Notes:
//   - hostPoolType is hardcoded to "Pooled" by AzureRM (the only ARM value).
//   - The schedule block maps to properties.schedules[*]; all its constrained
//     fields (daysOfWeek, *LoadBalancingAlgorithm, rampDownStopHostsWhen, and the
//     0-100 ramp percent ints) are array-element paths unsupported by azwise —
//     skipped. ramp_*_start_time / off_peak_start_time use an HH:MM regex which is
//     also an array-element path — skipped.
//   - host_pool block maps to properties.hostPoolReferences[*] (array element) —
//     no declarative rule emitted.
type VirtualDesktopScalingPlan struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualDesktopScalingPlan)(nil)

func NewVirtualDesktopScalingPlan() *VirtualDesktopScalingPlan {
	return &VirtualDesktopScalingPlan{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DesktopVirtualization/scalingPlans",
			ApiVersions:  []string{"2025-10-10"},
			// name (ForceNew) is an envelope field with no ARM body path.
			ForceNew:   []azwise.ForceNewRule{},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.hostPoolType",
					AllowedValues: []string{"Pooled"},
				},
				{PropertyPath: "properties.friendlyName", MinLength: 1, MaxLength: 64},
				{PropertyPath: "properties.description", MinLength: 1, MaxLength: 512},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.hostPoolType", Value: "Pooled"},
			},
			// timeZone is non-omitempty (required) in the ARM model; schedules is
			// required (MinItems 1) in the schema.
			RequiredFields: []string{
				"properties.timeZone",
				"properties.schedules",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualDesktopScalingPlan()) }
