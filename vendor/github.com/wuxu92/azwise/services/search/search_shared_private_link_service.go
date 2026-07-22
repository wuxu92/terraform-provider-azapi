package search

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SearchSharedPrivateLinkService provides resource knowledge for
// Microsoft.Search/searchServices/sharedPrivateLinkResources.
//
// Mirrors azurerm_search_shared_private_link_service. name and search_service_id
// are envelope / parent-reference fields (ForceNew) and are not modelled as body
// rules. subresource_name and target_resource_id are Required + ForceNew body
// fields.
//
// Sources:
//   - terraform-provider-azurerm internal/services/search/search_shared_private_link_service_resource.go:39-75
//     (schema), :98-149 (create -> sharedprivatelinkresources.SharedPrivateLinkResource),
//     :147/:196/:231/:253 (timeouts: Create 60m / Read 5m / Update 60m / Delete 60m).
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     subresource_name (networkValidate.PrivateLinkSubResourceName -> properties.groupId);
//     target_resource_id (azure.ValidateResourceID -> properties.privateLinkResourceId,
//     resource-id validator).
//   - go-azure-sdk resource-manager/search/2025-05-01/sharedprivatelinkresources
//     model_sharedprivatelinkresource.go / model_sharedprivatelinkresourceproperties.go /
//     constants.go.
type SearchSharedPrivateLinkService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SearchSharedPrivateLinkService)(nil)

// NewSearchSharedPrivateLinkService returns knowledge for the search shared
// private link resource.
func NewSearchSharedPrivateLinkService() *SearchSharedPrivateLinkService {
	return &SearchSharedPrivateLinkService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Search/searchServices/sharedPrivateLinkResources",
			ApiVersions:  []string{"2025-05-01"},
			ForceNew: []azwise.ForceNewRule{
				// subresource_name (ForceNew) -> properties.groupId.
				{PropertyPath: "properties.groupId"},
				// target_resource_id (ForceNew) -> properties.privateLinkResourceId.
				{PropertyPath: "properties.privateLinkResourceId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			RequiredFields: []string{
				"properties.groupId",
				"properties.privateLinkResourceId",
			},
			// Read-only body properties: present in the GET model, server-computed.
			ComputedFields: []string{
				"properties.status",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewSearchSharedPrivateLinkService()) }
