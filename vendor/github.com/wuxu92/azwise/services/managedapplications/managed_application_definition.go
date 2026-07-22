// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package managedapplications

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedApplicationDefinition provides resource knowledge for
// Microsoft.Solutions/applicationDefinitions.
//
// Sources:
//   - terraform-provider-azurerm internal/services/managedapplications/managed_application_definition_resource.go:25-164
//     (azurerm_managed_application_definition schema: ForceNew, timeouts, validators; create mapping
//     to applicationdefinitions.ApplicationDefinition / ApplicationDefinitionProperties)
//   - terraform-provider-azurerm internal/services/managedapplications/validate/application_definition_name.go
//     (ApplicationDefinitionName regex ^[^\W_]{3,64}$)
//   - terraform-provider-azurerm internal/services/managedapplications/validate/application_definition_display_name.go
//     (length 4-60), .../application_definition_description.go (max 200)
//   - terraform-provider-azurerm vendor/.../managedapplications/2021-07-01/applicationdefinitions/model_applicationdefinitionproperties.go:6-22
//   - terraform-provider-azurerm vendor/.../managedapplications/2021-07-01/applicationdefinitions/constants.go:105-109
//     (ApplicationLockLevel: CanNotDelete, None, ReadOnly)
//   - terraform-provider-azurerm vendor/.../managedapplications/2021-07-01/applicationdefinitions/id_applicationdefinition.go:104-118
//     (ARM type casing: Microsoft.Solutions/applicationDefinitions)
//
// Intentionally skipped here:
//   - authorization[].role_definition_id / service_principal_id: validation.IsUUID on
//     an array-element path (properties.authorizations[*].roleDefinitionId /
//     .principalId) — not expressible as a single body-path rule.
//   - create_ui_definition / main_template / package_file_uri: free-form JSON / URI
//     optional bodies with no fixed schema or default to express.
type ManagedApplicationDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedApplicationDefinition)(nil)

func NewManagedApplicationDefinition() *ManagedApplicationDefinition {
	return &ManagedApplicationDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Solutions/applicationDefinitions",
			ApiVersions:  []string{"2021-07-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				// lock_level -> properties.lockLevel (ForceNew).
				{PropertyPath: "properties.lockLevel"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name attribute (resource name): letters or numbers only.
					Regex:     `^[^\W_]{3,64}$`,
					MinLength: 3,
					MaxLength: 64,
					Message:   "must be between 3 and 64 characters in length and contain only letters or numbers",
				},
				{
					PropertyPath:  "properties.lockLevel",
					AllowedValues: []string{"CanNotDelete", "None", "ReadOnly"},
					Message:       "must be one of CanNotDelete, None or ReadOnly",
				},
				{
					PropertyPath: "properties.displayName",
					MinLength:    4,
					MaxLength:    60,
					Message:      "must be between 4 and 60 characters in length",
				},
				{
					PropertyPath: "properties.description",
					MaxLength:    200,
					Message:      "should not exceed 200 characters in length",
				},
			},
			RequiredFields: []string{
				"properties.displayName",
				"properties.lockLevel",
			},
			DefaultValues: []azwise.DefaultValue{
				// package_enabled Default:true -> properties.isEnabled true.
				{PropertyPath: "properties.isEnabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewManagedApplicationDefinition()) }
