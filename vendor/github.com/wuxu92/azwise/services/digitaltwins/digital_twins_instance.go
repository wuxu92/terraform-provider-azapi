package digitaltwins

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DigitalTwinsInstance provides resource knowledge for
// Microsoft.DigitalTwins/digitalTwinsInstances (azurerm_digital_twins_instance).
//
// Sources:
//   - AzureRM internal/services/digitaltwins/digital_twins_instance_resource.go
//     :32-37  (CRUD timeouts: create/update/delete 30m, read 5m)
//     :44-64  (schema: name ForceNew, location/resource_group_name envelope,
//     host_name Computed, identity, tags)
//     :93-97  (create payload: Location/Identity/Tags only — no settable body properties)
//   - AzureRM internal/services/digitaltwins/validate/digital_twins_instance_name.go
//     :11-33  (name length 3-63 + regex)
//   - go-azure-sdk resource-manager/digitaltwins/2023-01-31/digitaltwinsinstance:
//     model_digitaltwinsdescription.go, model_digitaltwinsproperties.go (ARM body paths),
//     model_digitaltwinspatchproperties.go (Update only exposes publicNetworkAccess),
//     constants.go (PublicNetworkAccess enum)
//
// Not encoded (deliberate):
//   - AzureRM does not expose any settable properties.* fields; publicNetworkAccess
//     exists in the ARM model (settable) but AzureRM has no schema field for it, so
//     it is neither Required nor a Default here (server decides).
//   - identity is a resource envelope block (SystemAssigned/UserAssigned), not a
//     properties.* body field, so no rule is emitted for it.
type DigitalTwinsInstance struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DigitalTwinsInstance)(nil)

// NewDigitalTwinsInstance returns knowledge for the digitalTwinsInstances resource.
func NewDigitalTwinsInstance() *DigitalTwinsInstance {
	return &DigitalTwinsInstance{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DigitalTwins/digitalTwinsInstances",
			ApiVersions:  []string{"2023-01-31"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath = name attribute).
					Regex:     `^[A-Za-z0-9][A-Za-z0-9-]+[A-Za-z0-9]$`,
					MinLength: 3,
					MaxLength: 63,
					Message:   "must be 3-63 chars, begin and end with a letter or number, and contain only letters, numbers, and hyphens",
				},
			},
			// Read-only ARM properties populated by Azure in the GET response.
			ComputedFields: []string{
				"properties.hostName",
				"properties.createdTime",
				"properties.lastUpdatedTime",
				"properties.provisioningState",
				"properties.privateEndpointConnections",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDigitalTwinsInstance()) }
