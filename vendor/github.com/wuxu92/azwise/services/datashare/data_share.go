package datashare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataShare provides resource knowledge for
// Microsoft.DataShare/accounts/shares.
//
// Mirrors azurerm_data_share.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datashare/data_share_resource.go
//     schema (lines 48-113: name/account_id/kind ForceNew, kind StringInSlice
//     CopyBased/InPlace, description/terms optional, snapshot_schedule block),
//     CreateUpdate (117-181: body = ShareProperties{ShareKind, Description, Terms}),
//     Read/Delete; timeouts 30m/5m/30m/30m.
//   - internal/services/datashare/validate/share_name.go (name regex, 2-90 chars).
//   - go-azure-sdk resource-manager/datashare/2019-11-01/share
//     ShareProperties (shareKind enum; description/terms user-settable;
//     createdAt/provisioningState/userEmail/userName read-only) and
//     constants.go PossibleValuesForShareKind (CopyBased, InPlace).
//
// NOTE: the `snapshot_schedule` block is NOT part of the share body — AzureRM
// writes it to the separate ARM API Microsoft.DataShare/accounts/shares/
// synchronizationSettings (SDK package synchronizationsetting). Rules for it
// belong on that sub-service resource, not here.
type DataShare struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataShare)(nil)

// NewDataShare returns knowledge for the accounts/shares resource.
func NewDataShare() *DataShare {
	return &DataShare{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataShare/accounts/shares",
			ApiVersions:  []string{"2019-11-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.shareKind"},
			},
			// kind is Required:true in AzureRM (shareKind is optional in the ARM model
			// but must be supplied to create a valid share).
			RequiredFields: []string{
				"properties.shareKind",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM ShareName():
					// numbers, letters, - and _, 2-90 characters.
					Regex:     `^[\w-]{2,90}$`,
					MinLength: 2,
					MaxLength: 90,
					Message:   "DataShare name can only contain numbers, letters, - and _, and must be between 2 and 90 characters long.",
				},
				{
					// ShareKind enum — full ARM SDK set (PossibleValuesForShareKind).
					PropertyPath:  "properties.shareKind",
					AllowedValues: []string{"CopyBased", "InPlace"},
				},
			},
			// Server-populated, read-only ARM properties (ShareProperties).
			ComputedFields: []string{
				"properties.createdAt",
				"properties.provisioningState",
				"properties.userEmail",
				"properties.userName",
			},
		},
	}
}

func init() { azwise.Register(NewDataShare()) }
