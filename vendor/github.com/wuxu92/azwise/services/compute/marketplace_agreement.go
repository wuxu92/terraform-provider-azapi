package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MarketplaceAgreement provides resource knowledge for
// Microsoft.MarketplaceOrdering/agreements/offers/plans.
//
// Contributing Terraform resource: azurerm_marketplace_agreement.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/marketplace_agreement_resource.go
//     (schema L21-70, Create body L72-128, Read L130-167)
//   - go-azure-sdk resource-manager/marketplaceordering/2015-06-01/agreements:
//     model_agreementterms.go, model_agreementproperties.go, id_plan.go (ID format
//     /providers/Microsoft.MarketplaceOrdering/agreements/{publisher}/offers/{offer}/plans/{plan}).
//
// Notes:
//   - publisher/offer/plan are ID (envelope) segments, each ForceNew but not ARM-body
//     properties; StringIsNotEmpty on them is a non-empty check on envelope fields, so no
//     body ForceNew/StringRule is emitted.
//   - The provider retrieves the existing agreement terms and PUTs them back with
//     properties.accepted = true; accepted is the only user-authored body field.
//   - This resource has no Update; the API accepts (create) or cancels (delete) the terms.
type MarketplaceAgreement struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MarketplaceAgreement)(nil)

// NewMarketplaceAgreement returns knowledge for the marketplace agreements resource.
func NewMarketplaceAgreement() *MarketplaceAgreement {
	return &MarketplaceAgreement{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.MarketplaceOrdering/agreements/offers/plans",
			ApiVersions:  []string{"2015-06-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.accepted", Value: true},
			},
			RequiredFields: []string{
				"properties.accepted",
			},
			// Read-only server-populated properties returned by GET but never authored.
			ComputedFields: []string{
				"properties.licenseTextLink",
				"properties.privacyPolicyLink",
				"properties.plan",
				"properties.product",
				"properties.publisher",
				"properties.retrieveDatetime",
				"properties.signature",
			},
		},
	}
}

func init() { azwise.Register(NewMarketplaceAgreement()) }
