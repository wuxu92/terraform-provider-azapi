package azurestackhci

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StackHCIStoragePath provides resource knowledge for Microsoft.AzureStackHCI/storageContainers.
//
// Mirrors azurerm_stack_hci_storage_path. name / resource_group_name / location live on the
// operational envelope; location is surfaced here as ForceNew (commonschema.Location).
//
// Sources:
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_storage_path_resource.go:52-84
//     (schema: ForceNew, name StringMatch regex, path StringIsNotEmpty, custom_location_id)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_storage_path_resource.go:114-125
//     (create mapping: custom_location_id->extendedLocation.name, path->properties.path)
//   - go-azure-sdk resource-manager/azurestackhci/2024-01-01/storagecontainers:
//     model_storagecontainers.go, model_storagecontainerproperties.go (properties.path),
//     model_extendedlocation.go (extendedLocation.name/type)
//
// Not encoded (deliberate):
//   - custom_location_id carries customlocations.ValidateCustomLocationID — a generic
//     resource-ID semantic validator that belongs in an azapin customizer
//     (validators.AzureResourceID) attached to extendedLocation.name, not a StringRule.
type StackHCIStoragePath struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StackHCIStoragePath)(nil)

// NewStackHCIStoragePath returns knowledge for the Azure Stack HCI storageContainers resource.
func NewStackHCIStoragePath() *StackHCIStoragePath {
	return &StackHCIStoragePath{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureStackHCI/storageContainers",
			ApiVersions:  []string{"2024-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "properties.path"},
			},
			RequiredFields: []string{
				"properties.path",
				// AzureRM always sends the custom location as the extended location.
				"extendedLocation.name",
				"extendedLocation.type",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: begin/end alphanumeric, 3-80 chars, alphanumeric + - . _ inside.
					Regex:   `^[a-zA-Z0-9][\-\.\_a-zA-Z0-9]{1,78}[a-zA-Z0-9]$`,
					Message: "name must begin and end with an alphanumeric character, be between 3 and 80 characters in length and can only contain alphanumeric characters, hyphens, periods or underscores",
				},
			},
		},
	}
}

func init() { azwise.Register(NewStackHCIStoragePath()) }
