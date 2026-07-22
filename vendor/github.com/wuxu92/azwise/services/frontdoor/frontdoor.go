package frontdoor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FrontDoor provides resource knowledge for Microsoft.Network/frontDoors
// (classic Azure Front Door — deprecated for new creation but still managed).
//
// Contributing Terraform resource: azurerm_frontdoor.
//
// Sources:
//   - terraform-provider-azurerm internal/services/frontdoor/frontdoor_resource.go
//     (resource L31-60, schema L1723-2217, create/update L62-239)
//   - internal/services/frontdoor/validate/front_door_name.go (name regex)
//   - go-azure-sdk resource-manager/frontdoor/2020-05-01/frontdoors:
//     model_frontdoor.go, model_frontdoorproperties.go, constants.go
//     (FrontDoorEnabledState / FrontDoorResourceState).
//
// Notes:
//   - name & resource_group_name are envelope-owned (both ForceNew); frontDoors has no
//     location field (the service is Global), so ForceNew is empty here.
//   - load_balancer_enabled (bool, default true) maps to properties.enabledState
//     (Enabled/Disabled); friendly_name maps to properties.friendlyName.
//   - routing_rule / backend_pool_load_balancing / backend_pool_health_probe /
//     backend_pool / frontend_endpoint are required arrays mapping to
//     properties.routingRules / .loadBalancingSettings / .healthProbeSettings /
//     .backendPools / .frontendEndpoints; their per-element enums (accepted protocols,
//     redirect/forwarding config, probe protocol/method, backend port/weight/priority
//     ranges) are array-element sub-trees, not expressible as scalar declarative rules.
type FrontDoor struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FrontDoor)(nil)

func NewFrontDoor() *FrontDoor {
	return &FrontDoor{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/frontDoors",
			ApiVersions:  []string{"2020-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 6 * time.Hour,
				Read:   5 * time.Minute,
				Update: 6 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorName
				{
					Regex:     `(^[\da-zA-Z])([-\da-zA-Z]{3,61})([\da-zA-Z]$)`,
					MinLength: 5,
					MaxLength: 63,
					Message:   "must be 5-63 characters, begin and end with a letter or number, and contain only letters, numbers or hyphens",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// load_balancer_enabled default true → properties.enabledState
				{PropertyPath: "properties.enabledState", Value: "Enabled"},
			},
			ComputedFields: []string{
				"properties.cname",
				"properties.frontdoorId",
				"properties.provisioningState",
				"properties.resourceState",
			},
			// The five block arrays are Required: true in the schema.
			RequiredFields: []string{
				"properties.routingRules",
				"properties.loadBalancingSettings",
				"properties.healthProbeSettings",
				"properties.backendPools",
				"properties.frontendEndpoints",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewFrontDoor()) }
