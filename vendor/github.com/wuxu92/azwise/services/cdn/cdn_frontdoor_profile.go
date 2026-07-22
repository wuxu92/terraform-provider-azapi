package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorProfile provides resource knowledge for Microsoft.Cdn/profiles
// (Azure Front Door Standard/Premium profile).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_profile.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_profile_resource.go
//     (schema L48-101, Create body L141-152, CustomizeDiff SKU-downgrade guard L103-116)
//   - internal/services/cdn/validate/front_door_name.go (FrontDoorName regex)
//   - go-azure-sdk resource-manager/cdn/2025-12-01/profiles:
//     model_profile.go (Sku is top-level `sku`), model_profileproperties.go,
//     model_profilepropertiesupdateparameters.go, constants.go (SkuName enum).
//
// Notes:
//   - name & resource_group_name are envelope-owned; not ARM-body ForceNew rules.
//   - sku_name (ForceNew) maps to top-level `sku.name` (not properties.*). Any SKU
//     change forces replacement; AzureRM additionally forbids Premium->Standard
//     downgrade via CustomizeDiff (a diff-time guard not expressible declaratively).
//   - SkuName enum uses the full ARM SDK set (shared with classic CDN); AzAPI sends
//     raw ARM values. AFD profiles in practice use Standard_AzureFrontDoor /
//     Premium_AzureFrontDoor.
type CdnFrontDoorProfile struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorProfile)(nil)

func NewCdnFrontDoorProfile() *CdnFrontDoorProfile {
	return &CdnFrontDoorProfile{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "sku.name"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorName
				{
					Regex:     `^[\da-zA-Z][-\da-zA-Z]{0,88}[\da-zA-Z]$`,
					MinLength: 2,
					MaxLength: 90,
					Message:   "must be 2-90 characters, begin and end with a letter or number, and contain only letters, numbers and hyphens",
				},
				// ── sku_name → sku.name ── full ARM SDK SkuName enum
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"Custom_Verizon",
						"Premium_AzureFrontDoor",
						"Premium_Verizon",
						"Standard_Akamai",
						"Standard_AvgBandWidth_ChinaCdn",
						"Standard_AzureFrontDoor",
						"Standard_ChinaCdn",
						"Standard_Microsoft",
						"Standard_955BandWidth_ChinaCdn",
						"StandardPlus_AvgBandWidth_ChinaCdn",
						"StandardPlus_ChinaCdn",
						"StandardPlus_955BandWidth_ChinaCdn",
						"Standard_Verizon",
					},
				},
			},
			IntRules: []azwise.IntRule{
				// response_timeout_seconds → properties.originResponseTimeoutSeconds
				{
					PropertyPath: "properties.originResponseTimeoutSeconds",
					MinValue:     azwise.Ptr(int64(16)),
					MaxValue:     azwise.Ptr(int64(240)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.originResponseTimeoutSeconds", Value: float64(120)},
			},
			ComputedFields: []string{
				"properties.frontDoorId",
				"properties.provisioningState",
				"properties.resourceState",
			},
			RequiredFields: []string{"sku.name"},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorProfile()) }
