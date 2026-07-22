// Copyright (c) HashiCorp, Inc.

package healthcare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Workspace provides resource knowledge for Microsoft.HealthcareApis/workspaces.
//
// Contributing TF resource:
//   - azurerm_healthcare_workspace
//
// Sources:
//   - AzureRM internal/services/healthcare/healthcare_workspace_resource.go
//     (schema :23-75, Create :77-112)
//   - AzureRM internal/services/healthcare/validate/workspace_name.go (name rule)
//   - go-azure-sdk .../healthcareapis/2024-03-31/workspaces:
//     id_workspace.go (segment casing: workspaces),
//     model_workspace.go, model_workspaceproperties.go, constants.go
type Workspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Workspace)(nil)

func NewWorkspace() *Workspace {
	return &Workspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HealthcareApis/workspaces",
			ApiVersions:  []string{"2024-03-31"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
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
					// Resource name (empty PropertyPath). validate.WorkspaceName:
					// 1-24 chars, lowercase letters or numbers only.
					Regex:     `^[a-z0-9]{1,24}$`,
					MinLength: 1,
					MaxLength: 24,
					Message:   "must be between 1 and 24 characters long and can contain only lowercase letters or numbers",
				},
			},
			// Azure-populated, read-only fields absent from the create body.
			// NOTE: properties.publicNetworkAccess IS present in the create model
			// (WorkspaceProperties) but is not surfaced by AzureRM; it is therefore
			// settable via AzAPI and intentionally NOT listed as computed.
			ComputedFields: []string{
				"properties.privateEndpointConnections",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewWorkspace()) }
