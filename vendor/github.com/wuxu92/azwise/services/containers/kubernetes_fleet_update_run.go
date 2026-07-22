package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KubernetesFleetUpdateRun provides resource knowledge for
// Microsoft.ContainerService/fleets/updateRuns.
//
// Mirrors azurerm_kubernetes_fleet_update_run. The update run is a child of a
// fleet (azurerm's kubernetes_fleet_manager_id) whose ARM name is azurerm's name;
// both live on the operational envelope (name + parent are Required + ForceNew).
//
// Notes:
//   - managed_cluster_update is Required and maps to properties.managedClusterUpdate;
//     its upgrade sub-object (required) carries the type enum. AzureRM restricts the
//     type to Full/NodeImageOnly, but azwise uses the full ARM SDK ManagedClusterUpgradeType
//     set since AzAPI sends raw ARM values.
//   - fleet_update_strategy_id (properties.updateStrategyId) and the inline stage list
//     (properties.strategy) are mutually exclusive (AzureRM ConflictsWith).
//   - stage/group are array-element paths (properties.strategy.stages[*]); their
//     StringIsNotEmpty validators target array elements and are not expressible as
//     flat rules, so they are skipped (noted, not dropped silently).
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/kubernetes_fleet_update_run_resource.go
//     (schema L72-174, create L180-226, expanders L346-405; timeouts 30m/5m/30m/30m)
//   - go-azure-sdk resource-manager/containerservice/2025-03-01/updateruns
//     UpdateRunProperties{managedClusterUpdate, strategy, updateStrategyId};
//     ManagedClusterUpgradeSpec.Type = ControlPlaneOnly|Full|NodeImageOnly;
//     NodeImageSelection.Type = Consistent|Custom|Latest
type KubernetesFleetUpdateRun struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KubernetesFleetUpdateRun)(nil)

// NewKubernetesFleetUpdateRun returns knowledge for the updateRuns resource.
func NewKubernetesFleetUpdateRun() *KubernetesFleetUpdateRun {
	return &KubernetesFleetUpdateRun{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerService/fleets/updateRuns",
			ApiVersions:  []string{"2025-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// managed_cluster_update.upgrade.type (StringInSlice -> ManagedClusterUpgradeType enum).
				{PropertyPath: "properties.managedClusterUpdate.upgrade.type", AllowedValues: []string{"ControlPlaneOnly", "Full", "NodeImageOnly"}},
				// managed_cluster_update.node_image_selection.type (StringInSlice -> NodeImageSelectionType enum).
				{PropertyPath: "properties.managedClusterUpdate.nodeImageSelection.type", AllowedValues: []string{"Consistent", "Custom", "Latest"}},
			},
			RequiredFields: []string{
				"properties.managedClusterUpdate",
				"properties.managedClusterUpdate.upgrade",
				"properties.managedClusterUpdate.upgrade.type",
			},
			// fleet_update_strategy_id <-> inline stage list are mutually exclusive.
			ConflictsWith: []azwise.RelationalRule{
				{Paths: []string{"properties.updateStrategyId", "properties.strategy"}},
			},
		},
	}
}

func init() { azwise.Register(NewKubernetesFleetUpdateRun()) }
