package iotcentral

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotCentralApplication provides resource knowledge for
// Microsoft.IoTCentral/iotApps (azurerm_iotcentral_application).
//
// Sources:
//   - AzureRM internal/services/iotcentral/iotcentral_application_resource.go
//     :46-51   (CRUD timeouts: create/update/delete 30m, read 5m)
//     :53-105  (schema: name ForceNew, sub_domain Required, display_name Optional+Computed,
//     public_network_access_enabled bool default true, sku default ST1, template ForceNew
//     default "iotc-pnp-preview@1.0.0", identity/location/resource_group_name/tags envelope)
//     :159-172 (create payload → App{Properties{DisplayName, PublicNetworkAccess, Subdomain,
//     Template}, Sku{Name}, Identity, Location, Tags})
//   - AzureRM internal/services/iotcentral/validate/application_name.go       :12-18 (name regex, len 2-63)
//   - AzureRM internal/services/iotcentral/validate/application_subdomain.go  :12-18 (subdomain regex, len 2-63)
//   - AzureRM internal/services/iotcentral/validate/application_display_name.go :12-16 (display_name len 1-200)
//   - AzureRM internal/services/iotcentral/validate/application_template_name.go :12-16 (template len 1-50)
//   - go-azure-sdk resource-manager/iotcentral/2021-11-01-preview/apps:
//     model_app.go, model_appproperties.go (ARM body paths), model_appskuinfo.go (sku.name),
//     constants.go (AppSku ST0/ST1/ST2, PublicNetworkAccess Enabled/Disabled)
//
// Not encoded (deliberate):
//   - azurerm_iotcentral_organization is a DATA-PLANE resource: its AzureRM implementation
//     uses the kermit IoT Central data-plane SDK
//     (jackofallops/kermit/sdk/iotcentral/2022-10-31-preview/iotcentral), not an ARM
//     resource-manager API. It has no distinct ARM resource type and no ARM body, so no
//     knowledge file is emitted for it.
//   - azurerm_iotcentral_application_network_rule_set is NOT a distinct ARM resource type.
//     It mutates the SAME Microsoft.IoTCentral/iotApps resource by setting
//     properties.networkRuleSets via the apps CreateOrUpdate call (see
//     iotcentral_application_network_rule_set_resource.go:141-149). Its IP-rule constraints
//     live at array-element paths (properties.networkRuleSets.ipRules[*].ipMask) and its
//     default_action/apply_to_device settings are managed as a sub-config of the app, so no
//     rules are emitted here (array-element / sub-config paths are out of scope).
//   - identity is a resource envelope block (SystemAssigned), not a properties.* body field.
//   - properties.applicationId, properties.provisioningState, properties.state and
//     properties.privateEndpointConnections are populated by Azure in the GET response.
type IotCentralApplication struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotCentralApplication)(nil)

// NewIotCentralApplication returns knowledge for the iotApps resource.
func NewIotCentralApplication() *IotCentralApplication {
	return &IotCentralApplication{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.IoTCentral/iotApps",
			ApiVersions:  []string{"2021-11-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// name attribute (empty PropertyPath = resource name).
				{PropertyPath: ""},
				{PropertyPath: "properties.template"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath = name attribute).
					Regex:     `^[a-z\d][a-z\d-]{0,61}[a-z\d]$`,
					MinLength: 2,
					MaxLength: 63,
					Message:   "must be 2-63 chars, lowercase alphanumeric or dashes, and begin/end with an alphanumeric character",
				},
				{
					PropertyPath: "properties.subdomain",
					Regex:        `^[a-z\d][a-z\d-]{0,61}[a-z\d]$`,
					MinLength:    2,
					MaxLength:    63,
					Message:      "must be 2-63 chars, lowercase alphanumeric or dashes, and begin/end with an alphanumeric character",
				},
				{
					PropertyPath: "properties.displayName",
					MinLength:    1,
					MaxLength:    200,
					Message:      "must be 1-200 characters",
				},
				{
					PropertyPath: "properties.template",
					MinLength:    1,
					MaxLength:    50,
					Message:      "must be 1-50 characters",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"ST0", "ST1", "ST2"},
					Message:       "must be one of ST0, ST1, ST2",
				},
			},
			// ARM requires the sku envelope (non-pointer in the SDK model) and AzureRM
			// requires sub_domain; both must be present on create.
			RequiredFields: []string{
				"properties.subdomain",
				"sku.name",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.name", Value: "ST1"},
				{PropertyPath: "properties.template", Value: "iotc-pnp-preview@1.0.0"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			// Read-only ARM properties populated by Azure in the GET response.
			ComputedFields: []string{
				"properties.applicationId",
				"properties.provisioningState",
				"properties.state",
				"properties.privateEndpointConnections",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewIotCentralApplication()) }
