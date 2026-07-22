package devtestlabs

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevTestLab provides resource knowledge for Microsoft.DevTestLab/labs.
//
// Contributing Terraform resource: azurerm_dev_test_lab.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devtestlabs/dev_test_lab_resource.go
//     (schema L51-96, Create body L127-130, Read mapping L167-191)
//   - terraform-provider-azurerm internal/services/devtestlabs/validate/devtest.go
//     (DevTestLabName L15-20)
//   - go-azure-sdk resource-manager/devtestlab/2018-09-15/labs:
//     model_lab.go, model_labproperties.go, id_lab.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned and ForceNew; only the name
//     naming regex is emitted (empty PropertyPath). There are no user-settable ARM body
//     properties in AzureRM's schema — the Create body only sets location + tags.
//   - The storage-account/vault/uniqueIdentifier properties are server-populated (Computed
//     in AzureRM), so they are listed as ComputedFields.
type DevTestLab struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevTestLab)(nil)

func NewDevTestLab() *DevTestLab {
	return &DevTestLab{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevTestLab/labs",
			ApiVersions:  []string{"2018-09-15"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name validation (empty PropertyPath validates the name).
				{Regex: `^[A-Za-z0-9_-]+$`, Message: "Lab Name can only include alphanumeric characters, underscores, hyphens."},
			},
			ComputedFields: []string{
				"properties.artifactsStorageAccount",
				"properties.defaultStorageAccount",
				"properties.defaultPremiumStorageAccount",
				"properties.vaultName",
				"properties.premiumDataDiskStorageAccount",
				"properties.uniqueIdentifier",
			},
		},
	}
}

func init() { azwise.Register(NewDevTestLab()) }
