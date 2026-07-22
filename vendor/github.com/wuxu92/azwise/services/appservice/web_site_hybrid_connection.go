package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebSiteHybridConnection provides resource knowledge for
// Microsoft.Web/sites/hybridConnectionNamespaces/relays.
//
// Merged from two AzureRM Terraform resources that map to the same ARM type
// (they differ only in the parent-site envelope reference — web_app_id vs
// function_app_id — and produce identical HybridConnection request bodies):
//   - azurerm_web_app_hybrid_connection
//   - azurerm_function_app_hybrid_connection
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/web_app_hybrid_connection_resource.go:27-131
//     (schema: web_app_id/relay_id ForceNew, required hostname/port, optional send_key_name default, computed outputs)
//   - terraform-provider-azurerm internal/services/appservice/web_app_hybrid_connection_resource.go:133-191
//     (create mapping to webapps.HybridConnection; 30m create timeout)
//   - terraform-provider-azurerm internal/services/appservice/function_app_hybrid_connection_resource.go:27-96
//     (identical schema/body via function_app_id envelope)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_hybridconnectionproperties.go:6-15
//
// Intentionally skipped here:
//   - web_app_id / function_app_id and relay_id: these are envelope/ID segments
//     (parent site name, namespace name, relay name). relay_id also populates the
//     body property properties.relayArmUri, which IS captured as ForceNew.
//   - properties.relayName / properties.serviceBusNamespace / properties.serviceBusSuffix:
//     GET-only outputs (AzureRM Computed attributes, absent from the create body).
//   - properties.sendKeyValue: AzureRM sets this in the create body (fetched via
//     helpers.GetSendKeyValue) so it is settable — marked Sensitive but NOT
//     ComputedFields (stripping it would discard the fetched key on raw ARM writes).
type WebSiteHybridConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebSiteHybridConnection)(nil)

// NewWebSiteHybridConnection returns knowledge for the site hybrid-connection relay resource.
func NewWebSiteHybridConnection() *WebSiteHybridConnection {
	return &WebSiteHybridConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites/hybridConnectionNamespaces/relays",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.relayArmUri"},
			},
			RequiredFields: []string{
				"properties.hostname",
				"properties.port",
				"properties.relayArmUri",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 5 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.hostname",
					MinLength:    1,
					Message:      "hostname must not be empty",
				},
				{
					PropertyPath: "properties.sendKeyName",
					MinLength:    1,
					Message:      "send key name must not be empty",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.port",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(65535)),
					Message:      "port must be between 0 and 65535",
				},
			},
			SensitiveFields: []string{
				"properties.sendKeyValue",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.sendKeyName", Value: "RootManageSharedAccessKey"},
			},
		},
	}
}

func init() { azwise.Register(NewWebSiteHybridConnection()) }
