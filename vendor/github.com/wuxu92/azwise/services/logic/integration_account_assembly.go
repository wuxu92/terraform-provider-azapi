package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccountAssembly provides resource knowledge for
// Microsoft.Logic/integrationAccounts/assemblies.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account_assembly.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_assembly_resource.go
//     (schema L42-94, Create body L120-140)
//   - terraform-provider-azurerm internal/services/logic/validate/
//     integration_account_assembly_artifact_name.go, integration_account_assembly_name.go,
//     integration_account_assembly_version.go
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccountassemblies:
//     model_assemblydefinition.go, model_assemblyproperties.go, model_contentlink.go,
//     id_assembly.go
//
// Notes:
//   - name (ForceNew artifact name), integration_account_name (ForceNew parent segment) and
//     resource_group_name are envelope-owned; only the artifact-name regex is emitted.
//   - content maps to the opaque properties.content blob; content_link_uri maps to
//     properties.contentLink.uri. AzureRM requires at-least-one-of content/content_link_uri
//     (a config-level AtLeastOneOf), not expressible as a single body RequiredField.
//   - assembly_version has AzureRM default "0.0.0.0" (Optional, not Computed); emitted as a
//     DefaultValue on properties.assemblyVersion.
//   - contentType is hardcoded by AzureRM to "application/octet-stream" and is server/provider
//     controlled, not user-facing.
type IntegrationAccountAssembly struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccountAssembly)(nil)

// NewIntegrationAccountAssembly returns knowledge for the assemblies child resource.
func NewIntegrationAccountAssembly() *IntegrationAccountAssembly {
	return &IntegrationAccountAssembly{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts/assemblies",
			ApiVersions:  []string{"2019-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9-().]+$`,
					MaxLength:    80,
					Message:      "assembly resource name contains only letters, numbers, dots, parentheses and hyphens, up to 80 characters",
				},
				{
					PropertyPath: "properties.assemblyName",
					Regex:        `^[A-Za-z0-9.]+$`,
					MaxLength:    80,
					Message:      "assembly_name contains only letters, numbers and dots, up to 80 characters",
				},
				{
					PropertyPath: "properties.assemblyVersion",
					Regex:        `^([0-9]+.[0-9]+.[0-9]+.[0-9]+)$|^([0-9]+.[0-9]+)$`,
					Message:      "assembly_version must be in the format `major.minor.build.revision` (`build` and `revision` optional)",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.assemblyVersion", Value: "0.0.0.0"},
			},
			RequiredFields: []string{
				"properties.assemblyName",
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccountAssembly()) }
