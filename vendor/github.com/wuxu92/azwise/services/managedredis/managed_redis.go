package managedredis

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedRedis provides resource knowledge for Microsoft.Cache/redisEnterprise
// (the Azure Managed Redis cluster).
//
// Mirrors azurerm_managed_redis (the cluster half). The default_database block is
// intentionally excluded here: in ARM it is the child resource
// Microsoft.Cache/redisEnterprise/databases, created via a separate PUT, so its
// property constraints do not belong in the cluster body knowledge.
//
// Sources:
//   - terraform-provider-azurerm internal/services/managedredis/managed_redis_resource.go
//   - Arguments(): name (ForceNew, validate.ManagedRedisClusterName), sku_name (Required,
//     enum), customer_managed_key -> properties.encryption, high_availability_enabled
//     (ForceNew, Default true -> properties.highAvailability), public_network_access
//     (Default Enabled -> properties.publicNetworkAccess)
//   - expandCreateForManagedRedis(): minimumTlsVersion hardcoded to "1.2"
//   - CustomizeDiff(): sku_name is conditionally ForceNew (only when the live SKU
//     transition is not an in-place scale — isSkuAllowedForScaling); NOT emitted as a
//     declarative ForceNew rule because forcing replacement on every SKU change would
//     break valid scale-up/down operations, and azwise cannot evaluate the live check.
//   - Create/Read/Update/Delete timeouts: 45m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/redisenterprise/2025-07-01/redisenterprise
//     model_cluster.go / model_clustercreateproperties.go / constants.go
//
// AMENDED for the legacy azurerm_redis_enterprise_cluster (same ARM type
// Microsoft.Cache/redisEnterprise, api-version 2024-10-01), which resolves here
// via the version-fallback in azwise.Get:
//   - terraform-provider-azurerm internal/services/redisenterprise/redis_enterprise_cluster_resource.go:52-107
//   - Only the genuinely-missing, ARM-universal `zones` ForceNew is unioned (below);
//     top-level zones (zones.Schema) is present and immutable in both the 2024-10-01
//     and 2025-07-01 Cluster models (model_cluster.go), so replacement on change holds
//     for every body.
//   - NOT unioned: sku_name ForceNew is kind-specific — redis_enterprise_cluster forces
//     replacement unconditionally, but azurerm_managed_redis allows in-place SKU scaling
//     (isSkuAllowedForScaling), so forcing it here would corrupt valid scale operations.
//   - NOT unioned: minimum_tls_version ForceNew is version-specific — the legacy resource
//     forces replacement, but the newer managed_redis hardcodes minimumTlsVersion to "1.2"
//     and does not force replacement (already captured as a DefaultValue below).
//   - name validation and the sku.name enum already cover the legacy resource.
type ManagedRedis struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedRedis)(nil)

// NewManagedRedis returns knowledge for the redisEnterprise cluster resource.
func NewManagedRedis() *ManagedRedis {
	return &ManagedRedis{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cache/redisEnterprise",
			ApiVersions:  []string{"2025-07-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; envelope name is
				// Required+RequiresReplace by construction.
				{PropertyPath: "location"},
				// high_availability_enabled maps to properties.highAvailability and is
				// ForceNew in AzureRM.
				{PropertyPath: "properties.highAvailability"},
				// Top-level `zones` is immutable for redisEnterprise across API
				// versions; unioned from the legacy redis_enterprise_cluster (zones
				// ForceNew) as an ARM-universal replacement trigger.
				{PropertyPath: "zones"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 45 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// sku.name is Required. location is envelope-owned; default_database is a
			// separate ARM child resource (databases) and is not part of the cluster body.
			RequiredFields: []string{
				"sku.name",
			},
			DefaultValues: []azwise.DefaultValue{
				// high_availability_enabled Default true -> "Enabled".
				{PropertyPath: "properties.highAvailability", Value: "Enabled"},
				// public_network_access Default Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// AzureRM hardcodes minimumTlsVersion to 1.2 on create.
				{PropertyPath: "properties.minimumTlsVersion", Value: "1.2"},
			},
			StringRules: []azwise.StringRule{
				// Resource name: validate.ManagedRedisClusterName. Length 3..63 and the
				// character regex are declarative; the "no consecutive hyphens" check is
				// semantic and not expressible here (documented, not silently dropped).
				{
					PropertyPath: "",
					MinLength:    3,
					MaxLength:    63,
					Regex:        `^[A-Za-z0-9][A-Za-z0-9-]+[A-Za-z0-9]$`,
					Message:      "name must be 3-63 chars, letters/numbers/hyphens, start and end alphanumeric, no consecutive hyphens",
				},
				// sku_name: full ARM SkuName enum. AzureRM hides Enterprise_/EnterpriseFlash_
				// for azurerm_managed_redis, but the redisEnterprise ARM type accepts them
				// (used by the legacy redis_enterprise_cluster), and AzAPI sends raw values.
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"Balanced_B0", "Balanced_B1", "Balanced_B3", "Balanced_B5",
						"Balanced_B10", "Balanced_B20", "Balanced_B50", "Balanced_B100",
						"Balanced_B150", "Balanced_B250", "Balanced_B350", "Balanced_B500",
						"Balanced_B700", "Balanced_B1000",
						"ComputeOptimized_X3", "ComputeOptimized_X5", "ComputeOptimized_X10",
						"ComputeOptimized_X20", "ComputeOptimized_X50", "ComputeOptimized_X100",
						"ComputeOptimized_X150", "ComputeOptimized_X250", "ComputeOptimized_X350",
						"ComputeOptimized_X500", "ComputeOptimized_X700",
						"Enterprise_E1", "Enterprise_E5", "Enterprise_E10", "Enterprise_E20",
						"Enterprise_E50", "Enterprise_E100", "Enterprise_E200", "Enterprise_E400",
						"EnterpriseFlash_F300", "EnterpriseFlash_F700", "EnterpriseFlash_F1500",
						"FlashOptimized_A250", "FlashOptimized_A500", "FlashOptimized_A700",
						"FlashOptimized_A1000", "FlashOptimized_A1500", "FlashOptimized_A2000",
						"FlashOptimized_A4500",
						"MemoryOptimized_M10", "MemoryOptimized_M20", "MemoryOptimized_M50",
						"MemoryOptimized_M100", "MemoryOptimized_M150", "MemoryOptimized_M250",
						"MemoryOptimized_M350", "MemoryOptimized_M500", "MemoryOptimized_M700",
						"MemoryOptimized_M1000", "MemoryOptimized_M1500", "MemoryOptimized_M2000",
					},
					Message: "sku.name must be a valid redisEnterprise SKU",
				},
				// public_network_access enum.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be Enabled or Disabled",
				},
			},
		},
	}
}

func init() { azwise.Register(NewManagedRedis()) }
