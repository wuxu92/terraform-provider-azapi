package bot

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BotConnection provides resource knowledge for
// Microsoft.BotService/botServices/connections.
//
// Mirrors azurerm_bot_connection. name/bot_name/resource_group_name are
// operational-envelope / parent-reference fields (part of the resource ID), so
// their AzureRM ForceNew flags are not encoded as body ForceNew rules.
//
// The management-plane body is botservice.ConnectionSettingProperties. AzureRM's
// service_provider_name argument (ForceNew) is not sent verbatim: AzureRM looks
// up the matching service provider and sends its id as
// properties.serviceProviderId — that ARM path is therefore ForceNew. (On
// Update, AzureRM instead sends properties.serviceProviderDisplayName, which is a
// mutable display name, so it is NOT ForceNew.)
//
// The `parameters` TypeMap expands into properties.parameters, an array of
// {key,value} ConnectionSettingParameter objects — a map-to-array reshaping that
// has no single scalar ARM body path, so no declarative rule is emitted for it.
//
// All schema ValidateFuncs on this resource are validation.StringIsNotEmpty
// (non-empty only), which is not expressible as an enum/regex/length StringRule,
// so no declarative StringRules are emitted.
//
// Sources:
//   - terraform-provider-azurerm internal/services/bot/bot_connection_resource.go
//     (schema + Create/Update/Read; timeouts 30m/5m/30m/30m; service_provider_name
//     resolved to serviceProviderId in the create body; client_secret sensitive)
//   - jackofallops/kermit sdk/botservice/2021-05-01-preview/botservice/models.go
//     ConnectionSettingProperties: settingId READ-ONLY; provisioningState
//     read-only; clientId/clientSecret/scopes/serviceProviderId/parameters
//     settable.
type BotConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BotConnection)(nil)

// NewBotConnection returns knowledge for the botServices/connections resource.
func NewBotConnection() *BotConnection {
	return &BotConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.BotService/botServices/connections",
			ApiVersions:  []string{"2021-05-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// service_provider_name (ForceNew) is resolved to serviceProviderId in
			// the create body; that ARM path is ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.serviceProviderId"},
			},
			// Required body inputs for creation (AzureRM Required: true).
			RequiredFields: []string{
				"properties.serviceProviderId",
				"properties.clientId",
				"properties.clientSecret",
			},
			SensitiveFields: []string{
				"properties.clientSecret",
			},
			// Returned by GET but absent from / read-only in the create body.
			ComputedFields: []string{
				"properties.settingId",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewBotConnection()) }
