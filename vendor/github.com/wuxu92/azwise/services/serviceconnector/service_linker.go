package serviceconnector

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceLinker provides resource knowledge for Microsoft.ServiceLinker/linkers.
//
// This ARM extension type ("linkers") is scoped onto a source resource (an app
// service, function app, spring cloud deployment, ...). AzureRM splits it into
// several typed resources that all create the SAME servicelinker.LinkerResource
// body; azwise is keyed by ARM type, so they MERGE into this single file.
// Only knowledge universal to every contributing resource is unioned here — the
// per-resource source scope (app_service_id / function_app_id / spring_cloud_id)
// is envelope/scope (part of the resource ID via NewScopedLinkerID) and is not a
// body field, so it contributes no body rules.
//
// Contributing TF resources:
//   - azurerm_app_service_connection    (app_service_connection_resource.go:38-92)
//   - azurerm_function_app_connection   (function_app_connection_resource.go:36-88)
//   - azurerm_spring_cloud_connection   (spring_cloud_connection_resource.go:37-111)
//
// Sources:
//   - helper.go:50-110 (authInfoSchema), :33-48 (secretStoreSchema),
//     :112-156 (expandServiceConnectorAuthInfoForCreate).
//   - app_service_connection_resource.go:106-182 (create -> servicelinker.LinkerResource),
//     :108/:186/:238/:257 timeouts (Create 30m / Read 5m / Delete 30m / Update 30m).
//   - Scoped extension ID: servicelinker id_scopedlinker.go:96-111 confirms casing
//     "Microsoft.ServiceLinker/linkers".
//   - client_type Default "none" is NOT unioned as a DefaultValue: app_service and
//     function_app default it, but spring_cloud (features.FivePointOh) drops the
//     default, so it is not universal.
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     target_resource_id (azure.ValidateResourceID -> properties.targetService.id,
//     resource-id validator); the per-resource source scope validators
//     (validate.AppServiceID / validate.FunctionAppID / validate.SpringCloudDeploymentID)
//     apply to the scope ID, not the body.
//   - go-azure-sdk resource-manager/servicelinker/2024-04-01/servicelinker
//     model_linkerproperties.go / model_authinfobase.go / model_secretauthinfo.go /
//     model_serviceprincipalsecretauthinfo.go / model_azureresource.go /
//     model_vnetsolution.go / model_secretstore.go / constants.go.
type ServiceLinker struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceLinker)(nil)

// NewServiceLinker returns knowledge for the ServiceLinker linkers extension resource.
func NewServiceLinker() *ServiceLinker {
	return &ServiceLinker{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceLinker/linkers",
			ApiVersions:  []string{"2024-04-01"},
			ForceNew: []azwise.ForceNewRule{
				// authentication.type (ForceNew) -> properties.authInfo.authType.
				{PropertyPath: "properties.authInfo.authType"},
				// target_resource_id (ForceNew) -> properties.targetService.id.
				{PropertyPath: "properties.targetService.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// authentication (Required) -> authInfo.authType discriminator;
			// target_resource_id (Required) -> targetService.id. targetService.type
			// is the ARM discriminator (hardcoded "AzureResource" by AzureRM).
			RequiredFields: []string{
				"properties.authInfo.authType",
				"properties.targetService.id",
			},
			StringRules: []azwise.StringRule{
				// authentication.type -> authInfo.authType: full ARM AuthType enum
				// (AzureRM restricts to a subset; ARM accepts the full set).
				{
					PropertyPath: "properties.authInfo.authType",
					AllowedValues: []string{
						"accessKey", "easyAuthMicrosoftEntraID", "secret",
						"servicePrincipalCertificate", "servicePrincipalSecret",
						"systemAssignedIdentity", "userAccount", "userAssignedIdentity",
					},
					Message: "authentication type must be a valid ServiceLinker AuthType",
				},
				// client_type -> properties.clientType: full ARM ClientType enum.
				{
					PropertyPath: "properties.clientType",
					AllowedValues: []string{
						"dapr", "django", "dotnet", "go", "java", "jms-springBoot",
						"kafka-springBoot", "nodejs", "none", "php", "python", "ruby",
						"springBoot",
					},
					Message: "client_type must be a valid ServiceLinker ClientType",
				},
				// vnet_solution -> properties.vNetSolution.type: VNetSolutionType enum.
				{
					PropertyPath:  "properties.vNetSolution.type",
					AllowedValues: []string{"privateLink", "serviceEndpoint"},
					Message:       "vnet_solution must be privateLink or serviceEndpoint",
				},
			},
			// authentication.secret / .certificate are Sensitive. The authInfo body is
			// polymorphic; these paths exist only for the matching authType and are
			// validated only when present.
			SensitiveFields: []string{
				"properties.authInfo.secret",
				"properties.authInfo.secretInfo.value",
				"properties.authInfo.certificate",
			},
			// Read-only body property present in LinkerProperties (GET) but
			// server-computed, not a user input.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewServiceLinker()) }
