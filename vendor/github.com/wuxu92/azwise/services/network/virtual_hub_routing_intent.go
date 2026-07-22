package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualHubRoutingIntent provides resource knowledge for
// Microsoft.Network/virtualHubs/routingIntent.
//
// Mirrors azurerm_virtual_hub_routing_intent. name, virtual_hub_id (parent) are
// envelope/parent references.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_hub_routing_intent_resource.go
//     (Arguments 57-105, timeouts 113/157/196/236)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     id_routingintent.go:110 (segment casing), model_routingintentproperties.go
//
// Not encoded (deliberate):
//   - routing_policy is a Required list expanding to properties.routingPolicies[*].
//     Its per-policy rules (name, destinations enum Internet/PrivateTraffic,
//     next_hop resource ID) live under an array element and cannot be lowered by
//     azwise/azapin, so they are documented here, not emitted.
type VirtualHubRoutingIntent struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualHubRoutingIntent)(nil)

// NewVirtualHubRoutingIntent returns knowledge for the routingIntent resource.
func NewVirtualHubRoutingIntent() *VirtualHubRoutingIntent {
	return &VirtualHubRoutingIntent{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualHubs/routingIntent",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.routingPolicies",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualHubRoutingIntent()) }
