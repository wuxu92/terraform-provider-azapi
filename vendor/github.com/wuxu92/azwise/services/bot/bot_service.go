package bot

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BotService provides resource knowledge for Microsoft.BotService/botServices.
//
// This single ARM type is exposed by AzureRM as three typed Terraform resources
// that differ ONLY by the Kind discriminator on the create body:
//   - azurerm_bot_service_azure_bot    -> Kind "azurebot"
//   - azurerm_bot_web_app              -> Kind "sdk"
//   - azurerm_bot_channels_registration-> Kind "bot"
//
// Only knowledge TRUE FOR EVERY kind is unioned here (sku/app validation, the
// ForceNew identity fields, universal read-only properties, timeouts). Kind-specific
// arguments are documented below but intentionally NOT encoded as declarative rules,
// because doing so would corrupt validation for the other kinds:
//
//   - azure_bot only: local_authentication_enabled -> properties.disableLocalAuth
//     (default false), public_network_access_enabled -> properties.publicNetworkAccess
//     (default "Enabled"), streaming_endpoint_enabled -> properties.isStreamingSupported
//     (default false), icon_url -> properties.iconUrl (default bot-framework png),
//     cmk_key_vault_key_url -> properties.cmekKeyVaultUrl/isCmekEnabled,
//     luis_app_ids -> properties.luisAppIds, luis_key -> properties.luisKey.
//   - channels_registration (KindBot) only: description -> properties.description,
//     icon_url -> properties.iconUrl (default png), streaming_endpoint_enabled ->
//     properties.isStreamingSupported (default false), public_network_access_enabled ->
//     properties.publicNetworkAccess (no default), cmk_key_vault_url ->
//     properties.cmekKeyVaultUrl/isCmekEnabled.
//   - web_app (KindSdk) only: luis_app_ids -> properties.luisAppIds,
//     luis_key -> properties.luisKey.
//
// The `name` argument is ForceNew in all three but is the operational-envelope
// resource name (not an ARM body property), so it is not encoded as a body ForceNew
// rule; only a non-empty name constraint is kept.
//
// Sources:
//   - terraform-provider-azurerm internal/services/bot/bot_service_azure_bot_resource.go
//     (Arguments L71-226, Create body L268-307; timeouts 30m/5m/30m/30m)
//   - terraform-provider-azurerm internal/services/bot/bot_web_app_resource.go
//     (Schema L68-172, Create body L217-243)
//   - terraform-provider-azurerm internal/services/bot/bot_channels_registration_resource.go
//     (Schema L71-188, Create body L236-269)
//   - jackofallops/kermit sdk/botservice/2021-05-01-preview/botservice
//     BotProperties (models.go L520-579), Sku (L4332), Kind/MsaAppType/SkuName enums
//     (enums.go L128-258). Read-only body fields: endpointVersion, configuredChannels,
//     enabledChannels, cmekEncryptionStatus, isDeveloperAppInsightsApiKeySet,
//     migrationToken, sku.tier.
type BotService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BotService)(nil)

// NewBotService returns knowledge for the botServices resource (all bot kinds).
func NewBotService() *BotService {
	return &BotService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.BotService/botServices",
			ApiVersions:  []string{"2021-05-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// ForceNew fields common to every bot kind. `name` is ForceNew too but is
			// the envelope resource name, not a body property, so it is omitted here.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "sku.name"},
				{PropertyPath: "properties.msaAppId"},
				{PropertyPath: "properties.msaAppType"},
				{PropertyPath: "properties.msaAppTenantId"},
				// microsoft_app_msi_id (azure_bot) / microsoft_app_user_assigned_identity_id
				// (web_app, channels) both map to msaAppMSIResourceId and are ForceNew.
				{PropertyPath: "properties.msaAppMSIResourceId"},
			},
			StringRules: []azwise.StringRule{
				{
					// name uses validation.StringIsNotEmpty in all three resources.
					PropertyPath: "",
					MinLength:    1,
					Message:      "bot name must not be empty",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"F0", "S1"},
					Message:       "sku must be one of F0 or S1",
				},
				{
					PropertyPath:  "properties.msaAppType",
					AllowedValues: []string{"MultiTenant", "SingleTenant", "UserAssignedMSI"},
					Message:       "microsoft app type must be MultiTenant, SingleTenant or UserAssignedMSI",
				},
			},
			// Required in every kind's create body (excludes envelope name/location).
			RequiredFields: []string{
				"sku.name",
				"properties.msaAppId",
				"properties.msaAppType",
			},
			// Sensitive body inputs. developerAppInsightsApiKey is present in all three;
			// luisKey is set only by azure_bot/web_app but the path is safe to mark since
			// it never appears in a channels_registration body.
			SensitiveFields: []string{
				"properties.developerAppInsightsApiKey",
				"properties.luisKey",
			},
			// Read-only properties returned by GET but absent from the create intent
			// (marked READ-ONLY in the SDK BotProperties/Sku models). Universal to all kinds.
			ComputedFields: []string{
				"properties.endpointVersion",
				"properties.configuredChannels",
				"properties.enabledChannels",
				"properties.cmekEncryptionStatus",
				"properties.isDeveloperAppInsightsApiKeySet",
				"properties.migrationToken",
				"sku.tier",
			},
		},
	}
}

func init() { azwise.Register(NewBotService()) }
