package maps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MapsAccount provides resource knowledge for Microsoft.Maps/accounts.
//
// Mirrors azurerm_maps_account. local_authentication_enabled inverts onto
// properties.disableLocalAuth (enabled=true -> disableLocalAuth=false).
//
// Sources:
//   - terraform-provider-azurerm internal/services/maps/maps_account_resource.go
//   - Schema(): name (Required, ForceNew, validate.AccountName StringMatch),
//     sku_name (Required, ForceNew, enum S0/S1/G2 -> sku.name),
//     local_authentication_enabled (Optional, Default true -> properties.disableLocalAuth inverted),
//     cors (-> properties.cors), data_store (-> properties.linkedResources)
//   - resourceMapsAccountCreate(): parameters.Properties.DisableLocalAuth = !local_authentication_enabled
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/maps/2023-06-01/accounts model_mapsaccountproperties.go /
//     constants.go (Name: S0, S1, G2)
//
// primary_access_key / secondary_access_key / x_ms_client_id are Computed TF-only
// attributes surfaced via ListKeys / uniqueId, not settable body fields, so no
// ComputedFields entry is needed (they are not part of the PUT body).
type MapsAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MapsAccount)(nil)

// NewMapsAccount returns knowledge for the Maps accounts resource.
func NewMapsAccount() *MapsAccount {
	return &MapsAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Maps/accounts",
			ApiVersions:  []string{"2023-06-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name is envelope.
				{PropertyPath: "location"},
				// sku_name is ForceNew.
				{PropertyPath: "sku.name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// sku.name is Required. location is envelope-owned.
			RequiredFields: []string{
				"sku.name",
			},
			DefaultValues: []azwise.DefaultValue{
				// local_authentication_enabled Default true -> disableLocalAuth false.
				{PropertyPath: "properties.disableLocalAuth", Value: false},
			},
			StringRules: []azwise.StringRule{
				// Resource name: validate.AccountName StringMatch.
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9]{1}[A-Za-z0-9._-]{1,}$`,
					Message:      "name must start with an alphanumeric character followed by alphanumerics, underscore, period, or hyphen",
				},
				// sku_name enum. The SDK Name enum is exactly S0/S1/G2 (full set).
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"S0", "S1", "G2"},
					Message:       "sku.name must be one of S0, S1, or G2",
				},
			},
		},
	}
}

func init() { azwise.Register(NewMapsAccount()) }
