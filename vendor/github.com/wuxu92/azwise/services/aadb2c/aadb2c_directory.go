package aadb2c

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AadB2cDirectory provides resource knowledge for
// Microsoft.AzureActiveDirectory/b2cDirectories.
//
// Mirrors azurerm_aadb2c_directory. Envelope fields live on the operational
// envelope: domain_name is the ARM resource name, resource_group_name the parent,
// and data_residency_location maps to the ARM envelope "location" field (a
// restricted data-residency enum, so its allowed values are still enforced here).
//
// Notes:
//   - country_code and display_name are Optional+Computed+ForceNew in AzureRM but
//     are mandatory at create time (enforced in the Create func) and are required,
//     non-pointer fields in the ARM CreateTenantProperties model, so they are listed
//     in RequiredFields and ForceNew.
//   - sku.tier is hardcoded by AzureRM to "A0" (the only SkuTier value); represented
//     as a RequiredField with a static default.
//   - sku_name uses the full ARM SkuName set (PremiumP1|PremiumP2|Standard); AzureRM
//     restricts the schema to PremiumP1/PremiumP2 but AzAPI sends raw ARM values.
//   - billing config and tenant id are server-computed read-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/aadb2c/aadb2c_directory_resource.go
//     (schema L53-131, create L134-210, read L245-306; data_residency_location
//     StringInSlice Location enum; sku_name StringInSlice; sku.tier hardcoded A0;
//     timeouts 30m/5m/30m/30m)
//   - go-azure-sdk resource-manager/aadb2c/2021-04-01-preview/tenants
//     CreateTenantProperties (countryCode/displayName required); Sku{name,tier};
//     TenantProperties.BillingConfig/TenantId server-computed; Location enum;
//     SkuName = PremiumP1|PremiumP2|Standard; SkuTier = A0
type AadB2cDirectory struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AadB2cDirectory)(nil)

// NewAadB2cDirectory returns knowledge for the b2cDirectories resource.
func NewAadB2cDirectory() *AadB2cDirectory {
	return &AadB2cDirectory{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureActiveDirectory/b2cDirectories",
			ApiVersions:  []string{"2021-04-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.countryCode"},
				{PropertyPath: "properties.displayName"},
			},
			StringRules: []azwise.StringRule{
				// domain_name (StringIsNotEmpty) is the ARM resource name.
				{MinLength: 1, Message: "domain_name must not be empty"},
				// data_residency_location -> envelope location (StringInSlice Location enum).
				{PropertyPath: "location", AllowedValues: []string{"Asia Pacific", "Australia", "Europe", "Global", "United States"}},
				// country_code (StringIsNotEmpty).
				{PropertyPath: "properties.countryCode", MinLength: 1, Message: "country_code must not be empty"},
				// display_name (StringIsNotEmpty).
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
				// sku_name (StringInSlice -> full ARM SkuName set).
				{PropertyPath: "sku.name", AllowedValues: []string{"PremiumP1", "PremiumP2", "Standard"}},
			},
			DefaultValues: []azwise.DefaultValue{
				// AzureRM hardcodes the SKU tier to "A0" (the only valid SkuTier).
				{PropertyPath: "sku.tier", Value: "A0"},
			},
			RequiredFields: []string{
				"properties.countryCode",
				"properties.displayName",
				"sku.name",
				"sku.tier",
			},
			ComputedFields: []string{
				"properties.billingConfig.billingType",
				"properties.billingConfig.effectiveStartDateUtc",
				"properties.tenantId",
			},
		},
	}
}

func init() { azwise.Register(NewAadB2cDirectory()) }
