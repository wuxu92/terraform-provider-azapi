// Copyright (c) HashiCorp, Inc.

package healthcare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FhirDestination provides resource knowledge for
// Microsoft.HealthcareApis/workspaces/iotconnectors/fhirdestinations.
//
// Contributing TF resource:
//   - azurerm_healthcare_medtech_service_fhir_destination
//
// Sources:
//   - AzureRM internal/services/healthcare/healthcare_medtech_service_fhir_destination_resource.go
//     (schema :30-94, Create :96-141)
//   - AzureRM internal/services/healthcare/validate/medtech_service_name.go (name rule)
//   - go-azure-sdk .../healthcareapis/2022-12-01/iotconnectors:
//     id_fhirdestination.go (segment casing: workspaces/iotConnectors/fhirDestinations),
//     model_iotfhirdestination.go, model_iotfhirdestinationproperties.go, constants.go
type FhirDestination struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FhirDestination)(nil)

func NewFhirDestination() *FhirDestination {
	return &FhirDestination{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HealthcareApis/workspaces/iotconnectors/fhirdestinations",
			ApiVersions:  []string{"2022-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				// NOTE: medtech_service_id is the parent resource reference (not a body path); omitted.
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). validate.MedTechServiceName:
					// 3-24 chars, starts/ends alphanumeric, dashes allowed.
					Regex:   `^[0-9a-zA-Z][-0-9a-zA-Z]{1,22}[0-9a-zA-Z]$`,
					Message: "must be 3-24 characters, start and end with a letter or number, and contain only letters, numbers, and dashes",
				},
				{
					// destination_identity_resolution_type -> StringInSlice(Create, Lookup).
					// Full ARM SDK set (IotIdentityResolutionType).
					PropertyPath:  "properties.resourceIdentityResolutionType",
					AllowedValues: []string{"Create", "Lookup"},
					Message:       "destination_identity_resolution_type must be one of Create or Lookup",
				},
				// NOTE: destination_fhir_service_id -> fhirservices.ValidateFhirServiceID
				// is a semantic ARM resource-ID validator, not representable as a declarative
				// StringRule — intentionally omitted.
			},
			RequiredFields: []string{
				// destination_fhir_service_id / destination_identity_resolution_type /
				// destination_fhir_mapping_json are all Required (non-pointer SDK fields).
				"properties.fhirServiceResourceId",
				"properties.resourceIdentityResolutionType",
				"properties.fhirMapping",
			},
			// Azure-populated, read-only field absent from the create body.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewFhirDestination()) }
