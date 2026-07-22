package chaosstudio

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ChaosStudioExperiment provides resource knowledge for Microsoft.Chaos/experiments.
//
// Contributing Terraform resource: azurerm_chaos_studio_experiment.
//
// Sources:
//   - terraform-provider-azurerm internal/services/chaosstudio/chaos_studio_experiment_resource.go
//     (schema L84-184, Create body L217-241, timeouts Create/Read/Update/Delete
//     30m/5m/30m/30m)
//   - go-azure-sdk resource-manager/chaosstudio/2023-11-01/experiments:
//     model_experiment.go, model_experimentproperties.go (selectors + steps are Required,
//     provisioningState is read-only), model_action.go (action type discriminator)
//
// Notes:
//   - name (ForceNew), location, and resource_group_name are all envelope-owned; the only
//     ForceNew body impact is the resource name.
//   - selectors -> properties.selectors and steps -> properties.steps are Required arrays
//     (MinItems 1 in AzureRM; ARM model marks both non-omitempty required).
//   - All AzureRM validators on this resource sit inside arrays and are therefore
//     array-element paths, which azwise/azapin cannot lower -> skipped with note:
//       * selectors[*].name, steps[*].name, steps[*].branch[*].name = StringIsNotEmpty
//       * selectors[*].chaos_studio_target_ids[*] = commonids.ValidateChaosStudioTargetID
//         (semantic resource-ID validator)
//       * steps[*].branch[*].actions[*].action_type = StringInSlice(continuous/delay/
//         discrete), mapping to properties.steps[*].branch[*].actions[*].type
//   - identity is the top-level ARM identity envelope, managed by AzAPI directly.
type ChaosStudioExperiment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ChaosStudioExperiment)(nil)

// NewChaosStudioExperiment returns knowledge for Microsoft.Chaos/experiments.
func NewChaosStudioExperiment() *ChaosStudioExperiment {
	return &ChaosStudioExperiment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Chaos/experiments",
			ApiVersions:  []string{"2023-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			// selectors and steps are Required arrays on the ARM body.
			RequiredFields: []string{
				"properties.selectors",
				"properties.steps",
			},
			// provisioningState is returned by GET but absent from the create/update body.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewChaosStudioExperiment()) }
