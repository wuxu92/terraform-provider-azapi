package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppPool provides resource knowledge for Microsoft.NetApp/netAppAccounts/capacityPools.
//
// Mirrors azurerm_netapp_pool.
//
// Sources:
//   - internal/services/netapp/netapp_pool_resource.go
//     (schema 46-117: name ForceNew+PoolName; service_level Required ForceNew enum;
//     size_in_tb IntBetween(1,2048) → properties.size in bytes; qos_type default Auto;
//     encryption_type Optional ForceNew default Single; cool_access_enabled default false;
//     custom_throughput_mibps IntAtLeast(128); timeouts C/U/D 30m R 5m;
//     create 150-193 → PoolProperties{ServiceLevel, Size, QosType, EncryptionType,
//     CoolAccess, CustomThroughputMibps}; CustomizeDiff 119-144).
//   - internal/services/netapp/validate/pool_name.go (regex ^[\da-zA-Z][-_\da-zA-Z]{2,63}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/capacitypools:
//     model_poolproperties.go (serviceLevel/size required; poolId/provisioningState/
//     totalThroughputMibps/utilizedThroughputMibps read-only),
//     constants.go (ServiceLevel Flexible/Premium/Standard/StandardZRS/Ultra,
//     QosType Auto/Manual, EncryptionType Single/Double),
//     id_capacitypool.go (type segment casing "capacityPools").
type NetAppPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppPool)(nil)

// NewNetAppPool returns knowledge for the capacityPools resource.
func NewNetAppPool() *NetAppPool {
	return &NetAppPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/capacityPools",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.serviceLevel"},
				{PropertyPath: "properties.encryptionType"},
			},
			// serviceLevel and size are required (no omitempty in the ARM model).
			RequiredFields: []string{
				"properties.serviceLevel",
				"properties.size",
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM PoolName: 3-64 chars, start alphanumeric.
					Regex:     `^[\da-zA-Z][-_\da-zA-Z]{2,63}$`,
					MinLength: 3,
					MaxLength: 64,
					Message:   "must be 3-64 characters, start with a letter or number, and contain only letters, numbers, underscores and hyphens",
				},
				{
					PropertyPath:  "properties.serviceLevel",
					AllowedValues: []string{"Flexible", "Premium", "Standard", "StandardZRS", "Ultra"},
				},
				{
					PropertyPath:  "properties.qosType",
					AllowedValues: []string{"Auto", "Manual"},
				},
				{
					PropertyPath:  "properties.encryptionType",
					AllowedValues: []string{"Single", "Double"},
				},
			},
			IntRules: []azwise.IntRule{
				{
					// custom_throughput_mibps IntAtLeast(128).
					PropertyPath: "properties.customThroughputMibps",
					MinValue:     azwise.Ptr(int64(128)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.qosType", Value: "Auto"},
				{PropertyPath: "properties.encryptionType", Value: "Single"},
				{PropertyPath: "properties.coolAccess", Value: false},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.poolId",
				"properties.provisioningState",
				"properties.totalThroughputMibps",
				"properties.utilizedThroughputMibps",
			},
			// NOTE: size_in_tb (IntBetween 1-2048 TiB) is converted to properties.size in
			// BYTES (sizeInTB * 1024^4) before send, so a declarative TiB IntRule on
			// properties.size would be wrong-unit and is intentionally omitted.
			// NOTE: CustomizeDiff enforces two value-conditional invariants: (1) cool_access
			// cannot be disabled once enabled (conditional ForceNew); (2) custom_throughput_mibps
			// is accepted only when qos_type=="Manual" AND service_level=="Flexible".
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppPool()) }
