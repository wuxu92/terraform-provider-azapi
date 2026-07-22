package applicationinsights

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationInsightsSmartDetectionRule provides resource knowledge for
// Microsoft.Insights/components/proactiveDetectionConfigs.
//
// Mirrors azurerm_application_insights_smart_detection_rule. The proactive
// detection configuration body is flat (no "properties" envelope): the update
// PUTs ApplicationInsightsComponentProactiveDetectionConfiguration directly, so
// enabled / sendEmailsToSubscriptionOwners / customEmails are top-level paths.
//
// Deliberately not encoded here:
//   - name is a StringInSlice over UI rule names that AzureRM converts to the API
//     configuration id via convertUiNameToApiName (+ a DiffSuppressFunc). It
//     targets the envelope resource name and undergoes a non-trivial UI->API
//     transform, so it is not expressible as a declarative StringRule; AzAPI users
//     supply the raw ARM configuration id.
//   - application_insights_id (components.ValidateComponentID) is the parent
//     reference on the envelope, not a body property.
//
// Sources:
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_smart_detection_rule_resource.go:24-128
//     (schema: name/application_insights_id ForceNew, enabled + send_emails
//     defaults true, timeouts 30m/5m/30m/30m; Update body mapping)
//   - go-azure-sdk resource-manager/applicationinsights/2015-05-01/componentproactivedetectionapis
//     model_applicationinsightscomponentproactivedetectionconfiguration.go
//     (flat json tags: enabled, sendEmailsToSubscriptionOwners, customEmails) and
//     id_proactivedetectionconfig.go (ARM type + 2015-05-01 API)
type ApplicationInsightsSmartDetectionRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationInsightsSmartDetectionRule)(nil)

// NewApplicationInsightsSmartDetectionRule returns knowledge for the
// proactiveDetectionConfigs resource.
func NewApplicationInsightsSmartDetectionRule() *ApplicationInsightsSmartDetectionRule {
	return &ApplicationInsightsSmartDetectionRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/components/proactiveDetectionConfigs",
			ApiVersions:  []string{"2015-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "enabled", Value: true},
				{PropertyPath: "sendEmailsToSubscriptionOwners", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewApplicationInsightsSmartDetectionRule()) }
