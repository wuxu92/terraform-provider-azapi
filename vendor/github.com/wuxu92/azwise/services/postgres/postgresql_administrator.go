package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlAdministrator provides resource knowledge for
// Microsoft.DBforPostgreSQL/servers/administrators
// (azurerm_postgresql_active_directory_administrator).
//
// Deprecated Single Server sub-resource; DISTINCT from
// Microsoft.DBforPostgreSQL/flexibleServers/administrators. It is a singleton whose
// ARM name is the literal "activeDirectory" (PUT .../servers/{name}/administrators/activeDirectory).
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_active_directory_administrator_resource.go
//     :24-76   (schema: server_name ForceNew, login AdminUsernames Required, object_id
//     IsUUID Required, tenant_id IsUUID Required)
//     :35-40   (timeouts: create/update/delete 30m, read 5m)
//     :102-109 (create mapping; administratorType hardcoded ActiveDirectory, login,
//     sid<-object_id, tenantId)
//   - go-azure-sdk resource-manager/postgresql/2017-12-01/serveradministrators:
//     model_serveradministratorproperties.go (properties.administratorType/login/sid/tenantId),
//     constants.go:14-16 (AdministratorType), method_createorupdate.go:33
//     (.../administrators/activeDirectory)
//
// Not encoded (deliberate):
//   - object_id (properties.sid) and tenant_id use validation.IsUUID; login uses
//     AdminUsernames (a disallowed-name blocklist). These are semantic validators that
//     belong in an azapin customizer, not declarative StringRules.
type PostgresqlAdministrator struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlAdministrator)(nil)

// NewPostgresqlAdministrator returns knowledge for the servers/administrators resource.
func NewPostgresqlAdministrator() *PostgresqlAdministrator {
	return &PostgresqlAdministrator{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/servers/administrators",
			ApiVersions:  []string{"2017-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.administratorType",
					AllowedValues: []string{"ActiveDirectory"},
					Message:       "administratorType must be ActiveDirectory",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.administratorType", Value: "ActiveDirectory"},
			},
			RequiredFields: []string{
				"properties.administratorType",
				"properties.login",
				"properties.sid",
				"properties.tenantId",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlAdministrator()) }
