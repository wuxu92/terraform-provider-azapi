package blueprints

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BlueprintAssignment provides resource knowledge for
// Microsoft.Blueprint/blueprintAssignments.
//
// Mirrors azurerm_blueprint_assignment. name, target_subscription_id, location and
// identity are operational-envelope inputs: name is the ARM resource name (ForceNew in
// AzureRM but not a body property), target_subscription_id is the assignment SCOPE
// (ForceNew; surfaced in the body only as the derived properties.scope, not a
// user-settable field), and identity is the top-level managed-identity envelope. None
// are encoded as body ForceNew rules.
//
// The settable body inputs are:
//   - version_id            -> properties.blueprintId (Required; the published-version id)
//   - parameter_values      -> properties.parameters (JSON object)
//   - resource_groups       -> properties.resourceGroups (JSON object)
//   - lock_mode             -> properties.locks.mode (enum, default "None")
//   - lock_exclude_principals -> properties.locks.excludedPrincipals (max 5, each a UUID)
//   - lock_exclude_actions  -> properties.locks.excludedActions (max 200)
//
// lock_exclude_principals validates each element as a UUID; that is an array-ELEMENT
// constraint (properties.locks.excludedPrincipals[*]) which azwise's declarative rules
// cannot express, so only the MaxItems structural bound is captured here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/blueprints/blueprint_assignment_resource.go
//     (Schema L48-140, CreateUpdate body L166-208; timeouts 30m/30m/5m/5m)
//   - go-azure-sdk resource-manager/blueprints/2018-11-01-preview/assignment
//     Assignment (model_assignment.go), AssignmentProperties (model_assignmentproperties.go:
//     blueprintId/parameters/resourceGroups/locks settable; description/displayName/
//     provisioningState/status output), AssignmentLockSettings (model_assignmentlocksettings.go:
//     mode/excludedPrincipals/excludedActions), AssignmentLockMode enum (constants.go
//     L53-59: AllResourcesDoNotDelete/AllResourcesReadOnly/None).
type BlueprintAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BlueprintAssignment)(nil)

// NewBlueprintAssignment returns knowledge for the blueprintAssignments resource.
func NewBlueprintAssignment() *BlueprintAssignment {
	return &BlueprintAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Blueprint/blueprintAssignments",
			ApiVersions:  []string{"2018-11-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 5 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name uses validation.StringIsNotEmpty.
					PropertyPath: "",
					MinLength:    1,
					Message:      "blueprint assignment name must not be empty",
				},
				{
					PropertyPath:  "properties.locks.mode",
					AllowedValues: []string{"AllResourcesDoNotDelete", "AllResourcesReadOnly", "None"},
					Message:       "lock_mode must be one of AllResourcesDoNotDelete, AllResourcesReadOnly or None",
				},
			},
			ArrayRules: []azwise.ArrayRule{
				{
					PropertyPath: "properties.locks.excludedPrincipals",
					MaxItems:     5,
					Message:      "lock_exclude_principals allows at most 5 entries",
				},
				{
					PropertyPath: "properties.locks.excludedActions",
					MaxItems:     200,
					Message:      "lock_exclude_actions allows at most 200 entries",
				},
			},
			// version_id is Required and maps to the create body.
			RequiredFields: []string{
				"properties.blueprintId",
			},
			// AzureRM defaults lock_mode to "None" when omitted.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.locks.mode", Value: "None"},
			},
			// Output-only body fields: provisioning state and the assignment status are
			// server-computed and returned by GET but never part of the create intent.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.status",
			},
		},
	}
}

func init() { azwise.Register(NewBlueprintAssignment()) }
