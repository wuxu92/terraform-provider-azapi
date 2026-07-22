package databricks

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AccessConnector provides resource knowledge for Microsoft.Databricks/accessConnectors.
//
// Contributing TF resource: azurerm_databricks_access_connector.
//
// Sources:
//   - AzureRM internal/services/databricks/databricks_access_connector_resource.go
//     (schema :43-60, Create :78-125, Read timeout :171-173, Delete timeout :220-222)
//   - AzureRM internal/services/databricks/validate/access_connector_name.go (name rule)
//   - go-azure-sdk .../databricks/2026-01-01/accessconnector: model_accessconnector.go,
//     model_accessconnectorproperties.go (ARM body paths)
type AccessConnector struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AccessConnector)(nil)

func NewAccessConnector() *AccessConnector {
	return &AccessConnector{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Databricks/accessConnectors",
			ApiVersions:  []string{"2026-01-01"},
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
					// Resource name (empty PropertyPath). validate.AccessConnectorName:
					// 1-64 chars, alphanumeric plus underscore and hyphen.
					Regex:     `^[a-zA-Z0-9_-]*$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters and contain only alphanumeric characters, underscores, and hyphens",
				},
			},
			// ProvisioningState is populated by Azure on GET and is absent from the
			// create/update body (AccessConnectorProperties, referedBy is service-managed).
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewAccessConnector()) }
