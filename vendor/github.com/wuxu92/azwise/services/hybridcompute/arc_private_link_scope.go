package hybridcompute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ArcPrivateLinkScope provides resource knowledge for
// Microsoft.HybridCompute/privateLinkScopes.
//
// Contributing TF resource: azurerm_arc_private_link_scope.
//
// Sources:
//   - AzureRM internal/services/hybridcompute/arc_private_link_scope_resource.go
//     (model :22-28, schema :35-52, Create body :90-103)
//   - go-azure-sdk .../hybridcompute/2022-11-10/privatelinkscopes:
//     model_hybridcomputeprivatelinkscope.go,
//     model_hybridcomputeprivatelinkscopeproperties.go (publicNetworkAccess :10,
//     privateEndpointConnections/privateLinkScopeId/provisioningState read-only),
//     constants.go (PublicNetworkAccessType :14-17)
type ArcPrivateLinkScope struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ArcPrivateLinkScope)(nil)

func NewArcPrivateLinkScope() *ArcPrivateLinkScope {
	return &ArcPrivateLinkScope{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HybridCompute/privateLinkScopes",
			ApiVersions:  []string{"2022-11-10"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
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
					// name: validation.StringIsNotEmpty.
					MinLength: 1,
					Message:   "must not be empty",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// public_network_access_enabled schema Default:false → ARM "Disabled".
				{PropertyPath: "properties.publicNetworkAccess", Value: "Disabled"},
			},
			// privateEndpointConnections/privateLinkScopeId/provisioningState are
			// Azure-populated read-only fields (present on GET, absent from the body).
			ComputedFields: []string{
				"properties.privateEndpointConnections",
				"properties.privateLinkScopeId",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewArcPrivateLinkScope()) }
