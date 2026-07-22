package applicationinsights

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationInsightsWebTest provides resource knowledge for
// Microsoft.Insights/webtests.
//
// This ARM type backs TWO Terraform resources, merged here (azwise keys by ARM
// resource type, so a single knowledge entry covers both):
//   - azurerm_application_insights_web_test          (kind: ping / multistep;
//     XML `configuration` body)
//   - azurerm_application_insights_standard_web_test (kind: standard; structured
//     `request` + `validation_rules` blocks)
//
// The ARM body uses PascalCase json tags (Frequency, Timeout, Enabled, Request,
// ValidationRules, …). kind is set at both the top-level envelope and
// properties.Kind. frequency/timeout and (for the standard kind) request/
// validation-rule defaults are unioned below.
//
// Kind-conditional shape (NOT encoded as unconditional RequiredFields, since each
// only applies to one kind and would wrongly gate the other):
//   - ping/multistep require properties.Configuration.WebTest (the XML config).
//   - standard requires properties.Request and properties.Request.RequestUrl.
//   The only cross-kind requirements are kind/properties.Kind and
//   properties.Locations (geo_locations).
//
// Deliberately not encoded here:
//   - frequency is validation.IntInSlice([300 600 900]) — a discrete allow-set
//     that IntRule (Min/Max only) cannot express; see the TODO.
//   - geo_locations is Required with MinItems:1; ArrayRule only carries MaxItems,
//     so the minimum-length constraint is not expressible.
//   - application_insights_id (components.ValidateComponentID) is the parent
//     reference on the envelope, not a body property.
//   - synthetic_monitor_id is Computed in AzureRM but the ARM create model sets
//     properties.SyntheticMonitorId (required, = web test name), so it is NOT a
//     strip-able ComputedField.
//   - the standard resource's CustomizeDiff (ssl_check_enabled / ssl_cert_
//     remaining_lifetime require an https request url) is a cross-field/array
//     semantic check that no declarative rule type expresses.
//
// Sources:
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_web_test_resource.go:31-209
//     (ping/multistep schema: kind ForceNew StringInSlice[multistep,ping],
//     frequency/timeout defaults, geo_locations Required, configuration Required XML,
//     timeouts 30m/5m/30m/30m; Create mapping to WebTestProperties)
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_standard_web_test_resource.go:37-389
//     (standard schema: request/validation_rules blocks, http_verb enum, ssl cert
//     lifetime IntBetween(1,365), defaults; kind hardcoded "standard")
//   - go-azure-sdk resource-manager/applicationinsights/2022-06-15/webtestsapis
//     model_webtestproperties.go / model_webtestpropertiesrequest.go /
//     model_webtestpropertiesvalidationrules.go /
//     model_webtestpropertiesvalidationrulescontentvalidation.go (ARM json tags) and
//     constants.go (WebTestKind: multistep/ping/standard)
type ApplicationInsightsWebTest struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationInsightsWebTest)(nil)

// NewApplicationInsightsWebTest returns knowledge for the webtests resource,
// covering both the classic (ping/multistep) and standard web tests.
func NewApplicationInsightsWebTest() *ApplicationInsightsWebTest {
	return &ApplicationInsightsWebTest{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/webtests",
			ApiVersions:  []string{"2022-06-15"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// kind is ForceNew (classic web_test); the standard resource hardcodes it
			// to "standard". Set at both the envelope and properties.Kind.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "kind"},
				{PropertyPath: "properties.Kind"},
			},
			// TODO: frequency is validation.IntInSlice([300 600 900]) — a discrete
			// allow-set IntRule (Min/Max bounds) cannot represent; no rule emitted for
			// properties.Frequency.
			IntRules: []azwise.IntRule{
				// standard: ssl_cert_remaining_lifetime validation.IntBetween(1,365).
				{PropertyPath: "properties.ValidationRules.SSLCertRemainingLifetimeCheck", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(365))},
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "kind", AllowedValues: []string{"ping", "multistep", "standard"}},
				{PropertyPath: "properties.Kind", AllowedValues: []string{"ping", "multistep", "standard"}},
				// standard: request.http_verb enum.
				{PropertyPath: "properties.Request.HttpVerb", AllowedValues: []string{
					"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
				}},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.Frequency", Value: int64(300)},
				{PropertyPath: "properties.Timeout", Value: int64(30)},
				// standard request block defaults.
				{PropertyPath: "properties.Request.FollowRedirects", Value: true},
				{PropertyPath: "properties.Request.HttpVerb", Value: "GET"},
				{PropertyPath: "properties.Request.ParseDependentRequests", Value: true},
				// standard validation_rules block defaults.
				{PropertyPath: "properties.ValidationRules.ExpectedHttpStatusCode", Value: int64(200)},
				{PropertyPath: "properties.ValidationRules.SSLCheck", Value: false},
				{PropertyPath: "properties.ValidationRules.ContentValidation.IgnoreCase", Value: false},
				{PropertyPath: "properties.ValidationRules.ContentValidation.PassIfTextFound", Value: false},
			},
			// Cross-kind requirements only (see doc comment for kind-conditional ones).
			RequiredFields: []string{
				"kind",
				"properties.Kind",
				"properties.Locations",
			},
		},
	}
}

func init() { azwise.Register(NewApplicationInsightsWebTest()) }
