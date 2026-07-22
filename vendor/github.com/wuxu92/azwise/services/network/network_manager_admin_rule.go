package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerAdminRule provides resource knowledge for
// Microsoft.Network/networkManagers/securityAdminConfigurations/ruleCollections/rules.
//
// Mirrors azurerm_network_manager_admin_rule. name and the parent
// admin_rule_collection_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_admin_rule_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/adminrules:
//     model_adminpropertiesformat.go, constants.go
//
// Not encoded (deliberate):
//   - source.address_prefix_type / destination.address_prefix_type are array-element enums
//     (sources[*] / destinations[*]); azwise cannot lower a path through an array element,
//     so their IPPrefix/ServiceTag/NetworkGroup enum is documented, not emitted.
type NetworkManagerAdminRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerAdminRule)(nil)

func NewNetworkManagerAdminRule() *NetworkManagerAdminRule {
	return &NetworkManagerAdminRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/securityAdminConfigurations/ruleCollections/rules",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.access",
					AllowedValues: []string{"Allow", "AlwaysAllow", "Deny"},
					Message:       "must be Allow, AlwaysAllow or Deny",
				},
				{
					PropertyPath:  "properties.direction",
					AllowedValues: []string{"Inbound", "Outbound"},
					Message:       "must be Inbound or Outbound",
				},
				{
					PropertyPath:  "properties.protocol",
					AllowedValues: []string{"Ah", "Any", "Esp", "Icmp", "Tcp", "Udp"},
					Message:       "must be one of Ah, Any, Esp, Icmp, Tcp, Udp",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.priority",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(4096)),
				},
			},
			RequiredFields: []string{
				"properties.access",
				"properties.direction",
				"properties.priority",
				"properties.protocol",
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerAdminRule()) }
