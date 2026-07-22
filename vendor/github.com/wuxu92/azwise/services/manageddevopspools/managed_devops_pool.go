// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package manageddevopspools

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedDevOpsPool provides resource knowledge for Microsoft.DevOpsInfrastructure/pools.
//
// Sources:
//   - terraform-provider-azurerm internal/services/manageddevopspools/managed_devops_pool_resource.go:50-387
//     (azurerm_managed_devops_pool arguments: ForceNew, timeouts, validators; create mapping
//     to pools.Pool / PoolProperties)
//   - terraform-provider-azurerm vendor/.../devopsinfrastructure/2025-09-20/pools/model_poolproperties.go:11-18
//   - terraform-provider-azurerm vendor/.../devopsinfrastructure/2025-09-20/pools/model_vmssfabricprofile.go
//     + model_storageprofile.go:6-9 (fabricProfile.storageProfile.osDiskStorageAccountType)
//   - terraform-provider-azurerm vendor/.../devopsinfrastructure/2025-09-20/pools/model_azuredevopsorganizationprofile.go:13-21
//     (organizationProfile.permissionProfile.kind)
//   - terraform-provider-azurerm vendor/.../devopsinfrastructure/2025-09-20/pools/constants.go
//     (OsDiskStorageAccountType, AzureDevOpsPermissionType)
//   - terraform-provider-azurerm vendor/.../devopsinfrastructure/2025-09-20/pools/id_pool.go:104-119
//     (ARM type casing: Microsoft.DevOpsInfrastructure/pools)
//
// Notes / intentionally skipped:
//   - permission block (properties.organizationProfile.permissionProfile) and its
//     administrator_account groups/users are ForceNew in AzureRM. The block itself is
//     conditional (organizationProfile is a discriminated union with kind "AzureDevOps");
//     permissionProfile.kind is captured below as a value rule, but the nested groups/users
//     ForceNew are array-element paths and not expressible.
//   - virtual_machine_scale_set_fabric.storage.* (caching, storage_account_type) map to
//     properties.fabricProfile.storageProfile.dataDisks[*].* — array-element paths, skipped.
//   - virtual_machine_scale_set_fabric.image[] / azure_devops_organization.organization[] are
//     array bodies; per-element validators (buffer regex, url, parallelism range) are not
//     expressible as single body-path rules.
//   - identity: envelope `identity` block, not a properties.* body field.
type ManagedDevOpsPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedDevOpsPool)(nil)

func NewManagedDevOpsPool() *ManagedDevOpsPool {
	return &ManagedDevOpsPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevOpsInfrastructure/pools",
			ApiVersions:  []string{"2025-09-20"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				// permission.kind is ForceNew when the permission block is set.
				{PropertyPath: "properties.organizationProfile.permissionProfile.kind"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name attribute (resource name).
					Regex:     `^[a-zA-Z0-9][a-zA-Z0-9-.]{1,42}[a-zA-Z0-9-]$`,
					MinLength: 3,
					MaxLength: 44,
					Message:   "name can only include alphanumeric characters, periods (.) and hyphens (-); it must start with an alphanumeric character, cannot end with a period, and be between 3 and 44 characters",
				},
				{
					PropertyPath:  "properties.fabricProfile.storageProfile.osDiskStorageAccountType",
					AllowedValues: []string{"Premium", "Standard", "StandardSSD"},
					Message:       "must be one of Premium, Standard or StandardSSD",
				},
				{
					PropertyPath:  "properties.organizationProfile.permissionProfile.kind",
					AllowedValues: []string{"Inherit", "SpecificAccounts"},
					Message:       "must be one of Inherit or SpecificAccounts",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.maximumConcurrency",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(10000)),
				},
			},
			RequiredFields: []string{
				"properties.maximumConcurrency",
				"properties.devCenterProjectResourceId",
				"properties.organizationProfile",
				"properties.fabricProfile",
				"properties.agentProfile",
			},
			DefaultValues: []azwise.DefaultValue{
				// os_disk_storage_account_type Default:Standard.
				{PropertyPath: "properties.fabricProfile.storageProfile.osDiskStorageAccountType", Value: "Standard"},
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewManagedDevOpsPool()) }
