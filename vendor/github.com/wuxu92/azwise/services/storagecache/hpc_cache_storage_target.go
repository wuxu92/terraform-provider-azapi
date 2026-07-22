package storagecache

import (
	"time"

	"github.com/wuxu92/azwise"
)

// HPCCacheStorageTarget provides resource knowledge for
// Microsoft.StorageCache/caches/storageTargets.
//
// MERGES three AzureRM TF resources that all map to the SAME ARM type
// (caches/storageTargets), discriminated by properties.targetType:
//   - azurerm_hpc_cache_blob_target      → targetType "clfs"    (properties.clfs)
//   - azurerm_hpc_cache_blob_nfs_target  → targetType "blobNfs" (properties.blobNfs)
//   - azurerm_hpc_cache_nfs_target       → targetType "nfs3"    (properties.nfs3)
//
// Only knowledge universal to every body is unioned at the top level (name
// validation, timeouts, the targetType discriminator enum, provisioning/state
// computed fields). Value constraints on a sub-object that only one kind sets
// (blobNfs/nfs3 usageModel enum, verification/write-back timers) are safe to
// include because they fire only when that sub-object is present. Kind-specific
// ForceNew (storage_container_id for blob/blob_nfs, target_host_name for nfs) is
// NOT unioned, since it would corrupt validation for the other kinds.
//
// Sources:
//   - internal/services/storagecache/hpc_cache_blob_target_resource.go
//     (schema 46-82: name ForceNew StorageTargetName; namespace_path Required;
//     storage_container_id Required ForceNew; access_policy_name Optional Default
//     "default"; timeouts C/U 30m R 5m D 30m; create 113-127 → targetType Clfs,
//     Clfs.Target, Junctions[*]).
//   - internal/services/storagecache/hpc_cache_blob_nfs_target_resource.go
//     (schema 46-112: usage_model Required StringInSlice(9 models);
//     verification_timer_in_seconds/write_back_timer_in_seconds IntBetween(1,31536000);
//     create 147-171 → targetType BlobNfs, BlobNfs.{Target,UsageModel,VerificationTimer,
//     WriteBackTimer}).
//   - internal/services/storagecache/hpc_cache_nfs_target_resource.go
//     (schema 45-133: namespace_junction Set MaxItems 10; target_host_name Required
//     ForceNew; usage_model Required StringInSlice(9 models); timers IntBetween(1,31536000);
//     create 161-170 → targetType NfsThree, Nfs3.{Target,UsageModel,...}).
//   - internal/services/storagecache/validate/storage_target_name.go
//     (regex ^[-0-9a-zA-Z_]{1,31}$).
//   - go-azure-sdk resource-manager/storagecache/2023-05-01/storagetargets:
//     model_storagetargetproperties.go (targetType/clfs/blobNfs/nfs3/junctions/
//     provisioningState/state), model_blobnfstarget.go + model_nfs3target.go
//     (target/usageModel/verificationTimer/writeBackTimer), model_clfstarget.go (target),
//     model_namespacejunction.go (namespacePath/nfsExport/targetPath/nfsAccessPolicy),
//     constants.go (StorageTargetType blobNfs/clfs/nfs3/unknown),
//     id_storagetarget.go (type segment casing "caches/storageTargets").
type HPCCacheStorageTarget struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*HPCCacheStorageTarget)(nil)

// NewHPCCacheStorageTarget returns knowledge for the caches/storageTargets resource.
func NewHPCCacheStorageTarget() *HPCCacheStorageTarget {
	return &HPCCacheStorageTarget{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageCache/caches/storageTargets",
			ApiVersions:  []string{"2023-05-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Universal ForceNew only. name is ForceNew across all three kinds.
			// (storage_container_id / target_host_name are kind-specific and omitted.)
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: validate.StorageTargetName.
					Regex:     `^[-0-9a-zA-Z_]{1,31}$`,
					MinLength: 1,
					MaxLength: 31,
					Message:   "must be 1-31 characters and contain only letters, numbers, dashes and underscores",
				},
				{
					// Discriminator, present in every body.
					PropertyPath:  "properties.targetType",
					AllowedValues: []string{"blobNfs", "clfs", "nfs3", "unknown"},
				},
				{
					// blob_nfs_target usage_model — fires only when the blobNfs sub-object is set.
					PropertyPath: "properties.blobNfs.usageModel",
					AllowedValues: []string{
						"READ_HEAVY_INFREQ",
						"READ_HEAVY_CHECK_180",
						"READ_ONLY",
						"READ_WRITE",
						"WRITE_WORKLOAD_15",
						"WRITE_AROUND",
						"WRITE_WORKLOAD_CHECK_30",
						"WRITE_WORKLOAD_CHECK_60",
						"WRITE_WORKLOAD_CLOUDWS",
					},
				},
				{
					// nfs_target usage_model — fires only when the nfs3 sub-object is set.
					PropertyPath: "properties.nfs3.usageModel",
					AllowedValues: []string{
						"READ_HEAVY_INFREQ",
						"READ_HEAVY_CHECK_180",
						"READ_ONLY",
						"READ_WRITE",
						"WRITE_WORKLOAD_15",
						"WRITE_AROUND",
						"WRITE_WORKLOAD_CHECK_30",
						"WRITE_WORKLOAD_CHECK_60",
						"WRITE_WORKLOAD_CLOUDWS",
					},
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.blobNfs.verificationTimer",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(31536000)),
				},
				{
					PropertyPath: "properties.blobNfs.writeBackTimer",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(31536000)),
				},
				{
					PropertyPath: "properties.nfs3.verificationTimer",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(31536000)),
				},
				{
					PropertyPath: "properties.nfs3.writeBackTimer",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(31536000)),
				},
			},
			// targetType is required (non-omitempty) in every body.
			RequiredFields: []string{
				"properties.targetType",
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.state",
			},
			// NOTE: access_policy_name (Default "default") and namespace_path/nfs_export/
			// target_path map to properties.junctions[*].{nfsAccessPolicy,namespacePath,
			// nfsExport,targetPath} — array-element paths, not declarative rules.
			// NOTE: kind-specific RequiredFields are intentionally omitted: blob/blob_nfs
			// require properties.clfs.target / properties.blobNfs.target, nfs requires
			// properties.nfs3.target + target_host_name; unioning them would wrongly force
			// every kind to carry the others' sub-objects.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewHPCCacheStorageTarget()) }
