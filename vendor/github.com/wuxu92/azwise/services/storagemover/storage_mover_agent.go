package storagemover

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageMoverAgent provides resource knowledge for
// Microsoft.StorageMover/storageMovers/agents.
//
// Mirrors azurerm_storage_mover_agent.
//
// Sources:
//   - internal/services/storagemover/storage_mover_agent_resource.go
//     (Arguments 55-91: name Required ForceNew StringIsNotEmpty; arc_virtual_machine_id
//     Required ForceNew HybridMachineID; arc_virtual_machine_uuid Required ForceNew IsUUID;
//     storage_mover_id ForceNew parent ref; description Optional; create 125-134 →
//     Agent{Properties{ArcResourceId, ArcVMUuid, Description}}; timeouts Create 30m).
//   - go-azure-sdk resource-manager/storagemover/2023-03-01/agents:
//     model_agentproperties.go (arcResourceId/arcVmUuid/description required for create;
//     agentStatus/agentVersion/errorDetails/lastStatusUpdate/localIPAddress/memoryInMB/
//     numberOfCores/provisioningState/uptimeInSeconds read-only),
//     id_agent.go (type segment casing "storageMovers/agents").
type StorageMoverAgent struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageMoverAgent)(nil)

// NewStorageMoverAgent returns knowledge for the storageMovers/agents resource.
func NewStorageMoverAgent() *StorageMoverAgent {
	return &StorageMoverAgent{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageMover/storageMovers/agents",
			ApiVersions:  []string{"2023-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.arcResourceId"},
				{PropertyPath: "properties.arcVmUuid"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: validation.StringIsNotEmpty.
					MinLength: 1,
					Message:   "must not be empty",
				},
				{
					// arc_virtual_machine_uuid: validation.IsUUID (expressed as a UUID regex).
					PropertyPath: "properties.arcVmUuid",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "must be a valid UUID",
				},
			},
			RequiredFields: []string{
				"properties.arcResourceId",
				"properties.arcVmUuid",
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.agentStatus",
				"properties.agentVersion",
				"properties.errorDetails",
				"properties.lastStatusUpdate",
				"properties.localIPAddress",
				"properties.memoryInMB",
				"properties.numberOfCores",
				"properties.uptimeInSeconds",
			},
			// NOTE: arc_virtual_machine_id maps to properties.arcResourceId and is validated
			// by compute/validate.HybridMachineID (a HybridCompute machine resource-ID check).
			// That is a semantic resource-ID validator, not a declarative rule; it belongs in
			// an azapin customizer (typegraph.Validator) rather than a declarative StringRule.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewStorageMoverAgent()) }
