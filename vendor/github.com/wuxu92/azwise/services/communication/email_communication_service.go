package communication

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EmailCommunicationService provides resource knowledge for
// Microsoft.Communication/emailServices.
//
// Mirrors azurerm_email_communication_service.
//
// Sources:
//   - terraform-provider-azurerm internal/services/communication/email_communication_service_resource.go
//     schema Arguments (lines 43-81: name/data_location ForceNew, data_location StringInSlice enum),
//     Create (95-140), Update (142-185); timeouts 30m/5m/30m/30m.
//   - internal/services/communication/validate/communication_service_name.go (name regex).
//   - go-azure-sdk resource-manager/communication/2023-03-31/emailservices
//     EmailServiceProperties (dataLocation required; provisioningState read-only).
type EmailCommunicationService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EmailCommunicationService)(nil)

// NewEmailCommunicationService returns knowledge for the emailServices resource.
func NewEmailCommunicationService() *EmailCommunicationService {
	return &EmailCommunicationService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Communication/emailServices",
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
				{PropertyPath: "properties.dataLocation"},
			},
			// dataLocation is required by the ARM model (json:"dataLocation", no omitempty).
			RequiredFields: []string{
				"properties.dataLocation",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). Same validator as communication_service.
					Regex:     `^(([a-zA-Z])|([a-zA-Z][0-9a-zA-Z-]{0,62}[0-9a-zA-Z]))$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters, start with a letter, contain only letters, numbers and hyphens, and not end with a hyphen",
				},
				{
					// dataLocation is a free string in the ARM model; AzureRM restricts it
					// to this fixed set of geographies.
					PropertyPath: "properties.dataLocation",
					AllowedValues: []string{
						"Africa", "Asia Pacific", "Australia", "Brazil", "Canada",
						"Europe", "France", "Germany", "India", "Japan", "Korea",
						"Norway", "Switzerland", "UAE", "UK", "United States", "usgov",
					},
				},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewEmailCommunicationService()) }
