package communication

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CommunicationService provides resource knowledge for
// Microsoft.Communication/communicationServices.
//
// Mirrors azurerm_communication_service.
//
// Sources:
//   - terraform-provider-azurerm internal/services/communication/communication_service_resource.go
//     schema Arguments (lines 61-129: name/data_location ForceNew, data_location StringInSlice enum),
//     Attributes (131-162: hostname computed; primary/secondary connection strings + keys are
//     data-plane ListKeys results with no ARM body path), Create (172-218), Read (265-318),
//     Delete DeleteThenPoll (320-338); timeouts 30m/5m/30m/30m.
//   - internal/services/communication/validate/communication_service_name.go (name regex).
//   - go-azure-sdk resource-manager/communication/2023-03-31/communicationservices
//     CommunicationServiceProperties (dataLocation required; hostName/immutableResourceId/
//     provisioningState/version read-only; linkedDomains is user-settable via the
//     email-domain-association resource, so it is NOT computed).
type CommunicationService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CommunicationService)(nil)

// NewCommunicationService returns knowledge for the communicationServices resource.
func NewCommunicationService() *CommunicationService {
	return &CommunicationService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Communication/communicationServices",
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
					// Resource name (empty PropertyPath). AzureRM CommunicationServiceName:
					// 1-64 chars, start with a letter, letters/numbers/hyphens, no trailing hyphen.
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
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.hostName",
				"properties.immutableResourceId",
				"properties.provisioningState",
				"properties.version",
			},
		},
	}
}

func init() { azwise.Register(NewCommunicationService()) }
