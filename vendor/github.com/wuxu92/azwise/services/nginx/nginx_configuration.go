package nginx

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NginxConfiguration provides resource knowledge for Nginx.NginxPlus/nginxDeployments/configurations.
//
// Mirrors azurerm_nginx_configuration. The configuration name is hardcoded "default".
//
// ARM type casing verified against go-azure-sdk nginxconfiguration/id_configuration.go Segments:
// "Nginx.NginxPlus" / "nginxDeployments" / "configurations".
//
// Sources:
//   - terraform-provider-azurerm internal/services/nginx/nginx_configuration_resource.go
//     Arguments (99-170: nginx_deployment_id parent ref ForceNew; config_file set with
//     content(StringIsBase64)/virtual_path AtLeastOneOf{config_file,package_data}; protected_file
//     set RequiredWith{config_file} with content(Sensitive)/virtual_path/content_hash(Computed);
//     package_data StringIsNotEmpty AtLeastOneOf{config_file,package_data} ConflictsWith
//     {protected_file,config_file}; root_file StringIsNotEmpty Required), ToSDKModel (74-93),
//     timeouts Create 30m / Read 5m / Update 10m / Delete 10m.
//   - go-azure-sdk resource-manager/nginx/2024-11-01-preview/nginxconfiguration
//     NginxConfigurationRequestProperties (files/package/protectedFiles/rootFile settable;
//     provisioningState read-only).
//
// Note: config_file/protected_file content(StringIsBase64), virtual_path(StringIsNotEmpty) and the
// computed content_hash live at array-element paths (properties.files[*], properties.protectedFiles[*]);
// azwise cannot resolve array-element paths so those constraints are documented but not emitted.
type NginxConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NginxConfiguration)(nil)

// NewNginxConfiguration returns knowledge for the Nginx configurations resource.
func NewNginxConfiguration() *NginxConfiguration {
	return &NginxConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Nginx.NginxPlus/nginxDeployments/configurations",
			ApiVersions:  []string{"2024-11-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			RequiredFields: []string{
				"properties.rootFile",
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.rootFile", MinLength: 1, Message: "root_file must not be empty"},
				{PropertyPath: "properties.package.data", MinLength: 1, Message: "package_data must not be empty"},
			},
			// config_file AtLeastOneOf {config_file, package_data} (schema line 111, 160).
			AtLeastOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.files", "properties.package.data"},
					Message: "at least one of `config_file` or `package_data` must be set",
				},
			},
			// protected_file RequiredWith config_file (schema line 132).
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.protectedFiles", "properties.files"},
					Message: "`protected_file` requires `config_file`",
				},
			},
			// package_data ConflictsWith {protected_file, config_file} (schema line 161).
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.package.data", "properties.protectedFiles", "properties.files"},
					Message: "`package_data` conflicts with `protected_file` and `config_file`",
				},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewNginxConfiguration()) }
