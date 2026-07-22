package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorRuleSet provides resource knowledge for
// Microsoft.Cdn/profiles/ruleSets (Azure Front Door rule set).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_rule_set.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_rule_set_resource.go
//     (schema L62-76, Create body L105 — empty RuleSet{})
//   - internal/services/cdn/validate/front_door_rule_set_name.go (name regex)
//   - go-azure-sdk resource-manager/cdn/2025-12-01/rulesets.
//
// Notes:
//   - name & cdn_frontdoor_profile_id are envelope/parent-owned (ForceNew). The rule
//     set carries no user-settable body properties (create sends an empty object).
//   - properties.batchMode distinguishes batch-managed rule sets
//     (azurerm_cdn_frontdoor_batch_rule_set); it is server-managed here.
type CdnFrontDoorRuleSet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorRuleSet)(nil)

func NewCdnFrontDoorRuleSet() *CdnFrontDoorRuleSet {
	return &CdnFrontDoorRuleSet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/ruleSets",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorRuleSetName
				{
					Regex:     `^[a-zA-Z][\da-zA-Z]{0,59}$`,
					MinLength: 1,
					MaxLength: 60,
					Message:   "must be 1-60 characters, begin with a letter, and contain only letters and numbers",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.deploymentStatus",
				"properties.profileName",
				"properties.batchMode",
			},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorRuleSet()) }
