package resources

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ResourceManagementPrivateLink provides resource knowledge for
// Microsoft.Authorization/resourceManagementPrivateLinks.
//
// Mirrors azurerm_resource_management_private_link. The resource is little more than a
// named location envelope — the ARM body carries only `location`. There is no Update, so
// every field is replace-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/resource/resource_management_private_link_resource.go:43-99,134-152
//     (name=StringIsNotEmpty ForceNew, resource_group_name + location envelope, 30m/5m/30m
//     timeouts, no Update, payload = ResourceManagementPrivateLinkLocation{location})
//   - go-azure-sdk resource-manager/resources/2020-05-01/resourcemanagementprivatelink/model_resourcemanagementprivatelinklocation.go:6-8
//     (ResourceManagementPrivateLinkLocation{location} — ARM body json tag)
//   - go-azure-sdk resource-manager/resources/2020-05-01/resourcemanagementprivatelink/id_resourcemanagementprivatelink.go:110-119
//     (.../providers/Microsoft.Authorization/resourceManagementPrivateLinks/{name} — ARM type casing)
//
// Intentionally skipped (documented, no rule emitted):
//   - location: Required+ForceNew but an envelope field (captured below only as ForceNew).
//   - resource_group_name: envelope / parent scope.
type ResourceManagementPrivateLink struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ResourceManagementPrivateLink)(nil)

// NewResourceManagementPrivateLink returns knowledge for the resourceManagementPrivateLinks resource.
func NewResourceManagementPrivateLink() *ResourceManagementPrivateLink {
	return &ResourceManagementPrivateLink{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/resourceManagementPrivateLinks",
			ApiVersions:  []string{"2020-05-01"},
			SoftDelete:   false,
			// No Update function: changing location replaces the resource.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM validation.StringIsNotEmpty.
					MinLength: 1,
					Message:   "name must not be empty",
				},
			},
		},
	}
}

func init() { azwise.Register(NewResourceManagementPrivateLink()) }
