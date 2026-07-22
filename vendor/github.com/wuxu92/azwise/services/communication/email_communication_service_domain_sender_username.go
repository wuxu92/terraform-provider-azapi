package communication

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EmailCommunicationServiceDomainSenderUsername provides resource knowledge for
// Microsoft.Communication/emailServices/domains/senderUsernames.
//
// Mirrors azurerm_email_communication_service_domain_sender_username.
//
// Sources:
//   - terraform-provider-azurerm internal/services/communication/email_communication_service_domain_sender_username_resource.go
//     schema Arguments (lines 40-57: name ForceNew + StringLenBetween(1,253),
//     email_service_domain_id ForceNew parent ref, display_name StringLenBetween(1,253)),
//     Create (75-127: username hardcoded equal to resource name), Update (129-174);
//     CreateOrUpdate is synchronous (no poller); timeouts 30m each.
//   - go-azure-sdk resource-manager/communication/2023-03-31/senderusernames
//     SenderUsernameProperties (username required; displayName optional;
//     dataLocation/provisioningState read-only).
type EmailCommunicationServiceDomainSenderUsername struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EmailCommunicationServiceDomainSenderUsername)(nil)

// NewEmailCommunicationServiceDomainSenderUsername returns knowledge for the
// emailServices/domains/senderUsernames resource.
func NewEmailCommunicationServiceDomainSenderUsername() *EmailCommunicationServiceDomainSenderUsername {
	return &EmailCommunicationServiceDomainSenderUsername{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Communication/emailServices/domains/senderUsernames",
			ApiVersions:  []string{"2023-03-31"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			// AzureRM hardcodes properties.username to the resource name; the ARM model
			// requires it (json:"username", no omitempty).
			RequiredFields: []string{
				"properties.username",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath): StringLenBetween(1, 253).
					MinLength: 1,
					MaxLength: 253,
				},
				{
					// display_name: StringLenBetween(1, 253).
					PropertyPath: "properties.displayName",
					MinLength:    1,
					MaxLength:    253,
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.dataLocation",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewEmailCommunicationServiceDomainSenderUsername()) }
