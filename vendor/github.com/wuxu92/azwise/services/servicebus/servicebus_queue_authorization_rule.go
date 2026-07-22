// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusQueueAuthorizationRule provides resource knowledge for
// Microsoft.ServiceBus/namespaces/queues/authorizationRules.
//
// Contributing Terraform resource: azurerm_servicebus_queue_authorization_rule.
//
// Scoping verified via SDK id parser: QueueAuthorizationRuleId.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s/queues/%s/authorizationRules/%s
// (queuesauthorizationrule/id_queueauthorizationrule.go L116-117).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_queue_authorization_rule_resource.go
//     (schema L48-65, Create body via expandAuthorizationRuleRights)
//   - terraform-provider-azurerm internal/services/servicebus/internal.go
//     (listen/send/manage schema L60-107, expandAuthorizationRuleRights L21-37, CustomizeDiff L109-123)
//   - terraform-provider-azurerm internal/services/servicebus/validate/authorization_rule_name.go (L13-18)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/queuesauthorizationrule:
//     model_sbauthorizationruleproperties.go (Rights []AccessRights), constants.go (AccessRights: Listen/Manage/Send)
//
// Notes:
//   - name/queue_id are envelope / parent-reference fields; not emitted as body rules.
//   - listen/send/manage booleans expand into the properties.rights array of AccessRights
//     enums; a many-to-array mapping not expressible as a single-path rule, so only the
//     required-array constraint is emitted. CustomizeDiff requires at least one right, and
//     manage implies both listen and send.
//   - primary_key / secondary_key / *_connection_string are data-plane keys (ListKeys), not
//     body properties; omitted.
type ServiceBusQueueAuthorizationRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusQueueAuthorizationRule)(nil)

// NewServiceBusQueueAuthorizationRule returns knowledge for the
// namespaces/queues/authorizationRules resource.
func NewServiceBusQueueAuthorizationRule() *ServiceBusQueueAuthorizationRule {
	return &ServiceBusQueueAuthorizationRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces/queues/authorizationRules",
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

func init() { azwise.Register(NewServiceBusQueueAuthorizationRule()) }
