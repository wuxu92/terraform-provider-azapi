package applicationinsights

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationInsights provides resource knowledge for
// Microsoft.Insights/components (Application Insights component).
//
// Mirrors azurerm_application_insights. name/resource_group_name/location live on
// the operational envelope; application_type is the only ForceNew property that
// maps to the ARM body (properties.Application_Type). AzureRM also hardcodes the
// top-level ARM kind to the application_type value.
//
// Deliberately not encoded here:
//   - workspace_id (properties.WorkspaceResourceId) carries a Log Analytics
//     workspace resource-ID validator (workspaces.ValidateWorkspaceID). That is a
//     semantic resource-ID rule and belongs in the resource customizer
//     (typegraph.Validator(validators.AzureResourceID)), not a declarative
//     StringRule.
//   - retention_in_days uses validation.IntInSlice (a discrete allow-set:
//     30/60/90/120/180/270/365/550/730), which IntRule (Min/Max only) cannot
//     express exactly; see the TODO below.
//   - daily_data_cap_in_gb and daily_data_cap_notifications_enabled are written
//     through the separate billing-features sub-API
//     (componentfeaturesandpricingapis, CurrentBillingFeatures/DataVolumeCap), not
//     the components body, so their rules do not belong on this resource.
//
// Sources:
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_resource.go:33-227
//     (schema: name/application_type/location ForceNew, retention/sampling/ip-mask/
//     local-auth/internet/force-storage defaults, timeouts 60m/5m/30m/30m)
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_resource.go:229-408
//     (Create mapping: Application_Type, SamplingPercentage, DisableIpMasking,
//     DisableLocalAuth, publicNetworkAccessFor{Ingestion,Query}, RetentionInDays,
//     WorkspaceResourceId, ForceCustomerStorageForProfiler; kind = application_type)
//   - go-azure-sdk resource-manager/applicationinsights/2020-02-02/componentsapis
//     model_applicationinsightscomponentproperties.go (ARM json tags) and
//     constants.go (ApplicationType/PublicNetworkAccessType enums)
type ApplicationInsights struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationInsights)(nil)

// NewApplicationInsights returns knowledge for the components resource.
func NewApplicationInsights() *ApplicationInsights {
	return &ApplicationInsights{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/components",
			ApiVersions:  []string{"2020-02-02"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.Application_Type"},
			},
			// TODO: retention_in_days is validation.IntInSlice([30 60 90 120 180 270
			// 365 550 730]) — a discrete allow-set that IntRule (Min/Max bounds only)
			// cannot represent. Bounding 30..730 would wrongly accept intermediate
			// values, so no IntRule is emitted for properties.RetentionInDays.
			FloatRules: []azwise.FloatRule{
				{PropertyPath: "properties.SamplingPercentage", MinValue: azwise.Ptr(float64(0)), MaxValue: azwise.Ptr(float64(100))},
			},
			StringRules: []azwise.StringRule{
				// AzureRM validation.StringInSlice accepts the extended kind set; the
				// 2020-02-02 SDK enum only defines "web"/"other" but best-effort-parses
				// the rest, and the Azure API accepts these documented component kinds.
				{PropertyPath: "properties.Application_Type", AllowedValues: []string{
					"web", "other", "java", "MobileCenter", "phone", "store", "ios", "Node.JS",
				}},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.RetentionInDays", Value: int64(90)},
				{PropertyPath: "properties.SamplingPercentage", Value: float64(100)},
				// ip_masking_enabled Default:true -> DisableIpMasking false (inverted).
				{PropertyPath: "properties.DisableIpMasking", Value: false},
				// local_authentication_enabled Default:true -> DisableLocalAuth false.
				{PropertyPath: "properties.DisableLocalAuth", Value: false},
				// internet_ingestion_enabled / internet_query_enabled Default:true.
				{PropertyPath: "properties.publicNetworkAccessForIngestion", Value: "Enabled"},
				{PropertyPath: "properties.publicNetworkAccessForQuery", Value: "Enabled"},
				{PropertyPath: "properties.ForceCustomerStorageForProfiler", Value: false},
			},
			// application_type is Required; AzureRM also hardcodes the top-level ARM
			// kind to the same value (insightProperties.Kind = application_type).
			RequiredFields: []string{
				"kind",
				"properties.Application_Type",
			},
			// app_id/instrumentation_key/connection_string are Computed-only in
			// AzureRM (server-generated identifiers), never user-settable.
			ComputedFields: []string{
				"properties.AppId",
				"properties.InstrumentationKey",
				"properties.ConnectionString",
			},
			SensitiveFields: []string{
				"properties.InstrumentationKey",
				"properties.ConnectionString",
			},
		},
	}
}

func init() { azwise.Register(NewApplicationInsights()) }
