package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationConnectionType provides resource knowledge for
// Microsoft.Automation/automationAccounts/connectionTypes.
//
// Contributing Terraform resource: azurerm_automation_connection_type.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_connection_type_resource.go
//     (Arguments L40-90, Create L104-159 timeout 30m, Read L161 timeout 5m, Delete L204 timeout 10m)
//   - go-azure-sdk resource-manager/automation/2024-10-23/connectiontype:
//     model_connectiontypecreateorupdateproperties.go (Create body: fieldDefinitions, isGlobal)
//     model_connectiontypeproperties.go (GET-only: creationTime, description, lastModifiedTime).
//   - validate/connection_type_name.go ConnectionTypeName regex.
//
// Notes:
//   - The whole resource is immutable: name, automation_account_name, is_global and the field
//     list are all ForceNew (there is no Update). properties.fieldDefinitions is a
//     map[string]FieldDefinition; its element fields (name/type/is_encrypted/is_optional) are
//     map values with no fixed ARM schema path, so their per-field requireds are not expressed.
type AutomationConnectionType struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationConnectionType)(nil)

// NewAutomationConnectionType returns knowledge for the automationAccounts/connectionTypes resource.
func NewAutomationConnectionType() *AutomationConnectionType {
	return &AutomationConnectionType{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/connectionTypes",
			ApiVersions:  []string{"2024-10-23"},
			// name & automation_account_name are envelope-owned (Required+ForceNew).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.isGlobal"},
				{PropertyPath: "properties.fieldDefinitions"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 10 * time.Minute,
			},
			// field block is Required → properties.fieldDefinitions.
			RequiredFields: []string{
				"properties.fieldDefinitions",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ConnectionTypeName
				{
					Regex:     `^[\w\-]{1,128}$`,
					MinLength: 1,
					MaxLength: 128,
					Message:   "must contain only letters, numbers, hyphens and underscores, 1-128 characters",
				},
			},
			// GET-only fields, absent from the Create/Update model.
			ComputedFields: []string{
				"properties.creationTime",
				"properties.description",
				"properties.lastModifiedTime",
			},
		},
	}
}

func init() { azwise.Register(NewAutomationConnectionType()) }
