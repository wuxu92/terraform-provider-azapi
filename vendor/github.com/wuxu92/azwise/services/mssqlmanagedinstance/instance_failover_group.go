// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package mssqlmanagedinstance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// InstanceFailoverGroup provides resource knowledge for
// Microsoft.Sql/locations/instanceFailoverGroups
// (Terraform azurerm_mssql_managed_instance_failover_group).
//
// NOTE the ARM scope: the failover group is a child of a LOCATION, not of the
// managed instance — id shape is
// /subscriptions/../resourceGroups/../providers/Microsoft.Sql/locations/{location}/instanceFailoverGroups/{name}
// (verified against SDK instancefailovergroups.InstanceFailoverGroupId parser and
// resourceids.go go:generate). Body is
// {properties:{readWriteEndpoint, readOnlyEndpoint, secondaryType,
// managedInstancePairs[], partnerRegions[]}} against
// instancefailovergroups.InstanceFailoverGroup.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssqlmanagedinstance/mssql_managed_instance_failover_group_resource.go:67-129
//     (Arguments: mode enum, grace_minutes IntAtLeast(60), secondary_type enum)
//   - .../mssql_managed_instance_failover_group_resource.go:155-238 (Create → properties mapping)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/instancefailovergroups:
//     model_instancefailovergroupproperties.go:6-14,
//     model_instancefailovergroupreadwriteendpoint.go, constants.go
//     (ReadWriteEndpointFailoverPolicy, ReadOnlyEndpointFailoverPolicy, SecondaryInstanceType)
//   - .../instancefailovergroups/id_instancefailovergroup.go:116-127 (locations/instanceFailoverGroups)
//
// Intentionally not encoded (documented, not emitted):
//   - partner_managed_instance_id (ForceNew) maps to
//     properties.managedInstancePairs[*].partnerManagedInstanceId — an array
//     element path, which azwise ForceNewRule cannot address; the managed instance
//     pairs and partnerRegions arrays are assembled by AzureRM from the ids, so no
//     declarative rule is emitted for them.
//   - name / location / managed_instance_id are envelope/parent RequiresReplace.
type InstanceFailoverGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*InstanceFailoverGroup)(nil)

// NewInstanceFailoverGroup returns knowledge for Microsoft.Sql/locations/instanceFailoverGroups.
func NewInstanceFailoverGroup() *InstanceFailoverGroup {
	return &InstanceFailoverGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/locations/instanceFailoverGroups",
			ApiVersions:  []string{"2023-08-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{},
			StringRules: []azwise.StringRule{
				// read_write_endpoint_failover_policy.mode → properties.readWriteEndpoint.failoverPolicy.
				{
					PropertyPath:  "properties.readWriteEndpoint.failoverPolicy",
					AllowedValues: []string{"Automatic", "Manual"},
					Message:       "read_write_endpoint_failover_policy.mode must be one of Automatic or Manual",
				},
				// readonly_endpoint_failover_policy_enabled → properties.readOnlyEndpoint.failoverPolicy.
				{
					PropertyPath:  "properties.readOnlyEndpoint.failoverPolicy",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "properties.readOnlyEndpoint.failoverPolicy must be one of Disabled or Enabled",
				},
				// secondary_type → properties.secondaryType.
				{
					PropertyPath:  "properties.secondaryType",
					AllowedValues: []string{"Geo", "Standby"},
					Message:       "secondary_type must be one of Geo or Standby",
				},
			},
			IntRules: []azwise.IntRule{
				// grace_minutes → properties.readWriteEndpoint.failoverWithDataLossGracePeriodMinutes (IntAtLeast 60).
				{
					PropertyPath: "properties.readWriteEndpoint.failoverWithDataLossGracePeriodMinutes",
					MinValue:     azwise.Ptr(int64(60)),
					Message:      "grace_minutes must be at least 60",
				},
			},
			FloatRules:      []azwise.FloatRule{},
			ArrayRules:      []azwise.ArrayRule{},
			SensitiveFields: []string{},
			ComputedFields:  []string{},
			DefaultValues: []azwise.DefaultValue{
				// readonly_endpoint_failover_policy_enabled default true → Enabled.
				{PropertyPath: "properties.readOnlyEndpoint.failoverPolicy", Value: "Enabled"},
				{PropertyPath: "properties.secondaryType", Value: "Geo"},
			},
			// read_write_endpoint_failover_policy is Required (mode Required within it).
			RequiredFields: []string{
				"properties.readWriteEndpoint.failoverPolicy",
			},
		},
	}
}

func init() { azwise.Register(NewInstanceFailoverGroup()) }
