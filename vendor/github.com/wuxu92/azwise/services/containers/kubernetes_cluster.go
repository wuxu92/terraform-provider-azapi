package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KubernetesCluster provides resource knowledge for
// Microsoft.ContainerService/managedClusters.
//
// This single ARM resource type backs two AzureRM Terraform resources that both
// PUT to managedClusters and therefore MUST be merged into one knowledge entry
// (azwise keys on ARM ResourceType + ApiVersion):
//   - azurerm_kubernetes_cluster            (the full, general-purpose AKS resource)
//   - azurerm_kubernetes_automatic_cluster  (a variant that hardcodes SKU = Automatic
//     and lets the AKS RP auto-manage node pools / networking)
//
// Only knowledge that is UNIVERSAL across both bodies is unioned here. The variant
// (automatic_cluster) has a different valid shape — it does not expose
// default_node_pool / node_provisioning_profile / dns_prefix / sku_tier, and the RP
// manages them automatically — so the regular cluster's kind-specific RequiredFields
// (properties.agentPoolProfiles, properties.nodeProvisioningProfile) and its
// hardcoded SKU are deliberately NOT declared as required. ForceNew and value
// constraints ARE safe to union: a ForceNew rule only fires when its field is
// present and changes, and an enum only fires when its sub-object is present, so
// neither corrupts validation for a body that omits the field.
//
// Envelope fields (name, location, resource_group_name, identity, edge_zone) live on
// the operational envelope and are Required / RequiresReplace by construction, so
// their ForceNew is not repeated as body rules.
//
// Notes / skipped (noted, not dropped silently):
//   - network_profile is a ForceNew block in AzureRM, so the entire
//     properties.networkProfile object is immutable; individual leaf ForceNew paths
//     are subsumed by the parent rule.
//   - dns_prefix / dns_prefix_private_cluster are ExactlyOneOf; they map to
//     properties.dnsPrefix and properties.fqdnSubdomain respectively.
//   - private_dns_zone_id accepts a private DNS zone resource ID OR the literals
//     "System"/"None" (validation.Any) — not a clean enum, so no StringRule is
//     emitted (semantic validator; belongs in a customizer).
//   - kubelet_identity maps to the properties.identityProfile map (keyed by
//     "kubeletidentity"); its ForceNew sub-fields target map values and are skipped.
//   - network_profile.ip_versions maps to properties.networkProfile.ipFamilies[*]
//     (array-element enum) and is skipped.
//   - IsUUID (aad tenant_id / admin_group_object_ids), CIDR, IPv4Address, base64,
//     Duration, and KubernetesDNSPrefix are semantic validators with no declarative
//     StringRule equivalent; they belong in customizer validators.
//   - service_principal ExactlyOneOf identity is an envelope-vs-body relation and is
//     not expressed as a body RelationalRule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/kubernetes_cluster_resource.go
//     (schema L226-1756, create L1900-2090, update L2280-2540, read L2960-3040;
//     timeouts 90m/5m/90m/90m)
//   - terraform-provider-azurerm internal/services/containers/kubernetes_automatic_cluster_resource.go
//     (variant; schema L230-471, create L553-610; hardcodes SKU=Automatic, exposes
//     api_server_access / hosted_system / private_cluster / service_mesh /
//     web_app_routing_ingress)
//   - terraform-provider-azurerm internal/services/containers/kubernetes_nodepool.go
//     (SchemaDefaultNodePool default_node_pool -> properties.agentPoolProfiles)
//   - go-azure-sdk resource-manager/containerservice/2025-10-01/managedclusters
//     ManagedClusterProperties + ContainerServiceNetworkProfile +
//     ManagedClusterAPIServerAccessProfile + ManagedClusterAutoUpgradeProfile +
//     ManagedClusterSKU + load balancer / NAT gateway profiles; enum constants.go
type KubernetesCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KubernetesCluster)(nil)

