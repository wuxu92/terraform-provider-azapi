package firewall

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FirewallPolicyRuleCollectionGroup provides resource knowledge for
// Microsoft.Network/firewallPolicies/ruleCollectionGroups.
//
// Mirrors azurerm_firewall_policy_rule_collection_group. name and
// firewall_policy_id are envelope/parent references (ForceNew owned by the
// envelope); only body-path knowledge is encoded.
//
// Sources:
//   - terraform-provider-azurerm internal/services/firewall/firewall_policy_rule_collection_group_resource.go
//     (resourceFirewallPolicyRuleCollectionGroup schema lines 31-460,
//     CreateUpdate lines 462-525, CRUD timeouts lines 44-49)
//   - terraform-provider-azurerm internal/services/firewall/validate/firewall_policy_rule_collection_group_name.go
//     (name regex, line 13)
//   - go-azure-sdk resource-manager/network/2025-01-01/firewallpolicyrulecollectiongroups:
//     model_firewallpolicyrulecollectiongroupproperties.go, constants.go
//
// Not encoded (deliberate):
//   - Every application_rule_collection / network_rule_collection /
//     nat_rule_collection entry and its rules (action enums, per-collection and
//     per-rule priority IntBetween(100,65000), protocol type/port validators) lives
//     under properties.ruleCollections[*] array elements; azwise cannot lower a path
//     through an array element, so these are skipped.
type FirewallPolicyRuleCollectionGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FirewallPolicyRuleCollectionGroup)(nil)

// NewFirewallPolicyRuleCollectionGroup returns knowledge for the
// firewallPolicies/ruleCollectionGroups resource.
func NewFirewallPolicyRuleCollectionGroup() *FirewallPolicyRuleCollectionGroup {
	return &FirewallPolicyRuleCollectionGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/firewallPolicies/ruleCollectionGroups",
			ApiVersions:  []string{"2025-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// resource name (empty PropertyPath = name attribute)
					Regex:   `^[^\W_][\w-.]*[\w]$`,
					Message: "must begin with a letter or number, end with a letter, number or underscore, and may contain only letters, numbers, underscores, periods, or hyphens",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.priority",
					MinValue:     azwise.Ptr(int64(100)),
					MaxValue:     azwise.Ptr(int64(65000)),
				},
			},
			// size and provisioningState are populated by Azure and never set by
			// the user.
			ComputedFields: []string{
				"properties.size",
				"properties.provisioningState",
			},
			// priority is Required in AzureRM.
			RequiredFields: []string{
				"properties.priority",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewFirewallPolicyRuleCollectionGroup()) }
