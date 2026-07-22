package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlFlexibleServerBackup provides resource knowledge for
// Microsoft.DBforPostgreSQL/flexibleServers/backups
// (azurerm_postgresql_flexible_server_backup).
//
// This is a genuine ARM child resource type (an on-demand backup), not a data-plane
// operation: it has a stable resource id (.../flexibleServers/{name}/backups/{backup})
// and a GET/DELETE lifecycle. The create is a bodyless PUT — the backup content is
// entirely server-populated — so there are no settable body properties.
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_flexible_server_backup_resource.go
//     :43-63   (Arguments: name FlexibleServerBackupName ForceNew, server_id ForceNew;
//     completed_time computed)
//     :67,109,144 (timeouts: create 30m, read 5m, delete 30m)
//     :97      (bodyless create — BackupsAutomaticAndOnDemandCreate takes no payload)
//   - go-azure-sdk resource-manager/postgresql/2025-08-01/backupautomaticandondemands:
//     model_backupautomaticandondemandproperties.go (backupType/completedTime/source —
//     all server-populated), id_backup.go:123-125 (staticFlexibleServers/staticBackups)
type PostgresqlFlexibleServerBackup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlFlexibleServerBackup)(nil)

// NewPostgresqlFlexibleServerBackup returns knowledge for the
// flexibleServers/backups resource.
func NewPostgresqlFlexibleServerBackup() *PostgresqlFlexibleServerBackup {
	return &PostgresqlFlexibleServerBackup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/flexibleServers/backups",
			ApiVersions:  []string{"2025-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[-\w\._]+$`,
					Message:      "backup name may contain only letters, numbers, '-', '_' and '.'",
				},
			},
			ComputedFields: []string{
				"properties.backupType",
				"properties.completedTime",
				"properties.source",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlFlexibleServerBackup()) }
