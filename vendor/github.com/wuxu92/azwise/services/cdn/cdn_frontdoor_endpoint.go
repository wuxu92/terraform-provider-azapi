package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorEndpoint provides resource knowledge for
// Microsoft.Cdn/profiles/afdEndpoints (Azure Front Door endpoint).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_endpoint.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_endpoint_resource.go
//     (schema L43-70, Create body L100-113)
//   - internal/services/cdn/validate/front_door_endpoint_name.go (FrontDoorEndpointName regex)
//   - go-azure-sdk resource-manager/cdn/2025-12-01/afdendpoints:
//     model_afdendpointproperties.go, constants.go (EnabledState enum).
//
// Notes:
//   - name & cdn_frontdoor_profile_id are envelope/parent-owned (ForceNew), not
//     ARM-body ForceNew rules.
//   - enabled (bool, default true) maps to properties.enabledState ("Enabled"/"Disabled");
//     the default is expressed in ARM enum form.
type CdnFrontDoorEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorEndpoint)(nil)

func NewCdnFrontDoorEndpoint() *CdnFrontDoorEndpoint {
	return &CdnFrontDoorEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/afdEndpoints",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorEndpointName
				{
					Regex:     `^[\da-zA-Z][-\da-zA-Z]{0,44}[\da-zA-Z]$`,
					MinLength: 2,
					MaxLength: 46,
					Message:   "must be 2-46 characters, begin and end with a letter or number, and contain only letters, numbers and hyphens",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled default true → properties.enabledState
				{PropertyPath: "properties.enabledState", Value: "Enabled"},
			},
			ComputedFields: []string{
				"properties.hostName",
				"properties.deploymentStatus",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorEndpoint()) }
