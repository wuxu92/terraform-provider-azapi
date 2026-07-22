// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CloudVmCluster provides resource knowledge for Oracle.Database/cloudVmClusters.
//
// Contributing TF resource:
//   - azurerm_oracle_cloud_vm_cluster (oracle_cloud_vm_cluster_resource.go)
//
// Sources:
//   - AzureRM oracle_cloud_vm_cluster_resource.go (schema + Create mapping)
//   - AzureRM validate/cloud_vm_cluster.go (name/cpu/storage/license/system-version rules)
//   - Azure SDK oracledatabase/2025-09-01/cloudvmclusters
//     model_cloudvmclusterproperties.go + constants.go
type CloudVmCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CloudVmCluster)(nil)

// NewCloudVmCluster returns a CloudVmCluster knowledge instance.
func NewCloudVmCluster() *CloudVmCluster {
	return &CloudVmCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Oracle.Database/cloudVmClusters",
			ApiVersions:  []string{"2025-09-01"},
			// Every non-envelope property in this resource is ForceNew; only tags are
			// updatable.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.cloudExadataInfrastructureId"},
				{PropertyPath: "properties.cpuCoreCount"},
				{PropertyPath: "properties.dataStorageSizeInTbs"},
				{PropertyPath: "properties.dbNodeStorageSizeInGbs"},
				{PropertyPath: "properties.dbServers"},
				{PropertyPath: "properties.displayName"},
				{PropertyPath: "properties.giVersion"},
				{PropertyPath: "properties.hostname"},
				{PropertyPath: "properties.licenseModel"},
				{PropertyPath: "properties.memorySizeInGbs"},
				{PropertyPath: "properties.sshPublicKeys"},
				{PropertyPath: "properties.subnetId"},
				{PropertyPath: "properties.vnetId"},
				{PropertyPath: "properties.backupSubnetCidr"},
				{PropertyPath: "properties.clusterName"},
				{PropertyPath: "properties.dataCollectionOptions"},
				{PropertyPath: "properties.dataStoragePercentage"},
				{PropertyPath: "properties.domain"},
				{PropertyPath: "properties.isLocalBackupEnabled"},
				{PropertyPath: "properties.isSparseDiskgroupEnabled"},
				{PropertyPath: "properties.scanListenerPortTcp"},
				{PropertyPath: "properties.scanListenerPortTcpSsl"},
				{PropertyPath: "properties.systemVersion"},
				{PropertyPath: "properties.timeZone"},
				{PropertyPath: "properties.zoneId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 24 * time.Hour,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Non-pointer required fields in CloudVMClusterProperties.
			RequiredFields: []string{
				"properties.cloudExadataInfrastructureId",
				"properties.cpuCoreCount",
				"properties.displayName",
				"properties.giVersion",
				"properties.hostname",
				"properties.sshPublicKeys",
				"properties.subnetId",
				"properties.vnetId",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (validate.CloudVMClusterName) ──
				// 1-255 chars, must start with a letter or underscore. AzureRM also
				// rejects consecutive hyphens ("--"); that check needs negative
				// lookahead which RE2 lacks, so only the start-char rule is encoded.
				{
					Regex:     `^[a-zA-Z_]`,
					MinLength: 1,
					MaxLength: 255,
					Message:   "must be 1-255 characters and start with a letter or underscore",
				},
				{
					PropertyPath: "properties.displayName",
					Regex:        `^[a-zA-Z_]`,
					MinLength:    1,
					MaxLength:    255,
					Message:      "must be 1-255 characters and start with a letter or underscore",
				},
				// ── properties.licenseModel (LicenseModel) ──
				{
					PropertyPath:  "properties.licenseModel",
					AllowedValues: []string{"BringYourOwnLicense", "LicenseIncluded"},
				},
				// ── properties.systemVersion (validate.SystemVersion regex) ──
				{
					PropertyPath: "properties.systemVersion",
					Regex:        `(?:19|22|23|24|25)\.[0-9]+(\.[0-9]+)*|[0-9]+(\.[0-9]+)*`,
					Message:      "must be a valid Exadata system version",
				},
			},
			IntRules: []azwise.IntRule{
				// cpu_core_count — validate.CpuCoreCount: minimum 2 (no upper bound)
				{PropertyPath: "properties.cpuCoreCount", MinValue: azwise.Ptr(int64(2))},
				// data_storage_percentage — validate.DataStoragePercentage: {35,40,60,80};
				// bounded here 35..80 (discrete membership not expressible declaratively).
				{PropertyPath: "properties.dataStoragePercentage", MinValue: azwise.Ptr(int64(35)), MaxValue: azwise.Ptr(int64(80))},
				// scan_listener_port_tcp — IntBetween(1024, 8999)
				{PropertyPath: "properties.scanListenerPortTcp", MinValue: azwise.Ptr(int64(1024)), MaxValue: azwise.Ptr(int64(8999))},
				// scan_listener_port_tcp_ssl — IntBetween(1024, 8999)
				{PropertyPath: "properties.scanListenerPortTcpSsl", MinValue: azwise.Ptr(int64(1024)), MaxValue: azwise.Ptr(int64(8999))},
			},
			FloatRules: []azwise.FloatRule{
				// data_storage_size_in_tbs — validate.DataStorageSizeInTbs: 2..192
				{PropertyPath: "properties.dataStorageSizeInTbs", MinValue: azwise.Ptr(2.0), MaxValue: azwise.Ptr(192.0)},
			},
			DefaultValues: []azwise.DefaultValue{
				// scan_listener_port_tcp Default 1521
				{PropertyPath: "properties.scanListenerPortTcp", Value: float64(1521)},
				// scan_listener_port_tcp_ssl Default 2484
				{PropertyPath: "properties.scanListenerPortTcpSsl", Value: float64(2484)},
			},
			// Server-computed read-only properties (SDK CloudVMClusterProperties).
			ComputedFields: []string{
				"properties.compartmentId",
				"properties.computeModel",
				"properties.computeNodes",
				"properties.diskRedundancy",
				"properties.exascaleDbStorageVaultId",
				"properties.iormConfigCache",
				"properties.lastUpdateHistoryEntryId",
				"properties.lifecycleDetails",
				"properties.lifecycleState",
				"properties.listenerPort",
				"properties.nodeCount",
				"properties.nsgUrl",
				"properties.ociUrl",
				"properties.ocid",
				"properties.ocpuCount",
				"properties.provisioningState",
				"properties.scanDnsName",
				"properties.scanDnsRecordId",
				"properties.scanIpIds",
				"properties.shape",
				"properties.storageManagementType",
				"properties.storageSizeInGbs",
				"properties.subnetOcid",
				"properties.timeCreated",
				"properties.vipIds",
			},
		},
	}
}

func init() { azwise.Register(NewCloudVmCluster()) }
