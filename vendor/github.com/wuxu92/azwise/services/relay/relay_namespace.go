package relay

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RelayNamespace provides resource knowledge for Microsoft.Relay/namespaces.
//
// Contributing Terraform resource: azurerm_relay_namespace.
//
// Sources:
//   - terraform-provider-azurerm internal/services/relay/relay_namespace_resource.go
//     (schema L25-95, Create body L120-129)
//   - go-azure-sdk resource-manager/relay/2021-11-01/namespaces:
//     model_relaynamespace.go, model_relaynamespaceproperties.go, model_sku.go, constants.go
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not emitted as body rules.
//   - sku_name maps to sku.name (SkuNameStandard is the only allowed value); AzureRM also
//     mirrors it into sku.tier.
//   - metric_id / primary_connection_string / secondary_connection_string / primary_key /
//     secondary_key are computed data-plane values (ListKeys / read-only), not settable
//     body properties; omitted rather than listed as ComputedFields on the namespace body.
type RelayNamespace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RelayNamespace)(nil)

// NewRelayNamespace returns knowledge for the namespaces resource.
func NewRelayNamespace() *RelayNamespace {
	return &RelayNamespace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Relay/namespaces",
			ApiVersions:  []string{"2021-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    6,
					MaxLength:    50,
					Message:      "relay namespace name must be between 6 and 50 characters long",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Standard"},
					Message:       "sku_name must be Standard",
				},
			},
			ComputedFields: []string{
				"properties.metricId",
				"properties.serviceBusEndpoint",
				"properties.provisioningState",
				"properties.status",
				"properties.createdAt",
				"properties.updatedAt",
			},
			RequiredFields: []string{
				"sku.name",
			},
		},
	}
}

func init() { azwise.Register(NewRelayNamespace()) }
