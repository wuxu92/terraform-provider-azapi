package connections

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiConnection provides resource knowledge for Microsoft.Web/connections.
//
// Contributing Terraform resource: azurerm_api_connection.
//
// NOTE: the `managedapis` SDK type (web/2016-06-01/managedapis) backs only the
// azurerm_managed_api DATA SOURCE (registration.go SupportedDataSources ->
// dataSourceManagedApi); it has no managed TF resource. The sole managed resource
// in this service is azurerm_api_connection, so only Microsoft.Web/connections is
// emitted here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/connections/api_connection_resource.go
//     (schema L45-82, Create body L113-126, Update L207-221, Read L156-178)
//   - terraform-provider-azurerm internal/services/connections/registration.go L52-62
//     (managed_api = data source only; api_connection = only managed resource)
//   - go-azure-sdk resource-manager/web/2016-06-01/connections:
//     model_apiconnectiondefinition.go, model_apiconnectiondefinitionproperties.go,
//     model_apireference.go
//
// Notes:
//   - managed_api_id (Required, ForceNew) maps to properties.api.id and also determines
//     the resource location (AzureRM normalizes managedApiId.LocationName). Its
//     ValidateFunc managedapis.ValidateManagedApiID is a resource-ID (semantic)
//     validator -> belongs in an azapin customizer (AzureResourceID), not a declarative
//     StringRule.
//   - display_name is Optional+Computed: Azure derives a per-managed-API default when it
//     is omitted (e.g. servicebus -> "Service Bus"), so it is a server default, not a
//     fixed value -> DefaultValue with nil Value.
//   - parameter_values -> properties.parameterValues (write direction). The GET response
//     echoes non-secret values in properties.nonSecretParameterValues instead, which is
//     therefore read-only.
type ApiConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiConnection)(nil)

// NewApiConnection returns knowledge for the Microsoft.Web/connections resource.
func NewApiConnection() *ApiConnection {
	return &ApiConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/connections",
			ApiVersions:  []string{"2016-06-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// name (resource name) and managed_api_id are ForceNew. managed_api_id maps
			// to properties.api.id.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.api.id"},
			},
			// managed_api_id (Required) -> properties.api.id.
			RequiredFields: []string{
				"properties.api.id",
			},
			StringRules: []azwise.StringRule{
				// name -> resource name attribute -- validation.StringIsNotEmpty
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "must not be empty",
				},
				// display_name -> properties.displayName -- validation.StringIsNotEmpty
				{
					PropertyPath: "properties.displayName",
					MinLength:    1,
					Message:      "must not be empty",
				},
			},
			// display_name is O+C: Azure fills a per-managed-API default when omitted.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.displayName"},
			},
			// Read-only response fields absent from the create/update write path.
			ComputedFields: []string{
				"properties.createdTime",
				"properties.changedTime",
				"properties.statuses",
				"properties.testLinks",
				"properties.nonSecretParameterValues",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewApiConnection()) }
