package appconfiguration

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AppConfiguration provides resource knowledge for
// Microsoft.AppConfiguration/configurationStores.
//
// Mirrors azurerm_app_configuration.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appconfiguration/app_configuration_resource.go
//     schema (lines 37-294) + Create (296-412): sku/soft_delete_retention_days ForceNew,
//     data-plane proxy, encryption, public/local auth, purge protection; timeouts 60m/5m/60m/60m;
//     soft-delete recovery via purgeDeletedPoller (Delete 691-778).
//   - go-azure-sdk resource-manager/appconfiguration/2024-05-01/configurationstores
//     ConfigurationStore (sku.name required), ConfigurationStoreProperties
//     (dataPlaneProxy.authenticationMode/privateLinkDelegation, disableLocalAuth,
//     enablePurgeProtection, encryption.keyVaultProperties, publicNetworkAccess,
//     softDeleteRetentionInDays, endpoint read-only), constants.go enums.
type AppConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AppConfiguration)(nil)

// NewAppConfiguration returns knowledge for the configurationStores resource.
func NewAppConfiguration() *AppConfiguration {
	return &AppConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppConfiguration/configurationStores",
			ApiVersions:  []string{"2024-05-01"},
			SoftDelete:   true,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			// soft_delete_retention_days is unconditionally ForceNew in AzureRM.
			//
			// sku is only ForceNew on a *downgrade* (premium/standard -> developer, or
			// anything -> free) via a CustomizeDiff ForceNewIfChange; upgrades are
			// in-place. That conditional replacement cannot be expressed as an
			// unconditional ForceNewRule without triggering spurious replacements on
			// upgrade, so it is documented here but not declared.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.softDeleteRetentionInDays"},
			},
			// ARM requires the sku on the create body; AzureRM defaults it to "free".
			RequiredFields: []string{
				"sku.name",
			},
			StringRules: []azwise.StringRule{
				// sku is validated against a fixed set (not an ARM enum;
				// https://github.com/Azure/azure-rest-api-specs/issues/23902).
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"free", "developer", "standard", "premium"},
					Message:       "sku must be one of free, developer, standard, premium",
				},
				// data_plane_proxy_authentication_mode -> AuthenticationMode enum.
				{
					PropertyPath:  "properties.dataPlaneProxy.authenticationMode",
					AllowedValues: []string{"Local", "Pass-through"},
					Message:       "authentication mode must be Local or Pass-through",
				},
				// data_plane_proxy_private_link_delegation_enabled -> PrivateLinkDelegation enum.
				{
					PropertyPath:  "properties.dataPlaneProxy.privateLinkDelegation",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "private link delegation must be Enabled or Disabled",
				},
				// public_network_access -> PublicNetworkAccess enum.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "public network access must be Enabled or Disabled",
				},
				// encryption.identity_client_id -> validation.IsUUID.
				{
					PropertyPath: "properties.encryption.keyVaultProperties.identityClientId",
					Regex:        "^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$",
					Message:      "identity client id must be a valid UUID",
				},
				// encryption.key_vault_key_identifier -> validation.IsURLWithHTTPorHTTPS.
				{
					PropertyPath: "properties.encryption.keyVaultProperties.keyIdentifier",
					Regex:        "^https?://.+",
					Message:      "key vault key identifier must be an http(s) URL",
				},
			},
			IntRules: []azwise.IntRule{
				// soft_delete_retention_days -> validation.IntBetween(1, 7).
				{
					PropertyPath: "properties.softDeleteRetentionInDays",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(7)),
					Message:      "soft delete retention must be between 1 and 7 days",
				},
			},
			// endpoint is returned by Azure, never set by the user.
			ComputedFields: []string{
				"properties.endpoint",
			},
			// AzureRM schema defaults; Azure fills these in when omitted.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.name", Value: "free"},
				{PropertyPath: "properties.dataPlaneProxy.authenticationMode", Value: "Local"},
				{PropertyPath: "properties.dataPlaneProxy.privateLinkDelegation", Value: "Disabled"},
				{PropertyPath: "properties.disableLocalAuth", Value: false}, // local_auth_enabled default true -> disableLocalAuth false
				{PropertyPath: "properties.enablePurgeProtection", Value: false},
				{PropertyPath: "properties.softDeleteRetentionInDays", Value: float64(7)},
				// public_network_access has no explicit default (Azure decides).
				{PropertyPath: "properties.publicNetworkAccess"},
			},
		},
	}
}

func init() { azwise.Register(NewAppConfiguration()) }
