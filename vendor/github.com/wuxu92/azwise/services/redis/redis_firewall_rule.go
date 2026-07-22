package redis

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RedisFirewallRule provides resource knowledge for Microsoft.Cache/Redis/firewallRules.
//
// Mirrors azurerm_redis_firewall_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/redis/redis_firewall_rule_resource.go:28-116
//   - Timeouts (:40-45): Create 30m / Read 5m / Update 30m / Delete 30m
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     start_ip / end_ip (validation.All(IsIPAddress, StringIsNotEmpty) ->
//     properties.startIP / properties.endIP — IP-address checks have no declarative form).
//   - name (validate.FirewallRuleName, regex ^\w+$) is emitted as a name StringRule.
//   - go-azure-sdk resource-manager/redis/2024-11-01/redisfirewallrules
//     model_redisfirewallrule.go / model_redisfirewallruleproperties.go
type RedisFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RedisFirewallRule)(nil)

// NewRedisFirewallRule returns knowledge for the Redis firewall rule resource.
func NewRedisFirewallRule() *RedisFirewallRule {
	return &RedisFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cache/Redis/firewallRules",
			ApiVersions:  []string{"2024-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Both IPs are Required for the rule body. name / redis_cache_name are
			// envelope/path.
			RequiredFields: []string{
				"properties.startIP",
				"properties.endIP",
			},
			StringRules: []azwise.StringRule{
				// name: validate.FirewallRuleName -> alphanumeric + underscore only.
				{
					PropertyPath: "",
					Regex:        `^\w+$`,
					Message:      "firewall rule name may only contain alphanumeric characters and underscores",
				},
			},
		},
	}
}

func init() { azwise.Register(NewRedisFirewallRule()) }
