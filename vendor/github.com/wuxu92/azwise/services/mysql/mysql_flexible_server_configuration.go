package mysql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MySQLFlexibleServerConfiguration provides resource knowledge for
// Microsoft.DBforMySQL/flexibleServers/configurations.
//
// Contributing Terraform resource: azurerm_mysql_flexible_server_configuration.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mysql/mysql_flexible_server_configuration_resource.go
//     (schema L48-68, Create/Update body L78-121, timeouts L37-42)
//   - go-azure-sdk resource-manager/mysql/2023-12-30/configurations:
//     model_configuration.go, model_configurationproperties.go, id_configuration.go
//
// Notes:
//   - name/server_name/resource_group_name are envelope / parent-reference fields; not emitted as body rules.
//   - value (properties.value) is the only user-settable body property; it is Required.
//   - allowedValues/currentValue/dataType/defaultValue/description/documentationLink/
//     isConfigPendingRestart/isDynamicConfig/isReadOnly/source are server-computed metadata.
type MySQLFlexibleServerConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MySQLFlexibleServerConfiguration)(nil)

// NewMySQLFlexibleServerConfiguration returns knowledge for the flexibleServers/configurations resource.
func NewMySQLFlexibleServerConfiguration() *MySQLFlexibleServerConfiguration {
	return &MySQLFlexibleServerConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforMySQL/flexibleServers/configurations",
			ApiVersions:  []string{"2023-12-30"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.value",
			},
			ComputedFields: []string{
				"properties.allowedValues",
				"properties.currentValue",
				"properties.dataType",
				"properties.defaultValue",
				"properties.description",
				"properties.documentationLink",
				"properties.isConfigPendingRestart",
				"properties.isDynamicConfig",
				"properties.isReadOnly",
				"properties.source",
			},
		},
	}
}

func init() { azwise.Register(NewMySQLFlexibleServerConfiguration()) }
