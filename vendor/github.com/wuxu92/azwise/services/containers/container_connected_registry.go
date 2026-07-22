package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerConnectedRegistry provides resource knowledge for
// Microsoft.ContainerRegistry/registries/connectedRegistries.
//
// Contributing Terraform resource: azurerm_container_connected_registry.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_connected_registry_resource.go
//     (schema L50-180, Create/expand L198-263)
//   - terraform-provider-azurerm internal/services/containers/validate/container_registry_name.go
//   - go-azure-sdk resource-manager/containerregistry/2025-11-01/connectedregistries:
//     model_connectedregistryproperties.go, model_syncproperties.go,
//     model_parentproperties.go, model_loggingproperties.go, constants.go (enums).
//
// Notes:
//   - mode is ForceNew → properties.mode (default ReadWrite).
//   - sync_token_id is Required + ForceNew → properties.parent.syncProperties.tokenId.
//   - parent_registry_id is ForceNew → properties.parent.id.
//   - sync_message_ttl → properties.parent.syncProperties.messageTtl (ISO8601, default P1D);
//     sync_schedule → ...schedule (default "* * * * *"); sync_window → ...syncWindow.
//   - notification blocks expand into properties.notificationsList ([]string) via a custom
//     name:tag@digest:action encoding — a one-to-many transform not expressible declaratively.
//   - audit_log_enabled (bool) maps to properties.logging.auditLogStatus (Enabled/Disabled).
type ContainerConnectedRegistry struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerConnectedRegistry)(nil)

// NewContainerConnectedRegistry returns knowledge for the connectedRegistries resource.
func NewContainerConnectedRegistry() *ContainerConnectedRegistry {
	return &ContainerConnectedRegistry{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries/connectedRegistries",
			ApiVersions:  []string{"2025-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.mode"},
				{PropertyPath: "properties.parent.id"},
				{PropertyPath: "properties.parent.syncProperties.tokenId"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ContainerRegistryName
				{
					Regex:     `^[a-zA-Z0-9]+$`,
					MinLength: 5,
					MaxLength: 50,
					Message:   "alpha numeric characters only, 5-50 chars",
				},
				// ── mode → properties.mode
				{
					PropertyPath:  "properties.mode",
					AllowedValues: []string{"Mirror", "ReadOnly", "ReadWrite", "Registry"},
					Message:       "must be one of Mirror, ReadOnly, ReadWrite or Registry",
				},
				// ── log_level → properties.logging.logLevel
				{
					PropertyPath:  "properties.logging.logLevel",
					AllowedValues: []string{"Debug", "Error", "Information", "None", "Warning"},
					Message:       "must be one of Debug, Error, Information, None or Warning",
				},
				// ── audit_log_enabled → properties.logging.auditLogStatus
				{
					PropertyPath:  "properties.logging.auditLogStatus",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// mode Default ReadWrite.
				{PropertyPath: "properties.mode", Value: "ReadWrite"},
				// sync_schedule Default "* * * * *".
				{PropertyPath: "properties.parent.syncProperties.schedule", Value: "* * * * *"},
				// sync_message_ttl Default P1D.
				{PropertyPath: "properties.parent.syncProperties.messageTtl", Value: "P1D"},
				// log_level Default None.
				{PropertyPath: "properties.logging.logLevel", Value: "None"},
				// audit_log_enabled Default false → Disabled.
				{PropertyPath: "properties.logging.auditLogStatus", Value: "Disabled"},
			},
			// sync_token_id and mode are Required for creation (mode has a default).
			RequiredFields: []string{
				"properties.mode",
				"properties.parent.syncProperties.tokenId",
				"properties.parent.syncProperties.messageTtl",
			},
			// Read-only properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.activation",
				"properties.connectionState",
				"properties.lastActivityTime",
				"properties.loginServer",
				"properties.provisioningState",
				"properties.statusDetails",
				"properties.version",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewContainerConnectedRegistry()) }
