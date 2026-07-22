package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteRecoveryReplicationRecoveryPlan provides resource knowledge for
// Microsoft.RecoveryServices/vaults/replicationRecoveryPlans
// (azurerm_site_recovery_replication_recovery_plan).
//
// The create body (CreateRecoveryPlanInputProperties) requires primaryFabricId,
// recoveryFabricId and the groups array. The recovery-group / action / A2A
// provider-specific structures are deeply nested arrays (properties.groups[*],
// providerSpecificInput[*]), which azwise cannot lower through array elements, so
// no per-element rules are encoded.
//
// Sources:
//   - AzureRM internal/services/recoveryservices/site_recovery_replication_recovery_plan_resource.go
//     :90-94   (schema: name StringMatch letters/numbers/hyphens, start letter, end letter/number)
//     :325-327,391,442,503 (timeouts create/update/delete 30m, read 5m)
//     :337-384 (expand: primaryFabricId = source fabric, recoveryFabricId = target fabric, groups)
//   - go-azure-sdk resource-manager/recoveryservicessiterecovery/2024-04-01/replicationrecoveryplans:
//     id_replicationrecoveryplan.go (Segments: .../vaults/{vaultName}/replicationRecoveryPlans/{name}),
//     model_createrecoveryplaninputproperties.go
//     (primaryFabricId, recoveryFabricId, groups required; providerSpecificInput optional array)
type SiteRecoveryReplicationRecoveryPlan struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteRecoveryReplicationRecoveryPlan)(nil)

// NewSiteRecoveryReplicationRecoveryPlan returns knowledge for the
// replicationRecoveryPlans resource.
func NewSiteRecoveryReplicationRecoveryPlan() *SiteRecoveryReplicationRecoveryPlan {
	return &SiteRecoveryReplicationRecoveryPlan{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/replicationRecoveryPlans",
			ApiVersions:  []string{"2024-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: start with a letter, contain letters/numbers/
					// hyphens, end with a letter or number.
					Regex:   `[a-zA-Z][a-zA-Z0-9-]{1,63}[a-zA-Z0-9]$`,
					Message: "recovery plan name must start with a letter, contain only letters, numbers and hyphens, and end with a letter or number",
				},
			},
			RequiredFields: []string{
				"properties.primaryFabricId",
				"properties.recoveryFabricId",
				"properties.groups",
			},
		},
	}
}

func init() { azwise.Register(NewSiteRecoveryReplicationRecoveryPlan()) }
