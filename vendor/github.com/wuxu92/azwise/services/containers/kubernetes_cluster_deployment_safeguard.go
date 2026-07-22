package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KubernetesClusterDeploymentSafeguard provides resource knowledge for
// Microsoft.ContainerService/deploymentSafeguards.
//
// Mirrors azurerm_kubernetes_cluster_deployment_safeguard. This is an extension
// resource scoped to a managed cluster: the ARM path is
// {clusterId}/providers/Microsoft.ContainerService/deploymentSafeguards/default
// (the name is always "default"). azurerm's kubernetes_cluster_id is the parent
// scope and lives on the operational envelope (Required + ForceNew).
//
// Notes:
//   - level (Required) and pod_security_standards_level use the full ARM SDK enum
//     sets; AzAPI sends raw ARM values.
//   - excluded_namespaces maps to properties.excludedNamespaces (list of strings);
//     the per-element StringIsNotEmpty validator targets array elements and is
//     skipped (noted, not dropped silently).
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/kubernetes_cluster_deployment_safeguard_resource.go
//     (schema L43-74, create L80-129, update L169-220; level Required enum;
//     pod_security_standards_level Default Privileged; timeouts 30m)
//   - go-azure-sdk resource-manager/containerservice/2025-07-01/deploymentsafeguards
//     DeploymentSafeguardsProperties{level(required), podSecurityStandardsLevel,
//     excludedNamespaces; provisioningState + systemExcludedNamespaces computed};
//     DeploymentSafeguardsLevel = Enforce|Warn;
//     PodSecurityStandardsLevel = Baseline|Privileged|Restricted
type KubernetesClusterDeploymentSafeguard struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KubernetesClusterDeploymentSafeguard)(nil)

// NewKubernetesClusterDeploymentSafeguard returns knowledge for the deploymentSafeguards resource.
func NewKubernetesClusterDeploymentSafeguard() *KubernetesClusterDeploymentSafeguard {
	return &KubernetesClusterDeploymentSafeguard{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerService/deploymentSafeguards",
			ApiVersions:  []string{"2025-07-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// level (StringInSlice -> DeploymentSafeguardsLevel enum).
				{PropertyPath: "properties.level", AllowedValues: []string{"Enforce", "Warn"}},
				// pod_security_standards_level (StringInSlice -> PodSecurityStandardsLevel enum).
				{PropertyPath: "properties.podSecurityStandardsLevel", AllowedValues: []string{"Baseline", "Privileged", "Restricted"}},
			},
			DefaultValues: []azwise.DefaultValue{
				// pod_security_standards_level Default Privileged.
				{PropertyPath: "properties.podSecurityStandardsLevel", Value: "Privileged"},
			},
			RequiredFields: []string{
				"properties.level",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.systemExcludedNamespaces",
			},
		},
	}
}

func init() { azwise.Register(NewKubernetesClusterDeploymentSafeguard()) }
