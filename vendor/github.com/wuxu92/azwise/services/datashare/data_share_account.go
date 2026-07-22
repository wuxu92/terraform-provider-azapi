package datashare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataShareAccount provides resource knowledge for
// Microsoft.DataShare/accounts.
//
// Mirrors azurerm_data_share_account.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datashare/data_share_account_resource.go
//     schema (lines 48-65: name/location/identity ForceNew, identity RequiredForceNew,
//     tags lower-case), Create (69-111: body = Name/Location/Identity/Tags),
//     Update (147-168: only tags mutable), Read/Delete; timeouts 30m/5m/30m/30m.
//   - internal/services/datashare/validate/account_name.go (name regex, 3-90 chars).
//   - go-azure-sdk resource-manager/datashare/2019-11-01/account
//     Account (Identity required json:"identity"; Location/Name/Tags) and
//     AccountProperties (createdAt/provisioningState/userEmail/userName all read-only).
type DataShareAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataShareAccount)(nil)

// NewDataShareAccount returns knowledge for the accounts resource.
func NewDataShareAccount() *DataShareAccount {
	return &DataShareAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataShare/accounts",
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
				{PropertyPath: "location"},
				// commonschema.SystemAssignedIdentityRequiredForceNew()
				{PropertyPath: "identity"},
			},
			// Account.Identity is required by the ARM model (json:"identity", no omitempty).
			RequiredFields: []string{
				"identity",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM AccountName():
					// 3-90 chars, cannot contain <>%&:\?/#*$^();,.|+={}[]!~@.
					Regex:     `^[^<>%&:\\?/#*$^();,.\|+={}\[\]!~@]{3,90}$`,
					MinLength: 3,
					MaxLength: 90,
					Message:   `Data share account name should have length of 3 - 90, and cannot contain <>%&:\?/#*$^();,.|+={}[]!~@.`,
				},
			},
			// Server-populated, read-only ARM properties (AccountProperties).
			ComputedFields: []string{
				"properties.createdAt",
				"properties.provisioningState",
				"properties.userEmail",
				"properties.userName",
			},
		},
	}
}

func init() { azwise.Register(NewDataShareAccount()) }
