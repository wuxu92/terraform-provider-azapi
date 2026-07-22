package relay

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RelayHybridConnectionAuthorizationRule provides resource knowledge for
// Microsoft.Relay/namespaces/hybridConnections/authorizationRules.
//
// This is a DISTINCT ARM type from the namespace-scoped
// Microsoft.Relay/namespaces/authorizationRules (see
// relay_namespace_authorization_rule.go).
//
// Contributing Terraform resource: azurerm_relay_hybrid_connection_authorization_rule.
//
// Scoping verified via SDK id parser: the resource builds its ID with
// hybridconnections.NewHybridConnectionAuthorizationRuleID
// (relay_hybrid_connection_authorization_rule_resource.go L72), whose ID() formats
// .../Microsoft.Relay/namespaces/%s/hybridConnections/%s/authorizationRules/%s
// (go-azure-sdk hybridconnections/id_hybridconnectionauthorizationrule.go L116).
//
// Sources:
//   - terraform-provider-azurerm internal/services/relay/relay_hybrid_connection_authorization_rule_resource.go
//     (schema L20-64, Create body L88-93)
//   - terraform-provider-azurerm internal/services/relay/internal.go
//     (listen/send/manage schema + CustomizeDiff L16-142)
//   - go-azure-sdk resource-manager/relay/2021-11-01/hybridconnections:
//     id_hybridconnectionauthorizationrule.go, model_authorizationruleproperties.go,
//     constants.go (AccessRights: Listen/Send/Manage)
//
// Notes:
//   - name/namespace_name/hybrid_connection_name/resource_group_name are envelope /
//     parent-reference fields; not emitted as body rules.
//   - listen/send/manage booleans expand into the properties.rights array of AccessRights
//     enums; a many-to-array mapping not expressible as a single-path rule, so only the
//     required-array constraint is emitted. CustomizeDiff (internal.go L129-142) requires
//     at least one right and manage implies listen+send.
//   - primary_key / secondary_key / *_connection_string are data-plane keys, not body
//     properties; omitted.
type RelayHybridConnectionAuthorizationRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RelayHybridConnectionAuthorizationRule)(nil)

// NewRelayHybridConnectionAuthorizationRule returns knowledge for the
// namespaces/hybridConnections/authorizationRules resource.
func NewRelayHybridConnectionAuthorizationRule() *RelayHybridConnectionAuthorizationRule {
	return &RelayHybridConnectionAuthorizationRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Relay/namespaces/hybridConnections/authorizationRules",
			ApiVersions:  []string{"2021-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.rights",
			},
		},
	}
}

func init() { azwise.Register(NewRelayHybridConnectionAuthorizationRule()) }
