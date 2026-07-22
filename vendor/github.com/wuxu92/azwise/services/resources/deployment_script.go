package resources

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DeploymentScript provides resource knowledge for Microsoft.Resources/deploymentScripts.
//
// The ARM type is a discriminated ROOT (kind: "AzureCLI" | "AzurePowerShell").
// AzureRM models the two variants as two separate resources that share one base
// schema; their shared operational knowledge is merged here for the single ARM type.
//
// Sources:
//   - terraform-provider-azurerm internal/services/resource/resource_deployment_script_common.go:30-250
//     (shared ResourceDeploymentScriptModel + getDeploymentScriptArguments/Attributes:
//     ForceNew flags, name regex, ISO8601 duration validators, cleanup_preference enum,
//     container/storage/environment_variable blocks, ExactlyOneOf script source, defaults)
//   - terraform-provider-azurerm internal/services/resource/resource_deployment_script_common.go:252-306
//     (updateDeploymentScript patches only Tags; deleteDeploymentScript — 30m timeouts, no soft-delete)
//   - terraform-provider-azurerm internal/services/resource/resource_deployment_script_azure_cli_resource.go:38-124,130-153
//     (AzureCli Arguments()/Create() mapping to deploymentscripts.AzureCliScript; Read 5m, Create 30m timeouts)
//   - terraform-provider-azurerm internal/services/resource/resource_deployment_script_azure_power_shell_resource.go:38-124
//     (AzurePowerShell variant mapping to deploymentscripts.AzurePowerShellScript)
//   - terraform-provider-azurerm internal/services/resource/validate/resource_deployment_script_azure_cli_version.go:11-23
//     (azCliVersion regex ^\d+\.\d+\.\d+$)
//   - terraform-provider-azurerm internal/services/resource/validate/resource_deployment_script_azure_power_shell_version.go:11-23
//     (azPowerShellVersion regex ^\d+\.\d+$)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/resources/2020-10-01/deploymentscripts/model_azurecliscriptproperties.go:6-22
//   - .../deploymentscripts/model_azurepowershellscriptproperties.go:6-22
//     (AzureCliScriptProperties / AzurePowerShellScriptProperties json tags — ARM body paths)
//   - .../deploymentscripts/model_containerconfiguration.go:6-8, model_storageaccountconfiguration.go:6-9,
//     model_environmentvariable.go:6-10, model_deploymentscript.go:21-30 (base envelope: kind/location/identity/tags)
//   - .../deploymentscripts/constants.go:12-26 (CleanupOptions: Always/OnExpiration/OnSuccess — full SDK set)
//   - internal/native/services/resources/resources_deployment_script_gen.go:186-811
//     (generated discriminated schema; property nesting azure_cli.properties.* / azure_power_shell.properties.*)
//
// Runtime note: the generated NATIVE resource (azapi_resources_deployment_script) has a
// discriminated body root, so typegraph.ApplyAzwise short-circuits (it only overlays a
// KindObject body). This knowledge therefore drives the generic azapi_resource path,
// where the raw ARM body is {kind, location, identity, tags, properties:{...}} and every
// path below resolves against the go-azure-sdk model. The native schema already bakes the
// per-variant Required (az_cli_version / az_power_shell_version), the cleanup_preference
// OneOf, and the azure_cli/azure_power_shell ExactlyOneOf — those are NOT re-emitted here.
//
// Intentionally skipped (documented, no rule emitted):
//   - environment_variable.secure_value (Sensitive): maps to
//     properties.environmentVariables[*].secureValue — an array-element path, which the
//     overlay/validator cannot resolve; left to a customizer if per-element redaction is needed.
//   - supporting_script_uris: properties.supportingScriptUris[*] is a string-array element
//     (no per-element rule); the whole array is still covered as ForceNew.
//   - retention_interval / timeout: AzureRM validates these with ISO8601DurationBetween
//     ("PT1H".."P1DT2H" and "PT1S".."P1D") — a semantic duration-range comparison, not a
//     regex/enum/length, so it is not expressible as a StringRule. Requires a customizer/hook
//     validator. (retention_interval is still captured as ForceNew + RequiredFields; timeout as
//     ForceNew + DefaultValue.)
//   - properties.azCliVersion / properties.azPowerShellVersion Required: variant-specific.
//     Marking both required would falsely flag the absent sibling on every body. The native
//     generator already enforces each as Required within its own variant block, so only the
//     always-present properties.retentionInterval is listed in RequiredFields.
type DeploymentScript struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DeploymentScript)(nil)

