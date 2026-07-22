package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseIntegrationRuntime provides resource knowledge for
// Microsoft.Synapse/workspaces/integrationRuntimes.
//
// This ARM type is discriminated by properties.type (Managed | SelfHosted). AzureRM
// splits it into two typed resources which are MERGED here (only universal knowledge is
// unioned; kind-specific ForceNew/name regexes are omitted — see notes):
//   - azurerm_synapse_integration_runtime_azure   (type = Managed)
//   - azurerm_synapse_integration_runtime_self_hosted (type = SelfHosted)
//
// Sources:
//   - internal/services/synapse/synapse_integration_runtime_azure_resource.go
//     (schema 49-110: name ForceNew (StringMatch); location ForceNew; compute_type enum
//     Default General; core_count IntInSlice(8,16,32,48,80,144,272) Default 8;
//     time_to_live_min Default 0; create 140-156 →
//     ManagedIntegrationRuntime.typeProperties.computeProperties.dataFlowProperties).
//   - internal/services/synapse/synapse_integration_runtime_self_hosted_resource.go
//     (schema 48-81: name ForceNew (StringMatch); description; authorization_key_* Computed;
//     create 110-116 → SelfHostedIntegrationRuntime).
//   - go-azure-sdk resource-manager (track1) synapse models: ManagedIntegrationRuntime,
//     IntegrationRuntimeComputeProperties, IntegrationRuntimeDataFlowProperties
//     (computeType/coreCount/timeToLive json tags), enums.go DataFlowComputeType
//     (General/ComputeOptimized/MemoryOptimized); resourceids.go IntegrationRuntime
//     (segment casing "integrationRuntimes").
//
// Notes:
//   - name is ForceNew in both siblings (universal) but the naming regex differs between
//     Managed (^([a-zA-Z0-9](-|-?[a-zA-Z0-9]+)+[a-zA-Z0-9])$) and SelfHosted
//     (^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$), so no shared name StringRule is emitted (it would
//     reject valid names for the other kind).
//   - compute_type enum lives under the Managed-only sub-object
//     properties.typeProperties.computeProperties.dataFlowProperties.computeType; the enum
//     rule fires only when that sub-object is present, so it is safe to union.
//   - core_count is a discrete allowed set {8,16,32,48,80,144,272}, not a range — not
//     expressible as an IntRule Min/Max, so it is documented here but no rule is emitted.
//   - description maps to properties.description (universal).
type SynapseIntegrationRuntime struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseIntegrationRuntime)(nil)

func NewSynapseIntegrationRuntime() *SynapseIntegrationRuntime {
	return &SynapseIntegrationRuntime{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/integrationRuntimes",
			ApiVersions:  []string{"2021-06-01"},
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
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.typeProperties.computeProperties.dataFlowProperties.computeType",
					AllowedValues: []string{"General", "ComputeOptimized", "MemoryOptimized"},
					Message:       "must be one of General, ComputeOptimized or MemoryOptimized",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseIntegrationRuntime()) }
