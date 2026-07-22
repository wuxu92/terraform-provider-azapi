package mongocluster

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MongoClusterFirewallRule provides resource knowledge for
// Microsoft.DocumentDB/mongoClusters/firewallRules
// (azurerm_mongo_cluster_firewall_rule).
//
// Sources:
//   - AzureRM internal/services/mongocluster/mongo_cluster_firewall_rule_resource.go
//     :45-53  (name StringMatch regex, 1-80 chars)
//     :55     (mongo_cluster_id parent reference, ForceNew)
//     :57-67  (end_ip_address / start_ip_address, Required, IsIPv4Address)
//     :105-110 (Create: ARM body mapping startIpAddress / endIpAddress)
//     :77,125,161,209 (timeouts: create 30m, read 5m, update 30m, delete 30m)
//   - go-azure-sdk resource-manager/mongocluster/2025-09-01/firewallrules:
//     model_firewallruleproperties.go (ARM body paths endIpAddress / startIpAddress,
//     both non-omitempty required; provisioningState read-only)
//
// Not encoded (deliberate):
//   - start_ip_address / end_ip_address use validation.IsIPv4Address, a semantic
//     validator that is not a clean regex/enum/length constraint. It cannot be
//     expressed as a declarative StringRule; it belongs in an azapin customizer
//     validator (properties.startIpAddress / properties.endIpAddress), handled
//     separately from this knowledge file.
//   - name and mongo_cluster_id are ForceNew in AzureRM but both are envelope /
//     parent-reference fields with no place in the resource body, so ForceNew is
//     empty.
type MongoClusterFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MongoClusterFirewallRule)(nil)

// NewMongoClusterFirewallRule returns knowledge for the firewallRules resource.
func NewMongoClusterFirewallRule() *MongoClusterFirewallRule {
	return &MongoClusterFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/mongoClusters/firewallRules",
			ApiVersions:  []string{"2025-09-01"},
			SoftDelete:   false,
			ForceNew:     []azwise.ForceNewRule{},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: 1-80 chars, alphanumeric start, then
					// alphanumeric / dot / hyphen / underscore.
					Regex:     `^[a-zA-Z0-9][a-zA-Z0-9.\-_]{0,79}$`,
					MinLength: 1,
					MaxLength: 80,
					Message:   "`name` must be between 1 and 80 characters, start with an alphanumeric character and contain only alphanumeric characters, dots, hyphens and underscores",
				},
			},
			SensitiveFields: []string{},
			// provisioningState is response-only (absent from the create model).
			ComputedFields: []string{
				"properties.provisioningState",
			},
			DefaultValues: []azwise.DefaultValue{},
			RequiredFields: []string{
				"properties.startIpAddress",
				"properties.endIpAddress",
			},
		},
	}
}

func init() { azwise.Register(NewMongoClusterFirewallRule()) }
