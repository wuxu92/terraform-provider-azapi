package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KubernetesClusterNodePool provides resource knowledge for
// Microsoft.ContainerService/managedClusters/agentPools.
//
// Mirrors azurerm_kubernetes_cluster_node_pool. The agent pool is a child of a
// managed cluster (azurerm's kubernetes_cluster_id) whose ARM name is azurerm's
// name; both live on the operational envelope (name + parent are Required + ForceNew).
// AgentPool.Properties is ManagedClusterAgentPoolProfileProperties, so every body
// path below is under "properties".
//
// Enums use the full ARM SDK constant sets (agentpools + managedclusters packages),
// not the AzureRM-restricted subset, since AzAPI sends raw ARM values.
//
// Notes / skipped (noted, not dropped silently):
//   - name uses containerValidate.KubernetesAgentPoolName and
//     temporary_name_for_rotation is a provider-internal rotation helper with no ARM
//     body field — both are omitted here (semantic / non-mappable).
//   - spot_max_price uses computeValidate.SpotMaxPrice (allows -1 or a positive price);
//     not expressible as a plain FloatRule range, so only its Default (-1) is emitted.
//   - node_public_ip_prefix_id RequiredWith node_public_ip_enabled maps to
//     properties.nodePublicIPPrefixID requiring properties.enableNodePublicIP.
//   - windows_profile.outbound_nat_enabled is INVERTED into
//     properties.windowsProfile.disableOutboundNat (Default true -> false).
//   - node_network_profile.allowed_host_ports.protocol (array element),
//     node_network_profile.node_public_ip_tags (map, ForceNew), and
//     net_ipv4_ip_local_port_range_{min,max} (combined into the single string
//     netIpv4IpLocalPortRange) have no flat scalar ARM path and are skipped.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/kubernetes_cluster_node_pool_resource.go
//     (schema L166-489, create L491-752, expand windows L1826-1836, upgrade_settings L1310-1347;
//     timeouts 60m/5m/60m/60m)
//   - terraform-provider-azurerm internal/services/containers/kubernetes_nodepool.go
//     (kubelet_config L290-368, linux_os_config L370-434, sysctl_config L436-618,
//     node_network_profile L620-676)
//   - go-azure-sdk resource-manager/containerservice/2025-10-01/agentpools
//     ManagedClusterAgentPoolProfileProperties + KubeletConfig / LinuxOSConfig /
//     SysctlConfig / AgentPoolUpgradeSettings / GPUProfile / CreationData /
//     AgentPoolWindowsProfile; enum constants.go
type KubernetesClusterNodePool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KubernetesClusterNodePool)(nil)

