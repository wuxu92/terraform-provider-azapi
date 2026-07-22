package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppVolumeQuotaRule provides resource knowledge for
// Microsoft.NetApp/netAppAccounts/capacityPools/volumes/volumeQuotaRules.
//
// Mirrors azurerm_netapp_volume_quota_rule.
//
// Sources:
//   - internal/services/netapp/netapp_volume_quota_rule_resource.go
//     (schema 41-79: name ForceNew+VolumeQuotaRuleName; volume_id parent ForceNew;
//     quota_type Required ForceNew enum; quota_size_in_kib IntAtLeast(4); quota_target
//     Optional ForceNew; timeouts Create 90m Update/Delete 120m Read 5m;
//     create 85-145 → VolumeQuotaRulesProperties{QuotaSizeInKiBs, QuotaType, QuotaTarget}).
//   - internal/services/netapp/validate/volume_quota_rule_name.go (regex ^[a-zA-Z][-_\da-zA-Z]{0,63}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/volumequotarules:
//     model_volumequotarulesproperties.go (quotaSizeInKiBs/quotaType/quotaTarget settable;
//     provisioningState read-only),
//     constants.go (QuotaType DefaultGroupQuota/DefaultUserQuota/IndividualGroupQuota/
//     IndividualUserQuota), id_volumequotarule.go (type segment casing "volumeQuotaRules").
type NetAppVolumeQuotaRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppVolumeQuotaRule)(nil)

// NewNetAppVolumeQuotaRule returns knowledge for the volumeQuotaRules resource.
func NewNetAppVolumeQuotaRule() *NetAppVolumeQuotaRule {
	return &NetAppVolumeQuotaRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/capacityPools/volumes/volumeQuotaRules",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 120 * time.Minute,
				Delete: 120 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.quotaType"},
				{PropertyPath: "properties.quotaTarget"},
			},
			// quotaSizeInKiBs and quotaType are Required.
			RequiredFields: []string{
				"properties.quotaSizeInKiBs",
				"properties.quotaType",
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM VolumeQuotaRuleName: 1-64 chars, start with a letter.
					Regex:     `^[a-zA-Z][-_\da-zA-Z]{0,63}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters, start with a letter, and contain only letters, numbers, underscores and hyphens",
				},
				{
					PropertyPath:  "properties.quotaType",
					AllowedValues: []string{"DefaultGroupQuota", "DefaultUserQuota", "IndividualGroupQuota", "IndividualUserQuota"},
				},
			},
			IntRules: []azwise.IntRule{
				{
					// quota_size_in_kib IntAtLeast(4).
					PropertyPath: "properties.quotaSizeInKiBs",
					MinValue:     azwise.Ptr(int64(4)),
				},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppVolumeQuotaRule()) }
