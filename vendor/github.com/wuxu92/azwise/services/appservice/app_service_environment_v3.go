package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AppServiceEnvironmentV3 provides resource knowledge for Microsoft.Web/hostingEnvironments.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/app_service_environment_v3_resource.go:32-58
//     (AppServiceEnvironmentV3Model fields)
//   - terraform-provider-azurerm internal/services/appservice/app_service_environment_v3_resource.go:70-153
//     (schema: name/subnet_id/dedicated_host_count/internal_load_balancing_mode/zone_redundant ForceNew, validators, defaults, ConflictsWith)
//   - terraform-provider-azurerm internal/services/appservice/app_service_environment_v3_resource.go:248-354
//     (create: 6h timeout, expandCreateForAppServiceEnvironmentV3 -> appserviceenvironments.AppServiceEnvironmentResource)
//   - terraform-provider-azurerm internal/services/appservice/app_service_environment_v3_resource.go:439-536
//     (read 5m, delete/update 6h timeouts)
//   - terraform-provider-azurerm internal/services/web/validate/app_service_environment_name.go:11-14
//     (name regex)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/appserviceenvironments/model_appserviceenvironmentresource.go:6-14
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/appserviceenvironments/model_appserviceenvironment.go:6-27
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/appserviceenvironments/constants.go:613-618
//     (LoadBalancingMode full ARM enum)
//
// Intentionally skipped here:
//   - subnet_id: an ARM resource ID captured as the required body property
//     properties.virtualNetwork.id (its own validator is the SubnetID syntax).
//   - allow_new_private_endpoint_connections / remote_debugging_enabled: AzureRM
//     manages these via a separate networkingConfiguration PUT
//     (UpdateAseNetworkingConfiguration), not the hostingEnvironments create body.
//   - dedicated_host_count <-> zone_redundant ConflictsWith: AzureRM's expand always
//     sends BOTH properties.dedicatedHostCount (0 when unset) and
//     properties.zoneRedundant (false when unset), so a presence-based azwise
//     ConflictsWith would fire trivially. The real constraint is value-based
//     (dedicatedHostCount > 0 conflicts with zoneRedundant == true), which the
//     current RelationalRule kinds (presence-only) cannot express, so it is omitted.
//   - location: not a user argument; AzureRM derives it from the Virtual Network.
//   - AzureRM computed attributes (dns_suffix, *_inbound_ip_addresses, *_outbound_ip_addresses,
//     ip_ssl_address_count, pricing_tier, inbound_network_dependencies) are not listed
//     as ComputedFields because the AppServiceEnvironment create model shares the same
//     struct; stripping them would risk discarding valid raw ARM input.
type AppServiceEnvironmentV3 struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AppServiceEnvironmentV3)(nil)

// NewAppServiceEnvironmentV3 returns knowledge for the hostingEnvironments resource.
func NewAppServiceEnvironmentV3() *AppServiceEnvironmentV3 {
	return &AppServiceEnvironmentV3{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/hostingEnvironments",
			ApiVersions:  []string{"2023-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.virtualNetwork.id"},
				{PropertyPath: "properties.dedicatedHostCount"},
				{PropertyPath: "properties.internalLoadBalancingMode"},
				{PropertyPath: "properties.zoneRedundant"},
			},
			RequiredFields: []string{
				"properties.virtualNetwork.id",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 6 * time.Hour,
				Read:   5 * time.Minute,
				Update: 6 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:   `^[0-9a-zA-Z][-0-9a-zA-Z]{0,61}[0-9a-zA-Z]$`,
					Message: "must be 2-63 characters, alphanumeric and dashes, starting and ending with an alphanumeric character",
				},
				{
					PropertyPath:  "properties.internalLoadBalancingMode",
					AllowedValues: []string{"None", "Publishing", "Web", "Web, Publishing"},
					Message:       "must be a valid App Service Environment load-balancing mode",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.dedicatedHostCount",
					MinValue:     azwise.Ptr(int64(2)),
					MaxValue:     azwise.Ptr(int64(2)),
					Message:      "dedicated host count is currently limited to 2 physical hosts",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.internalLoadBalancingMode", Value: "None"},
				{PropertyPath: "properties.zoneRedundant", Value: false},
				{PropertyPath: "properties.clusterSettings"},
			},
		},
	}
}

func init() { azwise.Register(NewAppServiceEnvironmentV3()) }
