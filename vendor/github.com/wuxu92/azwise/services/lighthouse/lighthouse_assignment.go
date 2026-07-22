// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lighthouse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LighthouseAssignment provides resource knowledge for
// Microsoft.ManagedServices/registrationAssignments.
//
// Sources:
//   - terraform-provider-azurerm internal/services/lighthouse/lighthouse_assignment_resource.go:26-150
//     (azurerm_lighthouse_assignment schema: ForceNew, timeouts, validators; create mapping in
//     resourceLighthouseAssignmentCreate to registrationassignments.RegistrationAssignment)
//   - vendor/.../managedservices/2022-10-01/registrationassignments/model_registrationassignment.go:10-16
//   - vendor/.../registrationassignments/model_registrationassignmentproperties.go:6-10
//
// ARM resource name = registrationAssignmentId (a UUID, TF `name`, server-generated when omitted).
// The assignment has no update path — every field is ForceNew.
//
// Intentionally skipped here:
//   - scope: envelope — scope segment of the ARM ID, not a properties.* body field.
//   - lighthouse_definition_id ValidateScopedRegistrationDefinitionID: a semantic scoped-resource-ID
//     validator (belongs in an azapin customizer, not a declarative StringRule).
//   - name Optional+Computed default: the server generates the UUID resource name when omitted;
//     the name is the ARM resource-name segment, not a body property, so no DefaultValue applies.
type LighthouseAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LighthouseAssignment)(nil)

func NewLighthouseAssignment() *LighthouseAssignment {
	return &LighthouseAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ManagedServices/registrationAssignments",
			ApiVersions:  []string{"2022-10-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				// lighthouse_definition_id (ForceNew) -> properties.registrationDefinitionId
				{PropertyPath: "properties.registrationDefinitionId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ARM resource name = registrationAssignmentId; TF name IsUUID.
				{
					Regex:   `^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$`,
					Message: "Lighthouse assignment name (registration assignment id) must be a UUID",
				},
			},
			RequiredFields: []string{
				"properties.registrationDefinitionId",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				// Read-only expansion of the referenced definition (only populated when
				// ExpandRegistrationDefinition is requested on GET).
				"properties.registrationDefinition",
			},
		},
	}
}

func init() { azwise.Register(NewLighthouseAssignment()) }
