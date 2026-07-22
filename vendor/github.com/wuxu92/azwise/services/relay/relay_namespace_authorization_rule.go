package relay

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RelayNamespaceAuthorizationRule provides resource knowledge for
// Microsoft.Relay/namespaces/authorizationRules.
//
// This is a DISTINCT ARM type from the hybrid-connection-scoped
// Microsoft.Relay/namespaces/hybridConnections/authorizationRules (see
// relay_hybrid_connection_authorization_rule.go).
//
// Contributing Terraform resource: azurerm_relay_namespace_authorization_rule.
//
// Scoping verified via SDK id parser: the resource builds its ID with
// namespaces.NewAuthorizationRuleID (relay_namespace_authorization_rule_resource.go L66),
// whose ID() formats .../Microsoft.Relay/namespaces/%s/authorizationRules/%s
// (go-azure-sdk namespaces/id_authorizationrule.go L110).
//
// Sources:
//   - terraform-provider-azurerm internal/services/relay/relay_namespace_authorization_rule_resource.go
//     (schema L20-58, Create body L82-87)
//   - terraform-provider-azurerm internal/services/relay/internal.go
//     (listen/send/manage schema + CustomizeDiff L16-142)
//   - go-azure-sdk resource-manager/relay/2021-11-01/namespaces:
//     id_authorizationrule.go, model_authorizationruleproperties.go,
//     constants.go (AccessRights: Listen/Send/Manage)
//
// Notes:
//   - name/namespace_name/resource_group_name are envelope / parent-reference fields;
//     not emitted as body rules.
//   - listen/send/manage booleans expand into the properties.rights array of AccessRights
//     enums; a many-to-array mapping not expressible as a single-path rule, so only the
//     required-array constraint is emitted. CustomizeDiff (internal.go L129-142) requires
//     at least one right and manage implies listen+send.
//   - primary_key / secondary_key / *_connection_string are data-plane keys, not body
//     properties; omitted.
type RelayNamespaceAuthorizationRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RelayNamespaceAuthorizationRule)(nil)

// NewRelayNamespaceAuthorizationRule returns knowledge for the
// namespaces/authorizationRules resource.
func NewRelayNamespaceAuthorizationRule() *RelayNamespaceAuthorizationRule {
	return &RelayNamespaceAuthorizationRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Relay/namespaces/authorizationRules",
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

func init() { azwise.Register(NewRelayNamespaceAuthorizationRule()) }
