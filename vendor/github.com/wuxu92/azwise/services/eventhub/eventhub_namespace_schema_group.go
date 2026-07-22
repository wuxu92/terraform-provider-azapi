package eventhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventHubNamespaceSchemaGroup provides resource knowledge for
// Microsoft.EventHub/namespaces/schemagroups.
//
// Contributing Terraform resource: azurerm_eventhub_namespace_schema_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventhub/eventhub_namespace_schema_group_resource.go
//     (schema L21-76, Create body L104-112)
//   - terraform-provider-azurerm internal/services/eventhub/validate/eventhub_names.go
//     (ValidateSchemaGroupName L42-47)
//   - go-azure-sdk resource-manager/eventhub/2024-01-01/schemaregistry:
//     model_schemagroup.go, model_schemagroupproperties.go, constants.go
//
// Notes:
//   - name/namespace_id are envelope / parent-reference fields; not emitted as body rules.
//   - schema_compatibility and schema_type are ForceNew body properties.
//   - schema_type allows "Json" in AzureRM in addition to the SDK enum (Avro/Unknown);
//     the API accepts it, so it is included in AllowedValues even though the SDK constants
//     list only Avro and Unknown.
//   - create-only resource (no Update); Update timeout left at the provider default.
type EventHubNamespaceSchemaGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventHubNamespaceSchemaGroup)(nil)

// NewEventHubNamespaceSchemaGroup returns knowledge for the schemagroups resource.
func NewEventHubNamespaceSchemaGroup() *EventHubNamespaceSchemaGroup {
	return &EventHubNamespaceSchemaGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventHub/namespaces/schemagroups",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.schemaCompatibility"},
				{PropertyPath: "properties.schemaType"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,254}[a-zA-Z0-9])?$`,
					Message:      "schema group name may contain only letters, numbers, periods, hyphens and underscores, up to 256 characters, beginning and ending with a letter or number",
				},
				{
					PropertyPath:  "properties.schemaCompatibility",
					AllowedValues: []string{"None", "Backward", "Forward"},
					Message:       "schema_compatibility must be None, Backward or Forward",
				},
				{
					PropertyPath:  "properties.schemaType",
					AllowedValues: []string{"Unknown", "Avro", "Json"},
					Message:       "schema_type must be Unknown, Avro or Json",
				},
			},
			RequiredFields: []string{
				"properties.schemaCompatibility",
				"properties.schemaType",
			},
		},
	}
}

func init() { azwise.Register(NewEventHubNamespaceSchemaGroup()) }
