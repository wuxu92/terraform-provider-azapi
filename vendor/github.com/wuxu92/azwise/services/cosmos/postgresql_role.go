package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgreSQLRole provides resource knowledge for
// Microsoft.DBforPostgreSQL/serverGroupsv2/roles (azurerm_cosmosdb_postgresql_role).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_postgresql_role_resource.go
//     :40-64   (schema: name RoleName ForceNew, cluster_id ForceNew, password
//     RolePassword Required+ForceNew+Sensitive)
//     :70-112  (create mapping to roles.RoleProperties: properties.password)
//     :72,116,152 (timeouts: create 30m, read 5m, delete 30m; no update)
//   - AzureRM internal/services/cosmos/validate/role_name.go:11-14 (name regex),
//     validate/role_password.go:10-19 (password 8..256 length)
//   - go-azure-sdk resource-manager/postgresqlhsc/2022-11-08/roles:
//     model_roleproperties.go (properties.password body path)
type PostgreSQLRole struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgreSQLRole)(nil)

// NewPostgreSQLRole returns knowledge for the serverGroupsv2/roles resource.
func NewPostgreSQLRole() *PostgreSQLRole {
	return &PostgreSQLRole{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/serverGroupsv2/roles",
			ApiVersions:  []string{"2022-11-08"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.password"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-z0-9]{1,63}$`,
					MinLength:    1,
					MaxLength:    63,
					Message:      "name must be 1-63 characters of lower case letters and numbers",
				},
				{
					PropertyPath: "properties.password",
					MinLength:    8,
					MaxLength:    256,
					Message:      "password must be between 8 and 256 characters",
				},
			},
			SensitiveFields: []string{"properties.password"},
			RequiredFields:  []string{"properties.password"},
		},
	}
}

func init() { azwise.Register(NewPostgreSQLRole()) }
