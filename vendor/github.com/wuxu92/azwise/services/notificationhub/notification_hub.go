package notificationhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NotificationHub provides resource knowledge for
// Microsoft.NotificationHubs/namespaces/notificationHubs.
//
// Contributing Terraform resource: azurerm_notification_hub.
//
// Sources:
//   - terraform-provider-azurerm internal/services/notificationhub/notification_hub_resource.go
//     (schema L36-182, CustomizeDiff L60-79, Create body L207-215,
//     credential expands L339-445)
//   - go-azure-sdk resource-manager/notificationhubs/2023-09-01/hubs:
//     model_notificationhubresource.go, model_notificationhubproperties.go,
//     model_apnscredential(properties).go, model_gcmcredential(properties).go,
//     model_browsercredential(properties).go
//
// Notes:
//   - name/namespace_name/resource_group_name/location are envelope / parent-reference
//     fields; not emitted as body rules.
//   - browser_credential is unconditionally ForceNew → properties.browserCredential.
//   - apns_credential and gcm_credential are ForceNew ONLY when the block is removed
//     (CustomizeDiff L60-79, a nil-value SDK workaround); this conditional replace is
//     not expressible as an unconditional declarative rule, so no ForceNew rule is
//     emitted for them.
//   - apns_credential.application_mode (Production/Sandbox) is transformed into
//     properties.apnsCredential.properties.endpoint (a push URL), not a plain enum
//     field, so it is not emitted as a declarative StringRule.
type NotificationHub struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NotificationHub)(nil)

// NewNotificationHub returns knowledge for the namespaces/notificationHubs resource.
func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NotificationHubs/namespaces/notificationHubs",
			ApiVersions:  []string{"2023-09-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.browserCredential"},
			},
			SensitiveFields: []string{
				"properties.apnsCredential.properties.token",
				"properties.gcmCredential.properties.googleApiKey",
				"properties.browserCredential.properties.vapidPrivateKey",
			},
		},
	}
}

func init() { azwise.Register(NewNotificationHub()) }
