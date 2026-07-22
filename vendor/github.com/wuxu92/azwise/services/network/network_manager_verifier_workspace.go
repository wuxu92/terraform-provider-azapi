package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerVerifierWorkspace provides resource knowledge for
// Microsoft.Network/networkManagers/verifierWorkspaces.
//
// Mirrors azurerm_network_manager_verifier_workspace. name and the parent
// network_manager_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_verifier_workspace_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/verifierworkspaces:
//     model_verifierworkspaceproperties.go
type NetworkManagerVerifierWorkspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerVerifierWorkspace)(nil)

func NewNetworkManagerVerifierWorkspace() *NetworkManagerVerifierWorkspace {
	return &NetworkManagerVerifierWorkspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/verifierWorkspaces",
			ApiVersions:  []string{"2025-01-01"},
			// location (commonschema.Location) replaces the workspace on change.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name validation (empty PropertyPath = the name attribute).
					Regex:     `^[a-zA-Z0-9\_\.\-]{1,64}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters and contain only letters, numbers, underscores, periods and hyphens",
				},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerVerifierWorkspace()) }
