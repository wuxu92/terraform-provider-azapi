package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KubernetesClusterTrustedAccessRoleBinding provides resource knowledge for
// Microsoft.ContainerService/managedClusters/trustedAccessRoleBindings.
//
// Mirrors azurerm_kubernetes_cluster_trusted_access_role_binding. The binding is a
// child of a managed cluster (azurerm's kubernetes_cluster_id) whose ARM name is
// azurerm's name; both live on the operational envelope (name + parent are
// Required + ForceNew).
//
// Notes:
//   - source_resource_id maps to properties.sourceResourceId and is ForceNew.
//   - roles maps to properties.roles (Required list of role strings; no enum
//     constraint in AzureRM).
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/kubernetes_cluster_trusted_access_role_binding_resource.go
//     (schema L44-69, create L75-118; source_resource_id Required+ForceNew;
//     roles Required; timeouts 30m/5m/30m/30m)
//   - go-azure-sdk resource-manager/containerservice/2025-10-01/trustedaccess
//     TrustedAccessRoleBindingProperties{roles(required), sourceResourceId(required);
//     provisioningState computed}
type KubernetesClusterTrustedAccessRoleBinding struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KubernetesClusterTrustedAccessRoleBinding)(nil)

// NewKubernetesClusterTrustedAccessRoleBinding returns knowledge for the trustedAccessRoleBindings resource.
func NewKubernetesClusterTrustedAccessRoleBinding() *KubernetesClusterTrustedAccessRoleBinding {
	return &KubernetesClusterTrustedAccessRoleBinding{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerService/managedClusters/trustedAccessRoleBindings",
			ApiVersions:  []string{"2025-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// source_resource_id (Required + ForceNew).
				{PropertyPath: "properties.sourceResourceId"},
			},
			RequiredFields: []string{
				"properties.roles",
				"properties.sourceResourceId",
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewKubernetesClusterTrustedAccessRoleBinding()) }
