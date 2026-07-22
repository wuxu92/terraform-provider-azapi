package communication

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EmailCommunicationServiceDomain provides resource knowledge for
// Microsoft.Communication/emailServices/domains.
//
// Mirrors azurerm_email_communication_service_domain.
//
// Sources:
//   - terraform-provider-azurerm internal/services/communication/email_communication_service_domain_resource.go
//     schema Arguments (lines 57-86: name/email_service_id/domain_management ForceNew,
//     domain_management StringInSlice(PossibleValuesForDomainManagement),
//     user_engagement_tracking_enabled bool -> properties.userEngagementTracking enum),
//     Attributes (88-114: from_sender_domain/mail_from_sender_domain/verification_records computed),
//     Create (124-181), Update (183-231); timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/communication/2023-03-31/domains
//     DomainProperties (domainManagement required enum; userEngagementTracking enum;
//     dataLocation/fromSenderDomain/mailFromSenderDomain/provisioningState/verificationRecords/
//     verificationStates read-only), constants.go DomainManagement + UserEngagementTracking enums.
type EmailCommunicationServiceDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EmailCommunicationServiceDomain)(nil)

// NewEmailCommunicationServiceDomain returns knowledge for the emailServices/domains resource.
func NewEmailCommunicationServiceDomain() *EmailCommunicationServiceDomain {
	return &EmailCommunicationServiceDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Communication/emailServices/domains",
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
				{PropertyPath: "properties.domainManagement"},
			},
			// domainManagement is required by the ARM model (json:"domainManagement", no omitempty).
			RequiredFields: []string{
				"properties.domainManagement",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.domainManagement",
					AllowedValues: []string{
						"AzureManaged",
						"CustomerManaged",
						"CustomerManagedInExchangeOnline",
					},
				},
				{
					// AzureRM exposes user_engagement_tracking_enabled as a bool, but the ARM
					// body property is an enum. AzAPI users set the raw ARM enum value.
					PropertyPath: "properties.userEngagementTracking",
					AllowedValues: []string{
						"Disabled",
						"Enabled",
					},
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.dataLocation",
				"properties.fromSenderDomain",
				"properties.mailFromSenderDomain",
				"properties.provisioningState",
				"properties.verificationRecords",
				"properties.verificationStates",
			},
			// AzureRM defaults userEngagementTracking to Disabled when unspecified.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.userEngagementTracking", Value: "Disabled"},
			},
		},
	}
}

func init() { azwise.Register(NewEmailCommunicationServiceDomain()) }
