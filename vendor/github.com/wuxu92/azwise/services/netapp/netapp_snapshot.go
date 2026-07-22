package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppSnapshot provides resource knowledge for
// Microsoft.NetApp/netAppAccounts/capacityPools/volumes/snapshots.
//
// Mirrors azurerm_netapp_snapshot.
//
// Sources:
//   - internal/services/netapp/netapp_snapshot_resource.go
//     (schema 26-74: name ForceNew+SnapshotName; parents account/pool/volume ForceNew;
//     no Update; timeouts Create/Delete 30m Read 5m; create 77-108 → Snapshot{Location}
//     with no user-settable properties).
//   - internal/services/netapp/validate/snapshot_name.go (regex ^[\da-zA-Z][-_\da-zA-Z]{3,63}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/snapshots:
//     model_snapshotproperties.go (created/provisioningState/snapshotId all read-only),
//     id_snapshot.go (type segment casing "snapshots").
type NetAppSnapshot struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppSnapshot)(nil)

// NewNetAppSnapshot returns knowledge for the snapshots resource.
func NewNetAppSnapshot() *NetAppSnapshot {
	return &NetAppSnapshot{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/capacityPools/volumes/snapshots",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM SnapshotName: 4-64 chars, start alphanumeric.
					Regex:     `^[\da-zA-Z][-_\da-zA-Z]{3,63}$`,
					MinLength: 4,
					MaxLength: 64,
					Message:   "must be 4-64 characters, start with a letter or number, and contain only letters, numbers, underscores and hyphens",
				},
			},
			// SnapshotProperties is entirely server-populated (read-only).
			ComputedFields: []string{
				"properties.snapshotId",
				"properties.provisioningState",
				"properties.created",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppSnapshot()) }
