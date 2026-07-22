package videoindexer

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VideoIndexerAccount provides resource knowledge for Microsoft.VideoIndexer/accounts.
//
// Mirrors azurerm_video_indexer_account.
//
// Sources:
//   - internal/services/videoindexer/video_indexer_account_resource.go
//     (schema 54-98: name ForceNew ServicePlanName (^[0-9a-zA-Z-_]{1,60}$);
//     storage Required MaxItems 1 {storage_account_id Required ForceNew
//     ValidateStorageAccountID; user_assigned_identity_id Optional
//     ValidateUserAssignedIdentityID}; identity SystemAssignedUserAssigned Required;
//     public_network_access Optional Default Enabled StringInSlice; timeouts
//     Create/Update/Delete 60m Read 5m;
//     create 146-154 → Account{Identity, Location, Properties{PublicNetworkAccess,
//     StorageServices{ResourceId, UserAssignedIdentity}}}).
//   - internal/services/appservice/validate/service_plan_name.go (regex ^[0-9a-zA-Z-_]{1,60}$).
//   - go-azure-sdk resource-manager/videoindexer/2025-04-01/accounts:
//     model_accountpropertiesforputrequest.go (publicNetworkAccess/storageServices/
//     openAiServices settable; accountId/tenantId/provisioningState/totalMinutesIndexed/
//     totalSecondsIndexed read-only), model_storageservicesforputrequest.go
//     (resourceId/userAssignedIdentity), constants.go (PublicNetworkAccess
//     Disabled/Enabled), id_account.go (segment casing "Microsoft.VideoIndexer"/"accounts").
type VideoIndexerAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VideoIndexerAccount)(nil)

// NewVideoIndexerAccount returns knowledge for the accounts resource.
func NewVideoIndexerAccount() *VideoIndexerAccount {
	return &VideoIndexerAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.VideoIndexer/accounts",
			ApiVersions:  []string{"2025-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.storageServices.resourceId"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: ServicePlanName regex, 1-60 chars.
					Regex:     `^[0-9a-zA-Z-_]{1,60}$`,
					MinLength: 1,
					MaxLength: 60,
					Message:   "must be 1-60 characters and contain only letters, numbers, hyphens and underscores",
				},
				{
					// public_network_access: StringInSlice(PossibleValuesForPublicNetworkAccess).
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
				},
			},
			RequiredFields: []string{
				"properties.storageServices.resourceId",
			},
			DefaultValues: []azwise.DefaultValue{
				// public_network_access Default "Enabled".
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			ComputedFields: []string{
				"properties.accountId",
				"properties.tenantId",
				"properties.provisioningState",
				"properties.totalMinutesIndexed",
				"properties.totalSecondsIndexed",
			},
			// NOTE: storage_account_id uses commonids.ValidateStorageAccountID and
			// user_assigned_identity_id uses commonids.ValidateUserAssignedIdentityID
			// (semantic ARM resource-ID checks). Map them to an AzureResourceID
			// validator on properties.storageServices.resourceId /
			// properties.storageServices.userAssignedIdentity in the resource customizer.
			// NOTE: identity (SystemAssignedUserAssigned) is Required in AzureRM but is a
			// top-level ARM envelope field (not under properties), so it is not listed as
			// a body RequiredField.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewVideoIndexerAccount()) }
