package naming

// ambiguousResourceNames assigns a canonical azapi noun to ARM types the
// azurerm-armtypes extractor leaves ambiguous: several unrelated AzureRM
// resources register the same ARM type (e.g. Microsoft.Compute/virtualMachines
// backs azurerm_linux_virtual_machine, azurerm_windows_virtual_machine and
// azurerm_virtual_machine), so azurerm_reference_gen.go omits the type and
// ResourceName would otherwise fall back to the mechanical service+segment
// derivation. A single native azapi_* resource for such a type is inherently
// os-/variant-agnostic, so each entry pins the name that resource should carry:
// the parent/primary noun with the discriminator dropped.
//
// These are hand-curated (see azurerm_reference_report.md "Ambiguous ARM types")
// and applied by init to keep the generated table (azurerm_reference_gen.go,
// DO NOT EDIT) machine-owned. Keys use go-azure-sdk casing; the lookup folds
// case. Values are the azurerm_ noun; the azapi_ prefix is substituted at lookup,
// exactly like the generated entries.
var ambiguousResourceNames = map[string]string{
	// Compute: os-agnostic VM / scale-set / restore-point resources.
	"Microsoft.Compute/restorePointCollections": "azurerm_virtual_machine_restore_point_collection",
	"Microsoft.Compute/virtualMachineScaleSets": "azurerm_virtual_machine_scale_set",
	"Microsoft.Compute/virtualMachines":         "azurerm_virtual_machine",

	// Data Factory data flow: the mapping data flow; flowlet is a variant.
	"Microsoft.DataFactory/factories/dataflows": "azurerm_data_factory_data_flow",

	// Application Insights web test: the classic web test; standard is a variant.
	"Microsoft.Insights/webTests": "azurerm_application_insights_web_test",

	// ML workspace: the workspace resource. AI Foundry (hub/project) is the same
	// ARM type with a different kind and keeps its mechanical name.
	"Microsoft.MachineLearningServices/workspaces": "azurerm_machine_learning_workspace",

	// Front Door WAF policy: the current CDN Front Door policy; the classic
	// (Front Door Service) policy is legacy.
	"Microsoft.Network/frontDoorWebApplicationFirewallPolicies": "azurerm_cdn_frontdoor_firewall_policy",

	// Virtual hub: the hub resource. Route Server is a virtual hub sku.
	"Microsoft.Network/virtualHubs":                "azurerm_virtual_hub",
	"Microsoft.Network/virtualHubs/bgpConnections": "azurerm_virtual_hub_bgp_connection",

	// Policy remediation: one resource applied at any scope (resource group,
	// subscription, management group); the scope-agnostic noun.
	"Microsoft.PolicyInsights/remediations": "azurerm_policy_remediation",

	// Site Recovery: the replication fabric, network mapping, protection-container
	// mapping and replication policy resources, with the hyperv/vmware
	// discriminators dropped.
	"Microsoft.RecoveryServices/vaults/replicationFabrics":                                                                        "azurerm_site_recovery_fabric",
	"Microsoft.RecoveryServices/vaults/replicationFabrics/replicationNetworks/replicationNetworkMappings":                         "azurerm_site_recovery_network_mapping",
	"Microsoft.RecoveryServices/vaults/replicationFabrics/replicationProtectionContainers/replicationProtectionContainerMappings": "azurerm_site_recovery_protection_container_mapping",
	"Microsoft.RecoveryServices/vaults/replicationPolicies":                                                                       "azurerm_site_recovery_replication_policy",

	// Storage Mover endpoint: one endpoint resource; source/target is a variant.
	"Microsoft.StorageMover/storageMovers/endpoints": "azurerm_storage_mover_endpoint",

	// App Service site (Microsoft.Web/sites) and its slot and hybrid-connection
	// children. The linux/windows/function/logic app variants collapse to web_site
	// — the mechanical noun the already-generated native azapi_web_site resource
	// carries — and the children follow that stem so the family stays consistent.
	// (Related standalone resources keep their own azurerm nouns: certificates ->
	// app_service_certificate below.)
	"Microsoft.Web/sites":                                   "azurerm_web_site",
	"Microsoft.Web/sites/slots":                             "azurerm_web_site_slot",
	"Microsoft.Web/sites/hybridConnectionNamespaces/relays": "azurerm_web_site_hybrid_connection",
}

// init patches the generated azurermResourceForARMType table with the curated
// names above, then rebuilds azurermReferenceLower. The rebuild is required:
// azurermReferenceLower is a package-level var initialized before any init runs,
// so it would otherwise index the un-patched table.
func init() {
	for armType, name := range ambiguousResourceNames {
		azurermResourceForARMType[armType] = name
	}
	azurermReferenceLower = buildAzurermReferenceLower()
}
