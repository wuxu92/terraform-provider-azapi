package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementNotificationRecipientUser provides resource knowledge for
// Microsoft.ApiManagement/service/notifications/recipientUsers.
//
// Mirrors azurerm_api_management_notification_recipient_user. This is a pure
// association resource: CreateOrUpdate takes only the resource ID (no request
// body). The user_id is the resource name, and notification_type is a path
// segment (the parent notification name) — both ForceNew in AzureRM.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_notification_recipient_user_resource.go
//     arguments (lines 31-62), Create (lines 76-120), timeouts (30m create / 5m read / 30m delete)
//   - go-azure-sdk apimanagement/2022-08-01/notificationrecipientuser
//     id_recipientuser.go + NotificationName constants
type ApiManagementNotificationRecipientUser struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementNotificationRecipientUser)(nil)

func NewApiManagementNotificationRecipientUser() *ApiManagementNotificationRecipientUser {
	return &ApiManagementNotificationRecipientUser{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/notifications/recipientUsers",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // user_id (resource name, ForceNew)
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (user_id) ── StringIsNotEmpty
				{MinLength: 1, Message: "user_id must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementNotificationRecipientUser()) }
