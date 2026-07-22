package datadog

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DatadogMonitorSsoConfiguration provides resource knowledge for
// Microsoft.Datadog/monitors/singleSignOnConfigurations.
//
// Mirrors azurerm_datadog_monitor_sso_configuration.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datadog/datadog_monitor_sso_configuration_resource.go
//     schema (lines 43-95): datadog_monitor_id parent reference (ForceNew, envelope);
//     name optional default "default" (resource name); enterprise_application_id required
//     (IsUUID) → enterpriseAppId; single_sign_on required StringInSlice(SingleSignOnStates)
//     → singleSignOnState; login_url computed → singleSignOnUrl; Create (100-142);
//     timeouts 30m/5m/30m/30m. (Pre-5.0 also exposes the deprecated single_sign_on_enabled
//     alias with an ExactlyOneOf against single_sign_on — both map to the same ARM
//     singleSignOnState, so no ARM-level relational rule.)
//   - go-azure-sdk resource-manager/datadog/2021-03-01/singlesignon:
//     DatadogSingleSignOnProperties (enterpriseAppId/singleSignOnState settable;
//     singleSignOnUrl/provisioningState read-only), constants
//     SingleSignOnStates{Disable,Enable,Existing,Initial};
//     id_singlesignonconfiguration.go type Microsoft.Datadog/monitors/singleSignOnConfigurations.
//
// No body-level ForceNew: the only ForceNew field is datadog_monitor_id (parent
// reference / envelope), not part of this resource's ARM body.
//
// TODO: enterprise_application_id uses validation.IsUUID — a generic semantic validator.
// It cannot be expressed as an azwise declarative rule; route it to
// typegraph.Validator(validators.UUID) on properties.enterpriseAppId in the azapin
// customizer (out of scope for this knowledge file).
type DatadogMonitorSsoConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DatadogMonitorSsoConfiguration)(nil)

// NewDatadogMonitorSsoConfiguration returns knowledge for the
// monitors/singleSignOnConfigurations resource.
func NewDatadogMonitorSsoConfiguration() *DatadogMonitorSsoConfiguration {
	return &DatadogMonitorSsoConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Datadog/monitors/singleSignOnConfigurations",
			ApiVersions:  []string{"2021-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// enterprise_application_id and single_sign_on are Required in the schema.
			RequiredFields: []string{
				"properties.enterpriseAppId",
				"properties.singleSignOnState",
			},
			StringRules: []azwise.StringRule{
				{
					// single_sign_on: StringInSlice(PossibleValuesForSingleSignOnStates).
					PropertyPath:  "properties.singleSignOnState",
					AllowedValues: []string{"Disable", "Enable", "Existing", "Initial"},
					Message:       "single_sign_on must be one of Disable, Enable, Existing, Initial",
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.singleSignOnUrl",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewDatadogMonitorSsoConfiguration()) }
