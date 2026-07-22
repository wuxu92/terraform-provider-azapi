package workloads

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WorkloadsSapVirtualInstance provides resource knowledge for
// Microsoft.Workloads/sapVirtualInstances.
//
// This ARM type is exposed by AzureRM as THREE typed Terraform resources, all of
// which create the same sapVirtualInstances resource discriminated by
// properties.configuration.configurationType:
//   - azurerm_workloads_sap_discovery_virtual_instance   → configurationType "Discovery"
//   - azurerm_workloads_sap_single_node_virtual_instance → configurationType "DeploymentWithOSConfig" (SingleServer)
//   - azurerm_workloads_sap_three_tier_virtual_instance  → configurationType "DeploymentWithOSConfig" (ThreeTier)
//
// Only UNIVERSAL knowledge (true for every deployment kind) is unioned here.
// Kind-specific configuration sub-trees (single_server_configuration /
// three_tier_configuration / discovery central_server_virtual_machine_id and their
// nested VM/disk/OS-profile blocks and semantic validators) are NOT unioned: they are
// deeply-nested, kind-specific object/array paths that would corrupt validation for
// the other kinds if applied universally.
//
// Sources:
//   - internal/services/workloads/workloads_sap_discovery_virtual_instance_resource.go
//     (schema 66-125; create 160-185 → SAPVirtualInstance{Identity, Location,
//     Properties{Environment, SapProduct, ManagedResourcesNetworkAccessType,
//     Configuration=DiscoveryConfiguration, ManagedResourceGroupConfiguration.Name}}).
//   - internal/services/workloads/workloads_sap_single_node_virtual_instance_resource.go
//     (schema 117-410; environment/sap_product/managed_resource_group_name/
//     managed_resources_network_access_type at 379-406).
//   - internal/services/workloads/workloads_sap_three_tier_virtual_instance_resource.go
//     (schema 117-1060; environment 196, sap_product 212, managed_resource_group_name
//     1047, managed_resources_network_access_type 1054).
//   - internal/services/workloads/validate/sap_virtual_instance_name.go
//     (regex ^[A-Z][A-Z0-9][A-Z0-9]$).
//   - go-azure-sdk resource-manager/workloads/2024-09-01/sapvirtualinstances:
//     model_sapvirtualinstanceproperties.go (configuration/environment/sapProduct
//     required; managedResourceGroupConfiguration/managedResourcesNetworkAccessType
//     optional; errors/health/provisioningState/state/status read-only),
//     model_sapconfiguration.go (discriminator json tag "configurationType"),
//     constants.go (SAPEnvironmentType NonProd/Prod; SAPProductType ECC/Other/S4HANA;
//     ManagedResourcesNetworkAccessType Private/Public; SAPConfigurationType
//     Deployment/DeploymentWithOSConfig/Discovery),
//     id_sapvirtualinstance.go (segment casing "Microsoft.Workloads"/"sapVirtualInstances").
type WorkloadsSapVirtualInstance struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WorkloadsSapVirtualInstance)(nil)

// NewWorkloadsSapVirtualInstance returns knowledge for the sapVirtualInstances resource.
func NewWorkloadsSapVirtualInstance() *WorkloadsSapVirtualInstance {
	return &WorkloadsSapVirtualInstance{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Workloads/sapVirtualInstances",
			ApiVersions:  []string{"2024-09-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			// Universal ForceNew across all three deployment kinds.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.environment"},
				{PropertyPath: "properties.sapProduct"},
				{PropertyPath: "properties.managedResourceGroupConfiguration.name"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: SAPVirtualInstanceName — 3 chars, first alphabetic uppercase,
					// second/third alphanumeric uppercase.
					Regex:     `^[A-Z][A-Z0-9][A-Z0-9]$`,
					MinLength: 3,
					MaxLength: 3,
					Message:   "must be three characters: first alphabetic, second and third alphanumeric, all uppercase",
				},
				{
					PropertyPath:  "properties.environment",
					AllowedValues: []string{"NonProd", "Prod"},
				},
				{
					PropertyPath:  "properties.sapProduct",
					AllowedValues: []string{"ECC", "Other", "S4HANA"},
				},
				{
					PropertyPath:  "properties.managedResourcesNetworkAccessType",
					AllowedValues: []string{"Private", "Public"},
				},
				{
					// Discriminator common to all deployment kinds; full SDK enum set.
					PropertyPath:  "properties.configuration.configurationType",
					AllowedValues: []string{"Deployment", "DeploymentWithOSConfig", "Discovery"},
				},
			},
			// Universal required body fields (present for every kind).
			RequiredFields: []string{
				"properties.environment",
				"properties.sapProduct",
			},
			DefaultValues: []azwise.DefaultValue{
				// managed_resources_network_access_type Default "Public".
				{PropertyPath: "properties.managedResourcesNetworkAccessType", Value: "Public"},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.state",
				"properties.status",
				"properties.health",
				"properties.errors",
			},
			// NOTE: properties.configuration is required in the SDK but its shape is
			// kind-specific (DiscoveryConfiguration vs DeploymentWithOSConfiguration), so
			// only its universal discriminator (configurationType) is constrained here.
			// NOTE: kind-specific semantic validators (central_server_virtual_machine_id
			// ValidateVirtualMachineID, managed_storage_account_name StorageAccountName,
			// subnet_id / VM image / OS-profile ssh key validators inside the nested
			// single_server_configuration / three_tier_configuration blocks) live on
			// array-element/object paths under properties.configuration and are not
			// expressible as declarative universal rules — attach them per-kind in a
			// resource customizer if needed.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewWorkloadsSapVirtualInstance()) }
