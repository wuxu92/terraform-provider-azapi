package elasticsan

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ElasticSan provides resource knowledge for Microsoft.ElasticSan/elasticSans.
//
// Mirrors azurerm_elastic_san.
//
// Sources:
//   - terraform-provider-azurerm internal/services/elasticsan/elastic_san_resource.go
//     schema Arguments (65-124: name ForceNew, zones OptionalForceNew, base_size_in_tib
//     IntBetween(1,100), extended_size_in_tib IntBetween(1,100), sku.name ForceNew +
//     StringInSlice(PossibleValuesForSkuName), sku.tier StringInSlice(PossibleValuesForSkuTier)
//     default Premium), Attributes (126-153: total_iops/total_mbps/total_size_in_tib/
//     total_volume_size_in_gib/volume_group_count computed), CustomizeDiff (155-179: zone
//     vs Premium_ZRS conflict; base/extended size cannot shrink), ExpandSku (342-356);
//     timeouts Create/Update/Delete 30m, Read 5m.
//   - internal/services/elasticsan/validate/elastic_san_name.go (name regex 3-24 chars).
//   - go-azure-sdk resource-manager/elasticsan/2023-01-01/elasticsans:
//     ElasticSanProperties (baseSizeTiB/sku required; totalIops/totalMBps/totalSizeTiB/
//     totalVolumeSizeGiB/volumeGroupCount/provisioningState read-only), Sku{Name,Tier},
//     constants.go (SkuName Premium_LRS/Premium_ZRS, SkuTier Premium),
//     id_elasticsan.go Segments (type segment casing "elasticSans").
type ElasticSan struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ElasticSan)(nil)

// NewElasticSan returns knowledge for the elasticSans resource.
func NewElasticSan() *ElasticSan {
	return &ElasticSan{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ElasticSan/elasticSans",
			ApiVersions:  []string{"2023-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.availabilityZones"},
				{PropertyPath: "properties.sku.name"},
			},
			// baseSizeTiB and sku.name are required (no omitempty in the ARM model).
			RequiredFields: []string{
				"properties.baseSizeTiB",
				"properties.sku.name",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM ElasticSanName:
					// 3-24 chars, lowercase alphanumeric plus '_'/'-', must start and
					// end with a lowercase letter or number.
					Regex:     `^[a-z0-9][a-z0-9_-]{1,22}[a-z0-9]$`,
					MinLength: 3,
					MaxLength: 24,
					Message:   "must be 3-24 characters, lowercase letters, numbers, underscores and hyphens only, and start/end with a lowercase letter or number",
				},
				{
					PropertyPath:  "properties.sku.name",
					AllowedValues: []string{"Premium_LRS", "Premium_ZRS"},
				},
				{
					PropertyPath:  "properties.sku.tier",
					AllowedValues: []string{"Premium"},
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.baseSizeTiB",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(100)),
				},
				{
					PropertyPath: "properties.extendedCapacitySizeTiB",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(100)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.sku.tier", Value: "Premium"},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.totalIops",
				"properties.totalMBps",
				"properties.totalSizeTiB",
				"properties.totalVolumeSizeGiB",
				"properties.volumeGroupCount",
				"properties.provisioningState",
			},
			// NOTE: CustomizeDiff enforces two value-conditional invariants that cannot be
			// expressed as declarative rules: (1) availabilityZones are rejected when
			// sku.name == Premium_ZRS; (2) baseSizeTiB / extendedCapacitySizeTiB may only
			// grow (new value >= old value) on update.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewElasticSan()) }
