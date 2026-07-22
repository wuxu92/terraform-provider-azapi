package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlFlexibleServerVirtualEndpoint provides resource knowledge for
// Microsoft.DBforPostgreSQL/flexibleServers/virtualEndpoints
// (azurerm_postgresql_flexible_server_virtual_endpoint).
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_flexible_server_virtual_endpoint_resource.go
//     :73-103  (Arguments: name ForceNew, source_server_id ForceNew, replica_server_id,
//     type ForceNew enum VirtualEndpointType)
//     :107,160,250,297 (timeouts: create/update 10m, read/delete 5m)
//     :140-146 (create mapping to virtualendpoints.VirtualEndpointResourceProperties;
//     endpointType from `type`, members from the replica server name)
//   - go-azure-sdk resource-manager/postgresql/2025-08-01/virtualendpoints:
//     model_virtualendpointresourceproperties.go (properties.endpointType / members),
//     constants.go:14-16 (VirtualEndpointType), id_virtualendpoint.go:125 (staticVirtualEndpoints)
//
// Not encoded (deliberate):
//   - properties.members is an array of flexible-server names derived from
//     replica_server_id; no scalar constraint to express.
type PostgresqlFlexibleServerVirtualEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlFlexibleServerVirtualEndpoint)(nil)

// NewPostgresqlFlexibleServerVirtualEndpoint returns knowledge for the
// flexibleServers/virtualEndpoints resource.
func NewPostgresqlFlexibleServerVirtualEndpoint() *PostgresqlFlexibleServerVirtualEndpoint {
	return &PostgresqlFlexibleServerVirtualEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/flexibleServers/virtualEndpoints",
			ApiVersions:  []string{"2025-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.endpointType"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 10 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 5 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.endpointType",
					AllowedValues: []string{"ReadWrite"},
					Message:       "type must be a valid VirtualEndpointType",
				},
			},
			RequiredFields: []string{
				"properties.endpointType",
				"properties.members",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlFlexibleServerVirtualEndpoint()) }
