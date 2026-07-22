package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SourceControlToken provides resource knowledge for Microsoft.Web/sourcecontrols.
//
// This is a subscription-level resource: the source control name (GitHub, Bitbucket,
// Dropbox, OneDrive) is the resource name, and the OAuth token/secret are body
// properties. It is distinct from Microsoft.Web/sites/sourcecontrols/web
// (azurerm_app_service_source_control), which configures a repository on a site.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/source_control_token_resource.go:22-60
//     (azurerm_source_control_token schema: required token, optional token_secret, type enum)
//   - terraform-provider-azurerm internal/services/appservice/source_control_token_resource.go:74-117
//     (create mapping to resourceproviders.SourceControl; 5m timeout)
//   - terraform-provider-azurerm internal/services/appservice/source_control_token_resource.go:119-216
//     (read/delete/update: same 5m timeouts, token/tokenSecret round-trip)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/resourceproviders/model_sourcecontrol.go:6-12
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/resourceproviders/model_sourcecontrolproperties.go:12-17
//
// Intentionally skipped here:
//   - type: this is the resource NAME segment (id.SourceControlName), not a body
//     property. Its enum is captured as a name-level StringRule (empty PropertyPath).
//   - properties.expirationTime / properties.refreshToken: present in the SDK model
//     but not managed by AzureRM (server-computed OAuth metadata); not listed as
//     ComputedFields because they are also settable on the ARM create/update body,
//     so stripping them would discard valid raw ARM input.
type SourceControlToken struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SourceControlToken)(nil)

// NewSourceControlToken returns knowledge for the sourcecontrols resource.
func NewSourceControlToken() *SourceControlToken {
	return &SourceControlToken{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sourcecontrols",
			ApiVersions:  []string{"2023-01-01"},
			RequiredFields: []string{
				"properties.token",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 5 * time.Minute,
				Read:   5 * time.Minute,
				Update: 5 * time.Minute,
				Delete: 5 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					AllowedValues: []string{"Bitbucket", "Dropbox", "GitHub", "OneDrive"},
					Message:       "source control name must be Bitbucket, Dropbox, GitHub, or OneDrive",
				},
				{
					PropertyPath: "properties.token",
					MinLength:    1,
					Message:      "token must not be empty",
				},
				{
					PropertyPath: "properties.tokenSecret",
					MinLength:    1,
					Message:      "token secret must not be empty",
				},
			},
			SensitiveFields: []string{
				"properties.token",
				"properties.tokenSecret",
			},
		},
	}
}

func init() { azwise.Register(NewSourceControlToken()) }
