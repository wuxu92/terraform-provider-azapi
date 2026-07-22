package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlFailoverGroup provides resource knowledge for Microsoft.Sql/servers/failoverGroups
// (azurerm_mssql_failover_group).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_failover_group_resource.go
//     :69-143  (name ForceNew; server_id ForceNew; partner_server Required;
//     read_write_endpoint_failover_policy Required: mode StringInSlice(Automatic/Manual),
//     grace_minutes IntAtLeast(60))
//     :151-171 (CustomizeDiff: grace_minutes must be >=60 for Automatic, unset for Manual)
//     :173-242 (create) / :244-292 (update, timeout 30m) / :294-347 (read 5m) / :349-375 (delete 30m)
//     (mapping: properties.databases / partnerServers / readOnlyEndpoint.failoverPolicy /
//     readWriteEndpoint.failoverPolicy / failoverWithDataLossGracePeriodMinutes)
//   - AzureRM internal/services/mssql/validate/mssql.go:30-36 (failover group name regex)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/failovergroups:
//     model_failovergroupproperties.go, model_failovergroupreadwriteendpoint.go,
//     constants.go (ReadWriteEndpointFailoverPolicy, ReadOnlyEndpointFailoverPolicy)
type MsSqlFailoverGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlFailoverGroup)(nil)

// NewMsSqlFailoverGroup returns knowledge for the servers/failoverGroups resource.
func NewMsSqlFailoverGroup() *MsSqlFailoverGroup {
	return &MsSqlFailoverGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/failoverGroups",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-z]([-0-9a-z]{0,61}[0-9a-z])?$`,
					Message:      "name can contain only lowercase letters, numbers and '-', and can't start or end with '-'",
				},
				{
					PropertyPath:  "properties.readWriteEndpoint.failoverPolicy",
					AllowedValues: []string{"Automatic", "Manual"},
					Message:       "read_write_endpoint_failover_policy.mode must be one of Automatic or Manual",
				},
			},
			IntRules: []azwise.IntRule{
				{
					// grace_minutes is only valid (and required) when mode is Automatic;
					// AzureRM enforces the conditional in CustomizeDiff. The declarative
					// floor of 60 holds whenever the field is present.
					PropertyPath: "properties.readWriteEndpoint.failoverWithDataLossGracePeriodMinutes",
					MinValue:     azwise.Ptr(int64(60)),
					Message:      "grace_minutes must be at least 60 when read_write_endpoint_failover_policy.mode is Automatic",
				},
			},
			RequiredFields: []string{
				"properties.partnerServers",
				"properties.readWriteEndpoint.failoverPolicy",
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlFailoverGroup()) }
