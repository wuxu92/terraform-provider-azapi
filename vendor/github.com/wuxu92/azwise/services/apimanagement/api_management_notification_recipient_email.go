package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementNotificationRecipientEmail provides resource knowledge for
// Microsoft.ApiManagement/service/notifications/recipientEmails.
//
// Mirrors azurerm_api_management_notification_recipient_email. This is a pure
// association resource: CreateOrUpdate takes only the resource ID (no request
// body). The email address is the resource name, and notification_type is a
// path segment (the parent notification name) — both ForceNew in AzureRM.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_notification_recipient_email_resource.go
//     arguments (lines 31-62), Create (lines 76-120), timeouts (30m create / 5m read / 30m delete)
//   - go-azure-sdk apimanagement/2022-08-01/notificationrecipientemail
//     id_recipientemail.go + NotificationName constants
type ApiManagementNotificationRecipientEmail struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementNotificationRecipientEmail)(nil)

func NewApiManagementNotificationRecipientEmail() *ApiManagementNotificationRecipientEmail {
	return &ApiManagementNotificationRecipientEmail{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/notifications/recipientEmails",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // email (resource name, ForceNew)
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (email) ── StringIsNotEmpty
				{MinLength: 1, Message: "email must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementNotificationRecipientEmail()) }
