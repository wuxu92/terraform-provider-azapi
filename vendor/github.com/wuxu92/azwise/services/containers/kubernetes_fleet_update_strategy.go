package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KubernetesFleetUpdateStrategy provides resource knowledge for
// Microsoft.ContainerService/fleets/updateStrategies.
//
// Mirrors azurerm_kubernetes_fleet_update_strategy. The update strategy is a child
// of a fleet (azurerm's kubernetes_fleet_manager_id) whose ARM name is azurerm's
// name; both live on the operational envelope (name + parent are Required + ForceNew).
//
// Notes:
//   - stage is Required and maps to properties.strategy.stages; the nested
//     stage.name / stage.group.name (StringIsNotEmpty) validators and
//     after_stage_wait_in_seconds target array-element paths
//     (properties.strategy.stages[*]) and are not expressible as flat rules, so they
//     are skipped (noted, not dropped silently).
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/kubernetes_fleet_update_strategy_resource.go
//     (schema L56-100, create L106-152, expanders L245-265; timeouts 30m/5m/30m/30m)
//   - go-azure-sdk resource-manager/containerservice/2025-03-01/fleetupdatestrategies
//     FleetUpdateStrategyProperties{strategy(required)}; UpdateRunStrategy.Stages
type KubernetesFleetUpdateStrategy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KubernetesFleetUpdateStrategy)(nil)

// NewKubernetesFleetUpdateStrategy returns knowledge for the updateStrategies resource.
func NewKubernetesFleetUpdateStrategy() *KubernetesFleetUpdateStrategy {
	return &KubernetesFleetUpdateStrategy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerService/fleets/updateStrategies",
			ApiVersions:  []string{"2025-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.strategy",
			},
		},
	}
}

func init() { azwise.Register(NewKubernetesFleetUpdateStrategy()) }
