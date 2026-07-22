package mysql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MySQLFlexibleServerActiveDirectoryAdministrator provides resource knowledge for
// Microsoft.DBforMySQL/flexibleServers/administrators.
//
// Contributing Terraform resource: azurerm_mysql_flexible_server_active_directory_administrator.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mysql/mysql_flexible_server_active_directory_administrator_resource.go
//     (Arguments L46-79, Create body L111-119, Update L162-176, timeouts L87/L136/L189/L234)
//   - terraform-provider-azurerm internal/services/mysql/parse/flexible_server_azure_active_directory_administrator.go
//     (ID format L42-45 -> administrators/<administratorType>)
//   - go-azure-sdk resource-manager/mysql/2023-12-30/azureadadministrators:
//     model_azureadadministrator.go, model_administratorproperties.go, constants.go
//
// Notes:
//   - server_id is a parent-reference / envelope field; not emitted as a body rule.
//   - administratorType is hardcoded by AzureRM to "ActiveDirectory" (the SDK enum has only this value);
//     emitted as a required field + default.
//   - object_id (properties.sid) and tenant_id (properties.tenantId) are validated as UUIDs by AzureRM
//     (validation.IsUUID); the UUID semantic check is not expressed declaratively here.
type MySQLFlexibleServerActiveDirectoryAdministrator struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MySQLFlexibleServerActiveDirectoryAdministrator)(nil)

// NewMySQLFlexibleServerActiveDirectoryAdministrator returns knowledge for the
// flexibleServers/administrators resource.
func NewMySQLFlexibleServerActiveDirectoryAdministrator() *MySQLFlexibleServerActiveDirectoryAdministrator {
	return &MySQLFlexibleServerActiveDirectoryAdministrator{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforMySQL/flexibleServers/administrators",
			ApiVersions:  []string{"2023-12-30"},
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
				"properties.identityResourceId",
			},
		},
	}
}

func init() { azwise.Register(NewMySQLFlexibleServerActiveDirectoryAdministrator()) }
