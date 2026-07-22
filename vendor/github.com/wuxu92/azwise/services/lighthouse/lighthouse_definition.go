// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lighthouse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LighthouseDefinition provides resource knowledge for
// Microsoft.ManagedServices/registrationDefinitions.
//
// Sources:
//   - terraform-provider-azurerm internal/services/lighthouse/lighthouse_definition_resource.go:26-406
//     (azurerm_lighthouse_definition schema: ForceNew, timeouts, validators; create mapping in
//     resourceLighthouseDefinitionCreateUpdate to registrationdefinitions.RegistrationDefinition)
//   - vendor/.../managedservices/2022-10-01/registrationdefinitions/model_registrationdefinition.go:10-17
//   - vendor/.../registrationdefinitions/model_registrationdefinitionproperties.go:6-16
//   - vendor/.../registrationdefinitions/model_plan.go:6-11
//   - vendor/.../registrationdefinitions/constants.go:14-17 (MultiFactorAuthProvider: Azure, None)
//
// ARM resource name = registrationDefinitionId (a UUID, TF `lighthouse_definition_id`);
// TF `name` is the display name mapped to properties.registrationDefinitionName.
//
// Intentionally skipped here:
//   - scope / lighthouse_definition_id: envelope — scope + resource-name segments of the ARM ID,
//     not properties.* body fields.
//   - authorization[] and eligible_authorization[] sub-field validators (principal_id/role_definition_id
//     IsUUID, principal_display_name StringIsNotEmpty, delegated_role_definition_ids IsUUID,
//     just_in_time_access_policy.multi_factor_auth_provider enum [Azure], maximum_activation_duration
//     ISO8601 default "PT8H", approver.*): array-element paths under
//     properties.authorizations[*] / properties.eligibleAuthorizations[*] — not declaratively expressible.
//   - authorization MinItems:1 — ArrayRule supports only MaxItems; presence is covered by
//     RequiredFields (properties.authorizations).
type LighthouseDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LighthouseDefinition)(nil)

func NewLighthouseDefinition() *LighthouseDefinition {
	return &LighthouseDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ManagedServices/registrationDefinitions",
			ApiVersions:  []string{"2022-10-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				// name (ForceNew) -> properties.registrationDefinitionName
				{PropertyPath: "properties.registrationDefinitionName"},
				// managing_tenant_id (ForceNew) -> properties.managedByTenantId
				{PropertyPath: "properties.managedByTenantId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ARM resource name = registrationDefinitionId; TF lighthouse_definition_id IsUUID.
				{
					Regex:   `^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$`,
					Message: "Lighthouse definition name (registration definition id) must be a UUID",
				},
				// name StringIsNotEmpty -> properties.registrationDefinitionName
				{
					PropertyPath: "properties.registrationDefinitionName",
					MinLength:    1,
					Message:      "must not be empty",
				},
				// managing_tenant_id IsUUID -> properties.managedByTenantId
				{
					PropertyPath: "properties.managedByTenantId",
					Regex:        `^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$`,
					Message:      "managing_tenant_id must be a UUID",
				},
				// plan.* fields StringIsNotEmpty (Plan is a top-level ARM object, not under properties).
				{PropertyPath: "plan.name", MinLength: 1, Message: "plan name must not be empty"},
				{PropertyPath: "plan.publisher", MinLength: 1, Message: "plan publisher must not be empty"},
				{PropertyPath: "plan.product", MinLength: 1, Message: "plan product must not be empty"},
				{PropertyPath: "plan.version", MinLength: 1, Message: "plan version must not be empty"},
			},
			RequiredFields: []string{
				"properties.authorizations",
				"properties.managedByTenantId",
				"properties.registrationDefinitionName",
			},
			ComputedFields: []string{
				// Server-populated response fields (echoed/read-only on the shared PUT/GET model).
				"properties.provisioningState",
				"properties.managedByTenantName",
				"properties.manageeTenantId",
				"properties.manageeTenantName",
			},
		},
	}
}

func init() { azwise.Register(NewLighthouseDefinition()) }
