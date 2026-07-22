package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorFirewallPolicy provides resource knowledge for
// Microsoft.Network/frontDoorWebApplicationFirewallPolicies (Azure Front Door WAF policy).
//
// Contributing Terraform resources:
//   - azurerm_cdn_frontdoor_firewall_policy (AFD Standard/Premium, this file's version)
//   - azurerm_frontdoor_firewall_policy (classic Front Door WAF, API 2020-04-01)
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_firewall_policy_resource.go
//     (schema L48-540)
//   - internal/services/cdn/validate/front_door_firewall_policy_name.go (name regex)
//   - go-azure-sdk resource-manager/frontdoor/2025-03-01/webapplicationfirewallpolicies:
//     model_webapplicationfirewallpolicy.go (top-level Sku), model_policysettings.go,
//     model_webapplicationfirewallpolicyproperties.go, constants.go
//     (SkuName / PolicyMode / PolicyEnabledState / PolicyRequestBodyCheck).
//   - classic: internal/services/frontdoor/frontdoor_firewall_policy_resource.go
//     (schema L30-449, create L451-525) + go-azure-sdk
//     resource-manager/frontdoor/2020-04-01/webapplicationfirewallpolicies/constants.go.
//     No new universal declarative rules: mode (Detection/Prevention) and enabledState
//     (Enabled/Disabled) already match; the classic version has no sku, mode is
//     optional (default Prevention) rather than required, and its
//     custom_block_response_status_code is an IntInSlice {200,403,405,406,429} that is
//     not expressible as an azwise IntRule range — none are unioned to avoid corrupting
//     the AFD Standard/Premium shape.
//
// Notes:
//   - name & resource_group_name are envelope-owned; sku_name (ForceNew) maps to the
//     top-level sku.name.
//   - enabled and request_body_check_enabled are bools mapped to Enabled/Disabled
//     ARM enums (properties.policySettings.enabledState / .requestBodyCheck).
//   - redirect_url (IsURLWithScheme http/https) and custom_block_response_body
//     (StringIsBase64) are semantic validators ported at the azapin layer.
//   - custom_rule -> properties.customRules.rules[*] and managed_rule ->
//     properties.managedRules.managedRuleSets[*] are array-element sub-trees; their
//     per-element enums/ranges are not expressible as scalar declarative rules here.
type CdnFrontDoorFirewallPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorFirewallPolicy)(nil)

func NewCdnFrontDoorFirewallPolicy() *CdnFrontDoorFirewallPolicy {
	return &CdnFrontDoorFirewallPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/frontDoorWebApplicationFirewallPolicies",
			ApiVersions:  []string{"2025-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "sku.name"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorFirewallPolicyName
				{
					Regex:     `^[a-zA-Z][\da-zA-Z]{0,127}$`,
					MinLength: 1,
					MaxLength: 128,
					Message:   "must be 1-128 characters, begin with a letter, and contain only letters and numbers",
				},
				// sku_name → sku.name
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Standard_AzureFrontDoor", "Premium_AzureFrontDoor"},
				},
				// mode → properties.policySettings.mode
				{
					PropertyPath:  "properties.policySettings.mode",
					AllowedValues: []string{"Detection", "Prevention"},
				},
			},
			IntRules: []azwise.IntRule{
				// js_challenge_cookie_expiration_in_minutes → ...javascriptChallengeExpirationInMinutes
				{PropertyPath: "properties.policySettings.javascriptChallengeExpirationInMinutes", MinValue: azwise.Ptr(int64(5)), MaxValue: azwise.Ptr(int64(1440))},
				// captcha_cookie_expiration_in_minutes → ...captchaExpirationInMinutes
				{PropertyPath: "properties.policySettings.captchaExpirationInMinutes", MinValue: azwise.Ptr(int64(5)), MaxValue: azwise.Ptr(int64(1440))},
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled default true → policySettings.enabledState
				{PropertyPath: "properties.policySettings.enabledState", Value: "Enabled"},
				// request_body_check_enabled default true → policySettings.requestBodyCheck
				{PropertyPath: "properties.policySettings.requestBodyCheck", Value: "Enabled"},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.resourceState",
				"properties.frontendEndpointLinks",
				"properties.routingRuleLinks",
				"properties.securityPolicyLinks",
			},
			// mode and sku.name are Required.
			RequiredFields: []string{
				"sku.name",
				"properties.policySettings.mode",
			},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorFirewallPolicy()) }