// NewKubernetesClusterNodePool returns knowledge for the agentPools resource.
func NewKubernetesClusterNodePool() *KubernetesClusterNodePool {
	return &KubernetesClusterNodePool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerService/managedClusters/agentPools",
			ApiVersions:  []string{"2025-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.hostGroupID"},
				{PropertyPath: "properties.capacityReservationGroupID"},
				{PropertyPath: "properties.scaleSetEvictionPolicy"},
				{PropertyPath: "properties.gpuInstanceProfile"},
				{PropertyPath: "properties.gpuProfile.driver"},
				{PropertyPath: "properties.nodePublicIPPrefixID"},
				{PropertyPath: "properties.osType"},
				{PropertyPath: "properties.scaleSetPriority"},
				{PropertyPath: "properties.proximityPlacementGroupID"},
				{PropertyPath: "properties.spotMaxPrice"},
				// windows_profile block is ForceNew (outbound_nat_enabled -> disableOutboundNat).
				{PropertyPath: "properties.windowsProfile.disableOutboundNat"},
			},
			StringRules: []azwise.StringRule{
				// vm_size (StringIsNotEmpty).
				{PropertyPath: "properties.vmSize", MinLength: 1, Message: "vm_size must not be empty"},
				// orchestrator_version (StringIsNotEmpty).
				{PropertyPath: "properties.orchestratorVersion", MinLength: 1, Message: "orchestrator_version must not be empty"},
				// eviction_policy (ScaleSetEvictionPolicy enum).
				{PropertyPath: "properties.scaleSetEvictionPolicy", AllowedValues: []string{"Deallocate", "Delete"}},
				// gpu_instance (GPUInstanceProfile enum).
				{PropertyPath: "properties.gpuInstanceProfile", AllowedValues: []string{"MIG1g", "MIG2g", "MIG3g", "MIG4g", "MIG7g"}},
				// gpu_driver (GPUDriver enum).
				{PropertyPath: "properties.gpuProfile.driver", AllowedValues: []string{"Install", "None"}},
				// kubelet_disk_type (KubeletDiskType enum).
				{PropertyPath: "properties.kubeletDiskType", AllowedValues: []string{"OS", "Temporary"}},
				// mode (AgentPoolMode enum).
				{PropertyPath: "properties.mode", AllowedValues: []string{"Gateway", "System", "User"}},
				// os_disk_type (OSDiskType enum).
				{PropertyPath: "properties.osDiskType", AllowedValues: []string{"Ephemeral", "Managed"}},
				// os_sku (OSSKU enum, full ARM SDK set).
				{PropertyPath: "properties.osSKU", AllowedValues: []string{"AzureLinux", "AzureLinux3", "CBLMariner", "Ubuntu", "Ubuntu2204", "Ubuntu2404", "Windows2019", "Windows2022"}},
				// os_type (OSType enum).
				{PropertyPath: "properties.osType", AllowedValues: []string{"Linux", "Windows"}},
				// priority (ScaleSetPriority enum).
				{PropertyPath: "properties.scaleSetPriority", AllowedValues: []string{"Regular", "Spot"}},
				// scale_down_mode (ScaleDownMode enum).
				{PropertyPath: "properties.scaleDownMode", AllowedValues: []string{"Deallocate", "Delete"}},
				// workload_runtime (WorkloadRuntime enum).
				{PropertyPath: "properties.workloadRuntime", AllowedValues: []string{"KataVmIsolation", "OCIContainer", "WasmWasi"}},
				// kubelet_config.cpu_manager_policy (StringInSlice).
				{PropertyPath: "properties.kubeletConfig.cpuManagerPolicy", AllowedValues: []string{"none", "static"}},
				// kubelet_config.topology_manager_policy (StringInSlice).
				{PropertyPath: "properties.kubeletConfig.topologyManagerPolicy", AllowedValues: []string{"none", "best-effort", "restricted", "single-numa-node"}},
				// linux_os_config.transparent_huge_page (StringInSlice).
				{PropertyPath: "properties.linuxOSConfig.transparentHugePageEnabled", AllowedValues: []string{"always", "madvise", "never"}},
				// linux_os_config.transparent_huge_page_defrag (StringInSlice).
				{PropertyPath: "properties.linuxOSConfig.transparentHugePageDefrag", AllowedValues: []string{"always", "defer", "defer+madvise", "madvise", "never"}},
				// upgrade_settings.undrainable_node_behavior (UndrainableNodeBehavior enum).
				{PropertyPath: "properties.upgradeSettings.undrainableNodeBehavior", AllowedValues: []string{"Cordon", "Schedule"}},
			},
			IntRules: []azwise.IntRule{
				// node_count / max_count / min_count (IntBetween 0-1000).
				{PropertyPath: "properties.count", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(1000))},
				{PropertyPath: "properties.maxCount", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(1000))},
				{PropertyPath: "properties.minCount", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(1000))},
				// os_disk_size_gb (IntAtLeast 1).
				{PropertyPath: "properties.osDiskSizeGB", MinValue: azwise.Ptr(int64(1))},
				// kubelet_config.image_gc_high_threshold / low_threshold (IntBetween 0-100).
				{PropertyPath: "properties.kubeletConfig.imageGcHighThreshold", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(100))},
				{PropertyPath: "properties.kubeletConfig.imageGcLowThreshold", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(100))},
				// kubelet_config.container_log_max_files (IntAtLeast 2).
				{PropertyPath: "properties.kubeletConfig.containerLogMaxFiles", MinValue: azwise.Ptr(int64(2))},
				// upgrade_settings.drain_timeout_in_minutes (IntAtLeast 0).
				{PropertyPath: "properties.upgradeSettings.drainTimeoutInMinutes", MinValue: azwise.Ptr(int64(0))},
				// upgrade_settings.node_soak_duration_in_minutes (IntBetween 0-30).
				{PropertyPath: "properties.upgradeSettings.nodeSoakDurationInMinutes", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(30))},
				// linux_os_config.sysctl_config.* (IntBetween).
				{PropertyPath: "properties.linuxOSConfig.sysctls.fsAioMaxNr", MinValue: azwise.Ptr(int64(65536)), MaxValue: azwise.Ptr(int64(6553500))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.fsFileMax", MinValue: azwise.Ptr(int64(8192)), MaxValue: azwise.Ptr(int64(12000500))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.fsInotifyMaxUserWatches", MinValue: azwise.Ptr(int64(781250)), MaxValue: azwise.Ptr(int64(2097152))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.fsNrOpen", MinValue: azwise.Ptr(int64(8192)), MaxValue: azwise.Ptr(int64(20000500))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.kernelThreadsMax", MinValue: azwise.Ptr(int64(20)), MaxValue: azwise.Ptr(int64(513785))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netCoreNetdevMaxBacklog", MinValue: azwise.Ptr(int64(1000)), MaxValue: azwise.Ptr(int64(3240000))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netCoreOptmemMax", MinValue: azwise.Ptr(int64(20480)), MaxValue: azwise.Ptr(int64(4194304))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netCoreRmemDefault", MinValue: azwise.Ptr(int64(212992)), MaxValue: azwise.Ptr(int64(134217728))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netCoreRmemMax", MinValue: azwise.Ptr(int64(212992)), MaxValue: azwise.Ptr(int64(134217728))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netCoreSomaxconn", MinValue: azwise.Ptr(int64(4096)), MaxValue: azwise.Ptr(int64(3240000))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netCoreWmemDefault", MinValue: azwise.Ptr(int64(212992)), MaxValue: azwise.Ptr(int64(134217728))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netCoreWmemMax", MinValue: azwise.Ptr(int64(212992)), MaxValue: azwise.Ptr(int64(134217728))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4NeighDefaultGcThresh1", MinValue: azwise.Ptr(int64(128)), MaxValue: azwise.Ptr(int64(80000))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4NeighDefaultGcThresh2", MinValue: azwise.Ptr(int64(512)), MaxValue: azwise.Ptr(int64(90000))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4NeighDefaultGcThresh3", MinValue: azwise.Ptr(int64(1024)), MaxValue: azwise.Ptr(int64(100000))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4TcpFinTimeout", MinValue: azwise.Ptr(int64(5)), MaxValue: azwise.Ptr(int64(120))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4TcpkeepaliveIntvl", MinValue: azwise.Ptr(int64(10)), MaxValue: azwise.Ptr(int64(90))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4TcpKeepaliveProbes", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(15))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4TcpKeepaliveTime", MinValue: azwise.Ptr(int64(30)), MaxValue: azwise.Ptr(int64(432000))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4TcpMaxSynBacklog", MinValue: azwise.Ptr(int64(128)), MaxValue: azwise.Ptr(int64(3240000))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netIpv4TcpMaxTwBuckets", MinValue: azwise.Ptr(int64(8000)), MaxValue: azwise.Ptr(int64(1440000))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netNetfilterNfConntrackBuckets", MinValue: azwise.Ptr(int64(65536)), MaxValue: azwise.Ptr(int64(524288))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.netNetfilterNfConntrackMax", MinValue: azwise.Ptr(int64(131072)), MaxValue: azwise.Ptr(int64(2097152))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.vmMaxMapCount", MinValue: azwise.Ptr(int64(65530)), MaxValue: azwise.Ptr(int64(262144))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.vmSwappiness", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(100))},
				{PropertyPath: "properties.linuxOSConfig.sysctls.vmVfsCachePressure", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(100))},
			},
			DefaultValues: []azwise.DefaultValue{
				// Explicit AzureRM Default: values.
				{PropertyPath: "properties.mode", Value: "User"},
				{PropertyPath: "properties.osDiskType", Value: "Managed"},
				{PropertyPath: "properties.osType", Value: "Linux"},
				{PropertyPath: "properties.scaleSetPriority", Value: "Regular"},
				{PropertyPath: "properties.scaleDownMode", Value: "Delete"},
				{PropertyPath: "properties.spotMaxPrice", Value: float64(-1)},
				{PropertyPath: "properties.enableUltraSSD", Value: false},
				{PropertyPath: "properties.kubeletConfig.cpuCfsQuota", Value: true},
				// windows_profile.outbound_nat_enabled Default true -> disableOutboundNat false (inverted).
				{PropertyPath: "properties.windowsProfile.disableOutboundNat", Value: false},
				// Optional+Computed: server supplies a value when omitted.
				{PropertyPath: "properties.vmSize", Value: nil},
				{PropertyPath: "properties.count", Value: nil},
				{PropertyPath: "properties.kubeletDiskType", Value: nil},
				{PropertyPath: "properties.maxPods", Value: nil},
				{PropertyPath: "properties.nodeLabels", Value: nil},
				{PropertyPath: "properties.orchestratorVersion", Value: nil},
				{PropertyPath: "properties.osDiskSizeGB", Value: nil},
				{PropertyPath: "properties.osSKU", Value: nil},
			},
			ComputedFields: []string{
				// node_image_version is Computed-only in AzureRM (server-assigned).
				"properties.nodeImageVersion",
			},
			// node_public_ip_prefix_id requires node_public_ip_enabled.
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.nodePublicIPPrefixID", "properties.enableNodePublicIP"}},
			},
		},
	}
}

func init() { azwise.Register(NewKubernetesClusterNodePool()) }
