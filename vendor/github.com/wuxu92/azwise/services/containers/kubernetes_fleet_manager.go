package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KubernetesFleetManager provides resource knowledge for
// Microsoft.ContainerService/fleets.
//
// Mirrors azurerm_kubernetes_fleet_manager. The fleet is a top-level resource
// whose ARM name and location live on the operational envelope (name is
// Required + ForceNew).
//
// Notes:
//   - hub_profile is deprecated in AzureRM and is NO LONGER sent to the API, so
//     no body rules are emitted for it (dns_prefix / fqdn / kubernetes_version).
//   - FleetProperties only carries hubProfile (deprecated) and the server-computed
//     provisioningState, so this resource has no user-settable body rules beyond
//     the envelope.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/kubernetes_fleet_manager_resource.go
//     (schema L46-80, create L86-124; name Required+ForceNew; hub_profile Deprecated;
//     timeouts 30m/5m/30m/30m)
//   - go-azure-sdk resource-manager/containerservice/2024-04-01/fleets
//     FleetProperties{hubProfile(deprecated), provisioningState(computed)}
type KubernetesFleetManager struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KubernetesFleetManager)(nil)

// NewKubernetesFleetManager returns knowledge for the fleets resource.
func NewKubernetesFleetManager() *KubernetesFleetManager {
	return &KubernetesFleetManager{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerService/fleets",
			ApiVersions:  []string{"2024-04-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewKubernetesFleetManager()) }
