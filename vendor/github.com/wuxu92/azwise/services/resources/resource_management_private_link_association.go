package resources

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateLinkAssociation provides resource knowledge for
// Microsoft.Authorization/privateLinkAssociations.
//
// Mirrors azurerm_resource_management_private_link_association. The association is a scoped
// resource under a management group; its name (a UUID, auto-generated when omitted) and the
// management group live on the ID, while privateLink + publicNetworkAccess form the ARM body
// under `properties`. There is no Update — every field is replace-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/resource/resource_management_private_link_association_resource.go:46-137,175-193
//     (name=IsUUID Optional ForceNew, management_group_id ForceNew scope, resource_management_private_link_id
//     Required ForceNew, public_network_access_enabled bool Required ForceNew, tenant_id Computed,
//     30m/5m/30m timeouts, no Update)
//   - go-azure-sdk resource-manager/resources/2020-05-01/privatelinkassociation/model_privatelinkassociationproperties.go:6-9
//     (PrivateLinkAssociationProperties{privateLink,publicNetworkAccess} — ARM body json tags)
//   - go-azure-sdk resource-manager/resources/2020-05-01/privatelinkassociation/model_privatelinkassociationpropertiesexpanded.go:6-11
//     (adds read-only scope + tenantID)
//   - go-azure-sdk resource-manager/resources/2020-05-01/privatelinkassociation/constants.go:12-24
//     (PublicNetworkAccessOptions: Disabled/Enabled — full SDK set)
//   - go-azure-sdk resource-manager/resources/2020-05-01/privatelinkassociation/id_privatelinkassociation.go:96-113
//     (.../Microsoft.Management/managementGroups/{groupId}/providers/Microsoft.Authorization/privateLinkAssociations/{plaId})
//
// Intentionally skipped (documented, no rule emitted):
//   - name: AzureRM validates with validation.IsUUID — a semantic UUID check. It targets the
//     resource-name attribute (auto-generated when omitted), not a body path, so it belongs in
//     an azapin customizer validator (schema/validators.UUID), not a declarative StringRule.
//   - management_group_id: parent scope; lives on the ID, not the body.
//   - resource_management_private_link_id: ValidateResourceManagementPrivateLinkID is a resource-ID
//     semantic validator; the value maps to properties.privateLink (captured as Required+ForceNew),
//     but the ID-format check belongs in a customizer, not a declarative StringRule.
type PrivateLinkAssociation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateLinkAssociation)(nil)

// NewPrivateLinkAssociation returns knowledge for the privateLinkAssociations resource.
func NewPrivateLinkAssociation() *PrivateLinkAssociation {
	return &PrivateLinkAssociation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/privateLinkAssociations",
			ApiVersions:  []string{"2020-05-01"},
			SoftDelete:   false,
			// No Update function: the body properties force replacement.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.privateLink"},
				{PropertyPath: "properties.publicNetworkAccess"},
			},
			RequiredFields: []string{
				"properties.privateLink",
				"properties.publicNetworkAccess",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// public_network_access_enabled (bool) -> properties.publicNetworkAccess enum.
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be one of Enabled, Disabled",
				},
			},
			ComputedFields: []string{
				// tenant_id is read-only (PrivateLinkAssociationPropertiesExpanded.tenantID).
				"properties.tenantID",
			},
		},
	}
}

func init() { azwise.Register(NewPrivateLinkAssociation()) }
