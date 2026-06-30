package azwise

import "time"

// WebServerFarm provides resource knowledge for Microsoft.Web/serverfarms.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/service_plan_resource.go:69-141
//     (azurerm_service_plan schema: ForceNew fields, required fields, validators, defaults)
//   - terraform-provider-azurerm internal/services/appservice/service_plan_resource.go:166-233
//     (create mapping to appserviceplans.AppServicePlan)
//   - terraform-provider-azurerm internal/services/appservice/service_plan_resource.go:236-340
//     (read/update/delete timeouts and update mapping)
//   - terraform-provider-azurerm internal/services/appservice/service_plan_resource.go:352-391
//     (CustomizeDiff SKU cross-field checks and conditional zone_balancing_enabled ForceNew)
//   - terraform-provider-azurerm internal/services/appservice/service_plan_resource.go:394-441
//     (flatten mapping from ARM response)
//   - terraform-provider-azurerm internal/services/appservice/helpers/service_plan.go:27-84,210-220
//     (known Service Plan SKU names, zone-balancing and scale-out SKU helpers)
//   - terraform-provider-azurerm internal/services/appservice/validate/service_plan_name.go:11-18
//     (name regex)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-helpers/resourcemanager/commonids/app_service_environment.go:64-76,98-113
//     (App Service Environment ID validator and ARM ID shape)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema/location.go:11-19
//     (location ForceNew)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema/resource_group_name.go:11-17
//     (resource-group envelope ForceNew)
//   - terraform-provider-azurerm internal/services/web/app_service_plan_resource.go:29-47,56-150,155-245,334-378
//     (deprecated azurerm_app_service_plan cross-check: same timeouts/name regex; legacy-only kind/sku shape)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/appserviceplans/model_appserviceplan.go:6-16
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/appserviceplans/model_appserviceplanproperties.go:12-36
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/appserviceplans/model_skudescription.go:6-15
//
// Intentionally skipped here:
//   - resource_group_name: AzureRM marks it ForceNew, but it is an AzAPI envelope/ID
//     segment rather than a Microsoft.Web/serverfarms body property.
//   - os_type StringInSlice([Linux, Windows, WindowsContainer]): current AzureRM maps this
//     Terraform-only discriminator to properties.reserved/properties.hyperV. Those ARM
//     booleans are covered as ForceNew, but there is no single ARM string property to validate.
//   - os_type Required: for the same reason, RequiredFields does not force a particular
//     reserved/hyperV boolean combination on raw ARM users.
//   - app_service_environment_id isolated-SKU check, premium_plan_auto_scale_enabled SKU check,
//     maximum_elastic_worker_count SKU/autoscale check, and zone_balancing_enabled SKU-family
//     support check: these are cross-field SKU-category rules and cannot be represented by the
//     current azwise StringRule/IntRule/FloatRule types. The ASE ID syntax itself is captured.
//   - Deprecated azurerm_app_service_plan legacy-only kind, reserved/is_xenon relationships,
//     sku.tier/sku.size list shape, maximum_number_of_workers, and zone_redundant ForceNew:
//     azurerm_service_plan supersedes that resource and maps the same ARM type through
//     sku.name, sku.capacity, reserved/hyperV, and conditional zoneRedundant replacement.
//   - AzureRM computed attributes kind and reserved are not listed as ComputedFields because
//     the 2023-12-01 SDK AppServicePlan create/update model includes kind and
//     properties.reserved; stripping them from AzAPI bodies would discard valid raw ARM input.
type WebServerFarm struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*WebServerFarm)(nil)

// CheckForceNew extends BaseKnowledge with AzureRM's conditional zone-balancing replacement:
// enabling zone_balancing_enabled requires replacement when worker_count is less than 2.
func (s *WebServerFarm) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	oldZoneRedundant := webServerFarmBoolValue(oldBody, "properties.zoneRedundant")
	newZoneRedundant := webServerFarmBoolValue(newBody, "properties.zoneRedundant")
	if oldZoneRedundant || !newZoneRedundant {
		return false
	}

	capacity, ok := webServerFarmInt64Value(newBody, "sku.capacity")
	if !ok {
		return true
	}
	return capacity < 2
}

// NewWebServerFarm returns knowledge for the serverfarms resource.
func NewWebServerFarm() *WebServerFarm {
	return &WebServerFarm{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.Web/serverfarms",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.reserved"},
				{PropertyPath: "properties.hyperV"},
			},
			RequiredFields: []string{
				"sku.name",
			},
			TimeoutsConfig: &Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []StringRule{
				{
					Regex:     `^[0-9a-zA-Z-_]{1,60}$`,
					MinLength: 1,
					MaxLength: 60,
					Message:   "must contain only alphanumeric characters, dashes, and underscores, up to 60 characters",
				},
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"B1", "B2", "B3",
						"Y1",
						"EP1", "EP2", "EP3",
						"FC1",
						"F1",
						"I1", "I2", "I3",
						"I1v2", "I2v2", "I3v2", "I4v2", "I5v2", "I6v2",
						"I1mv2", "I2mv2", "I3mv2", "I4mv2", "I5mv2",
						"P1v2", "P2v2", "P3v2",
						"P0v3", "P1v3", "P2v3", "P3v3",
						"P1mv3", "P2mv3", "P3mv3", "P4mv3", "P5mv3",
						"P0v4", "P1v4", "P2v4", "P3v4",
						"P1mv4", "P2mv4", "P3mv4", "P4mv4", "P5mv4",
						"D1", "SHARED",
						"WS1", "WS2", "WS3",
					},
					Message: "must be a known AzureRM Service Plan SKU name",
				},
				{
					PropertyPath: "properties.hostingEnvironmentProfile.id",
					Regex:        `(?i)^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft\.Web/hostingEnvironments/[^/]+$`,
					Message:      "must be an App Service Environment resource ID",
				},
			},
			IntRules: []IntRule{
				{
					PropertyPath: "sku.capacity",
					MinValue:     ptr(int64(1)),
					Message:      "worker count must be at least 1",
				},
				{
					PropertyPath: "properties.maximumElasticWorkerCount",
					MinValue:     ptr(int64(0)),
					Message:      "maximum elastic worker count must be non-negative",
				},
			},
			DefaultValues: []DefaultValue{
				{PropertyPath: "properties.perSiteScaling", Value: false},
				{PropertyPath: "properties.elasticScaleEnabled", Value: false},
				{PropertyPath: "properties.zoneRedundant", Value: false},
				{PropertyPath: "sku.capacity"},
				{PropertyPath: "properties.maximumElasticWorkerCount"},
			},
		},
	}
}

func webServerFarmBoolValue(body map[string]interface{}, path string) bool {
	value, ok := extractNestedValue(body, path).(bool)
	return ok && value
}

func webServerFarmInt64Value(body map[string]interface{}, path string) (int64, bool) {
	value := extractNestedValue(body, path)
	number, ok := toFloat64(value)
	if !ok {
		return 0, false
	}
	return int64(number), true
}
