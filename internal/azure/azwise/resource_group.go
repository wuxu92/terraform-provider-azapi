package azwise

import "time"

// ResourceGroup provides resource knowledge for Microsoft.Resources/resourceGroups.
//
// Mirrors azurerm_resource_group. The resource-group *name* carries AzureRM's
// validation, but the name is a DeployTimeConstant in the ARM body that the
// generator lifts onto the operational envelope — so that rule is attached by the
// customizer (internal/native/generator/customizers/resource_group.go), not here.
// Body-path knowledge (location is ForceNew, managed_by must be non-empty) lives
// below.
//
// Sources:
//   - terraform-provider-azurerm internal/services/resource/resource_group_resource.go:33-66
//     (name=ResourceGroupName ForceNew, location=Location ForceNew, managed_by=StringIsNotEmpty)
//   - go-azure-helpers resourcemanager/resourcegroups/validate.go:12-31 (name charset/length)
//   - go-azure-helpers resourcemanager/commonschema/{resource_group_name,location}.go (ForceNew)
type ResourceGroup struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*ResourceGroup)(nil)

// NewResourceGroup returns knowledge for the resourceGroups resource.
func NewResourceGroup() *ResourceGroup {
	return &ResourceGroup{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.Resources/resourceGroups",
			ApiVersions:  []string{"2025-04-01"},
			// Changing a resource group's location replaces it (commonschema.Location
			// is ForceNew). The name is likewise replace-only, but it lives on the
			// envelope, which is already Required+RequiresReplace by construction.
			ForceNew: []ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			// managed_by is Optional but, when set, must be non-empty
			// (azurerm validation.StringIsNotEmpty).
			StringRules: []StringRule{
				{
					PropertyPath: "managedBy",
					MinLength:    1,
					Message:      "managed_by must not be empty when set",
				},
			},
		},
	}
}