// NewKubernetesCluster returns knowledge for the managedClusters resource.
func NewKubernetesCluster() *KubernetesCluster {
	return &KubernetesCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerService/managedClusters",
			ApiVersions:  []string{"2025-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.diskEncryptionSetID"},
				{PropertyPath: "properties.dnsPrefix"},
				{PropertyPath: "properties.fqdnSubdomain"},
				{PropertyPath: "properties.apiServerAccessProfile.enablePrivateCluster"},
				{PropertyPath: "properties.apiServerAccessProfile.privateDNSZone"},
				{PropertyPath: "properties.nodeResourceGroup"},
				// network_profile is a ForceNew block: the whole object is immutable.
				{PropertyPath: "properties.networkProfile"},
			},
			StringRules: []azwise.StringRule{
				// automatic_upgrade_channel (UpgradeChannel enum, full ARM SDK set).
				{PropertyPath: "properties.autoUpgradeProfile.upgradeChannel", AllowedValues: []string{"node-image", "none", "patch", "rapid", "stable"}},
				// node_os_upgrade_channel (NodeOSUpgradeChannel enum).
				{PropertyPath: "properties.autoUpgradeProfile.nodeOSUpgradeChannel", AllowedValues: []string{"NodeImage", "None", "SecurityPatch", "Unmanaged"}},
				// sku_tier (ManagedClusterSKUTier enum). Note: SKU lives on the top-level envelope, not properties.
				{PropertyPath: "sku.tier", AllowedValues: []string{"Free", "Standard", "Premium"}},
				// support_plan (KubernetesSupportPlan enum).
				{PropertyPath: "properties.supportPlan", AllowedValues: []string{"KubernetesOfficial", "AKSLongTermSupport"}},
				// auto_scaler_profile.expander (Expander enum).
				{PropertyPath: "properties.autoScalerProfile.expander", AllowedValues: []string{"least-waste", "most-pods", "priority", "random"}},
				// network_profile.network_plugin (NetworkPlugin enum, Required).
				{PropertyPath: "properties.networkProfile.networkPlugin", AllowedValues: []string{"azure", "kubenet", "none"}},
				// network_profile.network_mode (NetworkMode enum).
				{PropertyPath: "properties.networkProfile.networkMode", AllowedValues: []string{"bridge", "transparent"}},
				// network_profile.network_policy (NetworkPolicy enum, full ARM SDK set).
				{PropertyPath: "properties.networkProfile.networkPolicy", AllowedValues: []string{"azure", "calico", "cilium", "none"}},
				// network_profile.network_data_plane (NetworkDataplane enum).
				{PropertyPath: "properties.networkProfile.networkDataplane", AllowedValues: []string{"azure", "cilium"}},
				// network_profile.network_plugin_mode (NetworkPluginMode enum).
				{PropertyPath: "properties.networkProfile.networkPluginMode", AllowedValues: []string{"overlay"}},
				// network_profile.load_balancer_sku (LoadBalancerSku enum).
				{PropertyPath: "properties.networkProfile.loadBalancerSku", AllowedValues: []string{"basic", "standard"}},
				// network_profile.outbound_type (OutboundType enum, full ARM SDK set).
				{PropertyPath: "properties.networkProfile.outboundType", AllowedValues: []string{"loadBalancer", "managedNATGateway", "none", "userAssignedNATGateway", "userDefinedRouting"}},
				// network_profile.load_balancer_profile.backend_pool_type (BackendPoolType enum).
				{PropertyPath: "properties.networkProfile.loadBalancerProfile.backendPoolType", AllowedValues: []string{"NodeIP", "NodeIPConfiguration"}},
			},
			IntRules: []azwise.IntRule{
				// image_cleaner_interval_hours (IntBetween 24-2160).
				{PropertyPath: "properties.securityProfile.imageCleaner.intervalHours", MinValue: azwise.Ptr(int64(24)), MaxValue: azwise.Ptr(int64(2160))},
				// network_profile.load_balancer_profile.outbound_ports_allocated (IntBetween 0-64000).
				{PropertyPath: "properties.networkProfile.loadBalancerProfile.allocatedOutboundPorts", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(64000))},
				// network_profile.load_balancer_profile.idle_timeout_in_minutes (IntBetween 4-100).
				{PropertyPath: "properties.networkProfile.loadBalancerProfile.idleTimeoutInMinutes", MinValue: azwise.Ptr(int64(4)), MaxValue: azwise.Ptr(int64(100))},
				// network_profile.load_balancer_profile.managed_outbound_ip_count (IntBetween 1-100).
				{PropertyPath: "properties.networkProfile.loadBalancerProfile.managedOutboundIPs.count", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(100))},
				// network_profile.load_balancer_profile.managed_outbound_ipv6_count (IntBetween 1-100).
				{PropertyPath: "properties.networkProfile.loadBalancerProfile.managedOutboundIPs.countIPv6", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(100))},
				// network_profile.nat_gateway_profile.idle_timeout_in_minutes (IntBetween 4-120).
				{PropertyPath: "properties.networkProfile.natGatewayProfile.idleTimeoutInMinutes", MinValue: azwise.Ptr(int64(4)), MaxValue: azwise.Ptr(int64(120))},
				// network_profile.nat_gateway_profile.managed_outbound_ip_count (IntBetween 1-100).
				{PropertyPath: "properties.networkProfile.natGatewayProfile.managedOutboundIPProfile.count", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(100))},
			},
			// FloatRules: auto_scaler_profile.max_unready_percentage is TypeFloat in AzureRM
			// but ARM stores every autoScalerProfile field as a kebab-case string
			// (e.g. "max-total-unready-percentage"), so no declarative FloatRule applies.
			ArrayRules: []azwise.ArrayRule{
				// custom_ca_trust_certificates_base64 (MaxItems 10).
				{PropertyPath: "properties.securityProfile.customCATrustCertificates", MaxItems: 10},
			},
			DefaultValues: []azwise.DefaultValue{
				// role_based_access_control_enabled Default true.
				{PropertyPath: "properties.enableRBAC", Value: true},
				// sku_tier Default Free.
				{PropertyPath: "sku.tier", Value: "Free"},
				// support_plan Default KubernetesOfficial.
				{PropertyPath: "properties.supportPlan", Value: "KubernetesOfficial"},
				// node_os_upgrade_channel Default NodeImage.
				{PropertyPath: "properties.autoUpgradeProfile.nodeOSUpgradeChannel", Value: "NodeImage"},
				// network_profile.load_balancer_sku Default standard.
				{PropertyPath: "properties.networkProfile.loadBalancerSku", Value: "standard"},
				// network_profile.outbound_type Default loadBalancer.
				{PropertyPath: "properties.networkProfile.outboundType", Value: "loadBalancer"},
				// network_profile.network_data_plane Default azure.
				{PropertyPath: "properties.networkProfile.networkDataplane", Value: "azure"},
			},
			ComputedFields: []string{
				"properties.fqdn",
				"properties.privateFQDN",
				"properties.azurePortalFQDN",
				"properties.currentKubernetesVersion",
				"properties.oidcIssuerProfile.issuerURL",
				"properties.powerState",
				"properties.provisioningState",
				"properties.maxAgentPools",
			},
		},
	}
}

func init() { azwise.Register(NewKubernetesCluster()) }
