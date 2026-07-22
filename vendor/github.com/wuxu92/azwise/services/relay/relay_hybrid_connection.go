package relay

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RelayHybridConnection provides resource knowledge for
// Microsoft.Relay/namespaces/hybridConnections.
//
// Contributing Terraform resource: azurerm_relay_hybrid_connection.
//
// Sources:
//   - terraform-provider-azurerm internal/services/relay/relay_hybrid_connection_resource.go
//     (schema L22-70, Create body L95-103)
//   - go-azure-sdk resource-manager/relay/2021-11-01/hybridconnections:
//     model_hybridconnection.go, model_hybridconnectionproperties.go, id_hybridconnection.go
//
// Notes:
//   - name/relay_namespace_name/resource_group_name are envelope / parent-reference
//     fields; not emitted as body rules.
//   - requires_client_authorization is ForceNew and defaults to true →
//     properties.requiresClientAuthorization.
type RelayHybridConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RelayHybridConnection)(nil)

// NewRelayHybridConnection returns knowledge for the namespaces/hybridConnections resource.
func NewRelayHybridConnection() *RelayHybridConnection {
	return &RelayHybridConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Relay/namespaces/hybridConnections",
			ApiVersions:  []string{"2021-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.requiresClientAuthorization"},
			},
			ComputedFields: []string{
				"properties.listenerCount",
				"properties.createdAt",
				"properties.updatedAt",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.requiresClientAuthorization", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewRelayHybridConnection()) }
