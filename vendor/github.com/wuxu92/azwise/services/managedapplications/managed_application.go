// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package managedapplications

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedApplication provides resource knowledge for Microsoft.Solutions/applications.
//
// Sources:
//   - terraform-provider-azurerm internal/services/managedapplications/managed_application_resource.go:55-202
//     (azurerm_managed_application schema: ForceNew, timeouts, validators; create mapping to
//     applications.Application / ApplicationProperties)
//   - terraform-provider-azurerm internal/services/managedapplications/validate/application_name.go
//     (ApplicationName regex ^[-\da-zA-Z]{3,64}$)
//   - terraform-provider-azurerm vendor/.../managedapplications/2021-07-01/applications/model_application.go:10-24
//     (Kind, Plan, Properties top-level layout)
//   - terraform-provider-azurerm vendor/.../managedapplications/2021-07-01/applications/model_applicationproperties.go:6-22
//   - terraform-provider-azurerm vendor/.../managedapplications/2021-07-01/applications/id_application.go:104-119
//     (ARM type casing: Microsoft.Solutions/applications)
//
// Intentionally skipped here:
//   - identity: envelope `identity` block, not a properties.* body field.
//   - parameter_values: optional+computed free-form JSON expanded into properties.parameters;
//     no fixed schema/default to express.
//   - outputs: computed read-only map (see ComputedFields).
type ManagedApplication struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedApplication)(nil)

func NewManagedApplication() *ManagedApplication {
	return &ManagedApplication{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Solutions/applications",
			ApiVersions:  []string{"2021-07-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				// kind is a top-level body field (not under properties).
				{PropertyPath: "kind"},
				// managed_resource_group_name -> properties.managedResourceGroupId (ForceNew).
				{PropertyPath: "properties.managedResourceGroupId"},
				// plan block is ForceNew in its entirety (top-level `plan` object).
				{PropertyPath: "plan"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name attribute (resource name).
					Regex:     `^[-\da-zA-Z]{3,64}$`,
					MinLength: 3,
					MaxLength: 64,
					Message:   "must be between 3 and 64 characters in length and contain only letters, numbers or hyphens",
				},
				{
					PropertyPath:  "kind",
					AllowedValues: []string{"MarketPlace", "ServiceCatalog"},
					Message:       "must be one of MarketPlace or ServiceCatalog",
				},
			},
			RequiredFields: []string{
				"kind",
				"properties.managedResourceGroupId",
			},
			ComputedFields: []string{
				"properties.outputs",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewManagedApplication()) }
