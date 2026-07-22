package arckubernetes

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ArcKubernetesCluster provides resource knowledge for
// Microsoft.Kubernetes/connectedClusters.
//
// This single ARM resource type backs two AzureRM Terraform resources that both
// PUT to connectedClusters and therefore MUST be merged into one knowledge entry
// (azwise keys on ARM ResourceType + ApiVersion):
//   - azurerm_arc_kubernetes_cluster            (kind defaults to unset)
//   - azurerm_arc_kubernetes_provisioned_cluster (kind = "ProvisionedCluster")
//
// The union of both schemas is captured here. Envelope fields (name,
// resource_group_name, location, identity) live on the operational envelope and
// are Required + RequiresReplace by construction, so their ForceNew is not
// repeated as body rules.
//
// Notes:
//   - properties.agentPublicKeyCertificate is Required + ForceNew for the regular
//     cluster but is NOT set by the provisioned-cluster variant (kind=ProvisionedCluster).
//     Because both variants share this ARM type, it is declared ForceNew but is
//     deliberately NOT added to RequiredFields — marking it required would break
//     provisioned-cluster creation.
//   - Base64EncodedString (agent_public_key_certificate) and IsUUID
//     (aadProfile.tenantID, aadProfile.adminGroupObjectIDs) are semantic validators
//     with no declarative StringRule equivalent; they are intentionally omitted here
//     (IsUUID over an array element is also unsupported by the overlay).
//
// Sources:
//   - terraform-provider-azurerm internal/services/arckubernetes/arc_kubernetes_cluster_resource.go
//     (schema L45-105, create L136-143; name regex ^[-_a-zA-Z0-9]{1,260}$;
//     agent_public_key_certificate Required+ForceNew Base64; timeouts 30m/5m/30m/30m)
//   - terraform-provider-azurerm internal/services/arckubernetes/arc_kubernetes_provisioned_cluster_resource.go
//     (schema L75-176, create L213-231, expand L387-406; azure_active_directory ->
//     properties.aadProfile; arc_agent_desired_version/arc_agent_auto_upgrade_enabled ->
//     properties.arcAgentProfile)
//   - go-azure-sdk resource-manager/hybridkubernetes/2024-01-01/connectedclusters
//     ConnectedClusterProperties (aadProfile/arcAgentProfile settable;
//     agentVersion/distribution/infrastructure/kubernetesVersion/offering/
//     totalCoreCount/totalNodeCount server-computed); AutoUpgradeOptions = Disabled|Enabled
type ArcKubernetesCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ArcKubernetesCluster)(nil)

// NewArcKubernetesCluster returns knowledge for the connectedClusters resource.
func NewArcKubernetesCluster() *ArcKubernetesCluster {
	return &ArcKubernetesCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Kubernetes/connectedClusters",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// agent_public_key_certificate (Required+ForceNew on the regular cluster).
				{PropertyPath: "properties.agentPublicKeyCertificate"},
			},
			StringRules: []azwise.StringRule{
				// name (StringMatch), shared by both TF resources.
				{Regex: "^[-_a-zA-Z0-9]{1,260}$", Message: "name may contain only alphanumeric characters, underscores and hyphens, with a maximum length of 260 characters"},
				// arc_agent_desired_version (StringIsNotEmpty).
				{PropertyPath: "properties.arcAgentProfile.desiredAgentVersion", MinLength: 1, Message: "desired agent version must not be empty"},
				// arc_agent_auto_upgrade_enabled maps to an AutoUpgradeOptions enum.
				{PropertyPath: "properties.arcAgentProfile.agentAutoUpgrade", AllowedValues: []string{"Disabled", "Enabled"}},
			},
			DefaultValues: []azwise.DefaultValue{
				// azure_active_directory.azure_rbac_enabled default false.
				{PropertyPath: "properties.aadProfile.enableAzureRBAC", Value: false},
				// arc_agent_auto_upgrade_enabled default true -> AutoUpgradeOptionsEnabled.
				{PropertyPath: "properties.arcAgentProfile.agentAutoUpgrade", Value: "Enabled"},
			},
			// Server-computed read-only properties surfaced by AzureRM as Attributes.
			ComputedFields: []string{
				"properties.agentVersion",
				"properties.distribution",
				"properties.infrastructure",
				"properties.kubernetesVersion",
				"properties.offering",
				"properties.totalCoreCount",
				"properties.totalNodeCount",
			},
		},
	}
}

func init() { azwise.Register(NewArcKubernetesCluster()) }
