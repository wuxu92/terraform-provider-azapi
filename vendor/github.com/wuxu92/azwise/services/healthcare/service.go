// Copyright (c) HashiCorp, Inc.

package healthcare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Service provides resource knowledge for the legacy
// Microsoft.HealthcareApis/services resource type.
//
// Contributing TF resource:
//   - azurerm_healthcare_service
//
// Sources:
//   - AzureRM internal/services/healthcare/healthcare_service_resource.go
//     (schema :30-231, CreateUpdate :233-303, expand cors/auth/cosmos :415-479)
//   - go-azure-sdk .../healthcareapis/2022-12-01/resource:
//     id_service.go (segment casing: services),
//     model_servicesdescription.go, model_servicesproperties.go,
//     model_servicecosmosdbconfigurationinfo.go, model_servicecorsconfigurationinfo.go,
//     model_serviceauthenticationconfigurationinfo.go, constants.go
type Service struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Service)(nil)

func NewService() *Service {
	return &Service{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HealthcareApis/services",
			ApiVersions:  []string{"2022-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				// cosmosdb_key_vault_key_versionless_id -> cosmosDbConfiguration.keyVaultKeyUri.
				{PropertyPath: "properties.cosmosDbConfiguration.keyVaultKeyUri"},
				// NOTE: resource_group_name is an envelope reference; omitted.
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). validation.StringIsNotEmpty.
					MinLength: 1,
					Message:   "name must not be empty",
				},
				{
					// kind -> StringInSlice(fhir, fhir-R4, fhir-Stu3). Full ARM SDK set (Kind).
					PropertyPath:  "kind",
					AllowedValues: []string{"fhir", "fhir-R4", "fhir-Stu3"},
					Message:       "kind must be one of fhir, fhir-R4, or fhir-Stu3",
				},
				// NOTE: access_policy_object_ids (validation.IsUUID) and
				// cors_configuration.allowed_methods (StringInSlice DELETE/GET/HEAD/MERGE/
				// POST/OPTIONS/PUT/PATCH) validate array elements
				// (properties.accessPolicies[*].objectId, .corsConfiguration.methods[*])
				// and are not representable as declarative StringRules — intentionally omitted.
			},
			IntRules: []azwise.IntRule{
				// cosmosdb_throughput -> validation.IntBetween(1, 100000).
				{PropertyPath: "properties.cosmosDbConfiguration.offerThroughput", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(100000)), Message: "cosmosdb_throughput must be between 1 and 100000"},
				// cors_configuration.max_age_in_seconds -> validation.IntBetween(0, 2000000000).
				{PropertyPath: "properties.corsConfiguration.maxAge", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(2000000000)), Message: "max_age_in_seconds must be between 0 and 2000000000"},
			},
			ArrayRules: []azwise.ArrayRule{
				// cors_configuration.allowed_origins / allowed_headers / allowed_methods MaxItems: 64.
				{PropertyPath: "properties.corsConfiguration.origins", MaxItems: 64, Message: "cors_configuration.allowed_origins allows at most 64 entries"},
				{PropertyPath: "properties.corsConfiguration.headers", MaxItems: 64, Message: "cors_configuration.allowed_headers allows at most 64 entries"},
				{PropertyPath: "properties.corsConfiguration.methods", MaxItems: 64, Message: "cors_configuration.allowed_methods allows at most 64 entries"},
			},
			DefaultValues: []azwise.DefaultValue{
				// kind default fhir.
				{PropertyPath: "kind", Value: "fhir"},
				// cosmosdb_throughput default 1000.
				{PropertyPath: "properties.cosmosDbConfiguration.offerThroughput", Value: 1000},
				// public_network_access_enabled default true -> PublicNetworkAccessEnabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			AtLeastOneOf: []azwise.RelationalRule{
				// authentication_configuration: at least one of authority/audience/smart_proxy_enabled.
				{Paths: []string{
					"properties.authenticationConfiguration.authority",
					"properties.authenticationConfiguration.audience",
					"properties.authenticationConfiguration.smartProxyEnabled",
				}},
				// cors_configuration: at least one of origins/headers/methods/maxAge/allowCredentials.
				{Paths: []string{
					"properties.corsConfiguration.origins",
					"properties.corsConfiguration.headers",
					"properties.corsConfiguration.methods",
					"properties.corsConfiguration.maxAge",
					"properties.corsConfiguration.allowCredentials",
				}},
			},
			// Azure-populated, read-only fields absent from the create body.
			ComputedFields: []string{
				"properties.privateEndpointConnections",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewService()) }
