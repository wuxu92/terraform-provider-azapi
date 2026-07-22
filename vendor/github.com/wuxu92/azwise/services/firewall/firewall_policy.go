package firewall

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FirewallPolicy provides resource knowledge for Microsoft.Network/firewallPolicies.
//
// Mirrors azurerm_firewall_policy. name, location and resource_group_name live on
// the operational envelope, so their ForceNew is envelope-owned; only body-path
// knowledge is encoded.
//
// Sources:
//   - terraform-provider-azurerm internal/services/firewall/firewall_policy_resource.go
//     (resourceFirewallPolicySchema lines 662-1015, resourceFirewallPolicyCreateUpdate
//     lines 60-157, CRUD timeouts lines 49-54)
//   - terraform-provider-azurerm internal/services/firewall/validate/firewall_policy_name.go
//     (FirewallPolicyName regex, line 13)
//   - go-azure-sdk resource-manager/network/2025-01-01/firewallpolicies:
//     model_firewallpolicypropertiesformat.go, model_firewallpolicysku.go,
//     model_dnssettings.go, model_firewallpolicysnat.go, model_firewallpolicysql.go,
//     model_explicitproxy.go, model_firewallpolicyinsights.go,
//     model_firewallpolicyintrusiondetection.go,
//     model_firewallpolicythreatintelwhitelist.go, constants.go (enum values)
//
// Not encoded (deliberate):
//   - intrusion_detection.signature_overrides[*].state and
//     intrusion_detection.traffic_bypass[*].protocol are enum validators living under
//     array-element paths (properties.intrusionDetection.configuration.*[*]); azwise
//     cannot lower a path through an array element, so they are skipped.
//   - tls_certificate expands into properties.transportSecurity.certificateAuthority
//     (a key-vault secret reference); no declarative rule applies.
type FirewallPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FirewallPolicy)(nil)

// NewFirewallPolicy returns knowledge for the firewallPolicies resource.
func NewFirewallPolicy() *FirewallPolicy {
	return &FirewallPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/firewallPolicies",
			ApiVersions:  []string{"2025-01-01"},
			// sku (properties.sku.tier) replaces the policy on change.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.sku.tier"},
			},
			SoftDelete: false,
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
				{
					PropertyPath:  "properties.sku.tier",
					AllowedValues: []string{"Basic", "Premium", "Standard"},
					Message:       "must be Basic, Premium or Standard",
				},
				{
					PropertyPath:  "properties.threatIntelMode",
					AllowedValues: []string{"Alert", "Deny", "Off"},
					Message:       "must be Alert, Deny or Off",
				},
				{
					PropertyPath:  "properties.intrusionDetection.mode",
					AllowedValues: []string{"Alert", "Deny", "Off"},
					Message:       "must be Alert, Deny or Off",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.insights.retentionDays",
					MinValue:     azwise.Ptr(int64(0)),
				},
				{
					PropertyPath: "properties.explicitProxy.httpPort",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(35536)),
				},
				{
					PropertyPath: "properties.explicitProxy.httpsPort",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(35536)),
				},
				{
					PropertyPath: "properties.explicitProxy.pacFilePort",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(35536)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.sku.tier", Value: "Standard"},
				{PropertyPath: "properties.threatIntelMode", Value: "Alert"},
				{PropertyPath: "properties.dnsSettings.enableProxy", Value: false},
			},
			// childPolicies, firewalls, ruleCollectionGroups, size and
			// provisioningState are populated by Azure and never set by the user.
			ComputedFields: []string{
				"properties.childPolicies",
				"properties.firewalls",
				"properties.ruleCollectionGroups",
				"properties.size",
				"properties.provisioningState",
			},
			// threat_intelligence_allowlist requires at least one of ip_addresses /
			// fqdns; the block is MaxItems:1 so it maps to object-level ARM paths.
			AtLeastOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.threatIntelWhitelist.ipAddresses",
						"properties.threatIntelWhitelist.fqdns",
					},
					Message: "at least one of ip_addresses or fqdns must be set in threat_intelligence_allowlist",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewFirewallPolicy()) }
