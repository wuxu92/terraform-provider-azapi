package hybridcompute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ArcMachine provides resource knowledge for Microsoft.HybridCompute/machines.
//
// Contributing TF resource: azurerm_arc_machine.
//
// Sources:
//   - AzureRM internal/services/hybridcompute/arc_machine_resource.go
//     (model :22-29, schema :45-69, Create body :106-111)
//   - go-azure-sdk .../hybridcompute/2024-07-10/machines:
//     model_machine.go (Kind at envelope level :14), constants.go
//     (ArcKindEnum :55-63)
type ArcMachine struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ArcMachine)(nil)

func NewArcMachine() *ArcMachine {
	return &ArcMachine{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HybridCompute/machines",
			ApiVersions:  []string{"2024-07-10"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "kind"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: validation.StringIsNotEmpty.
					MinLength: 1,
					Message:   "must not be empty",
				},
				{
					// kind: machines.PossibleValuesForArcKindEnum() (envelope-level `kind`).
					PropertyPath:  "kind",
					AllowedValues: []string{"AVS", "AWS", "EPS", "GCP", "HCI", "SCVMM", "VMware"},
					Message:       "must be a valid Arc machine kind",
				},
			},
			RequiredFields: []string{
				"kind",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewArcMachine()) }