// NewDeploymentScript returns knowledge for the deploymentScripts resource.
func NewDeploymentScript() *DeploymentScript {
	return &DeploymentScript{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Resources/deploymentScripts",
			ApiVersions:  []string{"2023-08-01"},
			SoftDelete:   false,
			// Deployment scripts are largely immutable: Update() patches only tags,
			// so every other property forces replacement.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "identity"},
				{PropertyPath: "properties.arguments"},
				{PropertyPath: "properties.azCliVersion"},
				{PropertyPath: "properties.azPowerShellVersion"},
				{PropertyPath: "properties.cleanupPreference"},
				{PropertyPath: "properties.containerSettings"},
				{PropertyPath: "properties.environmentVariables"},
				{PropertyPath: "properties.forceUpdateTag"},
				{PropertyPath: "properties.primaryScriptUri"},
				{PropertyPath: "properties.retentionInterval"},
				{PropertyPath: "properties.scriptContent"},
				{PropertyPath: "properties.storageAccountSettings"},
				{PropertyPath: "properties.supportingScriptUris"},
				{PropertyPath: "properties.timeout"},
			},
			RequiredFields: []string{
				"properties.retentionInterval",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM: 1-260 chars,
					// alphanumeric / underscore / parentheses / hyphen / period, not ending with a period.
					Regex:     `^[a-zA-Z0-9_()-.]{0,259}[a-zA-Z0-9_()-]$`,
					MinLength: 1,
					MaxLength: 260,
					Message:   "name must be 1-260 characters of alphanumeric, underscore, parentheses, hyphen and period, and cannot end with a period",
				},
				{
					PropertyPath:  "properties.cleanupPreference",
					AllowedValues: []string{"Always", "OnExpiration", "OnSuccess"},
					Message:       "must be one of Always, OnExpiration, OnSuccess",
				},
				{
					// AzureCLI variant. Absent on AzurePowerShell bodies, so the rule is skipped there.
					PropertyPath: "properties.azCliVersion",
					Regex:        `^\d+\.\d+\.\d+$`,
					Message:      "az_cli_version should be in the format X.Y.Z (e.g. 2.30.0)",
				},
				{
					// AzurePowerShell variant. Absent on AzureCLI bodies, so the rule is skipped there.
					PropertyPath: "properties.azPowerShellVersion",
					Regex:        `^\d+\.\d+$`,
					Message:      "az_power_shell_version should be in the format X.Y (e.g. 9.7)",
				},
			},
			SensitiveFields: []string{
				"properties.storageAccountSettings.storageAccountKey",
			},
			ComputedFields: []string{
				"properties.outputs",
				"properties.provisioningState",
				"properties.status",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.cleanupPreference", Value: "Always"},
				{PropertyPath: "properties.timeout", Value: "P1D"},
			},
			// ExactlyOneOf of the two script-source properties (AzureRM ExactlyOneOf on
			// primary_script_uri / script_content). Both resolve one level under properties
			// in the ARM body. Distinct from the generator's azure_cli/azure_power_shell
			// variant ExactlyOneOf, which is emitted natively and NOT duplicated here.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.primaryScriptUri", "properties.scriptContent"},
					Message: "exactly one of primary_script_uri or script_content must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewDeploymentScript()) }
