package authorization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PimEligibleRoleAssignment provides resource knowledge for
// Microsoft.Authorization/roleEligibilityScheduleRequests.
//
// Mirrors azurerm_pim_eligible_role_assignment. AzureRM models an "eligible" PIM
// assignment by submitting a role eligibility schedule request with
// requestType = "AdminAssign". Every configurable argument is ForceNew.
//
// Unlike the active variant, the eligible request also carries condition /
// condition_version (ABAC), which map to properties.condition and
// properties.conditionVersion. AzureRM restricts condition_version to "2.0";
// azwise keeps that single-value enum as a StringRule.
//
// principal_id uses validation.IsUUID; that semantic check is handled at the
// schema/customizer layer, so it is not repeated as a StringRule here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/authorization/pim_eligible_role_assignment_resource.go
//     schema (all args ForceNew; requestType hardcoded "AdminAssign";
//     condition_version restricted to "2.0"; timeouts Create 10m / Read 5m /
//     Delete 10m; no Update)
//   - go-azure-sdk resource-manager/authorization/2020-10-01/roleeligibilityschedulerequests
//     RoleEligibilityScheduleRequestProperties: principalId/requestType/
//     roleDefinitionId required; justification/condition/conditionVersion/scope/
//     scheduleInfo/ticketInfo settable; approvalId/createdOn/expandedProperties/
//     requestorId/status/target* are read-only.
type PimEligibleRoleAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PimEligibleRoleAssignment)(nil)

// NewPimEligibleRoleAssignment returns knowledge for the roleEligibilityScheduleRequests resource.
func NewPimEligibleRoleAssignment() *PimEligibleRoleAssignment {
	return &PimEligibleRoleAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/roleEligibilityScheduleRequests",
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
				{PropertyPath: "properties.condition"},
				{PropertyPath: "properties.conditionVersion"},
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
			// AzureRM hardcodes requestType to "AdminAssign" for an eligible assignment.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.requestType", Value: "AdminAssign"},
			},
			// condition requires condition_version and vice versa (azurerm RequiredWith).
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.condition", "properties.conditionVersion"}},
				{Paths: []string{"properties.conditionVersion", "properties.condition"}},
			},
			StringRules: []azwise.StringRule{
				// Full ARM SDK RequestType enum (AzAPI sends raw ARM values).
				{PropertyPath: "properties.requestType", AllowedValues: []string{
					"AdminAssign", "AdminExtend", "AdminRemove", "AdminRenew", "AdminUpdate",
					"SelfActivate", "SelfDeactivate", "SelfExtend", "SelfRenew",
				}, Message: "requestType must be a valid role eligibility schedule request type"},
				// Full ARM SDK expiration Type enum.
				{PropertyPath: "properties.scheduleInfo.expiration.type", AllowedValues: []string{
					"AfterDateTime", "AfterDuration", "NoExpiration",
				}, Message: "expiration type must be AfterDateTime, AfterDuration or NoExpiration"},
				// AzureRM restricts condition_version to "2.0".
				{PropertyPath: "properties.conditionVersion", AllowedValues: []string{"2.0"}, Message: "condition version must be 2.0"},
				{PropertyPath: "properties.condition", MinLength: 1, Message: "condition must not be empty"},
			},
			// Read-only request properties returned by GET but not settable.
			ComputedFields: []string{
				"properties.approvalId",
				"properties.createdOn",
				"properties.expandedProperties",
				"properties.requestorId",
				"properties.status",
				"properties.targetRoleEligibilityScheduleId",
				"properties.targetRoleEligibilityScheduleInstanceId",
			},
		},
	}
}

func init() { azwise.Register(NewPimEligibleRoleAssignment()) }
