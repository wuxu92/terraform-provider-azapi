package hsm

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DedicatedHardwareSecurityModule provides resource knowledge for
// Microsoft.HardwareSecurityModules/dedicatedHSMs.
//
// Contributing TF resource: azurerm_dedicated_hardware_security_module.
//
// Sources:
//   - AzureRM internal/services/hsm/dedicated_hardware_security_module_resource.go
//     (schema :48-140, Create body :171-192, expand network profile :294-323)
//   - AzureRM internal/services/hsm/validate/dedicated_hardware_security_module_name.go
//     (name regex :19, no-consecutive-hyphens :26)
//   - go-azure-sdk .../hardwaresecuritymodules/2021-11-30/dedicatedhsms:
//     model_dedicatedhsm.go, model_dedicatedhsmproperties.go, model_sku.go,
//     model_networkprofile.go, model_networkinterface.go, constants.go
//     (SkuName enum :70-78, ARM body paths)
type DedicatedHardwareSecurityModule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DedicatedHardwareSecurityModule)(nil)

func NewDedicatedHardwareSecurityModule() *DedicatedHardwareSecurityModule {
	return &DedicatedHardwareSecurityModule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HardwareSecurityModules/dedicatedHSMs",
			ApiVersions:  []string{"2021-11-30"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "sku.name"},
				{PropertyPath: "zones"},
				{PropertyPath: "properties.stampId"},
				// network_profile inner fields are ForceNew (subnet_id + IPs).
				{PropertyPath: "properties.networkProfile.subnet.id"},
				{PropertyPath: "properties.networkProfile.networkInterfaces"},
				// management_network_profile inner fields are ForceNew.
				{PropertyPath: "properties.managementNetworkProfile.subnet.id"},
				{PropertyPath: "properties.managementNetworkProfile.networkInterfaces"},
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
					// Resource name (empty PropertyPath). validate.DedicatedHardwareSecurityModuleName:
					// 3-24 alphanumeric chars, begin with a letter, end with a letter or digit,
					// no consecutive hyphens.
					Regex:     `^[a-zA-Z][a-zA-Z0-9-]{1,22}[a-zA-Z0-9]$`,
					MinLength: 3,
					MaxLength: 24,
					Message:   "must be 3-24 alphanumeric characters, begin with a letter, end with a letter or digit, and contain no consecutive hyphens",
				},
				{
					// dedicatedhsms.PossibleValuesForSkuName() (full ARM set).
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"payShield10K_LMK1_CPS60",
						"payShield10K_LMK1_CPS250",
						"payShield10K_LMK1_CPS2500",
						"payShield10K_LMK2_CPS60",
						"payShield10K_LMK2_CPS250",
						"payShield10K_LMK2_CPS2500",
						"SafeNet Luna Network HSM A790",
					},
					Message: "must be a valid dedicated HSM SKU name",
				},
				{
					// stamp_id: validation.StringInSlice{"stamp1","stamp2"}.
					PropertyPath:  "properties.stampId",
					AllowedValues: []string{"stamp1", "stamp2"},
					Message:       "must be stamp1 or stamp2",
				},
			},
			RequiredFields: []string{
				"sku.name",
				"properties.networkProfile.subnet.id",
				"properties.networkProfile.networkInterfaces",
			},
			// provisioningState/statusMessage are Azure-populated read-only status fields.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.statusMessage",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDedicatedHardwareSecurityModule()) }
