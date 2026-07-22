// Copyright (c) HashiCorp, Inc.

package healthcare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FhirService provides resource knowledge for
// Microsoft.HealthcareApis/workspaces/fhirservices.
//
// Contributing TF resource:
//   - azurerm_healthcare_fhir_service
//
// Sources:
//   - AzureRM internal/services/healthcare/healthcare_fhir_service_resource.go
//     (schema :31-231, Create :233-304)
//   - AzureRM internal/services/healthcare/validate/fhirservice_name.go (name rule)
//   - go-azure-sdk .../healthcareapis/2022-12-01/fhirservices:
//     id_fhirservice.go (segment casing: workspaces/fhirServices),
//     model_fhirservice.go, model_fhirserviceproperties.go,
//     model_fhirservicecorsconfiguration.go, model_fhirserviceauthenticationconfiguration.go,
//     model_fhirserviceexportconfiguration.go, constants.go
type FhirService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FhirService)(nil)

func NewFhirService() *FhirService {
	return &FhirService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HealthcareApis/workspaces/fhirservices",
			ApiVersions:  []string{"2022-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "kind"},
				// NOTE: workspace_id / resource_group_name are envelope/parent references; omitted.
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). validate.FhirServiceName:
					// 3-24 chars, starts/ends alphanumeric, dashes allowed.
					Regex:   `^[0-9a-zA-Z][-0-9a-zA-Z]{1,22}[0-9a-zA-Z]$`,
					Message: "must be 3-24 characters, start and end with a letter or number, and contain only letters, numbers, and dashes",
				},
				{
					// kind -> StringInSlice(fhir-R4, fhir-Stu3). Full ARM SDK set (FhirServiceKind).
					PropertyPath:  "kind",
					AllowedValues: []string{"fhir-R4", "fhir-Stu3"},
					Message:       "kind must be one of fhir-R4 or fhir-Stu3",
				},
				{
					// authentication.authority -> validation.StringIsNotEmpty.
					PropertyPath: "properties.authenticationConfiguration.authority",
					MinLength:    1,
					Message:      "authentication.authority must not be empty",
				},
				// NOTE: access_policy_object_ids (validation.IsUUID),
				// container_registry_login_server_url / oci_artifact.login_server
				// (StringIsNotEmpty), and cors.allowed_methods (StringInSlice DELETE/GET/
				// HEAD/MERGE/POST/OPTIONS/PUT/PATCH) all validate array elements
				// (properties.accessPolicies[*].objectId, .acrConfiguration.loginServers[*],
				// .corsConfiguration.methods[*]) and are not representable as declarative
				// StringRules — intentionally omitted.
			},
			IntRules: []azwise.IntRule{
				// cors.max_age_in_seconds -> validation.IntBetween(0, 2000000000).
				{PropertyPath: "properties.corsConfiguration.maxAge", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(2000000000)), Message: "max_age_in_seconds must be between 0 and 2000000000"},
			},
			ArrayRules: []azwise.ArrayRule{
				// cors.allowed_origins / allowed_headers / allowed_methods MaxItems: 64.
				{PropertyPath: "properties.corsConfiguration.origins", MaxItems: 64, Message: "cors.allowed_origins allows at most 64 entries"},
				{PropertyPath: "properties.corsConfiguration.headers", MaxItems: 64, Message: "cors.allowed_headers allows at most 64 entries"},
				{PropertyPath: "properties.corsConfiguration.methods", MaxItems: 64, Message: "cors.allowed_methods allows at most 64 entries"},
			},
			DefaultValues: []azwise.DefaultValue{
				// kind default fhir-R4.
				{PropertyPath: "kind", Value: "fhir-R4"},
				// cors.credentials_allowed default false.
				{PropertyPath: "properties.corsConfiguration.allowCredentials", Value: false},
			},
			RequiredFields: []string{
				// authentication block is Required; authority and audience are Required.
				"properties.authenticationConfiguration.authority",
				"properties.authenticationConfiguration.audience",
			},
			// Azure-populated, read-only fields absent from the create body.
			// NOTE: properties.publicNetworkAccess is Computed-only in AzureRM but IS
			// present in the create model (FhirServiceProperties); it is settable via
			// AzAPI and intentionally NOT listed as computed.
			ComputedFields: []string{
				"properties.eventState",
				"properties.privateEndpointConnections",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewFhirService()) }
