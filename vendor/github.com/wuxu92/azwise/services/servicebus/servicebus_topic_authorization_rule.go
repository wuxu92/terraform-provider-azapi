// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusTopicAuthorizationRule provides resource knowledge for
// Microsoft.ServiceBus/namespaces/topics/authorizationRules.
//
// Contributing Terraform resource: azurerm_servicebus_topic_authorization_rule.
//
// Scoping verified via SDK id parser: TopicAuthorizationRuleId.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s/topics/%s/authorizationRules/%s
// (topicsauthorizationrule/id_topicauthorizationrule.go L116-117).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_topic_authorization_rule_resource.go
//     (schema L48-65, Create body via expandAuthorizationRuleRights)
//   - terraform-provider-azurerm internal/services/servicebus/internal.go
//     (listen/send/manage schema L60-107, expandAuthorizationRuleRights L21-37, CustomizeDiff L109-123)
//   - terraform-provider-azurerm internal/services/servicebus/validate/authorization_rule_name.go (L13-18)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/topicsauthorizationrule:
//     model_sbauthorizationruleproperties.go (Rights []AccessRights), constants.go (AccessRights: Listen/Manage/Send)
//
// Notes:
//   - name/topic_id are envelope / parent-reference fields; not emitted as body rules.
//   - listen/send/manage booleans expand into the properties.rights array of AccessRights
//     enums; a many-to-array mapping not expressible as a single-path rule, so only the
//     required-array constraint is emitted. CustomizeDiff requires at least one right, and
//     manage implies both listen and send.
//   - primary_key / secondary_key / *_connection_string are data-plane keys (ListKeys), not
//     body properties; omitted.
type ServiceBusTopicAuthorizationRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusTopicAuthorizationRule)(nil)

// NewServiceBusTopicAuthorizationRule returns knowledge for the
// namespaces/topics/authorizationRules resource.
func NewServiceBusTopicAuthorizationRule() *ServiceBusTopicAuthorizationRule {
	return &ServiceBusTopicAuthorizationRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces/topics/authorizationRules",
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

func init() { azwise.Register(NewServiceBusTopicAuthorizationRule()) }
