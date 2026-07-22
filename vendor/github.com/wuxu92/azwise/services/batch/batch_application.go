package batch

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BatchApplication provides resource knowledge for
// Microsoft.Batch/batchAccounts/applications.
//
// Contributing Terraform resource: azurerm_batch_application.
//
// Sources:
//   - terraform-provider-azurerm internal/services/batch/batch_application_resource.go
//     (schema L41-75, Create body L107-113)
//   - terraform-provider-azurerm internal/services/batch/validate/application_name.go,
//     application_version.go, application_display_name.go
//   - go-azure-sdk resource-manager/batch/2024-07-01/application:
//     model_applicationproperties.go (Create body: allowUpdates, defaultVersion, displayName).
//
// Notes:
//   - name & account_name are envelope-owned (Required+ForceNew); resource_group_name is
//     the parent RG. None are ARM-body properties, so no body ForceNew rules are emitted.
//   - allow_updates/default_version/display_name are all Optional (settable on update),
//     so there are no body ForceNew rules and no ComputedFields.
type BatchApplication struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BatchApplication)(nil)

// NewBatchApplication returns knowledge for the batchAccounts/applications resource.
func NewBatchApplication() *BatchApplication {
	return &BatchApplication{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Batch/batchAccounts/applications",
			ApiVersions:  []string{"2024-07-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ApplicationName
				{
					Regex:     `^[-_\da-zA-Z]+$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "may contain any combination of alphanumeric characters, hyphens, and underscores (1-64 chars)",
				},
				// ── default_version → properties.defaultVersion ── validate.ApplicationVersion
				{
					PropertyPath: "properties.defaultVersion",
					Regex:        `^[-._\da-zA-Z]+$`,
					MinLength:    1,
					MaxLength:    64,
					Message:      "may contain any combination of alphanumeric characters, hyphens, underscores, and periods (1-64 chars)",
				},
				// ── display_name → properties.displayName ── validate.ApplicationDisplayName
				{
					PropertyPath: "properties.displayName",
					MinLength:    1,
					MaxLength:    1024,
					Message:      "must be between 1 and 1024 characters",
				},
			},
			// allow_updates Default true → properties.allowUpdates.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.allowUpdates", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewBatchApplication()) }
