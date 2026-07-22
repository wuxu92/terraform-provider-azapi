// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusNamespaceAuthorizationRule provides resource knowledge for
// Microsoft.ServiceBus/namespaces/authorizationRules.
//
// Contributing Terraform resource: azurerm_servicebus_namespace_authorization_rule.
//
// Scoping verified via SDK id parser: AuthorizationRuleId.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s/authorizationRules/%s
// (namespacesauthorizationrule/id_authorizationrule.go L110-111).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_namespace_authorization_rule_resource.go
//     (schema L53-69, Create body L98-103)
//   - terraform-provider-azurerm internal/services/servicebus/internal.go
//     (listen/send/manage schema L60-107, expandAuthorizationRuleRights L21-37, CustomizeDiff L109-123)
//   - terraform-provider-azurerm internal/services/servicebus/validate/authorization_rule_name.go (L13-18)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/namespacesauthorizationrule:
//     model_sbauthorizationruleproperties.go (Rights []AccessRights), constants.go (AccessRights: Listen/Manage/Send)
//
// Notes:
//   - name/namespace_id are envelope / parent-reference fields; not emitted as body rules.
//   - listen/send/manage booleans expand into the properties.rights array of AccessRights
//     enums (internal.go L21-37); a many-to-array mapping not expressible as a single-path
//     rule, so only the required-array constraint is emitted. CustomizeDiff (L109-123)
//     requires at least one right, and manage implies both listen and send.
//   - primary_key / secondary_key / *_connection_string / *_connection_string_alias are
//     data-plane keys (ListKeys), not body properties; omitted.
type ServiceBusNamespaceAuthorizationRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusNamespaceAuthorizationRule)(nil)

// NewServiceBusNamespaceAuthorizationRule returns knowledge for the
// namespaces/authorizationRules resource.
func NewServiceBusNamespaceAuthorizationRule() *ServiceBusNamespaceAuthorizationRule {
	return &ServiceBusNamespaceAuthorizationRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces/authorizationRules",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        "^[a-zA-Z0-9][-._a-zA-Z0-9]{0,48}([a-zA-Z0-9])?$",
					Message:      "authorization rule name can contain only letters, numbers, periods, hyphens and underscores, must start and end with a letter or number, and be at most 50 characters long",
				},
			},
			RequiredFields: []string{
				"properties.rights",
			},
		},
	}
}

func init() { azwise.Register(NewServiceBusNamespaceAuthorizationRule()) }
