package batch

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BatchPool provides resource knowledge for Microsoft.Batch/batchAccounts/pools.
//
// Contributing Terraform resource: azurerm_batch_pool.
//
// Large resource: processed block-by-block. ARM paths below were verified against the
// go-azure-sdk pool models rather than derived purely mechanically.
//
// Sources:
//   - terraform-provider-azurerm internal/services/batch/batch_pool_resource.go
//     (schema L51-820, Create body L898-988)
//   - terraform-provider-azurerm internal/services/batch/validate/pool_name.go
//   - go-azure-sdk resource-manager/batch/2024-07-01/pool:
//     model_poolproperties.go, model_deploymentconfiguration.go,
//     model_virtualmachineconfiguration.go, model_scalesettings.go,
//     model_fixedscalesettings.go, model_autoscalesettings.go,
//     model_networkconfiguration.go, model_containerconfiguration.go,
//     model_datadisk.go, model_diskencryptionconfiguration.go, model_useraccount.go,
//     model_securityprofile.go, constants.go.
//
// Key mappings (non-obvious renames):
//   - max_tasks_per_node          → properties.taskSlotsPerNode
//   - storage_image_reference     → properties.deploymentConfiguration.virtualMachineConfiguration.imageReference
//   - node_agent_sku_id           → properties.deploymentConfiguration.virtualMachineConfiguration.nodeAgentSkuId
//   - fixed_scale / auto_scale    → properties.scaleSettings.{fixedScale,autoScale}
//   - container_configuration     → properties.deploymentConfiguration.virtualMachineConfiguration.containerConfiguration
//   - data_disks / disk_encryption/ extensions / license_type / node_placement /
//     os_disk_placement / security_profile → under
//     properties.deploymentConfiguration.virtualMachineConfiguration.*
//
// Notes:
//   - name/account_name/resource_group_name are envelope-owned; not body fields.
//   - stop_pending_resize_operation and certificate (pre-5.0) are provider-side operational
//     concerns with no stable create-body property mapping and are not encoded.
//   - Several enum constraints live inside arrays (data_disks[*], user_accounts[*],
//     disk_encryption targets[*]); StringRule "[*]" paths are used where the array element
//     type is unambiguous. Endpoint/NSG-rule and extension enums nested several levels inside
//     network_configuration arrays are covered by AzureRM schema validation and are omitted
//     here to avoid emitting fragile deep array paths.
type BatchPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BatchPool)(nil)

