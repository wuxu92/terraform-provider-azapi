package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorRule provides resource knowledge for
// Microsoft.Cdn/profiles/ruleSets/rules (Azure Front Door rule).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_rule_resource.go
//     (schema L72-641, Create body L701-709)
//   - internal/services/cdn/validate/front_door_validation_helpers.go (CdnFrontDoorRuleName)
//   - go-azure-sdk resource-manager/cdn/2025-12-01/rules:
//     model_ruleproperties.go, constants.go (MatchProcessingBehavior).
//
// Notes:
//   - name & cdn_frontdoor_rule_set_id are envelope/parent-owned (ForceNew).
//   - actions -> properties.actions and conditions -> properties.conditions are
//     deeply polymorphic delivery-rule arrays (each element discriminated by name);
//     their sub-fields are array-element paths not expressible as scalar declarative
//     rules, so none are emitted here (they are validated by the AzureRM
//     expand/flatten logic and, where reusable, semantic azapin validators).
type CdnFrontDoorRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorRule)(nil)

func NewCdnFrontDoorRule() *CdnFrontDoorRule {
	return &CdnFrontDoorRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/ruleSets/rules",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.CdnFrontDoorRuleName
				{
					Regex:     `^[a-zA-Z][\da-zA-Z]{0,259}$`,
					MinLength: 1,
					MaxLength: 260,
					Message:   "must be 1-260 characters, begin with a letter, and contain only letters and numbers",
				},
				// behavior_on_match → properties.matchProcessingBehavior
				{
					PropertyPath:  "properties.matchProcessingBehavior",
					AllowedValues: []string{"Continue", "Stop"},
				},
			},
			IntRules: []azwise.IntRule{
				// order → properties.order (IntAtLeast 0)
				{PropertyPath: "properties.order", MinValue: azwise.Ptr(int64(0))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.matchProcessingBehavior", Value: "Continue"},
			},
			ComputedFields: []string{
				"properties.ruleSetName",
				"properties.provisioningState",
				"properties.deploymentStatus",
			},
			RequiredFields: []string{
				"properties.order",
				"properties.actions",
			},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorRule()) }
