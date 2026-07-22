package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteRecoveryProtectionContainer provides resource knowledge for
// Microsoft.RecoveryServices/vaults/replicationFabrics/replicationProtectionContainers
// (azurerm_site_recovery_protection_container).
//
// The resource has only Create/Read/Delete (no Update); every attribute is a
// name or parent reference (recovery_vault_name, recovery_fabric_name), so it is
// entirely replace-on-change. The create body
// (CreateProtectionContainerInputProperties) carries no user-settable ARM
// properties, so there are no declarative value rules to encode.
//
// Sources:
//   - AzureRM internal/services/recoveryservices/site_recovery_protection_container_resource.go
//     :33-37  (timeouts create/delete 30m, read 5m)
//     :39-60  (schema: name StringIsNotEmpty, recovery_vault_name, recovery_fabric_name)
//     :90-92  (expand: CreateProtectionContainerInput{Properties: empty})
//   - go-azure-sdk resource-manager/recoveryservicessiterecovery/2024-04-01/replicationprotectioncontainers:
//     id_replicationprotectioncontainer.go
//     (Segments: .../replicationFabrics/{fabric}/replicationProtectionContainers/{name})
type SiteRecoveryProtectionContainer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteRecoveryProtectionContainer)(nil)

// NewSiteRecoveryProtectionContainer returns knowledge for the
// replicationProtectionContainers resource.
func NewSiteRecoveryProtectionContainer() *SiteRecoveryProtectionContainer {
	return &SiteRecoveryProtectionContainer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/replicationFabrics/replicationProtectionContainers",
			ApiVersions:  []string{"2024-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewSiteRecoveryProtectionContainer()) }
