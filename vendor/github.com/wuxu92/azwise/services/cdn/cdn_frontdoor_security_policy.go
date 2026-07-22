package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorSecurityPolicy provides resource knowledge for
// Microsoft.Cdn/profiles/securityPolicies (Azure Front Door security policy).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_security_policy.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_security_policy_resource.go
//     (schema L47-135, Create body L187-191)
//   - internal/services/cdn/validate/front_door_security_policy_name.go (name regex)
//   - go-azure-sdk resource-manager/cdn/2025-12-01/securitypolicies.
//
// Notes:
//   - name & cdn_frontdoor_profile_id are envelope/parent-owned (ForceNew).
//   - security_policies (Required) maps to the polymorphic properties.parameters
//     (WebApplicationFirewall type). Nested firewall.cdn_frontdoor_firewall_policy_id,
//     association.domain[*].cdn_frontdoor_domain_id, and patterns_to_match are all
//     array-element / polymorphic paths — validated by semantic validators
//     (FrontDoorFirewallPolicyID, FrontDoorSecurityPolicyDomainID) at the azapin
//     layer, and patterns_to_match is a per-element enum ("/*"). None are expressible
//     as scalar declarative rules, so none are emitted here.
type CdnFrontDoorSecurityPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorSecurityPolicy)(nil)

func NewCdnFrontDoorSecurityPolicy() *CdnFrontDoorSecurityPolicy {
	return &CdnFrontDoorSecurityPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/securityPolicies",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorSecurityPolicyName
				{
					Regex:     `^[\da-zA-Z](?:[-\da-zA-Z]*[\da-zA-Z])?$`,
					MinLength: 1,
					Message:   "must begin and end with an alphanumeric character and may contain only alphanumeric characters and hyphens",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.deploymentStatus",
				"properties.profileName",
			},
			RequiredFields: []string{"properties.parameters"},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorSecurityPolicy()) }
