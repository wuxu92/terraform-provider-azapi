package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapsePrivateLinkHub provides resource knowledge for
// Microsoft.Synapse/privateLinkHubs.
//
// Mirrors azurerm_synapse_private_link_hub.
//
// Sources:
//   - internal/services/synapse/synapse_private_link_hub_resource.go
//     (schema 45-58: name ForceNew (PrivateLinkHubName validator); location/tags only;
//     timeouts Create/Update/Delete 30m Read 5m; create 82-85 → PrivateLinkHub{Location}).
//   - internal/services/synapse/validate/private_link_hub_name.go
//     (regex ^[a-z0-9]{1,45}$).
//   - go-azure-sdk resource-manager (track1) synapse model PrivateLinkHub;
//     resourceids.go PrivateLinkHub (top-level type, provider "Microsoft.Synapse",
//     segment casing "privateLinkHubs").
type SynapsePrivateLinkHub struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapsePrivateLinkHub)(nil)

func NewSynapsePrivateLinkHub() *SynapsePrivateLinkHub {
	return &SynapsePrivateLinkHub{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/privateLinkHubs",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// validate.PrivateLinkHubName
					Regex:     `^[a-z0-9]{1,45}$`,
					MinLength: 1,
					MaxLength: 45,
					Message:   "must be 1-45 chars and contain only lowercase letters or numbers",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapsePrivateLinkHub()) }
