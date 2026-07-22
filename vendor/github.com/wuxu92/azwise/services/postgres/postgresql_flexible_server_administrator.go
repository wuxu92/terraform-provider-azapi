package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlFlexibleServerAdministrator provides resource knowledge for
// Microsoft.DBforPostgreSQL/flexibleServers/administrators
// (azurerm_postgresql_flexible_server_active_directory_administrator).
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_flexible_server_active_directory_administrator_resource.go
//     :24-81   (schema: server_name/object_id/principal_name/principal_type/tenant_id all
//     Required+ForceNew; object_id & tenant_id IsUUID; principal_type enum)
//     :34-38   (timeouts: create/delete 30m, read 5m)
//     :111-117 (create mapping to AdministratorMicrosoftEntraPropertiesForAdd)
//   - go-azure-sdk resource-manager/postgresql/2025-08-01/administratormicrosoftentras:
//     model_administratormicrosoftentrapropertiesforadd.go (properties.principalName /
//     principalType / tenantId), constants.go:14-19 (PrincipalType),
//     id_administrator.go:125 (staticAdministrators; object_id is the name segment)
//
// Not encoded (deliberate):
//   - object_id (the administrators id name segment) and tenant_id use validation.IsUUID,
//     a semantic validator (validators.UUID in an azapin customizer), not declarative.
type PostgresqlFlexibleServerAdministrator struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlFlexibleServerAdministrator)(nil)

// NewPostgresqlFlexibleServerAdministrator returns knowledge for the
// flexibleServers/administrators resource.
func NewPostgresqlFlexibleServerAdministrator() *PostgresqlFlexibleServerAdministrator {
	return &PostgresqlFlexibleServerAdministrator{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/flexibleServers/administrators",
			ApiVersions:  []string{"2025-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.principalName"},
				{PropertyPath: "properties.principalType"},
				{PropertyPath: "properties.tenantId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.principalType",
					AllowedValues: []string{"Group", "ServicePrincipal", "Unknown", "User"},
					Message:       "principal_type must be a valid PrincipalType",
				},
			},
			RequiredFields: []string{
				"properties.principalName",
				"properties.principalType",
				"properties.tenantId",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlFlexibleServerAdministrator()) }
