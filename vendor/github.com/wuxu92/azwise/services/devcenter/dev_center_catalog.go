package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterCatalog provides resource knowledge for
// Microsoft.DevCenter/devCenters/catalogs.
//
// Contributing Terraform resource: azurerm_dev_center_catalog.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_catalog_resource.go
//     (schema L52-74, block schema L249-281, expand L283-294)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/catalogs:
//     id_devcentercatalog.go (type segment "catalogs", parent "devCenters"),
//     model_catalogproperties.go, model_gitcatalog.go.
//
// Key mappings (catalog_github → properties.gitHub, catalog_adogit → properties.adoGit):
//   - <block>.uri               → properties.{gitHub,adoGit}.uri
//   - <block>.branch            → properties.{gitHub,adoGit}.branch
//   - <block>.key_vault_key_url → properties.{gitHub,adoGit}.secretIdentifier
//   - <block>.path              → properties.{gitHub,adoGit}.path
//
// Notes:
//   - name/resource_group_name are envelope-owned; dev_center_id is the parent, not a body field.
//   - catalog_github and catalog_adogit are both Optional with no schema-level ExactlyOneOf,
//     so no relational rule is emitted.
type DevCenterCatalog struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterCatalog)(nil)

func NewDevCenterCatalog() *DevCenterCatalog {
	return &DevCenterCatalog{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/devCenters/catalogs",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterCatalog()) }
