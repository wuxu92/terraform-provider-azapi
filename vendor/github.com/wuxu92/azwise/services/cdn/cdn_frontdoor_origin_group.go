package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorOriginGroup provides resource knowledge for
// Microsoft.Cdn/profiles/originGroups (Azure Front Door origin group).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_origin_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_origin_group_resource.go
//     (schema L42-143, Create body L172-179)
//   - internal/services/cdn/validate/front_door_origin_group_name.go (name regex)
//   - go-azure-sdk resource-manager/cdn/2025-12-01/afdorigingroups:
//     model_afdorigingroupproperties.go, model_healthprobeparameters.go,
//     model_loadbalancingsettingsparameters.go, constants.go
//     (ProbeProtocol / HealthProbeRequestType enums).
//
// Notes:
//   - name & cdn_frontdoor_profile_id are envelope/parent-owned (ForceNew), not body rules.
//   - session_affinity_enabled (bool, default true) maps to
//     properties.sessionAffinityState ("Enabled"/"Disabled").
//   - ProbeProtocol / HealthProbeRequestType use the full ARM SDK enum (includes
//     "NotSet"); AzureRM narrows to Http/Https and GET/HEAD.
type CdnFrontDoorOriginGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorOriginGroup)(nil)

func NewCdnFrontDoorOriginGroup() *CdnFrontDoorOriginGroup {
	return &CdnFrontDoorOriginGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/originGroups",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorOriginGroupName
				{
					Regex:     `^[\da-zA-Z][-\da-zA-Z]{0,88}[\da-zA-Z]$`,
					MinLength: 2,
					MaxLength: 90,
					Message:   "must be 2-90 characters, begin and end with a letter or number, and contain only letters, numbers and hyphens",
				},
				// health_probe.protocol → properties.healthProbeSettings.probeProtocol
				{
					PropertyPath:  "properties.healthProbeSettings.probeProtocol",
					AllowedValues: []string{"Http", "Https", "NotSet"},
				},
				// health_probe.request_type → properties.healthProbeSettings.probeRequestType
				{
					PropertyPath:  "properties.healthProbeSettings.probeRequestType",
					AllowedValues: []string{"GET", "HEAD", "NotSet"},
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.loadBalancingSettings.additionalLatencyInMilliseconds", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(1000))},
				{PropertyPath: "properties.loadBalancingSettings.sampleSize", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(255))},
				{PropertyPath: "properties.loadBalancingSettings.successfulSamplesRequired", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(255))},
				{PropertyPath: "properties.healthProbeSettings.probeIntervalInSeconds", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(255))},
				{PropertyPath: "properties.trafficRestorationTimeToHealedOrNewEndpointsInMinutes", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(50))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.loadBalancingSettings.additionalLatencyInMilliseconds", Value: float64(50)},
				{PropertyPath: "properties.loadBalancingSettings.sampleSize", Value: float64(4)},
				{PropertyPath: "properties.loadBalancingSettings.successfulSamplesRequired", Value: float64(3)},
				{PropertyPath: "properties.healthProbeSettings.probeRequestType", Value: "HEAD"},
				{PropertyPath: "properties.healthProbeSettings.probePath", Value: "/"},
				{PropertyPath: "properties.sessionAffinityState", Value: "Enabled"},
				{PropertyPath: "properties.trafficRestorationTimeToHealedOrNewEndpointsInMinutes", Value: float64(10)},
			},
			ComputedFields: []string{
				"properties.deploymentStatus",
				"properties.provisioningState",
				"properties.profileName",
			},
			// load_balancing block is Required (its sub-fields are all optional+defaulted).
			RequiredFields: []string{"properties.loadBalancingSettings"},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorOriginGroup()) }
