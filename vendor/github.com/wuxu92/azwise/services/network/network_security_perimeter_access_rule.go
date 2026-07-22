package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkSecurityPerimeterAccessRule provides resource knowledge for
// Microsoft.Network/networkSecurityPerimeters/profiles/accessRules.
//
// Mirrors azurerm_network_security_perimeter_access_rule. name, direction and
// network_security_perimeter_profile_id are ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_security_perimeter_access_rule_resource.go
//     (typed schema, CustomizeDiff direction/address/fqdn coupling, Create 30m)
//   - go-azure-sdk resource-manager/network/2025-01-01/networksecurityperimeteraccessrules:
//     model_nspaccessruleproperties.go, constants.go (AccessRuleDirection),
//     id_accessrule.go (segment casing
//     "networkSecurityPerimeters/profiles/accessRules")
//
// Not encoded (deliberate):
//   - subscription_ids maps to properties.subscriptions[*].id (SubscriptionId
//     objects), an array-element path azwise/azapin cannot lower; the ExactlyOneOf
//     below therefore uses the settable top-level "properties.subscriptions" array
//     itself as the discriminating path.
//   - The AzureRM CustomizeDiff coupling (address_prefixes forbidden when
//     direction=Outbound; fqdns forbidden when direction=Inbound) is a
//     value-conditional cross-field rule with no ConflictsWith/RequiredWith
//     equivalent; it is enforced by AzureRM's diff and left to the API otherwise.
type NetworkSecurityPerimeterAccessRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkSecurityPerimeterAccessRule)(nil)

// NewNetworkSecurityPerimeterAccessRule returns knowledge for the
// networkSecurityPerimeters/profiles/accessRules resource.
func NewNetworkSecurityPerimeterAccessRule() *NetworkSecurityPerimeterAccessRule {
	return &NetworkSecurityPerimeterAccessRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkSecurityPerimeters/profiles/accessRules",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.direction"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:   `(^[a-zA-Z0-9]+[a-zA-Z0-9_.-]{0,78}[a-zA-Z0-9_]+$)|(^[a-zA-Z0-9]$)`,
					Message: "name must be 1-80 chars, start with a letter or number, end with a letter, number or underscore, and contain only letters, numbers, underscores, periods, or hyphens",
				},
				{
					PropertyPath:  "properties.direction",
					AllowedValues: []string{"Inbound", "Outbound"},
					Message:       "direction must be Inbound or Outbound",
				},
			},
			RequiredFields: []string{
				"properties.direction",
			},
			// AzureRM ExactlyOneOf over address_prefixes / fqdns / service_tags /
			// subscription_ids; each maps to a distinct top-level ARM array.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.addressPrefixes",
						"properties.fullyQualifiedDomainNames",
						"properties.serviceTags",
						"properties.subscriptions",
					},
					Message: "exactly one of address_prefixes, fqdns, service_tags or subscription_ids must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkSecurityPerimeterAccessRule()) }
