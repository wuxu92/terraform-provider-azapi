package eventhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventHubNamespace provides resource knowledge for Microsoft.EventHub/namespaces.
//
// Contributing Terraform resource: azurerm_eventhub_namespace.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventhub/eventhub_namespace_resource.go
//     (schema L44-280, Create body L282-385, sku Premium-migration CustomizeDiff L252-264)
//   - terraform-provider-azurerm internal/services/eventhub/validate/eventhub_names.go
//     (ValidateEventHubNamespaceName L14-19)
//   - go-azure-sdk resource-manager/eventhub/2024-01-01/namespaces:
//     model_ehnamespace.go, model_ehnamespaceproperties.go, model_sku.go, constants.go
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; not emitted as body rules.
//   - sku change is ForceNew only when migrating to/from Premium (CustomizeDiff); handled
//     in CheckForceNew below, not as an unconditional declarative rule.
//   - minimum_tls_version uses the full TlsVersion SDK enum (1.0/1.1/1.2); AzureRM 5.0
//     restricts the settable value to 1.2.
//   - network_rulesets is a separate ARM API (Microsoft.EventHub/namespaces/networkRuleSets)
//     and customer_managed_key drives properties.encryption via the distinct
//     azurerm_eventhub_namespace_customer_managed_key resource; neither is emitted here.
//   - default_primary_connection_string / *_key etc. are data-plane keys read from the
//     namespace's default authorization rule, not namespace body properties; omitted.
type EventHubNamespace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventHubNamespace)(nil)

// NewEventHubNamespace returns knowledge for the namespaces resource.
func NewEventHubNamespace() *EventHubNamespace {
	return &EventHubNamespace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventHub/namespaces",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.clusterArmId"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z][-a-zA-Z0-9]{4,48}[a-zA-Z0-9]$`,
					Message:      "namespace name may contain only letters, numbers and hyphens, must start with a letter, end with a letter or number and be 6-50 characters long",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Premium", "Standard"},
					Message:       "sku must be Basic, Standard or Premium",
				},
				{
					PropertyPath:  "properties.minimumTlsVersion",
					AllowedValues: []string{"1.0", "1.1", "1.2"},
					Message:       "minimum_tls_version must be 1.0, 1.1 or 1.2",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.maximumThroughputUnits", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(40))},
			},
			ComputedFields: []string{
				"properties.metricId",
				"properties.serviceBusEndpoint",
				"properties.provisioningState",
				"properties.status",
				"properties.createdAt",
				"properties.updatedAt",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.capacity", Value: 1},
				{PropertyPath: "properties.isAutoInflateEnabled", Value: false},
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				{PropertyPath: "properties.minimumTlsVersion", Value: "1.2"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			RequiredFields: []string{
				"sku.name",
			},
		},
	}
}

// CheckForceNew extends the declarative check with the namespace SKU rule: migrating a
// namespace to or from the Premium SKU forces replacement (mirrors the AzureRM
// CustomizeDiff in eventhub_namespace_resource.go L252-264).
func (s *EventHubNamespace) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}
	oldSku := azwise.ExtractStringValue(oldBody, "sku.name")
	newSku := azwise.ExtractStringValue(newBody, "sku.name")
	if oldSku != "" && newSku != "" && oldSku != newSku {
		if newSku == "Premium" || oldSku == "Premium" {
			return true
		}
	}
	return false
}

func init() { azwise.Register(NewEventHubNamespace()) }
