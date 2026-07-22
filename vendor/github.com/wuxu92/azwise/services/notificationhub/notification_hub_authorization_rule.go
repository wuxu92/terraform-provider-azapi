package notificationhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NotificationHubAuthorizationRule provides resource knowledge for
// Microsoft.NotificationHubs/namespaces/notificationHubs/authorizationRules.
//
// Contributing Terraform resource: azurerm_notification_hub_authorization_rule.
//
// Scoping: verified HUB-scoped via the SDK id parser. The resource builds its ID
// with hubs.NewNotificationHubAuthorizationRuleID (notification_hub_authorization_rule_resource.go
// L118), whose ID() formats
// .../Microsoft.NotificationHubs/namespaces/%s/notificationHubs/%s/authorizationRules/%s
// (go-azure-sdk hubs/id_notificationhubauthorizationrule.go L116). The distinct
// namespace-scoped .../namespaces/authorizationRules type (namespaces/id_authorizationrule.go)
// has NO corresponding azurerm resource, so only this single hub-scoped type is emitted.
//
// Sources:
//   - terraform-provider-azurerm internal/services/notificationhub/notification_hub_authorization_rule_resource.go
//     (schema L22-110, Create body L140-147)
//   - go-azure-sdk resource-manager/notificationhubs/2023-09-01/hubs:
//     id_notificationhubauthorizationrule.go, model_sharedaccessauthorizationruleproperties.go,
//     constants.go (AccessRights: Listen/Send/Manage)
//
// Notes:
//   - name/notification_hub_name/namespace_name/resource_group_name are envelope /
//     parent-reference fields; not emitted as body rules.
//   - manage/send/listen booleans expand into the properties.rights array of AccessRights
//     enums; this many-to-array mapping is not expressible as a single-path rule, so only
//     the required-array constraint is emitted.
//   - primary_access_key / secondary_access_key / *_connection_string are data-plane keys
//     (ListKeys), not body properties; omitted.
type NotificationHubAuthorizationRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NotificationHubAuthorizationRule)(nil)

// NewNotificationHubAuthorizationRule returns knowledge for the
// namespaces/notificationHubs/authorizationRules resource.
func NewNotificationHubAuthorizationRule() *NotificationHubAuthorizationRule {
	return &NotificationHubAuthorizationRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NotificationHubs/namespaces/notificationHubs/authorizationRules",
			ApiVersions:  []string{"2023-09-01"},
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

func init() { azwise.Register(NewNotificationHubAuthorizationRule()) }
