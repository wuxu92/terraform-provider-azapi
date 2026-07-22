package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebApplicationFirewallPolicy provides resource knowledge for
// Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies.
//
// Mirrors azurerm_web_application_firewall_policy. name and resource_group are
// envelope-owned; location is the only replacement-forcing envelope path.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/web_application_firewall_policy_resource.go
//     (schema, expandWebApplicationFirewallPolicyPolicySettings, CRUD timeouts 30/5/30/30)
//   - go-azure-sdk resource-manager/network/2025-01-01/webapplicationfirewallpolicies:
//     model_policysettings.go, model_webapplicationfirewallpolicypropertiesformat.go,
//     id_applicationgatewaywebapplicationfirewallpolicy.go (segment casing
//     "applicationGatewayWebApplicationFirewallPolicies"), constants.go
//
// Not encoded (deliberate):
//   - custom_rules and managed_rules expand into properties.customRules[*] and
//     properties.managedRules.* arrays; every rule/exclusion enum (action,
//     match_variable, operator, rule_type, transform, exclusion match variable /
//     operator) lives under array elements which azwise/azapin cannot lower
//     through "[*]". Skipped, not emitted.
//   - policy_settings is a single object (MaxItems 1) mapping to
//     properties.policySettings, so its scalar constraints ARE representable and
//     are encoded below.
type WebApplicationFirewallPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebApplicationFirewallPolicy)(nil)

// NewWebApplicationFirewallPolicy returns knowledge for the
// applicationGatewayWebApplicationFirewallPolicies resource.
func NewWebApplicationFirewallPolicy() *WebApplicationFirewallPolicy {
	return &WebApplicationFirewallPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.policySettings.mode",
					AllowedValues: []string{"Detection", "Prevention"},
					Message:       "policy_settings mode must be Detection or Prevention",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.policySettings.fileUploadLimitInMb",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(4000)),
					Message:      "file_upload_limit_in_mb must be between 1 and 4000",
				},
				{
					PropertyPath: "properties.policySettings.maxRequestBodySizeInKb",
					MinValue:     azwise.Ptr(int64(8)),
					MaxValue:     azwise.Ptr(int64(2000)),
					Message:      "max_request_body_size_in_kb must be between 8 and 2000",
				},
				{
					PropertyPath: "properties.policySettings.requestBodyInspectLimitInKB",
					MinValue:     azwise.Ptr(int64(0)),
					Message:      "request_body_inspect_limit_in_kb must be at least 0",
				},
				{
					PropertyPath: "properties.policySettings.jsChallengeCookieExpirationInMins",
					MinValue:     azwise.Ptr(int64(5)),
					MaxValue:     azwise.Ptr(int64(1440)),
					Message:      "js_challenge_cookie_expiration_in_minutes must be between 5 and 1440",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.policySettings.mode", Value: "Prevention"},
				{PropertyPath: "properties.policySettings.requestBodyCheck", Value: true},
				{PropertyPath: "properties.policySettings.fileUploadLimitInMb", Value: float64(100)},
				{PropertyPath: "properties.policySettings.maxRequestBodySizeInKb", Value: float64(128)},
				{PropertyPath: "properties.policySettings.requestBodyInspectLimitInKB", Value: float64(128)},
				{PropertyPath: "properties.policySettings.jsChallengeCookieExpirationInMins", Value: float64(30)},
			},
		},
	}
}

func init() { azwise.Register(NewWebApplicationFirewallPolicy()) }
