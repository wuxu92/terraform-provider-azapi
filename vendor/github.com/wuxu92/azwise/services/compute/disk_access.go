package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DiskAccess provides resource knowledge for Microsoft.Compute/diskAccesses.
//
// Contributing Terraform resource: azurerm_disk_access.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/disk_access_resource.go
//     (schema L23-56, Create body L80-83)
//   - go-azure-sdk resource-manager/compute/2022-03-02/diskaccesses:
//     model_diskaccess.go, model_diskaccessproperties.go.
//
// Notes:
//   - The resource carries only envelope fields (name/location/tags). name is ForceNew but
//     is not an ARM-body property; DiskAccessProperties has no user-settable create fields,
//     so there are no body ForceNew/validation rules to emit.
type DiskAccess struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DiskAccess)(nil)

// NewDiskAccess returns knowledge for the diskAccesses resource.
func NewDiskAccess() *DiskAccess {
	return &DiskAccess{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/diskAccesses",
			ApiVersions:  []string{"2022-03-02"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewDiskAccess()) }