// NewBatchPool returns knowledge for the batchAccounts/pools resource.
func NewBatchPool() *BatchPool {
	return &BatchPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Batch/batchAccounts/pools",
			ApiVersions:  []string{"2024-07-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Unconditionally-ForceNew create-body properties.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.vmSize"},
				{PropertyPath: "properties.displayName"},
				{PropertyPath: "properties.taskSlotsPerNode"}, // max_tasks_per_node
				{PropertyPath: "properties.deploymentConfiguration.virtualMachineConfiguration.nodeAgentSkuId"},
				{PropertyPath: "properties.deploymentConfiguration.virtualMachineConfiguration.imageReference"}, // storage_image_reference block
				{PropertyPath: "properties.deploymentConfiguration.virtualMachineConfiguration.securityProfile"}, // security_profile block
				{PropertyPath: "properties.deploymentConfiguration.virtualMachineConfiguration.containerConfiguration.containerImageNames"},
				{PropertyPath: "properties.deploymentConfiguration.virtualMachineConfiguration.containerConfiguration.containerRegistries"},
				{PropertyPath: "properties.networkConfiguration"}, // network_configuration block
			},
			// Required create-body properties (vm_size, storage_image_reference, node_agent_sku_id).
			RequiredFields: []string{
				"properties.vmSize",
				"properties.deploymentConfiguration.virtualMachineConfiguration.imageReference",
				"properties.deploymentConfiguration.virtualMachineConfiguration.nodeAgentSkuId",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.PoolName
				{
					Regex:     `^[a-zA-Z0-9_-]+$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "may contain any combination of alphanumeric characters, hyphens, and underscores (1-64 chars)",
				},
				// ── inter_node_communication → properties.interNodeCommunication
				{
					PropertyPath:  "properties.interNodeCommunication",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				// ── target_node_communication_mode → properties.targetNodeCommunicationMode (full ARM SDK set)
				{
					PropertyPath:  "properties.targetNodeCommunicationMode",
					AllowedValues: []string{"Classic", "Default", "Simplified"},
					Message:       "must be one of Classic, Default, or Simplified",
				},
				// ── task_scheduling_policy.node_fill_type → properties.taskSchedulingPolicy.nodeFillType
				{
					PropertyPath:  "properties.taskSchedulingPolicy.nodeFillType",
					AllowedValues: []string{"Spread", "Pack"},
					Message:       "must be one of Spread or Pack",
				},
				// ── fixed_scale.node_deallocation_method → properties.scaleSettings.fixedScale.nodeDeallocationOption
				{
					PropertyPath:  "properties.scaleSettings.fixedScale.nodeDeallocationOption",
					AllowedValues: []string{"Requeue", "RetainedData", "TaskCompletion", "Terminate"},
					Message:       "must be one of Requeue, RetainedData, TaskCompletion, or Terminate",
				},
				// ── network_configuration.dynamic_vnet_assignment_scope
				{
					PropertyPath:  "properties.networkConfiguration.dynamicVnetAssignmentScope",
					AllowedValues: []string{"none", "job"},
					Message:       "must be one of none or job",
				},
				// ── network_configuration.public_address_provisioning_type
				//    → properties.networkConfiguration.publicIPAddressConfiguration.provision
				{
					PropertyPath:  "properties.networkConfiguration.publicIPAddressConfiguration.provision",
					AllowedValues: []string{"BatchManaged", "UserManaged", "NoPublicIPAddresses"},
					Message:       "must be one of BatchManaged, UserManaged, or NoPublicIPAddresses",
				},
				// ── security_profile.security_type → ...securityProfile.securityType
				{
					PropertyPath:  "properties.deploymentConfiguration.virtualMachineConfiguration.securityProfile.securityType",
					AllowedValues: []string{"trustedLaunch", "confidentialVM"},
					Message:       "must be one of trustedLaunch or confidentialVM",
				},
				// ── node_placement.policy → ...nodePlacementConfiguration.policy
				{
					PropertyPath:  "properties.deploymentConfiguration.virtualMachineConfiguration.nodePlacementConfiguration.policy",
					AllowedValues: []string{"Regional", "Zonal"},
					Message:       "must be one of Regional or Zonal",
				},
				// ── data_disks[*].caching → ...dataDisks[*].caching
				{
					PropertyPath:  "properties.deploymentConfiguration.virtualMachineConfiguration.dataDisks[*].caching",
					AllowedValues: []string{"None", "ReadOnly", "ReadWrite"},
					Message:       "must be one of None, ReadOnly, or ReadWrite",
				},
				// ── data_disks[*].storage_account_type → ...dataDisks[*].storageAccountType (full ARM SDK set)
				{
					PropertyPath:  "properties.deploymentConfiguration.virtualMachineConfiguration.dataDisks[*].storageAccountType",
					AllowedValues: []string{"Standard_LRS", "Premium_LRS", "StandardSSD_LRS"},
					Message:       "must be one of Standard_LRS, Premium_LRS, or StandardSSD_LRS",
				},
				// ── disk_encryption[*].disk_encryption_target → ...diskEncryptionConfiguration.targets[*]
				{
					PropertyPath:  "properties.deploymentConfiguration.virtualMachineConfiguration.diskEncryptionConfiguration.targets[*]",
					AllowedValues: []string{"OsDisk", "TemporaryDisk"},
					Message:       "must be one of OsDisk or TemporaryDisk",
				},
				// ── user_accounts[*].elevation_level → properties.userAccounts[*].elevationLevel
				{
					PropertyPath:  "properties.userAccounts[*].elevationLevel",
					AllowedValues: []string{"NonAdmin", "Admin"},
					Message:       "must be one of NonAdmin or Admin",
				},
			},
			IntRules: []azwise.IntRule{
				// ── max_tasks_per_node → properties.taskSlotsPerNode ── IntAtLeast(1)
				{PropertyPath: "properties.taskSlotsPerNode", MinValue: azwise.Ptr(int64(1)), Message: "must be at least 1"},
				// ── fixed_scale.target_dedicated_nodes ── IntBetween(0, 2000)
				{PropertyPath: "properties.scaleSettings.fixedScale.targetDedicatedNodes", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(2000)), Message: "must be between 0 and 2000"},
				// ── fixed_scale.target_low_priority_nodes ── IntBetween(0, 1000)
				{PropertyPath: "properties.scaleSettings.fixedScale.targetLowPriorityNodes", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(1000)), Message: "must be between 0 and 1000"},
			},
			DefaultValues: []azwise.DefaultValue{
				// max_tasks_per_node Default 1 → taskSlotsPerNode.
				{PropertyPath: "properties.taskSlotsPerNode", Value: float64(1)},
				// inter_node_communication Default Enabled.
				{PropertyPath: "properties.interNodeCommunication", Value: "Enabled"},
				// fixed_scale defaults (present only when the fixed_scale block is set).
				{PropertyPath: "properties.scaleSettings.fixedScale.resizeTimeout", Value: "PT15M"},
				{PropertyPath: "properties.scaleSettings.fixedScale.targetDedicatedNodes", Value: float64(1)},
				{PropertyPath: "properties.scaleSettings.fixedScale.targetLowPriorityNodes", Value: float64(0)},
				// auto_scale default (present only when the auto_scale block is set).
				{PropertyPath: "properties.scaleSettings.autoScale.evaluationInterval", Value: "PT15M"},
				// network_configuration.dynamic_vnet_assignment_scope Default none.
				{PropertyPath: "properties.networkConfiguration.dynamicVnetAssignmentScope", Value: "none"},
			},
			// Read-only runtime/state properties returned by GET; sending them is meaningless
			// and produces perpetual diffs, so they are stripped.
			ComputedFields: []string{
				"properties.allocationState",
				"properties.allocationStateTransitionTime",
				"properties.autoScaleRun",
				"properties.creationTime",
				"properties.currentDedicatedNodes",
				"properties.currentLowPriorityNodes",
				"properties.currentNodeCommunicationMode",
				"properties.lastModified",
				"properties.provisioningState",
				"properties.provisioningStateTransitionTime",
				"properties.resizeOperationStatus",
			},
		},
	}
}

func init() { azwise.Register(NewBatchPool()) }
