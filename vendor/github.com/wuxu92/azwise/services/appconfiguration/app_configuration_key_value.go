package appconfiguration

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AppConfigurationKeyValue provides resource knowledge for
// Microsoft.AppConfiguration/configurationStores/keyValues.
//
// This ARM type is the merge target for two AzureRM resources that both write
// App Configuration key/value entries (they share the same ARM resource type;
// AzureRM implements them over the data-plane API, but AzAPI addresses the ARM
// control-plane keyValues resource):
//   - azurerm_app_configuration_key     (a plain kv or Key Vault reference)
//   - azurerm_app_configuration_feature  (a feature flag)
//
// The ARM keyValues body is intentionally thin: properties.value (string),
// properties.contentType (string), and tags. The key and label are encoded in the
// resource name ("<key>$<label>") rather than the body.
//
// Almost all of the AzureRM schema validation cannot be lowered to declarative
// ARM-path rules because it constrains fields that are serialized *into* the single
// properties.value JSON string rather than distinct ARM properties:
//   - key `type` (kv/vault), `vault_key_reference`, and value/vault_key_reference
//     ConflictsWith select the contentType and encode the value payload.
//   - feature `enabled`, `description`, `percentage_filter_value` (FloatBetween 0-100),
//     `targeting_filter`/`timewindow_filter`/`custom_filter` (IntBetween 0-100,
//     RFC3339 times, etc.) are all marshalled into the feature-flag JSON value.
//   - `locked` is a data-plane lock operation, read-only on the ARM control plane.
//
// TODO: none of the above are expressible as azwise StringRules/IntRules/FloatRules
// against ARM property paths — they validate the *contents* of properties.value.
// Only the envelope (name = key$label, parent = configuration store) and the
// coarse body shape are captured here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appconfiguration/app_configuration_key_resource.go
//     (schema 61-127, Create 141-252; data-plane KeyValue{Key,Label,Value,ContentType,Tags}; 45m timeout)
//   - terraform-provider-azurerm internal/services/appconfiguration/app_configuration_feature_resource.go
//     (schema 61-194, Create 208-366; feature flag encoded into value; 45m timeout)
//   - ARM Microsoft.AppConfiguration/configurationStores/keyValues (2024-05-01):
//     properties.value / properties.contentType / tags.
type AppConfigurationKeyValue struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AppConfigurationKeyValue)(nil)

// NewAppConfigurationKeyValue returns knowledge for the keyValues resource.
func NewAppConfigurationKeyValue() *AppConfigurationKeyValue {
	return &AppConfigurationKeyValue{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppConfiguration/configurationStores/keyValues",
			ApiVersions:  []string{"2024-05-01"},
			// Both AzureRM resources use a 45-minute operation timeout (data-plane
			// propagation waits). key/label are ForceNew but live on the resource
			// name, so no body ForceNew rule applies.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 45 * time.Minute,
				Read:   5 * time.Minute,
				Update: 45 * time.Minute,
				Delete: 45 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewAppConfigurationKeyValue()) }
