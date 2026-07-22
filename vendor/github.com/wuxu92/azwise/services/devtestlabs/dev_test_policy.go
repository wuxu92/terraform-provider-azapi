package devtestlabs

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevTestPolicy provides resource knowledge for Microsoft.DevTestLab/labs/policySets/policies.
//
// Contributing Terraform resource: azurerm_dev_test_policy.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devtestlabs/dev_test_policy_resource.go
//     (schema L49-110, Create body L114-162)
//   - go-azure-sdk resource-manager/devtestlab/2018-09-15/policies:
//     model_policyproperties.go, constants.go (PolicyEvaluatorType, PolicyFactName,
//     PolicyStatus), id_policy.go (segments labs/{labName}/policySets/{policySetName}/policies/{policyName}).
//
// Notes:
//   - name/lab_name/policy_set_name are envelope-owned (ARM name segment + parent
//     labs/{labName}/policySets/{policySetName}) and ForceNew. The resource name is also
//     copied into properties.factName by AzureRM, and is validated against the PolicyFactName
//     enum — emitted as a name rule (empty PropertyPath) using the full ARM SDK enum set.
//   - evaluator_type maps to properties.evaluatorType (ForceNew, enum).
//   - threshold uses StringIsNotEmpty (MinLength 1).
type DevTestPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevTestPolicy)(nil)

func NewDevTestPolicy() *DevTestPolicy {
	return &DevTestPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevTestLab/labs/policySets/policies",
			ApiVersions:  []string{"2018-09-15"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.evaluatorType"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name is copied to properties.factName and validated against the
				// PolicyFactName enum (empty PropertyPath validates the name). Full ARM SDK set.
				{AllowedValues: []string{
					"EnvironmentTemplate",
					"GalleryImage",
					"LabPremiumVmCount",
					"LabTargetCost",
					"LabVmCount",
					"LabVmSize",
					"ScheduleEditPermission",
					"UserOwnedLabPremiumVmCount",
					"UserOwnedLabVmCount",
					"UserOwnedLabVmCountInSubnet",
				}},
				{PropertyPath: "properties.evaluatorType", AllowedValues: []string{"AllowedValuesPolicy", "MaxValuePolicy"}},
				{PropertyPath: "properties.threshold", MinLength: 1, Message: "must not be empty"},
			},
			RequiredFields: []string{
				"properties.threshold",
				"properties.evaluatorType",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.uniqueIdentifier",
				"properties.createdDate",
				"properties.status",
			},
		},
	}
}

func init() { azwise.Register(NewDevTestPolicy()) }
