package authorization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PimActiveRoleAssignment provides resource knowledge for
// Microsoft.Authorization/roleAssignmentScheduleRequests.
//
// Mirrors azurerm_pim_active_role_assignment. AzureRM models an "active" PIM
// assignment by submitting a role assignment schedule request with
// requestType = "AdminAssign". Every configurable argument is ForceNew, so the
// mapped ARM body properties are all replace-on-change.
//
// The envelope carries the request name (a generated GUID) and the scope
// (azurerm's scope), so azurerm's ForceNew on scope also lives on properties.scope
// which is included below because the ARM body echoes the scope.
//
// principal_id uses validation.IsUUID; that semantic check is handled at the
// schema/customizer layer (ApplyAzwise does not apply a UUID rule here), so it is
// intentionally not repeated as a StringRule. The schedule expiration is derived
// from duration_days / duration_hours / end_date_time in azurerm; those collapse
// into properties.scheduleInfo.expiration.{duration,endDateTime,type} in the ARM
// body.
//
// Sources:
//   - terraform-provider-azurerm internal/services/authorization/pim_active_role_assignment_resource.go
//     schema (all args ForceNew; requestType hardcoded "AdminAssign"; timeouts
//     Create 10m / Read 5m / Delete 10m; no Update)
//   - go-azure-sdk resource-manager/authorization/2020-10-01/roleassignmentschedulerequests
//     RoleAssignmentScheduleRequestProperties: principalId/requestType/roleDefinitionId
//     required; justification/scope/scheduleInfo/ticketInfo settable; approvalId/
//     createdOn/expandedProperties/requestorId/status/target* are read-only.
type PimActiveRoleAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PimActiveRoleAssignment)(nil)

// NewPimActiveRoleAssignment returns knowledge for the roleAssignmentScheduleRequests resource.
func NewPimActiveRoleAssignment() *PimActiveRoleAssignment {
	return &PimActiveRoleAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/roleAssignmentScheduleRequests",
			ApiVersions:  []string{"2020-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 10 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 10 * time.Minute,
			},
			// AzureRM marks every configurable argument ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.roleDefinitionId"},
				{PropertyPath: "properties.principalId"},
				{PropertyPath: "properties.scope"},
				{PropertyPath: "properties.justification"},
				{PropertyPath: "properties.scheduleInfo.startDateTime"},
				{PropertyPath: "properties.scheduleInfo.expiration.type"},
				{PropertyPath: "properties.scheduleInfo.expiration.duration"},
				{PropertyPath: "properties.scheduleInfo.expiration.endDateTime"},
				{PropertyPath: "properties.ticketInfo.ticketNumber"},
				{PropertyPath: "properties.ticketInfo.ticketSystem"},
			},
			RequiredFields: []string{
				"properties.principalId",
				"properties.roleDefinitionId",
				"properties.requestType",
			},
			// AzureRM hardcodes requestType to "AdminAssign" for an active assignment.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.requestType", Value: "AdminAssign"},
			},
			StringRules: []azwise.StringRule{
				// Full ARM SDK RequestType enum (AzAPI sends raw ARM values).
				{PropertyPath: "properties.requestType", AllowedValues: []string{
					"AdminAssign", "AdminExtend", "AdminRemove", "AdminRenew", "AdminUpdate",
					"SelfActivate", "SelfDeactivate", "SelfExtend", "SelfRenew",
				}, Message: "requestType must be a valid role assignment schedule request type"},
				// Full ARM SDK expiration Type enum.
				{PropertyPath: "properties.scheduleInfo.expiration.type", AllowedValues: []string{
					"AfterDateTime", "AfterDuration", "NoExpiration",
				}, Message: "expiration type must be AfterDateTime, AfterDuration or NoExpiration"},
			},
			// Read-only request properties returned by GET but not settable.
			ComputedFields: []string{
				"properties.approvalId",
				"properties.createdOn",
				"properties.expandedProperties",
				"properties.requestorId",
				"properties.status",
				"properties.targetRoleAssignmentScheduleId",
				"properties.targetRoleAssignmentScheduleInstanceId",
			},
		},
	}
}

func init() { azwise.Register(NewPimActiveRoleAssignment()) }
