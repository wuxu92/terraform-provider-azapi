package advisor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AdvisorSuppression provides resource knowledge for
// Microsoft.Advisor/suppressions.
//
// Mirrors azurerm_advisor_suppression. The suppression is a scoped extension
// resource: its ARM name is the suppression name and its parent is the
// recommendation (which itself is scoped to a target resource). Name,
// recommendation, and target resource all live on the operational envelope
// (name + parent), which is Required + RequiresReplace by construction, so
// AzureRM's ForceNew on name/recommendation_id/resource_id is not repeated here.
//
// ttl is Optional + ForceNew and maps to properties.ttl. It is validated by
// AzureRM's validate.Duration (a "d.hh:mm:ss"-style duration string). That check is
// a resource-specific format validator with no equivalent declarative azwise rule,
// so it is documented but not lowered to a StringRule.
//
// suppression_id is Computed (server-assigned) and read from properties.suppressionId.
//
// Sources:
//   - terraform-provider-azurerm internal/services/advisor/advisor_suppression_resource.go
//     (schema 32-66, Create 76-119; ttl ForceNew; suppression_id computed; 30m create timeout)
//   - go-azure-sdk resource-manager/advisor/2023-01-01/suppressions
//     SuppressionProperties{suppressionId, ttl, expirationTimeStamp}.
type AdvisorSuppression struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AdvisorSuppression)(nil)

// NewAdvisorSuppression returns knowledge for the suppressions resource.
func NewAdvisorSuppression() *AdvisorSuppression {
	return &AdvisorSuppression{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Advisor/suppressions",
			ApiVersions:  []string{"2023-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// ttl is unconditionally ForceNew in AzureRM.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.ttl"},
			},
			// suppressionId and expirationTimeStamp are server-assigned read-only
			// properties returned by GET.
			ComputedFields: []string{
				"properties.suppressionId",
				"properties.expirationTimeStamp",
			},
		},
	}
}

func init() { azwise.Register(NewAdvisorSuppression()) }
