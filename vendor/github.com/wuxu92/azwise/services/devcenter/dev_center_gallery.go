package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterGallery provides resource knowledge for
// Microsoft.DevCenter/devCenters/galleries.
//
// Contributing Terraform resource: azurerm_dev_center_gallery.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_gallery_resource.go
//     (schema L39-57, Create body L95-99)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/galleries:
//     id_gallery.go (type segment "galleries", parent "devCenters"),
//     model_galleryproperties.go.
//
// Key mappings:
//   - shared_gallery_id → properties.galleryResourceId (Required, ForceNew)
//
// Notes:
//   - name is envelope-owned; dev_center_id is the parent, not a body field.
//   - No Update: every argument is ForceNew.
type DevCenterGallery struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterGallery)(nil)

func NewDevCenterGallery() *DevCenterGallery {
	return &DevCenterGallery{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/devCenters/galleries",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.galleryResourceId"},
			},
			RequiredFields: []string{
				"properties.galleryResourceId",
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterGallery()) }
