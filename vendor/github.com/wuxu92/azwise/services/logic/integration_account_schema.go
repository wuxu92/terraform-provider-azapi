package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccountSchema provides resource knowledge for
// Microsoft.Logic/integrationAccounts/schemas.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account_schema.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_schema_resource.go
//     (schema L42-77, Create body L103-113)
//   - terraform-provider-azurerm internal/services/logic/validate/
//     integration_account_schema_name.go, integration_account_schema_file_name.go
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccountschemas:
//     model_integrationaccountschema.go, model_integrationaccountschemaproperties.go,
//     constants.go (PossibleValuesForSchemaType), id_schema.go
//
// Notes:
//   - name (ForceNew), integration_account_name (ForceNew parent segment) and
//     resource_group_name are envelope-owned; only the schema-name regex is emitted.
//   - schema_type is hardcoded by AzureRM to Xml → properties.schemaType; the full SchemaType
//     SDK enum is emitted for the value constraint. content → properties.content;
//     contentType is hardcoded to application/xml (provider-controlled).
//   - file_name → properties.fileName, validated by AzureRM only as a `.xsd` suffix check;
//     that suffix rule is expressed as a regex.
type IntegrationAccountSchema struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccountSchema)(nil)

// NewIntegrationAccountSchema returns knowledge for the schemas child resource.
func NewIntegrationAccountSchema() *IntegrationAccountSchema {
	return &IntegrationAccountSchema{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts/schemas",
			ApiVersions:  []string{"2019-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9-()._]+$`,
					MaxLength:    80,
					Message:      "schema name contains only letters, numbers, dots, parentheses, hyphens and underscores, up to 80 characters",
				},
				{
					PropertyPath:  "properties.schemaType",
					AllowedValues: []string{"NotSpecified", "Xml"},
					Message:       "schema_type must be a valid SchemaType",
				},
				{
					PropertyPath: "properties.fileName",
					Regex:        `\.xsd$`,
					Message:      "file_name must end with `.xsd`",
				},
			},
			RequiredFields: []string{
				"properties.schemaType",
				"properties.content",
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccountSchema()) }
